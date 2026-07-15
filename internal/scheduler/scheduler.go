package scheduler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ButterHost69/sloper/internal/agent"
	myGit "github.com/ButterHost69/sloper/internal/git"
	myGithub "github.com/ButterHost69/sloper/internal/github"
	"github.com/ButterHost69/sloper/internal/logger"
	"github.com/ButterHost69/sloper/internal/models"
	"github.com/ButterHost69/sloper/internal/pipeline"
	"github.com/ButterHost69/sloper/internal/session"
	"github.com/ButterHost69/sloper/internal/slash"
	"github.com/ButterHost69/sloper/internal/storage"
	"github.com/ButterHost69/sloper/internal/worktree"
	"go.uber.org/zap"
)

const (
	PollInterval = 60 * time.Second
)

type Scheduler struct {
	RepoName   string
	RepoPath   string
	RepoUrl    string
	BotUser    string // GH_USERNAME — comments from this user are sloper's own
	sessionDir string // ~/.sloper/sessions/ — isolated pi session storage

	db        *storage.Repositories
	ghClient  *myGithub.GithubGateway
	gitClient *myGit.GitGateway
	agentGw   *agent.AgentGateway
	pl        *pipeline.Pipeline
	wtMgr     *worktree.Manager
	log       *zap.Logger
}

func New(repoPath string, db *storage.Repositories) *Scheduler {
	return &Scheduler{
		RepoPath: repoPath,
		db:       db,
		log:      logger.Default(),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.log.Info("scheduler: starting", zap.String("repo", s.RepoPath))

	s.ghClient = myGithub.NewGithubGateway(models.GithubOptions{CWD: s.RepoPath})

	s.gitClient = myGit.New(models.GitGatewayOptions{})

	repoName, err := s.gitClient.DetectGitHubRepo(ctx, s.RepoPath)
	if err != nil {
		s.log.Error("scheduler: detect repo failed", zap.Error(err))
		return
	}
	s.RepoName = repoName
	s.BotUser = os.Getenv("GH_USERNAME")
	s.log.Info("scheduler: repo detected",
		zap.String("repo", s.RepoName),
		zap.String("bot_user", s.BotUser))

	homeDir, _ := os.UserHomeDir()
	s.sessionDir = filepath.Join(homeDir, ".sloper", "sessions")
	if err := os.MkdirAll(s.sessionDir, 0o755); err != nil {
		s.log.Warn("scheduler: failed to create session dir", zap.Error(err))
	}

	s.agentGw = agent.NewAgentGateway(models.AgentOptions{
		CWD:        s.RepoPath,
		Model:      os.Getenv("AGENT_MODEL"),
		Thinking:   "high",
		APIKey:     os.Getenv("AGENT_KEY"),
		Provider:   os.Getenv("AGENT_PROVIDER"),
		SessionDir: s.sessionDir,
	})
	s.pl = pipeline.New(s.agentGw)

	s.wtMgr = worktree.NewManager("", s.gitClient)

	s.tick(ctx)

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler: stopping")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	s.log.Info("scheduler: tick started")
	s.db.AppendEvent(ctx, storage.EventRecord{
		EventType: "tick.started",
	})

	issues, err := s.ghClient.GetAllOpenIssuesRaw(ctx, models.GithubIssueOptions{
		Repo:  s.RepoName,
		CWD:   s.RepoPath,
		Limit: 30,
	})
	if err != nil {
		s.log.Error("scheduler: list issues failed", zap.Error(err))
		s.db.AppendEvent(ctx, storage.EventRecord{
			EventType: "tick.failed",
			Message:   err.Error(),
		})
		return
	}

	s.log.Info("scheduler: found open issues", zap.Int("count", len(issues)))

	for _, summary := range issues {
		if err := ctx.Err(); err != nil {
			return
		}
		if summary.IsPullRequest {
			continue
		}
		if err := s.processOne(ctx, summary); err != nil {
			s.log.Error("scheduler: process issue failed",
				logger.WithIssue(summary.Number), zap.Error(err))
		}
	}

	s.processApprovedIssues(ctx)
	s.processWorkDoneIssues(ctx)
	s.processReviewDoneIssues(ctx)

	s.db.AppendEvent(ctx, storage.EventRecord{
		EventType: "tick.completed",
	})
	s.log.Info("scheduler: tick completed")
}

func (s *Scheduler) processOne(ctx context.Context, summary models.GithubIssueSummary) error {
	log := s.log.With(logger.WithIssue(summary.Number))

	cached, err := s.db.GetIssue(ctx, summary.Number)
	if err != nil {
		return fmt.Errorf("get cached issue: %w", err)
	}

	if cached != nil && cached.UpdatedAt == summary.UpdatedAt {
		return nil
	}

	issue, err := s.ghClient.ViewIssue(ctx, models.ViewIssueInput{
		Repo:        s.RepoName,
		IssueNumber: summary.Number,
		CWD:         s.RepoPath,
	})
	if err != nil {
		return fmt.Errorf("view issue: %w", err)
	}

	// Cache all comments in DB, marking bot's own comments as processed.
	// Exception: bot's /sloper commands stay unprocessed so they get handled.
	for _, c := range issue.Comments {
		isBot := s.BotUser != "" && c.Author == s.BotUser
		isSloperCmd := slash.IsValidCommand(c.Body)
		_ = s.db.InsertComment(ctx, storage.CommentRecord{
			ID:          c.ID,
			IssueNumber: issue.Number,
			Author:      c.Author,
			Body:        c.Body,
			CreatedAt:   c.CreatedAt,
			Processed:   isBot && !isSloperCmd,
		})
	}

	maxCommentID := int64(0)
	for _, c := range issue.Comments {
		if c.ID > maxCommentID {
			maxCommentID = c.ID
		}
	}

	stage := models.StageNew
	if cached != nil {
		stage = cached.Stage
	}

	rec := storage.IssueRecordFromModel(issue, stage)
	rec.LastCommentID = maxCommentID
	if cached != nil {
		rec.SpecJSON = cached.SpecJSON
		rec.BranchName = cached.BranchName
		rec.PRNumber = cached.PRNumber
		rec.ReviewIterations = cached.ReviewIterations
	}

	if cached == nil {
		log.Info("scheduler: new issue detected",
			zap.String("title", issue.Title), zap.Int("comments", len(issue.Comments)))
		_ = s.db.UpsertIssue(ctx, rec)
		return s.runSpecStage(ctx, issue, "")
	}

	log.Info("scheduler: processing existing issue",
		zap.String("title", issue.Title),
		zap.String("stage", stage))

	// Update the cached issue record with the latest GitHub state.
	// This prevents re-processing on the next tick just because updatedAt changed.
	_ = s.db.UpsertIssue(ctx, rec)

	// Issues past SPEC stage are handled by dedicated processors
	// (processApprovedIssues, processWorkDoneIssues, etc.).
	// Don't re-triage or re-process slash commands for them.
	if stage == models.StageApproved || stage == models.StageWorkDone ||
		stage == models.StageReviewDone || stage == models.StageMerged {
		_ = s.db.UpdateIssueLastCommentID(ctx, issue.Number, maxCommentID)
		return nil
	}

	// Check for unprocessed comments (bot comments are auto-marked processed)
	unprocessed, err := s.db.GetUnprocessedComments(ctx, issue.Number)
	if err != nil {
		return fmt.Errorf("get unprocessed comments: %w", err)
	}

	// Filter to only HUMAN unprocessed comments, plus bot /sloper commands.
	var humanUnprocessed []storage.CommentRecord
	for _, c := range unprocessed {
		if s.BotUser != "" && c.Author == s.BotUser && !slash.IsValidCommand(c.Body) {
			_ = s.db.MarkCommentProcessed(ctx, c.ID)
			continue
		}
		humanUnprocessed = append(humanUnprocessed, c)
	}

	// Check for slash commands from human unprocessed comments
	commands := slash.ParseComments(issue.Comments)
	hasNewCommands := false
	for _, c := range humanUnprocessed {
		if slash.IsValidCommand(c.Body) {
			hasNewCommands = true
		}
	}

	if hasNewCommands {
		log.Info("scheduler: new slash commands detected",
			zap.Int("unprocessed", len(humanUnprocessed)))
		return s.handleSlashCommands(ctx, issue, commands, humanUnprocessed)
	}

	// Check for new human comments (not from bot) that would trigger re-triage
	newHumanCommentID := int64(0)
	for _, c := range issue.Comments {
		isBot := s.BotUser != "" && c.Author == s.BotUser
		if !isBot && c.ID > cached.LastCommentID {
			if c.ID > newHumanCommentID {
				newHumanCommentID = c.ID
			}
		}
	}

	if newHumanCommentID > 0 {
		log.Info("scheduler: new human comments detected, re-triaging")
		feedback := collectFeedbackFromCommentsExcludingBot(issue.Comments, cached.LastCommentID, s.BotUser)
		_ = s.db.UpdateIssueLastCommentID(ctx, issue.Number, maxCommentID)
		return s.runSpecStage(ctx, issue, feedback)
	}

	// No new human comments — just update last_comment_id to include bot comments
	_ = s.db.UpdateIssueLastCommentID(ctx, issue.Number, maxCommentID)

	return nil
}

func (s *Scheduler) runSpecStage(ctx context.Context, issue models.IssueDetail, feedback string) error {
	log := s.log.With(logger.WithIssue(issue.Number), logger.WithStage("spec"))

	runID, _ := s.db.StartRun(ctx, issue.Number, "spec")

	log.Info("scheduler: starting SPEC stage",
		zap.Bool("has_feedback", feedback != ""))
	s.db.AppendEvent(ctx, storage.EventRecord{
		IssueNumber: issue.Number,
		EventType:   "spec.started",
		Stage:       "spec",
	})

	spec, err := s.pl.SpecIssue(ctx, issue, feedback, s.specSessionID(issue.Number))
	if err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		s.db.AppendEvent(ctx, storage.EventRecord{
			IssueNumber: issue.Number,
			EventType:   "spec.failed",
			Stage:       "spec",
			Message:     err.Error(),
		})
		s.transitionIssueStage(ctx, issue.Number, models.StageFailed)
		return fmt.Errorf("spec: %w", err)
	}

	if spec.Summary == "" || len(spec.FilesToChange) == 0 || spec.ImplementationPlan == "" {
		log.Warn("scheduler: spec has empty fields — agent output may not have been parsed correctly",
			zap.String("summary", spec.Summary),
			zap.Int("files_count", len(spec.FilesToChange)),
			zap.Bool("has_plan", spec.ImplementationPlan != ""),
			zap.Int("raw_output_len", len(spec.RawOutput)))
		if len(spec.RawOutput) > 0 {
			preview := spec.RawOutput
			if len(preview) > 1000 {
				preview = preview[:1000] + "...(truncated)"
			}
			log.Warn("scheduler: raw agent output", zap.String("output", preview))
		}

		// Agent produced no usable output — likely an LLM error.
		// Post an error comment and mark as failed (don't proceed to WORK).
		isError := strings.Contains(spec.RawOutput, "\"stopReason\":\"error\"") ||
			strings.Contains(spec.RawOutput, "error") ||
			strings.Contains(spec.RawOutput, "Error")

		if isError {
			errMsg := spec.RawOutput
			if len(errMsg) > 2000 {
				errMsg = errMsg[:2000] + "..."
			}
			s.ghClient.PostIssueComment(ctx, s.RepoName, issue.Number,
				fmt.Sprintf("## Sloper encountered an error during analysis\n\n"+
					"```\n%s\n```\n\nPlease check the model/provider configuration "+
					"in `.env` (AGENT_MODEL, AGENT_PROVIDER, AGENT_KEY).\n\n"+
					"Use `/sloper retry` to try again after fixing the configuration.",
					errMsg))
		}

		s.transitionIssueStage(ctx, issue.Number, models.StageFailed)
		s.db.FailRun(ctx, runID, "agent produced empty output — likely LLM error")
		s.db.AppendEvent(ctx, storage.EventRecord{
			IssueNumber: issue.Number,
			EventType:   "spec.empty_output",
			Stage:       "spec",
			Message:     "agent produced empty spec output",
		})
		return fmt.Errorf("spec: agent produced empty output")
	}

	log.Info("scheduler: spec completed",
		zap.String("summary", spec.Summary),
		zap.Strings("files", spec.FilesToChange))

	specJSON := storage.SpecJSON(spec)
	_ = s.db.UpdateIssueSpec(ctx, issue.Number, specJSON)
	s.transitionIssueStage(ctx, issue.Number, models.StageSpecDone)
	_ = s.db.UpdateIssueLastCommentID(ctx, issue.Number, maxCommentID(issue.Comments))

	specComment := formatSpecComment(spec, issue.Number)
	if err := s.ghClient.PostIssueComment(ctx, s.RepoName, issue.Number, specComment); err != nil {
		log.Warn("scheduler: failed to post spec comment", zap.Error(err))
	}

	if err := s.ghClient.AddIssueLabel(ctx, s.RepoName, issue.Number, models.TRIAGED_LABEL); err != nil {
		log.Warn("scheduler: failed to add triaged label", zap.Error(err))
	}

	s.db.CompleteRun(ctx, runID, spec.RawOutput, "", "")
	s.db.AppendEvent(ctx, storage.EventRecord{
		IssueNumber: issue.Number,
		EventType:   "spec.completed",
		Stage:       "spec",
		Message:     spec.Summary,
	})

	return nil
}

func (s *Scheduler) handleSlashCommands(
	ctx context.Context,
	issue models.IssueDetail,
	commands []slash.Command,
	unprocessed []storage.CommentRecord,
) error {
	log := s.log.With(logger.WithIssue(issue.Number))

	for _, cmd := range commands {
		alreadyProcessed := true
		for _, u := range unprocessed {
			if u.ID == cmd.CommentID {
				alreadyProcessed = false
				break
			}
		}
		if alreadyProcessed {
			continue
		}

		log.Info("scheduler: processing slash command",
			zap.String("command", string(cmd.Type)),
			zap.String("author", cmd.Author))

		switch cmd.Type {
		case slash.CmdApprove:
			s.db.AppendEvent(ctx, storage.EventRecord{
				IssueNumber: issue.Number,
				EventType:   "approve.received",
				Message:     fmt.Sprintf("approved by @%s", cmd.Author),
			})
			s.transitionIssueStage(ctx, issue.Number, models.StageApproved)
			s.ghClient.PostIssueComment(ctx, s.RepoName, issue.Number,
				fmt.Sprintf("Plan approved by @%s. Starting implementation...", cmd.Author))
			s.cleanupSpecSession(issue.Number)

		case slash.CmdRevise:
			s.db.AppendEvent(ctx, storage.EventRecord{
				IssueNumber: issue.Number,
				EventType:   "revise.received",
				Message:     fmt.Sprintf("revision requested by @%s: %s", cmd.Author, cmd.Feedback),
			})
			s.ghClient.PostIssueComment(ctx, s.RepoName, issue.Number,
				fmt.Sprintf("Revising plan based on feedback from @%s...", cmd.Author))
			if err := s.runSpecStage(ctx, issue, cmd.Feedback); err != nil {
				return fmt.Errorf("revise spec: %w", err)
			}

		case slash.CmdAbort:
			s.db.AppendEvent(ctx, storage.EventRecord{
				IssueNumber: issue.Number,
				EventType:   "abort.received",
				Message:     fmt.Sprintf("aborted by @%s", cmd.Author),
			})
			s.transitionIssueStage(ctx, issue.Number, models.StageFailed)
			s.ghClient.PostIssueComment(ctx, s.RepoName, issue.Number,
				fmt.Sprintf("Aborted by @%s. No further action will be taken.", cmd.Author))
			s.cleanupIssueSessions(issue.Number)

		case slash.CmdStatus:
			cached, _ := s.db.GetIssue(ctx, issue.Number)
			stage := "unknown"
			if cached != nil {
				stage = cached.Stage
			}
			s.ghClient.PostIssueComment(ctx, s.RepoName, issue.Number,
				fmt.Sprintf("Current status: **%s**", stage))

		case slash.CmdRetry:
			s.db.AppendEvent(ctx, storage.EventRecord{
				IssueNumber: issue.Number,
				EventType:   "retry.received",
				Message:     fmt.Sprintf("retry requested by @%s", cmd.Author),
			})
			s.ghClient.PostIssueComment(ctx, s.RepoName, issue.Number,
				fmt.Sprintf("Retrying spec analysis as requested by @%s...", cmd.Author))
			// Clear the failed state and re-run from scratch
			s.transitionIssueStage(ctx, issue.Number, models.StageNew)
			_ = s.db.UpdateIssueSpec(ctx, issue.Number, "")
			if err := s.runSpecStage(ctx, issue, ""); err != nil {
				return fmt.Errorf("retry spec: %w", err)
			}
		}

		_ = s.db.MarkCommentProcessed(ctx, cmd.CommentID)
	}

	_ = s.db.UpdateIssueLastCommentID(ctx, issue.Number, maxCommentID(issue.Comments))
	return nil
}

func (s *Scheduler) processApprovedIssues(ctx context.Context) {
	approved, err := s.db.GetIssuesByStage(ctx, models.StageApproved)
	if err != nil {
		s.log.Error("scheduler: get approved issues failed", zap.Error(err))
		return
	}
	if len(approved) == 0 {
		return
	}

	s.log.Info("scheduler: processing approved issues",
		zap.Int("count", len(approved)),
		zap.String("stage", models.StageApproved))

	for _, rec := range approved {
		if err := ctx.Err(); err != nil {
			return
		}
		if err := s.runWorkStage(ctx, rec); err != nil {
			s.log.Error("scheduler: work stage failed",
				logger.WithIssue(rec.Number), zap.Error(err))
		}
	}
}

func (s *Scheduler) runWorkStage(ctx context.Context, rec storage.IssueRecord) error {
	log := s.log.With(logger.WithIssue(rec.Number), logger.WithStage("work"))

	spec := storage.ParseSpecJSON(rec.SpecJSON)
	if spec == nil {
		return fmt.Errorf("no spec found for issue %d", rec.Number)
	}

	runID, _ := s.db.StartRun(ctx, rec.Number, "work")
	log.Info("scheduler: starting WORK stage")
	s.db.AppendEvent(ctx, storage.EventRecord{
		IssueNumber: rec.Number,
		EventType:   "work.started",
		Stage:       "work",
	})

	baseBranch, err := s.gitClient.GetDefaultBranch(ctx, s.RepoPath)
	if err != nil {
		baseBranch = "main"
	}

	branchName := s.wtMgr.BranchNameForIssue(rec.Number, slugify(spec.Summary))

	wtPath, err := s.wtMgr.CreateWithBranch(ctx, s.RepoPath, branchName, baseBranch)
	if err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("create worktree: %w", err)
	}
	defer func() {
		if err := s.wtMgr.Remove(ctx, s.RepoPath, wtPath); err != nil {
			log.Warn("scheduler: failed to cleanup worktree", zap.String("path", wtPath), zap.Error(err))
		}
	}()

	work := &models.WorkResult{}
	hasExistingCommits := false
	if branchExists, _ := s.gitClient.BranchExists(ctx, s.RepoPath, branchName); branchExists {
		count, err := s.gitClient.CountCommits(ctx, wtPath, baseBranch, branchName)
		if err == nil && count > 0 {
			hasExistingCommits = true
			log.Info("scheduler: branch has existing commits, resuming from push",
				zap.Int("commit_count", count),
				zap.String("branch", branchName))
		}
	}

	if hasExistingCommits {
		// Skip the agent implementation — commits are already on the branch.
		// The agent may have committed in a previous run that failed at push/PR creation.
	} else {
		work, err = s.pl.ImplementFix(ctx, spec, wtPath, "", s.workSessionID(rec.Number))
		if err != nil {
			s.db.FailRun(ctx, runID, err.Error())
			s.db.AppendEvent(ctx, storage.EventRecord{
				IssueNumber: rec.Number,
				EventType:   "work.failed",
				Stage:       "work",
				Message:     err.Error(),
			})
			s.transitionIssueStage(ctx, rec.Number, models.StageFailed)
			return fmt.Errorf("implement: %w", err)
		}
	}

	hasChanges, _ := s.gitClient.HasUncommittedChanges(ctx, wtPath)
	if hasChanges {
		commitMsg := fmt.Sprintf("fix: %s (closes #%d)", spec.Summary, rec.Number)
		if err := s.gitClient.CommitAll(ctx, wtPath, commitMsg); err != nil {
			s.db.FailRun(ctx, runID, err.Error())
			return fmt.Errorf("auto-commit: %w", err)
		}
		log.Info("scheduler: auto-committed uncommitted changes")
	}

	if err := s.gitClient.Push(ctx, wtPath, "origin", branchName); err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("push: %w", err)
	}
	log.Info("scheduler: pushed branch", zap.String("branch", branchName))

	prTitle := fmt.Sprintf("Fix: %s", spec.Summary)
	prBody := formatPRBody(spec, rec.Number, work)
	prNumber, err := s.ghClient.CreatePR(ctx, s.RepoName, prTitle, prBody, branchName, baseBranch)
	if err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("create pr: %w", err)
	}

	log.Info("scheduler: PR created", logger.WithPR(prNumber))

	_ = s.db.UpdateIssuePR(ctx, rec.Number, prNumber, branchName)
	s.transitionIssueStage(ctx, rec.Number, models.StageWorkDone)

	prInfo, _ := s.ghClient.GetPR(ctx, s.RepoName, prNumber)
	if prInfo != nil {
		_ = s.db.UpsertPR(ctx, storage.PRRecord{
			Number:      prInfo.Number,
			IssueNumber: rec.Number,
			Title:       prInfo.Title,
			HeadSHA:     prInfo.HeadSHA,
			BaseSHA:     prInfo.BaseSHA,
			State:       prInfo.State,
			URL:         prInfo.URL,
			UpdatedAt:   prInfo.UpdatedAt,
		})
	}

	s.db.CompleteRun(ctx, runID, work.RawOutput, "", "")
	s.db.AppendEvent(ctx, storage.EventRecord{
		IssueNumber: rec.Number,
		PRNumber:    prNumber,
		EventType:   "work.completed",
		Stage:       "work",
		Message:     fmt.Sprintf("PR #%d created", prNumber),
	})

	s.ghClient.PostIssueComment(ctx, s.RepoName, rec.Number,
		fmt.Sprintf("Implementation complete. PR #%d has been created for review.", prNumber))

	return nil
}

func (s *Scheduler) processWorkDoneIssues(ctx context.Context) {
	workDone, err := s.db.GetIssuesByStage(ctx, models.StageWorkDone)
	if err != nil {
		s.log.Error("scheduler: get work-done issues failed", zap.Error(err))
		return
	}
	if len(workDone) == 0 {
		return
	}

	s.log.Info("scheduler: processing work-done issues",
		zap.Int("count", len(workDone)),
		zap.String("stage", models.StageWorkDone))

	for _, rec := range workDone {
		if err := ctx.Err(); err != nil {
			return
		}
		if rec.PRNumber == 0 {
			continue
		}
		if err := s.runReviewStage(ctx, rec); err != nil {
			s.log.Error("scheduler: review stage failed",
				logger.WithIssue(rec.Number), zap.Error(err))
		}
	}
}

func (s *Scheduler) processReviewDoneIssues(ctx context.Context) {
	reviewDone, err := s.db.GetIssuesByStage(ctx, models.StageReviewDone)
	if err != nil {
		s.log.Error("scheduler: get review-done issues failed", zap.Error(err))
		return
	}
	merged, err := s.db.GetIssuesByStage(ctx, models.StageMerged)
	if err != nil {
		s.log.Error("scheduler: get merged issues failed", zap.Error(err))
		return
	}
	reviewDone = append(reviewDone, merged...)
	if len(reviewDone) == 0 {
		return
	}

	s.log.Info("scheduler: checking review-done/merged PRs for cleanup",
		zap.Int("count", len(reviewDone)),
		zap.String("stage", models.StageReviewDone+"/"+models.StageMerged))

	for _, rec := range reviewDone {
		if err := ctx.Err(); err != nil {
			return
		}
		if rec.PRNumber == 0 {
			continue
		}
		pr, err := s.ghClient.GetPR(ctx, s.RepoName, rec.PRNumber)
		if err != nil {
			s.log.Warn("scheduler: failed to check PR state",
				logger.WithIssue(rec.Number), logger.WithPR(rec.PRNumber), zap.Error(err))
			continue
		}
		if pr == nil || pr.State == "open" {
			continue
		}

		s.log.Info("scheduler: PR closed/merged, cleaning up",
			logger.WithIssue(rec.Number), logger.WithPR(rec.PRNumber),
			zap.String("pr_state", pr.State))
		s.cleanupIssueSessions(rec.Number)
		s.transitionIssueStage(ctx, rec.Number, models.StageMerged)
		s.db.AppendEvent(ctx, storage.EventRecord{
			IssueNumber: rec.Number,
			PRNumber:    rec.PRNumber,
			EventType:   "cleanup.sessions_deleted",
			Stage:       "merged",
			Message:     fmt.Sprintf("PR %s, sessions cleaned up", pr.State),
		})
	}
}

func (s *Scheduler) runReviewStage(ctx context.Context, rec storage.IssueRecord) error {
	log := s.log.With(logger.WithIssue(rec.Number), logger.WithStage("review"),
		logger.WithPR(rec.PRNumber))

	prInfo, err := s.ghClient.GetPR(ctx, s.RepoName, rec.PRNumber)
	if err != nil {
		return fmt.Errorf("get pr: %w", err)
	}
	if prInfo == nil || prInfo.State != "open" {
		s.transitionIssueStage(ctx, rec.Number, models.StageMerged)
		return nil
	}

	iterations := rec.ReviewIterations
	if iterations >= models.MaxReviewIterations {
		log.Info("scheduler: max review iterations reached, requesting human review")
		s.ghClient.PostPRComment(ctx, s.RepoName, rec.PRNumber,
			fmt.Sprintf("Maximum review iterations (%d) reached. Please review manually.", models.MaxReviewIterations))
		s.transitionIssueStage(ctx, rec.Number, models.StageFailed)
		s.db.AppendEvent(ctx, storage.EventRecord{
			IssueNumber: rec.Number,
			PRNumber:    rec.PRNumber,
			EventType:   "review.max_iterations",
			Stage:       "review",
			Message:     "max iterations reached",
		})
		return nil
	}

	runID, _ := s.db.StartRun(ctx, rec.Number, "review")
	log.Info("scheduler: starting REVIEW stage",
		zap.Int("iteration", iterations+1))
	s.db.AppendEvent(ctx, storage.EventRecord{
		IssueNumber: rec.Number,
		PRNumber:    rec.PRNumber,
		EventType:   "review.started",
		Stage:       "review",
	})

	diff, err := s.ghClient.GetPRDiff(ctx, s.RepoName, rec.PRNumber)
	if err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("get pr diff: %w", err)
	}

	reviewWtName := s.wtMgr.ReviewWorktreeName(rec.Number, rec.PRNumber)
	reviewWtPath, err := s.wtMgr.CreateAtCommit(ctx, s.RepoPath, reviewWtName, prInfo.HeadSHA)
	if err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("create review worktree: %w", err)
	}
	defer func() {
		if err := s.wtMgr.Remove(ctx, s.RepoPath, reviewWtPath); err != nil {
			log.Warn("scheduler: failed to cleanup review worktree",
				zap.String("path", reviewWtPath), zap.Error(err))
		}
	}()

	review, err := s.pl.ReviewPR(ctx, diff, reviewWtPath, s.reviewSessionID(rec.Number))
	if err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		s.db.AppendEvent(ctx, storage.EventRecord{
			IssueNumber: rec.Number,
			PRNumber:    rec.PRNumber,
			EventType:   "review.failed",
			Stage:       "review",
			Message:     err.Error(),
		})
		return fmt.Errorf("review: %w", err)
	}

	log.Info("scheduler: review completed",
		zap.Bool("approved", review.Approved),
		zap.Int("issues", len(review.Issues)))

	if review.Approved {
		s.db.CompleteRun(ctx, runID, review.RawOutput, "", "")
		s.transitionIssueStage(ctx, rec.Number, models.StageReviewDone)
		s.db.AppendEvent(ctx, storage.EventRecord{
			IssueNumber: rec.Number,
			PRNumber:    rec.PRNumber,
			EventType:   "review.approved",
			Stage:       "review",
		})
		s.ghClient.PostPRComment(ctx, s.RepoName, rec.PRNumber,
			"Self-review passed. The PR is ready for human review and merge.")
		return nil
	}

	s.db.CompleteRun(ctx, runID, review.RawOutput, "", "")

	reviewComment := formatReviewComment(review)
	s.ghClient.PostPRComment(ctx, s.RepoName, rec.PRNumber, reviewComment)

	s.db.AppendEvent(ctx, storage.EventRecord{
		IssueNumber: rec.Number,
		PRNumber:    rec.PRNumber,
		EventType:   "review.changes_requested",
		Stage:       "review",
		Message:     fmt.Sprintf("%d issues found", len(review.Issues)),
	})

	if err := s.runFixStage(ctx, rec, review); err != nil {
		return fmt.Errorf("fix: %w", err)
	}

	return nil
}

func (s *Scheduler) runFixStage(ctx context.Context, rec storage.IssueRecord, review *models.ReviewResult) error {
	log := s.log.With(logger.WithIssue(rec.Number), logger.WithStage("fix"),
		logger.WithPR(rec.PRNumber))

	runID, _ := s.db.StartRun(ctx, rec.Number, "fix")
	log.Info("scheduler: starting FIX stage")
	s.db.AppendEvent(ctx, storage.EventRecord{
		IssueNumber: rec.Number,
		PRNumber:    rec.PRNumber,
		EventType:   "fix.started",
		Stage:       "fix",
	})

	baseBranch, _ := s.gitClient.GetDefaultBranch(ctx, s.RepoPath)
	if baseBranch == "" {
		baseBranch = "main"
	}

	fixWtPath, err := s.wtMgr.CreateWithBranch(ctx, s.RepoPath,
		rec.BranchName+"_fix", baseBranch)
	if err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("create fix worktree: %w", err)
	}
	defer func() {
		if err := s.wtMgr.Remove(ctx, s.RepoPath, fixWtPath); err != nil {
			log.Warn("scheduler: failed to cleanup fix worktree",
				zap.String("path", fixWtPath), zap.Error(err))
		}
	}()

	if err := s.gitClient.FetchBranch(ctx, s.RepoPath, "origin", rec.BranchName); err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("fetch work branch: %w", err)
	}

	if err := s.gitClient.ResetHard(ctx, fixWtPath, "origin/"+rec.BranchName); err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("reset to work branch: %w", err)
	}

	fixIteration := rec.ReviewIterations + 1
	work, err := s.pl.FixReviewIssues(ctx, review, fixWtPath, s.fixSessionID(rec.Number, fixIteration))
	if err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		s.db.AppendEvent(ctx, storage.EventRecord{
			IssueNumber: rec.Number,
			PRNumber:    rec.PRNumber,
			EventType:   "fix.failed",
			Stage:       "fix",
			Message:     err.Error(),
		})
		return fmt.Errorf("fix implement: %w", err)
	}

	hasChanges, _ := s.gitClient.HasUncommittedChanges(ctx, fixWtPath)
	if hasChanges {
		if err := s.gitClient.CommitAll(ctx, fixWtPath, "fix: address review feedback"); err != nil {
			s.db.FailRun(ctx, runID, err.Error())
			return fmt.Errorf("commit fix: %w", err)
		}
	}

	if err := s.gitClient.PushRef(ctx, fixWtPath, "origin", "HEAD", rec.BranchName); err != nil {
		s.db.FailRun(ctx, runID, err.Error())
		return fmt.Errorf("push fix: %w", err)
	}

	log.Info("scheduler: fix pushed to PR branch")

	iters, _ := s.db.IncrementReviewIterations(ctx, rec.Number)
	s.transitionIssueStage(ctx, rec.Number, models.StageWorkDone)

	s.db.CompleteRun(ctx, runID, work.RawOutput, "", "")
	s.db.AppendEvent(ctx, storage.EventRecord{
		IssueNumber: rec.Number,
		PRNumber:    rec.PRNumber,
		EventType:   "fix.completed",
		Stage:       "fix",
		Message:     fmt.Sprintf("fix pushed, iteration %d", iters),
	})

	return nil
}

// ─── Formatting helpers ──────────────────────────────────────────────

func formatSpecComment(spec *models.SpecResult, issueNumber int64) string {
	var b strings.Builder
	b.WriteString("## Sloper Spec Analysis\n\n")
	b.WriteString(fmt.Sprintf("**Summary:** %s\n\n", spec.Summary))
	b.WriteString("### Files to Change\n")
	if len(spec.FilesToChange) > 0 {
		for _, f := range spec.FilesToChange {
			b.WriteString(fmt.Sprintf("- `%s`\n", f))
		}
	} else {
		b.WriteString("_To be determined during implementation_\n")
	}
	b.WriteString("\n### Implementation Plan\n\n")
	b.WriteString(spec.ImplementationPlan)
	b.WriteString("\n\n---\n")
	b.WriteString("_To approve this plan, comment `/sloper approve`\n")
	b.WriteString("To request changes, comment `/sloper revise: <your feedback>`\n")
	b.WriteString("To abort, comment `/sloper abort`_")
	return b.String()
}

func formatPRBody(spec *models.SpecResult, issueNumber int64, work *models.WorkResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Summary\n\n%s\n\n", spec.Summary))
	b.WriteString(fmt.Sprintf("Closes #%d\n\n", issueNumber))
	b.WriteString("## Plan\n\n")
	b.WriteString(spec.ImplementationPlan)
	b.WriteString("\n\n## Files Changed\n\n")
	if len(spec.FilesToChange) > 0 {
		for _, f := range spec.FilesToChange {
			b.WriteString(fmt.Sprintf("- `%s`\n", f))
		}
	}
	b.WriteString("\n---\n_Generated by sloper_")
	return b.String()
}

func formatReviewComment(review *models.ReviewResult) string {
	var b strings.Builder
	b.WriteString("## Sloper Self-Review\n\n")
	if review.Approved {
		b.WriteString("✅ **Approved** — no issues found.\n")
	} else {
		b.WriteString("❌ **Changes requested**\n\n")
		if len(review.Issues) > 0 {
			b.WriteString("### Issues\n")
			for i, issue := range review.Issues {
				b.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
			}
		}
		if len(review.Suggestions) > 0 {
			b.WriteString("\n### Suggestions\n")
			for i, sug := range review.Suggestions {
				b.WriteString(fmt.Sprintf("%d. %s\n", i+1, sug))
			}
		}
	}
	b.WriteString("\n_Applying fixes automatically..._")
	return b.String()
}

func maxCommentID(comments []models.CommentInfo) int64 {
	var maxID int64
	for _, c := range comments {
		if c.ID > maxID {
			maxID = c.ID
		}
	}
	return maxID
}

func collectFeedbackFromCommentsExcludingBot(comments []models.CommentInfo, lastSeenID int64, botUser string) string {
	var parts []string
	for _, c := range comments {
		if c.ID <= lastSeenID {
			continue
		}
		if botUser != "" && c.Author == botUser {
			continue
		}
		parts = append(parts, fmt.Sprintf("[%s] %s: %s", c.CreatedAt, c.Author, c.Body))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n---\n\n")
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "'", "")
	var result []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}
	slug := strings.Trim(string(result), "-")
	if len(slug) > 30 {
		slug = slug[:30]
	}
	if slug == "" {
		slug = "fix"
	}
	return slug
}

func (s *Scheduler) specSessionID(issueNumber int64) string {
	return fmt.Sprintf("sloper-issue-%d-spec", issueNumber)
}

func (s *Scheduler) workSessionID(issueNumber int64) string {
	return fmt.Sprintf("sloper-issue-%d-work", issueNumber)
}

func (s *Scheduler) reviewSessionID(issueNumber int64) string {
	return fmt.Sprintf("sloper-issue-%d-review", issueNumber)
}

func (s *Scheduler) fixSessionID(issueNumber int64, iteration int) string {
	return fmt.Sprintf("sloper-issue-%d-fix-%d", issueNumber, iteration)
}

func (s *Scheduler) issueSessionPattern(issueNumber int64) string {
	return fmt.Sprintf("sloper-issue-%d", issueNumber)
}

func (s *Scheduler) cleanupIssueSessions(issueNumber int64) {
	_ = session.CleanupSessions(s.sessionDir, s.issueSessionPattern(issueNumber))
}

func (s *Scheduler) cleanupSpecSession(issueNumber int64) {
	_ = session.CleanupSessions(s.sessionDir, s.specSessionID(issueNumber))
}

func (s *Scheduler) transitionIssueStage(ctx context.Context, issueNumber int64, newStage string) {
	old, _ := s.db.GetIssue(ctx, issueNumber)
	oldStage := ""
	if old != nil {
		oldStage = old.Stage
	}

	_ = s.db.UpdateIssueStage(ctx, issueNumber, newStage)

	s.log.Info("scheduler: stage transition",
		logger.WithIssue(issueNumber),
		zap.String("old_stage", oldStage),
		zap.String("new_stage", newStage))
}

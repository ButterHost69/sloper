package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ButterHost69/sloper/internal/agent"
	"github.com/ButterHost69/sloper/internal/models"
)

type Pipeline struct {
	ag *agent.AgentGateway
}

func New(ag *agent.AgentGateway) *Pipeline {
	return &Pipeline{ag: ag}
}

// ─── Stage operations ───────────────────────────────────────────────

// SpecIssue runs the SPEC stage: analyzes an issue and produces a plan.
// feedback is optional — pass user feedback from /sloper revise when re-triaging.
func (p *Pipeline) SpecIssue(ctx context.Context, issue models.IssueDetail, feedback, sessionID string) (*models.SpecResult, error) {
	out, err := p.ag.RunStageWithCWD(ctx, buildSpecPrompt(issue, feedback), "", sessionID)
	if err != nil {
		return nil, fmt.Errorf("spec: %w", err)
	}
	return parseSpecResult(parseText(out)), nil
}

// Processes an IssueComment runs the SPEC stage: analyzes an issue comment and produces a plan.
func (p *Pipeline) ProcessIssueComment(ctx context.Context, issue models.IssueDetail, unprocessedComment string, sessionID string) (*models.ProcessCommentResult, error) {
	out, err := p.ag.RunStageWithCWD(ctx, buildProcessCommentPrompt(issue, unprocessedComment), "", sessionID)
	if err != nil {
		return nil, fmt.Errorf("process comment: %w", err)
	}
	return parseCommentSpecResult(parseText(out)), nil
}

// ImplementFix runs the WORK stage: implements the spec plan.
// worktreePath is the CWD for the agent (a git worktree).
// feedback is optional — pass user feedback when re-working based on comments.
func (p *Pipeline) ImplementFix(ctx context.Context, spec *models.SpecResult, worktreePath, feedback, sessionID string) (*models.WorkResult, error) {
	out, err := p.ag.RunStageWithCWD(ctx, buildWorkPrompt(spec, worktreePath, feedback), worktreePath, sessionID)
	if err != nil {
		return nil, fmt.Errorf("work: %w", err)
	}
	return parseWorkResult(out.Text), nil
}

// ReviewPR runs the REVIEW stage: reviews a diff.
// worktreePath is the CWD for the review agent (a separate worktree at the PR head).
func (p *Pipeline) ReviewPR(ctx context.Context, diff, worktreePath, sessionID string) (*models.ReviewResult, error) {
	out, err := p.ag.RunStageWithCWD(ctx, buildReviewPrompt(diff, worktreePath), worktreePath, sessionID)
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	return parseReviewResult(parseText(out)), nil
}

// FixReviewIssues runs the FIX stage: addresses review feedback.
// worktreePath is the CWD for the fix agent (the WORK worktree).
func (p *Pipeline) FixReviewIssues(ctx context.Context, review *models.ReviewResult, worktreePath, sessionID string) (*models.WorkResult, error) {
	out, err := p.ag.RunStageWithCWD(ctx, buildFixPrompt(review, worktreePath), worktreePath, sessionID)
	if err != nil {
		return nil, fmt.Errorf("fix: %w", err)
	}
	return parseWorkResult(out.Text), nil
}

// ─── Response parsers ───────────────────────────────────────────────

// parseText picks the text the JSON parsers should read: the agent's final
// answer (FinalText) when available, falling back to the full streamed text.
func parseText(out *agent.StageOutput) string {
	if strings.TrimSpace(out.FinalText) != "" {
		return out.FinalText
	}
	return out.Text
}

// extractJSONBlock scans every fenced code block (```json first, then bare ```)
// and unmarshals the first one that parses into dst. A closing fence is a ```
// that sits on its own line, so ``` appearing inside the JSON content (e.g. in
// markdown plans) does not truncate the block. Falls back to parsing the whole
// text as JSON.
func extractJSONBlock(text string, dst any) {
	for _, fence := range []string{"```json", "```"} {
		if extractFencedBlocks(text, fence, dst) {
			return
		}
	}
	_ = json.Unmarshal([]byte(text), dst)
}

func extractFencedBlocks(text, fence string, dst any) bool {
	rest := text
	for {
		start := strings.Index(rest, fence)
		if start == -1 {
			return false
		}
		contentStart := start + len(fence)
		if nl := strings.Index(rest[contentStart:], "\n"); nl != -1 {
			contentStart += nl + 1
		}
		search := contentStart
		foundClose := false
		for {
			end := strings.Index(rest[search:], "```")
			if end == -1 {
				break
			}
			closePos := search + end
			lineStart := strings.LastIndex(rest[:closePos], "\n") + 1
			if strings.TrimSpace(rest[lineStart:closePos]) == "" {
				block := strings.TrimSpace(rest[contentStart:closePos])
				if json.Unmarshal([]byte(block), dst) == nil {
					return true
				}
				rest = rest[closePos+3:]
				foundClose = true
				break
			}
			search = closePos + 3
		}
		if !foundClose {
			return false
		}
	}
}

func parseCommentSpecResult(text string) *models.ProcessCommentResult {
	r := &models.ProcessCommentResult{RawOutput: text}

	var parsedPropose models.ProcessCommentProposeParse
	var parsedGrill models.ProcessCommentGrillParse

	extractJSONBlock(text, &parsedPropose)
	if parsedPropose.FilesToChange != nil || parsedPropose.Summary != "" {
		r.Type = int(models.CommentPropose)
		r.Summary = parsedPropose.Summary
		r.FilesToChange = parsedPropose.FilesToChange
		return r
	}

	extractJSONBlock(text, &parsedGrill)
	if parsedGrill.GrillMe {
		r.Type = int(models.CommentGrillme)
		r.Questions = parsedGrill.Questions
		return r
	}
	r.Type = int(models.CommentUnknown)
	return r
}

func parseSpecResult(text string) *models.SpecResult {
	r := &models.SpecResult{RawOutput: text}
	var parsed models.SpecResult
	extractJSONBlock(text, &parsed)
	if parsed.Summary != "" {
		r.Summary = parsed.Summary
	}
	if len(parsed.FilesToChange) > 0 {
		r.FilesToChange = parsed.FilesToChange
	}
	if parsed.ImplementationPlan != "" {
		r.ImplementationPlan = parsed.ImplementationPlan
	}
	if r.Summary == "" {
		r.Summary = firstLine(text, 120)
	}
	return r
}

func parseWorkResult(text string) *models.WorkResult {
	return &models.WorkResult{RawOutput: text}
}

func parseReviewResult(text string) *models.ReviewResult {
	r := &models.ReviewResult{RawOutput: text}
	var parsed models.ReviewResult
	extractJSONBlock(text, &parsed)
	r.Approved = parsed.Approved
	r.Issues = parsed.Issues
	r.Suggestions = parsed.Suggestions
	return r
}

func firstLine(s string, maxLen int) string {
	s = strings.TrimLeft(s, " \t\r\n")
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[:idx]
	}
	if len(s) > maxLen {
		s = s[:maxLen] + "…"
	}
	return strings.TrimSpace(s)
}

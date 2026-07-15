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
	return parseSpecResult(out.Text), nil
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
	return parseReviewResult(out.Text), nil
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

func extractJSONBlock(text string, dst any) {
	start := strings.Index(text, "```json")
	if start == -1 {
		start = strings.Index(text, "```")
	}
	if start != -1 {
		start = strings.Index(text[start:], "\n") + start + 1
		end := strings.Index(text[start:], "```")
		if end != -1 {
			block := strings.TrimSpace(text[start : start+end])
			if json.Unmarshal([]byte(block), dst) == nil {
				return
			}
		}
	}
	_ = json.Unmarshal([]byte(text), dst)
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
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[:idx]
	}
	if len(s) > maxLen {
		s = s[:maxLen] + "…"
	}
	return strings.TrimSpace(s)
}

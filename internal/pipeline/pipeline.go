package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ButterHost69/sloper/internal/agent"
	"github.com/ButterHost69/sloper/internal/models"
)

// Pipeline runs the SPEC → WORK → REVIEW → FIX → MERGE chain.
//
// It composes the low-level AgentGateway (RPC transport) with stage-level
// prompt templates and response parsers so the scheduler can treat each
// stage as a single function call.
type Pipeline struct {
	ag *agent.AgentGateway
}

// New creates a Pipeline backed by the given agent gateway.
func New(ag *agent.AgentGateway) *Pipeline {
	return &Pipeline{ag: ag}
}

// ─── Stage operations ───────────────────────────────────────────────

// SpecIssue runs the SPEC stage: analyzes an issue and produces a plan.
func (p *Pipeline) SpecIssue(ctx context.Context, issue models.IssueDetail) (*models.SpecResult, error) {
	out, err := p.ag.RunStage(ctx, buildSpecPrompt(issue))
	if err != nil {
		return nil, fmt.Errorf("spec: %w", err)
	}
	return parseSpecResult(out.Text), nil
}

// ImplementFix runs the WORK stage: implements the spec plan.
func (p *Pipeline) ImplementFix(ctx context.Context, spec *models.SpecResult) (*models.WorkResult, error) {
	out, err := p.ag.RunStage(ctx, buildWorkPrompt(spec))
	if err != nil {
		return nil, fmt.Errorf("work: %w", err)
	}
	return parseWorkResult(out.Text), nil
}

// ReviewPR runs the REVIEW stage: reviews a diff.
func (p *Pipeline) ReviewPR(ctx context.Context, diff string) (*models.ReviewResult, error) {
	out, err := p.ag.RunStage(ctx, buildReviewPrompt(diff))
	if err != nil {
		return nil, fmt.Errorf("review: %w", err)
	}
	return parseReviewResult(out.Text), nil
}

// FixReviewIssues runs the FIX stage: addresses review feedback.
func (p *Pipeline) FixReviewIssues(ctx context.Context, review *models.ReviewResult) (*models.WorkResult, error) {
	out, err := p.ag.RunStage(ctx, buildFixPrompt(review))
	if err != nil {
		return nil, fmt.Errorf("fix: %w", err)
	}
	return parseWorkResult(out.Text), nil
}

// ─── Response parsers ───────────────────────────────────────────────

// extractJSONBlock finds the first ```json … ``` fenced block and
// unmarshals it into dst.  Falls back to the raw text if no JSON found.
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
	// Fallback: try raw text as JSON.
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

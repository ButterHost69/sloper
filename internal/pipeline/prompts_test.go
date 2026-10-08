package pipeline

import (
	"strings"
	"testing"

	"github.com/ButterHost69/sloper/internal/models"
)

func specTestIssue() models.IssueDetail {
	return models.IssueDetail{
		Number: 5,
		Title:  "modify /sloper spec to complete rewrite the entire spec with the comment / conversation",
		Body:   "The spec should be rewritten from the whole conversation.",
		Labels: []string{"urgent"},
		Comments: []models.CommentInfo{
			{ID: 1, Author: "ButterHost69", CreatedAt: "2026-08-04T21:39:37Z", Body: "Please make it a full rewrite."},
			{ID: 2, Author: "sloper-bot", CreatedAt: "2026-08-04T21:40:00Z", Body: "Proposed: change internal/scheduler/scheduler.go"},
		},
	}
}

var previousSpec = &models.SpecResult{
	Summary:            "Spec v1 summary",
	FilesToChange:      []string{"internal/scheduler/scheduler.go", "internal/pipeline/prompts.go"},
	ImplementationPlan: "Spec v1 implementation plan",
}

func TestBuildSpecPromptRewritesPreviousSpec(t *testing.T) {
	prompt := buildSpecPrompt(specTestIssue(), SpecOptions{Previous: previousSpec})

	for _, want := range []string{
		"## Current Spec (being replaced)",
		"Spec v1 summary",
		"Spec v1 implementation plan",
		"internal/scheduler/scheduler.go",
		"internal/pipeline/prompts.go",
		"Please make it a full rewrite.",
		"sloper-bot",
		"The result is always a COMPLETE, self-contained spec",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt is missing %q", want)
		}
	}
	if strings.Contains(prompt, "Targeted Feedback") {
		t.Error("prompt mentions targeted feedback when none was given")
	}
	if !strings.Contains(prompt, "```json\n{") {
		t.Error("example JSON fence must start its block on a new line")
	}
	if strings.Contains(prompt, "%!") {
		t.Errorf("prompt contains a formatting error: %s", prompt)
	}
}

func TestBuildSpecPromptWithoutPreviousSpec(t *testing.T) {
	prompt := buildSpecPrompt(specTestIssue(), SpecOptions{})

	if strings.Contains(prompt, "## Current Spec") {
		t.Error("prompt claims a spec is being replaced when the issue has none")
	}
	if !strings.Contains(prompt, "Please make it a full rewrite.") {
		t.Error("prompt is missing the issue conversation")
	}
	if !strings.Contains(prompt, "The result is always a COMPLETE, self-contained spec") {
		t.Error("prompt does not require a complete spec")
	}
}

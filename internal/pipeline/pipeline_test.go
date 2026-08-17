package pipeline

import (
	"testing"

	"github.com/ButterHost69/sloper/internal/agent"
)

func TestExtractJSONBlockSkipsStrayFenceAndPicksValidBlock(t *testing.T) {
	text := "```json\n{\"partial\": \"not valid\"\n```\nmore prose\n\n```json\n{\"summary\":\"ok\",\"files_to_change\":[\"a\"],\"implementation_plan\":\"p\"}\n```\n"
	var parsed struct {
		Summary string `json:"summary"`
	}
	extractJSONBlock(text, &parsed)
	if parsed.Summary != "ok" {
		t.Fatalf("summary = %q, want ok", parsed.Summary)
	}
}

func TestExtractJSONBlockIgnoresFenceInsideJSONString(t *testing.T) {
	block := "{\"summary\":\"s\",\"files_to_change\":[\"f\"],\"implementation_plan\":\"line1\\n```\\ncode\\n```\\nline2\"}"
	text := "```json\n" + block + "\n```"
	var parsed struct {
		Summary string `json:"summary"`
		Plan    string `json:"implementation_plan"`
	}
	extractJSONBlock(text, &parsed)
	if parsed.Summary != "s" {
		t.Fatalf("summary = %q, want s", parsed.Summary)
	}
	if parsed.Plan == "" {
		t.Fatal("plan not parsed; closing fence scan was confused by ``` inside the JSON string")
	}
}

func TestFirstLineSkipsLeadingNewlines(t *testing.T) {
	got := firstLine("\n\n\n\n\n\nLet me check the backend tracking constraints...\nmore", 120)
	want := "Let me check the backend tracking constraints..."
	if got != want {
		t.Fatalf("firstLine = %q, want %q", got, want)
	}
}

func TestParseSpecResultFromIssue42Shape(t *testing.T) {
	text := "\n\n\n\n\n\nLet me check the backend tracking constraints to know whether new event types are possible.\nI have a complete picture of the codebase now. Key findings:\n\n- **100% of styling** lives in a single file\n\nHere is the specification:\n\n```json\n{\n  \"summary\": \"Add dark mode to the React client\",\n  \"files_to_change\": [\"client/src/index.css\", \"client/src/App.tsx\"],\n  \"implementation_plan\": \"## Overview\\n\\nImplement a theme system.\"\n}\n```\n"
	spec := parseSpecResult(text)
	if spec.Summary != "Add dark mode to the React client" {
		t.Fatalf("summary = %q", spec.Summary)
	}
	if len(spec.FilesToChange) != 2 {
		t.Fatalf("files = %v, want 2", spec.FilesToChange)
	}
	if spec.ImplementationPlan == "" {
		t.Fatal("plan is empty")
	}
}

func TestParseTextPrefersFinalText(t *testing.T) {
	out := &agent.StageOutput{
		Text:      "\n\n\nnoisy streamed transcript",
		FinalText: "clean final answer",
	}
	if got := parseText(out); got != "clean final answer" {
		t.Fatalf("parseText = %q, want the final message", got)
	}

	out = &agent.StageOutput{Text: "only streamed text"}
	if got := parseText(out); got != "only streamed text" {
		t.Fatalf("parseText fallback = %q", got)
	}
}

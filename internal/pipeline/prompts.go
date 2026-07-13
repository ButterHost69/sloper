package pipeline

import (
	"fmt"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
)

// ─── SPEC ───────────────────────────────────────────────────────────

const specTemplate = `You are an expert software engineer triaging a GitHub issue.

## Working Directory
You are inside the repository.  Use the read, grep and find tools to
explore the codebase before writing your specification.

## Issue
Title: %s

%s

Labels: %s

Comments:
%s

## Task
1. Read relevant source files to understand the codebase.
2. Identify the root cause (for bugs) or scope of change (for features).
3. List every file that must be modified.
4. Write a detailed, step-by-step implementation plan with code-level
   specifics (function names, types, edge cases).

## Output Format
Wrap your final answer in a fenced JSON block:

` + "```json" + `
{
  "summary": "one-line summary of the issue and fix",
  "files_to_change": ["path/to/file1.go", "path/to/file2.go"],
  "implementation_plan": "detailed multi-paragraph plan"
}
` + "```" + `

IMPORTANT: Do NOT make any edits yet.  Only produce the specification.`

func buildSpecPrompt(issue models.IssueDetail) string {
	labels := "none"
	if len(issue.Labels) > 0 {
		labels = strings.Join(issue.Labels, ", ")
	}
	comments := "none"
	if len(issue.Comments) > 0 {
		parts := make([]string, 0, len(issue.Comments))
		for _, c := range issue.Comments {
			parts = append(parts, fmt.Sprintf("[%s] %s:\n%s", c.CreatedAt, c.Author, c.Body))
		}
		comments = strings.Join(parts, "\n\n---\n\n")
	}
	return fmt.Sprintf(specTemplate, issue.Title, issue.Body, labels, comments)
}

// ─── WORK ───────────────────────────────────────────────────────────

const workTemplate = `You are implementing a fix based on the following specification.

## Specification
%s

## Files to Change
%s

## Implementation Plan
%s

## Task
1. Read the current state of each affected file.
2. Implement every change described in the plan.  Use the **edit** tool
   for precise, targeted modifications.
3. After all edits, run the project's test suite with bash.
4. If tests fail, fix the failures and re-run until green.
5. When everything passes, commit your changes with a descriptive message
   that references the issue.

## Important Rules
- Do NOT open a pull request.  Only commit locally.
- Prefer many small, focused edits over one giant rewrite.
- If you're unsure about anything, read more code before editing.
- Commit message format: "fix: <summary> (closes #N)"`

func buildWorkPrompt(spec *models.SpecResult) string {
	files := "discover from the repository"
	if len(spec.FilesToChange) > 0 {
		files = strings.Join(spec.FilesToChange, "\n")
	}
	return fmt.Sprintf(workTemplate, spec.Summary, files, spec.ImplementationPlan)
}

// ─── REVIEW ─────────────────────────────────────────────────────────

const reviewTemplate = `You are reviewing a pull request for correctness and quality.

## PR Context
This PR was created to address an issue.  The diff is shown below.

## Diff
%s

## Task
1. Examine the diff for:
   - Logic bugs or incorrect behavior
   - Security vulnerabilities
   - Performance regressions
   - Missing tests or error handling
   - Style or convention violations
2. Determine whether the PR is safe to merge.
3. List specific issues and suggested fixes.

## Output Format
` + "```json" + `
{
  "approved": true or false,
  "issues": ["issue 1", "issue 2"],
  "suggestions": ["suggestion 1", "suggestion 2"]
}
` + "```" + `

If approved is false, each issue MUST be accompanied by a concrete
suggestion for how to fix it.`

func buildReviewPrompt(diff string) string { return fmt.Sprintf(reviewTemplate, diff) }

// ─── FIX ────────────────────────────────────────────────────────────

const fixTemplate = `You need to fix issues raised during code review.

## Review Feedback
Issues found:
%s

Suggestions:
%s

## Task
1. Read the affected files.
2. Apply the suggested fixes using the edit tool.
3. Run tests after each change.
4. Commit with message: "fix: address review feedback"

Do NOT create a PR.`

func buildFixPrompt(review *models.ReviewResult) string {
	return fmt.Sprintf(fixTemplate,
		bulletList(review.Issues),
		bulletList(review.Suggestions),
	)
}

func bulletList(items []string) string {
	if len(items) == 0 {
		return "  (none)"
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = fmt.Sprintf("  - %s", item)
	}
	return strings.Join(lines, "\n")
}

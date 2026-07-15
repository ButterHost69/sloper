package pipeline

import (
	"fmt"
	"strings"

	"github.com/ButterHost69/sloper/internal/models"
)

// ─── SPEC ───────────────────────────────────────────────────────────

const specTemplate = `You are an expert software engineer triaging a GitHub issue.

## Working Directory
You are inside the repository at the root.  Before writing your spec, you MUST
explore the codebase to understand the project structure and relevant code.

Use these tools:
- ls: List files and directories to understand project structure
- find: Find files by pattern (e.g. find "*.go" in src/)
- grep: Search for relevant keywords, function names, or patterns
- read: Read the contents of specific files you identify as relevant
- bash: Run any shell commands needed (e.g. cat, wc, tree, etc.)

Do NOT skip exploration.  A good spec is grounded in the actual code.

## Issue
Title: %s

%s

Labels: %s

Comments:
%s

## Previous Feedback
%s

## Task
1. Use ls and find to understand the project structure.
2. Use grep to search for code related to the issue (keywords from the title/body).
3. Read the relevant source files you discover.
4. Identify the root cause (for bugs) or scope of change (for features).
5. List every file that must be modified, with the specific changes needed.
6. Write a detailed, step-by-step implementation plan with code-level
   specifics (function names, types, edge cases, new files to create).
7. If there is previous feedback above, incorporate it into your revised plan.

## Output Format
You MUST wrap your final answer in a fenced JSON block.  Do not output
anything except the JSON block as your final message.

` + "```json" + `{
  "summary": "one-line summary of the issue and the proposed fix",
  "files_to_change": ["path/to/file1.go", "path/to/file2.go"],
  "implementation_plan": "detailed multi-paragraph plan describing exactly what changes to make in each file, including function signatures, edge cases, and test considerations"
}
` + "```" + `

IMPORTANT:
- Do NOT make any edits yet.  Only produce the specification.
- The "summary" field must be a real description, not empty.
- The "files_to_change" array must list actual file paths from the repo.
- The "implementation_plan" must be detailed and reference real code.`

func buildSpecPrompt(issue models.IssueDetail, feedback string) string {
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
	if feedback == "" {
		feedback = "(none)"
	}
	return fmt.Sprintf(specTemplate, issue.Title, issue.Body, labels, comments, feedback)
}

// ─── WORK ───────────────────────────────────────────────────────────

const workTemplate = `You are implementing a fix based on the following specification.

## Working Directory
You are working in a dedicated git worktree: %s
All your edits, commits, and test runs should happen here.

## Specification
%s

## Files to Change
%s

## Implementation Plan
%s

## Additional Feedback
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

func buildWorkPrompt(spec *models.SpecResult, worktreePath string, feedback string) string {
	files := "discover from the repository"
	if len(spec.FilesToChange) > 0 {
		files = strings.Join(spec.FilesToChange, "\n")
	}
	if feedback == "" {
		feedback = "(none)"
	}
	wtPath := worktreePath
	if wtPath == "" {
		wtPath = "(repository root)"
	}
	return fmt.Sprintf(workTemplate, wtPath, spec.Summary, files, spec.ImplementationPlan, feedback)
}

// ─── REVIEW ─────────────────────────────────────────────────────────

const reviewTemplate = `You are reviewing a pull request for correctness and quality.

## Review Environment
You are running in a dedicated review worktree at: %s
This is a clean checkout of the PR's head commit.  You can read source
files, run tests, and explore the codebase to verify the changes.

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
2. Read the affected files in the worktree to understand the full context.
3. Run tests if possible to verify correctness.
4. Determine whether the PR is safe to merge.
5. List specific issues and suggested fixes.

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

func buildReviewPrompt(diff string, worktreePath string) string {
	wtPath := worktreePath
	if wtPath == "" {
		wtPath = "(repository root)"
	}
	return fmt.Sprintf(reviewTemplate, wtPath, diff)
}

// ─── FIX ────────────────────────────────────────────────────────────

const fixTemplate = `You need to fix issues raised during code review.

## Working Directory
You are working in a git worktree: %s
Apply your fixes here, then commit.

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

func buildFixPrompt(review *models.ReviewResult, worktreePath string) string {
	wtPath := worktreePath
	if wtPath == "" {
		wtPath = "(repository root)"
	}
	return fmt.Sprintf(fixTemplate, wtPath,
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

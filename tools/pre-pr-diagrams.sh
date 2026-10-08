#!/usr/bin/env bash
# pre-pr-diagrams.sh — pre-PR hook: turn a branch diff into a before/after
# Mermaid section, have it verified by an agent that did not write it, and
# attach the verified result to the pull request body.
#
# Usage:
#   tools/pre-pr-diagrams.sh brief  [--base <ref>] [--out <dir>]
#   tools/pre-pr-diagrams.sh verify <diagrams.md> [--out <dir>]
#   tools/pre-pr-diagrams.sh attach <diagrams.md> [--pr <number>] [--verified-by <name>]
#   tools/pre-pr-diagrams.sh run    [--base <ref>] [--pr <number>]
#
# The author and the verifier must be different agents. Set
# PRE_PR_AUTHOR_CMD / PRE_PR_VERIFY_CMD to a command that reads the prompt on
# stdin and writes its answer on stdout; with neither set, the script writes the
# prompts under .pre-pr/<branch>/ and tells you which fresh agent gets which one.

set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

START_MARK="<!-- pre-pr-diagrams:start -->"
END_MARK="<!-- pre-pr-diagrams:end -->"
BASE="origin/main"
OUT=""
PR=""
VERIFIED_BY="an independent reviewer agent"
MAX_ROUNDS="${PRE_PR_MAX_ROUNDS:-3}"

die() { printf 'pre-pr-diagrams: %s\n' "$*" >&2; exit 1; }
note() { printf 'pre-pr-diagrams: %s\n' "$*" >&2; }

usage() {
	sed -n '2,15p' "$0" | sed 's/^# \{0,1\}//'
	exit "${1:-0}"
}

branch_slug() {
	local name
	name="$(git branch --show-current)"
	if [ -z "$name" ]; then
		name="detached-$(git rev-parse --short HEAD)"
	fi
	printf '%s' "${name//\//-}"
}

out_dir() {
	if [ -n "$OUT" ]; then
		printf '%s' "$OUT"
	else
		printf '.pre-pr/%s' "$(branch_slug)"
	fi
}

# changed_files prints the file list of the merge-base diff, one per line.
changed_files() {
	git diff --name-only "$MERGE_BASE" "$HEAD"
}

write_prompts() {
	local dir="$1" diff="$2" brief="$3" base_sha="$4" head_sha="$5"

	cat >"$dir/author-prompt.md" <<EOF
You are writing the "Before / after" section of a pull request.

Inputs (read them; do not guess):
- Full diff: $diff  (base $base_sha, head $head_sha)
- Change brief: $brief
- The repository is checked out at the head commit. Read the real files.

Produce a Mermaid diagram pair:
- "Before": the flow/structure of the subsystem this diff touches, at the base commit.
- "After": the same view at the head commit.

Rules:
- Both diagrams use the same node vocabulary and the same level of abstraction, so a
  reader can diff them by eye. Label only what the diff changes.
- Every node, edge and label must exist in the code at the commit it describes, and
  each diagram gets short evidence notes citing file:line.
- Show the flow the diff actually changes, not the whole system. Keep it under ~20 nodes.
- Pick one diagram type and use it for both: flowchart LR for control flow,
  sequenceDiagram for call ordering, stateDiagram-v2 for state.
- No invented components, no aspirational behaviour, no placeholder nodes.
- If the change cannot honestly be shown as a before/after pair, say so and explain
  why instead of drawing something misleading.

Output markdown, exactly this shape:

## Before / after

### Before

\`\`\`mermaid
<diagram>
\`\`\`

<evidence notes, file:line>

### After

\`\`\`mermaid
<diagram>
\`\`\`

<evidence notes, file:line>
EOF

	cat >"$dir/verify-prompt.md" <<EOF
You are an independent reviewer. You did not write these diagrams and you must not
trust their narrative: verify every claim against the code.

Inputs:
- Full diff: $diff  (base $base_sha, head $head_sha)
- Diagrams under review: $dir/diagrams.md
- The repository is checked out at the head commit; use git to read the base commit.

For each diagram check:
1. Every node, edge and label exists at the commit it claims to describe
   (Before = $base_sha, After = $head_sha). Cite file:line for each.
2. The After diagram contains everything the diff adds to that flow, and the Before
   diagram everything the diff removes. Nothing else.
3. The two diagrams use the same vocabulary and abstraction, so they are comparable.
4. No invented behaviour and no component the diff does not touch.

Return ONLY this JSON, no prose:

{"verdict":"approve","errors":[{"diagram":"before","claim":"...","evidence":"path:line","why":"..."}],"notes":"..."}

Approve only if every claim is backed by code you actually read. If a claim cannot
be verified, reject it.
EOF
}

cmd_brief() {
	local dir diff brief base_sha head_sha
	HEAD="$(git rev-parse HEAD)"
	MERGE_BASE="$(git merge-base "$BASE" "$HEAD")" || die "no merge base with $BASE"
	base_sha="$(git rev-parse --short "$MERGE_BASE")"
	head_sha="$(git rev-parse --short "$HEAD")"
	dir="$(out_dir)"
	diff="$dir/diff.patch"
	brief="$dir/brief.md"
	mkdir -p "$dir"

	git diff --no-color "$MERGE_BASE" "$HEAD" >"$diff"
	{
		printf '# Change brief\n\n'
		printf -- '- Branch: `%s`\n' "$(git branch --show-current)"
		printf -- '- Base: `%s` (%s)\n' "$BASE" "$base_sha"
		printf -- '- Head: `%s`\n' "$head_sha"
		printf -- '- Diff: `%s`\n\n' "$diff"
		printf '## Files\n\n```\n'
		git diff --stat "$MERGE_BASE" "$HEAD"
		printf '```\n\n## Commits\n\n```\n'
		git log --oneline "$MERGE_BASE..$HEAD"
		printf '```\n'
	} >"$brief"

	write_prompts "$dir" "$diff" "$brief" "$base_sha" "$head_sha"

	note "wrote $dir"
	printf '%s\n' "$dir"
}

cmd_verify() {
	local diagrams="$1" dir prompt
	[ -f "$diagrams" ] || die "no such diagrams file: $diagrams"
	dir="$(out_dir)"
	[ -f "$dir/verify-prompt.md" ] || die "run 'brief' first ($dir/verify-prompt.md missing)"
	prompt="$dir/verify-input.md"
	{ cat "$dir/verify-prompt.md"; printf '\n---\n\n'; cat "$diagrams"; } >"$prompt"

	if [ -z "${PRE_PR_VERIFY_CMD:-}" ]; then
		note "PRE_PR_VERIFY_CMD is not set."
		note "Hand $prompt to a fresh agent (one that did not write the diagrams) and"
		note "save its JSON verdict, then run: $0 attach $diagrams"
		exit 2
	fi
	eval "$PRE_PR_VERIFY_CMD" <"$prompt" | tee "$dir/verdict.txt"
}

cmd_attach() {
	local diagrams="$1" dir body clean block
	[ -f "$diagrams" ] || die "no such diagrams file: $diagrams"
	dir="$(out_dir)"

	if [ -z "$PR" ]; then
		PR="$(gh pr view --json number --jq .number 2>/dev/null)" ||
			die "no open PR for this branch; pass --pr <number>"
	fi
	[ -n "$PR" ] || die "no open PR for this branch; pass --pr <number>"

	body="$dir/pr-body.md"
	clean="$dir/pr-body.clean.md"
	block="$dir/pr-body.new.md"
	gh pr view "$PR" --json body --jq .body >"$body"

	awk -v s="$START_MARK" -v e="$END_MARK" '
		$0 == s { skip = 1; next }
		$0 == e { skip = 0; next }
		!skip { print }
	' "$body" >"$clean"

	{
		cat "$clean"
		printf '\n%s\n' "$START_MARK"
		cat "$diagrams"
		printf '\n_Verified against the diff at `%s` (base `%s`) by %s._\n' \
			"$(git rev-parse --short HEAD)" "$(git rev-parse --short "$(git merge-base "$BASE" HEAD)")" "$VERIFIED_BY"
		printf '%s\n' "$END_MARK"
	} >"$block"

	gh pr edit "$PR" --body-file "$block" >/dev/null
	note "updated PR #$PR body with the verified Before / after section"
}

cmd_run() {
	local dir diagrams round verdict
	dir="$(cmd_brief)"
	diagrams="$dir/diagrams.md"

	if [ -z "${PRE_PR_AUTHOR_CMD:-}" ]; then
		note "PRE_PR_AUTHOR_CMD is not set."
		note "Hand $dir/author-prompt.md to an agent, save its markdown as $diagrams,"
		note "then run: $0 verify $diagrams"
		exit 2
	fi

	round=0
	while :; do
		round=$((round + 1))
		[ "$round" -le "$MAX_ROUNDS" ] || die "no approved diagram after $MAX_ROUNDS rounds"
		eval "$PRE_PR_AUTHOR_CMD" <"$dir/author-prompt.md" >"$diagrams"
		verdict="$(cmd_verify "$diagrams" | tail -n 1)"
		if printf '%s' "$verdict" | grep -Eq '"verdict"[[:space:]]*:[[:space:]]*"approve"'; then
			cmd_attach "$diagrams"
			return 0
		fi
		note "round $round rejected: $verdict"
		printf '\n---\n\nReviewer rejected the previous attempt:\n\n%s\n' "$verdict" \
			>>"$dir/author-prompt.md"
	done
}

while [ $# -gt 0 ]; do
	case "$1" in
	brief | verify | attach | run) CMD="$1" ;;
	--base) BASE="${2:?--base needs a ref}"; shift ;;
	--out) OUT="${2:?--out needs a dir}"; shift ;;
	--pr) PR="${2:?--pr needs a number}"; shift ;;
	--verified-by) VERIFIED_BY="${2:?--verified-by needs a name}"; shift ;;
	-h | --help) usage 0 ;;
	-*) die "unknown flag: $1" ;;
	*) ARG="$1" ;;
	esac
	shift
done

case "${CMD:-}" in
brief) cmd_brief ;;
verify) cmd_verify "${ARG:?verify needs the diagrams file}" ;;
attach) cmd_attach "${ARG:?attach needs the diagrams file}" ;;
run) cmd_run ;;
*) usage 1 ;;
esac

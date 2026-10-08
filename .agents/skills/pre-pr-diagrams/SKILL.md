---
name: pre-pr-diagrams
description: Pre-PR hook that turns a branch diff into before/after Mermaid diagrams and has them verified by a second agent before they are attached to the pull request. Use before opening a PR, when updating one, or when a PR body has no Before / after section.
---

# Pre-PR before/after diagrams

Every pull request that changes behaviour carries a **Before / after** section in
its body: one Mermaid diagram of the affected flow at the base commit, one at the
head commit, drawn at the same level of abstraction. The section is only attached
after an agent that did **not** write it has checked it against the diff.

Skip the section only when the diff cannot change behaviour: documentation,
comments, formatting, or test-only changes. Say so in the PR body instead of
drawing something decorative.

## The hook

```bash
tools/pre-pr-diagrams.sh brief          # diff + prompts under .pre-pr/<branch>/
tools/pre-pr-diagrams.sh verify <diagrams.md>
tools/pre-pr-diagrams.sh attach <diagrams.md> [--pr <number>]
tools/pre-pr-diagrams.sh run            # brief → author → verify → attach
```

`run` drives the whole loop when `PRE_PR_AUTHOR_CMD` / `PRE_PR_VERIFY_CMD` are set
to commands that read a prompt on stdin and write an answer on stdout. Without
them, the script writes `.pre-pr/<branch>/author-prompt.md` and
`.pre-pr/<branch>/verify-prompt.md` and stops — the calling agent fills both
roles. `.pre-pr/` is scratch space and is never committed.

## Two agents, never one

1. **Author** — may be the agent doing the work. It reads the diff and the real
   files and writes `.pre-pr/<branch>/diagrams.md`.
2. **Verifier** — a *fresh* agent that has not seen the author's reasoning. Give
   it `.pre-pr/<branch>/verify-input.md` (or the verify prompt plus the diagrams)
   and nothing else: no summary of the change, no defence of the diagrams. It
   must read the code at both commits and cite `file:line` for every claim.

Dispatch the verifier with a separate context — in DSH, the `subagent` tool; in
Claude Code, a `Task`; anywhere else, a new CLI session. The verifier returns
only:

```json
{"verdict":"approve","errors":[{"diagram":"before","claim":"…","evidence":"path:line","why":"…"}],"notes":"…"}
```

`approve` requires every claim to be backed by code the verifier actually read.
A rejection goes back to the author with the errors verbatim; re-verify after
every fix. After three rejected rounds, stop and report the disagreement in the
PR body rather than attaching an unverified diagram.

## Diagram bar

- Same diagram type and the same node vocabulary before and after, so the pair
  can be diffed by eye. `flowchart LR` for control flow, `sequenceDiagram` for
  call ordering, `stateDiagram-v2` for state.
- Only the flow the diff changes — not the whole subsystem. Under ~20 nodes.
- Every node, edge and label must exist at the commit it describes; label only
  what changed.
- No invented components, no aspirational behaviour, no placeholder nodes.
- If the change cannot honestly be drawn as a before/after pair, write that
  instead of a misleading diagram.

## Attach

`attach` replaces the marked block in the PR body, so re-running it after a
review round is idempotent, and stamps the base/head SHAs plus who verified it.
Run it before requesting review, not after.

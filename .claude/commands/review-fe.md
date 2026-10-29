---
description: Review frontend changes (diff mode) or audit named paths in full (audit mode). Flags only, no code edits.
argument-hint: [staged | <base-branch> | audit <paths...>]
allowed-tools: Bash(git diff:*), Bash(git status:*), Bash(git log:*), Bash(mkdir:*), Read, Grep, Glob, Write
---

Arguments: $ARGUMENTS

Pick the mode from the first word of the arguments:

- First word is `audit`: **audit mode**. The remaining words are files or directories, relative to the repo root. If none follow, stop and ask which paths to audit.
- Anything else, or empty: **diff mode**. `staged` means staged changes only, a branch name means everything changed since that branch, and empty means all uncommitted changes (staged and unstaged) under `frontend/`.

Steps:

1. Read `.claude/review/fe-review-rubric.md` in full and follow it exactly: the mode rules, the checks, and the output format for the chosen mode.
2. Read the frontend section of `CONVENTIONS.md` (and `CLAUDE.md` if it has frontend rules), and treat its rules as part of the rubric.
3. Diff mode: run `git status` and the appropriate `git diff`, then read the changed files and the sibling components and hooks they touch.
4. Audit mode: read every file under the named paths in full. Do not use git diff.
5. Produce the report in the rubric's output format.

Audit mode only: after producing the report, save it to `.claude/review/reports/audit-<scope>.md`, where `<scope>` is a short slug of the paths (e.g. `features-gallery`). Create the folder with `mkdir -p` if needed and overwrite an existing report for the same scope. This report file is the only file you may create.

Never edit, create, or delete any other file.

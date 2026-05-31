---
name: pusher
description: >
  Full git push workflow: status, diff, commit, push. Reviews every changed file,
  generates a Conventional Commits message, stages all changes, commits, and pushes.
  Use when user says "push", "ship it", "commit and push", "/pusher", or invokes
  /pusher. Also auto-triggers on "deploy", "send it", "push changes".
---

Review, commit, push. No gimmicks.

## Workflow

Execute these steps in order. Report progress at each step.

1. **git status** — `git status --short`. Show what changed. If nothing staged or modified, stop and report "nothing to commit".

2. **git diff** — `git diff` for unstaged changes, `git diff --cached` for staged. Show every changed file's diff. Do NOT skip any file.

3. **Conventional Commits message** — write a single commit message following Conventional Commits:
   - Types: `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `chore`, `build`, `ci`, `style`, `revert`
   - `<type>(<scope>): <imperative summary>` — scope optional
   - Subject ≤50 chars, hard cap 72, no trailing period
   - Body only for non-obvious *why*, breaking changes, migrations
   - Never: "this commit", "I", "we", emoji, AI attribution

4. **Ask confirmation** — show the proposed commit message. Ask user to confirm before committing. Do NOT commit without confirmation.

5. **Stage** — `git add -A` (or specific files if user specifies). Never force-add ignored files.

6. **Commit** — `git commit -m "..."`. If body present, use `git commit -m "subject" -m "body" ...`.

7. **Push** — `git push`. If upstream not set, use `git push --set-upstream origin <branch>`.

## Boundaries

- Always ask confirmation before committing. Never commit without user approval.
- If push fails (rejected, non-fast-forward), stop and report the error. Do NOT force push.
- If there are merge conflicts or the working tree is dirty in unexpected ways, stop and ask.
- Do NOT amend previous commits. Do NOT rebase unless explicitly asked.
- Do NOT push to `main`/`master` without explicit user override when on a protected branch.

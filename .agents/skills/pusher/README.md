# pusher

Full git push workflow: status → diff → commit → push.

## What it does

1. Runs `git status` and `git diff` to review every changed file
2. Generates a Conventional Commits message for the changes
3. Asks for confirmation before committing
4. Stages, commits, and pushes to origin

## How to invoke

```
/pusher
```

Also triggers on: "push", "ship it", "commit and push", "deploy", "send it".

## See also

- [SKILL.md](./SKILL.md) — full LLM-facing instructions

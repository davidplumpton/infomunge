# GEMINI.md

# Agent Instructions

This project uses **br** (beads rust) for issue tracking.

## Quick Reference

```bash
bv --robot-triage     # Find available work
br show <id>          # View issue details
br update <id> --status in_progress  # Claim work
br close <id>         # Complete work
br sync --flush-only  # Export beads DB to .beads/issues.jsonl for jj to commit
```

## Landing the Plane (Session Completion)

**MANDATORY WORKFLOW:**

Important: use jj for version control. Never use git commands. Always commit with an appropriate message, don't call `jj new` with a message.

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Run br sync --flush-only** - Export issue updates to `.beads/issues.jsonl`; `br` never runs VCS commands
5. **Only work on one task at a time before committing to VCS**
6. **Clean up**
7. **Verify** - Create an appropriate jj description and then run `jj commit -m <description>`
8. **Hand off** - Provide context for next session

### Best Practices

- Check `bv --robot-triage` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `br create` when you discover tasks
- Use descriptive titles and set appropriate priority/type, and dependencies between related items
- Always run `br sync --flush-only` before committing beads changes with `jj`
- Commit between finishing one beads issue and starting another
- Use jj commit with a description, don't use jj new with a description
- Use a 5 minute timeout when running cucumber tests
- Do not install new software
- Use golang only
- Always verify new features or changes by adding a cucumber test
- Stay within the infomunge directory, put temp files in tmp
- Track agent mistakes in `MIND_MAP.md` as they are discovered (what happened, why, and how to avoid repeating it)
- Track user preferences in `MIND_MAP.md` as they become clear (for example workflow/style likes and dislikes)

## Repo Tour

- CLI entrypoint: `cmd/infomunge/main.go`, `internal/cli/app.go`
- Runner + header parsing: `internal/runner/runner.go`
- Preprocessor (syntax transforms): `internal/preprocessor/*`
- Evaluator (AST + builtins + lazy): `internal/evaluator/*`
- Formats (read/write): `pkg/formats/*`

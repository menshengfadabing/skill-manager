# skill-manager

Terminal CLI to keep agent skills lean: **warehouse** (`~/.agents/skills-all`) + **dual activity sets** (`.agents/skills` and `.claude/skills`) via direct symlinks.

Design: [docs/skill-manager-design.md](docs/skill-manager-design.md)

## Build

```bash
go build -o skill ./cmd/skill
# or
go install ./cmd/skill
```

Ensure `$(go env GOPATH)/bin` is on `PATH`.

## Quick start

```bash
skill paths
skill sync          # migrate real dirs / fix two-level links
skill doctor
skill list
skill               # TUI
skill use default
skill init          # core profile
skill install-bundled   # from repo ./skills into warehouse
```

## Env

| Variable | Meaning |
| --- | --- |
| `SKILL_MANAGER_HOME` | Override `~/.agents` |
| `SKILL_MANAGER_CLAUDE` | Override `~/.claude/skills` |

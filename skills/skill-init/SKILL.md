---
name: skill-init
description: Interview the user about the project, then recommend a minimal skill set. Use when the user runs skill init, wants a clean skill activity set, or asks which skills this project needs. After agreement, enable via TUI or edit profile YAML then `skill use` — do not edit skill directories manually. skill-manager is global-only; use named profiles for different projects/scenarios.
---

# skill-init

## Goal

Produce a **small** enabled skill list as a **named profile** under `~/.agents/profiles/`. Prefer 5–10 skills max. skill-manager is global-only — different projects/scenarios = different profiles, not project-vs-global scope.

## Interview (ask briefly)

1. What kind of project?
2. Which agents? (Cursor, Claude Code, Codex, Qwen, …)
3. Must-have workflows?
4. Skills that must stay off?

## Apply

```bash
skill create <project-or-scenario-profile>
skill use <project-or-scenario-profile>
skill   # TUI toggle, writes back current profile
# or edit ~/.agents/profiles/<name>.yaml then:
skill use <name>
```

From a wipe:

```bash
skill init --yes
# then create/use the agreed set
```

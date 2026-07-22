# 支持的工具与工作目录（v1）

首批支持：**Claude Code / Codex / Cursor / Qwen Code / Pi**。

skill-manager **只维护全局一套**。切换启用集请用 profile（`skill use`），不再有项目级 / `-g` 双 scope。

## 共享约定

| 角色 | 路径 | 说明 |
| --- | --- | --- |
| 仓库目录 | `~/.agents/skills-all` | 实体只存一份 |
| 共享工作目录 | `~/.agents/skills` | Codex / Cursor / Qwen / Pi 都可读 |
| 配置档 | `~/.agents/profiles/*.yaml` | 启用清单 |

## 各工具工作目录（sync / init / restore 都会管）

| ID | 路径 | 工具 |
| --- | --- | --- |
| agents | `~/.agents/skills` | Codex、Cursor、Qwen、Pi（共享） |
| claude | `~/.claude/skills` | Claude Code |
| cursor | `~/.cursor/skills` | Cursor 用户 skills |
| codex | `~/.codex/skills` | Codex 遗留路径 |
| qwen | `~/.qwen/skills` | Qwen Code |
| pi | `~/.pi/agent/skills` | Pi |

## 项目残留（仅 sync 摄入）

若当前目录在 git 仓库内，且存在例如 `<repo>/.agents/skills` 等旧路径，`skill sync` 会把其中的实体**迁入** `~/.agents/skills-all`，但**不会**再往项目目录写启用集。

## 明确不管

- Cursor 内置 `~/.cursor/skills-cursor`（改坏 IDE）
- 其它未列入工具（Copilot / OpenCode / …）——后续再加

## 行为

1. **`skill sync`**：从上述全局工作目录摄入实体/旧链 → 迁入仓库目录 → 在每个全局工作目录重建直连仓库的软链；可选顺带清理项目残留实体。
2. **`skill init`**：各全局工作目录只保留 core（`skill-manager` + `skill-init`）软链。
3. **快照**：按每个 ID 分别备份 `*-skills/`，恢复时按 ID 盖回。

## 说明

- Qwen / Pi 同时读 `.agents` 与自家目录：两边都会被镜像，避免「只清一边仍脏」。
- Codex 以 `~/.agents/skills` 为主，遗留 `~/.codex/skills` 一并纳入以免漏扫。

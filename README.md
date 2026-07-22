# skill-manager

管理 Agent Skills：用精简活动集减少 token 浪费。

**只维护全局一套**。不同场景用 **profile** 切换启用集；不再区分项目级 / `-g`，避免全局与项目双倍注入、两套心智负担。

## 模型

| | 路径 | 职责 |
| --- | --- | --- |
| **唯一真源** | `~/.agents/skills-all` | 全部 skill 实体 |
| 全局工作目录 | `~/.agents/skills` 等（见 `docs/tools-paths.md`） | 各工具实际读取的启用集（软链 → warehouse） |
| Profile | `~/.agents/profiles/*.yaml` | 启用清单；`skill use <名>` 切换 |

若在 git 仓库里执行 `skill sync`，会**顺带摄入**项目下残留的 skill 目录到全局仓库，但**不会**再写回项目级启用集。

| Agent | 读哪里 | skill-manager 是否镜像 |
| --- | --- | --- |
| Codex / Cursor / Qwen / Pi | `~/.agents/skills`（共享）+ 各工具专用目录 | 是（见 `docs/tools-paths.md`） |
| Claude Code | `~/.claude/skills` | 是 |
| Cursor 内置 | `~/.cursor/skills-cursor` | **否** |

## 安装

预编译包（含 Windows）托管在 CloudBase：

https://menshengfadabing-d3ep6tl006fe480-1372800586.tcloudbaseapp.com/

命令说明见同站 [文档](https://menshengfadabing-d3ep6tl006fe480-1372800586.tcloudbaseapp.com/docs.html)。官网源码在仓库 `web/`（Vite + React）。

```bash
# macOS / Linux
curl -fsSL https://menshengfadabing-d3ep6tl006fe480-1372800586.tcloudbaseapp.com/install.sh | bash
```

```powershell
# Windows（PowerShell，勿用 irm|iex）
iwr -UseBasicParsing https://menshengfadabing-d3ep6tl006fe480-1372800586.tcloudbaseapp.com/install.ps1 -OutFile "$env:TEMP\sm-install.ps1"; powershell -ExecutionPolicy Bypass -File "$env:TEMP\sm-install.ps1"
```

也可从源码或 Docker：

```bash
go install ./cmd/skill
docker build -t skill-manager:latest .
docker run --rm -v "$HOME/.agents:/root/.agents" skill-manager:latest list
```

## 上手

```bash
skill sync --yes          # 实体进仓库 + 重建全局软链
skill create lean
skill                    # TUI 勾选后回车写回当前配置档
skill use lean           # 或 skill lean
skill profile            # 看当前用哪个配置档
```

## 命令

| 命令 | 说明 |
| --- | --- |
| `skill` | 交互界面；写回当前配置档 |
| `skill list` | 仓库目录 + 工作目录启用状态 |
| `skill create` / `delete` / `profile` / `use` | 配置档 |
| `skill doctor` | 体检 |
| `skill sync` | 实体 → 仓库目录，再重建全局工作目录软链 |
| `skill init` | 切到 core |
| `skill log` | 列出快照（`*` = 用户初始） |
| `skill restore <id\|initial>` | 恢复到某次快照 |
| `skill uninstall [--restore-initial]` | 清理本工具痕迹，可选恢复用户初始 |

标志：`--yes` · `--force` · `--dry-run` · `--restore-initial`（`-g` 已废弃，可省略）

## Profile

```yaml
name: lean
skills:
  - skill-manager
  - tdd
```

## 安全

`sync`/`init` 先备份到 `~/.agents/backups/`；非 TTY 需 `--yes`。

## 环境变量

| 变量 | 含义 |
| --- | --- |
| `SKILL_MANAGER_HOME` | 覆盖 `~/.agents`（含共享 skills-all） |
| `SKILL_MANAGER_CLAUDE` | 覆盖全局 Claude 工作目录 |

```bash
go test ./...
```

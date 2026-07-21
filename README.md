# skill-manager

管理 Agent Skills：用精简活动集减少 token 浪费。

**默认操作当前 git 仓库（项目级）**；加 `-g` / `--global` 才动用户主目录活动集。

## 模型

| | 路径 | 职责 |
| --- | --- | --- |
| **唯一真源** | `~/.agents/skills-all` | 全部 skill 实体（全局共用） |
| 项目活动集 | `<repo>/.agents/skills` + `<repo>/.claude/skills` | 本项目启用哪些（软链 → 全局 warehouse） |
| 全局活动集 | `~/.agents/skills` + `~/.claude/skills` | 全局启用哪些（`-g`） |
| Profile | `<scope>/.agents/profiles/*.yaml` | 启用清单 |

**项目下不再维护 `skills-all`。** 项目 `skill sync` 会把活动集里的实体目录**迁入全局** `skills-all`，再重建项目软链；不改全局活动集。

| Agent | 读哪里 | 是否单独镜像 |
| --- | --- | --- |
| Cursor / Codex / Qwen | `.agents/skills` | 否 |
| Claude Code | `.claude/skills` | 是（唯一例外） |

## 安装

```bash
go install ./cmd/skill
# 或 Docker
docker build -t skill-manager:latest .
docker run --rm -v "$HOME/.agents:/root/.agents" -v "$PWD:/work" -w /work skill-manager:latest list
```

将二进制拷到 PATH：

```bash
docker create --name skill-extract skill-manager:latest
docker cp skill-extract:/usr/local/bin/skill /usr/local/bin/skill
docker rm skill-extract
```

## 上手

```bash
# 全局：整理真源 + 可选全局活动集
skill sync -g --yes

# 项目：只选启用集
cd <仓库>
skill list
skill create lean
skill use lean
skill
```

若项目里还有旧的 `<repo>/.agents/skills-all`，可自行删除（真源已迁到 `~/.agents/skills-all` 后）。

## 命令

默认项目；`-g` 全局活动集。

| 命令 | 说明 |
| --- | --- |
| `skill` | TUI 启停（循环滚动）；写回当前 profile |
| `skill list` | warehouse（全局）+ 本 scope agents/claude |
| `skill create` / `delete` / `profile` / `use` | YAML profile |
| `skill doctor` | 体检 |
| `skill sync` | 活动集实体 → **全局** skills-all，再重建本 scope 软链 |
| `skill init` | 切到 core |

标志：`-g` · `--yes` · `--force` · `--dry-run`

## Profile

```yaml
name: lean
skills:
  - skill-manager
  - tdd
```

## 安全

`sync`/`init` 先备份到本 scope 的 `.agents/backups/`；非 TTY 需 `--yes`。

## 环境变量

| 变量 | 含义 |
| --- | --- |
| `SKILL_MANAGER_HOME` | 覆盖 `~/.agents`（含共享 skills-all） |
| `SKILL_MANAGER_CLAUDE` | 覆盖全局 Claude 活动集路径 |

```bash
go test ./...
```

# Deploy — 2026-07-21 CloudBase 安装包分发

## 目标

仅暴露下载/安装入口（含 Windows），无业务 API。

## 环境

- EnvId: `menshengfadabing-d3ep6tl006fe480`（体验版）
- URL: https://menshengfadabing-d3ep6tl006fe480-1372800586.tcloudbaseapp.com/
- 版本: v0.1.0

## 制品（应有且仅有）

| 内容 | 说明 |
| --- | --- |
| `skill-*` 多平台二进制 | CLI 本体 |
| `skills/skill-manager`、`skills/skill-init` + `skills.zip` | **仅**两个配套 skill（不含 cloudbase 等） |
| `install.sh` / `install.ps1` | 装 CLI + 解压配套 skill |
| `index.html` | Go 前置 + 功能说明 + Win/Unix 一键安装 |

## 命令

```bash
tcb hosting deploy ./dist/install -e menshengfadabing-d3ep6tl006fe480 --yes
```

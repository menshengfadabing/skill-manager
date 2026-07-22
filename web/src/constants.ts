export const HOST =
  'https://menshengfadabing-d3ep6tl006fe480-1372800586.tcloudbaseapp.com'

export const PATHS = {
  workGlobal: '~/.agents/skills',
  workClaudeGlobal: '~/.claude/skills',
  workCursorGlobal: '~/.cursor/skills',
  workCodexGlobal: '~/.codex/skills',
  workQwenGlobal: '~/.qwen/skills',
  workPiGlobal: '~/.pi/agent/skills',
  warehouse: '~/.agents/skills-all',
  profiles: '~/.agents/profiles/',
} as const

export const TOOLS = ['Claude Code', 'Codex', 'Cursor', 'Qwen Code', 'Pi'] as const

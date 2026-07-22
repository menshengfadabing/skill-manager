export const HOST =
  'https://menshengfadabing-d3ep6tl006fe480-1372800586.tcloudbaseapp.com'

export const PATHS = {
  workHub: '~/.agents/skills',
  workClaude: '~/.claude/skills',
  workCursor: '~/.cursor/skills',
  workCodex: '~/.codex/skills',
  workQwen: '~/.qwen/skills',
  workPi: '~/.pi/agent/skills',
  warehouse: '~/.agents/skills-all',
  profiles: '~/.agents/profiles/',
} as const

export const TOOLS = ['Claude Code', 'Codex', 'Cursor', 'Qwen Code', 'Pi'] as const

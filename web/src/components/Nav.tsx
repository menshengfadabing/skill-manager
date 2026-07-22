type NavProps = {
  current: 'home' | 'docs'
}

export function Nav({ current }: NavProps) {
  return (
    <header className="topbar">
      <div className="topbar-left">
        <a className={`btn${current === 'home' ? ' active' : ''}`} href="./" aria-current={current === 'home' ? 'page' : undefined}>
          🏠 首页
        </a>
        <a className={`btn${current === 'docs' ? ' active' : ''}`} href="./docs.html" aria-current={current === 'docs' ? 'page' : undefined}>
          📘 文档
        </a>
      </div>
      <div className="topbar-right">
        <a className="btn" href="https://github.com/menshengfadabing/skill-manager">
          🐙 GitHub
        </a>
      </div>
    </header>
  )
}

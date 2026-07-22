import { useState } from 'react'

type Props = {
  id: string
  children: string
}

export function CopyBlock({ id, children }: Props) {
  const [ok, setOk] = useState(false)

  async function onCopy() {
    try {
      await navigator.clipboard.writeText(children)
      setOk(true)
      window.setTimeout(() => setOk(false), 1400)
    } catch {
      /* ignore */
    }
  }

  return (
    <div className="code-row">
      <pre><code id={id}>{children}</code></pre>
      <button type="button" className={`btn copy${ok ? ' ok' : ''}`} onClick={onCopy} aria-label="复制">
        {ok ? '已复制' : '复制'}
      </button>
    </div>
  )
}

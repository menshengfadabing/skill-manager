import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles/global.css'
import { DocsPage } from './pages/DocsPage'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <DocsPage />
  </StrictMode>,
)

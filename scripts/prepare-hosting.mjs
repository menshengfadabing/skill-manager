import { cpSync, mkdirSync, rmSync, existsSync, readdirSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const www = join(root, 'dist', 'www')
const artifacts = join(root, 'dist', 'artifacts')
const hosting = join(root, 'dist', 'hosting')

if (!existsSync(www)) {
  console.error('missing dist/www — run npm run build in web/ first')
  process.exit(1)
}
if (!existsSync(artifacts)) {
  console.error('missing dist/artifacts — keep install binaries/scripts there')
  process.exit(1)
}

rmSync(hosting, { recursive: true, force: true })
mkdirSync(hosting, { recursive: true })

function copyTree(src, dest) {
  for (const name of readdirSync(src, { withFileTypes: true })) {
    const from = join(src, name.name)
    const to = join(dest, name.name)
    if (name.isDirectory()) {
      mkdirSync(to, { recursive: true })
      copyTree(from, to)
    } else {
      cpSync(from, to)
    }
  }
}

copyTree(www, hosting)
copyTree(artifacts, hosting)
console.log('prepared', hosting)
console.log('contents:', readdirSync(hosting).join(', '))

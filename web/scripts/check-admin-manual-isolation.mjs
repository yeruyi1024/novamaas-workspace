import assert from 'node:assert/strict'
import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const webRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const repositoryRoot = path.dirname(webRoot)
const manualRoot = path.join(repositoryRoot, 'docs/admin-manual')
const markers = []

for (const name of (await fs.readdir(manualRoot)).filter((name) =>
  name.endsWith('.md')
)) {
  const markdown = await fs.readFile(path.join(manualRoot, name), 'utf8')
  const introduction = markdown.split('\n\n').slice(1, 3).join('\n\n')
  assert.ok(introduction, `${name} must introduce the integration workflow`)
  markers.push({ name, text: introduction })
}

function checkContent(filename, content) {
  const decoded = content
    .replaceAll(/\\u\{([\da-f]+)\}/gi, (_, code) =>
      String.fromCodePoint(Number.parseInt(code, 16))
    )
    .replaceAll(/\\u([\da-f]{4})/gi, (_, code) =>
      String.fromCharCode(Number.parseInt(code, 16))
    )
  for (const marker of markers) {
    const quoted = JSON.stringify(marker.text).slice(1, -1)
    assert.ok(
      !decoded.includes(marker.text) && !decoded.includes(quoted),
      `${filename} exposes administrator documentation from ${marker.name}`
    )
  }
}

async function checkDirectory(directory) {
  let count = 0
  for (const entry of await fs.readdir(directory, { withFileTypes: true })) {
    const filename = path.join(directory, entry.name)
    if (entry.isDirectory()) count += await checkDirectory(filename)
    else if (entry.isFile()) {
      checkContent(
        path.relative(repositoryRoot, filename),
        await fs.readFile(filename, 'utf8')
      )
      count++
    }
  }
  return count
}

for (const name of ['api.json', 'relay.json']) {
  const filename = path.join(repositoryRoot, 'docs/openapi', name)
  const document = JSON.parse(await fs.readFile(filename, 'utf8'))
  const content = JSON.stringify(document)
  assert.ok(
    !content.includes('/api/admin-manual'),
    `${name} exposes the protected manual route`
  )
  assert.ok(
    !content.includes('AdminManualEnabled'),
    `${name} exposes the protected manual setting`
  )
  checkContent(name, content)
}

const customerRoot = path.join(repositoryRoot, 'docs/platform-api')
const customerFiles = (await fs.readdir(customerRoot)).filter((name) =>
  name.endsWith('.md')
)
assert.equal(
  customerFiles.length,
  7,
  'Review the public documentation manifest before changing its audience'
)
for (const name of customerFiles) {
  checkContent(name, await fs.readFile(path.join(customerRoot, name), 'utf8'))
}

// Require a completed build: an absent or empty dist is not evidence of isolation.
await fs.access(path.join(webRoot, 'dist/index.html'))
const builtFiles = await checkDirectory(path.join(webRoot, 'dist'))
const publicFiles = await checkDirectory(path.join(webRoot, 'public'))
assert.ok(builtFiles > 0)
console.log(
  `Platform integration guide isolation passed: ${markers.length} chapters, ${builtFiles} built files, ${publicFiles} public files, 7 customer documents and 2 OpenAPI documents.`
)

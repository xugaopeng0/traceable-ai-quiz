import { readFile, writeFile } from 'node:fs/promises'

const source = new URL('../docs/openapi.json', import.meta.url)
const target = new URL('../backend/internal/transport/httpapi/openapi.json', import.meta.url)
const data = await readFile(source)
JSON.parse(data.toString())
if (process.argv.includes('--check')) {
  if (!(await readFile(target)).equals(data)) throw new Error('OpenAPI 嵌入副本过期，请运行 npm run api:sync')
  console.log('OpenAPI 副本一致')
} else {
  await writeFile(target, data)
  console.log('OpenAPI 已同步')
}

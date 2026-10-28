import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { mkdtempSync, readFileSync, writeFileSync, rmSync, mkdirSync, cpSync, existsSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { createServer } from 'node:net'

const root = resolve('.')
const work = mkdtempSync(join(tmpdir(), 'quiz-skeleton-'))
const project = `quiz-skeleton-${process.pid}-${Date.now()}`
function command(args, options = {}) {
  const result = spawnSync('docker', args, { cwd: work, encoding: 'utf8', maxBuffer: 8 * 1024 * 1024, ...options })
  if (result.error || result.status !== 0) throw new Error(`docker ${args[0]} 失败：${result.error?.message ?? result.stderr}`)
  return result.stdout
}
function compose(...args) { return command(['compose', ...args]) }
const port = await new Promise((accept, reject) => { const server = createServer(); server.on('error', reject); server.listen(0, '127.0.0.1', () => { const address = server.address(); server.close(() => accept(address.port)) }) })
let started = false
try {
  if (process.argv.includes('--from-git')) {
    const clone = spawnSync('git', ['clone', '--quiet', '--no-hardlinks', root, work], { encoding: 'utf8' })
    if (clone.status !== 0) throw new Error(`干净克隆失败：${clone.stderr}`)
    assert.equal(existsSync(join(work, '.env')), false)
    assert.equal(existsSync(join(work, 'node_modules')), false)
  } else {
    // Test uncommitted edits in a clean directory without copied dependencies.
    for (const name of ['backend', 'frontend', 'Dockerfile', '.dockerignore', 'compose.yaml', '.env.example']) {
      cpSync(join(root, name), join(work, name), { recursive: true, filter: source => !source.split(/[\\/]/).some(part => ['node_modules', 'dist'].includes(part)) })
    }
  }
  const config = readFileSync(join(work, '.env.example'), 'utf8')
    .replace(/^COMPOSE_PROJECT_NAME=.*$/m, `COMPOSE_PROJECT_NAME=${project}`)
    .replace(/^APP_ENV=.*$/m, 'APP_ENV=test')
    .replace(/^WEB_PORT=.*$/m, `WEB_PORT=${port}`)
  writeFileSync(join(work, '.env'), config)
  started = true
  compose('up', '--build', '-d', '--wait', '--wait-timeout', '90')
  const base = `http://127.0.0.1:${port}`
  for (const [path, status, code] of [
    ['/health', 200], ['/api/health', 200], ['/missing-resource', 404, 'NOT_FOUND'],
    ['/health?status=unknown', 422, 'VALIDATION_ERROR'], ['/api/classes/C1/quiz/runs', 401, 'UNAUTHENTICATED'],
  ]) {
    const response = await fetch(base + path, { signal: AbortSignal.timeout(6000) })
    assert.equal(response.status, status, path)
    const body = await response.json()
    if (code) assert.equal(body.code, code)
    else assert.equal(body.status, 'ok')
    assert.ok(response.headers.get('x-request-id'))
  }
  const method = await fetch(base + '/health', { method: 'POST' })
  assert.equal(method.status, 405)
  const html = await (await fetch(base)).text()
  assert.match(html, /lang="zh-CN"/)
  const assets = [...html.matchAll(/(?:src|href)="(\/assets\/[^\"]+)"/g)].map(match => match[1])
  assert.ok(assets.length >= 2)
  for (const asset of assets) assert.equal((await fetch(base + asset)).status, 200)
  for (const service of ['api', 'web']) {
    const id = compose('ps', '-q', service).trim()
    const [inspect] = JSON.parse(command(['inspect', id]))
    assert.notEqual(inspect.Config.User, '0')
    assert.notEqual(inspect.Config.User, '')
    assert.equal(inspect.HostConfig.ReadonlyRootfs, true)
  }
  const invalid = spawnSync('docker', ['compose', 'run', '--rm', '--no-deps', '-e', 'API_PORT=70000', 'api'], { cwd: work, encoding: 'utf8' })
  assert.equal(invalid.status, 1)
  const invalidRow = JSON.parse(invalid.stdout.trim())
  assert.equal(invalidRow.error_code, 'CONFIG_INVALID')
  compose('stop', 'api')
  const before = Date.now()
  const unavailable = await fetch(base + '/health', { signal: AbortSignal.timeout(6000) })
  assert.equal(unavailable.status, 503)
  assert.equal((await unavailable.json()).code, 'DEPENDENCY_UNAVAILABLE')
  assert.ok(Date.now() - before < 5000, `代理失败响应耗时 ${Date.now()-before} ms，超过 5 秒预算`)
  compose('stop', 'web')
  let logCount = 0
  for (const service of ['api', 'web']) {
    const logs = compose('logs', '--no-color', '--no-log-prefix', service)
    for (const line of logs.split(/\r?\n/).filter(Boolean)) {
      const row = JSON.parse(line)
      for (const key of ['timestamp', 'level', 'service', 'env', 'commit', 'event']) assert.ok(key in row)
      if (row.event === 'http_request') for (const key of ['request_id', 'duration_ms', 'status']) assert.ok(key in row)
      logCount++
    }
  }
  assert.ok(logCount > 0)
  mkdirSync(join(root, 'tmp'), { recursive: true })
  writeFileSync(join(root, 'tmp/skeleton-smoke.json'), JSON.stringify({ at: new Date().toISOString(), project, cleanCopy: true, cleanClone: process.argv.includes('--from-git'), services: 2, logCount, result: 'passed' }, null, 2) + '\n')
  console.log(`独立干净副本验收通过：健康/401/404/405/422/503、外置配置、非 root、只读根、${logCount} 条 JSON 日志。`)
} finally {
  if (started) {
    try { compose('down', '--remove-orphans') } catch (error) { console.error(error.message) }
  }
  // Remove only the exact unique directory created by this run.
  if (existsSync(work) && work.startsWith(join(tmpdir(), 'quiz-skeleton-'))) rmSync(work, { recursive: true, force: true })
}

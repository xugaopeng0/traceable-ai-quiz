import { spawnSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { parseEnv } from 'node:util'

const config = parseEnv(readFileSync('.env.example', 'utf8'))
export function run(command, args, options = {}) {
  const result = spawnSync(command, args, { stdio: 'inherit', ...options })
  if (result.error || result.status !== 0) throw new Error(`${command} 执行失败，门禁未通过：${result.error?.message ?? result.status}`)
}

run(process.execPath, ['node_modules/@fission-ai/openspec/bin/openspec.js', 'validate', '--all', '--strict'])
run(process.execPath, ['scripts/sync-openapi.mjs', '--check'])
const buildArgs = ['GO_IMAGE', 'NODE_IMAGE', 'NGINX_IMAGE', 'ALPINE_IMAGE'].flatMap(key => ['--build-arg', `${key}=${config[key]}`])
run('docker', ['build', ...buildArgs, '--target', 'backend-test', '.'])
run('docker', ['build', ...buildArgs, '--target', 'frontend-build', '.'])
if (process.argv.includes('--unit-only')) {
  console.log('骨架单元/契约/构建检查通过；未执行容器验收。')
} else {
  run(process.execPath, ['scripts/skeleton-smoke.mjs'])
  console.log('骨架检查通过；不包含未来数据库、AI 与业务门禁。')
}

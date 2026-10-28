import { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import './style.css'

type State = { kind: 'loading' } | { kind: 'ready'; version: string } | { kind: 'error'; message: string }
declare global { interface Window { __QUIZ_CONFIG__?: { health_timeout_ms: number } } }

function App() {
  const [state, setState] = useState<State>({ kind: 'loading' })
  const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    const timeoutMs = window.__QUIZ_CONFIG__?.health_timeout_ms
    if (typeof timeoutMs !== 'number' || !Number.isInteger(timeoutMs) || timeoutMs < 100 || timeoutMs > 30000) {
      setState({ kind: 'error', message: '缺少或无效的页面运行配置' })
      return
    }
    const controller = new AbortController()
    const timer = setTimeout(() => {
      controller.abort()
      setState({ kind: 'error', message: '连接超时，请重试' })
    }, timeoutMs)
    setState({ kind: 'loading' })
    fetch('/api/health', { signal: controller.signal })
      .then(async response => {
        if (!response.ok) throw new Error(`服务暂不可达（HTTP ${response.status}）`)
        const data: unknown = await response.json()
        if (typeof data !== 'object' || data === null || !('status' in data) || data.status !== 'ok' || !('version' in data) || typeof data.version !== 'string') throw new Error('健康响应格式不正确')
        if (!controller.signal.aborted) setState({ kind: 'ready', version: data.version })
      })
      .catch((error: unknown) => { if (!controller.signal.aborted) setState({ kind: 'error', message: error instanceof Error ? error.message : '连接失败，请重试' }) })
      .finally(() => clearTimeout(timer))
    return () => { clearTimeout(timer); controller.abort() }
  }, [attempt])
  return <main>
    <header><span className="mark">讲义 → 练习</span><span className="phase">工程骨架</span></header>
    <section>
      <p className="eyebrow">可溯源 AI 出题与生成质量评测</p>
      <h1>从这里，开始构建。</h1>
      <p className="intro">前后端工程壳已建立。此页面仅验证本站连接，组卷、讲义与评测功能将在后续变更中逐步实现。</p>
      <div className="status" role="status" aria-live="polite">
        {state.kind === 'loading' && <><span className="dot pending" />正在检查服务连接…</>}
        {state.kind === 'ready' && <><span className="dot" />API 已连接<span className="version">{state.version}</span></>}
        {state.kind === 'error' && <><span className="dot failed" />{state.message}<button onClick={() => setAttempt(value => value + 1)}>重试</button></>}
      </div>
      <p className="note">健康检查只表示进程可访问，不代表数据库或 AI 服务已经就绪。</p>
    </section>
    <footer><span>徐焜 · 徐高鹏</span><span>React + TypeScript / Go</span></footer>
  </main>
}

createRoot(document.getElementById('root')!).render(<App />)

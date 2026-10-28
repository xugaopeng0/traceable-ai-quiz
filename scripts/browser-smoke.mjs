import assert from 'node:assert/strict'
import { mkdirSync } from 'node:fs'
import { chromium } from '@playwright/test'

const base = process.env.BASE_URL
if (!base) throw new Error('请设置 BASE_URL，指向已经启动的骨架 web 地址')
const browser = await chromium.launch({ headless: true, ...(process.env.BROWSER_CHANNEL ? { channel: process.env.BROWSER_CHANNEL } : {}) })
try {
  const page = await browser.newPage()
  const errors = []
  page.on('pageerror', error => errors.push(error.message))
  mkdirSync('tmp', { recursive: true })
  for (const width of [1440, 768, 375]) {
    await page.setViewportSize({ width, height: 900 })
    await page.goto(base)
    await page.getByText('API 已连接', { exact: false }).waitFor()
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth))
    assert.equal(await page.locator('button').count(), 0, '正常壳不应显示假业务按钮')
    if (width !== 768) await page.screenshot({ path: `tmp/skeleton-${width}.png`, fullPage: true })
  }
  // A successful probe must not turn into a false timeout five seconds later.
  await page.waitForTimeout(5500)
  await page.getByText('API 已连接', { exact: false }).waitFor()
  await page.route('**/api/health', route => route.fulfill({ status: 503, contentType: 'application/json', body: '{"code":"DEPENDENCY_UNAVAILABLE"}' }))
  await page.reload()
  await page.getByText('服务暂不可达', { exact: false }).waitFor()
  await page.unroute('**/api/health')
  await page.getByRole('button', { name: '重试' }).click()
  await page.getByText('API 已连接', { exact: false }).waitFor()
  assert.deepEqual(errors, [])
  console.log('浏览器验收通过：1440/768/375 px、正常/错误/重试、无溢出及控制台异常。')
} finally { await browser.close() }

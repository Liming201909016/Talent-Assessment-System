const { chromium } = require('playwright')
const crypto = require('crypto')
const fs = require('fs')
const path = require('path')
const assert = require('assert')

const BASE = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const statePath = process.env.STATE_FILE
const templateFile = process.env.TEMPLATE_FILE
if (!statePath) throw new Error('STATE_FILE is required')
if (!templateFile) throw new Error('TEMPLATE_FILE is required')
const state = JSON.parse(fs.readFileSync(statePath, 'utf8'))
const expectedTemplate = fs.readFileSync(templateFile)
const expectedSha = crypto.createHash('sha256').update(expectedTemplate).digest('hex')

;(async () => {
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 }, acceptDownloads: true })
  await context.addCookies([{ name: 'Admin-Token', value: state.adminToken, url: BASE }])
  const injectedCookie = (await context.cookies(BASE)).find(item => item.name === 'Admin-Token')
  assert(injectedCookie && injectedCookie.value.length === state.adminToken.length, 'Admin-Token cookie injection failed')
  const page = await context.newPage()
  const consoleErrors = []
  const authDiagnostics = []
  const templateRequests = []
  page.on('console', message => { if (message.type() === 'error') consoleErrors.push(message.text()) })
  page.on('request', request => { if (request.url().includes('/exam/api/competency/reports/template')) templateRequests.push(request.url()) })
  page.on('response', response => {
    if (!/\/(getInfo|getRouters|logout)$/.test(new URL(response.url()).pathname)) return
    const authorization = response.request().headers()['authorization'] || ''
    authDiagnostics.push(`${new URL(response.url()).pathname}:${response.status()}:auth=${authorization.startsWith('Bearer ')}`)
  })

  try {
    let loaded = false
    for (const route of ['/qu/template', '/exam/template', '/exam/template/index', '/template']) {
      await page.goto(`${BASE}/?v=template-management#${route}`, { waitUntil: 'domcontentloaded', timeout: 30000 })
      try {
        await page.getByText('00401 一期胜任力报告模板', { exact: true }).waitFor({ timeout: 12000 })
        loaded = true
        break
      } catch (error) {
        // Try the next dynamic-menu route candidate.
      }
    }
    if (!loaded) {
      const body = await page.locator('body').innerText().catch(() => '')
      throw new Error(`template page marker missing: url=${page.url()} auth=${authDiagnostics.join(',')} body=${body.slice(0, 500)}`)
    }
    await page.getByText(expectedSha, { exact: true }).waitFor({ timeout: 30000 })
    const body = await page.locator('body').innerText()
    assert(body.includes('报告模板管理'), 'unified page heading missing')
    assert(body.includes('49 控件 / 12 图表 / 0 可见占位符'), 'phase-one contract summary missing')
    assert(body.includes('MBTI 报告模板管理'), 'existing MBTI section missing')

    const actions = page.locator('.phase1-card .phase1-actions')
    await actions.locator('button').first().click()
    await page.waitForTimeout(2000)
    assert(templateRequests.some(url => url.includes('/template/download')), `download request missing: ${templateRequests.join(',')}`)

    const input = page.locator('.phase1-card input[type=file]')
    await input.setInputFiles(templateFile)
    await page.locator('.phase1-card .selected-file').getByText(path.basename(templateFile), { exact: true }).waitFor()
    await actions.locator('button.el-button--primary').click()
    const confirm = page.locator('.el-message-box').getByRole('button', { name: '上传并生效', exact: true })
    await confirm.waitFor({ timeout: 10000 })
    await confirm.click()
    await page.getByText('胜任力报告模板已上传并生效', { exact: true }).waitFor({ timeout: 120000 })
    await page.getByText(expectedSha, { exact: true }).waitFor({ timeout: 30000 })

    await page.setViewportSize({ width: 390, height: 844 })
    await page.reload({ waitUntil: 'domcontentloaded' })
    await page.getByText('00401 一期胜任力报告模板', { exact: true }).waitFor({ timeout: 30000 })
    const widths = await page.evaluate(() => ({ viewport: window.innerWidth, document: document.documentElement.scrollWidth }))
    assert(widths.document <= widths.viewport + 1, `mobile horizontal overflow: ${JSON.stringify(widths)}`)
    assert(await page.locator('.phase1-card .phase1-actions button.el-button--primary').isVisible(), 'mobile upload action hidden')

    const relevantErrors = consoleErrors.filter(text => !text.includes('favicon') && !text.includes('DevTools'))
    assert.strictEqual(relevantErrors.length, 0, `browser console errors: ${relevantErrors.join(' | ')}`)
    console.log('STAGING_TEMPLATE_MANAGEMENT_BROWSER_PASS')
    console.log(`sha256=${expectedSha}|desktop=1440x900|mobile=390x844|overflow=0`)
  } finally {
    await browser.close()
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

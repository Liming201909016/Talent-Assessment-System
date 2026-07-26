const { chromium } = require('playwright')
const fs = require('fs')
const assert = require('assert')

const BASE = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const statePath = process.env.STATE_FILE
if (!statePath) throw new Error('STATE_FILE is required')
const state = JSON.parse(fs.readFileSync(statePath, 'utf8'))

async function load(page, hash, marker) {
  await page.goto(`${BASE}/?v=admin-e2e#${hash}`, { waitUntil: 'domcontentloaded', timeout: 30000 })
  try {
    await page.waitForFunction(text => document.body && document.body.innerText.includes(text), marker, { timeout: 30000 })
  } catch (error) {
    const text = await page.locator('body').innerText().catch(() => '')
    throw new Error(`page marker timeout: hash=${hash} marker=${marker} url=${page.url()} body=${text.slice(0, 500)}`)
  }
  const body = await page.locator('body').innerText()
  assert(!body.includes('系统接口401异常'), `${hash} returned 401`)
  return body
}

;(async () => {
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } })
  await context.addCookies([{ name: 'Admin-Token', value: state.adminToken, domain: new URL(BASE).hostname, path: '/' }])
  const page = await context.newPage()
  const consoleErrors = []
  page.on('console', message => { if (message.type() === 'error') consoleErrors.push(message.text()) })

  try {
    await load(page, '/index', '人才综合素质评估系统')
    await load(page, '/exam/exam', '测评名称')

    const repoBody = await load(page, '/exam/repo', '00401')
    assert(repoBody.includes('00401'), '00401 virtual repository missing')

    const questionBody = await load(page, '/exam/competency/questions', 'D01-Q01')
    assert(questionBody.includes('D01-Q01'), 'competency question list did not load')

    const resultBody = await load(page, `/exam/competency-results/${state.examId}`, 'Result High')
    assert(resultBody.includes('Result High') && resultBody.includes('Result Low'), 'competency result rows missing')
    await page.getByRole('button', { name: '测试报告' }).first().click()
    await page.waitForURL(/#\/exam\/competency\/report\/[^?]+/)
    await page.waitForFunction(() => document.body && document.body.innerText.includes('临时测试报告'), null, { timeout: 30000 })
    const reportBody = await page.locator('body').innerText()
    assert(reportBody.includes('临时测试报告'), 'temporary report marker missing')
    assert(reportBody.includes('不可作为人才决策依据'), 'temporary report disclaimer missing')

    const paperBody = await load(page, `/exam/exam/paper/${state.examId}`, '考试名称')
    assert(!paperBody.includes('考试截图'), 'orphaned paper capture entry is still visible')

    const relevantErrors = consoleErrors.filter(text => !text.includes('favicon') && !text.includes('DevTools'))
    assert.strictEqual(relevantErrors.length, 0, `browser console errors: ${relevantErrors.join(' | ')}`)
    console.log('STAGING_ADMIN_BROWSER_E2E_PASS')
    console.log('flows=dashboard|exam-list|repo-00401|questions|results|test-report|paper-list')
  } finally {
    await browser.close()
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

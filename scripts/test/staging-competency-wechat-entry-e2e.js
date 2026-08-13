const { chromium } = require('playwright')
const assert = require('assert')

const BASE = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const EXAM_ID = process.env.EXAM_ID || '1786588375737209899'
const STU_FLAG = process.env.STU_FLAG || '0'
const WECHAT_ANDROID_UA = 'Mozilla/5.0 (Linux; Android 13; Pixel 7 Build/TQ3A.230805.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/110.0.0.0 Mobile Safari/537.36 MicroMessenger/8.0.47.2560(0x28002F3F) WeChat/arm64 Weixin NetType/WIFI Language/zh_CN ABI/arm64'

;(async () => {
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true, userAgent: WECHAT_ANDROID_UA })
  const page = await context.newPage()
  const errors = []
  const failedRequests = []
  page.on('console', message => { if (message.type() === 'error') errors.push(message.text()) })
  page.on('pageerror', error => errors.push(error.message))
  page.on('requestfailed', request => failedRequests.push(`${request.url()}:${request.failure() ? request.failure().errorText : ''}`))

  try {
    const url = `${BASE}/exam-entry.html?examId=${encodeURIComponent(EXAM_ID)}&stuFlag=${encodeURIComponent(STU_FLAG)}&repoCode=00401&isOpen=1`
    await page.goto(url, { waitUntil: 'networkidle', timeout: 60000 })
    await page.getByText('考生信息', { exact: true }).waitFor({ timeout: 30000 })
    await page.waitForFunction(() => document.querySelectorAll('.el-form-item').length > 0, null, { timeout: 30000 })
    const body = await page.locator('body').innerText()
    assert(body.includes('姓名') && body.includes('手机号'), `candidate fields missing: ${body.slice(0, 500)}`)
    assert(!body.includes('系统接口401异常'), 'anonymous QR entry returned 401')
    assert((await page.locator('#app').evaluate(element => element.children.length)) > 0, '#app is blank')
    assert(page.url().includes(`/#/my/exam/candidate/${EXAM_ID}/${STU_FLAG}/00401`), `bridge redirect mismatch: ${page.url()}`)
    const relevantErrors = errors.filter(text => !text.includes('favicon') && !text.includes('DevTools'))
    assert.strictEqual(relevantErrors.length, 0, `console errors: ${relevantErrors.join(' | ')}`)
    assert.strictEqual(failedRequests.length, 0, `failed requests: ${failedRequests.join(' | ')}`)
    const widths = await page.evaluate(() => ({ viewport: window.innerWidth, document: document.documentElement.scrollWidth }))
    assert(widths.document <= widths.viewport + 1, `horizontal overflow: ${JSON.stringify(widths)}`)
    console.log('STAGING_COMPETENCY_WECHAT_ENTRY_PASS')
    console.log(`url=${url}|fields=loaded|viewport=390x844|overflow=0`)
  } finally {
    await browser.close()
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

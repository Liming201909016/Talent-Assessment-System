const { chromium } = require('playwright')
const assert = require('assert')

const BASE_URL = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const labels = ['非常不符合', '不太符合', '一般', '比较符合', '非常符合']

function questions(complete) {
  return Array.from({ length: 3 }, (_, index) => ({
    id: `submit-guard-pq-${index + 1}`,
    code: `D01-Q0${index + 1}`,
    content: `交卷门禁E2E第 ${index + 1} 题`,
    answered: complete || index === 0,
    rawValue: complete || index === 0 ? 3 : null,
    options: labels.map((label, optionIndex) => ({ rawValue: optionIndex + 1, label }))
  }))
}

;(async () => {
  const browser = await chromium.launch({ headless: true })
  try {
    const authContext = await browser.newContext({ viewport: { width: 390, height: 844 }, isMobile: true })
    const authPage = await authContext.newPage()
    let unauthorizedPaperRequests = 0
    await authPage.route('**/exam/api/competency/participant/paper-detail', route => {
      unauthorizedPaperRequests++
      return route.abort()
    })
    await authPage.goto(`${BASE_URL}/?v=submit-auth-e2e#/`)
    await authPage.goto(`${BASE_URL}/?v=submit-auth-e2e#/exam/competency/start/submit-guard-paper`)
    await authPage.waitForURL(url => !url.hash.includes('/exam/competency/start/'))
    assert.strictEqual(await authPage.evaluate(() => sessionStorage.getItem('competencyPaperToken')), null)
    assert.strictEqual(unauthorizedPaperRequests, 0, 'missing token must block paper API before request')
    await authContext.close()

    const context = await browser.newContext({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true })
    await context.addInitScript(() => sessionStorage.setItem('competencyPaperToken', 'submit-guard-token'))
    const page = await context.newPage()
    let answeredCount = 1
    let submitRequests = 0

    await page.route('**/exam/api/competency/participant/paper-detail', route => {
      const rows = questions(false)
      return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, data: {
        id: 'submit-guard-paper', state: 0, totalCount: 3, answeredCount,
        unansweredCount: 3 - answeredCount,
        limitTime: new Date(Date.now() + 45 * 60 * 1000).toISOString(), questions: rows
      } }) })
    })
    await page.route('**/exam/api/competency/participant/fill-answer', route => {
      answeredCount++
      return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, data: { answeredCount, expired: false } }) })
    })
    await page.route('**/exam/api/competency/participant/submit', route => {
      submitRequests++
      assert.deepStrictEqual(route.request().postDataJSON(), { paperId: 'submit-guard-paper', submitType: 'manual' })
      return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, data: { completed: true } }) })
    })

    await page.goto(`${BASE_URL}/?v=submit-guard-e2e#/exam/competency/start/submit-guard-paper`)
    await page.getByRole('button', { name: '交卷' }).click()
    await page.getByText('还有 2 道题未答').waitFor()
    assert.strictEqual(submitRequests, 0, 'incomplete paper must not call submit API')
    assert.strictEqual(await page.getByRole('button', { name: '第 2 题，未答' }).getAttribute('aria-current'), 'true', 'guard must locate first unanswered question')

    await page.getByRole('radio', { name: '一般' }).click()
    await page.getByRole('button', { name: '第 3 题，未答' }).click()
    await page.getByRole('radio', { name: '比较符合' }).click()
    await page.getByRole('button', { name: '第 3 题，已答' }).waitFor()
    await page.getByRole('button', { name: '交卷' }).click()
    await page.getByRole('button', { name: '确定' }).click()
    await page.waitForURL('**/exam/thank-you')
    assert.strictEqual(submitRequests, 1, 'complete paper must submit exactly once')
    assert.strictEqual(await page.evaluate(() => sessionStorage.getItem('competencyPaperToken')), null, 'paper token must be removed after submission')
    console.log('COMPETENCY_SUBMIT_GUARD_E2E_PASS')
    await context.close()
  } finally {
    await browser.close()
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

const { chromium } = require('playwright')
const assert = require('assert')

const labels = ['非常不符合', '不太符合', '一般', '比较符合', '非常符合']
const baseUrl = (process.env.BASE_URL || 'http://127.0.0.1:8089').replace(/\/$/, '')

function paperData() {
  return {
    id: 'mobile-preview-paper',
    state: 0,
    totalCount: 40,
    answeredCount: 8,
    unansweredCount: 32,
    limitTime: new Date(Date.now() + 58 * 60 * 1000).toISOString(),
    questions: Array.from({ length: 40 }, (_, index) => ({
      id: `pq-${index + 1}`,
      code: `D05-Q${String(index + 1).padStart(2, '0')}`,
      content: index === 8
        ? '分析事情时，我会先形成结论，再寻找能够支持它的依据。'
        : `这是用于移动端布局验收的第 ${index + 1} 道胜任力题目。`,
      answered: index < 8,
      rawValue: index < 8 ? 4 : null,
      options: labels.map((label, optionIndex) => ({ rawValue: optionIndex + 1, label }))
    }))
  }
}

async function verifyViewport(browser, width, height, expectedColumns) {
  const context = await browser.newContext({ viewport: { width, height }, deviceScaleFactor: 1, isMobile: width < 600, hasTouch: true })
  await context.addInitScript(() => sessionStorage.setItem('competencyPaperToken', 'mobile-preview-token'))
  const page = await context.newPage()
  await page.route('**/exam/api/competency/participant/paper-detail', route => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ code: 0, data: paperData() })
  }))
  await page.goto(`${baseUrl}/?v=mobile-ui-test#/exam/competency/start/mobile-preview-paper`)
  await page.waitForSelector('.question-card')

  const metrics = await page.evaluate(() => {
    const options = document.querySelector('.scale-options')
    const nav = document.querySelector('.question-nav')
    const navButton = nav.querySelector('button')
    const option = options.querySelector('.el-radio')
    return {
      innerWidth: window.innerWidth,
      clientWidth: document.documentElement.clientWidth,
      scrollWidth: document.documentElement.scrollWidth,
      optionColumns: getComputedStyle(options).gridTemplateColumns.split(' ').length,
      navColumns: getComputedStyle(nav).gridTemplateColumns.split(' ').length,
      optionHeight: option.getBoundingClientRect().height,
      navButtonHeight: navButton.getBoundingClientRect().height,
      actionsWidth: document.querySelector('.actions').getBoundingClientRect().width
    }
  })

  assert.strictEqual(metrics.innerWidth, width, `${width}px viewport width must be preserved`)
  assert(metrics.scrollWidth <= metrics.clientWidth, `${width}px layout must not overflow horizontally`)
  assert.strictEqual(metrics.optionColumns, expectedColumns, `${width}px option column count`)
  assert(metrics.optionHeight >= 44, `${width}px option touch target must be at least 44px`)
  assert(metrics.navButtonHeight >= 40, `${width}px question navigation target must be at least 40px`)
  assert(metrics.actionsWidth <= metrics.clientWidth, `${width}px actions must fit the viewport`)

  if (width === 390) {
    await page.screenshot({ path: 'screenshots/competency-mobile-390.png', fullPage: true })
  }
  await context.close()
  return metrics
}

;(async () => {
  const browser = await chromium.launch({ headless: true })
  try {
    const phone = await verifyViewport(browser, 390, 844, 1)
    const tablet = await verifyViewport(browser, 768, 1024, 3)
    const desktop = await verifyViewport(browser, 1440, 900, 5)
    console.log(JSON.stringify({ baseUrl, phone, tablet, desktop }, null, 2))
    console.log('COMPETENCY_MOBILE_UI_PASS')
  } finally {
    await browser.close()
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

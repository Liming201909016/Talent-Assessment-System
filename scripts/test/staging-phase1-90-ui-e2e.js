const assert = require('assert')
const fs = require('fs')
const puppeteer = require('puppeteer')

const BASE = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const statePath = process.env.STATE_FILE
if (!statePath) throw new Error('STATE_FILE is required')
const state = JSON.parse(fs.readFileSync(statePath, 'utf8'))
if (!state.adminToken) throw new Error('STATE_FILE must contain adminToken')

;(async () => {
  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-features=HttpsUpgrades,HttpsFirstBalancedModeAutoEnable,AutomaticHttpsUpgrades']
  })
  const page = await browser.newPage()
  await page.setViewport({ width: 1440, height: 900 })
  const consoleErrors = []
  const failedRequests = []
  const httpErrors = []
  page.on('console', message => { if (message.type() === 'error') consoleErrors.push(message.text()) })
  page.on('pageerror', error => consoleErrors.push(error.message))
  page.on('requestfailed', request => failedRequests.push(`${request.url()}:${request.failure() && request.failure().errorText}`))
  page.on('response', response => { if (response.status() >= 400) httpErrors.push(`${response.status()}:${response.url()}`) })
  try {
    await page.setCookie({ name: 'Admin-Token', value: state.adminToken, url: BASE })
    const injectedCookie = (await page.cookies(BASE)).find(item => item.name === 'Admin-Token')
    assert(injectedCookie && injectedCookie.value.length === state.adminToken.length, 'Admin-Token cookie injection failed')
    console.log('UI_STAGE=open-question-bank')
    const pagingResponse = page.waitForResponse(response =>
      response.url().includes('/exam/api/competency/questions/paging') && response.request().method() === 'POST'
    )
    await page.goto(`${BASE}/#/exam/competency/questions`, { waitUntil: 'domcontentloaded', timeout: 30000 })
    let response
    try {
      response = await pagingResponse
    } catch (error) {
      const diagnostic = await page.evaluate(() => ({ url: location.href, body: (document.body && document.body.innerText || '').slice(0, 500) }))
      diagnostic.httpErrors = httpErrors
      diagnostic.cookiePresent = (await page.cookies(BASE)).some(item => item.name === 'Admin-Token')
      throw new Error(`question paging response timeout: ${JSON.stringify(diagnostic)}; ${error.message}`)
    }
    console.log('UI_STAGE=verify-question-bank')
    assert.strictEqual(response.status(), 200, 'question paging HTTP status')
    const payload = await response.json()
    assert([0, 200].includes(payload.code), `question paging failed: ${payload.msg}`)
    assert.strictEqual(payload.data.total, 90, 'question paging total')
    const apiTypes = payload.data.records.reduce((counts, row) => {
      counts[row.competencyQuestionType] = (counts[row.competencyQuestionType] || 0) + 1
      return counts
    }, {})
    assert(apiTypes.dimension > 0 && apiTypes.validity > 0, `first page must contain both types: ${JSON.stringify(apiTypes)}`)

    await page.waitForFunction(() => document.querySelectorAll('.el-table__body-wrapper .el-table__row').length > 0, { timeout: 30000 })
    const result = await page.evaluate(() => {
      const body = document.body.innerText
      const headers = Array.from(document.querySelectorAll('.el-table__header-wrapper th')).map(node => (node.innerText || '').trim())
      const typeIndex = headers.indexOf('题目类型')
      const typeCells = Array.from(document.querySelectorAll('.el-table__body-wrapper .el-table__row')).map(row => {
        const cells = row.querySelectorAll('td')
        return typeIndex >= 0 && cells[typeIndex] ? (cells[typeIndex].innerText || '').trim() : ''
      })
      return { body, headers, typeCells }
    })
    assert(result.body.includes('00401 胜任力测验题库'), 'dedicated question-bank title missing')
    assert(result.headers.includes('题目类型'), 'question-type table header missing')
    assert(result.typeCells.includes('维度题'), 'dimension type tag missing')
    assert(result.typeCells.includes('效度题'), 'validity type tag missing')
    assert(result.body.includes('共 90 条'), '90-row pagination total missing')

    const clicked = await page.evaluate(() => {
      const button = Array.from(document.querySelectorAll('button')).find(node => (node.innerText || '').trim() === '导入题目')
      if (!button) return false
      button.click()
      return true
    })
    assert(clicked, 'import button missing')
    await page.waitForFunction(() => Array.from(document.querySelectorAll('.el-dialog')).some(node => {
      const style = getComputedStyle(node)
      return style.display !== 'none' && (node.innerText || '').includes('导入胜任力题目')
    }), { timeout: 10000 })
    const dialogText = await page.evaluate(() => Array.from(document.querySelectorAll('.el-dialog')).find(node => {
      const style = getComputedStyle(node)
      return style.display !== 'none' && (node.innerText || '').includes('导入胜任力题目')
    }).innerText)
    assert(dialogText.includes('专用十列模板'), 'import dialog does not describe the ten-column contract')
    assert(!dialogText.includes('专用九列模板'), 'import dialog still exposes the retired nine-column contract')
    assert.strictEqual(failedRequests.length, 0, `failed browser requests: ${failedRequests.join(' | ')}`)
    assert.strictEqual(consoleErrors.length, 0, `browser console errors: ${consoleErrors.join(' | ')}`)

    console.log('STAGING_PHASE1_90_UI_PASS')
    console.log(`paging_total=${payload.data.total}|first_page_dimension=${apiTypes.dimension}|first_page_validity=${apiTypes.validity}`)
    console.log(`dom_rows=${result.typeCells.length}|dimension_tag=true|validity_tag=true|pagination_total=true`)
    console.log('import_dialog=ten_column:true|retired_nine_column:false')
    console.log('browser_errors=0|failed_requests=0')
  } finally {
    await browser.close()
  }
})().catch(error => {
  console.error(error.stack || error)
  process.exitCode = 1
})

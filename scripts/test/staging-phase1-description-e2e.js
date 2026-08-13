const { chromium } = require('playwright')
const fs = require('fs')
const assert = require('assert')

const BASE = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const statePath = process.env.STATE_FILE
const examId = process.env.EXAM_ID
const participantId = process.env.PARTICIPANT_ID
if (!statePath || !examId || !participantId) throw new Error('STATE_FILE, EXAM_ID and PARTICIPANT_ID are required')
const state = JSON.parse(fs.readFileSync(statePath, 'utf8'))

;(async () => {
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({ viewport: { width: 900, height: 700 } })
  await context.addCookies([{ name: 'Admin-Token', value: state.adminToken, url: BASE }])
  const page = await context.newPage()
  const consoleErrors = []
  page.on('console', message => { if (message.type() === 'error') consoleErrors.push(message.text()) })

  try {
    await page.goto(`${BASE}/?v=phase1-description#/my/exam/prepare/${examId}/${participantId}`, { waitUntil: 'domcontentloaded', timeout: 30000 })
    await page.getByText('本测评旨在全面了解您在各项工作情景中的行为表现与内在倾向', { exact: false }).waitFor({ timeout: 30000 })
    const body = await page.locator('.pre-exam').innerText()
    for (const text of [
      '本测评旨在全面了解您在各项工作情景中的行为表现与内在倾向',
      '作答说明：',
      '所有题目无对错、好坏之分，仅反映不同的行为偏好与应对倾向',
      '请依据您的第一反应如实作答，不必猜测“应该选什么”',
      '您的作答结果将严格保密，仅用于个人发展评估及组织人才决策',
      '请注意，本次测验共计90道题目，每道题目都必须作答'
    ]) assert(body.includes(text), `description missing: ${text}`)
    assert(!body.includes('共计80道题目'), 'old 80-question notice remains')
    assert(!body.includes('本测评采用五级量表，请根据自身真实情况选择最符合的选项'), 'old short description remains')
    const relevantErrors = consoleErrors.filter(text => !text.includes('favicon') && !text.includes('DevTools'))
    assert.strictEqual(relevantErrors.length, 0, `browser console errors: ${relevantErrors.join(' | ')}`)
    console.log('STAGING_PHASE1_DESCRIPTION_PASS')
    console.log(`examId=${examId}|participantId=${participantId}|questions=90`)
  } finally {
    await browser.close()
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

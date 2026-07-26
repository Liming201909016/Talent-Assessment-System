const { chromium } = require('playwright')
const assert = require('assert')

const BASE_URL = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const labels = ['非常不符合', '不太符合', '一般', '比较符合', '非常符合']

function buildQuestions() {
  return Array.from({ length: 40 }, (_, index) => ({
    id: `answer-flow-pq-${index + 1}`,
    code: `D05-Q${String(index + 1).padStart(2, '0')}`,
    content: `答题保存E2E第 ${index + 1} 题`,
    answered: index < 8,
    rawValue: index < 8 ? 4 : null,
    options: labels.map((label, optionIndex) => ({ rawValue: optionIndex + 1, label }))
  }))
}

;(async () => {
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true })
  await context.addInitScript(() => sessionStorage.setItem('competencyPaperToken', 'answer-flow-token'))
  const page = await context.newPage()
  const questions = buildQuestions()
  const saveRequests = []

  await page.route('**/exam/api/competency/participant/paper-detail', route => {
    const answeredCount = questions.filter(question => question.answered).length
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ code: 0, data: {
        id: 'answer-flow-paper', state: 0, totalCount: 40, answeredCount,
        unansweredCount: 40 - answeredCount,
        limitTime: new Date(Date.now() + 45 * 60 * 1000).toISOString(), questions
      } })
    })
  })

  await page.route('**/exam/api/competency/participant/fill-answer', async route => {
    const payload = route.request().postDataJSON()
    saveRequests.push(payload)
    const question = questions.find(item => item.id === payload.paperQuestionId)
    assert(question, 'saved question must belong to the paper')
    question.rawValue = payload.rawValue
    question.answered = true
    const answeredCount = questions.filter(item => item.answered).length
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 0, data: { answeredCount, expired: false } }) })
  })

  try {
    await page.goto(`${BASE_URL}/?v=answer-flow-e2e#/exam/competency/start/answer-flow-paper`)
    await page.getByRole('button', { name: '第 9 题，未答' }).click()
    await page.getByRole('radio', { name: '非常符合' }).click()
    await page.getByRole('status').filter({ hasText: '已保存' }).waitFor()

    assert.strictEqual(saveRequests.length, 1, 'one option click must produce one save request')
    assert.deepStrictEqual(saveRequests[0], { paperId: 'answer-flow-paper', paperQuestionId: 'answer-flow-pq-9', rawValue: 5 })
    const stats = await page.locator('.exam-stats strong').allTextContents()
    assert.deepStrictEqual(stats.slice(0, 2), ['9', '31'], 'answered/unanswered counts must update to 9/31')
    assert((await page.getByRole('button', { name: '第 9 题，已答' }).getAttribute('class')).includes('answered'), 'question navigation must update to answered')

    await page.reload()
    await page.getByRole('button', { name: '第 9 题，已答' }).click()
    assert(await page.getByRole('radio', { name: '非常符合' }).isChecked(), 'saved value must be restored after reload')
    assert.strictEqual(saveRequests.length, 1, 'reload must not submit another answer')
    console.log('COMPETENCY_ANSWER_FLOW_E2E_PASS')
  } finally {
    await browser.close()
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

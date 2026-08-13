const { chromium } = require('playwright')
const fs = require('fs')
const assert = require('assert')

const BASE = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const statePath = process.env.STATE_FILE
if (!statePath) throw new Error('STATE_FILE is required')
const state = JSON.parse(fs.readFileSync(statePath, 'utf8'))

async function setCheckbox(page, label, checked) {
  const checkbox = page.locator('.el-checkbox', { hasText: label }).first()
  const current = (await checkbox.getAttribute('class') || '').includes('is-checked')
  if (current !== checked) await checkbox.click()
}

;(async () => {
  const browser = await chromium.launch({ headless: true })
  const adminContext = await browser.newContext({ viewport: { width: 1440, height: 900 } })
  await adminContext.addCookies([{ name: 'Admin-Token', value: state.adminToken, url: BASE }])
  const page = await adminContext.newPage()
  let examId = ''
  const suffix = Date.now().toString(36)
  const title = `FB118首次保存-${suffix}`

  try {
    await page.goto(`${BASE}/?v=fb118#/exam/exam/add`, { waitUntil: 'domcontentloaded', timeout: 30000 })
    try {
      await page.getByText('胜任力测评', { exact: true }).waitFor({ timeout: 30000 })
    } catch (error) {
      const body = await page.locator('body').innerText().catch(() => '')
      throw new Error(`exam form marker missing: url=${page.url()} body=${body.slice(0, 500)}`)
    }
    await page.getByText('胜任力测评', { exact: true }).click()
    await page.locator('.el-form-item', { hasText: '测评名称' }).locator('input').fill(title)

    for (const label of ['姓名', '性别', '手机号']) await setCheckbox(page, label, true)
    for (const label of ['年龄', '身份证号', '单位/学校', '岗位', '部门', '学历', '专业']) await setCheckbox(page, label, false)

    const saveResponse = page.waitForResponse(response => response.url().includes('/exam/api/exam/exam/save') && response.request().method() === 'POST')
    await page.getByRole('button', { name: '保存', exact: true }).click()
    await page.locator('.el-message-box').getByRole('button', { name: '确定', exact: true }).click()
    const response = await saveResponse
    const requestBody = response.request().postDataJSON()
    assert.strictEqual(requestBody.requiredFields, 'name,gender,telephone', `first save payload=${requestBody.requiredFields}`)
    const payload = await response.json()
    assert(payload.code === 0 && payload.data && payload.data.id, JSON.stringify(payload))
    examId = payload.data.id

    const detail = await page.evaluate(async id => {
      const response = await fetch('/prod-api/exam/api/exam/exam/detail', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ id })
      })
      return response.json()
    }, examId)
    assert.strictEqual(detail.data.requiredFields, 'name,gender,telephone', `detail=${detail.data.requiredFields}`)

    const candidateContext = await browser.newContext({ viewport: { width: 1440, height: 900 } })
    const candidatePage = await candidateContext.newPage()
    await candidatePage.goto(`${BASE}/?v=fb118-candidate#/my/exam/candidate/${examId}/0/00401`, { waitUntil: 'domcontentloaded', timeout: 30000 })
    await candidatePage.getByText('考生信息', { exact: true }).waitFor({ timeout: 30000 })
    await candidatePage.waitForFunction(() => document.querySelectorAll('.el-form-item').length === 3, null, { timeout: 30000 })
    const labels = await candidatePage.locator('.el-form-item__label').allTextContents()
    const candidateBody = await candidatePage.locator('body').innerText()
    for (const expected of ['姓名', '性别', '手机号']) assert(labels.some(value => value.includes(expected)), `${expected} missing: url=${candidatePage.url()} labels=${labels} body=${candidateBody.slice(0, 500)}`)
    for (const unexpected of ['年龄', '单位', '岗位']) assert(!labels.some(value => value.includes(unexpected)), `${unexpected} unexpectedly visible: ${labels}`)
    await candidateContext.close()

    console.log('STAGING_FB118_REQUIRED_FIELDS_PASS')
    console.log(`examId=${examId}|requiredFields=name,gender,telephone|candidateFields=3`)
  } finally {
    if (examId) {
      await page.evaluate(async ({ id, token }) => {
        await fetch('/prod-api/exam/api/exam/exam/delete', {
          method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }, body: JSON.stringify({ ids: [id] })
        })
      }, { id: examId, token: state.adminToken }).catch(() => {})
    }
    await browser.close()
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

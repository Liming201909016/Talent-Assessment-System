const { execFileSync } = require('child_process')
const fs = require('fs')
const crypto = require('crypto')
const assert = require('assert')

const BASE = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const state = JSON.parse(fs.readFileSync(process.env.STATE_FILE, 'utf8'))
const sshKey = process.env.SSH_KEY
const sshHost = process.env.SSH_HOST || 'liming@20.200.136.133'
if (!sshKey) throw new Error('SSH_KEY is required')

async function api(path, body, expectError = false) {
  const response = await fetch(BASE + '/prod-api' + path, {
    method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + state.adminToken },
    body: JSON.stringify(body)
  })
  const payload = await response.json()
  if (expectError) {
    assert(![0, 200].includes(payload.code), `${path} unexpectedly succeeded`)
    return payload
  }
  assert([0, 200].includes(payload.code), `${path}: ${payload.msg}`)
  return payload.data
}

function mysql(sql) {
  return execFileSync('ssh', ['-o', 'ConnectTimeout=10', '-i', sshKey, sshHost, `printf %s ${Buffer.from(sql).toString('base64')} | base64 -d | sudo mysql -N -B element`], { encoding: 'utf8' }).trim()
}

;(async () => {
  const suffix = crypto.randomBytes(5).toString('hex')
  const dimensions = await api('/exam/api/competency/dimensions/list', {})
  const d01 = dimensions.find(item => item.code === 'D01')
  const d48 = dimensions.find(item => item.code === 'D48')
  assert(d01 && d48, 'D01/D48 missing')
  let publishedExamId = ''
  try {
    mysql(`UPDATE el_competency_dimension SET status=1 WHERE id='${d48.id}';`)
    const disabled = await api('/exam/api/exam/exam/save', {
      title: 'NEG-DISABLED-' + suffix, content: 'disabled dimension rejection', assessmentType: 'competency',
      scoringMode: 'competency_average', competencyReportAudience: 'leader', dimensionIds: [d48.id],
      joinType: 1, openType: 1, isOpen: 1, answerType: 1, state: 0, totalTime: 30, repoList: [], departIds: []
    }, true)
    assert(disabled.msg.includes('停用') || disabled.msg.includes('不存在'), `unexpected disabled error: ${disabled.msg}`)
    assert.strictEqual(mysql(`SELECT COUNT(*) FROM el_exam WHERE title='NEG-DISABLED-${suffix}';`), '0')
    mysql(`UPDATE el_competency_dimension SET status=0 WHERE id='${d48.id}';`)

    const exam = await api('/exam/api/exam/exam/save', {
      title: 'NEG-PUBLISHED-' + suffix, content: 'published edit rejection', assessmentType: 'competency',
      scoringMode: 'competency_average', competencyReportAudience: 'leader', dimensionIds: [d01.id],
      joinType: 1, openType: 1, isOpen: 1, answerType: 1, state: 0, totalTime: 30, repoList: [], departIds: []
    })
    publishedExamId = exam.id
    await api('/exam/api/competency/exams/publish', { examId: publishedExamId })
    const changedAudience = await api('/exam/api/exam/exam/save', {
      id: publishedExamId, title: 'NEG-PUBLISHED-' + suffix, content: 'published edit rejection', assessmentType: 'competency',
      scoringMode: 'competency_average', competencyReportAudience: 'frontline_employee', dimensionIds: [d01.id],
      joinType: 1, openType: 1, isOpen: 1, answerType: 1, state: 0, totalTime: 30, repoList: [], departIds: []
    }, true)
    assert(changedAudience.msg.includes('不能修改报告版本'), `unexpected published edit error: ${changedAudience.msg}`)
    console.log('STAGING_COMPETENCY_CONFIG_NEGATIVE_PASS')
    console.log('branches=disabled-dimension|published-audience-edit')
  } finally {
    mysql(`UPDATE el_competency_dimension SET status=0 WHERE id='${d48.id}';`)
    if (publishedExamId) await api('/exam/api/exam/exam/delete', { ids: [publishedExamId] }).catch(() => {})
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

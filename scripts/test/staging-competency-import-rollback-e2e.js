const { execFileSync } = require('child_process')
const fs = require('fs')
const crypto = require('crypto')
const XLSX = require('xlsx')
const assert = require('assert')

const BASE = (process.env.BASE_URL || 'http://20.200.136.133').replace(/\/$/, '')
const state = JSON.parse(fs.readFileSync(process.env.STATE_FILE, 'utf8'))
const sshKey = process.env.SSH_KEY
const sshHost = process.env.SSH_HOST || 'liming@20.200.136.133'
if (!sshKey) throw new Error('SSH_KEY is required')

function mysql(sql) {
  return execFileSync('ssh', ['-o', 'ConnectTimeout=10', '-i', sshKey, sshHost, `printf %s ${Buffer.from(sql).toString('base64')} | base64 -d | sudo mysql -N -B element`], { encoding: 'utf8' }).trim()
}

async function upload(path, buffer, expectedHash) {
  const form = new FormData()
  form.append('file', new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }), 'rollback-test.xlsx')
  if (expectedHash) form.append('expectedHash', expectedHash)
  const response = await fetch(BASE + '/prod-api' + path, { method: 'POST', headers: { Authorization: 'Bearer ' + state.adminToken }, body: form })
  return response.json()
}

;(async () => {
  const suffix = crypto.randomBytes(6).toString('hex')
  const code = 'ROLLBACK-' + suffix
  const itemNo = 900000 + Math.floor(Math.random() * 90000)
  const headers = ['维度序号', '维度名称', '题目编号', '维度内题号', '题目内容', '考察点', '计分方向', '启用状态', '备注']
  const row = ['1', '沟通表达', code, String(itemNo), '数据库失败回滚临时题', '事务回滚', '正向', '启用', '自动化临时数据']
  const workbook = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(workbook, XLSX.utils.aoa_to_sheet([headers, row]), '胜任力题目')
  const buffer = XLSX.write(workbook, { type: 'buffer', bookType: 'xlsx' })
  const preview = await upload('/exam/api/competency/questions/import-preview', buffer)
  assert.strictEqual(preview.code, 0, preview.msg)
  assert.strictEqual(preview.data.successCount, 1)
  assert.strictEqual(preview.data.errorCount, 0)
  try {
    mysql("DROP TRIGGER IF EXISTS test_competency_import_fail; CREATE TRIGGER test_competency_import_fail BEFORE INSERT ON el_qu FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='injected competency import failure';")
    const imported = await upload('/exam/api/competency/questions/import', buffer, preview.data.sha256)
    assert.notStrictEqual(imported.code, 0, 'injected database failure must reject import')
    assert(imported.msg.includes('回滚'), `unexpected import error: ${imported.msg}`)
    assert.strictEqual(mysql(`SELECT COUNT(*) FROM el_qu WHERE question_code='${code}';`), '0')
    console.log('STAGING_COMPETENCY_IMPORT_ROLLBACK_PASS')
  } finally {
    mysql('DROP TRIGGER IF EXISTS test_competency_import_fail;')
    const residualId = mysql(`SELECT id FROM el_qu WHERE question_code='${code}' LIMIT 1;`)
    if (residualId) mysql(`DELETE FROM el_qu WHERE id='${residualId}';`)
  }
})().catch(error => {
  console.error(error)
  process.exitCode = 1
})

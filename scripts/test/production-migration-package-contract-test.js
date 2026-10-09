const assert = require('assert')
const crypto = require('crypto')
const cp = require('child_process')
const fs = require('fs')
const path = require('path')

const root = path.resolve(__dirname, '..', '..')
const pkg = path.join(root, 'scripts', 'data', 'production-migration-20261009')
const schema = path.join(pkg, 'schema')
const mng = path.join(pkg, 'mng005')
const expectedSchema = [
  'competency_007_versions.sql',
  'competency_008_phase1_structures.sql',
  'competency_010_phase1_report_framework.sql',
  'competency_011_result_runs.sql',
  'competency_012_v2_dimension_catalog.sql',
  'competency_013_report_result_run_binding.sql',
  'management_traits_001_runtime.sql',
  'management_traits_003_new_draft.sql',
  'management_traits_004_report_reissues.sql'
]
assert.deepStrictEqual(fs.readdirSync(schema).sort(), expectedSchema.sort())
assert(!fs.existsSync(path.join(schema, 'management_traits_002_formal_registry.sql')))
assert(!fs.existsSync(path.join(schema, 'competency_009_phase1_identity_reset.sql')))
assert(!fs.existsSync(path.join(schema, 'competency_014_v2_report_content_draft.sql')))
assert(!fs.existsSync(path.join(schema, 'competency_015_v2_report_content_staging_approval.sql')))

const packageManifest = fs.readFileSync(path.join(pkg, 'MANIFEST.txt'), 'utf8')
assert(packageManifest.includes('customer_activation=existing-el_exam-state; imported-005-exams-visible-disabled-state-1'))
assert(packageManifest.includes('formal_registry_included=false'))
assert(packageManifest.includes('rehearsal_status=mysql-5.7.44-restored-copy-green'))
assert(packageManifest.includes('rehearsal_cleanup=owned-schema-remaining-0; main-runtime-unchanged'))
assert(packageManifest.includes('linux_server_sha256=753fad7a6134139b11ed3285c418da092c160b4fe81b9baf53dbf70f6d3cf0fc'))
assert(packageManifest.includes('frontend_index_sha256=abf93dd1fcd6ca6d94a1da393cc492594f6c6d00117152cd987b9c490bbfcbc4'))
const deployScript = path.join(root, 'Go-based Refactored System', 'deploy', 'production-release-20261009.sh')
const deployScriptSHA = crypto.createHash('sha256').update(fs.readFileSync(deployScript)).digest('hex')
assert(packageManifest.includes(`deployment_script_sha256=${deployScriptSHA}`))

const productionDropIn = fs.readFileSync(path.join(root, 'Go-based Refactored System', 'deploy', 'talent-assessment-production-mng005.conf'), 'utf8')
for (const setting of [
  'REPORT_EFFECTIVE_ENV=production', 'MNG_TEST_REPORT_ENV=production',
  'MNG_TEST_REPORT_DIR=/opt/talent-assessment/private/management-traits-test-reports',
  'MNG_TEST_TEMPLATE_PATH=/opt/talent-assessment/configs/export-templates/management-traits-002-test-only-v2.docx',
  'MNG_TEST_CONTENT_PATH=/opt/talent-assessment/configs/export-templates/management-traits-002-test-content-v1.xlsx'
]) {
  assert(productionDropIn.includes(setting), `production drop-in missing ${setting}`)
}

const dataPath = path.join(mng, '010_mng005_test_baseline_data.sql')
const data = fs.readFileSync(dataPath, 'utf8')
const profileV2 = fs.readFileSync(path.join(mng, '015_mng005_00502_profile_v2.sql'), 'utf8')
assert(data.includes("title LIKE '[TEST-%'"))
assert(data.includes("'00501','管理特质测验基层员工新版'"))
assert(data.includes("'00502','管理特质测验干部新版'"))
for (const approved of [
  '当我接手具有挑战性的工作时，我通常能鼓励大家创新，并提出创新的解决方案。',
  '当下属反对我的某个决定或者工作安排时，我会保持冷静和理性来应对。'
]) {
  assert(data.includes(approved), `approved 00502 text is absent from source rows: ${approved}`)
  assert(profileV2.includes(Buffer.from(approved).toString('hex')), `approved 00502 text is absent from frozen profile: ${approved}`)
}
assert(profileV2.includes('mng-00502-db-current-v2'))
assert(profileV2.includes('MNG005_00502_PROFILE_V2_OK'))
assert.strictEqual((data.match(/INSERT INTO `el_repo`/g) || []).length, 2)
assert.strictEqual((data.match(/INSERT INTO `el_qu`/g) || []).length, 280)
assert.strictEqual((data.match(/INSERT INTO `el_qu_answer`/g) || []).length, 1400)
assert.strictEqual((data.match(/INSERT INTO `el_qu_repo`/g) || []).length, 280)
assert.strictEqual((data.match(/INSERT INTO `el_exam`/g) || []).length, 2)
assert.strictEqual((data.match(/INSERT INTO `el_paper`/g) || []).length, 2)
assert.strictEqual((data.match(/INSERT INTO `el_mng_result_run`/g) || []).length, 2)
assert.strictEqual((data.match(/INSERT INTO `el_mng_report_revision`/g) || []).length, 2)
assert.strictEqual((data.match(/INSERT INTO `el_mng_report_reissue`/g) || []).length, 2)
assert(!/management_traits_002_formal_registry|el_mng_formal_/.test(data))

const manifest = fs.readFileSync(path.join(mng, 'MANIFEST.txt'), 'utf8')
assert(manifest.includes('identity_class=synthetic-test-only'))
assert(manifest.includes('labels_must_retain=[TEST-保留]'))
assert(manifest.includes('formal_registry_included=false'))

const activation = fs.readFileSync(path.join(mng, '020_mng005_customer_activation_state.sql'), 'utf8')
assert(activation.includes('UPDATE el_exam SET state=1'))
assert(activation.includes('MNG005_VISIBLE_DISABLED_OK'))
assert(!/INSERT INTO|DELETE FROM|DROP |TRUNCATE/i.test(activation))

const sums = fs.readFileSync(path.join(mng, 'SHA256SUMS'), 'utf8').trim().split(/\r?\n/)
for (const line of sums) {
  const match = line.match(/^([a-f0-9]{64})  \.\/(.+)$/)
  assert(match, `invalid checksum line: ${line}`)
  const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(mng, match[2]))).digest('hex')
  assert.strictEqual(actual, match[1], match[2])
}

const rollback = fs.readFileSync(path.join(mng, '090_mng005_test_baseline_rollback.sql'), 'utf8')
assert(rollback.includes('MNG005_ROLLBACK_OWNERSHIP_OK'))
assert(rollback.includes('MNG005_ROLLBACK_OK'))
for (const exact of ['el_paper_qu_answer', 'el_mng_paper_question_snapshot', 'el_mng_result_dimension', 'el_mng_result_module', 'el_mng_report_audit', 'el_mng_reissue_audit', 'state IN (0,1)', ')=3 AND']) {
  assert(rollback.includes(exact), `rollback ownership missing ${exact}`)
}
assert(!/DROP TABLE|DROP DATABASE|TRUNCATE/i.test(rollback))

const rehearsal = fs.readFileSync(path.join(root, 'scripts', 'db', 'production-migration-rehearsal-20261009.sh'), 'utf8')
assert(rehearsal.includes('gzip -dc "$backup/element.sql.gz" > "$backup/element.restore.sql"'))
assert(rehearsal.includes("^[[:space:]]*(CREATE[[:space:]]+DATABASE|USE[[:space:]])"))
assert(rehearsal.includes("(`element`|element)[[:space:]]*\\."))
assert(!rehearsal.includes('gzip -dc "$backup/element.sql.gz" | grep'))
for (const invariant of ['stage_assets', 'verify_assets', 'file_sha,file_bytes', 'rollback-before-activation', 'rollback-after-activation', 'verify_00502_profile_v2', 'PROFILE_V2_140_QUESTIONS_700_OPTIONS_PASS=1', 'ASSET_INSTALL_VERIFY_CLEANUP_PASS=1']) {
  assert(rehearsal.includes(invariant), `rehearsal missing ${invariant}`)
}
assert(!rehearsal.includes('rm -f "$backup/client.cnf"\nprintf \'MYSQL57_RESTORE_PASS'))
assert(rehearsal.includes('OWNED_SCHEMA_REMAINING=%s'))

const packageSumsRaw = fs.readFileSync(path.join(pkg, 'SHA256SUMS'), 'utf8')
assert(!packageSumsRaw.includes('\r'), 'package checksum manifest must use LF only')
const packageSums = packageSumsRaw.trim().split('\n')
for (const line of packageSums) {
  const match = line.match(/^([a-f0-9]{64})  \.\/(.+)$/)
  assert(match, `invalid package checksum line: ${line}`)
  const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(pkg, match[2]))).digest('hex')
  assert.strictEqual(actual, match[1], match[2])
}
const shaTool = process.platform === 'win32' ? 'C:\\Program Files\\Git\\usr\\bin\\sha256sum.exe' : 'sha256sum'
const nativeCheck = cp.spawnSync(shaTool, ['-c', 'SHA256SUMS'], { cwd: pkg, encoding: 'utf8' })
assert.strictEqual(nativeCheck.status, 0, nativeCheck.stderr || nativeCheck.stdout)
console.log('PRODUCTION_MIGRATION_PACKAGE_CONTRACT=PASS')

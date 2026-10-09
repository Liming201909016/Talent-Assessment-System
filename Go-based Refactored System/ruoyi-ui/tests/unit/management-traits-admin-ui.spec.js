// MT-ADMIN-DIRECT / MT-ADMIN-UI: docs/regression-tests.md (local synthetic only).
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import compiler from 'vue-template-compiler'
import babel from '@babel/core'
import fs from 'fs'
import path from 'path'
import * as product from '@/utils/managementTraitsProduct'

const api = {
  fetchManagementTraitsExamConfig: (...args) => fetchDetail(...args),
  fetchManagementTraitsAdminAccess: (...args) => getInfo(...args),
  canManageManagementTraits: store => store?.state?.user?.userId === 1,
  managementTraitsExamKnown: () => false,
  fetchManagementTraitsProfile: vi.fn(), fetchManagementTraitsResults: vi.fn(), fetchManagementTraitsResult: vi.fn(),
  fetchManagementTraitsReissueQualification: vi.fn(), fetchManagementTraitsReissues: vi.fn(),
  generateManagementTraitsReissue: vi.fn(), viewManagementTraitsReissue: vi.fn(), downloadManagementTraitsReissue: vi.fn(),
  generateManagementTraitsTestReport: vi.fn(), viewManagementTraitsTestReport: vi.fn(), downloadManagementTraitsTestReport: vi.fn()
}
const fetchDetail = vi.fn(), getInfo = vi.fn(), legacyList = vi.fn()
const defer = () => { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
const flush = async () => { for (let i = 0; i < 30; i++) await Promise.resolve() }
const exam = (extra = {}) => ({ id: 'exam-1', title: 'Synthetic assessment', repoCode: '00201', assessmentType: 'legacy', scoringMode: 'legacy', managementTraitsProfileFrozen: true, requiredFields: 'name', ...extra })
const row = (extra = {}) => ({ id: 'run-1', examId: 'exam-1', paperId: 'paper-1', participantType: 'candidate', participantId: 'person-1', status: 'completed', totalQuestionCount: 140, answeredQuestionCount: 140, overallScore: '50.000000', ...extra })
const report = (extra = {}) => ({ id: 'report-1', runId: 'run-1', paperId: 'paper-1', examId: 'exam-1', kind: 'verified_frozen_result', status: 'completed', createdAt: '2026-10-08T00:00:00Z', ...extra })
function compile(file) {
  const sfc = compiler.parseComponent(fs.readFileSync(path.resolve('src/views', file), 'utf8'))
  const code = babel.transformSync(sfc.script.content, { babelrc: false, configFile: false, plugins: ['@babel/plugin-transform-modules-commonjs'] }).code
  const module = { exports: {} }
  new Function('require', 'module', 'exports', code)(name => {
    if (name === '@/utils/managementTraitsProduct') return product
    if (name === '@/api/managementTraits') return api
    if (name === '@/api/exam/exam') return { fetchDetail }
    if (name === '@/api/login') return { getInfo }
    if (name === '@/api/tester/tester') return { listTester: legacyList, getListTester: legacyList }
    if (name.endsWith('.vue')) return { __esModule: true, default: { render: h => h('div') } }
    return {}
  }, module, module.exports)
  return { ...module.exports.default, ...compiler.compileToFunctions(sfc.template.content) }
}
function mount(entry = false, params = {}, query = {}) {
  return shallowMount(compile(entry ? 'user/exam/index.vue' : 'exam/exam/managementTraitsResults.vue'), {
    directives: { loading: () => {} },
    stubs: ['el-alert', 'pagination', 'el-descriptions', 'el-descriptions-item', 'el-empty', 'el-dropdown', 'el-dropdown-menu', 'el-dropdown-item'],
    mocks: {
      $route: { params: { examId: 'exam-1', isOpen: '1', ...params }, query, meta: {} },
      $store: { state: { user: { userId: 1 }, app: { device: 'desktop' } }, getters: { permissions: [] } },
      $router: { replace: vi.fn(), back: vi.fn() },
      $message: { error: vi.fn(), warning: vi.fn(), success: vi.fn() },
      $confirm: vi.fn(() => Promise.resolve()), parseTime: value => value || '—'
    }
  })
}
beforeEach(() => {
  Object.values(api).forEach(fn => fn.mockReset?.())
  getInfo.mockReset().mockResolvedValue({ code: 200, user: { userId: 1 }, permissions: [] })
  fetchDetail.mockReset().mockResolvedValue({ data: exam() })
  legacyList.mockReset().mockResolvedValue({ rows: [], total: 0 })
  api.fetchManagementTraitsProfile.mockResolvedValue({ data: { examId: 'exam-1', frozenAt: '2026-10-08T00:00:00Z' } })
  api.fetchManagementTraitsResults.mockResolvedValue({ data: [row()] })
  api.fetchManagementTraitsReissueQualification.mockResolvedValue({ data: { eligible: true, kind: 'verified_frozen_result', purpose: '仅供系统测试，不可作为人才决策依据' } })
  api.fetchManagementTraitsReissues.mockResolvedValue({ data: [report()] })
  api.generateManagementTraitsReissue.mockResolvedValue({ data: { report: report(), reused: false, purpose: '仅供系统测试，不可作为人才决策依据' } })
  api.viewManagementTraitsReissue.mockResolvedValue({ blob: new Blob(['%PDF-1.7'], { type: 'application/pdf' }), filename: 'synthetic.pdf' })
  api.downloadManagementTraitsReissue.mockResolvedValue({ blob: new Blob(['%PDF-1.7'], { type: 'application/pdf' }), filename: 'synthetic.pdf' })
})

describe('MT-ADMIN-DIRECT trusted server entry', () => {
  it.each(['00501', '00502'])('server frozen %s routes to shared new results only after real admin/profile checks', async code => {
    fetchDetail.mockResolvedValue({ data: exam({ repoCode: code }) }); const w = mount(true); await flush()
    expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ManagementTraitsResults', params: { examId: 'exam-1' } }); expect(api.fetchManagementTraitsProfile).toHaveBeenCalledWith('exam-1'); expect(legacyList).not.toHaveBeenCalled(); w.destroy()
  })
  it.each(['00501', '00502'])('server unfrozen %s never loads old results', async code => {
    fetchDetail.mockResolvedValue({ data: exam({ repoCode: code, managementTraitsProfileFrozen: false }) }); const w = mount(true); await flush()
    expect(legacyList).not.toHaveBeenCalled(); expect(w.vm.entryError).not.toBe(''); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it.each(['00201', '00202'])('RED queryless frozen %s replaces old entry before old list', async repoCode => {
    fetchDetail.mockResolvedValue({ data: exam({ repoCode }) }); const w = mount(true); await flush()
    expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ManagementTraitsResults', params: { examId: 'exam-1' } })
    expect(legacyList).not.toHaveBeenCalled(); expect(getInfo).toHaveBeenCalled(); w.destroy()
  })
  it('explicit legacy false ignores spoofed URL intent', async () => {
    fetchDetail.mockResolvedValue({ data: exam({ managementTraitsProfileFrozen: false, isManagementTraits: false, managementTraitsLifecycle: 'legacy' }) }); const w = mount(true); w.vm.$route.query.mngTest = '1'; await flush()
    expect(legacyList).toHaveBeenCalledTimes(1); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it.each([undefined, null, 'true', 1])('unknown frozen %s never loads old list', async flag => {
    fetchDetail.mockResolvedValue({ data: exam({ managementTraitsProfileFrozen: flag }) }); const w = mount(true); await flush()
    expect(legacyList).not.toHaveBeenCalled(); expect(w.vm.entryError).not.toBe(''); w.destroy()
  })
  it.each([{ user: null, permissions: ['*:*:*'] }, { user: { userId: 0 }, permissions: ['*:*:*'] }, { user: { userId: 2 }, roles: ['admin'], permissions: [] }])('fresh getInfo fails closed %j', async auth => {
    getInfo.mockResolvedValue({ code: 200, ...auth }); const w = mount(true); await flush()
    expect(w.vm.$router.replace).not.toHaveBeenCalled(); expect(legacyList).not.toHaveBeenCalled(); w.destroy()
  })
  it('getInfo failure ignores cached admin', async () => {
    getInfo.mockRejectedValue(new Error('authentication failed')); const w = mount(true); await flush()
    expect(w.vm.entryError).toContain('authentication failed'); expect(legacyList).not.toHaveBeenCalled(); w.destroy()
  })
  it.each(['00101', '00301'])('non-002 %s preserves old routing without profile probes', async repoCode => {
    fetchDetail.mockResolvedValue({ data: exam({ repoCode, managementTraitsProfileFrozen: undefined }) }); const w = mount(true); await flush()
    expect(legacyList).toHaveBeenCalledTimes(1); expect(getInfo).not.toHaveBeenCalled(); w.destroy()
  })
  it('competency routing remains independent', async () => {
    fetchDetail.mockResolvedValue({ data: exam({ assessmentType: 'competency', scoringMode: 'competency_average', repoCode: '', managementTraitsProfileFrozen: false, isManagementTraits: false, managementTraitsLifecycle: 'legacy' }) }); const w = mount(true); await flush()
    expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'CompetencyResults', params: { examId: 'exam-1' } }); expect(legacyList).not.toHaveBeenCalled(); w.destroy()
  })
  it('rejects mismatched server exam identity', async () => {
    fetchDetail.mockResolvedValue({ data: exam({ id: 'other' }) }); const w = mount(true); await flush()
    expect(w.vm.$router.replace).not.toHaveBeenCalled(); expect(legacyList).not.toHaveBeenCalled(); w.destroy()
  })
  it('cross-exam late detail cannot redirect or enable old state', async () => {
    const d = defer(); fetchDetail.mockReturnValue(d.promise); const w = mount(true); w.vm.$route.params.examId = 'exam-2'; d.resolve({ data: exam() }); await flush()
    expect(w.vm.$router.replace).not.toHaveBeenCalled(); expect(w.vm.entryReady).toBe(false); expect(legacyList).not.toHaveBeenCalled(); w.destroy()
  })
})

describe('MT-ADMIN-UI customer reissue row actions', () => {
  it.each([undefined, '', 'name,name', 'unknown'])('missing/invalid identity contract fails closed %s', async requiredFields => {
    fetchDetail.mockResolvedValue({ data: exam({ requiredFields }) }); const w = mount(true); await flush()
    expect(legacyList).not.toHaveBeenCalled(); expect(w.vm.$router.replace).not.toHaveBeenCalled(); expect(w.vm.entryError).not.toBe(''); w.destroy()
  })
  it('multiple archival revisions require explicit selection, never the old current', async () => {
    api.fetchManagementTraitsReissues.mockResolvedValue({ data: [report(), report({ id: 'report-2' })] })
    const w = mount(); await flush(); await w.vm.openReport(row()); expect(api.viewManagementTraitsReissue).not.toHaveBeenCalled()
    expect(w.vm.reportsVisible).toBe(true); w.vm.reportState(row()).selectedId = 'report-2'; await w.vm.openReport(row())
    expect(api.viewManagementTraitsReissue).toHaveBeenCalledWith('report-2'); w.destroy()
  })
  it('empty metadata is ungenerated only after successful query and never opens a file', async () => {
    api.fetchManagementTraitsReissues.mockResolvedValue({ data: [] }); const w = mount(); await flush(); await w.vm.openReport(row())
    expect(w.vm.reportStatus(row())).toBe('未生成'); expect(api.viewManagementTraitsReissue).not.toHaveBeenCalled(); w.destroy()
  })
  it('eligibility false prevents POST, no old fallback', async () => {
    api.fetchManagementTraitsReissueQualification.mockResolvedValue({ data: { eligible: false, kind: 'verified_frozen_result', purpose: 'TEST' } })
    const w = mount(); await flush(); await w.vm.generateReport(row()); expect(api.generateManagementTraitsReissue).not.toHaveBeenCalled(); expect(api.generateManagementTraitsTestReport).not.toHaveBeenCalled(); w.destroy()
  })
  it('generation 404 leaves visible results and reports specific uninstalled service', async () => {
    api.generateManagementTraitsReissue.mockRejectedValue(Object.assign(new Error('not installed'), { status: 404 }))
    const w = mount(); await flush(); await w.vm.generateReport(row()); expect(w.vm.rows).toHaveLength(1); expect(w.vm.generating).toBe(false)
    expect(w.vm.$message.error).toHaveBeenCalledWith(expect.stringContaining('报告服务未发布')); w.destroy()
  })
  it('13 dimensions/4 modules preserve exact facts and HALF_UP display including zero and NULL', async () => {
    const Dimensions = Array.from({ length: 13 }, (_, i) => ({ Name: `Synthetic dimension ${i}`, Score: i === 0 ? '0' : '425/8', Norm: '705/13', ScoreSum: 12, AnsweredCount: 10, QuestionCount: 10 }))
    const Modules = Array.from({ length: 4 }, (_, i) => ({ Key: `module-${i}`, Score: '50', DimensionCount: 3 }))
    api.fetchManagementTraitsResult.mockResolvedValue({ data: { RunID: 'run-1', PaperID: 'paper-1', ExamID: 'exam-1', Result: { Dimensions, Modules, OverallScore: '50' } } })
    const w = mount(); await flush(); await w.vm.showDetail(row()); expect(w.vm.detail.Result.Dimensions).toHaveLength(13); expect(w.vm.detail.Result.Modules).toHaveLength(4)
    expect(w.vm.formatScore('425/8')).toBe('53.13'); expect(w.vm.formatScore('1.005000')).toBe('1.01'); expect(w.vm.formatScore(0)).toBe('0.00'); expect(w.vm.formatScore(null)).toBe('—'); w.destroy()
  })
  it('late metadata after same-exam refresh cannot repopulate new state', async () => {
    const d = defer(); const w = mount(); await flush(); api.fetchManagementTraitsReissues.mockReturnValue(d.promise)
    const first = w.vm.loadReports(row()); await w.vm.loadResults(); d.resolve({ data: [report()] }); await first
    expect(w.vm.reportStatus(row())).toBe('未查询'); w.destroy()
  })
  it('cross-exam late generate cannot query or display old PDF state', async () => {
    const d = defer(); api.generateManagementTraitsReissue.mockReturnValue(d.promise); const w = mount(); await flush()
    const first = w.vm.generateReport(row()); await flush(); w.vm.$route.params.examId = 'exam-2'; d.resolve({ data: { report: report(), reused: false, purpose: 'TEST' } }); await first
    expect(api.fetchManagementTraitsReissues).not.toHaveBeenCalled(); expect(w.vm.$message.success).not.toHaveBeenCalled(); w.destroy()
  })
  it('shows assessment title and real score; no handwritten ID or eager metadata N+1', async () => {
    const w = mount(); await flush(); expect(w.vm.examTitle).toBe('Synthetic assessment')
    expect(w.text()).not.toContain('已知报告 ID'); expect(w.vm.formatScore(0)).toBe('0.00'); expect(w.vm.formatScore(null)).toBe('—')
    expect(api.fetchManagementTraitsReissues).not.toHaveBeenCalled(); expect(api.fetchManagementTraitsReissueQualification).not.toHaveBeenCalled(); w.destroy()
  })
  it('local status filter/reset/pagination never invents server search fields', async () => {
    api.fetchManagementTraitsResults.mockResolvedValue({ data: Array.from({ length: 35 }, (_, i) => row({ id: `run-${i}`, status: i === 0 ? 'incomplete' : 'completed' })) })
    const w = mount(); await flush(); w.vm.query.status = 'incomplete'; w.vm.handleQuery(); expect(w.vm.pageRows).toHaveLength(1)
    w.vm.resetQuery(); expect(w.vm.pageRows).toHaveLength(20); w.vm.changePageSize(10); w.vm.page = 2; expect(w.vm.pageRows[0].id).toBe('run-10')
    expect(api.fetchManagementTraitsResults.mock.calls).toEqual([['exam-1']]); w.destroy()
  })
  it.each([row({ status: 'incomplete' }), row({ answeredQuestionCount: 139 }), row({ examId: 'other' })])('no generation for incomplete/wrong scope %j', async target => {
    const w = mount(); await flush(); await w.vm.generateReport(target); expect(api.generateManagementTraitsReissue).not.toHaveBeenCalled(); w.destroy()
  })
  it('explicit generation qualifies then uses nested new DTO and reloads that paper only', async () => {
    const w = mount(); await flush(); await w.vm.generateReport(row())
    expect(api.fetchManagementTraitsReissueQualification).toHaveBeenCalledWith('run-1'); expect(api.generateManagementTraitsReissue).toHaveBeenCalledWith('run-1')
    expect(api.fetchManagementTraitsReissues).toHaveBeenCalledWith('paper-1'); expect(api.generateManagementTraitsTestReport).not.toHaveBeenCalled(); w.destroy()
  })
  it.each([404, 409, 503])('unavailable metadata %s preserves loaded result and never old-fallbacks', async status => {
    api.fetchManagementTraitsReissues.mockRejectedValue(Object.assign(new Error('unavailable'), { status })); const w = mount(); await flush(); await w.vm.loadReports(row())
    expect(w.vm.ready).toBe(true); expect(w.vm.rows).toHaveLength(1); expect(w.vm.reportState(row()).error).toContain('报告服务')
    expect(api.generateManagementTraitsTestReport).not.toHaveBeenCalled(); w.destroy()
  })
  it('metadata cannot mix runs or papers and unknown is not ungenerated', async () => {
    const w = mount(); await flush(); expect(w.vm.reportStatus(row())).toBe('未查询')
    api.fetchManagementTraitsReissues.mockResolvedValue({ data: [report({ runId: 'other' })] }); await w.vm.loadReports(row())
    expect(w.vm.reportState(row()).reports).toEqual([]); expect(w.vm.reportState(row()).error).not.toBe(''); w.destroy()
  })
  it('views/downloads only selected row metadata, never input ID or old TEST', async () => {
    const w = mount(); await flush(); await w.vm.openReport(row()); expect(api.viewManagementTraitsReissue).toHaveBeenCalledWith('report-1')
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {}); await w.vm.downloadReport(row())
    expect(api.downloadManagementTraitsReissue).toHaveBeenCalledWith('report-1'); expect(click).toHaveBeenCalledTimes(1)
    expect(api.viewManagementTraitsTestReport).not.toHaveBeenCalled(); w.destroy(); click.mockRestore()
  })
  it('does not save a counterfeit PDF Blob', async () => {
    api.downloadManagementTraitsReissue.mockResolvedValue({ blob: new Blob(['{"msg":"unavailable"}'], { type: 'application/pdf' }), filename: 'fake.pdf' })
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {}); const w = mount(); await flush(); await w.vm.downloadReport(row())
    expect(click).not.toHaveBeenCalled(); expect(w.vm.$message.error).toHaveBeenCalled(); w.destroy(); click.mockRestore()
  })
  it('late results never overwrite a changed exam', async () => {
    const d = defer(); api.fetchManagementTraitsResults.mockReturnValue(d.promise); const w = mount(); await flush(); w.vm.$route.params.examId = 'exam-2'; d.resolve({ data: [row()] }); await flush()
    expect(w.vm.rows).toEqual([]); expect(w.vm.ready).toBe(false); w.destroy()
  })
})

// MT-ADMIN-REVIEW-3: actual compiled SFC, deferred responses, no source-regex oracle.
describe('MT-ADMIN-REVIEW-3 metadata admission', () => {
  it.each([
    {}, { id: 'exam-1' }, exam({ assessmentType: undefined }), exam({ scoringMode: undefined }),
    exam({ repoCode: undefined }), exam({ repoCode: 'unknown' }), exam({ id: 'other', repoCode: '00101', managementTraitsProfileFrozen: false }),
    exam({ repoCode: '00101', managementTraitsProfileFrozen: true }),
    exam({ repoCode: '00301', managementTraitsProfileFrozen: false, isManagementTraits: true }),
    exam({ repoCode: '00101', managementTraitsProfileFrozen: false, managementTraitsLifecycle: 'draft' }),
    exam({ managementTraitsProfileFrozen: false, isManagementTraits: true, managementTraitsLifecycle: 'legacy' }),
    exam({ managementTraitsProfileFrozen: false, isManagementTraits: false, managementTraitsLifecycle: 'draft' }),
    exam({ managementTraitsProfileFrozen: false, isManagementTraits: false, managementTraitsLifecycle: 'frozen' }),
    exam({ managementTraitsProfileFrozen: false, managementTraitsLifecycle: 'unknown' }),
    exam({ managementTraitsProfileFrozen: true, managementTraitsLifecycle: 'legacy', isManagementTraits: false }),
    exam({ repoCode: '00101', repoList: [{ repoCode: '00201' }], managementTraitsProfileFrozen: false })
  ])('unknown/crossed/conflicting server DTO never reaches any old branch: %j', async data => {
    fetchDetail.mockResolvedValue({ data }); const w = mount(true); await flush()
    expect(legacyList).not.toHaveBeenCalled(); expect(w.vm.$router.replace).not.toHaveBeenCalled()
    expect(w.vm.entryReady).toBe(false); expect(w.vm.entryError).toContain('重试'); w.destroy()
  })
  it.each(['00101', '00302'])('existing complete non002 classifier %s needs no new lifecycle fields', async repoCode => {
    fetchDetail.mockResolvedValue({ data: { id: 'exam-1', assessmentType: 'legacy', scoringMode: 'legacy', repoList: [{ repoCode }] } })
    const w = mount(true); await flush(); expect(legacyList).toHaveBeenCalledTimes(1); expect(w.vm.entryReady).toBe(true); w.destroy()
  })
  it('current legacy002 can use repoList repoCode and retry unknown metadata', async () => {
    fetchDetail.mockResolvedValueOnce({ data: {} }); const w = mount(true); await flush(); expect(legacyList).not.toHaveBeenCalled()
    fetchDetail.mockResolvedValue({ data: exam({ repoCode: '', repoList: [{ repoCode: '00202' }], managementTraitsProfileFrozen: false, managementTraitsLifecycle: 'legacy', isManagementTraits: false }) })
    await w.vm.retryEntry(); expect(legacyList).toHaveBeenCalledTimes(1); w.destroy()
  })
  it('same-exam new draft goes to existing editor, never old list', async () => {
    fetchDetail.mockResolvedValue({ data: exam({ managementTraitsProfileFrozen: false, managementTraitsLifecycle: 'draft', isManagementTraits: true }) })
    const w = mount(true); await flush(); expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'UpdateExam', params: { id: 'exam-1' } }); expect(legacyList).not.toHaveBeenCalled(); w.destroy()
  })
  it('direct new result URL with contradictory lifecycle displays error, not empty results', async () => {
    fetchDetail.mockResolvedValue({ data: exam({ isManagementTraits: false, managementTraitsLifecycle: 'legacy' }) })
    const w = mount(); await flush(); expect(w.vm.error).not.toBe(''); expect(api.fetchManagementTraitsResults).not.toHaveBeenCalled(); w.destroy()
  })
})

// MT-ADMIN-STAGING-BOOL: docs/regression-tests.md; safe receipt, no live requests.
describe('MT-ADMIN-STAGING-BOOL deployed contract compatibility', () => {
  // Frozen routing-only projection of frontend-compatibility.json (2026-10-08T04:35:44Z).
  // Keep reproducible without ignored receipts; no identity/answer/title payload.
  const receipt = { cases: [
    { examId: '1776822816300709851', assessmentType: 'legacy', scoringMode: 'legacy', repoCode: '00201', frozen: false, lifecycleProvided: false, newModeProvided: false },
    { examId: '1791298091700970647', assessmentType: 'legacy', scoringMode: 'legacy', repoCode: '00201', frozen: true, lifecycleProvided: false, newModeProvided: false }
  ] }
  const oldContract = extra => ({ id: 'exam-1', assessmentType: 'legacy', scoringMode: 'legacy', repoCode: '00201', managementTraitsProfileFrozen: false, ...extra })
  it.each(receipt.cases)('actual safe DTO $examId preserves its server-selected branch', async source => {
    const data = { id: source.examId, assessmentType: source.assessmentType, scoringMode: source.scoringMode, repoCode: source.repoCode, managementTraitsProfileFrozen: source.frozen }
    expect(source.lifecycleProvided).toBe(false); expect(source.newModeProvided).toBe(false)
    if (source.frozen) data.requiredFields = 'name,gender,telephone'
    fetchDetail.mockResolvedValue({ data }); getInfo.mockResolvedValue({ code: 200, user: { userId: source.frozen ? 1 : 2 }, permissions: [] })
    api.fetchManagementTraitsProfile.mockResolvedValue({ data: { examId: source.examId, frozenAt: '2026-10-08T00:00:00Z' } })
    const w = mount(true, { examId: source.examId }); await flush()
    expect(w.vm.entryError).toBe(''); expect(legacyList).toHaveBeenCalledTimes(source.frozen ? 0 : 1)
    if (source.frozen) {
      expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ManagementTraitsResults', params: { examId: source.examId } })
      expect(getInfo).toHaveBeenCalledTimes(1); expect(api.fetchManagementTraitsProfile).toHaveBeenCalledTimes(1)
    } else {
      expect(legacyList.mock.calls[0][0].examId).toBe(source.examId); expect(w.vm.entryReady).toBe(true)
      expect(getInfo).not.toHaveBeenCalled(); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); expect(w.vm.$router.replace).not.toHaveBeenCalled()
    }
    w.destroy()
  })
  it.each([{}, { isManagementTraits: false }, { managementTraitsLifecycle: 'legacy' }, { isManagementTraits: false, managementTraitsLifecycle: 'legacy' }])('strict false accepts absent or explicit consistent newer markers %j', async markers => {
    fetchDetail.mockResolvedValue({ data: oldContract(markers) }); const w = mount(true, {}, { mngTest: '1' }); await flush()
    expect(legacyList).toHaveBeenCalledTimes(1); expect(w.vm.entryError).toBe(''); expect(getInfo).not.toHaveBeenCalled(); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); expect(w.vm.$router.replace).not.toHaveBeenCalled(); w.destroy()
  })
  it.each([
    { managementTraitsProfileFrozen: undefined }, { managementTraitsProfileFrozen: null }, { managementTraitsProfileFrozen: 'false' }, { managementTraitsProfileFrozen: 0 },
    { id: 'other' }, { id: true }, { id: {} }, { id: [] }, { id: 0 }, { id: 1.5 }, { id: Number.MAX_SAFE_INTEGER + 1 },
    { assessmentType: undefined }, { scoringMode: undefined }, { scoringMode: 'competency_average' }, { repoCode: 'unknown' },
    { isManagementTraits: true }, { isManagementTraits: true, managementTraitsLifecycle: 'unknown' },
    { isManagementTraits: false, managementTraitsLifecycle: 'draft' }, { managementTraitsLifecycle: 'draft' }, { managementTraitsLifecycle: 'frozen' },
    { managementTraitsLifecycle: 'unknown' }, { managementTraitsLifecycle: null }, { managementTraitsLifecycle: undefined },
    { isManagementTraits: null }, { isManagementTraits: 'false' }, { isManagementTraits: undefined }
  ])('unknown/present-invalid/conflicting contract remains closed %j', async extra => {
    fetchDetail.mockResolvedValue({ data: oldContract(extra) }); const w = mount(true); await flush()
    expect(w.vm.entryReady).toBe(false); expect(w.vm.entryError).not.toBe(''); expect(legacyList).not.toHaveBeenCalled(); expect(w.vm.$router.replace).not.toHaveBeenCalled(); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it('same textual ID cannot hide an invalid server ID type', async () => {
    fetchDetail.mockResolvedValue({ data: oldContract({ id: true }) }); const w = mount(true, { examId: 'true' }); await flush()
    expect(legacyList).not.toHaveBeenCalled(); expect(w.vm.entryError).not.toBe(''); w.destroy()
  })
  it('positive safe numeric server ID matches the current textual scope', async () => {
    fetchDetail.mockResolvedValue({ data: oldContract({ id: 12 }) }); const w = mount(true, { examId: '12' }); await flush()
    expect(legacyList).toHaveBeenCalledTimes(1); expect(legacyList.mock.calls[0][0].examId).toBe('12'); w.destroy()
  })
  it('bool-only old response arriving after another exam cannot enable old list', async () => {
    const d = defer(); fetchDetail.mockReturnValue(d.promise); const w = mount(true); w.vm.$route.params.examId = 'exam-2'
    d.resolve({ data: oldContract() }); await flush(); expect(legacyList).not.toHaveBeenCalled(); expect(w.vm.entryReady).toBe(false); w.destroy()
  })
})

describe('MT-ADMIN-REVIEW-3 report request ownership', () => {
  it.each(['resolve', 'reject'])('generate invalidates pending old metadata %s and selects the exact returned report', async completion => {
    const w = mount(); await flush(); const target = w.vm.rows[0], old = defer(), fresh = defer()
    api.fetchManagementTraitsReissues.mockReturnValueOnce(old.promise).mockReturnValueOnce(fresh.promise)
    const pending = w.vm.loadReports(target), generated = w.vm.generateReport(target); await flush()
    expect(api.fetchManagementTraitsReissues).toHaveBeenCalledTimes(2)
    if (completion === 'resolve') old.resolve({ data: [] }); else old.reject(Object.assign(new Error('old failure'), { status: 503 }))
    await pending; expect(w.vm.reportState(target).loading).toBe(true); expect(w.vm.reportState(target).error).toBe('')
    fresh.resolve({ data: [report({ id: 'other-version', createdAt: '2026-10-08T01:00:00Z' }), report()] }); await generated
    expect(w.vm.reportState(target).reports).toHaveLength(2); expect(w.vm.reportState(target).selectedId).toBe('report-1'); expect(w.vm.reportState(target).loading).toBe(false)
    expect(w.vm.$message.success).toHaveBeenCalledTimes(1); w.destroy()
  })
  it('view joins the pending metadata promise instead of falsely reporting no report', async () => {
    const w = mount(); await flush(); const target = w.vm.rows[0], d = defer(); api.fetchManagementTraitsReissues.mockReturnValue(d.promise)
    const pending = w.vm.loadReports(target), reading = w.vm.openReport(target); await flush(); expect(api.viewManagementTraitsReissue).not.toHaveBeenCalled()
    d.resolve({ data: [report()] }); await Promise.all([pending, reading]); expect(api.fetchManagementTraitsReissues).toHaveBeenCalledTimes(1); expect(api.viewManagementTraitsReissue).toHaveBeenCalledWith('report-1'); w.destroy()
  })
  it('successful generation plus failed refresh remains known generated and retry selects exact ID', async () => {
    const w = mount(); await flush(); const target = w.vm.rows[0]
    api.fetchManagementTraitsReissues.mockRejectedValueOnce(Object.assign(new Error('closed'), { status: 503 })).mockResolvedValueOnce({ data: [report({ id: 'latest' }), report()] })
    await w.vm.generateReport(target); expect(w.vm.reportStatus(target)).toContain('已生成'); expect(w.vm.reportState(target).loaded).toBe(false)
    expect(w.vm.$message.warning).toHaveBeenCalledWith(expect.stringContaining('刷新失败')); expect(w.vm.rows).toHaveLength(1)
    await w.vm.loadReports(target); expect(w.vm.reportState(target).selectedId).toBe('report-1'); w.destroy()
  })
  it('fresh metadata omitting known generated ID cannot declare ungenerated', async () => {
    const w = mount(); await flush(); const target = w.vm.rows[0]; api.fetchManagementTraitsReissues.mockResolvedValue({ data: [] })
    await w.vm.generateReport(target); expect(w.vm.reportStatus(target)).toContain('已生成'); expect(w.vm.reportState(target).error).not.toBe(''); w.destroy()
  })
  it.each(['resolve', 'reject'])('replaced row reference with same IDs invalidates metadata %s and finally', async completion => {
    const w = mount(); await flush(); const target = w.vm.rows[0], d = defer(); api.fetchManagementTraitsReissues.mockReturnValue(d.promise)
    const pending = w.vm.loadReports(target); w.vm.rows = [row()]
    if (completion === 'resolve') d.resolve({ data: [report()] }); else d.reject(new Error('stale'))
    await pending; expect(w.vm.reportState(w.vm.rows[0]).reports).toEqual([]); expect(w.vm.reportState(w.vm.rows[0]).error).toBe(''); w.destroy()
  })
  it('destroyed component discards late report callbacks', async () => {
    const w = mount(); await flush(); const target = w.vm.rows[0], d = defer(); api.fetchManagementTraitsReissues.mockReturnValue(d.promise)
    const pending = w.vm.loadReports(target); w.destroy(); d.resolve({ data: [report()] }); await pending; expect(w.vm.reportState(target).reports).toEqual([])
  })
  it.each(['resolve', 'reject'])('fresh generated metadata survives old response arriving last: %s', async completion => {
    const w = mount(); await flush(); const target = w.vm.rows[0], old = defer(); api.fetchManagementTraitsReissues.mockReturnValueOnce(old.promise).mockResolvedValueOnce({ data: [report()] })
    const pending = w.vm.loadReports(target); await w.vm.generateReport(target)
    if (completion === 'resolve') old.resolve({ data: [] }); else old.reject(new Error('late failure'))
    await pending; expect(w.vm.reportState(target).selectedId).toBe('report-1'); expect(w.vm.reportState(target).reports).toEqual([report()]); expect(w.vm.reportState(target).error).toBe(''); w.destroy()
  })
  it('replaced same-ID row starts its own request rather than joining discarded promise', async () => {
    const w = mount(); await flush(); const target = w.vm.rows[0], old = defer(); api.fetchManagementTraitsReissues.mockReturnValueOnce(old.promise).mockResolvedValueOnce({ data: [report()] })
    const pending = w.vm.loadReports(target); w.vm.rows = [row()]; await w.vm.loadReports(w.vm.rows[0]); expect(api.fetchManagementTraitsReissues).toHaveBeenCalledTimes(2)
    old.resolve({ data: [] }); await pending; expect(w.vm.reportState(w.vm.rows[0]).selectedId).toBe('report-1'); w.destroy()
  })
  it('new exam reusing run ID cannot inherit the previous exam metadata', async () => {
    const w = mount(); await flush(); const d = defer(); api.fetchManagementTraitsReissues.mockReturnValueOnce(d.promise)
    const pending = w.vm.loadReports(w.vm.rows[0]); fetchDetail.mockResolvedValue({ data: exam({ id: 'exam-2' }) })
    api.fetchManagementTraitsProfile.mockResolvedValue({ data: { examId: 'exam-2', frozenAt: '2026-10-08T00:00:00Z' } })
    api.fetchManagementTraitsResults.mockResolvedValue({ data: [row({ examId: 'exam-2', paperId: 'paper-2' })] }); w.vm.$route.params.examId = 'exam-2'; await w.vm.$nextTick(); await flush()
    d.resolve({ data: [report()] }); await pending; expect(w.vm.rows[0].examId).toBe('exam-2'); expect(w.vm.reportStatus(w.vm.rows[0])).toBe('未查询'); w.destroy()
  })
})

describe('MT-ADMIN-REVIEW-3 legacy list request ownership', () => {
  const legacyExam = extra => exam({ repoCode: '00101', managementTraitsProfileFrozen: undefined, ...extra })
  it.each(['1', '2'])('list %s binds immutable query and ignores older query response/finally', async isOpen => {
    fetchDetail.mockResolvedValue({ data: legacyExam() }); const a = defer(), b = defer(); legacyList.mockReturnValueOnce(a.promise).mockReturnValueOnce(b.promise)
    const w = mount(true); w.vm.$route.params.isOpen = isOpen; await flush(); const queryA = legacyList.mock.calls[0][0]
    w.vm.queryParams.name = 'new-query'; w.vm.queryParams.pageNum = 2; const pending = w.vm.getList(isOpen)
    expect(queryA.name).toBeUndefined(); expect(queryA.pageNum).toBe(1)
    a.resolve({ rows: [{ id: 'old' }], total: 10 }); await flush(); expect(w.vm.testerList).toEqual([]); expect(w.vm.loading).toBe(true)
    b.resolve({ rows: [{ id: 'new' }], total: 1 }); await pending; await flush(); expect(w.vm.testerList).toEqual([{ id: 'new' }]); expect(w.vm.total).toBe(1); w.destroy()
  })
  it('mounted reactive route watcher prevents examA list replacing examB', async () => {
    fetchDetail.mockResolvedValueOnce({ data: legacyExam() }).mockResolvedValue({ data: legacyExam({ id: 'exam-2' }) })
    const a = defer(), b = defer(); legacyList.mockReturnValueOnce(a.promise).mockReturnValueOnce(b.promise); const w = mount(true); await flush()
    w.vm.$route.params.examId = 'exam-2'; await w.vm.$nextTick(); await flush(); expect(legacyList).toHaveBeenCalledTimes(2)
    expect(legacyList.mock.calls[0][0].examId).toBe('exam-1'); expect(legacyList.mock.calls[1][0].examId).toBe('exam-2')
    b.resolve({ rows: [{ id: 'examB' }], total: 1 }); await flush(); a.resolve({ rows: [{ id: 'examA' }], total: 9 }); await flush()
    expect(w.vm.testerList).toEqual([{ id: 'examB' }]); expect(w.vm.total).toBe(1); w.destroy()
  })
  it('route changes to frozen002 invalidate an old list before new entry resolves', async () => {
    const list = defer(), detail = defer(); fetchDetail.mockResolvedValueOnce({ data: legacyExam() }).mockReturnValueOnce(detail.promise); legacyList.mockReturnValue(list.promise)
    const w = mount(true); await flush(); w.vm.$route.params.examId = 'exam-2'; await w.vm.$nextTick(); await flush()
    list.resolve({ rows: [{ id: 'stale' }], total: 1 }); await flush(); expect(w.vm.testerList).toEqual([]); expect(w.vm.entryReady).toBe(false)
    api.fetchManagementTraitsProfile.mockResolvedValue({ data: { examId: 'exam-2', frozenAt: '2026-10-08T00:00:00Z' } }); detail.resolve({ data: exam({ id: 'exam-2' }) }); await flush()
    expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ManagementTraitsResults', params: { examId: 'exam-2' } }); expect(legacyList).toHaveBeenCalledTimes(1); w.destroy()
  })
  it('destroyed component ignores the legacy list response', async () => {
    fetchDetail.mockResolvedValue({ data: legacyExam() }); const d = defer(); legacyList.mockReturnValue(d.promise); const w = mount(true); await flush(); w.destroy()
    d.resolve({ rows: [{ id: 'stale' }], total: 1 }); await flush(); expect(w.vm.testerList).toEqual([])
  })
  it('late list rejection is caught and cannot end a newer loading request', async () => {
    fetchDetail.mockResolvedValue({ data: legacyExam() }); const a = defer(), b = defer(); legacyList.mockReturnValueOnce(a.promise).mockReturnValueOnce(b.promise)
    const w = mount(true); await flush(); const pending = w.vm.getList('1'); a.reject(new Error('old list failed')); await flush()
    expect(w.vm.loading).toBe(true); expect(w.vm.$message.error).not.toHaveBeenCalled(); b.resolve({ rows: [], total: 0 }); await pending; expect(w.vm.loading).toBe(false); w.destroy()
  })
})
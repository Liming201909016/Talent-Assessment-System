import { beforeEach, describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import compiler from 'vue-template-compiler'
import babel from '@babel/core'
import fs from 'fs'
import path from 'path'
import * as product from '@/utils/managementTraitsProduct'

const api = {
  fetchManagementTraitsExamConfig: (...args) => examApi.fetchDetail(...args),
  managementTraitsExamKnown: () => true, rememberManagementTraitsProfile: vi.fn(),
  fetchManagementTraitsProfile: vi.fn(), freezeManagementTraitsProfile: vi.fn(),
  setManagementTraitsExamState: vi.fn(),
  fetchManagementTraitsResults: vi.fn(), fetchManagementTraitsResult: vi.fn(),
  generateManagementTraitsTestReport: vi.fn(), viewManagementTraitsTestReport: vi.fn(),
  downloadManagementTraitsTestReport: vi.fn(),
  fetchManagementTraitsAdminAccess: vi.fn(), fetchManagementTraitsReissueQualification: vi.fn(), fetchManagementTraitsReissues: vi.fn(),
  generateManagementTraitsReissue: vi.fn(), viewManagementTraitsReissue: vi.fn(), downloadManagementTraitsReissue: vi.fn(),
  canManageManagementTraits: store => store?.state?.user?.userId === 1 || (store?.getters?.permissions || []).includes('*:*:*')
}
const examApi = { fetchDetail: vi.fn(), saveData: vi.fn() }
const repoApi = { fetchList: vi.fn(() => Promise.resolve({ data: [] })) }
const messages = () => ({ error: vi.fn(), warning: vi.fn(), success: vi.fn(), info: vi.fn() })
const deferred = () => { let resolve; let reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
const flush = async () => { for (let i = 0; i < 12; i++) await Promise.resolve() }

// Compile the actual SFC (including its template), injecting only external API
// dependencies. This also runs before the separately owned API module exists.
function component(file) {
  const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views', file), 'utf8')
  const sfc = compiler.parseComponent(source)
  const code = babel.transformSync(sfc.script.content, {
    babelrc: false, configFile: false, plugins: ['@babel/plugin-transform-modules-commonjs']
  }).code
  const module = { exports: {} }
  const requireMock = name => {
    if (name === '@/utils/managementTraitsProduct') return product
    if (name === '@/api/managementTraits') return api
    if (name === '@/api/exam/exam') return examApi
    if (name === '@/api/qu/repo') return repoApi
    if (name === '@/api/competency') return { fetchCompetencyDimensions: vi.fn(() => Promise.resolve({ data: [] })) }
    if (name === '@/utils/request') return { __esModule: true, default: { post: vi.fn(() => Promise.resolve({ data: {} })), get: vi.fn(() => Promise.resolve({})) } }
    if (name.endsWith('.vue') || name.startsWith('@/components/')) return { __esModule: true, default: { render: h => h('div') } }
    return {}
  }
  new Function('require', 'module', 'exports', code)(requireMock, module, module.exports)
  return { ...module.exports.default, ...compiler.compileToFunctions(sfc.template.content) }
}
function mount(file = 'exam/exam/managementTraitsResults.vue', options = {}) {
  const store = options.store || { state: { user: { userId: 1 }, app: { device: 'desktop' } }, getters: { permissions: [] } }
  if (options.store) api.fetchManagementTraitsAdminAccess.mockResolvedValue({ user: store.state.user, permissions: store.getters.permissions })
  return shallowMount(component(file), {
    ...options, directives: { loading: () => {} },
    mocks: { $store: store, $route: { params: { examId: 'exam-1' } },
      $router: { push: vi.fn(), replace: vi.fn(), back: vi.fn() }, $message: messages(),
      $notify: vi.fn(), $confirm: vi.fn(() => Promise.resolve()), parseTime: value => value || '—', ...options.mocks },
    stubs: ['el-card', 'el-alert', 'el-checkbox', 'el-checkbox-group', 'el-radio', 'el-radio-group', 'el-switch', 'el-input-number', 'el-date-picker', 'el-row', 'el-col', 'el-dropdown', 'el-dropdown-menu', 'el-dropdown-item', 'pagination', 'el-descriptions', 'el-descriptions-item', 'el-empty', ...(options.stubs || [])]
  })
}
const row = (overrides = {}) => ({ id: 'run-1', examId: 'exam-1', paperId: 'paper-1', status: 'completed', totalQuestionCount: 140, answeredQuestionCount: 140, userTimeSeconds: 900, submittedAt: '2026-10-03T01:00:00Z', ...overrides })
const frozen = { examId: 'exam-1', frozenAt: '2026-10-03T01:00:00Z' }
const reissue = { id: 'report-1', runId: 'run-1', paperId: 'paper-1', examId: 'exam-1', kind: 'verified_frozen_result', status: 'completed' }
const repo = { repoId: 'repo-1', repoCode: '00201', radioCount: 140, radioScore: 5, multiCount: 0, multiScore: 0, judgeCount: 0, judgeScore: 0, saqCount: 0 }
const newRepo = { ...repo, repoId: 'synthetic-00501-repo', repoCode: '00501' }
beforeEach(() => {
  Object.values(api).forEach(fn => { if (fn.mockReset) fn.mockReset() })
  api.fetchManagementTraitsProfile.mockResolvedValue({ data: frozen })
  api.fetchManagementTraitsAdminAccess.mockResolvedValue({ user: { userId: 1 }, permissions: [] })
  api.fetchManagementTraitsReissueQualification.mockResolvedValue({ data: { eligible: true, kind: 'verified_frozen_result', purpose: '仅供系统测试，不可作为人才决策依据' } })
  api.fetchManagementTraitsReissues.mockResolvedValue({ data: [reissue] })
  api.generateManagementTraitsReissue.mockResolvedValue({ data: { report: reissue, reused: false, purpose: '仅供系统测试，不可作为人才决策依据' } })
  api.viewManagementTraitsReissue.mockResolvedValue({ blob: new Blob(['%PDF-1.7'], { type: 'application/pdf' }), filename: 'test.pdf' })
  api.downloadManagementTraitsReissue.mockResolvedValue({ blob: new Blob(['%PDF-1.7'], { type: 'application/pdf' }), filename: 'test.pdf' })
  api.fetchManagementTraitsResults.mockResolvedValue({ data: [row()] })
  api.freezeManagementTraitsProfile.mockResolvedValue({ data: frozen })
  api.setManagementTraitsExamState.mockResolvedValue({ data: { examId: 'exam-1', state: 0 } })
  api.generateManagementTraitsTestReport.mockResolvedValue({ data: { id: 'report-1', runId: 'run-1', examId: 'exam-1', mode: 'test', testTitle: 'TEST', testLabel: '不可作为人才决策依据' } })
  api.viewManagementTraitsTestReport.mockResolvedValue(new Blob(['%PDF-1.7'], { type: 'application/pdf' }))
  api.downloadManagementTraitsTestReport.mockResolvedValue(new Blob(['%PDF-1.7'], { type: 'application/pdf' }))
  examApi.fetchDetail.mockReset().mockResolvedValue({ data: { id: 'exam-1', managementTraitsProfileFrozen: true, assessmentType: 'legacy', scoringMode: 'legacy', repoCode: '00201', repoList: [{ ...repo }], requiredFields: 'age,degree', joinType: 1 } })
  examApi.saveData.mockReset().mockResolvedValue({ data: { id: 'exam-1', isManagementTraits: true, managementTraitsLifecycle: 'draft', managementTraitsProfileFrozen: false } })
})

describe('002 TEST admin result components', () => {
  it('loads the explicit exam and uses 20-row local pagination with a 200 cap notice', async () => {
    api.fetchManagementTraitsResults.mockResolvedValue({ data: Array.from({ length: 45 }, (_, i) => row({ id: `run-${i}` })) })
    const w = mount(); await flush()
    expect(api.fetchManagementTraitsProfile).toHaveBeenCalledWith('exam-1')
    expect(api.fetchManagementTraitsResults).toHaveBeenCalledWith('exam-1')
    expect(w.vm.pageRows).toHaveLength(20)
    expect(w.text()).toContain('200')
    await w.setData({ page: 2, pageSize: 10 }); expect(w.vm.pageRows[0].id).toBe('run-10')
  })
  it.each([null, undefined])('fails closed without a frozen profile (%s)', async profile => {
    api.fetchManagementTraitsProfile.mockResolvedValue({ data: profile })
    const w = mount(); await flush(); expect(api.fetchManagementTraitsResults).not.toHaveBeenCalled(); expect(w.vm.error).not.toBe('')
  })
  it.each([401, 403, 404, 503])('shows profile failure %s without legacy or result fallback', async status => {
    api.fetchManagementTraitsProfile.mockRejectedValue(new Error(`HTTP ${status}`))
    const w = mount(); await flush(); expect(w.vm.error).toContain(`${status}`); expect(api.fetchManagementTraitsResults).not.toHaveBeenCalled(); expect(w.vm.loading).toBe(false)
  })
  it.each([{ userId: 2, permissions: [], roles: ['admin'] }, { userId: '1', permissions: [] }])('rejects role-only/string-ID privilege', async user => {
    const w = mount(undefined, { store: { state: { user }, getters: { permissions: user.permissions } } }); await flush()
    expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled()
    await w.vm.generateReport(row()); await w.vm.openReport(); expect(api.generateManagementTraitsTestReport).not.toHaveBeenCalled(); expect(api.viewManagementTraitsTestReport).not.toHaveBeenCalled()
  })
  it('allows wildcard privilege without an admin role', async () => {
    mount(undefined, { store: { state: { user: { userId: 2 } }, getters: { permissions: ['*:*:*'] } } }); await flush(); expect(api.fetchManagementTraitsResults).toHaveBeenCalled()
  })
  it.each([row({ status: 'incomplete' }), row({ answeredQuestionCount: 139 }), row({ totalQuestionCount: 141 }), row({ examId: 'other' })])('rejects incomplete/mismatched generation %j', async target => {
    const w = mount(); await flush(); await w.vm.generateReport(target); expect(api.generateManagementTraitsTestReport).not.toHaveBeenCalled()
  })
  it('generates only explicitly, records the real report ID and keeps TEST visible', async () => {
    const w = mount(); await flush(); expect(api.generateManagementTraitsReissue).not.toHaveBeenCalled()
    await w.vm.generateReport(row()); expect(api.generateManagementTraitsReissue).toHaveBeenCalledWith('run-1'); expect(w.vm.reportState(row()).selectedId).toBe('report-1'); expect(w.find('el-alert-stub').attributes('title')).toContain('不可作为人才决策依据')
  })
  it('prevents duplicate generation and restores loading after errors', async () => {
    const pending = deferred(); api.generateManagementTraitsReissue.mockReturnValue(pending.promise)
    const w = mount(); await flush(); const first = w.vm.generateReport(row()); await flush(); await w.vm.generateReport(row())
    expect(api.generateManagementTraitsReissue).toHaveBeenCalledTimes(1); expect(w.vm.generating).toBe(true)
    pending.reject(new Error('报告端点不可用')); await first; expect(w.vm.generating).toBe(false); expect(w.vm.$message.error).toHaveBeenCalledWith(expect.stringContaining('报告端点不可用'))
  })
  it('does not generate when confirmation is cancelled', async () => {
    const w = mount(); await flush(); w.vm.$confirm.mockRejectedValue('cancel'); await w.vm.generateReport(row()); expect(api.generateManagementTraitsReissue).not.toHaveBeenCalled()
  })
  it.each([{ id: '', mode: 'test' }, { id: 'bad', mode: 'formal' }, { id: 'bad', mode: 'test', runId: 'other' }])('rejects invalid report metadata %j', async report => {
    api.generateManagementTraitsReissue.mockResolvedValue({ data: { report, reused: false, purpose: '仅供系统测试，不可作为人才决策依据' } }); const w = mount(); await flush(); await w.vm.generateReport(row()); expect(w.vm.reportState(row()).selectedId).toBe(''); expect(w.vm.$message.error).toHaveBeenCalled()
  })
  it('reads uppercase detail and displays big.Rat JSON strings exactly, not as guessed percentages', async () => {
    api.fetchManagementTraitsResult.mockResolvedValue({ data: { RunID: 'run-1', PaperID: 'paper-1', ExamID: 'exam-1', Result: { IsComplete: true, OverallScore: '705/13', Dimensions: [{ Key: 'confidence', Name: '自信', Score: '375/8' }], Modules: [] } } })
    const w = mount(); await flush(); await w.vm.showDetail(row()); expect(w.vm.detail.RunID).toBe('run-1'); expect(w.vm.formatRat('705/13')).toBe('705/13'); expect(w.vm.formatRat(null)).toBe('—'); expect(w.vm.formatRat({ numerator: 705 })).toBe('—')
  })
  it('rejects cross-run detail and does not retain stale detail on error', async () => {
    api.fetchManagementTraitsResult.mockResolvedValue({ data: { RunID: 'other', ExamID: 'exam-1', Result: {} } }); const w = mount(); await flush(); await w.vm.showDetail(row()); expect(w.vm.detail).toBe(null); expect(w.vm.detailLoading).toBe(false)
  })
  it('keeps list loading stable and suppresses duplicate refresh requests', async () => {
    const pending = deferred(); api.fetchManagementTraitsResults.mockReturnValue(pending.promise); const w = mount(); await flush(); await w.vm.loadResults(); expect(api.fetchManagementTraitsResults).toHaveBeenCalledTimes(1)
    pending.reject(new Error('列表503')); await flush(); expect(w.vm.loading).toBe(false); expect(w.vm.rows).toEqual([]); expect(w.vm.error).toContain('503')
  })
  it('views an explicit existing report ID and revokes the URL on close/destroy', async () => {
    const create = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:mt'); const revoke = vi.spyOn(URL, 'revokeObjectURL')
    const w = mount(); await flush(); await w.vm.openReport(row()); expect(api.viewManagementTraitsReissue).toHaveBeenCalledWith('report-1'); expect(w.vm.pdfUrl).toBe('blob:mt'); w.vm.closePdf(); expect(revoke).toHaveBeenCalledWith('blob:mt'); w.destroy(); create.mockRestore(); revoke.mockRestore()
  })
  it('downloads a valid Blob by ID, removes its anchor and revokes its URL', async () => {
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {}); const revoke = vi.spyOn(URL, 'revokeObjectURL')
    const w = mount(); await flush(); await w.vm.downloadReport(row()); expect(api.downloadManagementTraitsReissue).toHaveBeenCalledWith('report-1'); expect(click).toHaveBeenCalledTimes(1); expect(document.querySelector('a[download]')).toBe(null); w.destroy(); click.mockRestore(); revoke.mockRestore()
  })
  it.each(['openReport', 'downloadReport'])('rejects missing ID and PDF/API failure in %s', async method => {
    const w = mount(); await flush(); await w.vm[method](); expect(api.viewManagementTraitsTestReport).not.toHaveBeenCalled(); expect(api.downloadManagementTraitsTestReport).not.toHaveBeenCalled()
    api.viewManagementTraitsReissue.mockRejectedValue(new Error('报告404')); api.downloadManagementTraitsReissue.mockRejectedValue(new Error('报告404')); await w.vm[method](row()); expect(w.vm.reportLoading).toBe(false); expect(w.vm.$message.error).toHaveBeenCalledWith(expect.stringContaining('404'))
  })
  it('rejects a non-PDF Blob without opening it', async () => {
    api.viewManagementTraitsReissue.mockResolvedValue({ blob: new Blob(['error'], { type: 'application/json' }) }); const w = mount(); await flush(); await w.vm.openReport(row()); expect(w.vm.pdfUrl).toBe(''); expect(w.vm.$message.error).toHaveBeenCalled()
  })
  it('suppresses duplicate report reads until the pending Blob resolves', async () => {
    const pending = deferred(); api.viewManagementTraitsReissue.mockReturnValue(pending.promise)
    const w = mount(); await flush(); const first = w.vm.openReport(row()); await flush(); await w.vm.openReport(row()); await w.vm.downloadReport(row())
    expect(api.viewManagementTraitsReissue).toHaveBeenCalledTimes(1); expect(api.downloadManagementTraitsReissue).not.toHaveBeenCalled(); expect(w.vm.reportLoading).toBe(true)
    pending.resolve({ blob: new Blob(['%PDF-1.7'], { type: 'application/pdf' }) }); await first; expect(w.vm.reportLoading).toBe(false); w.destroy()
  })
  it.each([Array.from({ length: 201 }, () => row()), [row({ examId: 'other' })]])('fails closed on an oversized or cross-exam list', async rows => {
    api.fetchManagementTraitsResults.mockResolvedValue({ data: rows }); const w = mount(); await flush(); expect(w.vm.rows).toEqual([]); expect(w.vm.ready).toBe(false); expect(w.vm.error).not.toBe('')
  })
})

describe('002 explicit configuration and entry routing', () => {
  it('MT-NEW-DRAFT direct legacy detail redirects draft to editable config without old participant loading', async () => {
    examApi.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', assessmentType: 'legacy', scoringMode: 'legacy', repoCode: '00201', isManagementTraits: true, managementTraitsLifecycle: 'draft', managementTraitsProfileFrozen: false } })
    const getList = vi.fn(); const w = mount('user/exam/index.vue', { methods: { getList } }); await flush()
    expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'UpdateExam', params: { id: 'exam-1' } }); expect(getList).not.toHaveBeenCalled(); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it('MT-NEW-DRAFT reload keeps draft mandatory and editable without pretending profile is frozen', async () => {
    examApi.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', assessmentType: 'legacy', scoringMode: 'legacy', repoCode: '00201', repoList: [{ ...repo }], joinType: 1, requiredFields: 'name,telephone', isManagementTraits: true, managementTraitsLifecycle: 'draft', managementTraitsProfileFrozen: false } })
    const w = mount('exam/exam/form.vue', { mocks: { $route: { params: {} } } }); await w.vm.fetchData('exam-1'); await flush()
    expect(w.vm.managementTraitsDraft).toBe(true); expect(w.vm.managementTraitsSelected).toBe(true); expect(w.vm.managementTraitsReadOnly).toBe(false); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled()
    w.vm.managementTraitsSelected = false; w.vm.handleManagementTraitsSelection(false); expect(w.vm.managementTraitsSelected).toBe(true); w.destroy()
  })
  function form() {
    const w = mount('exam/exam/form.vue', { mocks: { $route: { params: {} } } })
    w.vm.$refs.postForm.clearValidate = vi.fn()
    return w
  }
  // MT-NEW-DEFAULT: docs/business-branches.md; new creation only, historical profiles unchanged.
  const choice = (code = '00501', overrides = {}) => ({ id: `repo-${code}`, code, radioCount: 140, multiCount: 0, judgeCount: 0, saqCount: 0, ...overrides })
  async function choose(w, value) {
    w.vm.repoChange(value, w.vm.repoList[0], 0)
    w.vm.$set(w.vm.repoList[0], 'repoId', value ? value.id : '')
    await w.vm.$nextTick()
  }
  it.each(['00501', '00502'])('defaults independent new %s to TEST without saving or freezing and preserves the field subset', async code => {
    const w = form(); w.vm.requiredFieldsList = ['age', 'degree']; await choose(w, choice(code))
    expect(w.vm.managementTraitsSelected).toBe(true); expect(w.vm.postForm.totalTime).toBe(25); expect(w.vm.postForm.showPdf).toBe(false)
    expect(w.vm.postForm.assessmentType).toBe('legacy'); expect(w.vm.postForm.scoringMode).toBe('legacy'); expect(w.vm.requiredFieldsList).toEqual(['age', 'degree'])
    expect(examApi.saveData).not.toHaveBeenCalled(); expect(api.freezeManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it.each(['00201', '00202', '00101', '00301', '00401', '00203', '002010', '00503'])('does not default legacy or unsupported repository %s', async code => {
    const w = form(); await choose(w, choice(code)); expect(w.vm.managementTraitsSelected).toBe(false); w.destroy()
  })
  it.each([{ radioCount: 139 }, { radioCount: 141 }, { multiCount: 1 }, { judgeCount: 1 }, { saqCount: 1 }, { id: '' }])('does not default an ineligible repository %j', async overrides => {
    const w = form(); await choose(w, choice('00501', overrides)); expect(w.vm.managementTraitsSelected).toBe(false); w.destroy()
  })
  it('clears the new default when the repository is cleared or changed away, then defaults the other 005 code', async () => {
    const w = form(); await choose(w, choice()); expect(w.vm.managementTraitsSelected).toBe(true)
    await choose(w, null); expect(w.vm.managementTraitsSelected).toBe(false)
    await choose(w, choice('00502')); expect(w.vm.managementTraitsSelected).toBe(true)
    await choose(w, choice('00301')); expect(w.vm.managementTraitsSelected).toBe(false); w.destroy()
  })
  // MT-DEFAULT-CLEAR-P1: docs/management-traits-new-default-local-20261006.md.
  it.each(['00501', '00502'])('reselects the same new %s default after clearing without manual opt-out', async code => {
    const w = form(); const states = []
    for (const value of [choice(code), null, choice(code)]) {
      await choose(w, value); states.push(w.vm.managementTraitsSelected)
    }
    expect(states).toEqual([true, false, true])
    expect(w.vm.repoList[0].repoCode).toBe(code); expect(w.vm.repoCode).toBe(code)
    expect(examApi.saveData).not.toHaveBeenCalled(); expect(api.freezeManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it.each(['00501', '00502'])('rejects opt-out across consecutive same-code %s selections under MT-005 policy', async code => {
    const w = form(); await choose(w, choice(code)); expect(w.vm.managementTraitsSelected).toBe(true)
    w.vm.managementTraitsSelected = false; w.vm.handleManagementTraitsSelection(false)
    for (let i = 0; i < 2; i++) {
      await choose(w, choice(code)); expect(w.vm.managementTraitsSelected).toBe(true)
    }
    expect(examApi.saveData).not.toHaveBeenCalled(); expect(api.freezeManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it('restores mandatory new mode after attempted opt-out, while freeze remains separately cancellable', async () => {
    const w = form(); await choose(w, choice()); expect(w.vm.managementTraitsSelected).toBe(true)
    w.vm.managementTraitsSelected = false; w.vm.handleManagementTraitsSelection(false); await choose(w, choice())
    expect(w.vm.managementTraitsSelected).toBe(true); w.vm.$confirm.mockRejectedValue('cancel'); await w.vm.submitForm()
    expect(examApi.saveData).toHaveBeenCalledTimes(1); expect(examApi.saveData.mock.calls[0][0].managementTraitsTestOnly).toBe(true)
    expect(w.vm.managementTraitsDraft).toBe(true); expect(api.freezeManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it('does not default mixed repositories or selected non-radio counts', async () => {
    const w = form(); w.vm.repoList.push({ ...repo }); await choose(w, choice()); expect(w.vm.managementTraitsSelected).toBe(false)
    w.vm.repoList.pop(); w.vm.repoList[0].multiCount = 1; await choose(w, choice('00502')); expect(w.vm.managementTraitsSelected).toBe(false); w.destroy()
  })
  it.each(['route', 'saved'])('does not default historical editing identified by %s', async mode => {
    const w = form(); if (mode === 'route') w.vm.$route.params.id = 'exam-1'; else w.vm.$set(w.vm.postForm, 'id', 'exam-1')
    await choose(w, choice()); expect(w.vm.managementTraitsSelected).toBe(false); expect(w.vm.postForm.totalTime).toBeUndefined(); w.destroy()
  })
  it('preserves an old profile-less 002 edit after asynchronous detail loading and a repository change', async () => {
    api.fetchManagementTraitsProfile.mockResolvedValue({ data: null }); const w = form(); w.vm.$route.params.id = 'exam-1'
    await w.vm.fetchData('exam-1'); await flush(); expect(w.vm.managementTraitsSelected).toBe(false)
    await choose(w, choice('00202')); expect(w.vm.managementTraitsSelected).toBe(false); expect(api.freezeManagementTraitsProfile).not.toHaveBeenCalled(); w.destroy()
  })
  it('preserves frozen TEST editing without changing its repo, fields or frozen selection', async () => {
    const w = form(); await w.vm.fetchData('exam-1'); await flush(); const before = JSON.stringify(w.vm.postForm)
    w.vm.repoChange(choice('00202'), w.vm.repoList[0], 0); await w.vm.$nextTick(); expect(w.vm.managementTraitsSelected).toBe(true); expect(w.vm.managementTraitsFrozen).toBe(true)
    // The child selection is disabled; direct handler calls must also leave the code and fields intact.
    expect(w.vm.repoCode).toBe('00201'); expect(w.vm.requiredFieldsList).toEqual(['age', 'degree']); expect(JSON.stringify(w.vm.postForm)).toBe(before); w.destroy()
  })
  it('does not default without admin permission or within the 00401 configuration', async () => {
    const w = form(); w.vm.$store.state.user.userId = 2; await choose(w, choice()); expect(w.vm.managementTraitsSelected).toBe(false)
    w.vm.$store.state.user.userId = 1; w.vm.postForm.assessmentType = 'competency'; w.vm.handleAssessmentTypeChange('competency'); await w.vm.$nextTick()
    const versions = [w.vm.postForm.competencyProductVersion, w.vm.postForm.competencyScoringVersion, w.vm.postForm.dimensionIds.join(',')]
    await choose(w, choice()); expect(w.vm.managementTraitsSelected).toBe(false)
    expect([w.vm.postForm.competencyProductVersion, w.vm.postForm.competencyScoringVersion, w.vm.postForm.dimensionIds.join(',')]).toEqual(versions); w.destroy()
  })
  it.each(['cancel', 'confirm'])('new default still requires separate freeze confirmation: %s', async decision => {
    const w = form(); await choose(w, choice()); expect(w.vm.managementTraitsSelected).toBe(true)
    if (decision === 'cancel') w.vm.$confirm.mockRejectedValue('cancel')
    await w.vm.submitForm(); expect(examApi.saveData).toHaveBeenCalledTimes(1)
    expect(api.freezeManagementTraitsProfile).toHaveBeenCalledTimes(decision === 'confirm' ? 1 : 0)
    expect(w.vm.managementTraitsFrozen).toBe(decision === 'confirm'); w.destroy()
  })
  it('preserves arbitrary configured fields and submits only the legacy Save contract before explicit freeze', async () => {
    const w = form(); w.vm.repoList = [{ ...newRepo }]; w.vm.repoCode = '00501'; w.vm.requiredFieldsList = ['age', 'degree', 'major']; w.vm.managementTraitsSelected = true; await w.vm.submitForm()
    const payload = examApi.saveData.mock.calls[0][0]; expect(payload.assessmentType).toBe('legacy'); expect(payload.scoringMode).toBe('legacy'); expect(payload.totalTime).toBe(25); expect(payload.requiredFields).toBe('age,degree,major'); expect(payload).not.toHaveProperty('managementTraitsSelected'); expect(api.freezeManagementTraitsProfile).toHaveBeenCalledWith('exam-1'); expect(w.vm.managementTraitsFrozen).toBe(true)
  })
  it.each(['cancel', 'save-failed'])('never freezes after %s', async reason => {
    const w = form(); w.vm.repoList = [{ ...newRepo }]; w.vm.repoCode = '00501'; w.vm.managementTraitsSelected = true
    if (reason === 'cancel') w.vm.$confirm.mockRejectedValue('cancel'); else examApi.saveData.mockRejectedValue(new Error('保存失败'))
    await w.vm.submitForm(); expect(api.freezeManagementTraitsProfile).not.toHaveBeenCalled(); expect(w.vm.saving).toBe(false)
  })
  it.each([{ ...repo, radioCount: 139 }, { ...repo, repoCode: '00101' }, { ...repo, multiCount: 1 }])('rejects incompatible TEST repository %j', async target => {
    const w = form(); w.vm.repoList = [target]; w.vm.repoCode = target.repoCode; w.vm.managementTraitsSelected = true; await w.vm.submitForm(); expect(examApi.saveData).not.toHaveBeenCalled()
  })
  it('loads frozen configuration read-only and rejects direct submit attempts', async () => {
    const w = form(); await w.vm.fetchData('exam-1'); await flush(); expect(w.vm.managementTraitsFrozen).toBe(true); await w.vm.submitForm(); expect(examApi.saveData).not.toHaveBeenCalled()
  })
  it('does not replace a fetched field subset with defaults', async () => {
    api.fetchManagementTraitsProfile.mockResolvedValue({ data: null }); const w = form(); await w.vm.fetchData('exam-1'); await flush(); expect(w.vm.requiredFieldsList).toEqual(['age', 'degree'])
  })
  it('blocks save after failed profile detection until retry', async () => {
    api.fetchManagementTraitsProfile.mockRejectedValue(new Error('HTTP 503')); const w = form(); await w.vm.fetchData('exam-1'); await flush(); await w.vm.submitForm(); expect(examApi.saveData).not.toHaveBeenCalled(); expect(w.vm.managementTraitsError).toContain('503')
  })
  it('locks editing when freeze failed with an uncertain server outcome, then retries the real profile', async () => {
    api.freezeManagementTraitsProfile.mockRejectedValue(new Error('HTTP 503'))
    const w = form(); w.vm.repoList = [{ ...newRepo }]; w.vm.repoCode = '00501'; w.vm.managementTraitsSelected = true; await w.vm.submitForm()
    expect(w.vm.managementTraitsReadOnly).toBe(true); expect(w.vm.managementTraitsError).toContain('503')
    await w.vm.fetchData('exam-1'); expect(w.vm.managementTraitsFrozen).toBe(true); expect(w.vm.managementTraitsError).toBe('')
  })
  it('suppresses a duplicate save while waiting for explicit freeze confirmation', async () => {
    const pending = deferred(); const w = form(); w.vm.repoList = [{ ...newRepo }]; w.vm.repoCode = '00501'; w.vm.managementTraitsSelected = true; w.vm.$confirm.mockReturnValue(pending.promise)
    const first = w.vm.submitForm(); await flush(); await w.vm.submitForm(); expect(examApi.saveData).toHaveBeenCalledTimes(1); expect(api.freezeManagementTraitsProfile).not.toHaveBeenCalled()
    pending.reject('cancel'); await first; expect(w.vm.saving).toBe(false); expect(w.vm.postForm.id).toBe('exam-1')
  })
  it('does not force name/telephone or defaults on an explicitly empty configured subset', async () => {
    examApi.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', assessmentType: 'legacy', repoList: [repo], joinType: 1, requiredFields: '' } }); api.fetchManagementTraitsProfile.mockResolvedValue({ data: null })
    const w = form(); await w.vm.fetchData('exam-1'); expect(w.vm.requiredFieldsList).toEqual([])
  })
  it('does not save or freeze TEST when only the admin role is present', async () => {
    const w = mount('exam/exam/form.vue', { store: { state: { user: { userId: 2, roles: ['admin'] } }, getters: { permissions: [] } }, mocks: { $route: { params: {} } } })
    w.vm.managementTraitsSelected = true; w.vm.repoList = [{ ...repo }]; await w.vm.submitForm(); expect(examApi.saveData).not.toHaveBeenCalled(); expect(api.freezeManagementTraitsProfile).not.toHaveBeenCalled()
  })
  for (const [file, method] of [['exam/exam/index.vue', 'handleExamDetail'], ['index.vue', 'goExamDetail']]) {
    it.each([true, false])(`${file} routes frozen=%s without guessing`, async isFrozen => {
      api.fetchManagementTraitsProfile.mockResolvedValue({ data: isFrozen ? frozen : null }); const w = mount(file); await w.vm[method]({ id: 'exam-1', repoCode: '00201', assessmentType: 'legacy', isOpen: 1 }); expect(w.vm.$router.push).toHaveBeenCalledWith(isFrozen ? { name: 'ManagementTraitsResults', params: { examId: 'exam-1' } } : expect.any(Object)); if (!isFrozen) expect(w.vm.$router.push.mock.calls[0][0].name).not.toBe('ManagementTraitsResults')
    })
    it(`${file} stops on a missing endpoint`, async () => { api.fetchManagementTraitsProfile.mockRejectedValue(new Error('HTTP 404')); const w = mount(file); await w.vm[method]({ id: 'exam-1', repoCode: '00202' }); expect(w.vm.$router.push).not.toHaveBeenCalled(); expect(w.vm.$message.error).toHaveBeenCalled() })
    it(`${file} never probes non-002 or role-only users`, async () => { const w = mount(file, { store: { state: { user: { userId: 2, roles: ['admin'] } }, getters: { permissions: [] } } }); await w.vm[method]({ id: 'exam-1', repoCode: '00201' }); await w.vm[method]({ id: 'exam-2', repoCode: '00101' }); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled() })
  }
  it('redirects a stale participant-management URL before legacy participant loading', async () => {
    const getList = vi.fn(); const w = mount('user/exam/index.vue', { methods: { getList } }); await flush(); expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ManagementTraitsResults', params: { examId: 'exam-1' } }); expect(getList).not.toHaveBeenCalled()
  })
  it('blocks the legacy participant page after detection failure and provides retry', async () => {
    const getList = vi.fn(); api.fetchManagementTraitsProfile.mockRejectedValue(new Error('HTTP 403')); const w = mount('user/exam/index.vue', { methods: { getList } }); await flush(); expect(getList).not.toHaveBeenCalled(); expect(w.vm.entryError).toContain('403'); expect(w.vm.entryReady).toBe(false)
    examApi.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', assessmentType: 'legacy', scoringMode: 'legacy', repoCode: '00201', managementTraitsProfileFrozen: false, managementTraitsLifecycle: 'legacy', isManagementTraits: false } }); api.fetchManagementTraitsProfile.mockResolvedValue({ data: null }); await w.vm.retryEntry(); expect(getList).toHaveBeenCalledTimes(1); expect(w.vm.entryReady).toBe(true)
  })
  it.each(['handlePaperList', 'handleStatistics', 'handleExportRawData', 'handleExportRawAnswers'])('does not dispatch the old %s for a frozen TEST exam', async method => {
    const w = mount('exam/exam/index.vue', { mocks: { download: vi.fn() } }); await w.vm[method]({ id: 'exam-1', repoCode: '00202' }); expect(w.vm.$router.push).toHaveBeenCalledWith({ name: 'ManagementTraitsResults', params: { examId: 'exam-1' } }); expect(w.vm.download).not.toHaveBeenCalled(); expect(w.vm.$confirm).not.toHaveBeenCalled()
  })
  it.each([['enableManagementTraits', 1, 0], ['disableManagementTraits', 0, 1]])('confirms %s and refreshes the list', async (command, current, target) => {
    api.setManagementTraitsExamState.mockResolvedValue({ data: { examId: 'exam-1', state: target } })
    const w = mount('exam/exam/index.vue')
    const getList = vi.fn()
    w.vm.$refs.pagingTable = { getList }
    await w.vm.handleCommand(command, { id: 'exam-1', repoCode: '00501', state: current })
    await flush()
    expect(w.vm.$confirm).toHaveBeenCalled()
    expect(api.setManagementTraitsExamState).toHaveBeenCalledWith('exam-1', target)
    expect(getList).toHaveBeenCalledTimes(1)
    expect(w.vm.$message.success).toHaveBeenCalled()
  })
  it('keeps the current state and reports a failed 005 state change', async () => {
    api.setManagementTraitsExamState.mockRejectedValue(new Error('state rejected'))
    const w = mount('exam/exam/index.vue')
    const getList = vi.fn()
    w.vm.$refs.pagingTable = { getList }
    await w.vm.handleCommand('enableManagementTraits', { id: 'exam-1', repoCode: '00502', state: 1 })
    await flush()
    expect(getList).not.toHaveBeenCalled()
    expect(w.vm.$message.error).toHaveBeenCalledWith(expect.stringContaining('state rejected'))
  })
})
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import compiler from 'vue-template-compiler'
import babel from '@babel/core'
import fs from 'fs'
import path from 'path'
import * as realApi from '@/api/managementTraits'
import * as product from '@/utils/managementTraitsProduct'

vi.mock('axios', () => ({ default: { create: vi.fn(() => vi.fn()) } }))
vi.mock('@/utils/auth', () => ({ getToken: vi.fn() }))
const api = { ...realApi, fetchManagementTraitsPaper: vi.fn(), saveManagementTraitsAnswer: vi.fn(), submitManagementTraitsPaper: vi.fn(), createManagementTraitsPaper: vi.fn(), registerManagementTraitsCandidate: vi.fn(), loginManagementTraitsTester: vi.fn(), fetchManagementTraitsExamConfig: vi.fn() }
const legacy = { fetchDetail: vi.fn(), saveData: vi.fn(), testerLogin: vi.fn(), getTesterByIdNumber: vi.fn(), createPaper: vi.fn(), paperDetail: vi.fn(), updateData: vi.fn(), updateTester: vi.fn(), fetchCandidate: vi.fn(), paperResult: vi.fn(), paperResult2: vi.fn(), pdfPersistence: vi.fn(), pdfPersistence2: vi.fn() }
const jwt = (purpose, extra = {}) => `header.${btoa(JSON.stringify({ purpose, exam_id: 'exam-1', participant_id: 'person-1', participant_type: 'candidate', paper_id: purpose === 'management_traits_paper' ? 'paper-1' : '', ...extra }))}.signature`
const paper = (overrides = {}) => ({ paperId: 'paper-1', examId: 'exam-1', title: 'TEST 管理特质', paperToken: jwt('management_traits_paper'), serverTime: '2026-10-03T01:00:00Z', startedAt: '2026-10-03T01:00:00Z', deadline: '2026-10-03T01:25:00Z', reminderAt: '2026-10-03T01:20:00Z', state: 1, answered: 0, questions: Array.from({ length: 140 }, (_, i) => ({ id: `pq-${i}`, displayOrder: i + 1, content: `statement ${i}`, selectedOptionId: null, options: Array.from({ length: 5 }, (_, j) => ({ id: `op-${i}-${j}`, content: `choice ${j}`, displayOrder: j + 1 })) })), ...overrides })
function component(file) {
  const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/paper/exam', file), 'utf8')
  const sfc = compiler.parseComponent(source); const code = babel.transformSync(sfc.script.content, { babelrc: false, configFile: false, plugins: ['@babel/plugin-transform-modules-commonjs'] }).code
  const module = { exports: {} }
  new Function('require', 'module', 'exports', code)(name => {
    if (name === '@/utils/managementTraitsProduct') return product
    if (name === '@/api/managementTraits') return api
    if (name.startsWith('@/api/')) return legacy
    return { __esModule: true, default: {} }
  }, module, module.exports)
  return { ...module.exports.default, ...compiler.compileToFunctions(sfc.template.content) }
}
function mount(file = 'managementTraitsExam.vue', options = {}) {
  return shallowMount(component(file), { directives: { loading: () => {} }, ...options,
    mocks: { $route: { params: { paperId: 'paper-1', examId: 'exam-1', repoCode: '00201', stuFlag: '1', id: 'person-1' }, query: {} }, $router: { replace: vi.fn(), push: vi.fn(), go: vi.fn() }, $message: { error: vi.fn(), warning: vi.fn() }, $notify: vi.fn(), $alert: vi.fn(), $confirm: vi.fn(() => Promise.resolve()), ...options.mocks },
    stubs: { 'el-card': true, 'el-button': true, 'el-alert': { props: ['title'], template: '<div>{{ title }}</div>' }, 'el-progress': true,
      'el-form': { template: '<div><slot /></div>', methods: { clearValidate() {}, validate(callback) { callback(true) } } },
      'el-form-item': true, 'el-input': true, 'el-select': true, 'el-option': true, 'el-row': true, 'el-col': true, 'el-dialog': true } })
}
const flush = async () => { for (let i = 0; i < 15; i++) await Promise.resolve() }
const deferred = () => { let resolve; let reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
beforeEach(() => {
  sessionStorage.clear(); Object.values(api).forEach(fn => fn.mockReset && fn.mockReset()); Object.values(legacy).forEach(fn => fn.mockReset())
  api.fetchManagementTraitsPaper.mockResolvedValue({ data: paper() }); api.saveManagementTraitsAnswer.mockResolvedValue({ data: { answered: 1 } }); api.submitManagementTraitsPaper.mockResolvedValue({ data: { status: 'completed', complete: true, answered: 140 } })
  legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', isManagementTraits: false, managementTraitsLifecycle: 'legacy', requiredFields: 'name,gender', managementTraitsProfileFrozen: false, state: 0, repoCode: '00201' } }); legacy.getTesterByIdNumber.mockResolvedValue({ data: {} })
  api.fetchManagementTraitsExamConfig.mockResolvedValue({ data: { id: 'exam-1', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields: 'name,gender', managementTraitsProfileFrozen: true, title: 'TEST' } })
})

describe('independent management traits answering', () => {
  it('does not fetch without its own paper token and shows retryable error', async () => { const w = mount(); await flush(); expect(api.fetchManagementTraitsPaper).not.toHaveBeenCalled(); expect(w.vm.error).not.toBe(''); expect(w.vm.loading).toBe(false); w.destroy() })
  it('renders 140 navigation buttons, five scalar choices and one current card', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const w = mount(); await flush()
    expect(w.findAll('.question-nav button')).toHaveLength(140); expect(w.findAll('.scale-options button')).toHaveLength(5); expect(w.findAll('.question-card')).toHaveLength(1); expect(w.vm.remainingText).toBe('25:00'); w.destroy()
  })
  it('restores same paper order/deadline and opens first unanswered with renewed token', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const data = paper(); for (let i = 0; i < 10; i++) data.questions[i].selectedOptionId = `op-${i}-2`; data.answered = 10; data.paperToken = jwt('management_traits_paper', { exp: 9999999999 }); api.fetchManagementTraitsPaper.mockResolvedValue({ data })
    const w = mount(); await flush(); expect(w.vm.currentIndex).toBe(10); expect(w.vm.paper.deadline).toBe(data.deadline); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(data.paperToken); w.destroy()
  })
  it('waits for save success then advances and updates answered count', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const pending = deferred(); api.saveManagementTraitsAnswer.mockReturnValue(pending.promise); const w = mount(); await flush()
    const action = w.vm.handleAnswer('op-0-3'); expect(w.vm.currentIndex).toBe(0); expect(w.vm.paper.questions[0].selectedOptionId).toBe(null)
    await w.vm.handleAnswer('op-0-4'); expect(api.saveManagementTraitsAnswer).toHaveBeenCalledTimes(1)
    pending.resolve({ data: { answered: 1 } }); await action; expect(w.vm.currentIndex).toBe(1); expect(w.vm.paper.questions[0].selectedOptionId).toBe('op-0-3'); expect(w.vm.paper.answered).toBe(1); w.destroy()
  })
  it('preserves prior saved answer on failure and explicitly retries the attempted scalar', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const data = paper({ answered: 1 }); data.questions[0].selectedOptionId = 'op-0-1'; api.fetchManagementTraitsPaper.mockResolvedValue({ data }); const w = mount(); await flush(); w.vm.currentIndex = 0
    api.saveManagementTraitsAnswer.mockRejectedValueOnce(new Error('网络失败')); await w.vm.handleAnswer('op-0-4'); expect(w.vm.paper.questions[0].selectedOptionId).toBe('op-0-1'); expect(w.vm.currentIndex).toBe(0)
    await w.vm.retryAnswer(); expect(w.vm.paper.questions[0].selectedOptionId).toBe('op-0-4'); w.destroy()
  })
  it('supports numeric choice keys without turning them into numeric option IDs', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const w = mount(); await flush(); await w.vm.onKeydown({ key: '3', target: { tagName: 'BODY' }, preventDefault: vi.fn() }); expect(api.saveManagementTraitsAnswer).toHaveBeenCalledWith('paper-1', 'pq-0', 'op-0-2', expect.any(String)); w.destroy()
  })
  it('does not wrap after last saved question or auto-submit', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const w = mount(); await flush(); w.vm.currentIndex = 139; await w.vm.handleAnswer('op-139-0'); expect(w.vm.currentIndex).toBe(139); expect(api.submitManagementTraitsPaper).not.toHaveBeenCalled(); w.destroy()
  })
  it('uses monotonic elapsed time, not a changed client wall clock, for reminder/deadline', async () => {
    const clock = vi.spyOn(performance, 'now').mockReturnValue(500); sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const w = mount(); await flush()
    const wall = vi.spyOn(Date, 'now').mockReturnValue(1); clock.mockReturnValue(1200500); w.vm.tick(); expect(w.vm.remainingSeconds).toBe(300); expect(w.vm.reminderVisible).toBe(true); expect(api.submitManagementTraitsPaper).not.toHaveBeenCalled(); w.destroy(); wall.mockRestore(); clock.mockRestore()
  })
  it('at deadline sends one manual request; 409 missing never becomes finished, retry remains available', async () => {
    const clock = vi.spyOn(performance, 'now').mockReturnValue(0); sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const pending = deferred(); api.submitManagementTraitsPaper.mockReturnValue(pending.promise); const w = mount(); await flush()
    clock.mockReturnValue(1500000); w.vm.tick(); w.vm.tick(); expect(api.submitManagementTraitsPaper).toHaveBeenCalledTimes(1); expect(api.submitManagementTraitsPaper.mock.calls[0]).toEqual(['paper-1', expect.any(String)])
    pending.reject(new Error('409 missing')); await flush(); expect(w.vm.$router.replace).not.toHaveBeenCalled(); expect(w.vm.error).toContain('409'); api.submitManagementTraitsPaper.mockResolvedValue({ data: { status: 'incomplete', complete: false, answered: 0 } }); await w.vm.submit(); expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ExamThankYou' }); w.destroy(); clock.mockRestore()
  })
  it('manual confirmation locates missing questions rather than making a premature request', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const w = mount(); await flush(); w.vm.currentIndex = 40; await w.vm.confirmSubmit(); expect(w.vm.currentIndex).toBe(0); expect(api.submitManagementTraitsPaper).not.toHaveBeenCalled(); w.destroy()
  })
  it('prevents concurrent confirmation/submit and blocks submission during answer save', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); const data = paper({ answered: 140 }); data.questions.forEach((q, i) => { q.selectedOptionId = `op-${i}-0` }); api.fetchManagementTraitsPaper.mockResolvedValue({ data }); const w = mount(); await flush()
    const pending = deferred(); w.vm.$confirm.mockReturnValue(pending.promise); const action = w.vm.confirmSubmit(); await w.vm.confirmSubmit(); expect(w.vm.$confirm).toHaveBeenCalledTimes(1); pending.reject('cancel'); await action; w.vm.saving = true; await w.vm.submit(); expect(api.submitManagementTraitsPaper).not.toHaveBeenCalled(); w.destroy()
  })
  it('completed detail clears only management tokens and preserves the paper marker', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); sessionStorage.setItem('competencyPaperToken', '004'); api.fetchManagementTraitsPaper.mockResolvedValue({ data: paper({ state: 2 }) }); const w = mount(); await flush(); expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ExamThankYou' }); expect(sessionStorage.getItem('competencyPaperToken')).toBe('004'); expect(realApi.isManagementTraitsPaper('paper-1')).toBe(true); w.destroy()
  })
  it.each([paper({ paperId: 'foreign' }), paper({ questions: [] }), paper({ deadline: 'invalid' })])('fails closed on a malformed or cross-paper DTO', async data => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper')); api.fetchManagementTraitsPaper.mockResolvedValue({ data }); const w = mount(); await flush(); expect(w.vm.paper).toBe(null); expect(w.vm.error).not.toBe(''); w.destroy()
  })
})

describe('management traits identity and legacy namespace separation', () => {
    it.each(['00501', '00502'])('unfrozen server %s never submits to the old identity API even with false legacy metadata', async code => {
      legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', repoCode: code, isManagementTraits: false, managementTraitsLifecycle: 'legacy', managementTraitsProfileFrozen: false, requiredFields: 'name', state: 0 } })
      const w = mount('candidate.vue', { mocks: { $route: { params: { examId: 'exam-1', repoCode: code }, query: {} } } }); await flush(); await w.vm.submitForm()
      expect(w.vm.examBlocked).toBe(true); expect(legacy.saveData).not.toHaveBeenCalled(); expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); w.destroy()
    })
    it.each(['00501', '00502'])('queryless frozen server %s loads only the new configured identity subset', async code => {
      const data = { id: 'exam-1', repoCode: code, managementTraitsProfileFrozen: true, isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields: 'name,telephone', state: 0 }
      legacy.fetchDetail.mockResolvedValue({ data }); api.fetchManagementTraitsExamConfig.mockResolvedValue({ data })
      const w = mount('candidate.vue', { mocks: { $route: { params: { examId: 'exam-1', repoCode: code }, query: {} } } }); await flush()
      expect(w.vm.managementTraitsMode).toBe(true); expect(w.vm.requiredFields).toEqual(['name', 'telephone']); expect(legacy.saveData).not.toHaveBeenCalled(); w.destroy()
    })
  // MT-NEW-DRAFT: the real SFC must not treat frozen=false as legacy.
  it.each(['draft', 'unknown', undefined])('MT-NEW-DRAFT lifecycle %s never sends legacy identity', async lifecycle => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', repoCode: '00201', isManagementTraits: true, managementTraitsLifecycle: lifecycle, managementTraitsProfileFrozen: false, requiredFields: 'name,telephone', state: 0 } })
    const w = mount('candidate.vue'); await flush(); await w.vm.submitForm()
    expect(w.vm.examBlocked).toBe(true); expect(w.vm.requiredFields).toBe(null)
    expect(legacy.saveData).not.toHaveBeenCalled(); expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); w.destroy()
  })
  const newRoute = { params: { examId: 'exam-1', repoCode: '00201', stuFlag: '1', id: 'person-1' }, query: { mngTest: '1' } }
  it('MT-CANDIDATE-FROZEN forged non002 URL cannot bypass unknown server002 marker', async () => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', repoCode: '00201', requiredFields: 'name,telephone', state: 0 } })
    const route = { params: { examId: 'exam-1', repoCode: '00101', stuFlag: '1' }, query: {} }
    const w = mount('candidate.vue', { mocks: { $route: route } }); await flush(); await w.vm.submitForm()
    expect(w.vm.examBlocked).toBe(true); expect(legacy.saveData).not.toHaveBeenCalled(); expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); w.destroy()
  })
  it.each(['foreign', undefined, {}])('MT-CANDIDATE-FROZEN legacy false must belong to current exam, not %s', async responseId => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: responseId, repoCode: '00201', managementTraitsProfileFrozen: false, requiredFields: 'name,telephone', state: 0 } })
    const w = mount('candidate.vue'); await flush(); await w.vm.submitForm()
    expect(w.vm.examBlocked).toBe(true); expect(legacy.saveData).not.toHaveBeenCalled(); expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); w.destroy()
  })
  // MT-CANDIDATE-FROZEN: docs/regression-tests.md; actual SFC, no remote requests.
  it.each([false, undefined, null, 'true', 1])('MT-CANDIDATE-FROZEN explicit URL with untrusted frozen flag %s cannot register identity', async frozen => {
    api.fetchManagementTraitsExamConfig.mockResolvedValue({ data: { id: 'exam-1', requiredFields: 'name,telephone', managementTraitsProfileFrozen: frozen } })
    const w = mount('candidate.vue', { mocks: { $route: newRoute } }); await flush()
    w.vm.candidateForm.name = 'Synthetic'; w.vm.candidateForm.telephone = '13800000000'
    await w.vm.submitForm()
    expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); expect(legacy.saveData).not.toHaveBeenCalled()
    expect(w.vm.examBlocked).toBe(true); expect(w.vm.requiredFields).toBe(null)
    expect(w.vm.managementTraitsError).toBe('测评身份配置未确认冻结或与当前测评不一致，请重试或联系管理员。')
    expect(w.find('el-form-stub').exists()).toBe(false); expect(legacy.fetchDetail).not.toHaveBeenCalled(); w.destroy()
  })
  it.each([false, undefined])('MT-CANDIDATE-FROZEN normal first read true then second read %s remains new and blocked', async frozen => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', repoCode: '00201', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields: 'name,telephone', managementTraitsProfileFrozen: true, state: 0 } })
    api.fetchManagementTraitsExamConfig.mockResolvedValue({ data: { id: 'exam-1', requiredFields: 'name,telephone', managementTraitsProfileFrozen: frozen } })
    const w = mount('candidate.vue'); await flush()
    w.vm.candidateForm.name = 'Synthetic'; w.vm.candidateForm.telephone = '13800000000'
    await w.vm.submitForm()
    expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); expect(legacy.saveData).not.toHaveBeenCalled(); expect(legacy.fetchCandidate).not.toHaveBeenCalled()
    expect(w.vm.managementTraitsMode).toBe(true); expect(w.vm.examBlocked).toBe(true); expect(w.vm.requiredFields).toBe(null)
    expect(w.vm.managementTraitsError).toBe('测评身份配置未确认冻结或与当前测评不一致，请重试或联系管理员。')
    expect(legacy.fetchDetail).toHaveBeenCalledTimes(1); expect(api.fetchManagementTraitsExamConfig).toHaveBeenCalledTimes(1); w.destroy()
  })
  it.each(['foreign', undefined, {}, true, 9007199254740992])('MT-CANDIDATE-FROZEN true flag with invalid response ID %s stays closed', async responseId => {
    api.fetchManagementTraitsExamConfig.mockResolvedValue({ data: { id: responseId, requiredFields: 'name,telephone', managementTraitsProfileFrozen: true } })
    const w = mount('candidate.vue', { mocks: { $route: newRoute } }); await flush(); await w.vm.submitForm()
    expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); expect(legacy.saveData).not.toHaveBeenCalled(); expect(w.vm.examBlocked).toBe(true); w.destroy()
  })
  it.each(['123', 123])('MT-CANDIDATE-FROZEN same-exam ID %s accepts true frozen whitelist without extra reads', async responseId => {
    const route = { params: { examId: '123', repoCode: '00201', stuFlag: '1' }, query: { mngTest: '1' } }
    api.fetchManagementTraitsExamConfig.mockResolvedValue({ data: { id: responseId, isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields: 'name,telephone', managementTraitsProfileFrozen: true } })
    api.registerManagementTraitsCandidate.mockResolvedValue({ data: { id: 'person-1', participantToken: jwt('management_traits_participant', { exam_id: '123' }) } })
    const w = mount('candidate.vue', { mocks: { $route: route } }); await flush(); await w.vm.submitForm()
    expect(w.vm.examBlocked).toBe(false); expect(w.vm.requiredFields).toEqual(['name', 'telephone'])
    expect(api.registerManagementTraitsCandidate).toHaveBeenCalledWith('123', ['name', 'telephone'], expect.any(Object))
    expect(api.fetchManagementTraitsExamConfig).toHaveBeenCalledTimes(1); expect(legacy.fetchDetail).not.toHaveBeenCalled(); expect(legacy.saveData).not.toHaveBeenCalled(); w.destroy()
  })
  it.each([undefined, null, 'false', 'true', 0, 1, {}, []])('MT-CANDIDATE-FROZEN queryless unknown marker %s cannot fall through to legacy', async frozen => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', requiredFields: 'name,telephone', managementTraitsProfileFrozen: frozen, state: 0, repoCode: '00201' } })
    const w = mount('candidate.vue'); await flush(); await w.vm.submitForm()
    expect(w.vm.examBlocked).toBe(true); expect(w.vm.requiredFields).toBe(null)
    expect(legacy.saveData).not.toHaveBeenCalled(); expect(legacy.fetchCandidate).not.toHaveBeenCalled()
    expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); w.destroy()
  })
  // MT-CANDIDATE-FIELDS: docs/regression-tests.md; real compiled SFC, synthetic identities only.
  it.each(['name,telephone', 'name,gender,telephone'])('queryless server-frozen %s uses only configured rules and the existing strict API', async requiredFields => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields, managementTraitsProfileFrozen: true, state: 0, repoCode: '00201' } })
    api.fetchManagementTraitsExamConfig.mockResolvedValue({ data: { id: 'exam-1', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields, managementTraitsProfileFrozen: true } })
    api.registerManagementTraitsCandidate.mockResolvedValue({ data: { id: 'person-1', participantToken: jwt('management_traits_participant') } })
    const w = mount('candidate.vue'); await flush()
    expect(w.vm.managementTraitsMode).toBe(true)
    expect(Object.keys(w.vm.candidateRules).sort()).toEqual(requiredFields.split(',').sort())
    expect(w.vm.showField('gender')).toBe(requiredFields.includes('gender'))
    expect(w.findAll('el-form-item-stub').wrappers.map(item => item.attributes('prop')).sort()).toEqual(requiredFields.split(',').sort())
    w.vm.candidateForm.name = 'Synthetic'; w.vm.candidateForm.telephone = '13800000000'
    w.vm.candidateForm.gender = '0'; w.vm.candidateForm.depart = 'hidden'; w.vm.candidateForm.grade = 'stale'
    await w.vm.submitForm()
    expect(api.registerManagementTraitsCandidate).toHaveBeenCalledWith('exam-1', requiredFields.split(','), expect.any(Object))
    expect(legacy.saveData).not.toHaveBeenCalled(); expect(legacy.fetchCandidate).not.toHaveBeenCalled(); w.destroy()
  })
  it('queryless pending config blocks even a direct submit', async () => {
    const pending = deferred(); legacy.fetchDetail.mockReturnValue(pending.promise)
    const w = mount('candidate.vue'); await w.vm.submitForm()
    expect(legacy.saveData).not.toHaveBeenCalled(); expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled()
    expect(w.vm.examBlocked).toBe(true); pending.resolve({ data: { requiredFields: 'name', state: 0 } }); await flush(); w.destroy()
  })
  it('queryless metadata failure never sends a legacy full payload', async () => {
    legacy.fetchDetail.mockRejectedValue(new Error('config unavailable')); const w = mount('candidate.vue'); await flush()
    await w.vm.submitForm(); expect(w.vm.examBlocked).toBe(true); expect(legacy.saveData).not.toHaveBeenCalled(); w.destroy()
  })
  it('late previous-exam config cannot select fields or unblock the current exam', async () => {
    const pending = deferred(); legacy.fetchDetail.mockReturnValueOnce(pending.promise).mockResolvedValue({ data: { id: 'exam-2', repoCode: '00201', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields: 'name,telephone', managementTraitsProfileFrozen: true, state: 0 } })
    api.fetchManagementTraitsExamConfig.mockResolvedValue({ data: { id: 'exam-2', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields: 'name,telephone', managementTraitsProfileFrozen: true } })
    const route = { params: { examId: 'exam-1', repoCode: '00201', stuFlag: '1' }, query: {} }
    const w = mount('candidate.vue', { mocks: { $route: route } }); w.vm.candidateForm.gender = '0'
    route.params.examId = 'exam-2'; await w.vm.$nextTick(); await flush()
    pending.resolve({ data: { requiredFields: 'name,gender', managementTraitsProfileFrozen: true, state: 0 } }); await flush()
    expect(w.vm.examId).toBe('exam-2'); expect(w.vm.requiredFields).toEqual(['name', 'telephone']); expect(w.vm.candidateForm.gender).toBe(''); w.destroy()
  })
  it('queryless legacy 002 retains legacy payload and lookup after successful configuration', async () => {
    legacy.saveData.mockResolvedValue({ data: { id: 'legacy-person' } })
    const w = mount('candidate.vue'); await flush(); expect(w.vm.managementTraitsMode).toBe(false)
    await w.vm.submitForm(); expect(legacy.saveData).toHaveBeenCalledWith(w.vm.candidateForm); expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); w.destroy()
  })
  it('late submitted identity cannot navigate or store credentials for a different exam', async () => {
    const pending = deferred(); api.registerManagementTraitsCandidate.mockReturnValue(pending.promise)
    const route = { params: { examId: 'exam-1', repoCode: '00201', stuFlag: '1' }, query: { mngTest: '1' } }
    const w = mount('candidate.vue', { mocks: { $route: route } }); await flush(); const action = w.vm.submitForm()
    route.params.examId = 'exam-2'; await w.vm.$nextTick(); await flush()
    const warningCalls = w.vm.$message.error.mock.calls.length
    pending.resolve({ data: { id: 'person-1', participantToken: jwt('management_traits_participant') } }); await action
    expect(w.vm.$router.replace).not.toHaveBeenCalled(); expect(sessionStorage.getItem('managementTraitsParticipantToken')).toBe(null)
    expect(w.vm.$message.error).toHaveBeenCalledTimes(warningCalls); expect(w.vm.candidateForm.id).toBe(''); w.destroy()
  })
  it.each(['name,gender,unknown', 'name,name', '', 'name,mobile', 'name,sex'])('server-frozen invalid %s stays closed without fallback', async requiredFields => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields, managementTraitsProfileFrozen: true, state: 0 } })
    api.fetchManagementTraitsExamConfig.mockResolvedValue({ data: { id: 'exam-1', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields, managementTraitsProfileFrozen: true } })
    const w = mount('candidate.vue'); await flush(); await w.vm.submitForm()
    expect(w.vm.examBlocked).toBe(true); expect(legacy.saveData).not.toHaveBeenCalled(); expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); w.destroy()
  })
  it('explicit candidate mode loads configured fields only, no legacy person lookup', async () => {
    const w = mount('candidate.vue', { mocks: { $route: newRoute } }); await flush(); expect(api.fetchManagementTraitsExamConfig).toHaveBeenCalledWith('exam-1'); expect(legacy.fetchCandidate).not.toHaveBeenCalled(); expect(w.text()).toContain('TEST'); expect(w.vm.showField('telephone')).toBe(false); w.destroy()
  })
  it('explicit candidate configuration failure is closed, not default-all', async () => {
    api.fetchManagementTraitsExamConfig.mockRejectedValue(new Error('503 unavailable')); const w = mount('candidate.vue', { mocks: { $route: newRoute } }); await flush(); expect(w.vm.examBlocked).toBe(true); await w.vm.submitForm(); expect(api.registerManagementTraitsCandidate).not.toHaveBeenCalled(); w.destroy()
  })
  it('candidate receives dedicated participant token and forwards intent without 004 storage', async () => {
    api.registerManagementTraitsCandidate.mockResolvedValue({ data: { id: 'person-1', participantToken: jwt('management_traits_participant') } }); const w = mount('candidate.vue', { mocks: { $route: newRoute } }); await flush(); await w.vm.submitForm(); expect(sessionStorage.getItem('competencyParticipantToken')).toBe(null); expect(sessionStorage.getItem('managementTraitsParticipantToken')).toBe(jwt('management_traits_participant')); expect(w.vm.$router.replace.mock.calls[0][0].query).toEqual({ mngTest: '1' }); w.destroy()
  })
  it('explicit candidate rejects non-management token without legacy fallback', async () => {
    api.registerManagementTraitsCandidate.mockResolvedValue({ data: { id: 'person-1', participantToken: jwt('competency_participant') } }); const w = mount('candidate.vue', { mocks: { $route: newRoute } }); await flush(); await w.vm.submitForm(); expect(w.vm.$router.replace).not.toHaveBeenCalled(); expect(sessionStorage.getItem('competencyParticipantToken')).toBe(null); expect(legacy.saveData).not.toHaveBeenCalled(); w.destroy()
  })
  it('queryless candidate branches on returned management purpose, never 004 token keys', async () => {
    legacy.saveData.mockResolvedValue({ data: { id: 'person-1', participantToken: jwt('management_traits_participant') } }); const w = mount('candidate.vue'); await flush(); await w.vm.submitForm(); expect(sessionStorage.getItem('competencyParticipantToken')).toBe(null); expect(w.vm.$router.replace.mock.calls[0][0].query).toEqual({ mngTest: '1' }); w.destroy()
  })
  it('explicit tester uses only login identity and does not block existing papers based on client end time', async () => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', repoCode: '00201', managementTraitsProfileFrozen: true, state: 3, timeLimit: true, endTime: '2020-01-01 00:00:00' } })
    api.loginManagementTraitsTester.mockResolvedValue({ data: { id: 'person-1', participantToken: jwt('management_traits_participant') } }); const w = mount('tester.vue', { mocks: { $route: newRoute } }); await flush(); w.vm.testerFrom = { idNumber: '123', password: 'test', extra: 'drop' }; await w.vm.submitForm(); expect(api.loginManagementTraitsTester).toHaveBeenCalledWith('exam-1', '123', 'test'); expect(legacy.testerLogin).not.toHaveBeenCalled(); expect(legacy.fetchDetail).toHaveBeenCalledTimes(1); expect(sessionStorage.getItem('competencyParticipantToken')).toBe(null); w.destroy()
  })
  // MT-TESTER-METADATA: docs/regression-tests.md; URL/session are intent, not authority.
  it.each(['00201', '00202', '00101', '00301', '00401'])('MT-TESTER-METADATA URL intent cannot change normal %s login', async code => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', repoCode: code, managementTraitsProfileFrozen: false, state: 0 } })
    legacy.testerLogin.mockResolvedValue({ data: { id: 'legacy-person' } })
    const w = mount('tester.vue', { mocks: { $route: { params: { examId: 'exam-1', repoCode: code }, query: { mngTest: '1' } } } }); await flush()
    await w.vm.submitForm(); expect(legacy.fetchDetail).toHaveBeenCalledTimes(1); expect(legacy.testerLogin).toHaveBeenCalledTimes(1); expect(api.loginManagementTraitsTester).not.toHaveBeenCalled(); w.destroy()
  })
  it.each(['00501', '00502'])('MT-TESTER-METADATA URL intent never unlocks unfrozen %s', async code => {
    legacy.fetchDetail.mockResolvedValue({ data: { id: 'exam-1', repoCode: code, isManagementTraits: true, managementTraitsLifecycle: 'draft', managementTraitsProfileFrozen: false, state: 0 } })
    const w = mount('tester.vue', { mocks: { $route: { params: { examId: 'exam-1', repoCode: code }, query: { mngTest: '1' } } } }); await flush(); await w.vm.submitForm()
    expect(legacy.fetchDetail).toHaveBeenCalledTimes(1); expect(w.vm.examBlocked).toBe(true); expect(legacy.testerLogin).not.toHaveBeenCalled(); expect(api.loginManagementTraitsTester).not.toHaveBeenCalled(); w.destroy()
  })
  it.each(['foreign', undefined, {}, 9007199254740992])('MT-TESTER-METADATA cross-exam %s cannot authorize frozen login', async id => {
    legacy.fetchDetail.mockResolvedValue({ data: { id, repoCode: '00501', managementTraitsProfileFrozen: true, state: 0 } })
    const w = mount('tester.vue', { mocks: { $route: { params: { examId: 'exam-1', repoCode: '00101' }, query: { mngTest: '1' } } } }); await flush(); await w.vm.submitForm()
    expect(w.vm.examBlocked).toBe(true); expect(legacy.testerLogin).not.toHaveBeenCalled(); expect(api.loginManagementTraitsTester).not.toHaveBeenCalled(); w.destroy()
  })
  it('MT-TESTER-METADATA stored token cannot skip pending or failed metadata', async () => {
    realApi.rememberManagementTraitsParticipant(jwt('management_traits_participant'), 'exam-1')
    const pending = deferred(); legacy.fetchDetail.mockReturnValue(pending.promise)
    const w = mount('tester.vue'); await w.vm.submitForm(); expect(w.vm.configLoading).toBe(true); expect(api.loginManagementTraitsTester).not.toHaveBeenCalled()
    pending.reject(new Error('metadata unavailable')); await flush(); await w.vm.submitForm(); expect(w.vm.examBlocked).toBe(true); expect(legacy.testerLogin).not.toHaveBeenCalled(); expect(api.loginManagementTraitsTester).not.toHaveBeenCalled(); w.destroy()
  })
  it('prepare branches before ANY old detail/person/create/update calls', async () => {
    realApi.rememberManagementTraitsParticipant(jwt('management_traits_participant'), 'exam-1'); api.createManagementTraitsPaper.mockResolvedValue({ data: paper() }); const w = mount('preview.vue', { mocks: { $route: newRoute } }); await flush(); await w.vm.handleCreate()
    expect(legacy.fetchDetail).not.toHaveBeenCalled(); expect(legacy.getTesterByIdNumber).not.toHaveBeenCalled(); expect(legacy.createPaper).not.toHaveBeenCalled(); expect(legacy.updateData).not.toHaveBeenCalled(); expect(legacy.updateTester).not.toHaveBeenCalled(); expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ManagementTraitsExam', params: { paperId: 'paper-1' } }); w.destroy()
  })
  it('prepare explicit intent without authentication cannot call either creation API', async () => {
    const w = mount('preview.vue', { mocks: { $route: newRoute } }); await flush(); await w.vm.handleCreate(); expect(api.createManagementTraitsPaper).not.toHaveBeenCalled(); expect(legacy.createPaper).not.toHaveBeenCalled(); w.destroy()
  })
  it('legacy 002 prepare still uses existing detail/person chain when scope is unknown', async () => { const w = mount('preview.vue'); await flush(); expect(legacy.fetchDetail).toHaveBeenCalledTimes(1); expect(legacy.getTesterByIdNumber).toHaveBeenCalledTimes(1); expect(api.createManagementTraitsPaper).not.toHaveBeenCalled(); w.destroy() })
  it('known new paper cannot execute old result score, PDF rendering or upload after tokens cleared', async () => {
    realApi.rememberManagementTraitsPaper(paper()); realApi.clearManagementTraitsTokens(); const w = mount('result2.vue', { mocks: { $route: { params: { id: 'paper-1', testerId: 'person-1' }, query: {} } } }); await flush(); await w.vm.fetchScore('paper-1'); await w.vm.createPdf(); await w.vm.UploadPdf('data'); expect(legacy.paperResult).not.toHaveBeenCalled(); expect(legacy.paperResult2).not.toHaveBeenCalled(); expect(legacy.pdfPersistence).not.toHaveBeenCalled(); expect(legacy.pdfPersistence2).not.toHaveBeenCalled(); expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ExamThankYou' }); w.destroy()
  })
  it('constant routes register real isolated participant and admin components', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/router/index.js'), 'utf8'); expect(source).toContain("path: '/exam/management-traits/start/:paperId'"); expect(source).toContain("name: 'ManagementTraitsExam'"); expect(source).toContain("path: '/exam/management-traits-results'"); expect(source).toContain("name: 'ManagementTraitsResults'")
  })
})
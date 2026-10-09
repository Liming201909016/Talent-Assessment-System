import { beforeEach, describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import compiler from 'vue-template-compiler'
import babel from '@babel/core'
import fs from 'fs'
import path from 'path'
import axios from 'axios'
import * as realApi from '@/api/managementTraits'
import * as product from '@/utils/managementTraitsProduct'

vi.mock('axios', () => ({ default: { create: vi.fn(() => vi.fn()) } }))
vi.mock('@/utils/auth', () => ({ getToken: vi.fn(() => 'mock-admin') }))
const api = { ...realApi, fetchManagementTraitsProfile: vi.fn(), fetchManagementTraitsExamConfig: vi.fn(), fetchManagementTraitsPaper: vi.fn() }
const legacy = { fetchDetail: vi.fn(), fetchList: vi.fn(() => Promise.resolve({ data: [] })), saveData: vi.fn(), fetchCompetencyDimensions: vi.fn() }
const flush = async () => { for (let i = 0; i < 20; i++) await Promise.resolve() }
const jwt = (purpose, extra = {}) => `h.${btoa(JSON.stringify({ purpose, exam_id: 'exam-1', paper_id: 'paper-1', ...extra }))}.mock`
function mount(file, extra = {}) {
  const sfc = compiler.parseComponent(fs.readFileSync(path.resolve(process.cwd(), 'src/views', file), 'utf8'))
  const code = babel.transformSync(sfc.script.content, { babelrc: false, configFile: false, plugins: ['@babel/plugin-transform-modules-commonjs'] }).code
  const module = { exports: {} }
  new Function('require', 'module', 'exports', code)(name => name === '@/utils/managementTraitsProduct' ? product : name === '@/api/managementTraits' ? api : name.startsWith('@/api/') ? legacy : name === '@/utils/request' ? { __esModule: true, default: { post: vi.fn(() => Promise.resolve({ data: {} })), get: vi.fn(() => Promise.resolve({})) } } : { __esModule: true, default: {} }, module, module.exports)
  return shallowMount({ ...module.exports.default, ...compiler.compileToFunctions(sfc.template.content) }, {
    directives: { loading: () => {} }, ...extra,
    mocks: { $route: { params: {}, query: {} }, $store: { state: { user: { id: 1 } }, getters: { permissions: [] } }, $router: { push: vi.fn(), replace: vi.fn() }, $message: { error: vi.fn(), warning: vi.fn() }, $nextTick: vi.fn(), ...extra.mocks },
    stubs: ['el-card', 'el-alert', 'el-form', 'el-form-item', 'el-table', 'el-table-column', 'el-checkbox', 'el-checkbox-group', 'el-button', 'el-input', 'el-radio', 'el-radio-group', 'el-switch', 'el-input-number', 'el-date-picker', 'el-select', 'el-option', 'el-row', 'el-col', 'el-progress', 'el-dropdown', 'el-dropdown-menu', 'el-dropdown-item', 'data-table', 'repo-select']
  })
}
beforeEach(() => {
  sessionStorage.clear(); axios.create.mock.results[0].value.mockReset()
  api.fetchManagementTraitsProfile.mockReset().mockRejectedValue(new Error('409 no profile'))
  api.fetchManagementTraitsExamConfig.mockReset().mockResolvedValue({ data: { id: 'exam-1', isManagementTraits: true, managementTraitsLifecycle: 'frozen', requiredFields: 'name,age', managementTraitsProfileFrozen: true } })
  api.fetchManagementTraitsPaper.mockReset()
  legacy.fetchDetail.mockReset().mockResolvedValue({ data: { id: 'exam-1', assessmentType: 'legacy', scoringMode: 'legacy', repoCode: '00201', repoList: [{ repoCode: '00201' }], joinType: 1, requiredFields: 'name,age' } })
})
describe('verified delivery gaps, without broad 002 switching', () => {
  it('old002 list navigation does not probe a nonexistent profile or block old detail', async () => {
    const w = mount('exam/exam/index.vue'); await w.vm.handleExamDetail({ id: 'exam-1', repoCode: '00201', isOpen: 1 })
    expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); expect(w.vm.$router.push).toHaveBeenCalledWith(expect.objectContaining({ name: 'ListExamUser' })); w.destroy()
  })
  it('old002 edit remains editable when no explicit TEST scope exists', async () => {
    const w = mount('exam/exam/form.vue', { mocks: { $route: { params: { id: 'exam-1' }, query: {} } } }); await flush()
    expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); expect(w.vm.managementTraitsReadOnly).toBe(false); w.destroy()
  })
  it('old002 dashboard remains on the legacy route', async () => {
    const w = mount('index.vue'); await w.vm.goExamDetail({ id: 'exam-1', repoCode: '00201' }); expect(api.fetchManagementTraitsProfile).not.toHaveBeenCalled(); expect(w.vm.$router.push).toHaveBeenCalled(); w.destroy()
  })
  it('known TEST scope still fails closed on profile error and never routes old', async () => {
    sessionStorage.setItem('managementTraitsExam:exam-1', 'test')
    const w = mount('exam/exam/index.vue'); await w.vm.handleExamDetail({ id: 'exam-1', repoCode: '00201' }); expect(api.fetchManagementTraitsProfile).toHaveBeenCalled(); expect(w.vm.$router.push).not.toHaveBeenCalled(); w.destroy()
  })
  it('frozen form offers an explicit TEST candidate link, not the old public entry', async () => {
    const w = mount('exam/exam/form.vue'); await w.setData({ managementTraitsFrozen: true, repoCode: '00201', postForm: { id: 'exam-1', isOpen: 1, stuFlag: 1 } })
    expect(w.vm.managementTraitsEntry).toMatchObject({ name: 'candidateInfo', query: { mngTest: '1' }, params: { examId: 'exam-1', repoCode: '00201' } }); expect(w.find('.test-entry').exists()).toBe(true); w.destroy()
  })
  it('frozen closed form offers a tester TEST link', async () => {
    const w = mount('exam/exam/form.vue'); await w.setData({ managementTraitsFrozen: true, repoCode: '00202', postForm: { id: 'exam-1', isOpen: 2 } }); expect(w.vm.managementTraitsEntry).toMatchObject({ name: 'tester', query: { mngTest: '1' } }); w.destroy()
  })
  it.each([{ fields: [] }, { fields: ['idNumber'] }, { fields: ['depart'] }, { fields: ['age', 'age'] }])('rejects unsupported TEST field contract before save %j', async ({ fields }) => {
    const w = mount('exam/exam/form.vue'); await w.setData({ requiredFieldsList: fields, managementTraitsSelected: true, repoList: [{ repoId: 'repo-1', repoCode: '00201', radioCount: 140 }], postForm: { assessmentType: 'legacy', scoringMode: 'legacy', joinType: 1 } }); expect(w.vm.validateManagementTraitsConfiguration()).toBe(false); w.destroy()
  })
  it('allows the configured TEST age/degree subset without forcing name/phone', async () => {
    const w = mount('exam/exam/form.vue'); await w.setData({ requiredFieldsList: ['age','degree','stuFlag'], managementTraitsSelected: true, repoList: [{ repoId: 'repo-00502-synthetic', repoCode: '00502', radioCount: 140 }], postForm: { assessmentType: 'legacy', scoringMode: 'legacy', joinType: 1 } }); expect(w.vm.validateManagementTraitsConfiguration()).toBe(true); expect(w.vm.requiredFieldsList).toEqual(['age','degree','stuFlag']); w.destroy()
  })
  it('matches TEST positive integer age contract without inheriting legacy minimum14', async () => {
    const w = mount('paper/exam/candidate.vue', { mocks: { $route: { params: { examId: 'exam-1', repoCode: '00201' }, query: { mngTest: '1' } } } }); await flush()
    const callback = vi.fn(); w.vm.candidateRules.age.find(rule => rule.validator).validator({}, '10', callback); expect(callback).toHaveBeenCalledWith(undefined); w.destroy()
  })
  it('existing same-exam paper credential is tried before re-registering candidate', async () => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt('management_traits_paper'))
    axios.create.mock.results[0].value.mockResolvedValue({ data: { code: 0, success: true, data: { paperId: 'paper-1', examId: 'exam-1', state: 1, paperToken: jwt('management_traits_paper') } } })
    const w = mount('paper/exam/candidate.vue', { mocks: { $route: { params: { examId: 'exam-1', repoCode: '00201' }, query: { mngTest: '1' } } } }); await flush()
    expect(api.fetchManagementTraitsExamConfig).not.toHaveBeenCalled(); expect(w.vm.$router.replace).toHaveBeenCalledWith({ name: 'ManagementTraitsExam', params: { paperId: 'paper-1' } }); w.destroy()
  })
  it('preserves HTTP403 JSON Blob message, never opening a pseudo PDF', async () => {
    const client = axios.create.mock.results[0].value
    client.mockRejectedValue({ response: { status: 403, data: new Blob(['{"msg":"admin denied by service"}'], { type: 'application/json' }) } })
    await expect(realApi.downloadManagementTraitsTestReport('report-1')).rejects.toMatchObject({ message: 'admin denied by service', status: 403 })
  })
})
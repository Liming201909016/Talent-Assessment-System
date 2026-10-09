import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { shallowMount, createLocalVue } from '@vue/test-utils'
import VueRouter from 'vue-router'
import compiler from 'vue-template-compiler'
import babel from '@babel/core'
import fs from 'fs'
import path from 'path'
import { getToken } from '@/utils/auth'
import * as realApi from '@/api/managementTraits'
import * as product from '@/utils/managementTraitsProduct'

const transport = vi.hoisted(() => vi.fn())
vi.mock('axios', () => ({ default: { create: vi.fn(() => transport) } }))
vi.mock('@/utils/auth', () => ({ getToken: vi.fn(() => 'admin-test') }))
const now = 1790989200000
const claims = extra => ({ purpose: 'management_traits_paper', participant_type: 'candidate', participant_id: 'candidate-1', exam_id: 'exam-1', paper_id: 'paper-1', exp: now / 1000 + 120, ...extra })
const jwt = (extra = {}, alg = 'HS512') => `${btoa(JSON.stringify({ alg, typ: 'JWT' }))}.${btoa(JSON.stringify(claims(extra)))}.test-signature`
const binding = { examId: 'exam-1', participantId: 'candidate-1', paperId: 'paper-1' }
const issued = extra => ({ code: 0, success: true, msg: '', data: { id: 'candidate-1', examId: 'exam-1', paperId: 'paper-1', name: 'Candidate', participantToken: jwt(), ...extra } })
const paper = extra => ({ paperId: 'paper-1', examId: 'exam-1', title: 'TEST', paperToken: jwt({ exp: now / 1000 + 3300 }), state: 1, answered: 0, serverTime: new Date(now).toISOString(), startedAt: new Date(now).toISOString(), deadline: new Date(now + 1500000).toISOString(), reminderAt: new Date(now + 1200000).toISOString(), questions: Array.from({ length: 140 }, (_, i) => ({ id: `q-${i}`, displayOrder: i + 1, content: `Question ${i}`, selectedOptionId: null, options: Array.from({ length: 5 }, (_, j) => ({ id: `o-${i}-${j}`, content: `Option ${j}`, displayOrder: j + 1 })) })), ...extra })
const client = () => transport
const flush = async () => { for (let i = 0; i < 20; i++) await Promise.resolve() }
const deferred = () => { let resolve; let reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
function load(file, dependencies = {}) {
  const source = fs.readFileSync(path.resolve(process.cwd(), file), 'utf8')
  const sfc = file.endsWith('.vue') ? compiler.parseComponent(source) : null
  const code = babel.transformSync(sfc ? sfc.script.content : source, { babelrc: false, configFile: false, plugins: ['@babel/plugin-transform-modules-commonjs'] }).code
  const module = { exports: {} }
  new Function('require', 'module', 'exports', code)(name => {
    if (name === '@/utils/managementTraitsProduct') return product
    if (dependencies[name]) return dependencies[name]
    if (name === '@/api/managementTraits') return realApi
    if (name === '@/utils/managementTraitsResume') return helper()
    if (name.endsWith('.vue')) return { __esModule: true, default: load('src/views/exam/exam/managementTraitsResumeDialog.vue') }
    throw new Error(`Unmapped dependency ${name}`)
  }, module, module.exports)
  return sfc ? { ...module.exports.default, ...compiler.compileToFunctions(sfc.template.content) } : module.exports
}
const helper = () => load('src/utils/managementTraitsResume.js')
const store = (id = 1, permissions = []) => ({ state: { user: { id }, app: { device: 'desktop' } }, getters: { permissions } })
const router = () => ({ replace: vi.fn(() => Promise.resolve()), resolve: vi.fn(target => ({ href: `#/exam/management-traits/start/${target.params.paperId}?${new URLSearchParams(target.query)}` })) })
const wrappers = []
function mount(file, extra = {}, dependencies = {}) {
  const w = shallowMount(load(file, dependencies), { directives: { loading: () => {} }, propsData: { examId: 'exam-1' },
    mocks: { $store: store(), $route: { name: 'ManagementTraitsExam', params: { paperId: 'paper-1', examId: 'exam-1' }, query: {} }, $router: router(), $message: { error: vi.fn(), warning: vi.fn(), success: vi.fn() } },
    stubs: { 'el-button': { props: ['disabled', 'loading'], template: '<button :disabled="disabled || loading" @click="$emit(\'click\')"><slot /></button>' }, 'el-dialog': { props: ['visible'], template: '<section v-if="visible"><slot /><slot name="footer" /></section>' }, 'el-form': { template: '<div><slot /></div>', methods: { validate(cb) { cb(true) }, clearValidate() {} } }, 'el-form-item': { template: '<label><slot /></label>' }, 'el-input': true, 'el-alert': true, 'el-card': true, 'el-progress': true }, ...extra })
  wrappers.push(w); return w
}
beforeEach(() => { vi.spyOn(Date, 'now').mockReturnValue(now); sessionStorage.clear(); client().mockReset().mockResolvedValue({ data: issued() }); getToken.mockReturnValue('admin-test'); window.history.replaceState(null, '', '/app/?safe=1#/'); })
afterEach(() => { wrappers.splice(0).forEach(w => w.destroy()); vi.restoreAllMocks(); vi.useRealTimers() })

describe('candidate resume real API contract', () => {
  it('POSTs only the three explicit IDs with the admin Bearer header', async () => {
    expect(await realApi.issueManagementTraitsCandidateResume({ ...binding, telephone: 'drop', token: 'drop' })).toEqual(issued())
    expect(client().mock.calls[0][0]).toMatchObject({ method: 'post', url: '/exam/api/management-traits/admin/candidate/resume', data: binding, headers: { Authorization: 'Bearer admin-test' } })
    expect(client().mock.calls[0][0].params).toBeUndefined()
  })
  it.each(['examId', 'participantId', 'paperId'])('rejects missing %s before HTTP', async key => { await expect(realApi.issueManagementTraitsCandidateResume({ ...binding, [key]: '' })).rejects.toThrow(); expect(client()).not.toHaveBeenCalled() })
  it('does not issue without admin authentication', async () => { getToken.mockReturnValue(''); await expect(realApi.issueManagementTraitsCandidateResume(binding)).rejects.toThrow(); expect(client()).not.toHaveBeenCalled() })
  it('keeps the controlled expiry409 and no credential response', async () => { client().mockRejectedValue({ response: { status: 409, data: { code: 1, success: false, msg: '试卷已到期', data: null } } }); await expect(realApi.issueManagementTraitsCandidateResume(binding)).rejects.toMatchObject({ status: 409, message: '试卷已到期' }) })
})

describe('fragment-only short candidate credential', () => {
  it('builds the existing hash route using current origin/path, dropping HTTP search', () => {
    const r = router(); const link = helper().buildManagementTraitsResumeLink(r, issued(), binding)
    const url = new URL(link); expect(url.origin).toBe(window.location.origin); expect(url.pathname).toBe('/app/'); expect(url.search).toBe(''); expect(url.hash).toContain('resumeToken='); expect(url.hash).toContain('candidate-1'); expect(r.resolve.mock.calls[0][0].name).toBe('ManagementTraitsExam'); expect(sessionStorage.length).toBe(0)
  })
  it.each([{ id: 'other' }, { examId: 'other' }, { paperId: 'other' }, { extra: 'forbidden' }, { participantToken: jwt({ participant_id: 'other' }) }])('rejects mismatched/five-field response %j', extra => { expect(() => helper().buildManagementTraitsResumeLink(router(), issued(extra), binding)).toThrow() })
  it.each([{ purpose: 'management_traits_participant' }, { participant_type: 'tester' }, { exam_id: 'other' }, { paper_id: 'other' }, { participant_id: 'other' }, { exp: now / 1000 }, { exp: now / 1000 + 301 }, { exp: '1790989320' }, { exp: null }])('rejects scope/expiry %j', extra => { expect(() => helper().validateManagementTraitsResumeToken(jwt(extra), binding)).toThrow() })
  it('rejects non-HS512 and malformed JWTs', () => { for (const token of [jwt({}, 'HS256'), 'opaque', jwt().replace('.test-signature', '.')]) expect(() => helper().validateManagementTraitsResumeToken(token, binding)).toThrow() })
  it('awaits fragment replace before returning a usable credential', async () => {
    const pending = deferred(); const r = router(); r.replace.mockReturnValue(pending.promise)
    const route = { name: 'ManagementTraitsExam', params: { paperId: 'paper-1' }, query: { ...binding, resumeToken: jwt(), keep: '1' } }; delete route.query.paperId
    const action = helper().consumeManagementTraitsResumeRoute(route, r); let consumed = false; action.then(() => { consumed = true }); await flush(); expect(consumed).toBe(false); expect(r.replace.mock.calls[0][0]).toEqual({ name: 'ManagementTraitsExam', params: { paperId: 'paper-1' }, query: { keep: '1' } }); pending.resolve(); expect((await action).token).toBe(jwt()); expect(sessionStorage.length).toBe(0)
  })
  it('scrubs invalid credentials too and fails closed if replace fails', async () => {
    const r = router(); const route = { params: { paperId: 'paper-1' }, query: { examId: 'exam-1', participantId: 'candidate-1', resumeToken: 'opaque' } }
    await expect(helper().consumeManagementTraitsResumeRoute(route, r)).rejects.toThrow(); expect(r.replace).toHaveBeenCalledTimes(1)
    r.replace.mockRejectedValue(new Error('navigation failed')); await expect(helper().consumeManagementTraitsResumeRoute({ ...route, query: { ...route.query, resumeToken: jwt() } }, r)).rejects.toThrow(); expect(sessionStorage.length).toBe(0)
  })
})

describe('explicit admin dialog', () => {
  const file = 'src/views/exam/exam/managementTraitsResumeDialog.vue'
  it('does not issue on mount/open, allows manual IDs with no submitted result rows', async () => { const w = mount(file); w.vm.open(); await flush(); expect(client()).not.toHaveBeenCalled(); expect(w.text()).toContain('签发续答链接'); expect(w.vm.visible).toBe(true); expect(w.vm.form).toEqual({ participantId: '', paperId: '' }) })
  it.each([store('1'), store(2, ['exam:list']), { state: { user: { id: 2, roles: ['admin'] } }, getters: { permissions: [] } }])('refuses non-admin privilege', async value => { const w = mount(file, { mocks: { $store: value, $router: router() } }); w.vm.open(); await w.vm.confirmIssue(); expect(client()).not.toHaveBeenCalled(); expect(w.vm.visible).toBe(false) })
  it('uses the real wildcard predicate', () => { const w = mount(file, { mocks: { $store: store(2, ['*:*:*']), $router: router() } }); w.vm.open(); expect(w.vm.visible).toBe(true) })
  it('validates blank IDs without issuing', async () => { const w = mount(file); w.vm.open(); await w.vm.confirmIssue(); expect(client()).not.toHaveBeenCalled(); expect(w.vm.error).not.toBe('') })
  it('issues only on confirmation, prevents duplicates, exposes local expiry not storage', async () => {
    const pending = deferred(); client().mockReturnValue(pending.promise); const w = mount(file); w.vm.open(); w.vm.form = { participantId: 'candidate-1', paperId: 'paper-1' }
    await flush(); const first = w.vm.confirmIssue(); await flush(); await w.vm.confirmIssue(); expect(client()).toHaveBeenCalledTimes(1); expect(w.vm.loading).toBe(true); pending.resolve({ data: issued() }); await first; expect(w.vm.link).toContain('#/exam/management-traits/start/paper-1'); expect(w.vm.expiresAt).toBe(now + 120000); expect(sessionStorage.length).toBe(0)
    w.vm.now = now + 120000; expect(w.vm.linkExpired).toBe(true); expect(w.vm.copyDisabled).toBe(true)
  })
  it('clears stale link on error but retains manual IDs', async () => { const w = mount(file); w.vm.open(); w.vm.form = { participantId: 'candidate-1', paperId: 'paper-1' }; await flush(); await w.vm.confirmIssue(); expect(w.vm.link).not.toBe(''); client().mockRejectedValue({ response: { status: 409, data: { msg: '到期', data: null } } }); await w.vm.confirmIssue(); expect(w.vm.link).toBe(''); expect(w.vm.error).toContain('试卷已到期'); expect(w.vm.form.participantId).toBe('candidate-1'); expect(w.vm.loading).toBe(false) })
  it('closing/destroying clears IDs/token and ignores a late response', async () => { const pending = deferred(); client().mockReturnValue(pending.promise); const w = mount(file); w.vm.open(); w.vm.form = { participantId: 'candidate-1', paperId: 'paper-1' }; await flush(); const action = w.vm.confirmIssue(); await flush(); expect(client()).toHaveBeenCalledTimes(1); w.vm.close(); pending.resolve({ data: issued() }); await action; expect(w.vm.link).toBe(''); expect(w.vm.form.participantId).toBe(''); w.destroy(); expect(w.vm.link).toBe('') })
  it('timer expiry erases the credential and stops its interval', async () => {
    vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] }); const w = mount(file); w.vm.open(); w.vm.form = { participantId: 'candidate-1', paperId: 'paper-1' }; await flush(); await w.vm.confirmIssue(); expect(w.vm.link).not.toBe(''); Date.now.mockReturnValue(now + 120000); vi.advanceTimersByTime(1000); expect(w.vm.link).toBe(''); expect(w.vm.timer).toBe(null); expect(w.vm.linkExpired).toBe(true); expect(w.vm.copyDisabled).toBe(true)
  })
  it('results page installs the dedicated component without depending on completed rows', () => { const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/managementTraitsResults.vue'), 'utf8'); expect(source).toContain('<management-traits-resume-dialog'); expect(source).toContain(':exam-id="examId"'); expect(source).toContain('v-if="allowed"') })
})

describe('participant resume consumption and binding', () => {
  const file = 'src/views/paper/exam/managementTraitsExam.vue'
  const route = token => ({ name: 'ManagementTraitsExam', params: { paperId: 'paper-1' }, query: { examId: 'exam-1', participantId: 'candidate-1', resumeToken: token } })
  const apis = () => ({ ...realApi, fetchManagementTraitsPaper: vi.fn().mockResolvedValue({ data: paper() }), saveManagementTraitsAnswer: vi.fn().mockResolvedValue({ data: { answered: 1 } }), submitManagementTraitsPaper: vi.fn() })
  it('scrubs BEFORE HTTP then binds and stores only the short token, never creates/saves profile', async () => {
    const r = router(); const pending = deferred(); r.replace.mockReturnValueOnce(pending.promise); const api = apis(); const w = mount(file, { mocks: { $route: route(jwt()), $router: r } }, { '@/api/managementTraits': api }); await flush(); expect(api.fetchManagementTraitsPaper).not.toHaveBeenCalled(); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(null)
    pending.resolve(); await flush(); expect(api.fetchManagementTraitsPaper).toHaveBeenCalledWith('paper-1', jwt()); expect(w.vm.paper).not.toBe(null); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(jwt()); expect(w.vm.paper.paperToken).toBe(jwt()); await w.vm.handleAnswer('o-0-0'); expect(api.saveManagementTraitsAnswer.mock.calls[0][3]).toBe(jwt()); expect(client()).not.toHaveBeenCalled(); expect(getToken).not.toHaveBeenCalled()
  })
  it.each([paper({ examId: 'other' }), paper({ paperId: 'other' }), paper({ paperToken: jwt({ participant_id: 'other' }) }), paper({ paperToken: jwt({ participant_type: 'tester' }) }), paper({ paperToken: '' }), paper({ deadline: new Date(now + 60000).toISOString() })])('rejects returned binding/deadline before storage', async data => {
    const api = apis(); api.fetchManagementTraitsPaper.mockResolvedValue({ data }); const w = mount(file, { mocks: { $route: route(jwt()), $router: router() } }, { '@/api/managementTraits': api }); await flush(); expect(w.vm.paper).toBe(null); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(null); expect(w.vm.error).not.toBe('')
  })
  it.each([jwt({ exp: now / 1000 }), jwt({ participant_type: 'tester' }), 'opaque'])('scrubs invalid resume token and never falls back to stored credentials', async token => {
    sessionStorage.setItem('managementTraitsPaperToken', jwt()); const api = apis(); const r = router(); const w = mount(file, { mocks: { $route: route(token), $router: r } }, { '@/api/managementTraits': api }); await flush(); expect(r.replace).toHaveBeenCalled(); expect(api.fetchManagementTraitsPaper).not.toHaveBeenCalled(); expect(w.vm.paper).toBe(null)
  })
  it('fails closed when scrub fails without sending HTTP', async () => { const api = apis(); const r = router(); r.replace.mockRejectedValue(new Error('replace failed')); mount(file, { mocks: { $route: route(jwt()), $router: r } }, { '@/api/managementTraits': api }); await flush(); expect(api.fetchManagementTraitsPaper).not.toHaveBeenCalled() })
  it('uses the actual participant API with only header/body, no admin/query credential', async () => {
    client().mockResolvedValue({ data: { code: 0, success: true, data: paper() } }); mount(file, { mocks: { $route: route(jwt()), $router: router() } }); await flush(); const config = client().mock.calls[0][0]; expect(config).toMatchObject({ url: '/exam/api/management-traits/participant/paper-detail', data: { paperId: 'paper-1' }, headers: { 'X-Management-Traits-Token': jwt() } }); expect(config.params).toBeUndefined(); expect(config.url).not.toContain(jwt()); expect(getToken).not.toHaveBeenCalled()
  })
  it('reload of scrubbed route preserves resume scope and does not upgrade token', async () => {
    const api = apis(); const w = mount(file, { mocks: { $route: route(jwt()), $router: router() } }, { '@/api/managementTraits': api }); await flush(); w.destroy(); const next = mount(file, {}, { '@/api/managementTraits': api }); await flush(); expect(next.vm.paper).not.toBe(null); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(jwt())
  })
  it('local credential expiry blocks further answer/detail/submit even before frozen deadline', async () => {
    const api = apis(); const w = mount(file, { mocks: { $route: route(jwt()), $router: router() } }, { '@/api/managementTraits': api }); await flush(); Date.now.mockReturnValue(now + 120000); w.vm.tick(); await w.vm.handleAnswer('o-0-0'); await w.vm.submit(); await w.vm.loadPaper(); expect(api.saveManagementTraitsAnswer).not.toHaveBeenCalled(); expect(api.submitManagementTraitsPaper).not.toHaveBeenCalled(); expect(api.fetchManagementTraitsPaper).toHaveBeenCalledTimes(1); expect(w.vm.error).toContain('续答'); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(null)
  })
  it('a fresh fragment on the same component scrubs and resumes after local expiry', async () => {
    const api = apis(); const r = router(); const w = mount(file, { mocks: { $route: route(jwt()), $router: r } }, { '@/api/managementTraits': api }); await flush(); w.vm.$route = { ...route(jwt()), query: {} }; await flush(); Date.now.mockReturnValue(now + 120000); w.vm.tick(); expect(w.vm.paper).toBe(null)
    const fresh = jwt({ exp: now / 1000 + 240 }); const pending = deferred(); r.replace.mockReturnValueOnce(pending.promise); const previous = w.vm.$route; w.vm.$route = route(fresh); w.vm.$options.watch.$route.call(w.vm, w.vm.$route, previous); await flush(); expect(api.fetchManagementTraitsPaper).toHaveBeenCalledTimes(1); expect(r.replace).toHaveBeenCalledTimes(2); pending.resolve(); await flush(); expect(api.fetchManagementTraitsPaper).toHaveBeenLastCalledWith('paper-1', fresh); expect(w.vm.paper).not.toBe(null); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(fresh)
  })
  it('a late detail response cannot overwrite a newer resume route', async () => {
    const api = apis(); const pending = deferred(); api.fetchManagementTraitsPaper.mockReturnValueOnce(pending.promise); const w = mount(file, { mocks: { $route: route(jwt()), $router: router() } }, { '@/api/managementTraits': api }); await flush(); const fresh = jwt({ exp: now / 1000 + 180 }); const previous = w.vm.$route; w.vm.$route = route(fresh); w.vm.$options.watch.$route.call(w.vm, w.vm.$route, previous); await flush(); expect(api.fetchManagementTraitsPaper).toHaveBeenCalledTimes(2); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(fresh); pending.resolve({ data: paper({ examId: 'other' }) }); await flush(); expect(w.vm.paper.examId).toBe('exam-1'); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(fresh)
  })
  it('real hash-router admin issue → candidate consume → same-instance renewal stays fragment-only', async () => {
    const localVue = createLocalVue(); localVue.use(VueRouter)
    const r = new VueRouter({ mode: 'hash', routes: [{ path: '/exam/management-traits/start/:paperId', name: 'ManagementTraitsExam' }] })
    const response = await realApi.issueManagementTraitsCandidateResume(binding)
    const link = helper().buildManagementTraitsResumeLink(r, response, binding)
    await r.push(new URL(link).hash.slice(1))
    client().mockResolvedValue({ data: { code: 0, success: true, data: paper() } })
    const w = mount(file, { localVue, router: r, mocks: { $message: { warning: vi.fn(), error: vi.fn() } } }); await flush()
    expect(r.currentRoute.query.resumeToken).toBeUndefined(); expect(window.location.hash).not.toContain('resumeToken'); expect(w.vm.paper).not.toBe(null)
    expect(client().mock.calls[1][0]).toMatchObject({ url: '/exam/api/management-traits/participant/paper-detail', headers: { 'X-Management-Traits-Token': jwt() } })
    Date.now.mockReturnValue(now + 120000); w.vm.tick(); const fresh = jwt({ exp: now / 1000 + 240 }); await r.push(route(fresh)); await flush(); expect(w.vm.paper).not.toBe(null); expect(r.currentRoute.query.resumeToken).toBeUndefined(); expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(fresh)
    expect(client().mock.calls[2][0].headers).toEqual({ 'X-Management-Traits-Token': fresh }); expect(client().mock.calls.every(([config]) => config.params === undefined)).toBe(true)
  })
})
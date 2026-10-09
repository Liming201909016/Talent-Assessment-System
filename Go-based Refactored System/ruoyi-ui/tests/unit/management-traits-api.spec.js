import { beforeEach, describe, expect, it, vi } from 'vitest'
import axios from 'axios'
import { getToken } from '@/utils/auth'
import * as api from '@/api/managementTraits'

vi.mock('axios', () => ({ default: { create: vi.fn(() => vi.fn()) } }))
vi.mock('@/utils/auth', () => ({ getToken: vi.fn() }))
const client = () => axios.create.mock.results[0].value
const jwt = (purpose, extra = {}) => `header.${btoa(JSON.stringify({ purpose, exam_id: 'exam-1', participant_type: 'candidate', participant_id: 'person-1', paper_id: purpose === 'management_traits_paper' ? 'paper-1' : '', ...extra }))}.signature`
beforeEach(() => { client().mockReset(); getToken.mockReset().mockReturnValue('admin-test'); sessionStorage.clear() })

describe('management traits isolated real API contract', () => {
  it('MT-ADMIN-DIRECT repeated read-only config uses isolated client without global repeat-submit state', async () => {
    client().mockResolvedValue({ data: { code: 0, success: true, data: { id: 'exam-1', managementTraitsProfileFrozen: true, requiredFields: 'name' } } })
    for (let i = 0; i < 2; i++) expect((await api.fetchManagementTraitsExamConfig('exam-1')).data.id).toBe('exam-1')
    expect(client()).toHaveBeenCalledTimes(2)
    expect(client().mock.calls[0][0]).toEqual({ method: 'post', url: '/exam/api/exam/exam/detail', data: { id: 'exam-1' } })
    expect(getToken).not.toHaveBeenCalled()
  })
  it.each([
    ['fetchManagementTraitsReissueQualification', 'get', '/report-reissues/qualification', { params: { runId: 'exam-1' } }],
    ['fetchManagementTraitsReissues', 'get', '/report-reissues', { params: { paperId: 'exam-1' } }],
    ['generateManagementTraitsReissue', 'post', '/report-reissues/generate', { data: { runId: 'exam-1' } }]
  ])('MT-ADMIN-UI %s uses independent exact contract', async (name, method, suffix, fields) => {
    const body = { code: 0, success: true, data: { report: { id: 'report-1' }, reused: false } }
    client().mockResolvedValue({ data: body }); expect(await api[name]('exam-1')).toEqual(body)
    expect(client()).toHaveBeenCalledWith(expect.objectContaining({ method, url: `/exam/api/management-traits${suffix}`, ...fields }))
  })
  it.each(['viewManagementTraitsReissue', 'downloadManagementTraitsReissue'])('MT-ADMIN-UI %s preserves PDF and RFC5987 filename', async name => {
    const blob = new Blob(['%PDF-1.7'], { type: 'application/pdf' })
    client().mockResolvedValue({ data: blob, headers: { 'content-disposition': "attachment; filename*=UTF-8''%E6%B5%8B%E8%AF%95%20TEST.pdf" } })
    expect(await api[name]('report-1')).toEqual({ blob, filename: '测试 TEST.pdf' })
    expect(client().mock.calls[0][0]).toMatchObject({ url: expect.stringContaining('/report-reissues/'), params: { reportId: 'report-1' } })
  })
  it.each([0, -1, null, undefined, '1'])('MT-ADMIN-UI positive authenticated identity required for wildcard %s', async userId => {
    client().mockResolvedValue({ data: { code: 200, user: { userId }, permissions: ['*:*:*'] } })
    await expect(api.fetchManagementTraitsAdminAccess()).rejects.toThrow()
  })
  it('creates its own client with the configured API base', () => {
    expect(axios.create).toHaveBeenCalledTimes(1)
    expect(axios.create).toHaveBeenCalledWith(expect.objectContaining({ baseURL: process.env.VUE_APP_BASE_API }))
  })
  it.each([
    ['fetchManagementTraitsProfile', 'get', '/profile/detail', { params: { examId: 'exam-1' } }],
    ['freezeManagementTraitsProfile', 'post', '/profile/freeze', { data: { examId: 'exam-1' } }],
    ['fetchManagementTraitsResults', 'post', '/results/list', { data: { examId: 'exam-1' } }],
    ['fetchManagementTraitsResult', 'get', '/results/detail', { params: { runId: 'exam-1' } }],
    ['generateManagementTraitsTestReport', 'post', '/reports/generate-test', { data: { runId: 'exam-1' } }]
  ])('%s preserves the success envelope and exact IDs', async (name, method, suffix, fields) => {
    const body = { code: 0, success: true, data: name === 'fetchManagementTraitsResults' ? [] : { id: 'real-id' } }
    client().mockResolvedValue({ data: body }); expect(await api[name]('exam-1')).toEqual(body)
    expect(client()).toHaveBeenCalledWith(expect.objectContaining({ method, url: `/exam/api/management-traits${suffix}`, headers: { Authorization: 'Bearer admin-test' }, ...fields }))
    if (name.startsWith('generate')) expect(client().mock.calls[0][0].timeout).toBeGreaterThan(90000)
  })
  it('preserves a genuinely successful null profile without inventing it on error', async () => {
    client().mockResolvedValue({ data: { code: 0, success: true, data: null } }); expect((await api.fetchManagementTraitsProfile('exam-1')).data).toBe(null)
    client().mockRejectedValue({ response: { status: 409, data: { msg: 'profile validation failed' } } }); await expect(api.fetchManagementTraitsProfile('exam-1')).rejects.toThrow('profile validation failed')
  })
  it.each([[0, '0'], [1, '1']])('sets customer-controlled 005 state %s with the exact dedicated contract', async (state, wire) => {
    const body = { code: 0, success: true, data: { examId: 'exam-1', state } }
    client().mockResolvedValue({ data: body })
    expect(await api.setManagementTraitsExamState('exam-1', state)).toEqual(body)
    expect(client()).toHaveBeenCalledWith(expect.objectContaining({ method: 'post', url: '/exam/api/management-traits/admin/exam/state', data: { examId: 'exam-1', state: wire }, headers: { Authorization: 'Bearer admin-test' } }))
  })
  it.each([-1, 2, '0', null])('rejects invalid 005 state %s before HTTP', async state => {
    await expect(api.setManagementTraitsExamState('exam-1', state)).rejects.toThrow()
    expect(client()).not.toHaveBeenCalled()
  })
  it.each([401, 403, 404, 405, 503])('retains HTTP %s failures, never retries or returns null', async status => {
    client().mockRejectedValue({ response: { status, data: { msg: 'closed' } } })
    await expect(api.fetchManagementTraitsProfile('exam-1')).rejects.toMatchObject({ status })
    expect(client()).toHaveBeenCalledTimes(1)
  })
  it.each([{ code: 1, success: false, msg: '拒绝', data: null }, { code: 0, success: false, msg: '拒绝' }, { data: {} }])('rejects unsuccessful or malformed runtime envelope', async body => {
    client().mockResolvedValue({ data: body }); await expect(api.freezeManagementTraitsProfile('exam-1')).rejects.toThrow()
  })
  it('does not send admin requests without a token', async () => { getToken.mockReturnValue(''); await expect(api.fetchManagementTraitsResults('exam-1')).rejects.toThrow(); expect(client()).not.toHaveBeenCalled() })
  it.each([{ state: { user: { userId: 1 } }, getters: {} }, { state: { user: { id: 1 } }, getters: {} }, { state: { user: { id: 2 } }, getters: { permissions: ['*:*:*'] } }])('uses actual numeric identity or wildcard permission', store => expect(api.canManageManagementTraits(store)).toBe(true))
  it.each([{ state: { user: { userId: '1', roles: ['admin'] } }, getters: { permissions: [] } }, { state: { user: { id: 2, roles: ['admin'] } }, getters: { permissions: ['exam:list'] } }])('does not trust role names or string user IDs', store => expect(api.canManageManagementTraits(store)).toBe(false))
  it('routes by decoded purpose only, not a claimed signature verification', () => {
    expect(api.managementTraitsTokenClaims(jwt('management_traits_participant')).purpose).toBe('management_traits_participant')
    expect(api.managementTraitsTokenClaims('opaque')).toBe(null)
    expect(api.managementTraitsTokenClaims(jwt('competency_participant'))).toBe(null)
  })
  it.each([
    ['createManagementTraitsPaper', ['exam-1'], 'management_traits_participant', '/participant/create-paper', { examId: 'exam-1' }],
    ['fetchManagementTraitsPaper', ['paper-1'], 'management_traits_paper', '/participant/paper-detail', { paperId: 'paper-1' }],
    ['saveManagementTraitsAnswer', ['paper-1', 'pq-1', 'option-3'], 'management_traits_paper', '/participant/fill-answer', { paperId: 'paper-1', paperQuestionId: 'pq-1', optionId: 'option-3' }],
    ['submitManagementTraitsPaper', ['paper-1'], 'management_traits_paper', '/participant/submit', { paperId: 'paper-1', submitType: 'manual' }]
  ])('%s sends only its strict body and dedicated header', async (name, args, purpose, suffix, data) => {
    const token = jwt(purpose); client().mockResolvedValue({ data: { code: 0, success: true, data: {} } }); await api[name](...args, token)
    expect(client()).toHaveBeenCalledWith(expect.objectContaining({ url: `/exam/api/management-traits${suffix}`, method: 'post', data, headers: { 'X-Management-Traits-Token': token } }))
    expect(getToken).not.toHaveBeenCalled()
  })
  it.each(['', 'opaque', jwt('competency_paper'), jwt('management_traits_participant')])('rejects missing/wrong-purpose paper token before HTTP', async token => {
    await expect(api.fetchManagementTraitsPaper('paper-1', token)).rejects.toThrow(); expect(client()).not.toHaveBeenCalled()
  })
  it('rejects foreign paper, multi-select and numeric option before HTTP', async () => {
    await expect(api.fetchManagementTraitsPaper('other-paper', jwt('management_traits_paper'))).rejects.toThrow()
    for (const option of [[], ['a', 'b'], 3, '']) await expect(api.saveManagementTraitsAnswer('paper-1', 'pq-1', option, jwt('management_traits_paper'))).rejects.toThrow()
    expect(client()).not.toHaveBeenCalled()
  })
  it('keeps opaque paper markers after clearing only management tokens', () => {
    sessionStorage.setItem('competencyPaperToken', 'untouched'); api.rememberManagementTraitsPaper({ paperId: 'paper-1', examId: 'exam-1', paperToken: jwt('management_traits_paper') })
    api.clearManagementTraitsTokens(); expect(api.isManagementTraitsPaper('paper-1')).toBe(true); expect(sessionStorage.getItem('competencyPaperToken')).toBe('untouched')
    expect(sessionStorage.getItem('managementTraitsPaperToken')).toBe(null)
  })
  it('constructs only configured candidate fields and sends no model or default stuFlag', async () => {
    client().mockResolvedValue({ data: { code: 0, success: true, data: {} } })
    await api.registerManagementTraitsCandidate('exam-1', ['name', 'age'], { name: 'Test', age: 30, telephone: 'secret', stuFlag: 1, pdfFlag: 1 })
    expect(client().mock.calls[0][0]).toMatchObject({ url: '/exam/api/candidate/save', data: { examId: 'exam-1', name: 'Test', age: '30' } })
  })
  it('rejects an unknown configured field rather than sending it', async () => { await expect(api.registerManagementTraitsCandidate('exam-1', ['unknown'], {})).rejects.toThrow(); expect(client()).not.toHaveBeenCalled() })
  // MT-CANDIDATE-FIELDS: only configured identity keys, never hidden legacy defaults.
  it.each([['name', 'telephone'], ['name', 'gender', 'telephone'], ['age', 'degree', 'major', 'stuFlag']].map(fields => [fields]))('strict candidate payload for %j drops all other form properties', async fields => {
    client().mockResolvedValue({ data: { code: 0, success: true, data: {} } })
    const values = { name: 'Synthetic', telephone: '13800000000', gender: '0', age: '', degree: null, major: '', stuFlag: '0', idNumber: null, depart: '', grade: 'old', professionStatus: 'old', examId: 'foreign', pdfFlag: 1 }
    await api.registerManagementTraitsCandidate('exam-1', fields, values)
    const expected = { examId: 'exam-1' }; fields.forEach(field => { expected[field] = values[field] == null ? '' : String(values[field]) })
    expect(client().mock.calls[0][0].data).toEqual(expected); expect(getToken).not.toHaveBeenCalled()
  })
  it.each(['viewManagementTraitsTestReport', 'downloadManagementTraitsTestReport'])('%s accepts only PDF header and explicit report ID', async name => {
    const blob = new Blob(['%PDF-1.7\nbody'], { type: 'application/pdf' }); client().mockResolvedValue({ data: blob }); expect(await api[name]('report-1')).toBe(blob)
    expect(client().mock.calls[0][0]).toMatchObject({ params: { reportId: 'report-1' }, responseType: 'blob' })
  })
  it.each([new Blob(['{"msg":"not ready"}'], { type: 'application/json' }), new Blob(['{"msg":"not ready"}'], { type: 'application/pdf' }), new Blob(['%PDF'], { type: 'application/pdf' })])('rejects JSON/error/truncated PDF Blob without saving', async blob => {
    client().mockResolvedValue({ data: blob }); await expect(api.downloadManagementTraitsTestReport('report-1')).rejects.toThrow()
  })
})
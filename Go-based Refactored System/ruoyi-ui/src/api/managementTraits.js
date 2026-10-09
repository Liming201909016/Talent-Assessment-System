import axios from 'axios'
import { getToken } from '@/utils/auth'

// Deliberately independent of the admin request/401/logout interceptors.
const client = axios.create({ baseURL: process.env.VUE_APP_BASE_API, timeout: 30000 })
const root = '/exam/api/management-traits'
const participantPurpose = 'management_traits_participant'
const paperPurpose = 'management_traits_paper'
const identityFields = ['name', 'gender', 'telephone', 'affiliation', 'post', 'age', 'degree', 'major', 'stuFlag']

function failure(message, status) {
  const error = new Error(message || '管理特质请求失败，请重试或联系管理员。')
  error.status = status
  return error
}
function readBlob(blob) {
  if (typeof blob.text === 'function') return blob.text()
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = () => reject(failure('无法读取报告响应'))
    reader.readAsText(blob)
  })
}
async function request(config, identity = false) {
  let response
  try { response = await client(config) }
  catch (error) {
    const status = error.response && error.response.status
    let body = error.response && error.response.data
    if (body instanceof Blob) {
      try { body = JSON.parse(await readBlob(body)) } catch (_) { body = null }
    }
    const unavailable = status === 404 || status === 405 ? '管理特质接口未提供或已关闭，请联系管理员。' : ''
    throw failure(unavailable || (body && body.msg) || '管理特质请求失败，请重试或重新登录。', status)
  }
  const body = response.data
  if (!body || !Object.prototype.hasOwnProperty.call(body, 'data') ||
      (identity ? ![0, 200].includes(body.code) || body.success === false : body.code !== 0 || body.success !== true)) {
    throw failure(body && body.msg)
  }
  return body
}
function admin(config) {
  const token = getToken()
  if (!token) throw failure('管理员认证已失效，请重新登录。', 401)
  return { timeout: 30000, ...config, headers: { Authorization: `Bearer ${token}` } }
}
function id(value) { if (typeof value !== 'string' || !value.trim()) throw failure('缺少有效资源ID'); return value }

export function canManageManagementTraits(store) {
  const user = store && store.state && store.state.user
  // This store commits the real user.userId to state.user.id; do not use roles.
  const userId = user && (user.userId !== undefined ? user.userId : user.id)
  const permissions = store && store.getters && store.getters.permissions
  return userId === 1 || (Array.isArray(permissions) && permissions.includes('*:*:*'))
}

export async function fetchManagementTraitsProfile(examId) { return request(admin({ method: 'get', url: `${root}/profile/detail`, params: { examId: id(examId) } })) }
export async function freezeManagementTraitsProfile(examId) { return request(admin({ method: 'post', url: `${root}/profile/freeze`, data: { examId: id(examId) } })) }
export async function setManagementTraitsExamState(examId, state) {
  if (!Number.isInteger(state) || (state !== 0 && state !== 1)) throw failure('005 TEST状态仅允许进行中或禁用')
  return request(admin({ method: 'post', url: `${root}/admin/exam/state`, data: { examId: id(examId), state: String(state) } }))
}
export async function fetchManagementTraitsResults(examId) { return request(admin({ method: 'post', url: `${root}/results/list`, data: { examId: id(examId) } })) }
export async function issueManagementTraitsCandidateResume(input) {
  const { examId, participantId, paperId } = input || {}
  return request(admin({ method: 'post', url: `${root}/admin/candidate/resume`, data: { examId: id(examId), participantId: id(participantId), paperId: id(paperId) } }))
}
export async function fetchManagementTraitsResult(runId) { return request(admin({ method: 'get', url: `${root}/results/detail`, params: { runId: id(runId) } })) }
export async function generateManagementTraitsTestReport(runId) { return request(admin({ method: 'post', url: `${root}/reports/generate-test`, data: { runId: id(runId) }, timeout: 120000 })) }

export async function fetchManagementTraitsTemplateInfo() {
  return request(admin({ method: 'get', url: `${root}/reports/template` }))
}
export async function downloadManagementTraitsTemplate() {
  let response
  try {
    response = await client(admin({ method: 'get', url: `${root}/reports/template/download`, responseType: 'blob', timeout: 120000 }))
  } catch (error) {
    const status = error.response && error.response.status
    let body = error.response && error.response.data
    if (body instanceof Blob && body.size <= 65536) {
      try {
        body = JSON.parse(await readBlob(body))
      } catch (_) {
        body = null
      }
    }
    throw failure((body && body.msg) || '下载00501/00502共用报告模板失败。', status)
  }
  const blob = response.data
  const contentType = blob instanceof Blob ? blob.type.split(';')[0].toLowerCase() : ''
  const header = blob instanceof Blob ? await readBlob(blob.slice(0, 2)) : ''
  if (contentType !== 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' || header !== 'PK') {
    let message = '模板响应不是有效DOCX，未保存文件。'
    if (blob instanceof Blob && blob.size <= 65536) {
      try {
        message = JSON.parse(await readBlob(blob)).msg || message
      } catch (_) {
        // Keep the generic invalid-DOCX message when the response is not JSON.
      }
    }
    throw failure(message)
  }
  return blob
}
export async function uploadManagementTraitsTemplate(file) {
  const data = new FormData()
  const headers = { 'Content-Type': 'multipart/form-data' }
  data.append('file', file)
  return request(admin({ method: 'post', url: `${root}/reports/template/upload`, data, headers, timeout: 120000 }))
}

// Fresh server identity; cached Vue roles and URL hints are not authorization.
export async function fetchManagementTraitsAdminAccess() {
  let response
  try { response = await client(admin({ method: 'get', url: '/getInfo' })) }
  catch (_) { throw failure('管理员认证核验失败，请重新登录或重试。', 401) }
  const body = response && response.data
  const userId = body && body.user && body.user.userId
  if (!body || body.code !== 200 || !Number.isSafeInteger(userId) || userId <= 0 ||
      !(userId === 1 || (Array.isArray(body.permissions) && body.permissions.includes('*:*:*')))) {
    throw failure('仅有效管理员或全局权限账号可访问管理特质结果。', 403)
  }
  return body
}
// Independent archival reports; never select the old TEST current slot.
export async function fetchManagementTraitsReissueQualification(runId) { return request(admin({ method: 'get', url: `${root}/report-reissues/qualification`, params: { runId: id(runId) }, timeout: 120000 })) }
export async function fetchManagementTraitsReissues(paperId) { return request(admin({ method: 'get', url: `${root}/report-reissues`, params: { paperId: id(paperId) } })) }
export async function generateManagementTraitsReissue(runId) { return request(admin({ method: 'post', url: `${root}/report-reissues/generate`, data: { runId: id(runId) }, timeout: 120000 })) }

async function report(action, reportId, namespace = 'reports') {
  const config = admin({ method: 'get', url: `${root}/${namespace}/${action}`, params: { reportId: id(reportId) }, responseType: 'blob', timeout: 120000 })
  let response
  try { response = await client(config) }
  catch (error) {
    const status = error.response && error.response.status
    let body = error.response && error.response.data
    if (body instanceof Blob) {
      try { body = body.size <= 65536 ? JSON.parse(await readBlob(body)) : null } catch (_) { body = null }
    }
    throw failure(status === 404 || status === 405 ? '管理特质报告接口未提供或已关闭。' : (body && body.msg) || '报告读取失败，请重试或联系管理员。', status)
  }
  const blob = response.data
  if (!(blob instanceof Blob)) throw failure('报告响应不是有效PDF')
  const header = await readBlob(blob.slice(0, 5))
  if (blob.type.split(';')[0].toLowerCase() !== 'application/pdf' || header !== '%PDF-') {
    let message = '报告响应不是有效PDF，未保存文件。'
    if (blob.size <= 65536) { try { message = JSON.parse(await readBlob(blob)).msg || message } catch (_) {} }
    throw failure(message)
  }
  if (namespace === 'reports') return blob
  let filename = `management-traits-TEST-${reportId.replace(/[^a-zA-Z0-9_-]/g, '_')}.pdf`
  const disposition = response.headers && response.headers['content-disposition'] || ''
  const encoded = /filename\*=UTF-8''([^;]+)/i.exec(disposition)
  if (encoded) { try { filename = decodeURIComponent(encoded[1].trim()) } catch (_) {} }
  filename = filename.replace(/[\\/\x00-\x1f\x7f]/g, '_')
  return { blob, filename }
}
export async function viewManagementTraitsTestReport(reportId) { return report('view', reportId) }
export async function downloadManagementTraitsTestReport(reportId) { return report('download', reportId) }
export async function viewManagementTraitsReissue(reportId) { return report('view', reportId, 'report-reissues') }
export async function downloadManagementTraitsReissue(reportId) { return report('download', reportId, 'report-reissues') }

// Decoding is ONLY a routing/namespace hint. The backend verifies signature,
// expiry, purpose, owner and resource binding on every request.
export function managementTraitsTokenClaims(token) {
  try {
    if (typeof token !== 'string' || token.split('.').length !== 3) return null
    const segment = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    const value = JSON.parse(atob(segment.padEnd(Math.ceil(segment.length / 4) * 4, '=')))
    return [participantPurpose, paperPurpose].includes(value.purpose) ? value : null
  } catch (_) { return null }
}
function scopedToken(token, purpose, resourceId) {
  const claims = managementTraitsTokenClaims(token)
  const key = purpose === participantPurpose ? 'exam_id' : 'paper_id'
  if (!claims || claims.purpose !== purpose || claims[key] !== resourceId) throw failure('管理特质认证已失效，请重新登录。', 401)
  return { 'X-Management-Traits-Token': token }
}
export function rememberManagementTraitsParticipant(token, examId) {
  scopedToken(token, participantPurpose, examId)
  sessionStorage.setItem('managementTraitsParticipantToken', token)
}
export function managementTraitsParticipantToken(examId) {
  const token = sessionStorage.getItem('managementTraitsParticipantToken') || ''
  const claims = managementTraitsTokenClaims(token)
  return claims && claims.purpose === participantPurpose && claims.exam_id === examId ? token : ''
}
export function rememberManagementTraitsPaper(paper) {
  const claims = managementTraitsTokenClaims(paper.paperToken)
  scopedToken(paper.paperToken, paperPurpose, id(paper.paperId))
  if (claims.exam_id !== paper.examId) throw failure('试卷认证与测评不一致')
  sessionStorage.setItem('managementTraitsPaperToken', paper.paperToken)
  sessionStorage.setItem(`managementTraitsPaper:${paper.paperId}`, paper.examId)
  sessionStorage.setItem(`managementTraitsExam:${paper.examId}`, 'test')
}
export function rememberManagementTraitsProfile(profile) {
  if (!profile || !profile.frozenAt) throw failure('未确认TEST冻结状态')
  sessionStorage.setItem(`managementTraitsExam:${id(profile.examId)}`, 'test')
}
export function managementTraitsExamKnown(examId, route) {
  return !!examId && (managementTraitsIntent(route) || sessionStorage.getItem(`managementTraitsExam:${examId}`) === 'test')
}
export async function resumeManagementTraitsPaper(examId) {
  const token = sessionStorage.getItem('managementTraitsPaperToken') || ''
  const claims = managementTraitsTokenClaims(token)
  if (!claims || claims.purpose !== paperPurpose || claims.exam_id !== examId || !claims.paper_id) return null
  const response = await fetchManagementTraitsPaper(claims.paper_id, token)
  if (!response.data || response.data.examId !== examId || response.data.paperId !== claims.paper_id || ![1, 2].includes(response.data.state)) throw failure('恢复试卷身份不匹配')
  rememberManagementTraitsPaper(response.data)
  return response.data
}
export function isManagementTraitsPaper(paperId) {
  if (!paperId) return false
  const claims = managementTraitsTokenClaims(sessionStorage.getItem('managementTraitsPaperToken'))
  return !!sessionStorage.getItem(`managementTraitsPaper:${paperId}`) || !!(claims && claims.purpose === paperPurpose && claims.paper_id === paperId)
}
export function clearManagementTraitsTokens() {
  sessionStorage.removeItem('managementTraitsParticipantToken')
  sessionStorage.removeItem('managementTraitsPaperToken')
}
export function managementTraitsIntent(route) {
  return !!(route && ((route.query && route.query.mngTest === '1') || (route.meta && route.meta.managementTraitsTest === true)))
}
export async function createManagementTraitsPaper(examId, token) { return request({ method: 'post', url: `${root}/participant/create-paper`, data: { examId: id(examId) }, headers: scopedToken(token, participantPurpose, examId) }) }
export async function fetchManagementTraitsPaper(paperId, token) { return request({ method: 'post', url: `${root}/participant/paper-detail`, data: { paperId: id(paperId) }, headers: scopedToken(token, paperPurpose, paperId) }) }
export async function saveManagementTraitsAnswer(paperId, paperQuestionId, optionId, token) {
  return request({ method: 'post', url: `${root}/participant/fill-answer`, data: { paperId: id(paperId), paperQuestionId: id(paperQuestionId), optionId: id(optionId) }, headers: scopedToken(token, paperPurpose, paperId) })
}
export async function submitManagementTraitsPaper(paperId, token) { return request({ method: 'post', url: `${root}/participant/submit`, data: { paperId: id(paperId), submitType: 'manual' }, headers: scopedToken(token, paperPurpose, paperId) }) }

// Existing public identity/config endpoints, not a guessed profile detector.
export async function fetchManagementTraitsExamConfig(examId) { return request({ method: 'post', url: '/exam/api/exam/exam/detail', data: { id: id(examId) } }, true) }
export async function registerManagementTraitsCandidate(examId, fields, values) {
  if (!Array.isArray(fields) || fields.some(field => !identityFields.includes(field))) throw failure('身份配置不受支持，请联系管理员。')
  const data = { examId: id(examId) }
  if (values.id) data.id = id(values.id)
  fields.forEach(field => { data[field] = values[field] == null ? '' : String(values[field]) })
  const token = managementTraitsParticipantToken(examId)
  return request({ method: 'post', url: '/exam/api/candidate/save', data, headers: token ? { 'X-Management-Traits-Token': token } : {} }, true)
}
export async function loginManagementTraitsTester(examId, idNumber, password) {
  return request({ method: 'post', url: '/exam/api/tester/login', data: { examId: id(examId), idNumber, password } }, true)
}
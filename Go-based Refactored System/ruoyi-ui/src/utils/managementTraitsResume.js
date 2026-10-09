import { managementTraitsTokenClaims, rememberManagementTraitsPaper } from '@/api/managementTraits'

const resumeKey = 'managementTraitsResumeBinding'
const invalid = () => new Error('续答凭据或试卷绑定无效/已到期，请联系管理员重新签发；不会创建新试卷。')
function resource(value) { return typeof value === 'string' && !!value.trim() && value === value.trim() }
function header(token) {
  try {
    const parts = token.split('.')
    if (parts.length !== 3 || !parts.every(Boolean)) return null
    const segment = parts[0].replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(segment.padEnd(Math.ceil(segment.length / 4) * 4, '=')))
  } catch (_) { return null }
}
function boundClaims(token, binding) {
  const claims = managementTraitsTokenClaims(token)
  const meta = header(token)
  if (!meta || meta.alg !== 'HS512' || !claims || claims.purpose !== 'management_traits_paper' ||
      claims.participant_type !== 'candidate' || claims.participant_id !== binding.participantId ||
      claims.exam_id !== binding.examId || claims.paper_id !== binding.paperId ||
      ![binding.examId, binding.participantId, binding.paperId].every(resource) || !Number.isSafeInteger(claims.exp)) throw invalid()
  return claims
}

// This is a local scope/expiry check, NOT signature authentication. Every
// participant HTTP request still passes the original token to the backend.
export function validateManagementTraitsResumeToken(token, binding, now = Date.now()) {
  const claims = boundClaims(token, binding)
  if (claims.exp * 1000 <= now || claims.exp * 1000 > now + 300000) throw invalid()
  return claims
}
export function buildManagementTraitsResumeLink(router, response, binding) {
  const data = response && response.data
  const fields = ['id', 'examId', 'paperId', 'name', 'participantToken']
  if (!response || response.code !== 0 || response.success !== true || !data ||
      Object.keys(data).length !== fields.length || fields.some(key => !Object.prototype.hasOwnProperty.call(data, key)) ||
      data.id !== binding.participantId || data.examId !== binding.examId || data.paperId !== binding.paperId || typeof data.name !== 'string') throw invalid()
  validateManagementTraitsResumeToken(data.participantToken, binding)
  const resolved = router.resolve({ name: 'ManagementTraitsExam', params: { paperId: binding.paperId },
    query: { examId: binding.examId, participantId: binding.participantId, resumeToken: data.participantToken } })
  const hash = resolved.href.indexOf('#')
  if (hash < 0) throw invalid()
  // Never retain the admin page's search or accept a different host/base path.
  return `${window.location.origin}${window.location.pathname}${resolved.href.slice(hash)}`
}
export async function consumeManagementTraitsResumeRoute(route, router) {
  const query = { ...route.query }
  const token = query.resumeToken
  const binding = { examId: query.examId, participantId: query.participantId, paperId: route.params.paperId }
  delete query.resumeToken; delete query.examId; delete query.participantId
  // Await navigation before even validating: invalid tokens must be scrubbed too.
  try { await router.replace({ name: 'ManagementTraitsExam', params: { paperId: binding.paperId }, query }) }
  catch (_) { throw invalid() } // Router errors can contain the credential-bearing fullPath.
  const claims = validateManagementTraitsResumeToken(token, binding)
  return { ...binding, token, expiresAt: claims.exp * 1000 }
}
export function storedManagementTraitsResume(paperId) {
  const stored = sessionStorage.getItem(resumeKey)
  if (!stored) return null
  let binding
  try { binding = JSON.parse(stored) } catch (_) { throw invalid() }
  if (binding.paperId !== paperId) throw invalid()
  const token = sessionStorage.getItem('managementTraitsPaperToken') || ''
  const claims = validateManagementTraitsResumeToken(token, binding)
  return { ...binding, token, expiresAt: claims.exp * 1000 }
}
export function bindManagementTraitsResumePaper(paper, context) {
  validateManagementTraitsResumeToken(context.token, context)
  const claims = boundClaims(paper && paper.paperToken, context)
  const serverTime = Date.parse(paper && paper.serverTime)
  const deadline = Date.parse(paper && paper.deadline)
  if (!paper || paper.paperId !== context.paperId || paper.examId !== context.examId ||
      !Number.isFinite(serverTime) || !Number.isFinite(deadline) || claims.exp * 1000 <= serverTime ||
      context.expiresAt <= serverTime || context.expiresAt > serverTime + 300000 || context.expiresAt > deadline) throw invalid()
  // Detail may return a longer-lived token. Verify its owner, but never adopt it.
  const bound = { ...paper, paperToken: context.token }
  rememberManagementTraitsPaper(bound)
  sessionStorage.setItem(resumeKey, JSON.stringify({ examId: context.examId, participantId: context.participantId, paperId: context.paperId }))
  return bound
}
export function clearManagementTraitsResume() { sessionStorage.removeItem(resumeKey) }
// A code identifies configuration intent only. Frozen routing requires current
// server metadata; query/session flags are deliberately not accepted here.
export function classifyManagementTraitsProduct(code) {
  if (['00201', '00202'].includes(code)) return 'LEGACY002'
  if (['00501', '00502'].includes(code)) return 'NEW005'
  return 'OTHER'
}

export function isManagementTraitsProduct(code) { return classifyManagementTraitsProduct(code) !== 'OTHER' }

export function classifyManagementTraitsExam(exam, examId) {
  if (!exam || !(typeof exam.id === 'string' && exam.id.length > 0 && exam.id.trim() === exam.id || Number.isSafeInteger(exam.id) && exam.id > 0) || String(exam.id) !== String(examId)) return 'UNKNOWN'
  const codes = [exam.repoCode, ...(Array.isArray(exam.repoList) ? exam.repoList.map(r => r && (r.repoCode || r.code)) : [])].filter(c => c !== undefined && c !== '')
  const code = codes[0], family = classifyManagementTraitsProduct(code)
  if (!code || codes.some(c => !isManagementTraitsProduct(c) || c !== code && !(family === 'LEGACY002' && classifyManagementTraitsProduct(c) === 'LEGACY002' && exam.managementTraitsProfileFrozen === false && (exam.isManagementTraits === undefined || exam.isManagementTraits === false)))) return 'UNKNOWN'
  if (exam.assessmentType !== undefined && exam.assessmentType !== 'legacy' || exam.scoringMode !== undefined && exam.scoringMode !== 'legacy') return 'UNKNOWN'
  const frozen = exam.managementTraitsProfileFrozen, lifecycle = exam.managementTraitsLifecycle, mode = exam.isManagementTraits
  if (typeof frozen !== 'boolean' || (Object.prototype.hasOwnProperty.call(exam, 'isManagementTraits') && typeof mode !== 'boolean') || (Object.prototype.hasOwnProperty.call(exam, 'managementTraitsLifecycle') && !['legacy', 'draft', 'frozen'].includes(lifecycle))) return 'UNKNOWN'
  if (frozen === true && (lifecycle === undefined || lifecycle === 'frozen') && mode !== false) return family === 'NEW005' ? 'NEW005' : 'FROZEN_COMPAT002'
  if (frozen === false && lifecycle === 'draft' && mode === true) return family === 'NEW005' ? 'DRAFT005' : 'DRAFT_COMPAT002'
  if (family === 'LEGACY002' && frozen === false && (lifecycle === undefined || lifecycle === 'legacy') && (mode === undefined || mode === false)) return 'LEGACY002'
  return 'UNKNOWN'
}
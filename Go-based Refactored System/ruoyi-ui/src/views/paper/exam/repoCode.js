export function resolveExamRepoCode(row) {
  if (row && row.assessmentType === 'competency') {
    return '00401'
  }
  return (row && row.repoCode) || ''
}

export function buildExamEntryURL(row, origin) {
  const repoCode = resolveExamRepoCode(row)
  if (row.isOpen !== 1 && row.isOpen !== 2) return ''
  const examId = encodeURIComponent(row.id || '')
  const stuFlag = encodeURIComponent(String(row.stuFlag == null ? 0 : row.stuFlag))
  const encodedRepoCode = encodeURIComponent(repoCode)
  return `${origin}/exam-entry.html?examId=${examId}&stuFlag=${stuFlag}&repoCode=${encodedRepoCode}&isOpen=${row.isOpen}`
}

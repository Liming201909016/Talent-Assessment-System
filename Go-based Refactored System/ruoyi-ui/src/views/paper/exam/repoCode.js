export function resolveExamRepoCode(row) {
  if (row && row.assessmentType === 'competency') {
    return '00401'
  }
  return (row && row.repoCode) || ''
}

export function buildExamEntryURL(row, origin) {
  const repoCode = resolveExamRepoCode(row)
  if (row.isOpen === 1) {
    return `${origin}/#/my/exam/candidate/${row.id}/${row.stuFlag}/${repoCode}`
  }
  if (row.isOpen === 2) {
    return `${origin}/#/my/exam/tester/${row.id}/${repoCode}`
  }
  return ''
}

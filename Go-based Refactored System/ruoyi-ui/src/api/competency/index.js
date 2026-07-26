import request, { post } from '@/utils/request'

/** 管理员创建胜任力测评时获取 48 维度主数据。 */
export function fetchCompetencyDimensions() {
  return post('/exam/api/competency/dimensions/list', {})
}

export function updateCompetencyDimension(data) {
  return post('/exam/api/competency/dimensions/update', data)
}

export function fetchCompetencyQuestions(data) {
  return post('/exam/api/competency/questions/paging', data)
}

export function updateCompetencyQuestion(data) {
  return post('/exam/api/competency/questions/update', data)
}

export function downloadCompetencyQuestionTemplate() {
  return request({ url: '/exam/api/competency/questions/import-template', method: 'get', responseType: 'blob' })
}

export function previewCompetencyQuestions(file) {
  const data = new FormData()
  data.append('file', file)
  return request({ url: '/exam/api/competency/questions/import-preview', method: 'post', data, headers: { 'Content-Type': 'multipart/form-data' } })
}

export function importCompetencyQuestions(file, expectedHash) {
  const data = new FormData()
  data.append('file', file)
  data.append('expectedHash', expectedHash)
  return request({ url: '/exam/api/competency/questions/import', method: 'post', data, headers: { 'Content-Type': 'multipart/form-data' } })
}

export function publishCompetencyExam(examId) { return post('/exam/api/competency/exams/publish', { examId }) }
export function createCompetencyPaper(data) { return request({ url: '/exam/api/competency/participant/create-paper', method: 'post', data, headers: { isToken: false } }) }
function paperRequest(path, data, token) { return request({ url: `/exam/api/competency/participant/${path}`, method: 'post', data, headers: { isToken: false, 'X-Competency-Token': token } }) }
export function fetchCompetencyPaper(paperId, token) { return paperRequest('paper-detail', { paperId }, token) }
export function saveCompetencyAnswer(paperId, paperQuestionId, rawValue, token) { return paperRequest('fill-answer', { paperId, paperQuestionId, rawValue }, token) }
export function submitCompetencyPaper(paperId, submitType, token) { return paperRequest('submit', { paperId, submitType }, token) }
export function fetchCompetencyResults(data) { return post('/exam/api/competency/results/paging', data) }
export function fetchCompetencyResultDetail(paperId) { return post('/exam/api/competency/results/detail', { paperId }) }
export function fetchCompetencyReportData(paperId) { return request({ url: '/exam/api/competency/admin/report-data', method: 'get', params: { paperId } }) }
export function fetchCompetencyInternalReportData(paperId, token) { return request({ url: '/exam/api/competency/internal/report-data', method: 'get', params: { paperId, token }, headers: { isToken: false } }) }
export function generateCompetencyReport(data) { return post('/exam/api/competency/reports/generate', data) }
export function downloadCompetencyReport(paperId) { return request({ url: '/exam/api/competency/reports/download', method: 'get', params: { paperId }, responseType: 'blob' }) }

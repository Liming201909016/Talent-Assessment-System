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

export function downloadCompetencyQuestions() {
  return request({ url: '/exam/api/competency/questions/export', method: 'get', responseType: 'blob' })
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
export function fetchCompetencyInternalReportData(paperId, token) { return request({ url: '/exam/api/competency/internal/report-data', method: 'get', params: { paperId }, headers: { isToken: false, 'X-Internal-Token': token } }) }
export function generateCompetencyReport(data) { return post('/exam/api/competency/reports/generate', data) }

function readBlobText(blob) {
  if (typeof blob.text === 'function') return blob.text()
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result || '')
    reader.onerror = () => reject(reader.error)
    reader.readAsText(blob)
  })
}

export async function downloadCompetencyReport(paperId) {
  const blob = await request({ url: '/exam/api/competency/reports/download', method: 'get', params: { paperId }, responseType: 'blob' })
  if ((blob.type || '').toLowerCase().includes('application/pdf')) return blob

  let message = '下载胜任力报告失败'
  try {
    const payload = JSON.parse(await readBlobText(blob))
    message = payload.msg || message
  } catch (error) {
    // Keep the controlled fallback for a non-PDF response that is not JSON.
  }
  throw new Error(message)
}

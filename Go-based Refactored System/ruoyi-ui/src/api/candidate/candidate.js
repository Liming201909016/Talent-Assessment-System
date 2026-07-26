import request, { post } from '@/utils/request'
import id from "element-ui/src/locale/lang/id";

/**
 * 试卷列表
 * @param id
 */

// 删除用户
export function delCandidate(id) {
  return request({
    url: '/exam/api/candidate/' + id,
    method: 'delete'
  })
}

// 逻辑删除用户
export function logisticDelCandidate(id) {
  return request({
    url: '/exam/api/candidate/logistic/' + id,
    method: 'delete'
  })
}

//逻辑删除报告
export function logicDeletePdfByIds(id) {
  return request({
    url: '/exam/api/candidate/logicDeletePdfByIds/' + id,
    method: 'delete'
  })
}

// 批量删除用户
export function batchDelCandidate(ids) {
  return request({
    url: '/exam/api/candidate/batch',
    method: 'delete',
    data: ids
  })
}

// // 批量逻辑删除用户
// export function batchLogisticDelCandidate(ids) {
//   return request({
//     url: '/exam/api/candidate/logicDeletePdfByIds',
//     method: 'put',
//     data: ids
//   })
// }



// 修改用户
export function updateCandidate(data) {
  return request({
    url: '/exam/api/candidate',
    method: 'put',
    data: data,
  })
}

export function batchDownloadCandidatePdf(data) {
  return request.post('/exam/api/candidate/batch-download', data, {responseType: "blob", timeout: 10000000})
}

export function saveData(data) {
  return post('/exam/api/candidate/save', data)
}

export function updateData(data) {
  // console.log(telephone)
  return post('/exam/api/candidate/update', data)
}

export function updateEndTime(data) {
  // console.log(telephone)
  return post('/exam/api/candidate/end-time', data)
}

export function fetchCandidate(examId, testerId) {
  console.log(testerId)
  return post('/exam/api/candidate/info', {"examId": examId, "id": testerId})
}

export function testerInfo(data) {
  return post('/exam/api/candidate/tester-info', data)
}

export function paperResult(data) {
  return post('/exam/api/candidate/stand-score', data)
}

export function pdfPersistence(data) {
  return request.post('/exam/api/candidate/pdf-persistence', data, {headers: {'Content-Type': 'multipart/form-data'}})
}

export function pdfDownload(data) {
  return request.post('/exam/api/candidate/pdf-upload', data, {responseType: "blob", timeout: 10000000})
}

// export function pdfDownload(data) {
//   return request({
//     url: '/exam/api/candidate/pdf-upload',
//     method: 'POST',
//     params: data,
//     responseType: "blob",
//     timeout: 20000000
//   })
// }

export function getTeamScore(data) {
  return request({
    url: '/exam/api/candidate/team-score',
    method: 'GET',
    params: data
  })
}

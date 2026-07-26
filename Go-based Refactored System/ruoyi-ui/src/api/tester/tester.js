import request, {post} from '@/utils/request'

export function editEndTime(data) {
  // console.log(telephone)
  return post('/exam/api/tester/end-time', data)
}

export function batchDownloadTesterPdf(data) {
  return request.post('/exam/api/tester/batch-download', data, {responseType: "blob", timeout: 10000000})
}

export function pdfPersistence2(data) {
  return request.post('/exam/api/tester/pdf-persistence', data, {headers: {'Content-Type': 'multipart/form-data'}})
}

export function testerLogin(data) {
  return request({
    url: '/exam/api/tester/login',
    method: 'POST',
    params: data
  })
}

export function paperResult2(data) {
  return post('/exam/api/tester/stand-score', data)
}

// 查询岗位列表
export function listTester(query) {
  return request({
    url: '/exam/api/tester/tester-list',
    method: 'GET',
    params: query
  })
}

// 查询用户列表
export function getListTester(query) {
  return request({
    url: '/exam/api/tester/list',
    method: 'get',
    params: query
  })
}

// 查询用户详细
export function getTester(id) {
  return request({
    url: '/exam/api/tester/' + id,
    method: 'get'
  })
}

export function getTesterByIdNumber(idNumber, examId) {
  return request({
    url: '/exam/api/tester/idNumber/' + idNumber,
    method: 'get',
    params: {
      examId: examId
    },
    silentError: true
  })
}

// 新增用户
export function addTester(data) {
  return request({
    url: '/exam/api/tester',
    method: 'post',
    data: data
  })
}

// 修改用户
export function updateTester(data) {
  return request({
    url: '/exam/api/tester',
    method: 'put',
    data: data,
  })
}

// 删除用户
export function delTester(id) {
  return request({
    url: '/exam/api/tester/' + id,
    method: 'delete'
  })
}

// 逻辑删除用户
export function logisticDelTester(id) {
    return request({
        url: '/exam/api/tester/logistic/' + id,
        method: 'delete'
    })
}




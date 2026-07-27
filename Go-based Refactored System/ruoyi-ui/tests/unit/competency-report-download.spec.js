import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/utils/request', () => ({
  default: vi.fn(),
  post: vi.fn()
}))

import request from '@/utils/request'
import { downloadCompetencyReport, fetchCompetencyInternalReportData } from '@/api/competency'

// TestBugFB077_RejectsJsonBusinessErrorBlob
// 对应：docs/regression-tests.md #FB-077
// 复现：报告未生成时后端以 HTTP 200 返回 JSON Blob。
// 期望：下载 API 拒绝该响应并保留后端错误消息，页面不得保存伪 PDF。
describe('FB-077 competency report Blob validation', () => {
  beforeEach(() => {
    request.mockReset()
  })

  it('rejects an HTTP 200 JSON business-error Blob instead of returning a fake PDF', async () => {
    request.mockResolvedValue(new Blob([
      JSON.stringify({ code: 1, msg: '报告尚未生成', success: false })
    ], { type: 'application/json' }))

    await expect(downloadCompetencyReport('paper-missing')).rejects.toThrow('报告尚未生成')
  })

  it('returns a real PDF Blob unchanged', async () => {
    const pdf = new Blob(['%PDF-1.7\ncontent'], { type: 'application/pdf' })
    request.mockResolvedValue(pdf)

    await expect(downloadCompetencyReport('paper-ready')).resolves.toBe(pdf)
    expect(request).toHaveBeenCalledWith(expect.objectContaining({
      url: '/exam/api/competency/reports/download',
      method: 'get',
      params: { paperId: 'paper-ready' },
      responseType: 'blob'
    }))
  })
})

// TestBugFB079_InternalReportTokenUsesHeader
// 对应：docs/regression-tests.md #FB-079
// 期望：内部令牌不进入 URL/query，仅使用 X-Internal-Token 请求头。
describe('FB-079 competency internal report token transport', () => {
  it('sends the internal token only in a request header', () => {
    fetchCompetencyInternalReportData('paper-1', 'internal-secret')
    expect(request).toHaveBeenCalledWith({
      url: '/exam/api/competency/internal/report-data',
      method: 'get',
      params: { paperId: 'paper-1' },
      headers: { isToken: false, 'X-Internal-Token': 'internal-secret' }
    })
  })
})

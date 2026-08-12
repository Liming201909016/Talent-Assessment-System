import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/api/competency', () => ({
  fetchPhase1WordTemplate: vi.fn(),
  downloadPhase1WordTemplate: vi.fn(),
  uploadPhase1WordTemplate: vi.fn()
}))
vi.mock('@/utils/request', () => ({ default: { get: vi.fn(() => Promise.resolve({ data: [] })) } }))
vi.mock('@/utils/auth', () => ({ getToken: vi.fn(() => 'test-token') }))

import ReportTemplates from '@/views/exam/template/index.vue'
import {
  downloadPhase1WordTemplate,
  fetchPhase1WordTemplate,
  uploadPhase1WordTemplate
} from '@/api/competency'

// 新功能：一期胜任力Word模板与既有MBTI模板在同一报告模板页面管理。
describe('phase-one competency report template management', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads current template metadata and contract status', async () => {
    fetchPhase1WordTemplate.mockResolvedValue({ data: {
      exists: true,
      fileName: 'competency-phase1-report.docx',
      size: 568208,
      modTime: '2026-08-12 18:10:00',
      sha256: 'abc123',
      contentControls: 49,
      charts: 12,
      visibleTokens: 0,
      valid: true
    } })
    const vm = {
      ...ReportTemplates.data(),
      $message: { error: vi.fn() }
    }
    await ReportTemplates.methods.fetchPhase1Template.call(vm)
    expect(fetchPhase1WordTemplate).toHaveBeenCalled()
    expect(vm.phase1Template).toEqual(expect.objectContaining({ valid: true, contentControls: 49, charts: 12 }))
    expect(vm.phase1Loading).toBe(false)
  })

  it('downloads the active DOCX with its configured file name', async () => {
    const blob = new Blob(['docx'], { type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' })
    downloadPhase1WordTemplate.mockResolvedValue(blob)
    const click = vi.fn()
    const link = { click, style: {}, setAttribute: vi.fn() }
    vi.spyOn(document, 'createElement').mockReturnValueOnce(link)
    vi.spyOn(document.body, 'appendChild').mockImplementation(() => link)
    vi.spyOn(document.body, 'removeChild').mockImplementation(() => link)
    vi.spyOn(window.URL, 'createObjectURL').mockReturnValue('blob:template')
    vi.spyOn(window.URL, 'revokeObjectURL').mockImplementation(() => {})
    const vm = { phase1Template: { fileName: 'competency-phase1-report.docx' }, $message: { error: vi.fn() } }
    await ReportTemplates.methods.downloadPhase1Template.call(vm)
    expect(downloadPhase1WordTemplate).toHaveBeenCalled()
    expect(link.setAttribute).toHaveBeenCalledWith('download', 'competency-phase1-report.docx')
    expect(click).toHaveBeenCalled()
  })

  it('confirms, uploads, refreshes metadata and preserves dialog state on failure', async () => {
    uploadPhase1WordTemplate.mockResolvedValue({ data: { valid: true } })
    const file = { name: 'phase1.docx', size: 1024 }
    const vm = {
      ...ReportTemplates.data(),
      phase1File: file,
      $confirm: vi.fn(() => Promise.resolve()),
      $message: { success: vi.fn(), error: vi.fn() },
      fetchPhase1Template: vi.fn()
    }
    await ReportTemplates.methods.uploadPhase1Template.call(vm)
    expect(vm.$confirm).toHaveBeenCalled()
    expect(uploadPhase1WordTemplate).toHaveBeenCalledWith(file)
    expect(vm.phase1File).toBe(null)
    expect(vm.fetchPhase1Template).toHaveBeenCalled()

    uploadPhase1WordTemplate.mockRejectedValueOnce(new Error('invalid template'))
    vm.phase1File = file
    await ReportTemplates.methods.uploadPhase1Template.call(vm)
    expect(vm.phase1File).toBe(file)
    expect(vm.$message.error).toHaveBeenCalledWith('invalid template')
  })
})

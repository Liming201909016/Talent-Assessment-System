import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/api/competency', () => ({
  fetchPhase1V2WordTemplate: vi.fn(),
  downloadPhase1V2WordTemplate: vi.fn(),
  uploadPhase1V2WordTemplate: vi.fn()
}))
vi.mock('@/utils/request', () => ({ default: { get: vi.fn(() => Promise.resolve({ data: [] })) } }))
vi.mock('@/utils/auth', () => ({ getToken: vi.fn(() => 'test-token') }))

import ReportTemplates from '@/views/exam/template/index.vue'
import {
  downloadPhase1V2WordTemplate,
  fetchPhase1V2WordTemplate,
  uploadPhase1V2WordTemplate
} from '@/api/competency'

// 新功能：一期胜任力Word模板与既有MBTI模板在同一报告模板页面管理。
describe('phase-one competency report template management', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads current template metadata and contract status', async () => {
    fetchPhase1V2WordTemplate.mockResolvedValue({ data: {
      exists: true,
      fileName: 'competency-phase1-report-v2.docx',
      size: 568208,
      modTime: '2026-08-12 18:10:00',
      sha256: 'abc123',
      schemaVersion: 'competency-phase1-template-schema-v2',
      contentControls: 49,
      registeredFields: 75,
      usedFields: 49,
      charts: 12,
      businessCharts: 12,
      embeddedWorkbooks: 1,
      externalLinks: 0,
      visibleTokens: 0,
      semanticFields: [{ key: 'participant.name', name: '姓名', description: '受测者姓名', repeatable: true }],
      valid: true
    } })
    const vm = {
      ...ReportTemplates.data(),
      $message: { error: vi.fn() }
    }
    await ReportTemplates.methods.fetchPhase1Template.call(vm)
    expect(fetchPhase1V2WordTemplate).toHaveBeenCalled()
    expect(vm.phase1Template).toEqual(expect.objectContaining({ valid: true, schemaVersion: 'competency-phase1-template-schema-v2', registeredFields: 75, businessCharts: 12 }))
    expect(vm.phase1Template.semanticFields).toEqual([expect.objectContaining({ key: 'participant.name', name: '姓名' })])
    expect(vm.phase1Loading).toBe(false)
  })

  it('opens the 00401 semantic-field dialog with names and descriptions', () => {
    const vm = {
      ...ReportTemplates.data(),
      phase1Template: { semanticFields: [{ key: 'validity.text', name: '效度说明', description: '效度评价说明；可选' }] }
    }
    ReportTemplates.methods.openPhase1SemanticFields.call(vm)
    expect(vm.semanticFieldDialogVisible).toBe(true)
    expect(vm.semanticFieldDialogTitle).toContain('00401')
    expect(vm.semanticFieldRows).toEqual([expect.objectContaining({ key: 'validity.text', name: '效度说明', description: expect.any(String) })])
  })

  it('explains the transparent V2 field and chart contract', () => {
    const vm = {
      phase1Template: {
        exists: true,
        schemaVersion: 'competency-phase1-template-schema-v2',
        contentControls: 49,
        registeredFields: 75,
        usedFields: 49,
        charts: 12,
        businessCharts: 12,
        embeddedWorkbooks: 1,
        externalLinks: 0,
        visibleTokens: 0
      }
    }
    const text = ReportTemplates.computed.phase1ContractText.call(vm)
    expect(text).toContain('V2')
    expect(text).toContain('49/75 字段')
    expect(text).toContain('12 业务图表')
    expect(text).toContain('0 外链')
  })

  it('downloads the active DOCX with its configured file name', async () => {
    const blob = new Blob(['docx'], { type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' })
    downloadPhase1V2WordTemplate.mockResolvedValue(blob)
    const click = vi.fn()
    const link = { click, style: {}, setAttribute: vi.fn() }
    vi.spyOn(document, 'createElement').mockReturnValueOnce(link)
    vi.spyOn(document.body, 'appendChild').mockImplementation(() => link)
    vi.spyOn(document.body, 'removeChild').mockImplementation(() => link)
    vi.spyOn(window.URL, 'createObjectURL').mockReturnValue('blob:template')
    vi.spyOn(window.URL, 'revokeObjectURL').mockImplementation(() => {})
    const vm = {
      phase1Template: { fileName: 'competency-phase1-report-v2.docx', sha256: '3b6a83fd4a2f' },
      phase1DownloadFileName: ReportTemplates.methods.phase1DownloadFileName,
      $message: { error: vi.fn() }
    }
    await ReportTemplates.methods.downloadPhase1Template.call(vm)
    expect(downloadPhase1V2WordTemplate).toHaveBeenCalled()
    expect(link.setAttribute).toHaveBeenCalledWith('download', 'competency-phase1-report-v2-3b6a83fd.docx')
    expect(click).toHaveBeenCalled()
  })

  it('confirms, uploads, refreshes metadata and preserves dialog state on failure', async () => {
    uploadPhase1V2WordTemplate.mockResolvedValue({ data: { valid: true } })
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
    expect(uploadPhase1V2WordTemplate).toHaveBeenCalledWith(file)
    expect(vm.phase1File).toBe(null)
    expect(vm.fetchPhase1Template).toHaveBeenCalled()

    uploadPhase1V2WordTemplate.mockRejectedValueOnce(new Error('invalid template'))
    vm.phase1File = file
    await ReportTemplates.methods.uploadPhase1Template.call(vm)
    expect(vm.phase1File).toBe(file)
    expect(vm.$message.error).toHaveBeenCalledWith('invalid template')
  })
})

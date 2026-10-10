import { beforeEach, describe, expect, it, vi } from 'vitest'
import fs from 'fs'
import path from 'path'

vi.mock('@/api/competency', () => ({
  fetchPhase1V2WordTemplate: vi.fn(),
  downloadPhase1V2WordTemplate: vi.fn(),
  uploadPhase1V2WordTemplate: vi.fn()
}))
vi.mock('@/api/managementTraits', () => ({
  fetchManagementTraitsTemplateInfo: vi.fn(),
  downloadManagementTraitsTemplate: vi.fn(),
  uploadManagementTraitsTemplate: vi.fn()
}))
vi.mock('@/utils/request', () => ({ default: { get: vi.fn(() => Promise.resolve({ data: [] })) } }))
vi.mock('@/utils/auth', () => ({ getToken: vi.fn(() => 'test-token') }))

import ReportTemplates from '@/views/exam/template/index.vue'
import {
  downloadManagementTraitsTemplate,
  fetchManagementTraitsTemplateInfo,
  uploadManagementTraitsTemplate
} from '@/api/managementTraits'

describe('00501/00502 shared report template management', () => {
  beforeEach(() => vi.clearAllMocks())

  it('loads one shared template card for both products', async () => {
    fetchManagementTraitsTemplateInfo.mockResolvedValue({ data: {
      exists: true,
      fileName: 'management-traits-00501-00502-shared.docx',
      size: 530000,
      modTime: '2026-10-09 12:00:00',
      sha256: 'a'.repeat(64),
      valid: true,
      productCodes: ['00501', '00502'],
      contentControls: 90,
      businessCharts: 6,
      numericLabels: 5,
      externalLinks: 0,
      semanticFields: [{ key: 'participant.name', name: '姓名', description: '受测者姓名', repeatable: true }]
    } })
    const vm = { ...ReportTemplates.data(), $message: { error: vi.fn() } }
    await ReportTemplates.methods.fetchManagementTraitsTemplate.call(vm)
    expect(fetchManagementTraitsTemplateInfo).toHaveBeenCalledTimes(1)
    expect(vm.managementTraitsTemplate).toEqual(expect.objectContaining({ valid: true, productCodes: ['00501', '00502'], contentControls: 90 }))
    expect(vm.managementTraitsTemplate.semanticFields).toHaveLength(1)
    expect(vm.managementTraitsTemplateLoading).toBe(false)
  })

  it('opens the 00501/00502 semantic-field dialog and explains repeated tags', () => {
    const vm = {
      ...ReportTemplates.data(),
      managementTraitsTemplate: { semanticFields: [{ key: 'overall.diagnosis', name: '总体诊断', description: '总体表现诊断', repeatable: true }] }
    }
    ReportTemplates.methods.openManagementTraitsSemanticFields.call(vm)
    expect(vm.semanticFieldDialogVisible).toBe(true)
    expect(vm.semanticFieldDialogTitle).toContain('00501 / 00502')
    expect(vm.semanticFieldRows[0]).toEqual(expect.objectContaining({ repeatable: true }))
  })

  it('downloads the shared DOCX with a hash-qualified file name', async () => {
    const blob = new Blob(['PK-docx'], { type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' })
    downloadManagementTraitsTemplate.mockResolvedValue(blob)
    const click = vi.fn()
    const link = { click, style: {}, setAttribute: vi.fn() }
    vi.spyOn(document, 'createElement').mockReturnValueOnce(link)
    vi.spyOn(document.body, 'appendChild').mockImplementation(() => link)
    vi.spyOn(document.body, 'removeChild').mockImplementation(() => link)
    vi.spyOn(window.URL, 'createObjectURL').mockReturnValue('blob:mng-template')
    vi.spyOn(window.URL, 'revokeObjectURL').mockImplementation(() => {})
    const vm = {
      managementTraitsTemplate: { fileName: 'management-traits-00501-00502-shared.docx', sha256: '12345678abcdef' },
      managementTraitsTemplateDownloadFileName: ReportTemplates.methods.managementTraitsTemplateDownloadFileName,
      $message: { error: vi.fn() }
    }
    await ReportTemplates.methods.downloadManagementTraitsTemplate.call(vm)
    expect(link.setAttribute).toHaveBeenCalledWith('download', 'management-traits-00501-00502-shared-12345678.docx')
    expect(click).toHaveBeenCalledTimes(1)
  })

  it('confirms upload, refreshes on success and preserves the selected file on failure', async () => {
    const file = { name: 'shared.docx', size: 2048 }
    uploadManagementTraitsTemplate.mockResolvedValue({ data: { template: { valid: true } } })
    const vm = {
      ...ReportTemplates.data(),
      managementTraitsTemplateFile: file,
      $confirm: vi.fn(() => Promise.resolve()),
      $message: { success: vi.fn(), error: vi.fn() },
      fetchManagementTraitsTemplate: vi.fn()
    }
    await ReportTemplates.methods.uploadManagementTraitsTemplate.call(vm)
    expect(uploadManagementTraitsTemplate).toHaveBeenCalledWith(file)
    expect(vm.managementTraitsTemplateFile).toBe(null)
    expect(vm.fetchManagementTraitsTemplate).toHaveBeenCalledTimes(1)

    uploadManagementTraitsTemplate.mockRejectedValueOnce(new Error('template rejected'))
    vm.managementTraitsTemplateFile = file
    await ReportTemplates.methods.uploadManagementTraitsTemplate.call(vm)
    expect(vm.managementTraitsTemplateFile).toBe(file)
    expect(vm.$message.error).toHaveBeenCalledWith('template rejected')
  })

  it('renders one common card rather than two independent product upload areas', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/template/index.vue'), 'utf8')
    expect(source).toContain('00501 / 00502 共用报告模板')
    expect(source.match(/management-traits-card/g)?.length).toBeGreaterThan(0)
    expect(source).not.toContain('00501 独立报告模板')
    expect(source).not.toContain('00502 独立报告模板')
  })
})

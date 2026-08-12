import { describe, expect, it, vi } from 'vitest'
import CompetencyQuestionList from '@/views/qu/competency/index.vue'
import { saveAs } from 'file-saver'
import { fetchCompetencyDimensions, fetchCompetencyQuestions, updateCompetencyQuestion, downloadCompetencyQuestions, previewCompetencyQuestions, importCompetencyQuestions } from '@/api/competency'

vi.mock('file-saver', () => ({ saveAs: vi.fn() }))

vi.mock('@/api/competency', () => ({
  fetchCompetencyDimensions: vi.fn(() => Promise.resolve({ data: [] })),
  fetchCompetencyQuestions: vi.fn(() => Promise.resolve({ data: { records: [], total: 0 } })),
  updateCompetencyQuestion: vi.fn(() => Promise.resolve({ data: {} })),
  downloadCompetencyQuestions: vi.fn(() => Promise.resolve(new Blob(['xlsx']))),
  downloadCompetencyQuestionTemplate: vi.fn(() => Promise.resolve(new Blob(['xlsx']))),
  previewCompetencyQuestions: vi.fn(() => Promise.resolve({ data: { sha256: 'abc', successCount: 1, errorCount: 0, successRows: [{ questionCode: 'D01-Q09' }], errorRows: [] } })),
  importCompetencyQuestions: vi.fn(() => Promise.resolve({ data: { importedCount: 1 } }))
}))

describe('00401 competency question bank', () => {
  it('loads the dedicated competency question paging endpoint', async () => {
    const vm = { ...CompetencyQuestionList.data() }
    vm.$set = vi.fn()
    await CompetencyQuestionList.methods.loadQuestions.call(vm)
    expect(fetchCompetencyQuestions).toHaveBeenCalledWith(expect.objectContaining({ current: 1, size: 20 }))
    expect(vm.rows).toEqual([])
    expect(vm.total).toBe(0)
  })

  it('loads all competency dimensions for filtering', async () => {
    const vm = { ...CompetencyQuestionList.data() }
    await CompetencyQuestionList.methods.loadDimensions.call(vm)
    expect(fetchCompetencyDimensions).toHaveBeenCalled()
    expect(vm.dimensions).toEqual([])
  })

  it('exports all 00401 source questions as xlsx', async () => {
    downloadCompetencyQuestions.mockClear()
    saveAs.mockClear()
    const vm = { ...CompetencyQuestionList.data() }
    await CompetencyQuestionList.methods.exportQuestions.call(vm)
    expect(downloadCompetencyQuestions).toHaveBeenCalled()
    expect(saveAs).toHaveBeenCalledWith(expect.any(Blob), '00401-competency-questions.xlsx')
  })

  it('opens an editable copy while preserving identity context', () => {
    const vm = { ...CompetencyQuestionList.data() }
    const row = { id: 'q1', questionCode: 'D01-Q01', dimensionCode: 'D01', dimensionName: '沟通表达', dimensionItemNo: 1, content: '题目', observationPoint: '考察点', scoringDirection: 'forward', questionStatus: 0, remark: '备注' }
    CompetencyQuestionList.methods.openEdit.call(vm, row)
    expect(vm.editVisible).toBe(true)
    expect(vm.editForm).toEqual(expect.objectContaining({ id: 'q1', content: '题目', questionStatus: 0 }))
    expect(vm.editIdentity).toEqual(expect.objectContaining({ questionCode: 'D01-Q01', dimensionName: '沟通表达' }))
  })

  it('distinguishes dimension and validity question types in the dedicated page', () => {
    const source = CompetencyQuestionList.render.toString()
    expect(source).toContain('competencyQuestionType')
    expect(source).toContain('validity')
  })

  it('shows the ten-column import contract after adding the question-type column', () => {
    // Regression FB-104: docs/regression-tests.md
    const source = CompetencyQuestionList.render.toString()
    expect(source).toContain('专用十列模板')
    expect(source).not.toContain('专用九列模板')
  })

  it('saves editable fields and refreshes the current page', async () => {
    updateCompetencyQuestion.mockClear()
    const loadQuestions = vi.fn()
    const notify = vi.fn()
    const vm = {
      ...CompetencyQuestionList.data(),
      editVisible: true,
      editForm: { id: 'q1', content: '新题目', observationPoint: '新考察点', scoringDirection: 'reverse', questionStatus: 1, remark: '新备注' },
      $refs: { editForm: { validate: callback => callback(true) } },
      $notify: notify,
      loadQuestions
    }
    await CompetencyQuestionList.methods.saveEdit.call(vm)
    expect(updateCompetencyQuestion).toHaveBeenCalledWith(vm.editForm)
    expect(vm.editVisible).toBe(false)
    expect(loadQuestions).toHaveBeenCalled()
    expect(notify).toHaveBeenCalled()
  })

  it('rejects non-xlsx and oversized files before preview', () => {
    const message = { warning: vi.fn() }
    const vm = { ...CompetencyQuestionList.data(), $message: message }
    expect(CompetencyQuestionList.methods.selectImportFile.call(vm, { name: 'bad.csv', size: 10, raw: {} })).toBe(false)
    expect(CompetencyQuestionList.methods.selectImportFile.call(vm, { name: 'large.xlsx', size: 11 * 1024 * 1024, raw: {} })).toBe(false)
    expect(message.warning).toHaveBeenCalledTimes(2)
  })

  it('previews the selected file and retains its digest', async () => {
    previewCompetencyQuestions.mockClear()
    const vm = { ...CompetencyQuestionList.data(), importFile: { name: 'valid.xlsx' } }
    await CompetencyQuestionList.methods.previewImport.call(vm)
    expect(previewCompetencyQuestions).toHaveBeenCalledWith(vm.importFile)
    expect(vm.importPreview).toEqual(expect.objectContaining({ sha256: 'abc', successCount: 1, errorCount: 0 }))
  })

  it('formally imports only a valid preview and refreshes data', async () => {
    importCompetencyQuestions.mockClear()
    const loadQuestions = vi.fn()
    const loadDimensions = vi.fn()
    const notify = vi.fn()
    const vm = {
      ...CompetencyQuestionList.data(), importVisible: true,
      importFile: { name: 'valid.xlsx' }, importPreview: { sha256: 'abc', errorCount: 0 },
      $confirm: vi.fn(() => Promise.resolve()), $notify: notify, loadQuestions, loadDimensions
    }
    await CompetencyQuestionList.methods.confirmImport.call(vm)
    expect(importCompetencyQuestions).toHaveBeenCalledWith(vm.importFile, 'abc')
    expect(vm.importVisible).toBe(false)
    expect(loadQuestions).toHaveBeenCalled()
    expect(loadDimensions).toHaveBeenCalled()
  })
})

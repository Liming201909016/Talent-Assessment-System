import { describe, expect, it, vi } from 'vitest'
import CompetencyDimensionMaintenance from '@/views/qu/competency/dimensions.vue'
import { fetchCompetencyDimensions, updateCompetencyDimension } from '@/api/competency'

vi.mock('@/api/competency', () => ({
  fetchCompetencyDimensions: vi.fn(() => Promise.resolve({ data: [
    { id: 'd1', code: 'D01', name: '沟通表达', virdLevel: 'Versatility 胜任力', applicableCategory: '基层通用', coreMeaning: '清晰表达', displayOrder: 1, status: 0, questionCount: 8 }
  ] })),
  updateCompetencyDimension: vi.fn(() => Promise.resolve({ data: {} }))
}))

describe('competency dimension maintenance', () => {
  it('loads all dimensions and applies local filters', async () => {
    const vm = { ...CompetencyDimensionMaintenance.data() }
    await CompetencyDimensionMaintenance.methods.loadDimensions.call(vm)
    expect(fetchCompetencyDimensions).toHaveBeenCalled()
    expect(vm.rows).toHaveLength(1)
    vm.query.keyword = 'D01'
    expect(CompetencyDimensionMaintenance.computed.filteredRows.call(vm)).toHaveLength(1)
    vm.query.keyword = '不存在'
    expect(CompetencyDimensionMaintenance.computed.filteredRows.call(vm)).toHaveLength(0)
  })

  it('opens an editable copy while keeping code immutable', () => {
    const vm = { ...CompetencyDimensionMaintenance.data(), $nextTick: vi.fn() }
    const row = { id: 'd1', code: 'D01', name: '沟通表达', virdLevel: 'Versatility 胜任力', applicableCategory: '基层通用', coreMeaning: '清晰表达', displayOrder: 1, status: 0 }
    CompetencyDimensionMaintenance.methods.openEdit.call(vm, row)
    expect(vm.editVisible).toBe(true)
    expect(vm.editIdentity).toEqual({ code: 'D01' })
    expect(vm.editForm).toEqual(expect.objectContaining({ id: 'd1', name: '沟通表达', displayOrder: 1, status: 0 }))
    expect(vm.editForm.code).toBeUndefined()
  })

  it('confirms status impact, saves and refreshes the list', async () => {
    updateCompetencyDimension.mockClear()
    const loadDimensions = vi.fn()
    const vm = {
      ...CompetencyDimensionMaintenance.data(), editVisible: true,
      editForm: { id: 'd1', name: '沟通表达', virdLevel: 'Versatility 胜任力', applicableCategory: '基层通用', coreMeaning: '清晰表达', displayOrder: 1, status: 1 },
      originalStatus: 0,
      $refs: { editForm: { validate: callback => callback(true) } },
      $confirm: vi.fn(() => Promise.resolve()), $notify: vi.fn(), loadDimensions
    }
    await CompetencyDimensionMaintenance.methods.saveEdit.call(vm)
    expect(vm.$confirm).toHaveBeenCalled()
    expect(updateCompetencyDimension).toHaveBeenCalledWith(vm.editForm)
    expect(vm.editVisible).toBe(false)
    expect(loadDimensions).toHaveBeenCalled()
  })

  it('keeps dialog data when save fails', async () => {
    updateCompetencyDimension.mockRejectedValueOnce(new Error('failed'))
    const vm = {
      ...CompetencyDimensionMaintenance.data(), editVisible: true,
      editForm: { id: 'd1', name: '沟通表达', virdLevel: 'V', applicableCategory: '基层通用', coreMeaning: '含义', displayOrder: 1, status: 0 },
      originalStatus: 0,
      $refs: { editForm: { validate: callback => callback(true) } },
      $notify: vi.fn(), loadDimensions: vi.fn()
    }
    await CompetencyDimensionMaintenance.methods.saveEdit.call(vm)
    expect(vm.editVisible).toBe(true)
    expect(vm.editForm.name).toBe('沟通表达')
  })
})
import { describe, expect, it, vi } from 'vitest'
import CompetencyDimensionMaintenance from '@/views/qu/competency/dimensions.vue'
import { fetchCompetencyDimensions, updateCompetencyDimension } from '@/api/competency'

vi.mock('@/api/competency', () => ({
  fetchCompetencyDimensions: vi.fn(() => Promise.resolve({ data: [
    { id: 'competency-a1-01', code: 'A1-01', name: '逻辑思维', virdLevel: '通用能力', applicableCategory: '基层员工', coreMeaning: '逻辑分析严谨，推理判断有据', displayOrder: 1, status: 0, questionCount: 0 }
  ] })),
  updateCompetencyDimension: vi.fn(() => Promise.resolve({ data: {} }))
}))

describe('competency dimension maintenance', () => {
  it('loads all dimensions and applies local filters', async () => {
    const vm = { ...CompetencyDimensionMaintenance.data() }
    await CompetencyDimensionMaintenance.methods.loadDimensions.call(vm)
    expect(fetchCompetencyDimensions).toHaveBeenCalled()
    expect(vm.rows).toHaveLength(1)
    vm.query.keyword = 'A1-01'
    expect(CompetencyDimensionMaintenance.computed.filteredRows.call(vm)).toHaveLength(1)
    vm.query.keyword = '不存在'
    expect(CompetencyDimensionMaintenance.computed.filteredRows.call(vm)).toHaveLength(0)
  })

  it('opens an editable copy while keeping code immutable', () => {
    const vm = { ...CompetencyDimensionMaintenance.data(), $nextTick: vi.fn() }
    const row = { id: 'competency-a1-01', code: 'A1-01', name: '逻辑思维', virdLevel: '通用能力', applicableCategory: '基层员工', coreMeaning: '逻辑分析严谨，推理判断有据', displayOrder: 1, status: 0 }
    CompetencyDimensionMaintenance.methods.openEdit.call(vm, row)
    expect(vm.editVisible).toBe(true)
    expect(vm.editIdentity).toEqual({ code: 'A1-01' })
    expect(vm.editForm).toEqual(expect.objectContaining({ id: 'competency-a1-01', name: '逻辑思维', displayOrder: 1, status: 0 }))
    expect(vm.editForm.code).toBeUndefined()
  })

  it('uses the confirmed phase-1 layers, audience and ten-position range', () => {
    const state = CompetencyDimensionMaintenance.data()
    expect(state.allVirdLevels).toEqual(['通用能力', '心理素养'])
    expect(state.allApplicableCategories).toEqual(['基层员工'])
    expect(state.maxDisplayOrder).toBe(10)
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
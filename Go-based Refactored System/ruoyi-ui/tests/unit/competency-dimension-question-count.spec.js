import { shallowMount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import CompetencyDimensionSelector from '@/views/exam/exam/components/CompetencyDimensionSelector.vue'

const dimensions = [
  { id: 'd1', code: 'D01', name: '沟通表达', virdLevel: 'V', applicableCategory: '基层通用', status: 0, questionCount: 8 },
  { id: 'd2', code: 'D02', name: '人际交往', virdLevel: 'V', applicableCategory: '基层通用', status: 0, questionCount: 0 },
  { id: 'd3', code: 'D03', name: '数字应用', virdLevel: 'V', applicableCategory: '基层通用', status: 1, questionCount: 4 }
]

function mountSelector(value = ['d1']) {
  return shallowMount(CompetencyDimensionSelector, {
    propsData: { value, dimensions },
    stubs: { 'el-checkbox-group': true, 'el-checkbox': true },
    directives: { loading: () => {} }
  })
}

describe('CompetencyDimensionSelector question counts', () => {
  it('shows selected question total and exposes zero-count disable state', () => {
    const wrapper = mountSelector()
    expect(wrapper.vm.selectedQuestionCount).toBe(8)
    expect(wrapper.vm.isDimensionDisabled(dimensions[0])).toBe(false)
    expect(wrapper.vm.isDimensionDisabled(dimensions[1])).toBe(true)
    expect(wrapper.vm.isDimensionDisabled(dimensions[2])).toBe(true)
    expect(wrapper.text()).toContain('共 8 道启用题目')
  })

  it('does not count disabled or zero-count dimensions even if stale IDs are present', () => {
    const wrapper = mountSelector(['d1', 'd2', 'd3'])
    expect(wrapper.vm.selectedQuestionCount).toBe(8)
  })
})

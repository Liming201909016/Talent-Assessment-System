import { shallowMount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import CompetencyDimensionSelector from '@/views/exam/exam/components/CompetencyDimensionSelector.vue'

const dimensions = [
  { id: 'd1', code: 'D01', name: '沟通表达', virdLevel: 'Versatility 胜任力', applicableCategory: '基层通用', status: 0 },
  { id: 'd2', code: 'D11', name: '战略思维', virdLevel: 'Versatility 胜任力', applicableCategory: '管理通用', status: 0 },
  { id: 'd3', code: 'D21', name: '敬业奉献', virdLevel: 'Integrity 信念力', applicableCategory: '管理通用', status: 1 }
]

function mountSelector(propsData = {}) {
  return shallowMount(CompetencyDimensionSelector, {
    propsData: {
      value: ['d1'],
      dimensions,
      ...propsData
    },
    stubs: {
      'el-checkbox-group': true,
      'el-checkbox': true
    },
    directives: {
      loading: () => {}
    }
  })
}

describe('CompetencyDimensionSelector', () => {
  it('groups dimensions by VIRD level and preserves reference order', () => {
    const wrapper = mountSelector()
    expect(wrapper.vm.groupedDimensions).toEqual([
      { name: 'Versatility 胜任力', items: [dimensions[0], dimensions[1]] },
      { name: 'Integrity 信念力', items: [dimensions[2]] }
    ])
    expect(wrapper.vm.selectedCount).toBe(1)
  })

  it('emits v-model updates without mutating the prop', () => {
    const wrapper = mountSelector()
    wrapper.vm.selectedValues = ['d1', 'd2']
    expect(wrapper.emitted('input')[0]).toEqual([['d1', 'd2']])
    expect(wrapper.props('value')).toEqual(['d1'])
  })

  it('shows a clear empty state when dimensions have not been migrated', () => {
    const wrapper = mountSelector({ value: [], dimensions: [], loading: false })
    expect(wrapper.text()).toContain('暂无可配置的胜任力维度')
  })
})

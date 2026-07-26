// 示例：Vue 组件测试模板（演示如何用 @vue/test-utils 测试 .vue 组件）
//
// 本文件展示常见测试模式：
//  - props 渲染验证
//  - 事件触发验证
//  - 计算属性验证
//  - 用户交互模拟
//
// 复制此模式为其他组件创建测试。
import { describe, it, expect } from 'vitest'
import { mount, shallowMount } from '@vue/test-utils'
import Vue from 'vue'

// 内联定义一个示例组件用于演示（实际项目应导入真实组件）
const SampleButton = Vue.component('SampleButton', {
  template: `
    <button
      :class="['btn', 'btn-' + type, { 'is-disabled': disabled }]"
      :disabled="disabled"
      @click="handleClick">
      <slot>{{ label }}</slot>
    </button>
  `,
  props: {
    label: { type: String, default: '按钮' },
    type: { type: String, default: 'default' },
    disabled: { type: Boolean, default: false }
  },
  computed: {
    isPrimary() {
      return this.type === 'primary'
    }
  },
  methods: {
    handleClick(e) {
      if (this.disabled) return
      this.$emit('click', e)
    }
  }
})

describe('SampleButton 组件', () => {
  it('渲染默认 label', () => {
    const wrapper = mount(SampleButton)
    expect(wrapper.text()).toBe('按钮')
  })

  it('props label 自定义文本', () => {
    const wrapper = mount(SampleButton, { propsData: { label: '保存' } })
    expect(wrapper.text()).toBe('保存')
  })

  it('插槽内容覆盖 label', () => {
    const wrapper = mount(SampleButton, { slots: { default: '提交' } })
    expect(wrapper.text()).toBe('提交')
  })

  it('type=primary 时添加 btn-primary class', () => {
    const wrapper = mount(SampleButton, { propsData: { type: 'primary' } })
    expect(wrapper.classes()).toContain('btn-primary')
  })

  it('disabled 时添加 is-disabled class 且 button 被禁用', () => {
    const wrapper = mount(SampleButton, { propsData: { disabled: true } })
    expect(wrapper.classes()).toContain('is-disabled')
    expect(wrapper.attributes('disabled')).toBeDefined()
  })

  it('点击触发 click 事件', async () => {
    const wrapper = mount(SampleButton)
    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toBeTruthy()
    expect(wrapper.emitted('click').length).toBe(1)
  })

  it('disabled 时点击不触发 click 事件', async () => {
    const wrapper = mount(SampleButton, { propsData: { disabled: true } })
    await wrapper.trigger('click')
    // disabled 的 button 浏览器原生不会触发 click，但 trigger 仍会调用
    // 但 handleClick 内部 return，不会 emit
    expect(wrapper.emitted('click')).toBeFalsy()
  })

  it('计算属性 isPrimary 正确', () => {
    const wrapper = mount(SampleButton, { propsData: { type: 'primary' } })
    expect(wrapper.vm.isPrimary).toBe(true)

    const wrapper2 = mount(SampleButton, { propsData: { type: 'default' } })
    expect(wrapper2.vm.isPrimary).toBe(false)
  })
})

// ============================================================
// 实际项目组件测试模板（取消注释并修改路径以使用）
// ============================================================
//
// import MyComponent from '@/views/exam/exam/form.vue'
//
// describe('exam/form.vue', () => {
//   it('渲染所有必填字段', () => {
//     const wrapper = shallowMount(MyComponent, {
//       propsData: { requiredFields: ['name', 'telephone'] }
//     })
//     expect(wrapper.find('input[name="name"]').exists()).toBe(true)
//   })
// })

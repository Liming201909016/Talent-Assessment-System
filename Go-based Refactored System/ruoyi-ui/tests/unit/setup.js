// Vitest 全局测试设置
// 由 vitest.config.js 中的 setupFiles 自动加载

import { config } from '@vue/test-utils'

// Element UI 组件在测试中的简化处理：用 stub 替代复杂组件
// 避免引入完整 element-ui（会引入大量副作用）
config.stubs = {
  'el-button': true,
  'el-input': true,
  'el-form': true,
  'el-form-item': true,
  'el-select': true,
  'el-option': true,
  'el-table': true,
  'el-table-column': true,
  'el-tag': true,
  'el-dialog': true,
  'el-pagination': true,
  'el-tooltip': true,
  'el-popconfirm': true,
  'el-icon': true,
  'router-link': true,
  'router-view': true
}

// Mock window 对象的常用方法
if (typeof window !== 'undefined') {
  window.URL.createObjectURL = window.URL.createObjectURL || (() => 'blob:mock')
  window.URL.revokeObjectURL = window.URL.revokeObjectURL || (() => {})
}

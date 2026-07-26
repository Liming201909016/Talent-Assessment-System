// UX 修复回归守护测试
// 防止以下 UX 决策被意外回退：
//  - DataTable 默认 pageSize=20（之前是 10，密度太低）
//  - RepoSelect prop value 支持 String + Array
//  - 关键图标按钮有 title/aria-label
//
// 测试方式：扫描源代码，断言关键 UI 决策没被改回旧值
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'fs'
import { resolve } from 'path'

// 读组件源码（vitest 在 ruoyi-ui 目录运行）
function readSrc(relPath) {
  return readFileSync(resolve(process.cwd(), relPath), 'utf-8')
}

describe('UX 决策回归守护', () => {
  describe('DataTable 默认 pageSize', () => {
    it('listQuery.size 必须是 20（防回退到 10）', () => {
      const src = readSrc('src/components/DataTable/index.vue')
      // 必须包含 size: 20，不能包含 size: 10
      expect(src).toMatch(/size:\s*20\b/)
      expect(src).not.toMatch(/size:\s*10\b/)
    })
  })

  describe('RepoSelect 类型兼容性', () => {
    const src = readSrc('src/components/RepoSelect/index.vue')

    it('value prop 必须支持 String + Array（防 multi=true 时类型警告）', () => {
      // 必须是 [String, Array] 数组形式或 type: [String, Array]
      const hasMultiType = /type:\s*\[\s*String,\s*Array\s*\]/.test(src)
      expect(hasMultiType, 'value prop 必须 type: [String, Array]').toBe(true)
    })

    it('禁止 value: String 单类型（FB-026 回归）', () => {
      // 禁止单独的 value: String 声明
      const oldPattern = /^\s*value:\s*String\s*$/m
      expect(src).not.toMatch(oldPattern)
    })

    it('currentValue 默认值必须按 multi 区分', () => {
      // 必须有 this.multi ? [] : '' 模式
      expect(src).toMatch(/this\.multi\s*\?\s*\[\s*\]\s*:\s*['"]['"]/)
    })
  })

  describe('图标按钮无障碍标签', () => {
    const components = [
      { file: 'src/components/Hamburger/index.vue', name: 'Hamburger' },
      { file: 'src/components/Screenfull/index.vue', name: 'Screenfull' },
      { file: 'src/components/HeaderSearch/index.vue', name: 'HeaderSearch' },
      { file: 'src/components/SizeSelect/index.vue', name: 'SizeSelect' }
    ]

    components.forEach(({ file, name }) => {
      it(`${name} 必须有 title 或 aria-label`, () => {
        const src = readSrc(file)
        const hasA11y = src.includes('aria-label') || src.includes(':title=')
        expect(hasA11y, `${name} 应有 aria-label 或 title 提供可访问文本`).toBe(true)
      })
    })
  })

  describe('业务列表 pageSize（核心列表组件抽样）', () => {
    const lists = [
      'src/views/system/user/index.vue',
      'src/views/system/role/index.vue',
      'src/views/system/dict/index.vue',
      'src/views/qu/qu/index.vue',
      'src/views/qu/repo/index.vue',
      'src/views/exam/exam/index.vue',
      'src/views/user/exam/index.vue'
    ]

    lists.forEach(path => {
      it(`${path.split('/').slice(-2).join('/')} pageSize/size 默认值 = 20`, () => {
        const src = readSrc(path)
        // 文件中如果包含 pageSize: 或 size: 数字，必须是 20
        const pageSizeMatch = src.match(/pageSize:\s*(\d+)/)
        const sizeMatch = src.match(/^[\s]*size:\s*(\d+),/m)

        if (pageSizeMatch) {
          expect(parseInt(pageSizeMatch[1]), `pageSize 应为 20，实际 ${pageSizeMatch[1]}`).toBeGreaterThanOrEqual(20)
        }
        if (sizeMatch) {
          expect(parseInt(sizeMatch[1]), `size 应为 20，实际 ${sizeMatch[1]}`).toBeGreaterThanOrEqual(20)
        }
      })
    })
  })
})

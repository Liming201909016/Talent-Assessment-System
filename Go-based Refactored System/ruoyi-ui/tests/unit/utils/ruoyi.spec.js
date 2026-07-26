// 测试 src/utils/ruoyi.js 的核心工具函数
import { describe, it, expect } from 'vitest'
import { parseTime, addDateRange, selectDictLabel } from '@/utils/ruoyi.js'

describe('parseTime — 日期格式化', () => {
  it('null/empty 返回 null', () => {
    expect(parseTime(null)).toBeNull()
    expect(parseTime(undefined)).toBeNull()
    expect(parseTime('')).toBeNull()
  })

  it('Date 对象格式化', () => {
    const date = new Date(2026, 3, 21, 12, 30, 45) // 月份 0-based, 4月
    expect(parseTime(date, '{y}-{m}-{d}')).toBe('2026-04-21')
    expect(parseTime(date, '{y}-{m}-{d} {h}:{i}:{s}')).toBe('2026-04-21 12:30:45')
  })

  it('默认格式包含完整时间', () => {
    const date = new Date(2026, 3, 21, 12, 30, 45)
    expect(parseTime(date)).toBe('2026-04-21 12:30:45')
  })

  it('单位数字补零', () => {
    const date = new Date(2026, 0, 5, 7, 8, 9)
    expect(parseTime(date, '{y}-{m}-{d} {h}:{i}:{s}')).toBe('2026-01-05 07:08:09')
  })

  it('解析 RFC3339 字符串（关键：避免 RFC3339 显示问题）', () => {
    // 项目已知陷阱：后端返回 "2026-04-21T12:00:00Z" 格式
    const result = parseTime('2026-04-21T12:00:00Z', '{y}-{m}-{d}')
    expect(result).toBe('2026-04-21')
  })

  it('解析"yyyy-MM-dd HH:mm:ss"字符串', () => {
    const result = parseTime('2026-04-21 12:00:00', '{y}-{m}-{d} {h}:{i}')
    expect(result).toBe('2026-04-21 12:00')
  })

  it('星期格式化', () => {
    // 2026-04-21 是周二
    const date = new Date(2026, 3, 21)
    const day = date.getDay() // 0=日, 1=一, 2=二
    const expected = ['日', '一', '二', '三', '四', '五', '六'][day]
    expect(parseTime(date, '{a}')).toBe(expected)
  })

  it('10 位时间戳（秒）自动转毫秒', () => {
    const ts = Math.floor(new Date(2026, 3, 21, 12, 0, 0).getTime() / 1000)
    expect(parseTime(ts, '{y}-{m}-{d}')).toBe('2026-04-21')
  })
})

describe('addDateRange — 日期范围参数注入', () => {
  it('默认注入 beginTime/endTime', () => {
    const params = {}
    const result = addDateRange(params, ['2026-04-01', '2026-04-30'])
    expect(result.params.beginTime).toBe('2026-04-01')
    expect(result.params.endTime).toBe('2026-04-30')
  })

  it('自定义 propName 注入 beginXxx/endXxx', () => {
    const params = {}
    const result = addDateRange(params, ['2026-04-01', '2026-04-30'], 'CreateTime')
    expect(result.params.beginCreateTime).toBe('2026-04-01')
    expect(result.params.endCreateTime).toBe('2026-04-30')
  })

  it('dateRange 非数组时仍能处理', () => {
    const params = {}
    const result = addDateRange(params, null)
    expect(result.params.beginTime).toBeUndefined()
  })

  it('保留原 params 中的其他属性', () => {
    const params = { name: 'test', params: { existing: 'value' } }
    const result = addDateRange(params, ['2026-04-01', '2026-04-30'])
    expect(result.name).toBe('test')
    expect(result.params.existing).toBe('value')
    expect(result.params.beginTime).toBe('2026-04-01')
  })
})

describe('selectDictLabel — 字典回显', () => {
  const dict = [
    { value: '0', label: '正常' },
    { value: '1', label: '禁用' },
    { value: '2', label: '已删除' }
  ]

  it('精确匹配返回 label', () => {
    expect(selectDictLabel(dict, '1')).toBe('禁用')
    expect(selectDictLabel(dict, '0')).toBe('正常')
  })

  it('数字与字符串值兼容（== 比较）', () => {
    expect(selectDictLabel(dict, 1)).toBe('禁用')
    expect(selectDictLabel(dict, 0)).toBe('正常')
  })

  it('value 为 undefined 返回空字符串', () => {
    expect(selectDictLabel(dict, undefined)).toBe('')
  })

  it('未匹配时返回原始 value（不是空字符串）', () => {
    // 实际行为：actions 为空时 push(value)，所以返回 value 本身
    expect(selectDictLabel(dict, '99')).toBe('99')
  })
})

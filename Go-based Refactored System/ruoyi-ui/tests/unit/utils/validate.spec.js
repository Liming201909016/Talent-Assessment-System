// 测试 src/utils/validate.js 的所有验证函数
import { describe, it, expect } from 'vitest'
import {
  isExternal,
  validUsername,
  validURL,
  validLowerCase,
  validUpperCase,
  validAlphabets,
  validEmail,
  isString,
  isArray
} from '@/utils/validate.js'

describe('isExternal', () => {
  it('识别 http/https/mailto/tel 为外部链接', () => {
    expect(isExternal('http://example.com')).toBe(true)
    expect(isExternal('https://example.com')).toBe(true)
    expect(isExternal('mailto:a@b.com')).toBe(true)
    expect(isExternal('tel:13800138000')).toBe(true)
  })

  it('识别相对路径/绝对路径为内部链接', () => {
    expect(isExternal('/path/to/page')).toBe(false)
    expect(isExternal('./relative')).toBe(false)
    expect(isExternal('#hash')).toBe(false)
    expect(isExternal('?query=1')).toBe(false)
  })
})

describe('validUsername', () => {
  it('admin/editor 为合法用户名', () => {
    expect(validUsername('admin')).toBe(true)
    expect(validUsername('editor')).toBe(true)
  })

  it('其他用户名不合法', () => {
    expect(validUsername('user')).toBe(false)
    expect(validUsername('')).toBe(false)
  })

  it('忽略首尾空格', () => {
    expect(validUsername('  admin  ')).toBe(true)
  })
})

describe('validURL', () => {
  it('合法的 http/https URL', () => {
    expect(validURL('http://www.example.com')).toBe(true)
    expect(validURL('https://example.com/path?q=1')).toBe(true)
  })

  it('不合法的 URL', () => {
    expect(validURL('not-a-url')).toBe(false)
    expect(validURL('')).toBe(false)
  })
})

describe('validLowerCase / validUpperCase / validAlphabets', () => {
  it('全小写', () => {
    expect(validLowerCase('abc')).toBe(true)
    expect(validLowerCase('aBc')).toBe(false)
    expect(validLowerCase('a1')).toBe(false)
  })

  it('全大写', () => {
    expect(validUpperCase('ABC')).toBe(true)
    expect(validUpperCase('AbC')).toBe(false)
  })

  it('全字母', () => {
    expect(validAlphabets('aBc')).toBe(true)
    expect(validAlphabets('a1')).toBe(false)
    expect(validAlphabets('')).toBe(false)
  })
})

describe('validEmail', () => {
  it('合法邮箱', () => {
    expect(validEmail('user@example.com')).toBe(true)
    expect(validEmail('a.b+tag@sub.example.co')).toBe(true)
  })

  it('不合法邮箱', () => {
    expect(validEmail('plain')).toBe(false)
    expect(validEmail('@nodomain.com')).toBe(false)
    expect(validEmail('no@')).toBe(false)
    expect(validEmail('')).toBe(false)
  })
})

describe('isString / isArray', () => {
  it('isString 识别字符串', () => {
    expect(isString('abc')).toBe(true)
    expect(isString('')).toBe(true)
    expect(isString(new String('x'))).toBe(true)
    expect(isString(123)).toBe(false)
    expect(isString(null)).toBe(false)
    expect(isString([])).toBe(false)
  })

  it('isArray 识别数组', () => {
    expect(isArray([])).toBe(true)
    expect(isArray([1, 2])).toBe(true)
    expect(isArray('abc')).toBe(false)
    expect(isArray({})).toBe(false)
    expect(isArray(null)).toBe(false)
  })
})

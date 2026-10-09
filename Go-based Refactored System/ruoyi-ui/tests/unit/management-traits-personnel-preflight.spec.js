import { describe, expect, it, vi } from 'vitest'
import compiler from 'vue-template-compiler'
import babel from '@babel/core'
import fs from 'fs'
import path from 'path'

// UF-049 / MT-ADMIN-PERSONNEL-LIST-403: replay the actual personnel SFC,
// not a replacement business component. No database, credentials or remote writes.
function personnel(existing = null) {
  const lists = []
  const api = {
    getListTester: vi.fn(query => {
      lists.push({ ...query })
      return Promise.resolve({ code: 200, rows: [], total: 0 })
    }),
    addTester: vi.fn(() => Promise.resolve({ code: 200, data: true })),
    getTester: vi.fn(() => Promise.resolve({ code: 200, data: { ...existing } })),
    updateTester: vi.fn(() => Promise.resolve({ code: 200, data: true }))
  }
  const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/tester/tester/index.vue'), 'utf8')
  const sfc = compiler.parseComponent(source)
  const code = babel.transformSync(sfc.script.content, {
    babelrc: false, configFile: false, plugins: ['@babel/plugin-transform-modules-commonjs']
  }).code
  const module = { exports: {} }
  new Function('require', 'module', 'exports', code)(name => {
    if (name === '@/api/tester/tester') return api
    if (name === '@/utils/auth') return { getToken: () => '' }
    return {}
  }, module, module.exports)
  const component = module.exports.default
  const vm = component.data()
  Object.entries(component.methods).forEach(([name, method]) => { vm[name] = method.bind(vm) })
  vm.resetForm = vi.fn()
  vm.$modal = { msgSuccess: vi.fn() }
  vm.$refs = { form: { validate: fn => fn(true) } }
  return { vm, api, lists }
}
async function flush() { for (let n = 0; n < 8; n++) await Promise.resolve() }

describe('UF-049 owned tester preparation before its profile freeze', () => {
  it('UF-050 normal scoped personnel creation supplies an enabled status for subsequent TEST login', async () => {
    // Real UI/SQL reproduction: normal add succeeds but stores status NULL;
    // the strict TEST identity admission requires an explicit enabled status.
    const { vm, api } = personnel()
    vm.queryParams.examId = 'owned-00201-tester-draft'
    vm.handleQuery()
    await flush()
    vm.handleAdd()
    Object.assign(vm.form, { name: 'Synthetic owned tester', telephone: '13800000000', stuFlag: 0 })
    vm.submitForm()
    await flush()
    expect(api.addTester).toHaveBeenCalledTimes(1)
    expect(api.addTester.mock.calls[0][0].examId).toBe('owned-00201-tester-draft')
    expect(api.addTester.mock.calls[0][0].status).toBe('0')
  })
  it.each(['0', '1', null, '2'])('UF-050 preserves existing status %s through actual edit and save', async status => {
    const existing = { id: 'owned-tester', examId: 'owned-draft', name: 'Synthetic owned tester', status }
    const { vm, api } = personnel(existing)
    vm.queryParams.examId = existing.examId
    vm.examList = [{ id: existing.examId }]
    vm.handleUpdate({ id: existing.id })
    await flush()
    expect(api.getTester).toHaveBeenCalledWith(existing.id)
    expect(vm.form.status).toBe(status)
    vm.form.name = 'Synthetic edited tester'
    vm.submitForm()
    await flush()
    expect(api.addTester).not.toHaveBeenCalled()
    expect(api.updateTester).toHaveBeenCalledTimes(1)
    expect(api.updateTester.mock.calls[0][0]).toEqual({ ...existing, name: 'Synthetic edited tester' })
    vm.handleAdd()
    expect(vm.form.id).toBeUndefined()
    expect(vm.form.status).toBe('0')
    expect(existing.status).toBe(status)
  })
  it.each(['00201', '00202'])('keeps the exact owned draft filter through normal %s add and refresh', async code => {
    const { vm, api, lists } = personnel()
    const examId = `owned-${code}-tester-draft`
    // Correct the driver, not the component/guard: select the owned draft in
    // the query form BEFORE adding, and prepare all testers before any freeze.
    vm.queryParams.examId = examId
    vm.handleQuery()
    await flush()
    vm.handleAdd()
    Object.assign(vm.form, { examId, name: 'Synthetic owned tester', telephone: '13800000000', stuFlag: 0 })
    vm.submitForm()
    await flush()
    expect(api.addTester).toHaveBeenCalledTimes(1)
    expect(api.addTester.mock.calls[0][0].examId).toBe(examId)
    expect(lists.length).toBeGreaterThan(0)
    expect(lists.every(query => query.examId === examId)).toBe(true)
    expect(lists.at(-1).pageNum).toBe(1)
    expect(vm.open).toBe(false)
    expect(vm.loading).toBe(false)
    expect(vm.testerList).toEqual([])
  })
  it('retains the old unfiltered request as a negative scope observation', async () => {
    const { vm, lists } = personnel()
    vm.handleAdd()
    Object.assign(vm.form, { examId: 'owned-draft', name: 'Synthetic owned tester' })
    vm.submitForm()
    await flush()
    expect(lists).toHaveLength(1)
    expect(lists[0].examId).toBe('')
    // HTTP403 is verified separately by the real Go guard tests. This mock
    // proves request scope only and must never claim a real HTTP200/403.
  })
})
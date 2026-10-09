import { describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import compiler from 'vue-template-compiler'
import babel from '@babel/core'
import fs from 'fs'
import path from 'path'
import * as product from '@/utils/managementTraitsProduct'

describe('management traits independent product policy', () => {
  it.each(['00201', '00202'])('preserves ordinary legacy %s', code => {
    expect(product.classifyManagementTraitsProduct(code)).toBe('LEGACY002')
    expect(product.classifyManagementTraitsExam({ id: 'e', repoCode: code, managementTraitsProfileFrozen: false }, 'e')).toBe('LEGACY002')
    expect(product.classifyManagementTraitsExam({ id: 'e', repoCode: code, managementTraitsProfileFrozen: true }, 'e')).toBe('FROZEN_COMPAT002')
    expect(product.classifyManagementTraitsExam({ id: 'other', repoCode: code, managementTraitsProfileFrozen: true }, 'e')).toBe('UNKNOWN')
  })
  it.each(['00501', '00502'])('code %s is intent, not frozen authority', code => {
    expect(product.classifyManagementTraitsProduct(code)).toBe('NEW005')
    for (const frozen of [undefined, null, 'true', 1, false]) expect(product.classifyManagementTraitsExam({ id: 'e', repoCode: code, managementTraitsProfileFrozen: frozen }, 'e')).toBe('UNKNOWN')
    expect(product.classifyManagementTraitsExam({ id: 'e', repoCode: code, managementTraitsProfileFrozen: false, managementTraitsLifecycle: 'draft', isManagementTraits: true }, 'e')).toBe('DRAFT005')
    expect(product.classifyManagementTraitsExam({ id: 'e', repoCode: code, managementTraitsProfileFrozen: true }, 'e')).toBe('NEW005')
  })
  it.each(['005', '00503', '005010', '00501 ', '00101', '00401', null])('does not guess code %s', code => {
    expect(product.classifyManagementTraitsProduct(code)).toBe('OTHER')
  })
  it('rejects aliases, cross-family and lifecycle contradictions', () => {
    const exam = { id: 'e', repoCode: '00501', managementTraitsProfileFrozen: true }
    for (const extra of [{ repoList: [{ repoCode: '00201' }] }, { isManagementTraits: false }, { managementTraitsLifecycle: 'legacy' }, { managementTraitsLifecycle: null }, { id: {} }]) expect(product.classifyManagementTraitsExam({ ...exam, ...extra }, 'e')).toBe('UNKNOWN')
  })
  it('keeps a normal multi-002 legacy exam but never a mixed frozen or 005 exam', () => {
    const exam = { id: 'e', repoCode: '00201', repoList: [{ repoCode: '00202' }], managementTraitsProfileFrozen: false }
    expect(product.classifyManagementTraitsExam(exam, 'e')).toBe('LEGACY002')
    expect(product.classifyManagementTraitsExam({ ...exam, managementTraitsProfileFrozen: true }, 'e')).toBe('UNKNOWN')
    expect(product.classifyManagementTraitsExam({ ...exam, assessmentType: 'competency' }, 'e')).toBe('UNKNOWN')
  })
})

describe('005 actual configuration reload', () => {
  it.each(['00501', '00502'])('unknown %s closes while persistent draft stays editable', async code => {
    const sfc = compiler.parseComponent(fs.readFileSync(path.resolve('src/views/exam/exam/form.vue'), 'utf8'))
    const js = babel.transformSync(sfc.script.content, { babelrc: false, configFile: false, plugins: ['@babel/plugin-transform-modules-commonjs'] }).code
    const data = { id: 'e', repoCode: code, repoList: [{ repoId: 'synthetic-' + code, repoCode: code, radioCount: 140 }], assessmentType: 'legacy', scoringMode: 'legacy', joinType: 1, managementTraitsProfileFrozen: false, requiredFields: 'name,telephone' }
    const module = { exports: {} }
    const profile = vi.fn()
    new Function('require', 'module', 'exports', js)(name => {
      if (name === '@/utils/managementTraitsProduct') return product
      if (name === '@/api/exam/exam') return { fetchDetail: vi.fn(async () => ({ data: { ...data } })) }
      if (name === '@/api/managementTraits') return { canManageManagementTraits: () => true, managementTraitsExamKnown: () => false, fetchManagementTraitsProfile: profile }
      return { __esModule: true, default: { render: h => h('div') } }
    }, module, module.exports)
    const w = shallowMount({ ...module.exports.default, ...compiler.compileToFunctions(sfc.template.content) }, { directives: { loading: () => {} }, mocks: { $route: { params: {} }, $store: {}, $message: { error: vi.fn() } }, stubs: ['el-card', 'el-alert', 'el-checkbox', 'el-checkbox-group', 'el-radio', 'el-radio-group', 'el-input-number', 'el-date-picker', 'el-switch'] })
    await w.vm.fetchData('e')
    expect(w.vm.managementTraitsReadOnly).toBe(true); expect(w.vm.managementTraitsError).not.toBe(''); expect(profile).not.toHaveBeenCalled()
    data.managementTraitsLifecycle = 'draft'; data.isManagementTraits = true
    await w.vm.fetchData('e')
    expect(w.vm.managementTraitsReadOnly).toBe(false); expect(w.vm.managementTraitsDraft).toBe(true); expect(w.vm.managementTraitsSelected).toBe(true); expect(profile).not.toHaveBeenCalled(); w.destroy()
  })
})
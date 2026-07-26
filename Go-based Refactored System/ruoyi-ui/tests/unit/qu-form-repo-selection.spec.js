import { describe, expect, it, vi } from 'vitest'
import QuForm from '@/views/qu/qu/form.vue'

vi.mock('@/api/qu/qu', () => ({ fetchDetail: vi.fn(), saveData: vi.fn() }))
vi.mock('@/components/RepoSelect', () => ({ default: { name: 'RepoSelect', render: h => h('div') } }))
vi.mock('@/components/FileUpload', () => ({ default: { name: 'FileUpload', render: h => h('div') } }))

describe('Question form repository selection regression', () => {
  it('validates the field bound by the single repository selector', () => {
    const state = QuForm.data()
    expect(state.rules.repoId).toEqual([
      expect.objectContaining({ required: true, message: '必须选择一个题库！' })
    ])
    expect(state.rules.repoIds).toBeUndefined()
  })

  it('always synchronizes a changed repository before save', () => {
    expect(typeof QuForm.methods.syncRepoSelection).toBe('function')
    const vm = { postForm: { repoId: 'repo-new', repoIds: ['repo-old'] } }
    QuForm.methods.syncRepoSelection.call(vm)
    expect(vm.postForm.repoIds).toEqual(['repo-new'])

    vm.postForm.repoId = ''
    QuForm.methods.syncRepoSelection.call(vm)
    expect(vm.postForm.repoIds).toEqual([])
  })
})

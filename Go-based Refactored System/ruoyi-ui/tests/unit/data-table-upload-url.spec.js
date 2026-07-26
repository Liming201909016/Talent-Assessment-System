import { describe, expect, it, vi } from 'vitest'
import DataTable from '@/components/DataTable/index.vue'

vi.mock('@/api/common', () => ({ fetchList: vi.fn(), deleteData: vi.fn(), changeState: vi.fn() }))
vi.mock('@/utils/auth', () => ({ getToken: vi.fn(() => 'token') }))
vi.mock('@/components/Pagination', () => ({ default: { name: 'Pagination', render: h => h('div') } }))

describe('DataTable question import upload URL regression', () => {
  it('uses the configured API base instead of a development-only prefix', () => {
    const state = DataTable.data()
    expect(state.upload.url).toBe((process.env.VUE_APP_BASE_API || '') + '/exam/api/qu/qu/import-excel')
    expect(state.upload.url).not.toContain('/dev-api/')
  })

  it('sends the current admin token in the upload headers', () => {
    const state = DataTable.data()
    expect(state.upload.headers).toEqual({ Authorization: 'Bearer token' })
  })
})

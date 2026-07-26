import { describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import CompetencyResults from '@/views/exam/exam/competencyResults.vue'
import { downloadCompetencyReport, fetchCompetencyResults, generateCompetencyReport } from '@/api/competency'

vi.mock('file-saver', () => ({ saveAs: vi.fn() }))

vi.mock('@/api/exam/exam', () => ({
  fetchDetail: vi.fn(() => Promise.resolve({ data: { title: '胜任力测评', competencyDimensions: [] } }))
}))

vi.mock('@/api/competency', () => ({
  fetchCompetencyResults: vi.fn(() => Promise.resolve({ data: { records: [], total: 0 } })),
  fetchCompetencyResultDetail: vi.fn(),
  generateCompetencyReport: vi.fn(() => Promise.resolve({ data: { id: 'report-1' } })),
  downloadCompetencyReport: vi.fn(() => Promise.resolve(new Blob(['pdf'], { type: 'application/pdf' })))
}))

const mountPage = (methods = {}) => shallowMount(CompetencyResults, {
  methods,
  mocks: {
    $route: { params: { examId: 'exam-1' } },
    $router: { back: vi.fn(), push: vi.fn() }
  },
  stubs: ['el-form', 'el-form-item', 'el-select', 'el-option', 'el-radio-group', 'el-radio-button', 'el-table', 'el-table-column', 'el-tag', 'el-button', 'el-dialog', 'el-descriptions', 'el-descriptions-item', 'el-tabs', 'el-tab-pane', 'el-dropdown', 'el-dropdown-menu', 'el-dropdown-item', 'pagination', 'el-empty']
})

describe('Competency result management', () => {
  it('loads results using the safe default sorting contract', async () => {
    fetchCompetencyResults.mockClear()
    mountPage()
    await Promise.resolve()
    expect(fetchCompetencyResults).toHaveBeenCalledWith(expect.objectContaining({
      examId: 'exam-1', sortBy: 'submittedAt', sortDirection: 'desc', dimensionId: ''
    }))
  })

  it('selects the first measured dimension before dimension-score sorting', () => {
    const loadResults = vi.fn()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults })
    wrapper.setData({ dimensions: [{ dimensionId: 'dimension-1', dimensionName: '计划组织' }] })
    wrapper.vm.handleSortByChange('dimensionScore')
    expect(wrapper.vm.query.dimensionId).toBe('dimension-1')
    expect(loadResults).toHaveBeenCalled()
  })

  it('does not request dimension sorting without a dimension id', () => {
    fetchCompetencyResults.mockClear()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.setData({ query: { ...wrapper.vm.query, sortBy: 'dimensionScore', dimensionId: '' } })
    wrapper.vm.loadResults = CompetencyResults.methods.loadResults.bind(wrapper.vm)
    wrapper.vm.loadResults()
    expect(fetchCompetencyResults).not.toHaveBeenCalled()
    expect(wrapper.vm.rows).toEqual([])
  })

  // FB-067: 命名路由 /exam/competency/report/:paperId 必须通过 params 传入 paperId。
  it('opens a complete competency test report with the required path parameter', () => {
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.showReport({ paperId: 'paper-complete-1', isComplete: 1 })
    expect(wrapper.vm.$router.push).toHaveBeenCalledWith({
      name: 'CompetencyReport',
      params: { paperId: 'paper-complete-1' }
    })
  })

  it('generates and downloads a complete temporary competency PDF report', async () => {
    generateCompetencyReport.mockClear()
    downloadCompetencyReport.mockClear()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), error: vi.fn() }
    await wrapper.vm.generateReport({ paperId: 'paper-complete-1', isComplete: 1 })
    expect(generateCompetencyReport).toHaveBeenCalledWith({ paperId: 'paper-complete-1', force: false })
    await wrapper.vm.downloadReport({ paperId: 'paper-complete-1', isComplete: 1 })
    expect(downloadCompetencyReport).toHaveBeenCalledWith('paper-complete-1')
  })
})

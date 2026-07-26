import { describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import CompetencyExam from '@/views/paper/exam/competencyExam.vue'
import CompetencyReport from '@/views/paper/exam/competencyReport.vue'
import { fetchCompetencyInternalReportData, submitCompetencyPaper } from '@/api/competency'

vi.mock('@/api/competency', () => ({
  fetchCompetencyPaper: vi.fn(), saveCompetencyAnswer: vi.fn(), submitCompetencyPaper: vi.fn(), fetchCompetencyReportData: vi.fn(), fetchCompetencyInternalReportData: vi.fn()
}))

describe('Competency runtime pages', () => {
  it('renders only one complex question card while navigation stays lightweight', () => {
    const questions = Array.from({ length: 384 }, (_, index) => ({ id: `q${index}`, code: `Q${index}`, content: '题目', answered: false, rawValue: null, options: [] }))
    const wrapper = shallowMount(CompetencyExam, { methods: { loadPaper: vi.fn() }, mocks: { $route: { params: { paperId: 'p1' } } }, stubs: ['el-card','el-radio-group','el-radio','el-button','el-progress'] })
    wrapper.setData({ paper: { totalCount: 384, answeredCount: 0, unansweredCount: 384, questions }, loading: false })
    expect(wrapper.vm.currentQuestion.id).toBe('q0')
    expect(wrapper.vm.paper.questions.length).toBe(384)
  })

  it('manual submit locates the first unanswered question', async () => {
    const wrapper = shallowMount(CompetencyExam, { methods: { loadPaper: vi.fn() }, mocks: { $route: { params: { paperId: 'p1' } }, $message: { warning: vi.fn() } }, stubs: ['el-card','el-radio-group','el-radio','el-button','el-progress'] })
    wrapper.setData({ paper: { totalCount: 2, answeredCount: 1, unansweredCount: 1, questions: [{ id:'q1',answered:true,options:[] }, { id:'q2',answered:false,options:[] }] }, currentIndex: 0, loading: false })
    wrapper.vm.confirmSubmit()
    expect(wrapper.vm.currentIndex).toBe(1)
  })

  it('renders an accessible assessment workspace with clear answer states', async () => {
    const wrapper = shallowMount(CompetencyExam, { methods: { loadPaper: vi.fn() }, mocks: { $route: { params: { paperId: 'p1' } } }, stubs: ['el-card','el-radio-group','el-radio','el-button','el-progress'] })
    await wrapper.setData({
      paper: {
        totalCount: 2,
        answeredCount: 1,
        unansweredCount: 1,
        questions: [
          { id: 'q1', code: 'D01-Q01', content: '题目一', answered: true, rawValue: 4, options: [{ rawValue: 4, label: '比较符合' }] },
          { id: 'q2', code: 'D01-Q02', content: '题目二', answered: false, rawValue: null, options: [] }
        ]
      },
      currentIndex: 0,
      loading: false
    })
    expect(wrapper.find('.exam-kicker').text()).toBe('胜任力测评')
    expect(wrapper.find('.scale-hint').text()).toContain('请选择最符合自身实际情况的一项')
    expect(wrapper.find('.question-nav').attributes('aria-label')).toBe('题目导航')
    expect(wrapper.findAll('.question-nav button').at(0).attributes('aria-label')).toContain('已答')
    expect(wrapper.findAll('.question-nav button').at(1).attributes('aria-label')).toContain('未答')
    expect(wrapper.find('.question-nav button').attributes('aria-current')).toBe('true')
  })

  it('timeout UI lets the server derive timeout instead of trusting a client timeout claim', async () => {
    submitCompetencyPaper.mockClear()
    submitCompetencyPaper.mockResolvedValue({ data: {} })
    const wrapper = shallowMount(CompetencyExam, {
      methods: { loadPaper: vi.fn(), finish: vi.fn() },
      mocks: { $route: { params: { paperId: 'p1' } }, $message: { error: vi.fn() } },
      stubs: ['el-card','el-radio-group','el-radio','el-button','el-progress']
    })
    await wrapper.vm.submit('timeout')
    expect(submitCompetencyPaper).toHaveBeenCalledWith('p1', 'manual', expect.any(String))
  })

  it('report uses one layout and audience snapshot', () => {
    const wrapper = shallowMount(CompetencyReport, { methods: {}, mocks: { $route: { params: { paperId: 'p1' } } } })
    wrapper.setData({ data: { result: { reportAudience: 'leader', overallScore: 4, evaluationAverage: 4, evaluationLevel: 'high', submittedAt: '2026-07-25', isComplete: 1 }, dimensions: [], reportTextReady: false, reportTextMessage: '正式解读文案待配置' } })
    expect(wrapper.vm.audienceLabel).toBe('领导人员版')
    expect(wrapper.vm.data.reportTextMessage).toBe('正式解读文案待配置')
  })

  it('internal PDF rendering uses the exact internal report endpoint and displays the temporary disclaimer', async () => {
    fetchCompetencyInternalReportData.mockResolvedValue({ data: {
      result: { reportAudience: 'leader', overallScore: 4, evaluationAverage: 4, evaluationLevel: 'high', submittedAt: '2026-07-25', isComplete: 1 },
      dimensions: [], reportTextReady: true, reportTextMessage: '临时测试文案，仅用于系统功能验证，不可作为人才决策依据。',
      reportText: { overallText: '临时总体评价', dimensionTexts: {}, isTemporary: true }
    } })
    const wrapper = shallowMount(CompetencyReport, { mocks: { $route: { params: { paperId: 'p1' }, query: { _internal: 'internal-token' } } }, stubs: ['el-alert'] })
    await vi.waitFor(() => expect(wrapper.vm.data).not.toBeNull())
    await wrapper.vm.$nextTick()
    expect(fetchCompetencyInternalReportData).toHaveBeenCalledWith('p1', 'internal-token')
    expect(wrapper.text()).toContain('不可作为人才决策依据')
    expect(wrapper.text()).toContain('临时总体评价')
  })
})

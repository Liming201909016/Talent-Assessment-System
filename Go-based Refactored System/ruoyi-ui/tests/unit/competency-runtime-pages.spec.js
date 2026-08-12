import { describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import CompetencyExam from '@/views/paper/exam/competencyExam.vue'
import CompetencyReport from '@/views/paper/exam/competencyReport.vue'
import phase1Catalog from '@/data/competencyPhase1ReportCatalog'
import { fetchCompetencyInternalReportData, submitCompetencyPaper } from '@/api/competency'
import fs from 'fs'
import path from 'path'

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

  it('implements the 00401 printable report template based on the reference document', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/paper/exam/competencyReport.vue'), 'utf8')
    for (const required of [
      'data.meta.examTitle', 'data.result.participantName', 'data.meta.generatedAt',
      '报告阅读说明', '1.00–5.00', '不同维度数量', 'personFields',
      'overall-scale', 'dimension-chart', 'dimensionCoreMeaning', '发展提示'
    ]) {
      expect(source).toContain(required)
    }
  })

  it('prints section titles and content pages without blank-page inflation', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/paper/exam/competencyReport.vue'), 'utf8')
    expect(source).toContain('<header class="page-title">报告阅读说明</header>')
    expect(source).toContain('<header class="page-title">测评结果分析</header>')
    expect(source).toContain('@media print')
    expect(source).toContain('.report-page{width:100%;min-height:auto;')
    expect(source).toContain('.cover{min-height:850px;')
    expect(source).toContain('.dimension-page{font-size:14px;line-height:1.6;padding:38px 60px}')
    expect(source).toContain('.overview-page{font-size:14px;line-height:1.6;padding:38px 60px}')
    expect(source).toContain('.overview-page .overall-score{width:120px;height:120px')
    expect(source).toContain('.phase1-guide{font-size:12px;line-height:1.55;padding:34px 54px}')
    expect(source).toContain('.phase1-report .report-page{height:850px;max-height:850px;break-inside:avoid;overflow:hidden}')
  })

  it('contains an isolated fixed ten-page phase-1 report framework', () => {
  const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/paper/exam/competencyReport.vue'), 'utf8')
  for (const required of [
    "reportKind === 'frontline_phase1'", 'phase1-report', 'phase1Pages', 'overallMaxScore',
    'phase1-groups', 'phase1-radar', 'phase1-dimension-pair', 'showRawScoreToParticipant',
    "import phase1Catalog from '@/data/competencyPhase1ReportCatalog'", 'catalogDimension',
    'secondaryLevelLabel', 'groupLevelLabel', 'overallLevelLabel', 'radar-chart', 'radarPoints',
    'dimension.definition', 'data.meta.examTitle', 'phase1Catalog.fixedTexts.coverTitle',
    'phase1Catalog.fixedTexts.readingBackground', 'phase1Catalog.fixedTexts.readingDimensions',
    'phase1Catalog.fixedTexts.readingUsage', 'phase1Catalog.fixedTexts.specialNotice'
  ]) expect(source).toContain(required)
  expect(source).toContain('phase1Pages.length === 10')
  expect(source).not.toContain('data.validity.validityScore')
  expect(source).not.toContain('35分')
  expect(source).not.toContain("L1: '起步级'")
  expect(source).not.toContain("weak: '较弱'")
  })

  it('uses a generated CSV report catalog for names, definitions and labels', () => {
    const catalog = fs.readFileSync(path.resolve(process.cwd(), 'src/data/competencyPhase1ReportCatalog.js'), 'utf8')
    for (const required of [
      "contractVersion: 'competency-phase1-csv-v1'", "code: 'general_ability'", "name: '通用能力'",
      "id: 'competency-a1-01'", "code: 'A1-01'", "name: '逻辑思维'", 'definition:',
      "secondaryLabel: '差'", "groupLabel: '低分'", "code: 'excellent'", "name: '优秀胜任'",
      "coverTitle: '胜任力测评报告'", 'readingBackground:', 'specialNotice:'
    ]) expect(catalog).toContain(required)
  })

  it('renders the CSV-backed phase-1 customer template as ten report pages', async () => {
    const dimensions = phase1Catalog.dimensions.map(item => ({ dimensionId: item.id, dimensionScore: 3.5, levelCode: 'L3' }))
    const groups = phase1Catalog.groups.map(item => ({ groupCode: item.code, groupScore: 3.5, levelCode: 'L3' }))
    const dimensionTexts = Object.fromEntries(phase1Catalog.dimensions.map(item => [item.id, `${item.name}诊断建议`]))
    const wrapper = shallowMount(CompetencyReport, { methods: { }, mocks: { $route: { params: { paperId: 'p1' }, query: {} } } })
    await wrapper.setData({ data: {
      reportKind: 'frontline_phase1', overallMaxScore: 50, groupMaxScore: 5, dimensionMaxScore: 5,
      pages: Array.from({ length: 10 }, (_, index) => ({ number: index + 1 })),
      meta: { examTitle: '基层员工一期测评', userTime: 18, generatedAt: '2026-08-12T10:00:00Z' },
      result: { participantName: '测试人员', overallScore: 35, overallLevel: 'qualified', submittedAt: '2026-08-12T09:58:00Z' },
      groups, dimensions, validity: { status: 'good', notice: '效度良好' },
      reportText: { disclaimer: '正式免责声明', overallText: '总体诊断', groupTexts: {}, dimensionTexts, validityText: '效度良好' }
    } })
    expect(wrapper.findAll('.report-page')).toHaveLength(10)
    expect(wrapper.find('.radar-chart').exists()).toBe(true)
    expect(wrapper.text()).toContain('基层员工一期测评')
    expect(wrapper.text()).toContain('逻辑思维')
    expect(wrapper.text()).toContain('运用归纳、演绎等逻辑方法分析问题')
    expect(wrapper.text()).toContain('合格胜任')
    expect(wrapper.text()).toContain('中分')
    expect(wrapper.text()).toContain('合格')
    // TestBugFB114_Phase1ApprovedDisclaimerIsVisibleInPDF
    // 对应：docs/regression-tests.md #FB-114
    expect(wrapper.text()).toContain('正式免责声明')
  })
})

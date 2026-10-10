import { describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import fs from 'fs'
import path from 'path'
import CompetencyResults from '@/views/exam/exam/competencyResults.vue'
import { downloadCompetencyReport, downloadCompetencyReportsArchive, fetchCompetencyResultDetail, fetchCompetencyResults, generateCompetencyReport } from '@/api/competency'
import { saveAs } from 'file-saver'

vi.mock('file-saver', () => ({ saveAs: vi.fn() }))

vi.mock('@/api/exam/exam', () => ({
  fetchDetail: vi.fn(() => Promise.resolve({ data: { title: '胜任力测评', competencyDimensions: [] } }))
}))

vi.mock('@/api/competency', () => ({
  fetchCompetencyResults: vi.fn(() => Promise.resolve({ data: { records: [], total: 0 } })),
  fetchCompetencyResultDetail: vi.fn(),
  generateCompetencyReport: vi.fn(() => Promise.resolve({ data: { id: 'report-1' } })),
  downloadCompetencyReport: vi.fn(() => Promise.resolve(new Blob(['pdf'], { type: 'application/pdf' }))),
  downloadCompetencyReportsArchive: vi.fn(() => Promise.resolve(new Blob(['zip'], { type: 'application/zip' })))
}))

const mountPage = (methods = {}) => shallowMount(CompetencyResults, {
  methods,
  mocks: {
    $route: { params: { examId: 'exam-1' } },
    $router: { back: vi.fn(), push: vi.fn() },
    $store: { state: { app: { device: 'desktop' } } }
  },
  stubs: ['el-form', 'el-form-item', 'el-select', 'el-option', 'el-radio-group', 'el-radio-button', 'el-table', 'el-table-column', 'el-tag', 'el-button', 'el-dialog', 'el-descriptions', 'el-descriptions-item', 'el-tabs', 'el-tab-pane', 'el-dropdown', 'el-dropdown-menu', 'el-dropdown-item', 'pagination', 'el-empty']
})

const deferred = () => {
  let resolve
  const promise = new Promise(done => { resolve = done })
  return { promise, resolve }
}

const flushPromises = async () => {
  await Promise.resolve()
  await Promise.resolve()
}

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

  // TestBugFB113_Phase1CompleteResultReachesReportApprovalGate
  // 对应：docs/regression-tests.md #FB-113
  it('allows a complete phase-1 result to reach the backend report approval gate', () => {
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    expect(wrapper.vm.isReportSelectable({ isComplete: 1, productVersion: 'competency-frontline-phase1-v1' })).toBe(true)
    expect(wrapper.vm.isReportSelectable({ isComplete: 1, productVersion: 'competency-generic-v1' })).toBe(true)
    expect(wrapper.vm.isReportSelectable({ isComplete: 0, productVersion: 'competency-frontline-phase1-v1' })).toBe(false)
  })

  it('shows the backend approval reason when phase-1 batch generation is rejected', async () => {
    generateCompetencyReport.mockReset().mockRejectedValue(new Error('一期正式报告内容尚未完成双重批准'))
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), warning: vi.fn(), error: vi.fn() }
    wrapper.vm.selectedRows = [{
      paperId: 'paper-phase1-complete', participantName: '一期受测者', isComplete: 1,
      productVersion: 'competency-frontline-phase1-v1'
    }]

    await wrapper.vm.batchGenerateReports()

    expect(generateCompetencyReport).toHaveBeenCalledWith({ paperId: 'paper-phase1-complete', force: true })
    expect(wrapper.vm.$message.warning).toHaveBeenCalledWith('批量生成完成：成功0份，失败1份；原因：一期正式报告内容尚未完成双重批准')
    const apiSource = fs.readFileSync(path.resolve(process.cwd(), 'src/api/competency/index.js'), 'utf8')
    expect(apiSource).toContain("rejectWithBusinessMessage: true")
  })

  // TestBugFB151_ExplicitGenerateRebuildsCompletedReport
  // 对应：docs/regression-tests.md #FB-151
  // 复现：模板更新后点击“生成报告”仍发送force=false，后端直接复用旧PDF。
  // 期望：单份和批量“生成报告”都显式force=true，下载动作仍只下载当前实例。
  it('regenerates and downloads a complete temporary competency PDF report', async () => {
    generateCompetencyReport.mockReset().mockResolvedValue({ data: { id: 'report-1' } })
    downloadCompetencyReport.mockReset().mockResolvedValue(new Blob(['pdf'], { type: 'application/pdf' }))
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), error: vi.fn() }
    await wrapper.vm.generateReport({ paperId: 'paper-complete-1', isComplete: 1 })
    expect(generateCompetencyReport).toHaveBeenCalledWith({ paperId: 'paper-complete-1', force: true })
    await wrapper.vm.downloadReport({ paperId: 'paper-complete-1', isComplete: 1 })
    expect(downloadCompetencyReport).toHaveBeenCalledWith('paper-complete-1')
  })

  it('shows start time, completion time, duration and dimension score sum', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/competencyResults.vue'), 'utf8')
    expect(source).toContain('label="开始时间"')
    expect(source).toContain('scope.row.startedAt')
    expect(source).toContain('label="完成时间"')
    expect(source).toContain('scope.row.userTime')
    expect(source).toContain('label="得分合计"')
    expect(source).toContain('prop="scoreSum"')
  })

  // TestBugFB221_ManagementResultsLabelPercentageScores
  // 对应：docs/regression-tests.md #FB-221
  it('labels the latest overall, module and dimension results as percentage scores', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/competencyResults.vue'), 'utf8')
    for (const label of ['整体分（百分制）', '模块分（百分制）', '维度分（百分制）']) {
      expect(source).toContain(label)
    }
    expect(source).toContain("row.scoreScale === 'percentage'")
    expect(source).toContain("this.detail.scoreScale === 'percentage'")
  })

  // TestBugFB159_ResultListHidesEvaluationAverage
  // 对应：docs/regression-tests.md #FB-159
  // 范围：只删除胜任力结果列表列，不改变报告、详情、导出或评分字段。
  it('does not display the evaluation-average column in the competency result list', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/competencyResults.vue'), 'utf8')
    expect(source).not.toContain('label="评价均值"')
    expect(source).not.toContain('scope.row.evaluationAverage')
  })

  // TestBugFB123_CompetencyScoresUseTwoDecimals
  // 对应：docs/regression-tests.md #FB-123
  it('formats all aggregate scores with two decimal places', () => {
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    expect(wrapper.vm.formatScore(4)).toBe('4.00')
    expect(wrapper.vm.formatScore(3.5)).toBe('3.50')
    expect(wrapper.vm.formatScore('4.625')).toBe('4.63')
    expect(wrapper.vm.formatScore(null)).toBe('—')
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/competencyResults.vue'), 'utf8')
    for (const value of ['scope.row.overallScore', 'scope.row.sortDimensionScore', 'selectedRow.overallScore', 'scope.row.groupScore', 'scope.row.scoreSum', 'scope.row.dimensionScore']) {
      expect(source).toContain(`formatScore(${value})`)
    }
  })

  it('shows and filters phase-1 group and validity results for administrators', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/competencyResults.vue'), 'utf8')
    for (const required of ['label="效度状态"', 'v-model="query.validity"', 'value="questionable"', 'scope.row.validityStatus', 'detail.groups', 'detail.validity.validityScore', '35分及以下为效度良好']) {
      expect(source).toContain(required)
    }
  })

  it('matches legacy filtering and row action labels without legacy report APIs', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/competencyResults.vue'), 'utf8')
    for (const text of ['姓名', '电话', '完成状态', '查询', '重置', '批量生成报告', '批量下载', '查看', '答题详情', '下载']) {
      expect(source).toContain(text)
    }
    expect(source).not.toContain('/exam/api/exam/exam/generate-report')
    expect(source).not.toContain('查看团队报告')
    expect(source).not.toContain('删除报告')
  })

  // TestBugFB087_CompetencyResultsMobileLayout
  // 对应：docs/regression-tests.md #FB-087
  it('uses a mobile full-screen detail dialog and responsive single-column layout', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/competencyResults.vue'), 'utf8')
    expect(source).toContain(":fullscreen=\"$store.state.app.device === 'mobile'\"")
    expect(source).toContain(":column=\"$store.state.app.device === 'mobile' ? 1 : 4\"")
    expect(source).toContain('@media (max-width: 768px)')
    expect(source).toContain('.search-form .el-form-item')
    expect(source).toContain('width: 100%')
  })

  it('resets filters and sorting before reloading the first page', () => {
    const loadResults = vi.fn()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults })
    wrapper.setData({ query: {
      ...wrapper.vm.query, current: 3, name: '张三', telephone: '139', completion: 'complete', validity: 'questionable',
      sortBy: 'dimensionScore', sortDirection: 'asc', dimensionId: 'dimension-1'
    } })
    wrapper.vm.resetQuery()
    expect(wrapper.vm.query).toEqual(expect.objectContaining({
      current: 1, name: '', telephone: '', completion: '', validity: '', sortBy: 'submittedAt', sortDirection: 'desc', dimensionId: ''
    }))
    expect(loadResults).toHaveBeenCalled()
  })

  it('batch-generates and batch-downloads every selected complete result', async () => {
    generateCompetencyReport.mockClear()
    downloadCompetencyReportsArchive.mockClear()
    saveAs.mockClear()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), warning: vi.fn(), error: vi.fn() }
    wrapper.setData({ selectedRows: [
      { paperId: 'paper-1', participantName: '甲', isComplete: 1 },
      { paperId: 'paper-2', participantName: '乙', isComplete: 1 }
    ] })
    await wrapper.vm.batchGenerateReports()
    expect(generateCompetencyReport).toHaveBeenCalledTimes(2)
    expect(generateCompetencyReport).toHaveBeenNthCalledWith(1, { paperId: 'paper-1', force: true })
    generateCompetencyReport.mockClear()
    await wrapper.vm.batchDownloadReports()
    expect(generateCompetencyReport).toHaveBeenCalledTimes(2)
    expect(generateCompetencyReport).toHaveBeenNthCalledWith(1, { paperId: 'paper-1', force: false })
    expect(generateCompetencyReport).toHaveBeenNthCalledWith(2, { paperId: 'paper-2', force: false })
    expect(downloadCompetencyReportsArchive).toHaveBeenCalledTimes(1)
    expect(downloadCompetencyReportsArchive).toHaveBeenCalledWith(['paper-1', 'paper-2'])
    expect(saveAs).toHaveBeenCalledTimes(1)
    expect(saveAs.mock.calls[0][0].type).toBe('application/zip')
    expect(saveAs.mock.calls[0][1]).toMatch(/\.zip$/)
    expect(wrapper.vm.reportLoading).toBe(false)
  })

  // TestBugFB157_BatchDownloadUsesOneZipArchive
  // 对应：docs/regression-tests.md #FB-157
  // 复现：勾选多份报告后，页面循环保存多个PDF，浏览器连续触发多个下载。
  // 期望：一次请求全部paperId，浏览器只保存一个包含所选PDF的ZIP。
  it('downloads selected reports as one zip instead of triggering one browser download per PDF', async () => {
    downloadCompetencyReportsArchive.mockClear()
    saveAs.mockClear()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), warning: vi.fn(), error: vi.fn() }
    wrapper.vm.selectedRows = [
      { paperId: 'paper-zip-1', participantName: '甲', isComplete: 1 },
      { paperId: 'paper-zip-2', participantName: '乙', isComplete: 1 }
    ]

    await wrapper.vm.batchDownloadReports()

    expect(downloadCompetencyReportsArchive).toHaveBeenCalledTimes(1)
    expect(downloadCompetencyReportsArchive).toHaveBeenCalledWith(['paper-zip-1', 'paper-zip-2'])
    expect(saveAs).toHaveBeenCalledTimes(1)
    expect(wrapper.vm.$message.success).toHaveBeenCalledWith('批量下载完成，共2份，已打包为ZIP')
  })

  // TestBugFB220_BatchDownloadGeneratesMissingReports
  // 对应：docs/regression-tests.md #FB-220
  // 复现：完整答卷尚未生成报告时直接点击“批量下载”，ZIP接口返回“报告尚未生成”。
  // 期望：批量下载先为冻结的所选答卷幂等确保报告存在，再请求一次ZIP。
  it('generates missing reports before requesting the batch-download archive', async () => {
    generateCompetencyReport.mockClear()
    downloadCompetencyReportsArchive.mockClear()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), warning: vi.fn(), error: vi.fn() }
    wrapper.vm.selectedRows = [
      { paperId: 'paper-missing-1', participantName: '甲', isComplete: 1 },
      { paperId: 'paper-missing-2', participantName: '乙', isComplete: 1 }
    ]

    await wrapper.vm.batchDownloadReports()

    expect(generateCompetencyReport).toHaveBeenCalledTimes(2)
    expect(generateCompetencyReport).toHaveBeenNthCalledWith(1, { paperId: 'paper-missing-1', force: false })
    expect(generateCompetencyReport).toHaveBeenNthCalledWith(2, { paperId: 'paper-missing-2', force: false })
    expect(downloadCompetencyReportsArchive).toHaveBeenCalledTimes(1)
    expect(generateCompetencyReport.mock.invocationCallOrder[1]).toBeLessThan(downloadCompetencyReportsArchive.mock.invocationCallOrder[0])
  })

  it('does not request a partial archive when preparing one selected report fails', async () => {
    generateCompetencyReport.mockReset()
      .mockResolvedValueOnce({ data: { id: 'report-1' } })
      .mockRejectedValueOnce(new Error('报告内容未批准'))
    downloadCompetencyReportsArchive.mockClear()
    saveAs.mockClear()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), warning: vi.fn(), error: vi.fn() }
    wrapper.vm.selectedRows = [
      { paperId: 'paper-ready', participantName: '甲', isComplete: 1 },
      { paperId: 'paper-blocked', participantName: '乙', isComplete: 1 }
    ]

    await wrapper.vm.batchDownloadReports()

    expect(downloadCompetencyReportsArchive).not.toHaveBeenCalled()
    expect(saveAs).not.toHaveBeenCalled()
    expect(wrapper.vm.$message.error).toHaveBeenCalledWith('报告内容未批准')
    expect(wrapper.vm.reportLoading).toBe(false)
  })

  // TestBugFB082_BatchGenerationUsesSelectionSnapshot
  // 对应：docs/regression-tests.md #FB-082
  it('keeps the original batch-generation targets when selection changes during the task', async () => {
    const first = deferred()
    generateCompetencyReport.mockReset()
      .mockReturnValueOnce(first.promise)
      .mockResolvedValueOnce({ data: { id: 'report-2' } })
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), warning: vi.fn(), error: vi.fn() }
    wrapper.vm.selectedRows = [
      { paperId: 'paper-1', participantName: '甲', isComplete: 1 },
      { paperId: 'paper-2', participantName: '乙', isComplete: 1 }
    ]

    const task = wrapper.vm.batchGenerateReports()
    await flushPromises()
    wrapper.vm.selectedRows = [{ paperId: 'paper-3', participantName: '丙', isComplete: 1 }]
    first.resolve({ data: { id: 'report-1' } })
    await task

    expect(generateCompetencyReport).toHaveBeenCalledTimes(2)
    expect(generateCompetencyReport).toHaveBeenNthCalledWith(1, { paperId: 'paper-1', force: true })
    expect(generateCompetencyReport).toHaveBeenNthCalledWith(2, { paperId: 'paper-2', force: true })
    expect(wrapper.vm.$message.success).toHaveBeenCalledWith('批量生成完成，共2份')
  })

  // TestBugFB082_BatchDownloadUsesSelectionSnapshot
  // 对应：docs/regression-tests.md #FB-082
  it('keeps the original batch-download targets when selection changes during the task', async () => {
    const first = deferred()
    const archive = new Blob(['zip'], { type: 'application/zip' })
    downloadCompetencyReportsArchive.mockReset().mockReturnValueOnce(first.promise)
    saveAs.mockClear()
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.$message = { success: vi.fn(), warning: vi.fn(), error: vi.fn() }
    wrapper.vm.selectedRows = [
      { paperId: 'paper-1', participantName: '甲', isComplete: 1 },
      { paperId: 'paper-2', participantName: '乙', isComplete: 1 }
    ]

    const task = wrapper.vm.batchDownloadReports()
    await flushPromises()
    wrapper.vm.selectedRows = []
    first.resolve(archive)
    await task

    expect(downloadCompetencyReportsArchive).toHaveBeenCalledTimes(1)
    expect(downloadCompetencyReportsArchive).toHaveBeenCalledWith(['paper-1', 'paper-2'])
    expect(saveAs).toHaveBeenCalledTimes(1)
    expect(wrapper.vm.$message.success).toHaveBeenCalledWith('批量下载完成，共2份，已打包为ZIP')
  })

  // TestBugFB085_ReportDownloadNamesAreUnique
  // 对应：docs/regression-tests.md #FB-085
  it('uses paper ids to keep same-name participant downloads unique', async () => {
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    const names = [
      wrapper.vm.reportDownloadFileName({ paperId: 'paper-unique-1', participantName: '同名人员' }),
      wrapper.vm.reportDownloadFileName({ paperId: 'paper-unique-2', participantName: '同名人员' })
    ]
    expect(new Set(names).size).toBe(2)
    expect(names[0]).toContain('paper-unique-1')
    expect(names[1]).toContain('paper-unique-2')
  })

  // TestBugFB081_LatestResultRequestWins
  // 对应：docs/regression-tests.md #FB-081
  it('ignores an older result response that arrives after the latest filter response', async () => {
    const older = deferred()
    const latest = deferred()
    fetchCompetencyResults.mockReset()
      .mockReturnValueOnce(older.promise)
      .mockReturnValueOnce(latest.promise)
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })
    wrapper.vm.loadResults = CompetencyResults.methods.loadResults.bind(wrapper.vm)

    wrapper.vm.query.name = '旧条件'
    wrapper.vm.loadResults()
    wrapper.vm.query.name = '新条件'
    wrapper.vm.loadResults()
    latest.resolve({ data: { records: [{ paperId: 'latest' }], total: 1 } })
    await flushPromises()
    older.resolve({ data: { records: [{ paperId: 'older' }], total: 1 } })
    await flushPromises()

    expect(wrapper.vm.rows.map(row => row.paperId)).toEqual(['latest'])
    expect(wrapper.vm.loading).toBe(false)
  })

  // TestBugFB081_LatestDetailRequestWins
  // 对应：docs/regression-tests.md #FB-081
  it('ignores an older detail response after another participant is opened', async () => {
    const older = deferred()
    const latest = deferred()
    fetchCompetencyResultDetail.mockReset()
      .mockReturnValueOnce(older.promise)
      .mockReturnValueOnce(latest.promise)
    const wrapper = mountPage({ loadExam: vi.fn(), loadResults: vi.fn() })

    wrapper.vm.showDetail({ paperId: 'older', participantName: '甲' })
    wrapper.vm.showDetail({ paperId: 'latest', participantName: '乙' })
    latest.resolve({ data: { result: { paperId: 'latest' }, dimensions: [], questions: [] } })
    await flushPromises()
    older.resolve({ data: { result: { paperId: 'older' }, dimensions: [], questions: [] } })
    await flushPromises()

    expect(wrapper.vm.selectedRow.paperId).toBe('latest')
    expect(wrapper.vm.detail.result.paperId).toBe('latest')
    expect(wrapper.vm.detailLoading).toBe(false)
  })
})

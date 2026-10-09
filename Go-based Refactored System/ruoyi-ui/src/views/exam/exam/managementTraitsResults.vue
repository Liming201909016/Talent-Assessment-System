<template>
  <div v-loading="loading" class="app-container management-traits-results" :aria-busy="loading">
    <div class="toolbar">
      <div><h2>管理特质测评结果 <el-tag type="warning" size="small">TEST</el-tag></h2><p class="hint">{{ examTitle || '正在加载测评名称…' }}</p></div>
      <div><management-traits-resume-dialog v-if="allowed" :exam-id="examId" /><el-button size="small" icon="el-icon-refresh" :loading="loading" @click="loadResults">刷新</el-button><el-button size="small" icon="el-icon-back" @click="$router.back()">返回</el-button></div>
    </div>
    <el-alert title="客户模板报告（TEST）仅用于系统测试，不可作为人才决策依据；不提供正式报告、旧API回退或历史重算。" type="warning" :closable="false" show-icon />
    <p class="hint">当前接口最多返回 200 条已提交结果；筛选和分页仅作用于本次返回数据。姓名、手机号和逐题答案未由当前接口提供，不以人员编号冒充姓名。</p>
    <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
    <template v-if="ready && allowed">
      <el-form :inline="true" :model="query" size="small" class="search-form" @submit.native.prevent="handleQuery">
        <el-form-item label="完成状态"><el-select v-model="query.status" clearable placeholder="全部"><el-option label="已完成" value="completed" /><el-option label="未完整作答" value="incomplete" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">查询</el-button><el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button></el-form-item>
      </el-form>
      <div class="table-scroll">
        <el-table :data="pageRows" border size="small" empty-text="暂无已提交管理特质结果">
          <el-table-column label="人员 / 分数详情" min-width="150"><template slot-scope="scope"><span class="hint">身份信息未提供</span><br /><el-button size="mini" type="text" icon="el-icon-tickets" :disabled="detailLoading" @click="showDetail(scope.row)">分数详情</el-button></template></el-table-column>
          <el-table-column label="人员类型" width="105" align="center"><template slot-scope="scope">{{ participantType(scope.row.participantType) }}</template></el-table-column>
          <el-table-column label="完成状态" width="115" align="center"><template slot-scope="scope"><el-tag :type="scope.row.status === 'completed' ? 'success' : 'warning'" size="mini">{{ statusText(scope.row.status) }}</el-tag></template></el-table-column>
          <el-table-column label="已答/题数" width="100" align="right"><template slot-scope="scope">{{ scope.row.answeredQuestionCount }}/{{ scope.row.totalQuestionCount }}</template></el-table-column>
          <el-table-column label="总体分" width="95" align="right"><template slot-scope="scope">{{ formatScore(scope.row.overallScore) }}</template></el-table-column>
          <el-table-column label="完成时间" min-width="160"><template slot-scope="scope">{{ scope.row.submittedAt ? parseTime(scope.row.submittedAt) : '—' }}</template></el-table-column>
          <el-table-column label="用时(秒)" width="95" align="right"><template slot-scope="scope">{{ scope.row.userTimeSeconds == null ? '—' : scope.row.userTimeSeconds }}</template></el-table-column>
          <el-table-column label="报告状态" min-width="125" align="center"><template slot-scope="scope"><el-button type="text" size="mini" :loading="reportState(scope.row).loading" @click="inspectReports(scope.row)">{{ reportStatus(scope.row) }}</el-button></template></el-table-column>
          <el-table-column label="报告操作" width="210" align="center"><template slot-scope="scope"><el-button size="mini" type="text" icon="el-icon-document" :title="generationReason(scope.row)" :disabled="!canGenerate(scope.row) || generating || reportLoading" :loading="generating && generatingRunId === scope.row.id" @click="generateReport(scope.row)">生成</el-button><el-button size="mini" type="text" icon="el-icon-view" :disabled="reportLoading || generating" @click="openReport(scope.row)">查看</el-button><el-button size="mini" type="text" icon="el-icon-download" :disabled="reportLoading || generating" @click="downloadReport(scope.row)">下载</el-button></template></el-table-column>
        </el-table>
      </div>
      <el-pagination class="pagination" :current-page="page" :page-size="pageSize" :page-sizes="[10,20,50,100]" :total="filteredRows.length" layout="total, sizes, prev, pager, next, jumper" @current-change="page = $event" @size-change="changePageSize" />
      <p class="hint" aria-live="polite">报告按行点击查询，不自动对全表请求；未查询不代表未生成。未完整作答不能生成报告。</p>
    </template>
    <el-dialog title="客户模板报告（TEST）" :visible.sync="reportsVisible" width="70%" :fullscreen="isMobile" append-to-body>
      <template v-if="selectedRow">
        <el-alert title="仅用于系统测试，不可作为人才决策依据；独立版本不会覆盖旧报告。" type="warning" :closable="false" show-icon />
        <el-alert v-if="reportState(selectedRow).error" :title="reportState(selectedRow).error" type="error" :closable="false" show-icon />
        <el-button size="small" :loading="reportState(selectedRow).loading" @click="loadReports(selectedRow)">刷新报告状态</el-button>
        <el-select v-if="reportState(selectedRow).reports.length" v-model="reportState(selectedRow).selectedId" placeholder="请选择已归档报告版本" class="report-select">
          <el-option v-for="(item, index) in reportState(selectedRow).reports" :key="item.id" :value="item.id" :label="`版本 ${index + 1} · ${item.createdAt} · ${item.id}`" />
        </el-select>
        <p v-else-if="reportState(selectedRow).loaded && !reportState(selectedRow).error">尚无客户模板报告；仅完整140题已完成结果可生成。</p>
        <el-button size="small" :disabled="generating || reportLoading || !canGenerate(selectedRow)" :loading="generating" @click="generateReport(selectedRow)">生成 TEST</el-button>
        <el-button size="small" :disabled="reportLoading || generating || !reportState(selectedRow).selectedId" @click="openReport(selectedRow)">查看所选版本</el-button>
        <el-button size="small" :disabled="reportLoading || generating || !reportState(selectedRow).selectedId" @click="downloadReport(selectedRow)">下载所选版本</el-button>
      </template>
    </el-dialog>
    <el-dialog title="管理特质分数详情（不是百分位）" :visible.sync="detailVisible" width="85%" :fullscreen="isMobile" append-to-body>
      <div v-loading="detailLoading">
        <el-alert v-if="detailError" :title="detailError" type="error" :closable="false" show-icon />
        <template v-if="detail">
          <p class="identity">Run ID：{{ detail.RunID }}；Paper ID：{{ detail.PaperID }}；Exam ID：{{ detail.ExamID }}</p>
          <p>已答/题数：{{ detail.Result.AnsweredQuestionCount }}/{{ detail.Result.TotalQuestionCount }}；用时 {{ detail.UserTimeSeconds }} 秒；提交 {{ detail.SubmittedAt }}</p>
          <p>总体分：{{ formatScore(detail.Result.OverallScore) }}；常模：{{ formatScore(detail.Result.OverallNorm) }}；等级：{{ detail.Result.OverallLevel || '—' }}。精确值 {{ formatRat(detail.Result.OverallScore) }}；不完整作答不产生分数或报告。</p>
          <el-table :data="detail.Result.Dimensions || []" border size="small">
            <el-table-column label="维度" prop="Name" min-width="120" />
            <el-table-column label="已答/题数" width="100" align="right"><template slot-scope="scope">{{ scope.row.AnsweredCount }}/{{ scope.row.QuestionCount }}</template></el-table-column>
            <el-table-column label="计分合计" prop="ScoreSum" width="100" align="right" />
            <el-table-column label="得分" min-width="110" align="right"><template slot-scope="scope">{{ formatScore(scope.row.Score) }}</template></el-table-column>
            <el-table-column label="常模" min-width="110" align="right"><template slot-scope="scope">{{ formatScore(scope.row.Norm) }}</template></el-table-column>
            <el-table-column label="等级" prop="Level" min-width="100" />
          </el-table>
          <el-table :data="detail.Result.Modules || []" border size="small" class="module-table"><el-table-column label="模块" prop="Key" /><el-table-column label="维度数" prop="DimensionCount" align="right" /><el-table-column label="得分" align="right"><template slot-scope="scope">{{ formatScore(scope.row.Score) }}</template></el-table-column></el-table>
        </template>
      </div>
    </el-dialog>
    <el-dialog title="TEST PDF — 不可作为人才决策依据" :visible.sync="pdfVisible" width="90%" :fullscreen="isMobile" append-to-body @closed="closePdf"><iframe v-if="pdfUrl" :src="pdfUrl" title="管理特质 TEST PDF" class="pdf-frame" /></el-dialog>
  </div>
</template>

<script>
import { fetchManagementTraitsAdminAccess, fetchManagementTraitsExamConfig, fetchManagementTraitsProfile, fetchManagementTraitsResults, fetchManagementTraitsResult, fetchManagementTraitsReissueQualification, fetchManagementTraitsReissues, generateManagementTraitsReissue, viewManagementTraitsReissue, downloadManagementTraitsReissue } from '@/api/managementTraits'
import ManagementTraitsResumeDialog from './managementTraitsResumeDialog.vue'
import { classifyManagementTraitsExam } from '@/utils/managementTraitsProduct'

const managementTraitsTestPurpose = '仅供系统测试，不可作为人才决策依据'

export default {
  name: 'ManagementTraitsResults',
  components: { ManagementTraitsResumeDialog },
  data() {
    return { rows: [], page: 1, pageSize: 20, query: { status: '' }, examTitle: '', adminAccess: false, resultSequence: 0, detailSequence: 0, loadingExamId: '', reportStates: {}, selectedRow: null, reportsVisible: false, loading: false, ready: false, error: '', detail: null, detailError: '', detailVisible: false, detailLoading: false, generating: false, generatingRunId: '', reportLoading: false, pdfUrl: '', pdfVisible: false }
  },
  computed: {
    examId() { return this.$route.params.examId },
    allowed() { return this.adminAccess },
    isMobile() { return this.$store.state.app && this.$store.state.app.device === 'mobile' },
    filteredRows() { return this.rows.filter(row => !this.query.status || row.status === this.query.status) },
    pageRows() { return this.filteredRows.slice((this.page - 1) * this.pageSize, this.page * this.pageSize) }
  },
  created() { this.loadResults() },
  beforeDestroy() { this.resultSequence++; this.detailSequence++; this.closePdf() },
  watch: { examId() { this.loadResults() } },
  methods: {
    changePageSize(size) { this.pageSize = size; this.page = 1 },
    handleQuery() { this.page = 1 },
    resetQuery() { this.query.status = ''; this.page = 1 },
    participantType(value) { return value === 'candidate' ? '开放测评人员' : value === 'tester' ? '封闭测评人员' : '未知类型' },
    statusText(value) { return value === 'completed' ? '已完成' : value === 'incomplete' ? '未完整作答' : '未知状态' },
    formatRat(value) { return typeof value === 'string' && /^\d+(?:\/[1-9]\d*)?$/.test(value) ? value : '—' },
    formatScore(value) {
      if (value === null || value === undefined || value === '') return '—'
      const raw = String(value)
      if (raw.length > 80 || !/^\d+(?:\.\d+|\/[1-9]\d*)?$/.test(raw)) return '—'
      const parts = raw.split('/'), decimal = parts[0].split('.')
      const numerator = BigInt(decimal.join(''))
      const denominator = parts.length === 2 ? BigInt(parts[1]) : BigInt('1' + '0'.repeat((decimal[1] || '').length))
      const two = BigInt(2), hundred = BigInt(100)
      const scaled = (numerator * hundred * two + denominator) / (denominator * two)
      return `${scaled / hundred}.${String(scaled % hundred).padStart(2, '0')}`
    },
    async loadResults() {
      if (this.loading && this.loadingExamId === this.examId) return
      const sequence = ++this.resultSequence, examId = this.examId
      const current = () => sequence === this.resultSequence && examId === this.examId
      this.loadingExamId = examId
      this.ready = false; this.adminAccess = false; this.rows = []; this.error = ''; this.examTitle = ''; this.reportStates = {}; this.selectedRow = null
      this.detailSequence++; this.detail = null; this.detailVisible = false; this.detailLoading = false; this.reportsVisible = false; this.closePdf()
      this.generating = false; this.reportLoading = false
      if (!examId) { this.error = '缺少明确测评 ID。'; return }
      this.loading = true
      try {
        const auth = await fetchManagementTraitsAdminAccess()
        if (!current()) return
        const uid = auth && auth.user && auth.user.userId
        if (!Number.isSafeInteger(uid) || uid <= 0 || !(uid === 1 || (Array.isArray(auth.permissions) && auth.permissions.includes('*:*:*')))) throw new Error('当前账号没有管理员权限。')
        this.adminAccess = true
        const info = await fetchManagementTraitsExamConfig(examId)
        if (!current()) return
        const exam = info && info.data
        if (!['NEW005', 'FROZEN_COMPAT002'].includes(classifyManagementTraitsExam(exam, examId))) throw new Error('服务器未确认此管理特质测评已冻结。')
        if (exam.assessmentType !== 'legacy' || exam.scoringMode !== 'legacy' || (exam.managementTraitsLifecycle !== undefined && exam.managementTraitsLifecycle !== 'frozen') || (exam.isManagementTraits !== undefined && exam.isManagementTraits !== true)) throw new Error('测评类型与冻结生命周期不一致，请核验配置。')
        const fields = typeof exam.requiredFields === 'string' ? exam.requiredFields.split(',') : []
        if (!fields.length || new Set(fields).size !== fields.length || fields.some(f => !['name', 'gender', 'telephone', 'affiliation', 'post', 'age', 'degree', 'major', 'stuFlag'].includes(f))) throw new Error('冻结身份字段配置无效。')
        this.examTitle = exam.title || '管理特质测评'
        const profile = await fetchManagementTraitsProfile(examId)
        if (!current()) return
        if (!profile || !profile.data || !profile.data.frozenAt || profile.data.examId !== examId) throw new Error('此测评未冻结 TEST profile，请返回配置页核验。')
        const response = await fetchManagementTraitsResults(examId)
        if (!current()) return
        if (!response || !Array.isArray(response.data) || response.data.length > 200 || response.data.some(r => r.examId !== examId || !r.id || !r.paperId || !['completed', 'incomplete'].includes(r.status))) throw new Error('结果列表响应无效或超过200条上限。')
        this.rows = response.data; this.page = 1; this.ready = true
      } catch (err) { if (current()) this.error = `加载失败：${err.message || err}；可点击刷新重试，不回退旧接口。` }
      finally { if (current()) this.loading = false }
    },
    validRow(row) { return this.allowed && this.ready && row && row.examId === this.examId && !!row.id && !!row.paperId && this.rows.some(r => r.id === row.id && r.paperId === row.paperId && r.examId === row.examId) },
    generationReason(row) { return this.canGenerate(row) ? '生成客户模板报告（TEST），不覆盖旧报告' : '仅完整140/140且已完成的结果可生成；不完整结果无报告。' },
    canGenerate(row) { return this.validRow(row) && row.status === 'completed' && row.totalQuestionCount === 140 && row.answeredQuestionCount === 140 },
    async showDetail(row) {
      if (!this.validRow(row) || this.detailLoading) return
      this.detailLoading = true; this.detail = null; this.detailError = ''; this.detailVisible = true
      const sequence = ++this.detailSequence, scope = this.resultSequence, examId = this.examId
      const current = () => scope === this.resultSequence && sequence === this.detailSequence && examId === this.examId
      try {
        const response = await fetchManagementTraitsResult(row.id), data = response && response.data
        if (!current()) return
        if (!data || data.RunID !== row.id || data.ExamID !== examId || data.PaperID !== row.paperId || !data.Result) throw new Error('结果详情身份不匹配。')
        this.detail = data
      } catch (err) { if (current()) { this.detailError = `详情读取失败：${err.message || err}`; this.$message.error(this.detailError) } }
      finally { if (current()) this.detailLoading = false }
    },
    reportState(row) { const state = this.reportStates[row && row.id]; return state && this.rows.some(r => r === state.owner && r.id === row.id && r.paperId === row.paperId && r.examId === this.examId) ? state : { reports: [], selectedId: '', loaded: false, loading: false, error: '' } },
    reportStatus(row) { const s = this.reportState(row); return s.generatedId && (s.error || !s.loaded) ? '已生成；状态待刷新' : s.error ? '报告服务不可用' : !s.loaded ? '未查询' : s.reports.length ? `已生成 ${s.reports.length} 版` : '未生成' },
    reportMatches(report, row) { return report && report.id && report.runId === row.id && report.paperId === row.paperId && report.examId === row.examId && report.kind === 'verified_frozen_result' && report.status === 'completed' },
    reportError(err) { return [404, 405, 503].includes(err.status) ? '报告服务未发布或未安装，请联系管理员；结果列表仍可使用。' : `报告服务不可用：${err.message || err}；可重试，不回退旧生成器。` },
    loadReports(row, { force = false, generatedId = '' } = {}) {
      if (!this.validRow(row)) return Promise.resolve()
      const previous = this.reportState(row)
      if (previous.loading && !force) return previous.promise
      const scope = this.resultSequence, examId = this.examId
      const owner = this.rows.find(r => r.id === row.id && r.paperId === row.paperId && r.examId === examId)
      const sequence = (previous.requestSeq || 0) + 1
      const state = { ...previous, owner, requestSeq: sequence, generatedId: generatedId || previous.generatedId || '', loading: true, error: '', promise: null }
      const current = () => scope === this.resultSequence && examId === this.examId && this.rows.includes(owner) && this.reportState(row) === state && state.requestSeq === sequence
      this.$set(this.reportStates, row.id, state)
      state.promise = fetchManagementTraitsReissues(row.paperId).then(response => {
        if (!current()) return
        if (!response || !Array.isArray(response.data) || response.data.length > 100 || response.data.some(r => !this.reportMatches(r, row)) || new Set(response.data.map(r => r.id)).size !== response.data.length) throw new Error('报告元数据身份不匹配。')
        if (state.generatedId && !response.data.some(r => r.id === state.generatedId)) throw new Error('已生成报告未出现在当前元数据，请重试刷新。')
        state.reports = response.data; state.loaded = true
        if (generatedId || (!state.selectedId && state.generatedId)) state.selectedId = state.generatedId
        else if (!state.reports.some(r => r.id === state.selectedId)) state.selectedId = state.reports.length === 1 ? state.reports[0].id : ''
      }).catch(err => { if (current()) { state.reports = []; state.selectedId = ''; state.loaded = false; state.error = state.generatedId ? `报告已生成，但元数据刷新失败；可重试刷新。${this.reportError(err)}` : this.reportError(err) } })
        .finally(() => { if (current()) state.loading = false })
      return state.promise
    },
    async inspectReports(row) { if (!this.validRow(row)) return; this.selectedRow = row; this.reportsVisible = true; await this.loadReports(row) },
    async generateReport(row) {
      if (this.generating || this.reportLoading || !this.canGenerate(row)) return
      this.generating = true; this.generatingRunId = row.id
      const scope = this.resultSequence, examId = this.examId
      const owner = this.rows.find(r => r.id === row.id && r.paperId === row.paperId && r.examId === examId)
      const current = () => scope === this.resultSequence && examId === this.examId && this.rows.includes(owner)
      try {
        try { await this.$confirm('生成客户模板报告（TEST）？仅用于系统测试，不可作为人才决策依据；不覆盖旧报告。', '生成 TEST', { type: 'warning' }) } catch (_) { return }
        if (!current() || !this.canGenerate(row)) return
        const qualified = await fetchManagementTraitsReissueQualification(row.id)
        if (!current()) return
        if (!qualified || !qualified.data || qualified.data.eligible !== true || qualified.data.kind !== 'verified_frozen_result' || qualified.data.purpose !== managementTraitsTestPurpose) throw new Error('来源资格尚未通过，不允许生成。')
        const response = await generateManagementTraitsReissue(row.id)
        if (!current()) return
        const result = response && response.data
        if (!result || !this.reportMatches(result.report, row) || typeof result.reused !== 'boolean' || result.purpose !== managementTraitsTestPurpose) throw new Error('客户模板报告元数据无效。')
        await this.loadReports(row, { force: true, generatedId: result.report.id })
        if (!current()) return
        const state = this.reportState(row)
        if (state.reports.some(r => r.id === result.report.id)) state.selectedId = result.report.id
        this.$message.success(`客户模板报告（TEST）${result.reused ? '已复用' : '已生成'}；不可作为人才决策依据。`)
        if (state.error) this.$message.warning(state.error)
      } catch (err) { if (current()) this.$message.error(`生成 TEST 失败：${this.reportError(err)}`) }
      finally { if (current()) { this.generating = false; this.generatingRunId = '' } }
    },
    async readReport(download, row) {
      if (!this.validRow(row) || this.reportLoading || this.generating) return
      const scope = this.resultSequence, examId = this.examId
      const owner = this.rows.find(r => r.id === row.id && r.paperId === row.paperId && r.examId === examId)
      const current = () => scope === this.resultSequence && examId === this.examId && this.rows.includes(owner)
      this.reportLoading = true
      try {
        if (this.reportState(row).loading || !this.reportState(row).loaded || this.reportState(row).error) await this.loadReports(row)
        if (!current()) return
        const state = this.reportState(row)
        if (state.error) throw new Error(state.error)
        const selected = state.reports.find(r => r.id === state.selectedId)
        if (!selected) { this.selectedRow = row; this.reportsVisible = true; this.$message.warning(state.reports.length ? '请先选择已归档报告版本。' : '尚未生成客户模板报告。'); return }
        if (!this.reportMatches(selected, row)) throw new Error('报告身份不匹配。')
        const file = await (download ? downloadManagementTraitsReissue(selected.id) : viewManagementTraitsReissue(selected.id))
        if (!current()) return
        const blob = file && file.blob
        if (!(blob instanceof Blob) || blob.type.split(';')[0] !== 'application/pdf' || blob.size === 0) throw new Error('未返回有效 PDF。')
        const header = await new Promise((resolve, reject) => { const reader = new FileReader(); reader.onload = () => resolve(reader.result); reader.onerror = reject; reader.readAsText(blob.slice(0, 5)) })
        if (!current()) return
        if (header !== '%PDF-') throw new Error('未返回有效 PDF，未保存文件。')
        if (download) {
          const url = URL.createObjectURL(blob), link = document.createElement('a')
          try { link.href = url; link.download = file.filename; document.body.appendChild(link); link.click() }
          finally { link.remove(); setTimeout(() => URL.revokeObjectURL(url), 1000) }
        } else { this.closePdf(); this.pdfUrl = URL.createObjectURL(blob); this.pdfVisible = true }
      } catch (err) { if (current()) this.$message.error(`TEST PDF ${download ? '下载' : '查看'}失败：${this.reportError(err)}`) }
      finally { if (current()) this.reportLoading = false }
    },
    openReport(row) { return this.readReport(false, row) },
    downloadReport(row) { return this.readReport(true, row) },
    closePdf() { if (this.pdfUrl) URL.revokeObjectURL(this.pdfUrl); this.pdfUrl = ''; this.pdfVisible = false }
  }
}
</script>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 12px; }
h2 { font-size: 20px; margin: 0; }
.hint { font-size: 12px; color: #606266; line-height: 1.5; overflow-wrap: anywhere; }
.identity { overflow-wrap: anywhere; }
.pagination, .module-table, .search-form { margin-top: 12px; }
.table-scroll { max-width: 100%; overflow-x: auto; }
.management-traits-results { min-width: 0; max-width: 100%; }
.report-select { display: block; width: 100%; margin: 12px 0; }
.pdf-frame { border: 0; width: 100%; height: 70vh; }
@media (max-width: 767px) { .toolbar { flex-wrap: wrap; } .search-form ::v-deep .el-form-item { display: block; margin-right: 0; } .pagination { overflow-x: auto; } }
</style>
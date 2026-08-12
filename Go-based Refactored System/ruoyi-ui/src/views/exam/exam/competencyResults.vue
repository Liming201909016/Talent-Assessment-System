<template>
  <div class="app-container competency-results">
    <div class="page-heading">
      <div>
        <h2>胜任力测评结果</h2>
        <p>{{ examTitle || '加载中...' }}</p>
      </div>
      <el-button icon="el-icon-back" @click="$router.back()">返回</el-button>
    </div>
    <el-alert v-if="examProductVersion === 'competency-frontline-phase1-v1'" title="一期十页报告模板已接入；仅完整答卷可选择生成。内容包未完成双重批准时，系统会返回具体门禁原因。" type="warning" :closable="false" show-icon class="mb12" />

    <el-form :inline="true" size="small" class="search-form">
      <el-form-item label="姓名">
        <el-input v-model.trim="query.name" placeholder="请输入姓名" clearable @keyup.enter.native="handleQuery" />
      </el-form-item>
      <el-form-item label="电话">
        <el-input v-model.trim="query.telephone" placeholder="请输入电话" clearable @keyup.enter.native="handleQuery" />
      </el-form-item>
      <el-form-item label="完成状态">
        <el-select v-model="query.completion" clearable placeholder="全部" style="width:120px">
          <el-option label="全部" value="all" />
          <el-option label="完整" value="complete" />
          <el-option label="不完整" value="incomplete" />
        </el-select>
      </el-form-item>
    <el-form-item label="效度状态">
    <el-select v-model="query.validity" clearable placeholder="全部" style="width:130px">
      <el-option label="全部" value="all" />
      <el-option label="效度良好" value="good" />
      <el-option label="效度存疑" value="questionable" />
      <el-option label="效度未完成" value="incomplete" />
    </el-select>
    </el-form-item>
      <el-form-item label="排序指标">
        <el-select v-model="query.sortBy" @change="handleSortByChange">
          <el-option label="提交时间" value="submittedAt" />
          <el-option label="整体分" value="overallScore" />
          <el-option label="维度分" value="dimensionScore" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="query.sortBy === 'dimensionScore'" label="胜任力维度">
        <el-select v-model="query.dimensionId" filterable placeholder="请选择维度" style="width:220px" @change="handleQuery">
          <el-option v-for="item in dimensions" :key="item.dimensionId" :label="item.dimensionName" :value="item.dimensionId" />
        </el-select>
      </el-form-item>
      <el-form-item label="排序方向">
        <el-radio-group v-model="query.sortDirection" @change="handleQuery">
          <el-radio-button label="desc">降序</el-radio-button>
          <el-radio-button label="asc">升序</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">查询</el-button>
        <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
      </el-form-item>
    </el-form>

    <div class="batch-toolbar">
      <el-button
        type="primary"
        plain
        icon="el-icon-document"
        size="mini"
        :loading="reportAction === 'generate'"
        :disabled="!selectedRows.length || reportLoading"
        @click="batchGenerateReports"
      >批量生成报告</el-button>
      <el-button
        type="success"
        plain
        icon="el-icon-download"
        size="mini"
        :loading="reportAction === 'download'"
        :disabled="!selectedRows.length || reportLoading"
        @click="batchDownloadReports"
      >批量下载</el-button>
      <span v-if="reportLoading" class="batch-progress">{{ reportProgress }}</span>
    </div>

    <el-table v-loading="loading" :data="rows" border @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="50" align="center" :selectable="isReportSelectable" />
      <el-table-column label="姓名" prop="participantName" min-width="100" />
      <el-table-column label="手机号" prop="participantTelephone" min-width="120" />
      <el-table-column label="参与者类型" width="100" align="center">
        <template slot-scope="scope">{{ scope.row.participantType === 'tester' ? '测评人员' : '候选人' }}</template>
      </el-table-column>
      <el-table-column label="完成情况" width="130" align="center">
        <template slot-scope="scope">
          <el-tag :type="scope.row.isComplete === 1 ? 'success' : 'warning'" size="mini">
            {{ scope.row.answeredQuestionCount }}/{{ scope.row.totalQuestionCount }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="整体分" prop="overallScore" width="100" align="center" sortable="custom" />
    <el-table-column label="效度状态" width="110" align="center">
    <template slot-scope="scope">
      <el-tag :type="scope.row.validityStatus === 'good' ? 'success' : scope.row.validityStatus === 'questionable' ? 'danger' : 'warning'" size="mini">
      {{ scope.row.validityStatus === 'good' ? '效度良好' : scope.row.validityStatus === 'questionable' ? '效度存疑' : '效度未完成' }}
      </el-tag>
    </template>
    </el-table-column>
    <el-table-column label="效度分" prop="validityScore" width="85" align="center" />
      <el-table-column v-if="query.sortBy === 'dimensionScore'" label="所选维度分" prop="sortDimensionScore" width="120" align="center" />
      <el-table-column label="评价均值" prop="evaluationAverage" width="100" align="center" />
      <el-table-column label="提交方式" width="90" align="center">
        <template slot-scope="scope">{{ scope.row.submitType === 'timeout' ? '到时提交' : '手工提交' }}</template>
      </el-table-column>
      <el-table-column label="开始时间" min-width="160" align="center">
        <template slot-scope="scope">{{ parseTime(scope.row.startedAt) }}</template>
      </el-table-column>
      <el-table-column label="完成时间" min-width="160" align="center">
        <template slot-scope="scope">{{ parseTime(scope.row.submittedAt) }}</template>
      </el-table-column>
      <el-table-column label="答题时长" width="100" align="center">
        <template slot-scope="scope">{{ scope.row.userTime }} 分钟</template>
      </el-table-column>
      <el-table-column label="操作" width="220" align="center" fixed="right">
        <template slot-scope="scope">
          <el-button size="mini" type="text" icon="el-icon-view" :disabled="!isReportSelectable(scope.row)" @click="showReport(scope.row)">查看</el-button>
          <el-button size="mini" type="text" icon="el-icon-tickets" @click="showDetail(scope.row)">答题详情</el-button>
          <el-button size="mini" type="text" icon="el-icon-download" :disabled="!isReportSelectable(scope.row)" @click="downloadReport(scope.row)">下载</el-button>
        </template>
      </el-table-column>
      <template slot="empty"><el-empty description="暂无已提交结果" /></template>
    </el-table>

    <pagination v-show="total > 0" :total="total" :page.sync="query.current" :limit.sync="query.size" @pagination="loadResults" />

    <el-dialog title="胜任力结果详情" :visible.sync="detailVisible" width="85%" :fullscreen="$store.state.app.device === 'mobile'" append-to-body>
      <div v-loading="detailLoading" class="detail-content">
        <el-descriptions v-if="selectedRow" :column="$store.state.app.device === 'mobile' ? 1 : 4" border size="small">
          <el-descriptions-item label="姓名">{{ selectedRow.participantName }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ selectedRow.participantTelephone || '—' }}</el-descriptions-item>
          <el-descriptions-item label="整体分">{{ selectedRow.overallScore }}</el-descriptions-item>
          <el-descriptions-item label="完整性">{{ selectedRow.isComplete === 1 ? '完整' : '不完整' }}</el-descriptions-item>
        </el-descriptions>
        <el-tabs v-if="detail" v-model="detailTab" style="margin-top:16px">
          <el-tab-pane label="一级维度" name="groups">
          <el-table :data="detail.groups" border size="mini">
            <el-table-column label="一级维度" prop="groupName" min-width="150" />
            <el-table-column label="完成维度" width="110" align="center">
            <template slot-scope="scope">{{ scope.row.effectiveDimensionCount }}/{{ scope.row.totalDimensionCount }}</template>
            </el-table-column>
            <el-table-column label="一级得分" prop="groupScore" width="100" align="center" />
            <el-table-column label="等级" prop="levelCode" width="100" align="center" />
          </el-table>
          </el-tab-pane>
          <el-tab-pane label="维度得分" name="dimensions">
            <el-table :data="detail.dimensions" border size="mini" max-height="480">
              <el-table-column label="维度编码" prop="dimensionCode" width="100" />
              <el-table-column label="维度名称" prop="dimensionName" min-width="150" />
              <el-table-column label="完成题数" width="110" align="center">
                <template slot-scope="scope">{{ scope.row.answeredQuestionCount }}/{{ scope.row.totalQuestionCount }}</template>
              </el-table-column>
              <el-table-column label="得分合计" prop="scoreSum" width="100" align="center" />
              <el-table-column label="维度分" prop="dimensionScore" width="100" align="center" />
              <el-table-column label="等级" prop="levelCode" width="100" align="center" />
            </el-table>
          </el-tab-pane>
          <el-tab-pane label="效度" name="validity">
          <el-descriptions v-if="detail.validity" :column="2" border size="small">
            <el-descriptions-item label="效度原始分">{{ detail.validity.validityScore == null ? '—' : detail.validity.validityScore }}</el-descriptions-item>
            <el-descriptions-item label="效度状态">{{ detail.validity.validityStatus === 'good' ? '效度良好' : detail.validity.validityStatus === 'questionable' ? '效度存疑' : '效度未完成' }}</el-descriptions-item>
            <el-descriptions-item label="完成题数">{{ detail.validity.answeredQuestionCount }}/{{ detail.validity.totalQuestionCount }}</el-descriptions-item>
            <el-descriptions-item label="判断阈值">35分及以下为效度良好，36分及以上为效度存疑</el-descriptions-item>
          </el-descriptions>
          </el-tab-pane>
          <el-tab-pane label="逐题审计" name="questions">
            <el-table :data="detail.questions" border size="mini" max-height="480">
              <el-table-column label="#" prop="sort" width="55" align="center" />
              <el-table-column label="题号" prop="questionCode" width="100" />
              <el-table-column label="维度" prop="dimensionName" width="130" />
              <el-table-column label="题目" prop="questionContent" min-width="260" show-overflow-tooltip />
              <el-table-column label="方向" width="70" align="center">
                <template slot-scope="scope">{{ scope.row.scoringDirection === 'reverse' ? '反向' : '正向' }}</template>
              </el-table-column>
              <el-table-column label="原始值" prop="rawValue" width="75" align="center" />
              <el-table-column label="计分值" prop="finalScore" width="75" align="center" />
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { fetchDetail } from '@/api/exam/exam'
import { downloadCompetencyReport, fetchCompetencyResults, fetchCompetencyResultDetail, generateCompetencyReport } from '@/api/competency'
import { saveAs } from 'file-saver'

export default {
  name: 'CompetencyResults',
  data() {
    return {
      loading: false,
      detailLoading: false,
      detailVisible: false,
      detailTab: 'dimensions',
      examTitle: '',
      examProductVersion: '',
      dimensions: [],
      rows: [],
      total: 0,
      selectedRow: null,
      detail: null,
      selectedRows: [],
      reportLoading: false,
      reportAction: '',
      reportProgress: '',
      resultRequestSequence: 0,
      detailRequestSequence: 0,
      query: {
        examId: this.$route.params.examId,
        current: 1,
        size: 20,
        name: '',
        telephone: '',
        completion: '',
        validity: '',
        sortBy: 'submittedAt',
        sortDirection: 'desc',
        dimensionId: ''
      }
    }
  },
  created() {
    this.loadExam()
    this.loadResults()
  },
  methods: {
    loadExam() {
      fetchDetail(this.query.examId).then(response => {
        const exam = response.data || {}
        this.examTitle = exam.title || ''
        this.examProductVersion = exam.competencyProductVersion || ''
        this.dimensions = exam.competencyDimensions || []
      })
    },
    loadResults() {
      const requestSequence = ++this.resultRequestSequence
      this.selectedRows = []
      if (this.query.sortBy === 'dimensionScore' && !this.query.dimensionId) {
        this.rows = []
        this.total = 0
        this.loading = false
        return
      }
      this.loading = true
      fetchCompetencyResults({ ...this.query }).then(response => {
        if (requestSequence !== this.resultRequestSequence) return
        const data = response.data || {}
        this.rows = data.records || []
        this.total = data.total || 0
      }).finally(() => {
        if (requestSequence === this.resultRequestSequence) this.loading = false
      })
    },
    handleQuery() {
      this.query.current = 1
      this.loadResults()
    },
    resetQuery() {
      Object.assign(this.query, {
        current: 1,
        name: '',
        telephone: '',
        completion: '',
        validity: '',
        sortBy: 'submittedAt',
        sortDirection: 'desc',
        dimensionId: ''
      })
      this.loadResults()
    },
    handleSortByChange(value) {
      if (value !== 'dimensionScore') this.query.dimensionId = ''
    if (value === 'overallScore' || value === 'dimensionScore') {
    if (!this.query.completion) this.query.completion = 'complete'
    if (!this.query.validity) this.query.validity = 'good'
      }
      if (value === 'dimensionScore' && !this.query.dimensionId && this.dimensions.length) {
        this.query.dimensionId = this.dimensions[0].dimensionId
      }
      this.handleQuery()
    },
    showDetail(row) {
      const requestSequence = ++this.detailRequestSequence
      this.selectedRow = row
      this.detail = null
      this.detailTab = 'dimensions'
      this.detailVisible = true
      this.detailLoading = true
      fetchCompetencyResultDetail(row.paperId).then(response => {
        if (requestSequence !== this.detailRequestSequence) return
        this.detail = response.data || null
      }).finally(() => {
        if (requestSequence === this.detailRequestSequence) this.detailLoading = false
      })
    },
    showReport(row) {
      if (!this.isReportSelectable(row)) return
      this.$router.push({ name: 'CompetencyReport', params: { paperId: row.paperId } })
    },
    isReportSelectable(row) {
      return row.isComplete === 1
    },
    handleSelectionChange(rows) {
      this.selectedRows = rows.filter(this.isReportSelectable)
    },
    reportDownloadFileName(row) {
      return `${row.participantName || '受测者'}-${row.paperId}-胜任力临时测试报告.pdf`
    },
    async batchGenerateReports() {
      if (!this.selectedRows.length || this.reportLoading) return
      const targetRows = this.selectedRows.filter(this.isReportSelectable).slice()
      if (!targetRows.length) return
      this.reportLoading = true
      this.reportAction = 'generate'
      let succeeded = 0
      let firstError = ''
      try {
        for (let index = 0; index < targetRows.length; index++) {
          this.reportProgress = `正在生成测评报告（${index + 1}/${targetRows.length}）`
          try {
            await generateCompetencyReport({ paperId: targetRows[index].paperId, force: false })
            succeeded++
          } catch (error) {
            if (!firstError) firstError = error.message || '未知错误'
            // Continue so one failed report does not block the remaining selected rows.
          }
        }
        const failed = targetRows.length - succeeded
        if (failed === 0) this.$message.success(`批量生成完成，共${succeeded}份`)
        else this.$message.warning(`批量生成完成：成功${succeeded}份，失败${failed}份；原因：${firstError}`)
      } finally {
        this.reportLoading = false
        this.reportAction = ''
        this.reportProgress = ''
      }
    },
    async batchDownloadReports() {
      if (!this.selectedRows.length || this.reportLoading) return
      const targetRows = this.selectedRows.filter(this.isReportSelectable).slice()
      if (!targetRows.length) return
      this.reportLoading = true
      this.reportAction = 'download'
      let succeeded = 0
      try {
        for (let index = 0; index < targetRows.length; index++) {
          const row = targetRows[index]
          this.reportProgress = `正在下载测评报告（${index + 1}/${targetRows.length}）`
          try {
            const blob = await downloadCompetencyReport(row.paperId)
            saveAs(blob, this.reportDownloadFileName(row))
            succeeded++
          } catch (error) {
            // Continue so one missing report does not block other downloads.
          }
        }
        const failed = targetRows.length - succeeded
        if (failed === 0) this.$message.success(`批量下载完成，共${succeeded}份`)
        else this.$message.warning(`批量下载完成：成功${succeeded}份，失败${failed}份；失败项请先生成报告`)
      } finally {
        this.reportLoading = false
        this.reportAction = ''
        this.reportProgress = ''
      }
    },
    async generateReport(row) {
      if (!this.isReportSelectable(row)) return
      await generateCompetencyReport({ paperId: row.paperId, force: false })
      this.$message.success('临时测试PDF生成成功')
    },
    async downloadReport(row) {
      if (!this.isReportSelectable(row)) return
      try {
        const blob = await downloadCompetencyReport(row.paperId)
        saveAs(blob, this.reportDownloadFileName(row))
      } catch (error) {
        this.$message.error(error.message || '下载胜任力报告失败')
      }
    }
  }
}
</script>

<style scoped>
.page-heading { display:flex; align-items:flex-start; justify-content:space-between; margin-bottom:14px; }
.page-heading h2 { margin:0 0 4px; color:#303133; font-size:20px; }
.page-heading p { margin:0; color:#909399; }
.search-form { margin-bottom:-2px; }
.search-form .el-input { width:170px; }
.batch-toolbar { display:flex; align-items:center; justify-content:flex-end; gap:8px; margin:4px 0 10px; }
.batch-progress { color:#606266; font-size:13px; }
.detail-content { min-height:240px; }

@media (max-width: 768px) {
  .competency-results { padding: 10px; }
  .page-heading { flex-wrap: wrap; gap: 10px; }
  .page-heading > div { width: 100%; }
  .search-form { display: block; margin-bottom: 8px; }
  .search-form .el-form-item { display: flex; width: 100%; margin-right: 0; }
  .search-form ::v-deep .el-form-item__label { width: 88px; flex: 0 0 88px; }
  .search-form ::v-deep .el-form-item__content { flex: 1; min-width: 0; }
  .search-form .el-input,
  .search-form ::v-deep .el-select { width: 100% !important; }
  .batch-toolbar { justify-content: flex-start; flex-wrap: wrap; }
  .batch-progress { width: 100%; }
  .detail-content { min-height: calc(100vh - 70px); }
}
</style>

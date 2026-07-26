<template>
  <div class="app-container competency-results">
    <div class="page-heading">
      <div>
        <h2>胜任力测评结果</h2>
        <p>{{ examTitle || '加载中...' }}</p>
      </div>
      <el-button icon="el-icon-back" @click="$router.back()">返回</el-button>
    </div>

    <el-form :inline="true" size="small" class="filter-bar">
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
    </el-form>

    <el-table v-loading="loading" :data="rows" border stripe>
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
      <el-table-column v-if="query.sortBy === 'dimensionScore'" label="所选维度分" prop="sortDimensionScore" width="120" align="center" />
      <el-table-column label="评价均值" prop="evaluationAverage" width="100" align="center" />
      <el-table-column label="提交方式" width="90" align="center">
        <template slot-scope="scope">{{ scope.row.submitType === 'timeout' ? '到时提交' : '手工提交' }}</template>
      </el-table-column>
      <el-table-column label="提交时间" min-width="160" align="center">
        <template slot-scope="scope">{{ parseTime(scope.row.submittedAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="210" align="center" fixed="right">
        <template slot-scope="scope">
          <el-button type="text" icon="el-icon-view" @click="showDetail(scope.row)">详情</el-button>
          <el-button v-if="scope.row.isComplete === 1" type="text" icon="el-icon-data-analysis" @click="showReport(scope.row)">测试报告</el-button>
          <el-dropdown v-if="scope.row.isComplete === 1" style="margin-left:10px" @command="handleReportCommand($event, scope.row)">
            <el-button type="text">PDF<i class="el-icon-arrow-down el-icon--right" /></el-button>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item command="generate" icon="el-icon-document-add">生成临时PDF</el-dropdown-item>
              <el-dropdown-item command="download" icon="el-icon-download">下载临时PDF</el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </template>
      </el-table-column>
      <template slot="empty"><el-empty description="暂无已提交结果" /></template>
    </el-table>

    <pagination v-show="total > 0" :total="total" :page.sync="query.current" :limit.sync="query.size" @pagination="loadResults" />

    <el-dialog title="胜任力结果详情" :visible.sync="detailVisible" width="85%" append-to-body>
      <div v-loading="detailLoading" class="detail-content">
        <el-descriptions v-if="selectedRow" :column="4" border size="small">
          <el-descriptions-item label="姓名">{{ selectedRow.participantName }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ selectedRow.participantTelephone || '—' }}</el-descriptions-item>
          <el-descriptions-item label="整体分">{{ selectedRow.overallScore }}</el-descriptions-item>
          <el-descriptions-item label="完整性">{{ selectedRow.isComplete === 1 ? '完整' : '不完整' }}</el-descriptions-item>
        </el-descriptions>
        <el-tabs v-if="detail" v-model="detailTab" style="margin-top:16px">
          <el-tab-pane label="维度得分" name="dimensions">
            <el-table :data="detail.dimensions" border size="mini" max-height="480">
              <el-table-column label="维度编码" prop="dimensionCode" width="100" />
              <el-table-column label="维度名称" prop="dimensionName" min-width="150" />
              <el-table-column label="完成题数" width="110" align="center">
                <template slot-scope="scope">{{ scope.row.answeredQuestionCount }}/{{ scope.row.totalQuestionCount }}</template>
              </el-table-column>
              <el-table-column label="维度分" prop="dimensionScore" width="100" align="center" />
              <el-table-column label="等级" prop="levelCode" width="100" align="center" />
            </el-table>
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
      dimensions: [],
      rows: [],
      total: 0,
      selectedRow: null,
      detail: null,
      query: {
        examId: this.$route.params.examId,
        current: 1,
        size: 10,
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
        this.dimensions = exam.competencyDimensions || []
      })
    },
    loadResults() {
      if (this.query.sortBy === 'dimensionScore' && !this.query.dimensionId) {
        this.rows = []
        this.total = 0
        return
      }
      this.loading = true
      fetchCompetencyResults(this.query).then(response => {
        const data = response.data || {}
        this.rows = data.records || []
        this.total = data.total || 0
      }).finally(() => { this.loading = false })
    },
    handleQuery() {
      this.query.current = 1
      this.loadResults()
    },
    handleSortByChange(value) {
      if (value !== 'dimensionScore') this.query.dimensionId = ''
      if (value === 'dimensionScore' && !this.query.dimensionId && this.dimensions.length) {
        this.query.dimensionId = this.dimensions[0].dimensionId
      }
      this.handleQuery()
    },
    showDetail(row) {
      this.selectedRow = row
      this.detail = null
      this.detailTab = 'dimensions'
      this.detailVisible = true
      this.detailLoading = true
      fetchCompetencyResultDetail(row.paperId).then(response => {
        this.detail = response.data || null
      }).finally(() => { this.detailLoading = false })
    },
    showReport(row) {
      this.$router.push({ name: 'CompetencyReport', params: { paperId: row.paperId } })
    },
    handleReportCommand(command, row) {
      if (command === 'generate') return this.generateReport(row)
      if (command === 'download') return this.downloadReport(row)
    },
    async generateReport(row) {
      await generateCompetencyReport({ paperId: row.paperId, force: false })
      this.$message.success('临时测试PDF生成成功')
    },
    async downloadReport(row) {
      const blob = await downloadCompetencyReport(row.paperId)
      saveAs(blob, `${row.participantName || '受测者'}-胜任力临时测试报告.pdf`)
    }
  }
}
</script>

<style scoped>
.page-heading { display:flex; align-items:flex-start; justify-content:space-between; margin-bottom:18px; }
.page-heading h2 { margin:0 0 6px; color:#303133; font-size:22px; }
.page-heading p { margin:0; color:#909399; }
.filter-bar { padding:14px 16px 0; margin-bottom:16px; background:#f7f9fc; border:1px solid #ebeef5; border-radius:4px; }
.detail-content { min-height:240px; }
</style>

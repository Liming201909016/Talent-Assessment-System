<template>
  <div class="app-container competency-question-list">
    <div class="page-heading">
      <div>
        <h2>00401 胜任力测验题库</h2>
        <p>按48个胜任力维度管理的专用题目视图</p>
      </div>
      <div class="heading-actions">
        <el-button icon="el-icon-setting" @click="$router.push({ name: 'CompetencyDimensionMaintenance' })">维度维护</el-button>
        <el-button type="success" plain icon="el-icon-download" @click="downloadTemplate">下载模板</el-button>
        <el-button type="primary" icon="el-icon-upload2" @click="openImport">导入题目</el-button>
        <el-button icon="el-icon-back" @click="$router.back()">返回题库</el-button>
      </div>
    </div>

    <el-form :inline="true" :model="query" size="small" class="filter-bar">
      <el-form-item label="维度">
        <el-select v-model="query.params.dimensionId" clearable filterable placeholder="全部维度" style="width:220px" @change="handleQuery">
          <el-option v-for="item in dimensions" :key="item.id" :label="`${item.code} ${item.name}`" :value="item.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.params.status" clearable placeholder="全部" style="width:110px" @change="handleQuery">
          <el-option label="启用" :value="0" />
          <el-option label="停用" :value="1" />
        </el-select>
      </el-form-item>
      <el-form-item label="题目编号">
        <el-input v-model="query.params.questionCode" clearable placeholder="如 D01-Q01" @keyup.enter.native="handleQuery" />
      </el-form-item>
      <el-form-item label="题干">
        <el-input v-model="query.params.content" clearable placeholder="输入题干关键词" @keyup.enter.native="handleQuery" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="el-icon-search" @click="handleQuery">查询</el-button>
        <el-button icon="el-icon-refresh" @click="resetQuery">重置</el-button>
      </el-form-item>
    </el-form>

    <el-alert title="题号、所属维度和维度内题号为身份字段，不可修改；内容修改仅影响未来发布，已发布测评继续使用原快照。" type="info" :closable="false" show-icon class="mb12" />

    <el-table v-loading="loading" :data="rows" border stripe>
      <el-table-column label="题目编号" prop="questionCode" width="120" align="center" />
      <el-table-column label="维度" min-width="150">
        <template slot-scope="scope">{{ scope.row.dimensionCode }} {{ scope.row.dimensionName }}</template>
      </el-table-column>
      <el-table-column label="维度内题号" prop="dimensionItemNo" width="100" align="center" />
      <el-table-column label="题目内容" prop="content" min-width="300" show-overflow-tooltip />
      <el-table-column label="考察点" prop="observationPoint" min-width="160" show-overflow-tooltip />
      <el-table-column label="计分方向" width="90" align="center">
        <template slot-scope="scope">
          <el-tag :type="scope.row.scoringDirection === 'reverse' ? 'warning' : 'success'" size="mini">
            {{ scope.row.scoringDirection === 'reverse' ? '反向' : '正向' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80" align="center">
        <template slot-scope="scope">
          <el-tag :type="scope.row.questionStatus === 0 ? 'success' : 'info'" size="mini">
            {{ scope.row.questionStatus === 0 ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80" align="center" fixed="right">
        <template slot-scope="scope"><el-button type="text" size="mini" icon="el-icon-edit" @click="openEdit(scope.row)">编辑</el-button></template>
      </el-table-column>
      <template slot="empty"><el-empty description="暂无胜任力题目" /></template>
    </el-table>

    <pagination v-show="total > 0" :total="total" :page.sync="query.current" :limit.sync="query.size" @pagination="loadQuestions" />

    <el-dialog title="编辑胜任力题目" :visible.sync="editVisible" width="650px" append-to-body>
      <el-descriptions v-if="editVisible" :column="3" border size="small" class="identity-block">
        <el-descriptions-item label="题目编号">{{ editIdentity.questionCode }}</el-descriptions-item>
        <el-descriptions-item label="维度">{{ editIdentity.dimensionCode }} {{ editIdentity.dimensionName }}</el-descriptions-item>
        <el-descriptions-item label="维度内题号">{{ editIdentity.dimensionItemNo }}</el-descriptions-item>
      </el-descriptions>
      <el-form ref="editForm" :model="editForm" :rules="editRules" label-width="90px">
        <el-form-item label="题目内容" prop="content"><el-input v-model="editForm.content" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="考察点" prop="observationPoint"><el-input v-model="editForm.observationPoint" /></el-form-item>
        <el-form-item label="计分方向" prop="scoringDirection">
          <el-radio-group v-model="editForm.scoringDirection"><el-radio label="forward">正向</el-radio><el-radio label="reverse">反向</el-radio></el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="questionStatus"><el-radio-group v-model="editForm.questionStatus"><el-radio :label="0">启用</el-radio><el-radio :label="1">停用</el-radio></el-radio-group></el-form-item>
        <el-form-item label="备注"><el-input v-model="editForm.remark" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <div slot="footer"><el-button @click="editVisible=false">取消</el-button><el-button type="primary" :loading="editSaving" @click="saveEdit">保存</el-button></div>
    </el-dialog>

    <el-dialog title="导入胜任力题目" :visible.sync="importVisible" width="760px" append-to-body @closed="resetImport">
      <el-upload ref="importUpload" action="#" accept=".xlsx" :auto-upload="false" :limit="1" :on-change="selectImportFile" :on-remove="clearImportFile" :file-list="importFileList">
        <el-button icon="el-icon-folder-opened">选择xlsx文件</el-button>
        <span slot="tip" class="el-upload__tip">必须使用专用九列模板，文件不超过10 MiB；先预览校验，再正式导入。</span>
      </el-upload>
      <div class="preview-actions">
        <el-button type="primary" plain :disabled="!importFile" :loading="previewLoading" @click="previewImport">预览校验</el-button>
      </div>
      <template v-if="importPreview">
        <el-alert :title="`校验完成：可导入 ${importPreview.successCount} 行，错误 ${importPreview.errorCount} 行`" :type="importPreview.errorCount ? 'error' : 'success'" :closable="false" show-icon />
        <el-table v-if="importPreview.errorCount" :data="importPreview.errorRows" border size="mini" max-height="300" class="preview-table">
          <el-table-column label="Excel行" prop="rowNumber" width="90" align="center" />
          <el-table-column label="错误信息"><template slot-scope="scope">{{ (scope.row.messages || []).join('；') }}</template></el-table-column>
        </el-table>
        <el-table v-else :data="(importPreview.successRows || []).slice(0, 20)" border size="mini" max-height="300" class="preview-table">
          <el-table-column label="题目编号" prop="questionCode" width="120" />
          <el-table-column label="维度" prop="dimensionName" width="140" />
          <el-table-column label="题目内容" prop="content" show-overflow-tooltip />
          <el-table-column label="方向" width="70"><template slot-scope="scope">{{ scope.row.direction === 'reverse' ? '反向' : '正向' }}</template></el-table-column>
        </el-table>
        <div v-if="!importPreview.errorCount && importPreview.successCount > 20" class="preview-note">仅展示前20行，共{{ importPreview.successCount }}行。</div>
      </template>
      <div slot="footer">
        <el-button @click="importVisible=false">取消</el-button>
        <el-button type="primary" :disabled="!canFormalImport" :loading="importLoading" @click="confirmImport">确认导入</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { saveAs } from 'file-saver'
import { fetchCompetencyDimensions, fetchCompetencyQuestions, updateCompetencyQuestion, downloadCompetencyQuestionTemplate, previewCompetencyQuestions, importCompetencyQuestions } from '@/api/competency'

export default {
  name: 'CompetencyQuestionList',
  data() {
    return {
      loading: false,
      dimensions: [],
      rows: [],
      total: 0,
      editVisible: false,
      editSaving: false,
      editIdentity: {},
      editForm: { id: '', content: '', observationPoint: '', scoringDirection: 'forward', questionStatus: 0, remark: '' },
      editRules: {
        content: [{ required: true, message: '题目内容不能为空', trigger: 'blur' }],
        observationPoint: [{ required: true, message: '考察点不能为空', trigger: 'blur' }],
        scoringDirection: [{ required: true, message: '请选择计分方向', trigger: 'change' }],
        questionStatus: [{ required: true, message: '请选择状态', trigger: 'change' }]
      },
      importVisible: false,
      importFile: null,
      importFileList: [],
      importPreview: null,
      previewLoading: false,
      importLoading: false,
      query: {
        current: 1,
        size: 20,
        params: { dimensionId: '', status: '', questionCode: '', content: '' }
      }
    }
  },
  created() {
    this.loadDimensions()
    this.loadQuestions()
  },
  computed: {
    canFormalImport() {
      return !!(this.importFile && this.importPreview && this.importPreview.sha256 && this.importPreview.errorCount === 0 && !this.previewLoading)
    }
  },
  methods: {
    loadDimensions() {
      fetchCompetencyDimensions().then(response => { this.dimensions = response.data || [] })
    },
    loadQuestions() {
      this.loading = true
      fetchCompetencyQuestions(this.query).then(response => {
        const data = response.data || {}
        this.rows = data.records || []
        this.total = data.total || 0
      }).finally(() => { this.loading = false })
    },
    handleQuery() {
      this.query.current = 1
      this.loadQuestions()
    },
    resetQuery() {
      this.query.params = { dimensionId: '', status: '', questionCode: '', content: '' }
      this.handleQuery()
    },
    openEdit(row) {
      this.editIdentity = { questionCode: row.questionCode, dimensionCode: row.dimensionCode, dimensionName: row.dimensionName, dimensionItemNo: row.dimensionItemNo }
      this.editForm = { id: row.id, content: row.content, observationPoint: row.observationPoint, scoringDirection: row.scoringDirection, questionStatus: row.questionStatus, remark: row.remark || '' }
      this.editVisible = true
      this.$nextTick && this.$nextTick(() => this.$refs.editForm && this.$refs.editForm.clearValidate())
    },
    saveEdit() {
      return new Promise(resolve => {
        this.$refs.editForm.validate(valid => {
          if (!valid) return resolve(false)
          this.editSaving = true
          updateCompetencyQuestion(this.editForm).then(() => {
            this.editVisible = false
            this.$notify({ title: '成功', message: '胜任力题目保存成功', type: 'success' })
            this.loadQuestions()
            resolve(true)
          }).catch(() => resolve(false)).finally(() => { this.editSaving = false })
        })
      })
    },
    downloadTemplate() {
      return downloadCompetencyQuestionTemplate().then(blob => saveAs(blob, 'competency-question-import-template.xlsx'))
    },
    openImport() {
      this.resetImport()
      this.importVisible = true
    },
    selectImportFile(file) {
      const name = String(file.name || '')
      if (!name.toLowerCase().endsWith('.xlsx')) {
        this.$message.warning('只支持xlsx格式文件')
        this.importFile = null
        this.importFileList = []
        this.importPreview = null
        return false
      }
      if (file.size <= 0 || file.size > 10 * 1024 * 1024) {
        this.$message.warning('文件必须大于0且不超过10 MiB')
        this.importFile = null
        this.importFileList = []
        this.importPreview = null
        return false
      }
      this.importFile = file.raw || file
      this.importFileList = [file]
      this.importPreview = null
      return true
    },
    clearImportFile() {
      this.importFile = null
      this.importFileList = []
      this.importPreview = null
    },
    resetImport() {
      this.clearImportFile()
      this.previewLoading = false
      this.importLoading = false
      if (this.$refs && this.$refs.importUpload) this.$refs.importUpload.clearFiles()
    },
    previewImport() {
      if (!this.importFile) {
        this.$message.warning('请先选择xlsx文件')
        return Promise.resolve(false)
      }
      this.previewLoading = true
      return previewCompetencyQuestions(this.importFile).then(response => {
        this.importPreview = response.data || null
        return true
      }).catch(() => false).finally(() => { this.previewLoading = false })
    },
    confirmImport() {
      if (!(this.importFile && this.importPreview && this.importPreview.sha256 && this.importPreview.errorCount === 0 && !this.previewLoading)) return Promise.resolve(false)
      return this.$confirm(`确认导入校验通过的 ${this.importPreview.successCount} 道题目吗？`, '确认导入', { type: 'warning' }).then(() => {
        this.importLoading = true
        return importCompetencyQuestions(this.importFile, this.importPreview.sha256).then(response => {
          const importedCount = response.data && response.data.importedCount
          this.importVisible = false
          this.$notify({ title: '成功', message: `成功导入 ${importedCount || 0} 道胜任力题目`, type: 'success' })
          this.loadQuestions()
          this.loadDimensions()
          return true
        }).catch(() => false).finally(() => { this.importLoading = false })
      }).catch(() => false)
    }
  }
}
</script>

<style scoped>
.page-heading { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:16px; }
.page-heading h2 { margin:0 0 6px; font-size:20px; color:#303133; }
.page-heading p { margin:0; color:#909399; }
.filter-bar { padding:14px 16px 0; margin-bottom:12px; background:#f7f9fc; border:1px solid #ebeef5; border-radius:4px; }
.mb12 { margin-bottom:12px; }
.identity-block { margin-bottom:18px; }
.heading-actions { display:flex; gap:8px; }
.preview-actions { margin:14px 0; }
.preview-table { margin-top:12px; }
.preview-note { margin-top:8px; color:#909399; font-size:12px; }
</style>

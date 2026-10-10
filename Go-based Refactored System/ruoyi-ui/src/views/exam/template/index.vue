<template>
  <div class="app-container" style="padding: 8px 8px;">
    <div class="page-heading">
      <h3>报告模板管理</h3>
      <p>集中管理胜任力、管理特质与 MBTI Word 报告模板。上传后将立即用于后续报告生成。</p>
    </div>

    <el-card class="phase1-card" shadow="never" v-loading="phase1Loading">
      <div slot="header" class="card-header">
        <div>
          <strong>00401 一期胜任力 v2 报告模板</strong>
          <span class="card-subtitle">基层员工版 · 60个可选语义字段 · 12业务图表 · value-only</span>
        </div>
        <el-tag v-if="phase1Template.exists" :type="phase1Template.valid ? 'success' : 'danger'" size="small">
          {{ phase1Template.valid ? '校验通过' : '校验失败' }}
        </el-tag>
        <el-tag v-else type="danger" size="small">模板缺失</el-tag>
      </div>

      <el-row :gutter="16" class="template-meta">
        <el-col :xs="24" :sm="12" :md="6"><span class="meta-label">文件</span><span class="meta-value">{{ phase1Template.fileName || '—' }}</span></el-col>
        <el-col :xs="24" :sm="12" :md="4"><span class="meta-label">大小</span><span class="meta-value">{{ formatFileSize(phase1Template.size) }}</span></el-col>
        <el-col :xs="24" :sm="12" :md="7"><span class="meta-label">修改时间</span><span class="meta-value">{{ phase1Template.modTime || '—' }}</span></el-col>
        <el-col :xs="24" :sm="12" :md="7"><span class="meta-label">模板契约</span><span class="meta-value">{{ phase1ContractText }}</span></el-col>
      </el-row>
      <div class="sha-row"><span class="meta-label">SHA-256</span><code>{{ phase1Template.sha256 || '—' }}</code></div>
      <el-alert v-if="phase1Template.validationError" :title="phase1Template.validationError" type="error" :closable="false" show-icon class="contract-alert" />

      <div class="phase1-actions">
        <el-button icon="el-icon-document" @click="openPhase1SemanticFields">模板字段</el-button>
        <el-button icon="el-icon-download" :loading="phase1Downloading" :disabled="!phase1Template.exists" @click="downloadPhase1Template">下载模板</el-button>
        <el-upload
          ref="phase1Upload"
          action="#"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="selectPhase1File"
          :before-upload="beforePhase1Upload"
          accept=".docx"
        >
          <el-button icon="el-icon-folder-opened">选择文件</el-button>
        </el-upload>
        <span class="selected-file" :title="phase1File ? phase1File.name : ''">{{ phase1File ? phase1File.name : '未选择文件' }}</span>
        <el-button type="primary" icon="el-icon-upload2" :loading="phase1Uploading" :disabled="!phase1File" @click="uploadPhase1Template">上传并生效</el-button>
      </div>
      <p class="upload-hint">仅支持 20MB 以内 DOCX。60个语义字段均可按需添加，同一Tag可在多处重复并显示相同内容；系统仍严格校验已使用Tag、12个业务图表、零公式/引用/外部链接及页码契约。</p>
    </el-card>

    <el-card class="management-traits-card" shadow="never" v-loading="managementTraitsTemplateLoading">
      <div slot="header" class="card-header">
        <div>
          <strong>00501 / 00502 共用报告模板</strong>
          <span class="card-subtitle">基层员工新版 / 干部新版 · 共用一份模板</span>
        </div>
        <el-tag v-if="managementTraitsTemplate.exists" :type="managementTraitsTemplate.valid ? 'success' : 'danger'" size="small">
          {{ managementTraitsTemplate.valid ? '校验通过' : '校验失败' }}
        </el-tag>
        <el-tag v-else type="danger" size="small">模板缺失</el-tag>
      </div>

      <el-row :gutter="16" class="template-meta">
        <el-col :xs="24" :sm="12" :md="6"><span class="meta-label">文件</span><span class="meta-value">{{ managementTraitsTemplate.fileName || '—' }}</span></el-col>
        <el-col :xs="24" :sm="12" :md="4"><span class="meta-label">大小</span><span class="meta-value">{{ formatFileSize(managementTraitsTemplate.size) }}</span></el-col>
        <el-col :xs="24" :sm="12" :md="7"><span class="meta-label">修改时间</span><span class="meta-value">{{ managementTraitsTemplate.modTime || '—' }}</span></el-col>
        <el-col :xs="24" :sm="12" :md="7"><span class="meta-label">模板契约</span><span class="meta-value">{{ managementTraitsContractText }}</span></el-col>
      </el-row>
      <div class="sha-row"><span class="meta-label">SHA-256</span><code>{{ managementTraitsTemplate.sha256 || '—' }}</code></div>
      <el-alert v-if="managementTraitsTemplate.validationError" :title="managementTraitsTemplate.validationError" type="error" :closable="false" show-icon class="contract-alert" />

      <div class="phase1-actions">
        <el-button icon="el-icon-document" @click="openManagementTraitsSemanticFields">模板字段</el-button>
        <el-button icon="el-icon-download" :loading="managementTraitsTemplateDownloading" :disabled="!managementTraitsTemplate.exists" @click="downloadManagementTraitsTemplate">下载模板</el-button>
        <el-upload
          ref="managementTraitsTemplateUpload"
          action="#"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="selectManagementTraitsTemplateFile"
          accept=".docx"
        >
          <el-button icon="el-icon-folder-opened">选择文件</el-button>
        </el-upload>
        <span class="selected-file" :title="managementTraitsTemplateFile ? managementTraitsTemplateFile.name : ''">{{ managementTraitsTemplateFile ? managementTraitsTemplateFile.name : '未选择文件' }}</span>
        <el-button type="primary" icon="el-icon-upload2" :loading="managementTraitsTemplateUploading" :disabled="!managementTraitsTemplateFile" @click="uploadManagementTraitsTemplate">上传并生效</el-button>
      </div>
      <p class="upload-hint">仅支持20MB以内DOCX。模板支持95个可填充语义字段，同一Tag可重复使用；系统校验必需字段、6个业务图表、5个数字标签和零外部关系。上传后仅影响新生成或重发报告。</p>
    </el-card>

    <el-dialog
      :title="semanticFieldDialogTitle"
      :visible.sync="semanticFieldDialogVisible"
      width="760px"
      custom-class="semantic-field-dialog"
      append-to-body
    >
      <el-alert title="在 Word“开发工具 → 控件属性 → 标记”中填写Tag。同一Tag可添加到多个位置，生成报告时这些位置会显示相同内容。" type="info" :closable="false" show-icon />
      <el-input v-model.trim="semanticFieldKeyword" clearable prefix-icon="el-icon-search" placeholder="搜索Tag、名称或描述" class="semantic-field-search" />
      <el-table :data="filteredSemanticFields" border size="mini" max-height="460" empty-text="暂无匹配字段">
        <el-table-column label="Tag" prop="key" min-width="230" show-overflow-tooltip>
          <template slot-scope="scope"><code>{{ scope.row.key }}</code></template>
        </el-table-column>
        <el-table-column label="名称" prop="name" min-width="130" show-overflow-tooltip />
        <el-table-column label="描述" prop="description" min-width="260" show-overflow-tooltip />
      </el-table>
      <span slot="footer"><el-button type="primary" @click="semanticFieldDialogVisible = false">关闭</el-button></span>
    </el-dialog>

    <div style="margin: 16px 0 8px;">
      <h3 style="margin:0 0 4px; font-size:15px">MBTI 报告模板管理</h3>
      <p style="color:#909399; font-size:12px; margin:0">管理 16 种 MBTI 类型的 Word 报告模板。上传 .docx 格式文件，系统会自动用于生成对应类型的报告。</p>
    </div>

    <el-table :data="templates" v-loading="loading" border size="mini">
      <el-table-column label="MBTI 类型" prop="type" width="100" align="center">
        <template slot-scope="scope">
          <el-tag size="medium" :type="scope.row.exists ? '' : 'danger'">{{ scope.row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="完整版" align="center">
        <el-table-column label="文件" prop="fileName" width="160" show-overflow-tooltip />
        <el-table-column label="大小" width="80" align="right">
          <template slot-scope="scope">
            <span v-if="scope.row.exists">{{ (scope.row.size / 1024).toFixed(0) }}KB</span>
            <span v-else style="color:#F56C6C">缺失</span>
          </template>
        </el-table-column>
        <el-table-column label="修改日期" width="160" align="center">
          <template slot-scope="scope">
            <span v-if="scope.row.exists">{{ scope.row.modTime }}</span>
            <span v-else style="color:#C0C4CC">--</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="center">
          <template slot-scope="scope">
            <el-button size="mini" type="text" icon="el-icon-download" :disabled="!scope.row.exists" @click="handleDownload(scope.row)">下载</el-button>
            <el-upload :action="uploadUrl" :data="{ type: scope.row.type }" :headers="uploadHeaders" :show-file-list="false" :on-success="handleUploadSuccess" :on-error="handleUploadError" :before-upload="beforeUpload" accept=".docx" style="display:inline-block; margin-left:4px">
              <el-button size="mini" type="text" icon="el-icon-upload2">上传</el-button>
            </el-upload>
          </template>
        </el-table-column>
      </el-table-column>
      <el-table-column label="简版" align="center">
        <el-table-column label="文件" width="180" show-overflow-tooltip>
          <template slot-scope="scope">{{ scope.row.simpleFileName }}</template>
        </el-table-column>
        <el-table-column label="大小" width="80" align="right">
          <template slot-scope="scope">
            <span v-if="scope.row.simpleExists">{{ (scope.row.simpleSize / 1024).toFixed(0) }}KB</span>
            <span v-else style="color:#F56C6C">缺失</span>
          </template>
        </el-table-column>
        <el-table-column label="修改日期" width="160" align="center">
          <template slot-scope="scope">
            <span v-if="scope.row.simpleExists">{{ scope.row.simpleModTime }}</span>
            <span v-else style="color:#C0C4CC">--</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="center">
          <template slot-scope="scope">
            <el-button size="mini" type="text" icon="el-icon-download" :disabled="!scope.row.simpleExists" @click="handleDownloadSimple(scope.row)">下载</el-button>
            <el-upload :action="uploadUrl" :data="{ type: scope.row.type, variant: 'simple' }" :headers="uploadHeaders" :show-file-list="false" :on-success="handleUploadSuccess" :on-error="handleUploadError" :before-upload="beforeUpload" accept=".docx" style="display:inline-block; margin-left:4px">
              <el-button size="mini" type="text" icon="el-icon-upload2">上传</el-button>
            </el-upload>
          </template>
        </el-table-column>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import request from '@/utils/request'
import { getToken } from '@/utils/auth'
import { downloadPhase1V2WordTemplate, fetchPhase1V2WordTemplate, uploadPhase1V2WordTemplate } from '@/api/competency'
import {
  downloadManagementTraitsTemplate as downloadManagementTraitsTemplateFile,
  fetchManagementTraitsTemplateInfo,
  uploadManagementTraitsTemplate as uploadManagementTraitsTemplateFile
} from '@/api/managementTraits'

export default {
  name: 'MbtiTemplates',
  data() {
    return {
      loading: false,
      templates: [],
      phase1Loading: false,
      phase1Downloading: false,
      phase1Uploading: false,
      phase1File: null,
      phase1Template: {
        exists: false,
        fileName: 'competency-phase1-report-v2.docx',
        size: 0,
        modTime: '',
        sha256: '',
        valid: false,
        validationError: '',
        schemaVersion: '',
        contentControls: 0,
        registeredFields: 0,
        usedFields: 0,
        charts: 0,
        businessCharts: 0,
        embeddedWorkbooks: 0,
        externalLinks: 0,
        visibleTokens: 0,
        semanticFields: []
      },
      managementTraitsTemplateLoading: false,
      managementTraitsTemplateDownloading: false,
      managementTraitsTemplateUploading: false,
      managementTraitsTemplateFile: null,
      managementTraitsTemplate: {
        exists: false,
        fileName: 'management-traits-00501-00502-shared.docx',
        size: 0,
        modTime: '',
        sha256: '',
        valid: false,
        validationError: '',
        productCodes: ['00501', '00502'],
        contentControls: 0,
        businessCharts: 0,
        numericLabels: 0,
        externalLinks: 0,
        semanticFields: []
      },
      semanticFieldDialogVisible: false,
      semanticFieldDialogTitle: '',
      semanticFieldKeyword: '',
      semanticFieldRows: []
    }
  },
  computed: {
    uploadUrl() {
      return process.env.VUE_APP_BASE_API + '/exam/api/mbti/templates/upload'
    },
    uploadHeaders() {
      return { Authorization: 'Bearer ' + getToken() }
    },
    phase1ContractText() {
      if (!this.phase1Template.exists) return '—'
      if (['competency-phase1-template-schema-v2', 'competency-phase1-report-template-v2'].includes(this.phase1Template.schemaVersion)) {
        const workbook = this.phase1Template.embeddedWorkbooks ? ` · ${this.phase1Template.embeddedWorkbooks} 内嵌工作簿` : ''
        return `V2 · ${this.phase1Template.usedFields || 0}/${this.phase1Template.registeredFields || 0} 字段 · ${this.phase1Template.businessCharts || 0} 业务图表${workbook} · ${this.phase1Template.externalLinks || 0} 外链`
      }
      return `${this.phase1Template.contentControls || 0} 控件 / ${this.phase1Template.charts || 0} 图表 / ${this.phase1Template.visibleTokens || 0} 可见占位符`
    },
    managementTraitsContractText() {
      if (!this.managementTraitsTemplate.exists) return '—'
      return `${this.managementTraitsTemplate.contentControls || 0} 控件 / ${this.managementTraitsTemplate.businessCharts || 0} 图表 / ${this.managementTraitsTemplate.numericLabels || 0} 数字标签 / ${this.managementTraitsTemplate.externalLinks || 0} 外链`
    },
    filteredSemanticFields() {
      const keyword = (this.semanticFieldKeyword || '').toLowerCase()
      if (!keyword) return this.semanticFieldRows
      return this.semanticFieldRows.filter(field => [field.key, field.name, field.description].some(value => String(value || '').toLowerCase().includes(keyword)))
    }
  },
  created() {
    this.fetchList()
    this.fetchPhase1Template()
    this.fetchManagementTraitsTemplate()
  },
  methods: {
    openPhase1SemanticFields() {
      this.semanticFieldDialogTitle = '00401 胜任力模板字段'
      this.semanticFieldKeyword = ''
      this.semanticFieldRows = [...(this.phase1Template.semanticFields || [])]
      this.semanticFieldDialogVisible = true
    },
    openManagementTraitsSemanticFields() {
      this.semanticFieldDialogTitle = '00501 / 00502 共用模板字段'
      this.semanticFieldKeyword = ''
      this.semanticFieldRows = [...(this.managementTraitsTemplate.semanticFields || [])]
      this.semanticFieldDialogVisible = true
    },
    async fetchPhase1Template() {
      this.phase1Loading = true
      try {
        const response = await fetchPhase1V2WordTemplate()
        this.phase1Template = { ...this.phase1Template, ...(response.data || {}) }
      } catch (error) {
        this.$message.error(error.message || '读取胜任力模板失败')
      } finally {
        this.phase1Loading = false
      }
    },
    formatFileSize(size) {
      if (!size) return '—'
      return size >= 1024 * 1024 ? `${(size / 1024 / 1024).toFixed(2)} MB` : `${(size / 1024).toFixed(0)} KB`
    },
    phase1DownloadFileName() {
      const base = (this.phase1Template.fileName || 'competency-phase1-report-v2.docx').replace(/\.docx$/i, '')
      const version = (this.phase1Template.sha256 || '').slice(0, 8)
      return `${base}${version ? '-' + version : ''}.docx`
    },
    beforePhase1Upload(file) {
      const isDocx = /\.docx$/i.test(file.name || '')
      const isValidSize = file.size > 0 && file.size <= 20 * 1024 * 1024
      if (!isDocx) this.$message.error('只支持 .docx 格式')
      else if (!isValidSize) this.$message.error('文件大小必须在 20MB 以内')
      return isDocx && isValidSize
    },
    selectPhase1File(uploadFile) {
      const file = uploadFile.raw || uploadFile
      if (!this.beforePhase1Upload(file)) {
        this.phase1File = null
        if (this.$refs && this.$refs.phase1Upload) this.$refs.phase1Upload.clearFiles()
        return
      }
      this.phase1File = file
    },
    async downloadPhase1Template() {
      this.phase1Downloading = true
      try {
        const blob = await downloadPhase1V2WordTemplate()
        const blobUrl = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.style.display = 'none'
        link.href = blobUrl
        link.setAttribute('download', this.phase1DownloadFileName())
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(blobUrl)
      } catch (error) {
        this.$message.error(error.message || '下载胜任力模板失败')
      } finally {
        this.phase1Downloading = false
      }
    },
    async uploadPhase1Template() {
      if (!this.phase1File) return
      try {
        await this.$confirm('上传校验通过后将立即替换当前胜任力报告模板，是否继续？', '确认上传', {
          confirmButtonText: '上传并生效',
          cancelButtonText: '取消',
          type: 'warning'
        })
      } catch (error) {
        if (error !== 'cancel' && error !== 'close') this.$message.error(error.message || '上传已取消')
        return
      }
      this.phase1Uploading = true
      try {
        await uploadPhase1V2WordTemplate(this.phase1File)
        this.$message.success('胜任力报告模板已上传并生效')
        this.phase1File = null
        if (this.$refs && this.$refs.phase1Upload) this.$refs.phase1Upload.clearFiles()
        await this.fetchPhase1Template()
      } catch (error) {
        this.$message.error(error.message || '上传胜任力模板失败')
      } finally {
        this.phase1Uploading = false
      }
    },
    async fetchManagementTraitsTemplate() {
      this.managementTraitsTemplateLoading = true
      try {
        const response = await fetchManagementTraitsTemplateInfo()
        this.managementTraitsTemplate = { ...this.managementTraitsTemplate, ...(response.data || {}) }
      } catch (error) {
        this.$message.error(error.message || '读取00501/00502共用报告模板失败')
      } finally {
        this.managementTraitsTemplateLoading = false
      }
    },
    managementTraitsTemplateDownloadFileName() {
      const base = (this.managementTraitsTemplate.fileName || 'management-traits-00501-00502-shared.docx').replace(/\.docx$/i, '')
      const version = (this.managementTraitsTemplate.sha256 || '').slice(0, 8)
      return `${base}${version ? '-' + version : ''}.docx`
    },
    selectManagementTraitsTemplateFile(uploadFile) {
      const file = uploadFile.raw || uploadFile
      if (!this.beforePhase1Upload(file)) {
        this.managementTraitsTemplateFile = null
        if (this.$refs && this.$refs.managementTraitsTemplateUpload) this.$refs.managementTraitsTemplateUpload.clearFiles()
        return
      }
      this.managementTraitsTemplateFile = file
    },
    async downloadManagementTraitsTemplate() {
      this.managementTraitsTemplateDownloading = true
      try {
        const blob = await downloadManagementTraitsTemplateFile()
        const blobUrl = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.style.display = 'none'
        link.href = blobUrl
        link.setAttribute('download', this.managementTraitsTemplateDownloadFileName())
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(blobUrl)
      } catch (error) {
        this.$message.error(error.message || '下载00501/00502共用报告模板失败')
      } finally {
        this.managementTraitsTemplateDownloading = false
      }
    },
    async uploadManagementTraitsTemplate() {
      if (!this.managementTraitsTemplateFile) return
      try {
        await this.$confirm('上传校验通过后将立即替换00501/00502共用报告模板；历史PDF保持不变。是否继续？', '确认上传', {
          confirmButtonText: '上传并生效',
          cancelButtonText: '取消',
          type: 'warning'
        })
      } catch (error) {
        if (error !== 'cancel' && error !== 'close') this.$message.error(error.message || '上传已取消')
        return
      }
      this.managementTraitsTemplateUploading = true
      try {
        await uploadManagementTraitsTemplateFile(this.managementTraitsTemplateFile)
        this.$message.success('00501/00502共用报告模板已上传并生效')
        this.managementTraitsTemplateFile = null
        if (this.$refs && this.$refs.managementTraitsTemplateUpload) this.$refs.managementTraitsTemplateUpload.clearFiles()
        await this.fetchManagementTraitsTemplate()
      } catch (error) {
        this.$message.error(error.message || '上传00501/00502共用报告模板失败')
      } finally {
        this.managementTraitsTemplateUploading = false
      }
    },
    fetchList() {
      this.loading = true
      request.get('/exam/api/mbti/templates').then(res => {
        this.templates = res.data || []
        this.loading = false
      }).catch(() => { this.loading = false })
    },
    handleDownload(row) {
      const url = process.env.VUE_APP_BASE_API + '/exam/api/mbti/templates/download/' + row.type
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', row.fileName)
      // 需要带 token
      fetch(url, { headers: { Authorization: 'Bearer ' + getToken() } })
        .then(res => res.blob())
        .then(blob => {
          const blobUrl = window.URL.createObjectURL(blob)
          link.href = blobUrl
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          window.URL.revokeObjectURL(blobUrl)
        })
    },
    handleDownloadSimple(row) {
      const url = process.env.VUE_APP_BASE_API + '/exam/api/mbti/templates/download/' + row.type + '?variant=simple'
      fetch(url, { headers: { Authorization: 'Bearer ' + getToken() } })
        .then(res => res.blob())
        .then(blob => {
          const blobUrl = window.URL.createObjectURL(blob)
          const link = document.createElement('a')
          link.style.display = 'none'
          link.href = blobUrl
          link.setAttribute('download', row.simpleFileName)
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          window.URL.revokeObjectURL(blobUrl)
        })
    },
    beforeUpload(file) {
      const isDocx = file.name.endsWith('.docx')
      const isLt10M = file.size / 1024 / 1024 < 10
      if (!isDocx) {
        this.$message.error('只支持 .docx 格式！')
        return false
      }
      if (!isLt10M) {
        this.$message.error('文件大小不能超过 10MB！')
        return false
      }
      return true
    },
    handleUploadSuccess(res) {
      if (res.code === 0 && res.data) {
        this.$message.success(`MBTI-${res.data.type} 模板上传成功`)
        this.fetchList()
      } else {
        this.$message.error(res.msg || '上传失败')
      }
    },
    handleUploadError() {
      this.$message.error('上传失败，请重试')
    }
  }
}
</script>

<style scoped>
.page-heading { margin-bottom: 12px; }
.page-heading h3 { margin: 0 0 4px; font-size: 18px; }
.page-heading p, .upload-hint { margin: 0; color: #909399; font-size: 12px; line-height: 1.5; }
.phase1-card, .management-traits-card { border-color: #dcdfe6; }
.management-traits-card { margin-top: 12px; }
.card-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.card-subtitle { margin-left: 8px; color: #909399; font-size: 12px; font-weight: normal; }
.template-meta .el-col { display: flex; gap: 8px; min-height: 30px; align-items: center; }
.meta-label { flex: none; color: #909399; font-size: 12px; }
.meta-value { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sha-row { display: flex; gap: 8px; align-items: center; margin-top: 4px; }
.sha-row code { overflow-wrap: anywhere; color: #606266; font-size: 12px; }
.contract-alert { margin-top: 10px; }
.phase1-actions { display: flex; align-items: center; gap: 8px; margin-top: 12px; }
.selected-file { flex: 1; min-width: 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #606266; }
.upload-hint { margin-top: 8px; }
.semantic-field-search { margin: 12px 0 8px; }
.semantic-field-dialog code { color: #303133; font-size: 12px; }
@media (max-width: 767px) {
  .card-header, .phase1-actions { align-items: flex-start; flex-wrap: wrap; }
  .selected-file { flex-basis: 100%; }
  .phase1-actions .el-button { min-height: 40px; }
  ::v-deep .semantic-field-dialog { width: calc(100% - 24px) !important; margin-top: 4vh !important; }
}
</style>

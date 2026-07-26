<template>
  <div class="app-container" style="padding: 8px 8px;">
    <div style="margin-bottom: 8px;">
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

export default {
  name: 'MbtiTemplates',
  data() {
    return {
      loading: false,
      templates: []
    }
  },
  computed: {
    uploadUrl() {
      return process.env.VUE_APP_BASE_API + '/exam/api/mbti/templates/upload'
    },
    uploadHeaders() {
      return { Authorization: 'Bearer ' + getToken() }
    }
  },
  created() {
    this.fetchList()
  },
  methods: {
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

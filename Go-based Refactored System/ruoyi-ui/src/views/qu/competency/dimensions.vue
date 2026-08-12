<template>
  <div class="app-container competency-dimension-maintenance">
    <div class="page-heading">
      <div>
        <h2>胜任力维度维护</h2>
        <p>维护第一期基层员工测评使用的10个A/B维度。</p>
      </div>
      <el-button icon="el-icon-back" @click="$router.push({ name: 'CompetencyQuestionList' })">返回题库</el-button>
    </div>

    <el-form :inline="true" :model="query" size="small" class="filter-bar">
      <el-form-item label="关键词">
        <el-input v-model="query.keyword" clearable placeholder="编号或名称" style="width:180px" @input="query.current=1" />
      </el-form-item>
      <el-form-item label="能力层级">
        <el-select v-model="query.virdLevel" clearable placeholder="全部" style="width:190px" @change="query.current=1">
          <el-option v-for="item in virdOptions" :key="item" :label="item" :value="item" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" clearable placeholder="全部" style="width:110px" @change="query.current=1">
          <el-option label="启用" :value="0" />
          <el-option label="停用" :value="1" />
        </el-select>
      </el-form-item>
      <el-form-item><el-button icon="el-icon-refresh" @click="resetQuery">重置</el-button></el-form-item>
    </el-form>

    <el-alert title="编号和稳定ID不可修改。当前仅维护已确认的第一期10个维度；第二、三期维度待总体矩阵定稿后另行扩展。" type="info" :closable="false" show-icon class="mb12" />

    <el-table v-loading="loading" :data="pagedRows" border stripe>
      <el-table-column label="顺序" prop="displayOrder" width="70" align="right" />
      <el-table-column label="编号" prop="code" width="80" align="center" />
      <el-table-column label="维度名称" prop="name" min-width="120" />
      <el-table-column label="能力层级" prop="virdLevel" min-width="120" />
      <el-table-column label="适用对象" prop="applicableCategory" width="100" align="center" />
      <el-table-column label="核心含义" prop="coreMeaning" min-width="220" show-overflow-tooltip />
      <el-table-column label="启用题数" prop="questionCount" width="90" align="right" />
      <el-table-column label="状态" width="80" align="center">
        <template slot-scope="scope">
          <el-tag :type="scope.row.status === 0 ? 'success' : 'info'" size="mini">{{ scope.row.status === 0 ? '启用' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80" align="center" fixed="right">
        <template slot-scope="scope"><el-button type="text" size="mini" icon="el-icon-edit" @click="openEdit(scope.row)">编辑</el-button></template>
      </el-table-column>
      <template slot="empty"><el-empty description="暂无胜任力维度" /></template>
    </el-table>
    <pagination v-show="filteredRows.length > 0" :total="filteredRows.length" :page.sync="query.current" :limit.sync="query.size" />

    <el-dialog title="编辑胜任力维度" :visible.sync="editVisible" width="680px" :fullscreen="$store.state.app.device === 'mobile'" append-to-body>
      <el-descriptions v-if="editVisible" :column="1" border size="small" class="identity-block">
        <el-descriptions-item label="维度编号">{{ editIdentity.code }}</el-descriptions-item>
      </el-descriptions>
      <el-form ref="editForm" :model="editForm" :rules="editRules" label-width="100px">
        <el-row :gutter="16">
          <el-col :xs="24" :sm="12"><el-form-item label="维度名称" prop="name"><el-input v-model="editForm.name" maxlength="100" show-word-limit /></el-form-item></el-col>
          <el-col :xs="24" :sm="12"><el-form-item label="显示顺序" prop="displayOrder"><el-input-number v-model="editForm.displayOrder" :min="1" :max="maxDisplayOrder" controls-position="right" /></el-form-item></el-col>
          <el-col :xs="24" :sm="12"><el-form-item label="能力层级" prop="virdLevel"><el-select v-model="editForm.virdLevel" style="width:100%"><el-option v-for="item in allVirdLevels" :key="item" :label="item" :value="item" /></el-select></el-form-item></el-col>
          <el-col :xs="24" :sm="12"><el-form-item label="适用对象" prop="applicableCategory"><el-select v-model="editForm.applicableCategory" style="width:100%"><el-option v-for="item in allApplicableCategories" :key="item" :label="item" :value="item" /></el-select></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="核心含义" prop="coreMeaning"><el-input v-model="editForm.coreMeaning" type="textarea" :rows="3" maxlength="500" show-word-limit /></el-form-item></el-col>
          <el-col :span="24"><el-form-item label="状态" prop="status"><el-radio-group v-model="editForm.status"><el-radio :label="0">启用</el-radio><el-radio :label="1">停用</el-radio></el-radio-group></el-form-item></el-col>
        </el-row>
      </el-form>
      <div slot="footer"><el-button @click="editVisible=false">取消</el-button><el-button type="primary" :loading="saving" @click="saveEdit">保存</el-button></div>
    </el-dialog>
  </div>
</template>

<script>
import { fetchCompetencyDimensions, updateCompetencyDimension } from '@/api/competency'

export default {
  name: 'CompetencyDimensionMaintenance',
  data() {
    return {
      loading: false,
      saving: false,
      rows: [],
      query: { keyword: '', virdLevel: '', status: '', current: 1, size: 20 },
      allVirdLevels: ['通用能力', '心理素养'],
      allApplicableCategories: ['基层员工'],
      maxDisplayOrder: 10,
      editVisible: false,
      editIdentity: {},
      originalStatus: 0,
      editForm: { id: '', name: '', virdLevel: '', applicableCategory: '', coreMeaning: '', displayOrder: 1, status: 0 },
      editRules: {
        name: [{ required: true, message: '维度名称不能为空', trigger: 'blur' }],
        virdLevel: [{ required: true, message: '请选择能力层级', trigger: 'change' }],
        applicableCategory: [{ required: true, message: '请选择适用对象', trigger: 'change' }],
        coreMeaning: [{ required: true, message: '核心含义不能为空', trigger: 'blur' }],
        displayOrder: [{ required: true, message: '显示顺序不能为空', trigger: 'change' }],
        status: [{ required: true, message: '请选择状态', trigger: 'change' }]
      }
    }
  },
  computed: {
    virdOptions() {
      return [...new Set(this.rows.map(item => item.virdLevel).filter(Boolean))]
    },
    filteredRows() {
      const keyword = String(this.query.keyword || '').trim().toLowerCase()
      return this.rows.filter(item => {
        if (this.query.virdLevel && item.virdLevel !== this.query.virdLevel) return false
        if (this.query.status !== '' && item.status !== this.query.status) return false
        return !keyword || String(item.code || '').toLowerCase().includes(keyword) || String(item.name || '').toLowerCase().includes(keyword)
      })
    },
    pagedRows() {
      const start = (this.query.current - 1) * this.query.size
      return this.filteredRows.slice(start, start + this.query.size)
    }
  },
  created() {
    this.loadDimensions()
  },
  methods: {
    loadDimensions() {
      this.loading = true
      return fetchCompetencyDimensions().then(response => { this.rows = response.data || [] }).finally(() => { this.loading = false })
    },
    resetQuery() {
      this.query = { keyword: '', virdLevel: '', status: '', current: 1, size: 20 }
    },
    openEdit(row) {
      this.editIdentity = { code: row.code }
      this.originalStatus = row.status
      this.editForm = {
        id: row.id, name: row.name, virdLevel: row.virdLevel,
        applicableCategory: row.applicableCategory, coreMeaning: row.coreMeaning,
        displayOrder: row.displayOrder, status: row.status
      }
      this.editVisible = true
      this.$nextTick && this.$nextTick(() => this.$refs.editForm && this.$refs.editForm.clearValidate())
    },
    saveEdit() {
      return new Promise(resolve => {
        this.$refs.editForm.validate(valid => {
          if (!valid) return resolve(false)
          const confirm = this.editForm.status !== this.originalStatus
            ? this.$confirm('状态变更只影响未来创建或发布的测评，已发布测评和历史结果不受影响。确认继续吗？', '确认状态变更', { type: 'warning' })
            : Promise.resolve()
          confirm.then(() => {
            this.saving = true
            updateCompetencyDimension(this.editForm).then(() => {
              this.editVisible = false
              this.$notify({ title: '成功', message: '胜任力维度保存成功', type: 'success' })
              this.loadDimensions()
              resolve(true)
            }).catch(() => resolve(false)).finally(() => { this.saving = false })
          }).catch(() => resolve(false))
        })
      })
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
</style>
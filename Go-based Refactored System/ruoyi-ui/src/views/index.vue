<template>
  <div class="app-container home">
    <!-- Header -->
    <div class="dashboard-header">
      <h2>人才综合素质评估系统</h2>
      <p class="sub-title">系统概览</p>
    </div>

    <!-- Stats Cards -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="12" :sm="6">
        <div class="stat-card" style="border-left: 4px solid #409EFF;" @click="$router.push('/exam/exam')">
          <div class="stat-icon" style="background: #ecf5ff; color: #409EFF;">
            <i class="el-icon-document"></i>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.examCount }}</div>
            <div class="stat-label">测评总数</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card" style="border-left: 4px solid #67C23A;" @click="$router.push('/my/my')">
          <div class="stat-icon" style="background: #f0f9eb; color: #67C23A;">
            <i class="el-icon-monitor"></i>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.onlineCount }}</div>
            <div class="stat-label">在线测评</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card" style="border-left: 4px solid #E6A23C;" @click="$router.push('/qu/qu/repo')">
          <div class="stat-icon" style="background: #fdf6ec; color: #E6A23C;">
            <i class="el-icon-folder-opened"></i>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.repoCount }}</div>
            <div class="stat-label">题库总数</div>
          </div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="6">
        <div class="stat-card" style="border-left: 4px solid #F56C6C;" @click="$router.push('/system/user')">
          <div class="stat-icon" style="background: #fef0f0; color: #F56C6C;">
            <i class="el-icon-user"></i>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.userCount }}</div>
            <div class="stat-label">系统用户</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- Quick Access -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :sm="24" :lg="12">
        <el-card shadow="hover">
          <div slot="header"><i class="el-icon-s-operation"></i> 快捷操作</div>
          <el-row :gutter="10">
            <el-col :span="8" style="margin-bottom: 10px;">
              <el-button type="primary" plain style="width:100%" @click="$router.push({name:'AddExam'})">
                <i class="el-icon-plus"></i> 新建测评
              </el-button>
            </el-col>
            <el-col :span="8" style="margin-bottom: 10px;">
              <el-button type="success" plain style="width:100%" @click="$router.push('/my/my')">
                <i class="el-icon-monitor"></i> 在线测评
              </el-button>
            </el-col>
            <el-col :span="8" style="margin-bottom: 10px;">
              <el-button type="warning" plain style="width:100%" @click="$router.push('/qu/qu/repo')">
                <i class="el-icon-folder"></i> 题库管理
              </el-button>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
      <el-col :sm="24" :lg="12">
        <el-card shadow="hover">
          <div slot="header"><i class="el-icon-data-line"></i> 最近测评</div>
          <el-table :data="recentExams" size="small" :show-header="false" style="width:100%; cursor:pointer" @row-click="goExamDetail">
            <template slot="empty">
              <div style="padding: 20px 0; color: #909399">
                <i class="el-icon-document" style="font-size: 24px; display: block; margin-bottom: 8px"></i>
                暂无测评数据
              </div>
            </template>
            <el-table-column prop="title" show-overflow-tooltip>
              <template slot-scope="scope">
                <span style="color:#409EFF">{{ scope.row.title }}</span>
              </template>
            </el-table-column>
            <el-table-column width="80" align="center">
              <template slot-scope="scope">
                <el-tag size="mini" :type="scope.row.state === 0 ? 'success' : scope.row.state === 2 ? 'warning' : scope.row.state === 3 ? 'danger' : 'info'">
                  {{ scope.row.state === 0 ? '进行中' : scope.row.state === 1 ? '已禁用' : scope.row.state === 2 ? '未开始' : scope.row.state === 3 ? '已结束' : '未知' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import request from '@/utils/request'

export default {
  name: "Index",
  data() {
    return {
      stats: {
        examCount: '-',
        onlineCount: '-',
        repoCount: '-',
        userCount: '-'
      },
      recentExams: []
    }
  },
  created() {
    this.loadStats()
  },
  methods: {
    loadStats() {
      // Fetch exam count
      request.post('/exam/api/exam/exam/paging', { current: 1, size: 1, params: {} }).then(res => {
        this.stats.examCount = res.data ? res.data.total : 0
      }).catch(() => {})
      // Fetch online exam count
      request.post('/exam/api/exam/exam/online-paging', { current: 1, size: 1 }).then(res => {
        this.stats.onlineCount = res.data ? res.data.total : 0
      }).catch(() => {})
      // Fetch repo count
      request.post('/exam/api/repo/paging', { current: 1, size: 1, params: {} }).then(res => {
        this.stats.repoCount = res.data ? res.data.total : 0
      }).catch(() => {})
      // Fetch user count
      request.get('/system/user/list', { params: { pageNum: 1, pageSize: 1 } }).then(res => {
        this.stats.userCount = res.total || 0
      }).catch(() => {})
      // Fetch recent exams
      request.post('/exam/api/exam/exam/paging', { current: 1, size: 5, params: {} }).then(res => {
        this.recentExams = (res.data && res.data.records) ? res.data.records : []
      }).catch(() => {})
    },
    goExamDetail(row) {
      if (row.assessmentType === 'competency') {
        this.$router.push({ name: 'CompetencyResults', params: { examId: row.id } })
        return
      }
      const isOpen = row.isOpen || 1
      this.$router.push({
        path: `/exam/exam/users/${row.id}/${isOpen}/${encodeURIComponent(row.title || '')}`
      })
    }
  }
};
</script>

<style scoped lang="scss">
.home {
  font-family: "open sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
  color: #303133;
}
.dashboard-header {
  margin-bottom: 20px;
  h2 {
    font-size: 24px;
    font-weight: 600;
    margin: 0 0 4px 0;
    color: #303133;
  }
  .sub-title {
    font-size: 14px;
    color: #909399;
    margin: 0;
  }
}
.stats-row {
  margin-bottom: 10px;
}
.stat-card {
  background: #fff;
  border-radius: 4px;
  padding: 20px;
  display: flex;
  align-items: center;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: box-shadow 0.3s;
  &:hover {
    box-shadow: 0 4px 16px 0 rgba(0, 0, 0, 0.12);
  }
}
.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin-right: 16px;
  flex-shrink: 0;
}
.stat-info {
  .stat-value {
    font-size: 28px;
    font-weight: 700;
    line-height: 1.2;
    color: #303133;
  }
  .stat-label {
    font-size: 13px;
    color: #909399;
    margin-top: 2px;
  }
}
</style>
</style>


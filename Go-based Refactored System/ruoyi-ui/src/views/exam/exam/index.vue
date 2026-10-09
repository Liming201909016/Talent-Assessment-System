<template>

  <data-table
    ref="pagingTable"
    :options="options"
    :list-query="listQuery"
  >
    <template slot="filter-content">

      <el-form :inline="true" size="small" label-width="80px" style="margin-bottom: -10px;">
        <el-form-item label="测评名称">
          <el-input v-model="listQuery.params.title" placeholder="请输入名称" clearable style="width: 180px;" />
        </el-form-item>
        <el-form-item label="测评类型">
          <el-select v-model="repoSelected" multiple clearable placeholder="全部" style="width: 200px;" @change="repoSelChange">
            <el-option v-for="item in repoOptions" :key="item.id" :label="item.title" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间">
          <el-date-picker v-model="listQuery.params.startTime" value-format="yyyy-MM-dd HH:mm" format="yyyy-MM-dd HH:mm" type="datetime" placeholder="选择开始时间" style="width: 200px;" />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-date-picker v-model="listQuery.params.endTime" value-format="yyyy-MM-dd HH:mm" format="yyyy-MM-dd HH:mm" type="datetime" placeholder="选择结束时间" style="width: 200px;" />
        </el-form-item>
      </el-form>

    </template>

    <template slot="data-columns">

      <el-table-column
        label="测评名称"
        prop="title"
        min-width="160"
        show-overflow-tooltip
      />
      <el-table-column
        label="测评类型"
        width="100"
        align="center"
      >
        <template slot-scope="scope">
          {{ managementTraitsVersionLabel(scope.row) }}
        </template>
      </el-table-column>

      <el-table-column
        label="测评时间"
        align="center"
        width="170"
      >
        <template slot-scope="scope">
          <template v-if="scope.row.timeLimit">
            <div>{{ parseTime(scope.row.startTime, '{y}-{m}-{d} {h}:{i}') }}</div>
            <div style="color:#999;">至 {{ parseTime(scope.row.endTime, '{y}-{m}-{d} {h}:{i}') }}</div>
          </template>
          <span v-else>不限时</span>
        </template>
      </el-table-column>

      <el-table-column
        label="考试总分"
        prop="totalScore"
        align="center"
        v-if="flag"
      />

      <el-table-column
        label="及格线"
        prop="qualifyScore"
        align="center"
        v-if="flag"
      />

      <el-table-column
        label="测评建立时间"
        align="center"
        width="155"
      >
        <template slot-scope="scope">
          {{ scope.row.createTime ? parseTime(scope.row.createTime, '{y}-{m}-{d} {h}:{i}') : '—' }}
        </template>
      </el-table-column>

      <el-table-column
        label="开放类型"
        align="center"
        width="90"
      >

        <template slot-scope="scope">
          {{ scope.row.isOpen | examOpenType(scope.row.isOpen) }}
        </template>

      </el-table-column>

      <el-table-column
        label="答题类型"
        align="center"
        width="90"
      >

        <template slot-scope="scope">
          {{ scope.row.answerType | examAnswerType(scope.row.answerType) }}
        </template>

      </el-table-column>

      <el-table-column
        label="状态"
        align="center"
        width="90"
      >

        <template slot-scope="scope">
          <el-tag v-if="scope.row.state === 0" type="success" size="mini">进行中</el-tag>
          <el-tag v-else-if="scope.row.state === 1" type="info" size="mini">已禁用</el-tag>
          <el-tag v-else-if="scope.row.state === 2" type="warning" size="mini">未开始</el-tag>
          <el-tag v-else-if="scope.row.state === 3" type="danger" size="mini">已结束</el-tag>
          <el-tag v-else type="info" size="mini">{{ scope.row.state }}</el-tag>
        </template>

      </el-table-column>

      <el-table-column
        label="操作"
        align="center"
        width="180px"
        fixed="right"
      >
        <template slot-scope="scope">
          <el-button type="warning" size="mini" :disabled="traitsEntryLoading" @click="handleExamDetail(scope.row)">详情</el-button>
          <el-dropdown size="mini" trigger="click" @command="cmd => handleCommand(cmd, scope.row)" style="margin-left:8px">
            <el-button size="mini" type="info">更多<i class="el-icon-arrow-down el-icon--right"></i></el-button>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item command="edit" icon="el-icon-edit">修改</el-dropdown-item>
              <el-dropdown-item v-if="scope.row.assessmentType === 'competency'" command="competencyResults" icon="el-icon-data-analysis">胜任力结果</el-dropdown-item>
              <el-dropdown-item v-else command="papers" icon="el-icon-document">测试记录</el-dropdown-item>
              <el-dropdown-item v-if="scope.row.assessmentType !== 'competency' && isManagementTraitsProduct(scope.row.repoCode) && canManageTraits" command="managementTraitsResults" icon="el-icon-view">TEST 结果（显式）</el-dropdown-item>
              <el-dropdown-item v-if="canManageTraits && isManagementTraitsProduct(scope.row.repoCode) && scope.row.state === 1" command="enableManagementTraits" icon="el-icon-video-play">启用 TEST 测评</el-dropdown-item>
              <el-dropdown-item v-if="canManageTraits && isManagementTraitsProduct(scope.row.repoCode) && scope.row.state === 0" command="disableManagementTraits" icon="el-icon-video-pause">禁用 TEST 测评</el-dropdown-item>
              <el-dropdown-item command="export" icon="el-icon-download">导出汇总</el-dropdown-item>
              <el-dropdown-item command="exportAnswers" icon="el-icon-document-copy">导出原始答题</el-dropdown-item>
              <el-dropdown-item v-if="scope.row.assessmentType !== 'competency'" command="stats" icon="el-icon-data-analysis">统计</el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </template>
      </el-table-column>

    </template>

  </data-table>

</template>

<script>
import DataTable from '@/components/DataTable'
import PieChart from "@/views/exam/exam/components/PieChart.vue";
import { fetchList } from '@/api/qu/repo'
import { canManageManagementTraits, fetchManagementTraitsProfile, managementTraitsExamKnown, setManagementTraitsExamState } from '@/api/managementTraits'
import { isManagementTraitsProduct } from '@/utils/managementTraitsProduct'

export default {
  name: 'ListExam',
  components: {PieChart, DataTable },
  data() {
    return {

      flag: false,
      traitsEntryLoading: false,
      listQuery: {
        current: 1,
        size: 20,
        params: {
          title: '',

        }
      },

      options: {
        // 可批量操作
        multi: true,
        // 批量操作列表
        multiActions: [
          {
            value: 'delete',
            label: '删除'
          },
        ],
        // 列表请求URL
        listUrl: '/exam/api/exam/exam/paging',
        // 删除请求URL
        deleteUrl: '/exam/api/exam/exam/delete',
        // 删除请求URL
        stateUrl: '/exam/exam/state',
        addRoute: 'AddExam'
      },

      sysUserId: undefined,
      repoOptions:[],
      repoSelected:[]
    }
  },

  created() {
    this.sysUserId = this.$store.state.user.id
    this.getRepoList()
  },

  mounted() {
    // this.wsUrl = "ws://127.0.0.1:8091/ws/" + this.sysUserId;
    // this.$store.dispatch("startWebSocket", {url: this.wsUrl, user: this.$options.name}).then(() => {
    //   // 添加socket通知监听
    //   window.addEventListener('onmessageWS', this.getSocketData)
    // })
  },

  computed: { canManageTraits() { return canManageManagementTraits(this.$store) } },
  methods: {
    isManagementTraitsProduct,
    managementTraitsVersionLabel(row) {
      if (row.repoCode === '00501') return '基层员工新版'
      if (row.repoCode === '00502') return '干部新版'
      if (isManagementTraitsProduct(row.repoCode)) return row.stuFlag == 1 ? '基层员工版' : '管理干部版'
      return row.stuFlag == 1 ? '学生版' : '职场版'
    },
    async routeManagementTraits(row) {
      const code = row.repoCode || (row.repoList && row.repoList[0] && row.repoList[0].repoCode) || ''
      if (row.assessmentType === 'competency' || !managementTraitsExamKnown(row.id) || !canManageManagementTraits(this.$store)) return false
      if (this.traitsEntryLoading) return true
      this.traitsEntryLoading = true
      try {
        const response = await fetchManagementTraitsProfile(row.id)
        if (!response || !Object.prototype.hasOwnProperty.call(response, 'data') || response.data === undefined) throw new Error('profile响应无效。')
        if (response.data === null) return false
        if (response.data.examId !== row.id || !response.data.frozenAt) throw new Error('profile身份或冻结状态无效。')
        await this.$router.push({ name: 'ManagementTraitsResults', params: { examId: row.id } })
        return true
      } catch (err) {
        this.$message.error(`002入口探测失败：${err.message || err}；已停止操作，请重试。`)
        return true
      } finally { this.traitsEntryLoading = false }
    },
    getRepoList() {
      fetchList({}).then(response => {
        this.repoOptions = response.data
      })
    },
    repoSelChange(){
      this.$set(this.listQuery.params,'repoIds',this.repoSelected)
      console.log(this.listQuery.params)
    },
    // 收到消息处理
    getSocketData (res) {
      // if (res.detail.data === 'success' || res.detail.data === 'heartBath') return
      // // ...业务处理
      // let msg = JSON.parse(res.detail.data)
      // if (msg.type === 1 && msg.roomId === this.roomId) {
      //   this.candidate = JSON.parse(msg.data)
      //   console.log(this.$options.name + ": " + this.sysUserName + "===receive message", this.candidate)
      // }
    },

    handleCommand(cmd, row) {
      if (cmd === 'edit') this.handleUpdateExam(row.id)
      else if (cmd === 'managementTraitsResults' && this.canManageTraits) this.$router.push({ name: 'ManagementTraitsResults', params: { examId: row.id } })
      else if (cmd === 'enableManagementTraits') return this.setManagementTraitsState(row, 0)
      else if (cmd === 'disableManagementTraits') return this.setManagementTraitsState(row, 1)
      else if (cmd === 'competencyResults') this.handleCompetencyResults(row)
      else if (cmd === 'papers') this.handlePaperList(row)
      else if (cmd === 'export') this.handleExportRawData(row)
      else if (cmd === 'exportAnswers') this.handleExportRawAnswers(row)
      else if (cmd === 'stats') this.handleStatistics(row)
    },

    async setManagementTraitsState(row, state) {
      if (!this.canManageTraits || !row || !isManagementTraitsProduct(row.repoCode) || ![0, 1].includes(state) || ![0, 1].includes(row.state) || row.state === state || this.traitsEntryLoading) return
      const action = state === 0 ? '启用' : '禁用'
      try {
        await this.$confirm(`确定${action}「${row.title || row.repoCode}」TEST 测评？`, `${action} TEST 测评`, {
          confirmButtonText: action,
          cancelButtonText: '取消',
          type: state === 0 ? 'warning' : 'info'
        })
        this.traitsEntryLoading = true
        const response = await setManagementTraitsExamState(row.id, state)
        if (!response || !response.data || response.data.examId !== row.id || response.data.state !== state) throw new Error('状态响应与请求不一致')
        this.$message.success(`${action}成功`)
        if (this.$refs.pagingTable && this.$refs.pagingTable.getList) this.$refs.pagingTable.getList()
      } catch (error) {
        if (error !== 'cancel' && error !== 'close') this.$message.error(`${action}失败：${error.message || error}`)
      } finally {
        this.traitsEntryLoading = false
      }
    },

    handleCompetencyResults(row) {
      this.$router.push({ name: 'CompetencyResults', params: { examId: row.id }})
    },

    async handlePaperList(row) {
      if (await this.routeManagementTraits(row)) return
      this.$router.push({ name: 'ListPaper', params: { examId: row.id }})
    },

    async handleStatistics(row) {
      if (await this.routeManagementTraits(row)) return
      this.$router.push({ name: 'StatisticsExam', params: { examId: row.id, title: row.title, state: row.state, isOpen: row.isOpen }})
    },

    async handleExamDetail(row) {
      console.log(row)
      if (row.assessmentType === 'competency') {
        this.$router.push({ name: 'CompetencyResults', params: { examId: row.id }})
        return
      }
      if (await this.routeManagementTraits(row)) return
      this.$router.push({ name: 'ListExamUser', params: { examId: row.id, isOpen: row.isOpen, title: row.title, stuFlag: row.stuFlag}})
    },

    handleUpdateExam(examId) {
      this.$router.push({ name: 'UpdateExam', params: { id: examId }})
    },

    async handleExportRawData(row) {
      if (await this.routeManagementTraits(row)) return
      const competency = row.assessmentType === 'competency'
      const message = competency ? '确定导出「' + row.title + '」的结果汇总、逐题明细和题目字典？' : '确定导出「' + row.title + '」的原始数据？'
      const fileName = competency ? row.title + '-胜任力结果明细.xlsx' : row.title + '-原始数据.xlsx'
      this.$confirm(message, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }).then(() => {
        this.download('/exam/api/exam/exam/export-raw-data?examId=' + row.id, {}, fileName)
      }).catch(() => {})
    },

    async handleExportRawAnswers(row) {
      if (await this.routeManagementTraits(row)) return
      const competency = row.assessmentType === 'competency'
      const message = competency ? '确定导出「' + row.title + '」的结果汇总、逐题明细和题目字典？' : '确定导出「' + row.title + '」全体考生的逐题答题原始记录？数据可能较大。'
      const fileName = competency ? row.title + '-胜任力结果明细.xlsx' : row.title + '-原始答题记录.xlsx'
      this.$confirm(message, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }).then(() => {
        this.download('/exam/api/exam/exam/export-raw-answers?examId=' + row.id, {}, fileName)
      }).catch(() => {})
    }
  },
  filters: {
      examStateFilter(value){
          if(value===0) return "进行中"
          else if(value===1) return "已禁用"
          else if(value===2) return "尚未开放"
          else if(value===3) return "已结束"
          return "未知(" + value + ")"
      },

      examOpenType(value) {
        if (value === 1) return "开放"
        else if (value === 2) return "封闭"
      },

      examAnswerType(value) {
        if (value === 1) return "滚动"
        else if (value === 2) return "点击"
      }
  }
}
</script>

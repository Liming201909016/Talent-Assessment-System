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
          {{ scope.row.stuFlag == 1 ? (scope.row.repoCode.startsWith("002") ? "基层员工版" : "学生版") : (scope.row.repoCode.startsWith("002") ? "管理干部版" : "职场版") }}
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
          <el-button type="warning" size="mini" @click="handleExamDetail(scope.row)">详情</el-button>
          <el-dropdown size="mini" trigger="click" @command="cmd => handleCommand(cmd, scope.row)" style="margin-left:8px">
            <el-button size="mini" type="info">更多<i class="el-icon-arrow-down el-icon--right"></i></el-button>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item command="edit" icon="el-icon-edit">修改</el-dropdown-item>
              <el-dropdown-item v-if="scope.row.assessmentType === 'competency'" command="competencyResults" icon="el-icon-data-analysis">胜任力结果</el-dropdown-item>
              <el-dropdown-item v-else command="papers" icon="el-icon-document">测试记录</el-dropdown-item>
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

export default {
  name: 'ListExam',
  components: {PieChart, DataTable },
  data() {
    return {

      flag: false,
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

  methods: {
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
      else if (cmd === 'competencyResults') this.handleCompetencyResults(row)
      else if (cmd === 'papers') this.handlePaperList(row)
      else if (cmd === 'export') this.handleExportRawData(row)
      else if (cmd === 'exportAnswers') this.handleExportRawAnswers(row)
      else if (cmd === 'stats') this.handleStatistics(row)
    },

    handleCompetencyResults(row) {
      this.$router.push({ name: 'CompetencyResults', params: { examId: row.id }})
    },

    handlePaperList(row) {
      this.$router.push({ name: 'ListPaper', params: { examId: row.id }})
    },

    handleStatistics(row) {
      this.$router.push({ name: 'StatisticsExam', params: { examId: row.id, title: row.title, state: row.state, isOpen: row.isOpen }})
    },

    handleExamDetail(row) {
      console.log(row)
      if (row.assessmentType === 'competency') {
        this.$router.push({ name: 'CompetencyResults', params: { examId: row.id }})
        return
      }
      this.$router.push({ name: 'ListExamUser', params: { examId: row.id, isOpen: row.isOpen, title: row.title, stuFlag: row.stuFlag}})
    },

    handleUpdateExam(examId) {
      this.$router.push({ name: 'UpdateExam', params: { id: examId }})
    },

    handleExportRawData(row) {
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

    handleExportRawAnswers(row) {
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

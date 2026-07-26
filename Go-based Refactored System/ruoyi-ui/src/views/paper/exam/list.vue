<template>

  <div>
    <el-dialog

      :show-close="false"
      :close-on-click-modal="false"
      :visible.sync="dialogVisible1"
      :destroy-on-close="true"
      width="30%"
      :before-close="handleClose">
      <div style="display: flex; justify-content: center; align-items: center;" class="qrcode" ref="qrCodeUrl"></div>
      <div style="text-align: center;margin-top: 10px;">
        <el-link type="primary" :underline="false" @click="toDoExam">{{QRCodeURL}}</el-link>
      </div>
      <div v-if="qrExamInfo" style="text-align: center; margin-top: 15px; padding: 10px; background: #f5f7fa; border-radius: 4px;">
        <div style="font-weight: bold; font-size: 15px; margin-bottom: 6px;">{{ qrExamInfo.title }}</div>
        <div v-if="qrExamInfo.timeLimit" style="color: #e6a23c; font-size: 13px;">
          测评时间：{{ parseTime(qrExamInfo.startTime, '{y}-{m}-{d} {h}:{i}') }} 至 {{ parseTime(qrExamInfo.endTime, '{y}-{m}-{d} {h}:{i}') }}
        </div>
        <div v-else style="color: #909399; font-size: 13px;">不限时</div>
        <div style="color: #909399; font-size: 12px; margin-top: 4px;">请考生在规定时间内扫码参加测评</div>
      </div>
      <span slot="footer" class="dialog-footer">
        <el-button :loading="loading" type="primary" @click="closeDialog" ref="autoClose">确认</el-button>
      </span>
    </el-dialog>

    <data-table
      ref="pagingTable"
      :options="options"
      :list-query="listQuery"
    >
      <template slot="filter-content">


        <el-input v-model="listQuery.params.title" placeholder="查询考试名称" style="width: 200px;" class="filter-item" />

      </template>

      <template slot="data-columns">

        <el-table-column
          label="考试名称"
          prop="title"
          show-overflow-tooltip
        />
        <el-table-column
          label="测评类型"
          prop="stuFlag"
          width="100"
        >
          <template slot-scope="scope">
            {{ scope.row.stuFlag == 1 ? (scope.row.repoCode.startsWith("002") ? "基层员工版" : "学生版") : (scope.row.repoCode.startsWith("002") ? "管理干部版" : "职场版") }}

          </template>
        </el-table-column>

        <el-table-column
          label="考试时间"
          width="220px"
          align="center"
        >

          <template slot-scope="scope">
          <span v-if="scope.row.timeLimit">
            {{ parseTime(scope.row.startTime, '{y}-{m}-{d} {h}:{i}') }} ~ {{ parseTime(scope.row.endTime, '{y}-{m}-{d} {h}:{i}') }}
          </span>
            <span v-else>不限时</span>
          </template>

        </el-table-column>

        <el-table-column
          label="考试时长"
          align="center"
          width="100"
        >

          <template slot-scope="scope">
            {{ scope.row.totalTime }}分钟
          </template>

        </el-table-column>

        <el-table-column
          label="测评开放类型"
          align="center"
          width="100"
        >

          <template slot-scope="scope">
            {{ scope.row.isOpen | examOpenType(scope.row.isOpen) }}
          </template>

        </el-table-column>

        <el-table-column
          label="测评答题类型"
          align="center"
          width="100"
        >

          <template slot-scope="scope">
            {{ scope.row.answerType | examAnswerType(scope.row.answerType) }}
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
          label="操作"
          align="center"
          width="150"
        >
          <template slot-scope="scope">
            <el-button v-if="scope.row.state===0" icon="el-icon-caret-right" type="primary" size="mini" @click="handlePre(scope.row)">去考试</el-button>
            <el-button v-if="scope.row.state===1" icon="el-icon-s-release" size="mini" disabled>已禁用</el-button>
            <el-button v-if="scope.row.state===2" icon="el-icon-s-fold" size="mini" disabled>待开始</el-button>
            <el-button v-if="scope.row.state===3" icon="el-icon-s-unfold" size="mini" disabled>已结束</el-button>
          </template>

        </el-table-column>

        <el-table-column
          label="考试二维码"
          align="center"
          width="120"
        >
          <template slot-scope="scope">
            <el-button type="primary" size="mini" @click="creatQrCode(scope.row)">二维码</el-button>
          </template>

        </el-table-column>

      </template>

    </data-table>
  </div>

</template>

<script>
import DataTable from '@/components/DataTable'
import QRCode from 'qrcodejs2'
import fa from "element-ui/src/locale/lang/fa";

export default {
  components: { DataTable },
  data() {
    return {
      QRCodeURL:'',
      qrExamInfo: null,
      flag: false,
      dialogVisible1: false,
      loading: false,
      // isGetData: false,

      listQuery: {
        current: 1,
        size: 20,
        params: {
        }
      },

      options: {
        // 可批量操作
        multi: false,
        // 列表请求URL
        listUrl: '/exam/api/exam/exam/online-paging'
      }
    }
  },

  methods: {

    // // 开始考试
    // handlePre(examId) {
    //   this.$router.push({ name: 'PreExam', params: { examId: examId }})
    // }
     creatQrCode(row) {
      const examId = row.id
      const isOpen = row.isOpen
      const stuFlag = row.stuFlag
      const repoCode = row.repoCode

      this.qrExamInfo = row
      this.dialogVisible1 = true
      let that = this;

      if (isOpen === 1) {
        // V-004 修复 (2026-05-11): 协议自适应，HTTPS 站点不再产生混合内容警告
        that.QRCodeURL = window.location.protocol + '//' + window.location.host + '/#/my/exam/candidate/' + examId + '/' + stuFlag + '/' + repoCode
        that.$nextTick(() => {
          let qrcode = new QRCode(this.$refs.qrCodeUrl, {
            // text: 'http://172.22.222.143:81/my/exam/candidate/' + examId, // 需要转换为二维码的内容
            // text: window.location.host+ window.location.pathname + '/my/exam/candidate/' + examId, // 需要转换为二维码的内容
            text: that.QRCodeURL, // 需要转换为二维码的内容
            // text: window.location.host + window.location.pathname + examId, // 需要转换为二维码的内容
            width: 150,
            height: 150,
            colorDark: '#000000',
            colorLight: '#ffffff',
            correctLevel: QRCode.CorrectLevel.H,
          })
        });
      } else if (isOpen === 2) {
        // V-004 修复 (2026-05-11): 协议自适应
        that.QRCodeURL = window.location.protocol + '//' + window.location.host + '/#/my/exam/tester/' + examId + '/' + repoCode
        that.$nextTick(() => {
          let qrcode = new QRCode(this.$refs.qrCodeUrl, {
            // text: 'http://172.22.222.143:81/my/exam/candidate/' + examId, // 需要转换为二维码的内容
            // text: window.location.host+ window.location.pathname + '/my/exam/candidate/' + examId, // 需要转换为二维码的内容
            text: that.QRCodeURL, // 需要转换为二维码的内容
            // text: window.location.host + window.location.pathname + examId, // 需要转换为二维码的内容
            width: 150,
            height: 150,
            colorDark: '#000000',
            colorLight: '#ffffff',
            correctLevel: QRCode.CorrectLevel.H,
          })
        });
      }
    },

    handlePre(row) {
      if (row.isOpen === 1) {
        this.$router.push({ name: 'candidateInfo', params: { examId: row.id ,stuFlag: row.stuFlag,repoCode: row.repoCode}})
      } else if (row.isOpen === 2) {
        this.$router.push({ name: 'tester', params: { examId: row.id ,stuFlag: row.stuFlag,repoCode: row.repoCode}})
      }
    },

    closeDialog() {
      this.dialogVisible1 = false
      this.QRCodeURL=''
    },

    handleClose() {
      // this.dialogVisible1 = false
    },
    toDoExam(){
      window.open(this.QRCodeURL)
    }
  },

  filters: {
    examStateFilter(value){
      if(value===0) return "进行中"
      else if(value===2) return "尚未开放"
      else if(value===3) return "已结束"
    },

    examOpenType(value) {
      if (value === 1) return "开放"
      else if (value === 2) return "封闭"
    },

    examAnswerType(value) {
      if (value === 1) return "滚动"
      else if (value === 2) return "点击"
    }
  },

  mounted() {

  },

}
</script>

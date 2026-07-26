<template>
  <div class="app-container" style="margin-left: auto; margin-right: auto; width: 100%;">

    <template v-if="!examBlocked">
    <el-card style="margin-top: 20px; ">
      <h3 align="center"  style="margin-bottom: 20px;">基本信息</h3>
      <el-form ref="testerFrom" :model="testerFrom" :rules="rules" label-position="left" label-width="120px"
      >

        <el-form-item label="手机号码" prop="idNumber">
          <el-input v-model="testerFrom.idNumber" :max-width="100" placeholder="请输入手机号码" />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input v-model="testerFrom.password" />
        </el-form-item>

      </el-form>
      <div style="margin-top: 20px; text-align: right;" >
        <el-button type="primary" @click="handleLogin" style="margin-left: auto; margin-right: auto">登录</el-button>
      </div>
    </el-card>
    </template>

  </div>
</template>

<script>

import {testerLogin} from "@/api/tester/tester";
import {fetchDetail} from "@/api/exam/exam";

export default {
  name: 'Tester',

  data() {
    return {
      testerFrom: {},
      repoCode: '',
      examBlocked: false,
      rules: {
        idNumber: [
          { required: true, message: '手机号码不能为空！' },
        ],

        password: [
          { required: true, message: '密码不能为空！' },
        ],
      }
    }
  },

  created() {

    this.examId = this.$route.params.examId
    this.repoCode = this.$route.params.repoCode

    // 考试状态检查 — 同时检查 state 和实际时间
    if (this.examId) {
      fetchDetail(this.examId).then(res => {
        if (res.data) {
          const state = res.data.state
          const now = new Date()
          const startTime = res.data.startTime ? new Date(res.data.startTime.replace(' ', 'T')) : null
          const endTime = res.data.endTime ? new Date(res.data.endTime.replace(' ', 'T')) : null
          const timeLimit = res.data.timeLimit

          if ((timeLimit && startTime && now < startTime) || state === 2) {
            this.examBlocked = true
            const msg = res.data.startTime
              ? `考试尚未开始，开放时间：${res.data.startTime}，请在规定时间内参加测评。`
              : '考试尚未开始，请在规定时间内参加测评。'
            this.$alert(msg, '提示', {
              confirmButtonText: '关闭', type: 'warning', showClose: false, closeOnClickModal: false
            })
          } else if ((timeLimit && endTime && now > endTime) || state === 3) {
            this.examBlocked = true
            this.$alert('考试已结束，无法继续参加测评。', '提示', {
              confirmButtonText: '关闭', type: 'error', showClose: false, closeOnClickModal: false
            })
          }
        }
      }).catch(() => {})
    }
  },

  // mounted() {
  //   const _this = this;
  //   this.bodyScale();
  //   window.onresize = function() {
  //     _this.bodyScale();
  //   }.bind(this);
  // },

  methods: {

    // bodyScale() {
    //   let devicewidth = document.documentElement.clientWidth //获取当前分辨率下的可是区域宽度
    //   let scale = devicewidth / 1000 // 分母——设计稿的尺寸
    //   document.body.style.zoom = scale //放大缩小相应倍数
    // },

    handleLogin() {
      this.$refs.testerFrom.validate((valid) => {
        if (!valid) {
          return
        }

        this.testerFrom.examId = this.examId
        console.log(this.testerFrom)
        this.submitForm()

      })
    },

    submitForm() {

      testerLogin(this.testerFrom).then(response => {
        this.testerFrom = response.data
        if (this.testerFrom.participantToken) {
          sessionStorage.setItem('competencyParticipantToken', this.testerFrom.participantToken)
          sessionStorage.setItem('competencyParticipantType', 'tester')
        }
        console.log(this.testerFrom)
        this.$notify({
          title: '成功',
          message: '登录成功！',
          type: 'success',
          duration: 2000
        })
        // console.log("====================================")
        this.$router.replace({ name: 'PreExam', params: { examId: this.examId, id: this.testerFrom.id,stuFlag: this.testerFrom.stuFlag,repoCode: this.repoCode}})
      })

      // this.$router.push({ name: 'PreExam', params: { id: this.examId }})
    },

  }
}
</script>

<style scoped>

</style>



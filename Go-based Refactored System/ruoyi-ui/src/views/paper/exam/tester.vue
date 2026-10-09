<template>
  <div class="app-container" style="margin-left: auto; margin-right: auto; width: 100%;">
    <el-alert v-if="managementTraitsMode" title="TEST · 管理特质，仅供测试，不可作为人才决策依据" type="warning" :closable="false" />
    <template v-if="!examBlocked">
    <el-card style="margin-top: 20px; ">
      <h3 align="center"  style="margin-bottom: 20px;">基本信息</h3>
      <el-form ref="testerFrom" :model="testerFrom" :rules="rules" label-position="left" label-width="120px"
      >

        <el-form-item label="手机号码" prop="idNumber">
          <el-input v-model="testerFrom.idNumber" :max-width="100" placeholder="请输入手机号码" />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input v-model="testerFrom.password" type="password" show-password />
        </el-form-item>

      </el-form>
      <div style="margin-top: 20px; text-align: right;" >
        <el-button type="primary" :loading="loggingIn" @click="handleLogin" style="margin-left: auto; margin-right: auto">登录</el-button>
      </div>
    </el-card>
    </template>

  </div>
</template>

<script>

import {testerLogin} from "@/api/tester/tester";
import {fetchDetail} from "@/api/exam/exam";
import { managementTraitsTokenClaims, rememberManagementTraitsParticipant, loginManagementTraitsTester } from '@/api/managementTraits'
import { classifyManagementTraitsExam, isManagementTraitsProduct } from '@/utils/managementTraitsProduct'

export default {
  name: 'Tester',

  data() {
    return {
      testerFrom: {},
      repoCode: '',
      examBlocked: false,
      examId: '',
      loggingIn: false,
      configLoading: false,
      managementTraitsDetected: false,
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

  computed: {
    managementTraitsMode() { return this.managementTraitsDetected }
  },

  created() {

    this.examId = this.$route.params.examId
    this.repoCode = this.$route.params.repoCode

    // 考试状态检查 — 同时检查 state 和实际时间
    if (this.examId) {
      this.configLoading = true
      fetchDetail(this.examId).then(res => {
        if (res.data) {
      if (isManagementTraitsProduct(this.repoCode) || isManagementTraitsProduct(res.data.repoCode)) {
        const classification = classifyManagementTraitsExam(res.data, this.examId)
        if (!['LEGACY002', 'NEW005', 'FROZEN_COMPAT002'].includes(classification)) {
          this.examBlocked = true
          this.$message.error('新版草稿尚未冻结或身份状态未确认，请联系管理员。')
          return
        }
        this.managementTraitsDetected = classification !== 'LEGACY002'
        if (this.managementTraitsDetected) return
      }
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
        } else this.examBlocked = true
      }).catch(() => { this.examBlocked = true }).finally(() => { this.configLoading = false })
    } else this.examBlocked = true
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
      if (this.loggingIn || this.examBlocked || this.configLoading) return
      this.$refs.testerFrom.validate((valid) => {
        if (!valid) {
          return
        }

        this.testerFrom.examId = this.examId
        this.submitForm()

      })
    },

    async submitForm() {
      if (this.loggingIn || this.examBlocked || this.configLoading) return
      this.loggingIn = true
      try {
        const response = this.managementTraitsMode
          ? await loginManagementTraitsTester(this.examId, this.testerFrom.idNumber, this.testerFrom.password)
          : await testerLogin(this.testerFrom)
        if (this.managementTraitsMode || (response.data && managementTraitsTokenClaims(response.data.participantToken))) {
          if (!response.data || !response.data.id) throw new Error('身份响应无效，请重新登录。')
          rememberManagementTraitsParticipant(response.data.participantToken, this.examId)
          this.$router.replace({ name: 'PreExam', params: { examId: this.examId, id: response.data.id, repoCode: this.repoCode }, query: { mngTest: '1' } })
          return
        }
        this.testerFrom = response.data
        if (this.testerFrom.participantToken) {
          sessionStorage.setItem('competencyParticipantToken', this.testerFrom.participantToken)
          sessionStorage.setItem('competencyParticipantType', 'tester')
        }
        this.$notify({
          title: '成功',
          message: '登录成功！',
          type: 'success',
          duration: 2000
        })
        // console.log("====================================")
        this.$router.replace({ name: 'PreExam', params: { examId: this.examId, id: this.testerFrom.id,stuFlag: this.testerFrom.stuFlag,repoCode: this.repoCode}})
      } catch (error) { this.$message.error(error.message || '登录失败，请重试。') }
      finally { this.loggingIn = false }
    },

  }
}
</script>

<style scoped>

</style>



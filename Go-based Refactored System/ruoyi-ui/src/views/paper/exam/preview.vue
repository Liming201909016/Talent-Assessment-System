<template>
  <div class="app-container"  style="width:100%; margin: 0px auto">
    <el-dialog

      :show-close="false"
      title="指导语"
      :close-on-click-modal="false"
      :visible.sync="dialogVisible1"
      width="50%"
      :before-close="handleClose">
      <span style="font-size: large">  本份问卷每道题目均由一对语句构成，请从每对语句中选择一个与您个人情况更相吻合或您更赞同的一种说法。每道题的不同选项之间并没有好坏、正误之分，只需您如实作答。如果您对题目中的两个语句都赞同或都不赞同，就请选择您相对而言更可接受的说法。</span><br><br>
<!--      <span style="font-size: large">我郑重承诺：</span><br><br>-->
<!--      <span style="font-size: large">一、保证在网上报名和网上确认（现场确认）时，严格按照报考条件及相关政策要求选择填报志愿，如实、准确提交报考信息和各项材料。如提供虚假、错误信息或弄虚作假，本人承担由此造成的一切后果。</span><br><br>-->
<!--      <span style="font-size: large">二、自觉服从考试组织管理部门的统一安排，接受监考人员的管理、监督和检查。</span><br><br>-->
<!--      <span style="font-size: large">三、自觉遵守相关法律和考试纪律、考场规则，诚信考试，不作弊。</span>-->
      <span slot="footer" class="dialog-footer">
        <el-button :loading="loading" type="primary" @click="closeDialog">确认</el-button>
      </span>
    </el-dialog>

    <el-row :gutter="24">

      <el-col :span="24" style="margin-bottom: 20px">

        <el-alert
          :title="startDisabledReason || '点击`开始测评`后将自动进入测评！'"
          :type="startDisabledReason ? 'warning' : 'error'"
          style="margin-bottom: 10px"
        />

        <el-card class="pre-exam">

          <div><strong>测评名称：</strong>{{ detailData.title }}</div>
          <div><strong>测评时长：</strong>{{detailData.totalTime}}分钟</div>
<!--          <div><strong>试卷总分：</strong>{{ detailData.totalScore }}分</div>-->
<!--          <div><strong>及格分数：</strong>{{ detailData.qualifyScore }}分</div>-->
<!--          <div><strong>测评描述：</strong>{{ detailData.content }}</div>-->
          <div>
            <strong>测评描述：</strong>
            <span v-html="displayDesc"></span>
          </div>

          <el-col :span="24" style="margin-top: 20px; margin-bottom: 20px; text-align: right;">
            <el-button :loading="loading" :disabled="!canStart" type="primary" icon="el-icon-caret-right" @click="handleCreate">
              开始测评
            </el-button>
            <el-button @click="handleBack">返回</el-button>
          </el-col>
        </el-card>

      </el-col>



    </el-row>
  </div>
</template>

<script>
import { fetchDetail } from '@/api/exam/exam'
import { createPaper } from '@/api/paper/exam'
import { paperDetail } from '@/api/paper/exam'
import {updateData} from "@/api/candidate/candidate";
import th from "element-ui/src/locale/lang/th";
import {getTesterByIdNumber, updateTester} from "@/api/tester/tester";
import { createCompetencyPaper } from '@/api/competency'

const desc1 = '本份问卷每道题目均由一对语句构成，请从每对语句中选择一个与您个人情况更相吻合或您更赞同的一种说法。每道题的不同选项之间并没有好坏、正误之分，只需您如实作答。<br><span style="color:red;">请注意，本次测验共计90道题目，每道题目都必须回答。如果您对题目中的两个语句都赞同或都不赞同，就请选择您相对而言更可接受的说法。</span>'
const desc2 = '在进行测验时，请仔细阅读每一个陈述，并根据自己的真实感受和经历，选择最符合您情况的选项。每道题的不同选项之间并没有好坏、正误之分，只需诚实地反映您的当前状态和感受。<br><span style="color:red;">请注意，本次测验共计140道题目，每道题目都必须回答，您必须且只能选择一个符合选项。</span>'
const desc3 = '本测验每道题目均设置A、B两个选项，请根据自身情况的符合程度选择对应圆圈，圆圈距离A或B越近，代表符合程度越高。每道题的不同选项之间并没有好坏、正误之分，只需诚实地反映您当前的状态和感受。<br><span style="color:red;">请注意，本次测验共计48道题目，每道题目都必须作答。</span>'
export default {
  data() {
    return {
      dialogVisible1: false,
      detailData: {},
      telephone: '',
      testerId: '',
      postForm: {
        examId: '',
        password: ''
      },
      testerInfo: {},
      examId: '',
      rules: {
        password: [
          { required: true, message: '测评密码不能为空！' }
        ]
      },

      loading: false,
      stuFlag: 0,
      repoCode: '',
      desc1 : desc1,
      desc2 : desc2,
      desc3 : desc3,
    }
  },

  computed: {
    canStart() {
      return !(this.detailData.assessmentType === 'competency' && Number(this.detailData.publishStatus) !== 1)
    },
    startDisabledReason() {
      return this.canStart ? '' : '该胜任力测评尚未发布，请联系管理员先执行“发布并冻结题目”。'
    },
    displayDesc() {
      if (this.detailData.assessmentType === 'competency') {
        const count = (this.detailData.competencyDimensions || []).reduce((total, item) => total + Number(item.questionCount || 0), 0)
        return `本测评采用五级量表，请根据自身真实情况选择最符合的选项。<br><span style="color:red;">请注意，本次测验共计${count}道题目，每道题目都必须作答。</span>`
      }
      const code = this.repoCode || ''
      let raw
      if (code.startsWith('003')) {
        raw = this.desc3
      } else if (code.startsWith('002')) {
        raw = this.desc2
      } else {
        raw = this.desc1
      }
      return (raw || '')
    }
  },

  created() {
    this.examId = this.$route.params.examId
    this.postForm.examId = this.examId
    this.testerId = this.$route.params.id
    this.stuFlag = this.$route.params.stuFlag || 0
    this.repoCode = this.$route.params.repoCode || ''
    this.getTesterInfo()
    this.fetchData()
  },

  methods: {

    handleClose() {
      // this.dialogVisible1 = false
    },

    getTesterInfo() {
      getTesterByIdNumber(this.testerId, this.examId).then(response => {
        this.testerInfo = response.data
      }).catch(() => {})
    },

    fetchData() {
      fetchDetail(this.postForm.examId).then(response => {
        this.detailData = response.data
        console.log(this.detailData)
        // 从 exam detail 获取 repoCode（路由参数可能为空）
        if (response.data && response.data.repoCode && !this.repoCode) {
          this.repoCode = response.data.repoCode
        }
        // 考试状态检查
        const state = this.detailData.state
        if (state === 2) {
          this.$alert('考试尚未开始，请在规定时间内参加测评。', '提示', {
            confirmButtonText: '返回',
            type: 'warning',
            callback: () => { this.$router.go(-1) }
          })
        } else if (state === 3) {
          this.$alert('考试已结束，无法继续参加测评。', '提示', {
            confirmButtonText: '返回',
            type: 'warning',
            callback: () => { this.$router.go(-1) }
          })
        }
      })
    },

    async handleCreate() {
      if (this.loading) return
      if (!this.canStart) {
        this.$message.warning(this.startDisabledReason)
        return
      }
      this.loading = true

      if (this.detailData.assessmentType === 'competency') {
        const participantToken = sessionStorage.getItem('competencyParticipantToken') || ''
        const participantType = sessionStorage.getItem('competencyParticipantType') || ''
        if (!participantToken || !participantType) {
          this.loading = false
          this.$message.error('参与者认证已失效，请返回重新登录或填写信息。')
          return
        }
        try {
          const response = await createCompetencyPaper({
            examId: this.examId,
            participantId: this.testerId,
            participantType,
            participantToken
          })
          sessionStorage.setItem('competencyPaperToken', response.data.paperToken)
          this.loading = false
          if (response.data.completed) {
            this.$router.replace({ name: 'ExamThankYou' })
          } else {
            this.$router.replace({ name: 'CompetencyExam', params: { paperId: response.data.paperId } })
          }
        } catch (error) {
          this.loading = false
        }
        return
      }

      // R08: 检查是否有已存在的试卷（断点续答 / 已完成拦截）
      if (this.testerInfo && this.testerInfo.paperId) {
        try {
          const paperRes = await paperDetail({ id: this.testerInfo.paperId })
          if (paperRes && paperRes.data) {
            const paperState = paperRes.data.state
            if (paperState === 2) {
              // 已完成答题
              this.loading = false
              this.$alert('已完成答题，不可重复答题。', '提示', {
                confirmButtonText: '返回',
                type: 'warning',
                callback: () => { this.$router.go(-1) }
              })
              return
            }
            if (paperState === 0) {
              // 进行中 → 断点续答
              this.loading = false
              this.toExam({ code: 0, data: { id: this.testerInfo.paperId } })
              return
            }
          }
        } catch (e) {
          console.log('paper check failed, creating new', e)
        }
      }

      await createPaper(this.postForm).then(res => {
        console.log(res)

        if (this.testerInfo && this.testerInfo.id) {
          let data = {id: this.testerInfo.id, examId: this.examId, paperId: res.data.id, idNumber: this.testerInfo.idNumber, pdfFlag: 0}
          data = JSON.stringify(data)
          updateTester(data).then(response => {
            console.log(response)
            this.toExam(res)
          })
        } else {
          let data = {examId: this.postForm.examId, id: this.testerId, paperId: res.data.id, pdfFlag: 0}
          data = JSON.stringify(data)
          updateData(data).then(response => {
            console.log(response)
            this.toExam(res)
          })
        }
      }).catch(() => {
        this.loading = false
      })
    },
    toExam(response){
      if (response.code === 0) {
        this.loading = false
        const repoCode = this.repoCode || ''
        // MBTI 走独立答题页
        if (repoCode.startsWith('003')) {
          this.$router.push({ name: 'MbtiExam', params: { id: response.data.id, testerId: this.testerId }})
        } else if (this.detailData.answerType === 1) {
          this.$router.push({ name: 'StartExam', params: { id: response.data.id, examId: this.examId, testerId: this.testerId,stuFlag: this.stuFlag,repoCode: this.repoCode}})
        } else if (this.detailData.answerType === 2) {
          this.$router.push({ name: 'StartExamClick', params: { id: response.data.id, examId: this.examId, testerId: this.testerId,stuFlag: this.stuFlag,repoCode: this.repoCode }})
        }
      }
      // setTimeout(function () {
      //   this.dialogVisible1=false;
      // },5000)
      this.dialogVisible1 = false
    },
    async closeDialog() {

      this.loading = true

      await createPaper(this.postForm).then(response => {
        console.log(response)

        if (this.testerInfo && this.testerInfo.id) {
          let data = {id: this.testerInfo.id, examId: this.examId, paperId: response.data.id, idNumber: this.testerInfo.idNumber}
          data = JSON.stringify(data)
          updateTester(data).then(response => {
            console.log(response)
          })
        } else {
          updateData(this.postForm.examId, this.testerId, response.data.id).then(response => {
            console.log(response)
          })
        }

        if (response.code === 0) {
          this.loading = false
          if (this.detailData.answerType === 1) {
            this.$router.push({ name: 'StartExam', params: { id: response.data.id, examId: this.examId, testerId: this.testerId, stuFlag: this.stuFlag, repoCode: this.repoCode }})
          } else if (this.detailData.answerType === 2) {
            this.$router.push({ name: 'StartExamClick', params: { id: response.data.id, examId: this.examId, testerId: this.testerId, stuFlag: this.stuFlag, repoCode: this.repoCode }})
          }

        }
        // setTimeout(function () {
        //   this.dialogVisible1=false;
        // },5000)
        this.dialogVisible1 = false
      }).catch(() => {
        this.loading = false
      })
    },

    handleBack() {
      this.$router.push({ name: 'candidateInfo', params: { examId: this.postForm.examId, testerId: this.testerId }})
    }

  }
}
</script>

<style scoped>

  .pre-exam div {

    line-height: 42px;
    color: #555555;
  }


</style>


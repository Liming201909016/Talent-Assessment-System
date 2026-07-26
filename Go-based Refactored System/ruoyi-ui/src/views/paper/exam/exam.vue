<template>

<!--  <div class="app-container"  style="display: flex;">-->
  <div class="app-container">
<!--    <div style="z-index: 999; position: fixed">
      <el-card style="margin-bottom: 10px;">

        距离测评结束还有：
        <exam-timer v-model="paperQuData.leftSeconds" @timeout="doHandler()" />

        <el-button style="float: right; margin-left: 15px;" type="primary" size="mini"
                   icon="el-icon-plus" :loading="loading" @click="handHandExam()">
          {{ handleText }}
        </el-button>

      </el-card>
    </div>-->
    <el-row style="margin-bottom: 30px;">
      <el-col>
        <el-form ref="form" :model="answers">
          <el-card :id="setQuId(index)" v-for="(item, index) in paperQuData.radioList" :key="index" class="qu-content">
            <p v-if="item.content"><span style="color: #00afff">{{ index + 1 }}.{{item.title?item.title:''}}</span></p>
            <div v-if="item.quType === 1 || item.quType===3">
              <el-radio-group :class="{'mng-radio-group': repoCode.startsWith('002')}" v-model="answers[index]" @change="handSelect(item, answers[index])" v-removeAriaHidden>
                <el-radio v-for="(c, j) in item.answerList" :key="c.id" :label="c.id" font-size="28px">{{ c.abc }}.&emsp; {{ c.content }}</el-radio>
              </el-radio-group>
            </div>
          </el-card>
        </el-form>
      </el-col>
    </el-row>
    <el-row style="position:fixed;bottom:0;width:100%;padding: 10px 0;text-align: center;">
      <el-button type="primary" size="mini"
                 icon="el-icon-plus" :loading="loading" @click="handHandExam()">
        {{ handleText }}
      </el-button>
    </el-row>

    <el-dialog

      :show-close="false"
      :close-on-click-modal="false"
      :visible.sync="dialogVisible1"
      width="50%"
      :before-close="handleClose">
      <span style="font-size: large">  答题已完成!</span><br><br>
      <span slot="footer" class="dialog-footer">
<!--        <el-button @click="candleDialog">否</el-button>-->
        <el-button type="primary" @click="closeDialog">确认</el-button>
      </span>
    </el-dialog>
  </div>

</template>

<script>
import {paperDetail, quDetail, handExam, fillAnswer, paperQuDetail, getShowPdf} from '@/api/paper/exam'
import { Loading } from 'element-ui'
import ExamTimer from '@/views/paper/exam/components/ExamTimer'
import item from "../../../layout/components/Sidebar/Item.vue";
import {testerInfo, updateEndTime} from "@/api/candidate/candidate";
import el from "element-ui/src/locale/lang/el";
import {editEndTime, getTesterByIdNumber} from "@/api/tester/tester";
import smoothscroll from 'smoothscroll-polyfill'

export default {
  name: 'ExamProcess',
  computed: {
    item() {
      return item
    }
  },
  components: { ExamTimer },
  data() {
    return {

      flag:false,
      dialogVisible1: false,

      answers: {},
      testerId: '',

      paperQuData: {
        leftSeconds: 99999,
        radioList: [],
        multiList: [],
        judgeList: [],
        repo:{}//题库信息
      },

      // 全屏/不全屏
      isFullscreen: false,
      showPrevious: false,
      showNext: true,
      loading: false,
      handleText: '交卷',
      pageLoading: false,
      // 试卷ID
      paperId: '',
      // 当前答题卡
      cardItem: {},
      allItem: [],

      // 单选选定值
      radioValue: [],
      // 多选选定值
      multiValue: [],
      // 已答ID
      answeredIds: [],
      // 是否生成报告
      idAndShowPdf: {
        showPdf: true,
      },

      randomUserId: undefined,
      tester: {},
      examId: '',
      stuFlag:0,
      repoCode:''
    }
  },
  created() {
    const id = this.$route.params.id
    this.examId = this.$route.params.examId
    this.testerId = this.$route.params.testerId
    this.stuFlag = this.$route.params.stuFlag
    this.repoCode = this.$route.params.repoCode
    if (typeof id !== 'undefined') {
      this.paperId = id
      this.fetchPaperQuData(id)
    }
  },

  mounted() {
    this.randomUserId = Math.floor(Math.random() * 2000) + 1;
    // let wsUrl = `${location.protocol === 'https' ? 'wss' : 'ws'}://${location.host}/ws/` + this.randomUserId;
    // let wsUrl = "ws://localhost:8091/ws/" + this.randomUserId;
    let wsUrl = "ws://" + window.location.hostname + ":8091/ws/" + this.randomUserId;
    this.$store.dispatch("startWebSocket", {url: wsUrl, user: this.$options.name})
  },

  methods: {

    sendMessage(message) {
      this.$store.dispatch('sendWebSocketMessage', {user:this.$options.name, msg:message});
    },

    formatDate(dt) {
      if (!dt) return ''
      return dt.substring(0, 10)
    },

    handleClose() {
      if (this.idAndShowPdf.isOpen === 1) {
        this.$router.push({ name: 'candidateInfo'})
        this.dialogVisible1 = false
      } else {
        this.$router.push({ name: 'tester'})
      }
    },

    setQuId(index) {
      return "question" + (index + 1)
    },

    fetchPaperQuData(id) {
      const params = { id: id }
      paperQuDetail(params).then(response => {
        // 试卷内容
        this.paperQuData = response.data

        if (this.paperQuData.radioList) {
          this.cardItem = this.paperQuData.radioList[0]
        }

        const that = this

        this.paperQuData.radioList.forEach(item => {
          item.answerList.forEach(answer => {
            if (answer.checked === true) {
              let index = parseInt(item.content.substring(1)) - 1
              this.answers[index] = answer.id
            }
          })
        })
        console.log(this.answers)
      })

      this.fetchTester();

      getShowPdf(params).then(response => {
        this.idAndShowPdf = response.data
        console.log("showPdf:", this.idAndShowPdf)
      })
    },

    handSelect(item, value, callback) {

      const answers = this.multiValue
      if (value !== '') {
        answers.push(value)
      }
      const params = { paperId: this.paperId, quId: item.quId, answers: answers, answer: '' }
      fillAnswer(params).then(() => {
        // 必须选择一个值
        if (value !== '') {
          // 加入已答列表
          this.cardItem.answered = true
        }

        // 最后一个动作，交卷
        if (callback) {
          callback()
        }

        this.cardItem = item

        // this.fetchPaperQuData(this.paperId)

        console.log(this.answers)
        // console.log(this.answers[0])
      })

      this.$forceUpdate()
    },

    /**
     * 统计有多少题没答的
     * @returns {*[]}
     */
    countNotAnswered() {
      let notAnswered = []

      // this.paperQuData.radioList.forEach(function(item) {
      //   if (!item.answered) {
      //     notAnswered += 1
      //   }
      // })

      for (let i = 0; i < this.paperQuData.radioList.length; i++) {
        if (typeof this.answers[i] !== "string") {
          notAnswered.push(i + 1)
        }
      }

      return notAnswered
    },


    doHandler() {
      this.handleText = '正在交卷，请等待...'
      this.loading = true

      const cParams = {paperId: this.paperId}
      if (this.testerId.length !== 18) {
        updateEndTime(cParams)
      } else {
        let data = {idNumber: this.testerId, examId: this.examId}
        editEndTime(data)
      }

      const params = { id: this.paperId }
      handExam(params).then(() => {
        this.$message({
          message: '试卷提交成功！',
          type: 'success'
        })

        this.dialogVisible1 = true
        // this.$router.push({ name: 'ShowExam', params: { id: this.paperId, testerId: this.testerId }})

        // if (this.idAndShowPdf.showPdf) {
        //   this.$router.push({ name: 'ShowExam', params: { id: this.paperId, testerId: this.testerId }})
        // } else {
        //   this.dialogVisible1 = true
        // }
      })
    },

    candleDialog() {
      this.$router.push({name: 'Finish'})
    },

    closeDialog() {

      if (this.idAndShowPdf.showPdf) {
        let routName = this.repoCode.startsWith('002') ? 'ShowMngExam' : 'ShowExam'
        this.$router.push({ name: routName, params: { id: this.paperId, examId: this.examId, testerId: this.testerId,stuFlag: this.stuFlag ,repoCode:this.paperQuData.repo.code}})
      } else {
        // if (this.idAndShowPdf.isOpen === 1) {
        //   this.$router.push({ name: 'candidateInfo'})
        //   this.dialogVisible1 = false
        // } else {
        //   this.$router.push({ name: 'tester'})
        // }
        this.$router.push({name: 'Finish'})
      }
      // this.$router.push({ name: 'ShowExam', params: { id: this.paperId, testerId: this.testerId }})
      // if (this.idAndShowPdf.isOpen === 1) {
      //   this.$router.push({ name: 'candidateInfo'})
      //   this.dialogVisible1 = false
      // } else {
      //   this.$router.push({ name: 'tester'})
      // }
    },

    // 交卷操作
    handHandExam() {
      const that = this

      const notAnswered = that.countNotAnswered()
      let questionId = "question" + notAnswered[0];

      console.log(notAnswered)
      let quStr = ''
      for (let i = 0; i < notAnswered.length; i++) {
        if (notAnswered[i] !== this.paperQuData.radioList.length) {
          quStr = quStr + notAnswered[i] + '、'
        } else {
          quStr = quStr + notAnswered[i]
        }
      }

      if (notAnswered.length > 0) {
        console.log("tester===========", this.tester)
        let data = {tester: this.tester.name, notAnswered: notAnswered.length}
        let send_msg = {type: 1, roomId: 1, data: JSON.stringify(data)}
        // this.sendMessage(send_msg)
        that.$confirm('第' + quStr + '题未作答，请全部作答完毕后再交卷！', '提示', {
          confirmButtonText: '定位到未作答',
          type: 'warning'
        }).then(() => {
          smoothscroll.polyfill()
          document.getElementById(questionId).scrollIntoView();
        }).catch((e) => {
        })
      }else{
        that.$confirm('确认要交卷吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          that.doHandler()
        }).catch((e) => {
          that.$message({
            type: 'info',
            message: '交卷已取消，您可以继续作答！'
          })
        })
      }


    },

    fetchTester() {
      // console.log(this.testerId)
      const params = { paperId: this.paperId }
      if (this.testerId.length === 18) {
        getTesterByIdNumber(this.testerId, this.examId).then(response => {

          if (response.data !== null) {
            this.tester = response.data
            if (this.tester.gender === "0") {
              this.tester.gender = "男"
            }

            if (this.tester.gender === "1") {
              this.tester.gender = "女"
            }

          }

          if (this.tester.createTime !== undefined && this.tester.createTime != null) {
            let tem = this.formatDate(this.tester.createTime).split('-')
            this.coverTime = tem[0] + ' 年 ' + tem[1] + ' 月 ' + tem[2] + ' 日 '
          }
        })
      } else {
        testerInfo(params).then(response => {

          this.tester = response.data
          if (this.tester.gender === "0") {
            this.tester.gender = "男"
          }

          if (this.tester.gender === "1") {
            this.tester.gender = "女"
          }

          if (this.tester.createTime !== undefined && this.tester.createTime != null) {
            let tem = this.formatDate(this.tester.createTime).split('-')
            this.coverTime = tem[0] + ' 年 ' + tem[1] + ' 月 ' + tem[2] + ' 日 '
          }
        })
      }

      console.log("this.tester>>>>", this.tester)
    },

    // 试卷详情
    fetchQuData(item) {
      // 打开
      const loading = Loading.service({
        text: '拼命加载中',
        background: 'rgba(0, 0, 0, 0.7)'
      })

      // 获得详情
      this.cardItem = item

      // 查找下个详情
      const params = { paperId: this.paperId, quId: item.quId }
      quDetail(params).then(response => {
        console.log(response)
        this.quData = response.data
        this.radioValue = ''
        this.multiValue = []

        // 填充该题目的答案
        this.quData.answerList.forEach((item) => {
          if ((this.quData.quType === 1 || this.quData.quType === 3) && item.checked) {
            this.radioValue = item.id
          }

          if (this.quData.quType === 2 && item.checked) {
            this.multiValue.push(item.id)
          }
        })

        // 关闭详情
        loading.close()
      })
    },

    // 试卷详情
    fetchData(id) {
      const params = { id: id }
      paperDetail(params).then(response => {
        // 试卷内容
        this.paperData = response.data

        // 获得第一题内容
        if (this.paperData.radioList) {
          this.cardItem = this.paperData.radioList[0]
        } else if (this.paperData.multiList) {
          this.cardItem = this.paperData.multiList[0]
        } else if (this.paperData.judgeList) {
          this.cardItem = this.paperData.judgeList[0]
        }

        const that = this

        this.paperData.radioList.forEach(function(item) {
          that.allItem.push(item)
        })

        this.paperData.multiList.forEach(function(item) {
          that.allItem.push(item)
        })

        this.paperData.judgeList.forEach(function(item) {
          that.allItem.push(item)
        })

        // 当前选定
        this.fetchQuData(this.cardItem)
      })
    }

  }
}
</script>

<style scoped>

  .qu-content div{
    line-height: 30px;
  }

  .el-checkbox-group label,.el-radio-group label{
    width: 100%;
  }
  .mng-radio-group label{
    width: 120px !important;
  }

  .card-title{
    background: #eee;
    line-height: 35px;
    text-align: center;
    font-size: 14px;
  }
  .card-line{
    padding-left: 10px
  }
  .card-line span {
    cursor: pointer;
    margin: 2px;
  }

  ::v-deep
  .el-radio, .el-checkbox{
    //padding: 9px 20px 9px 10px;
    border-radius: 4px;
    border: 1px solid #dcdfe6;
    margin-bottom: 10px;
  }

  .is-checked{
    border: #409eff 1px solid;
  }

  .el-radio img, .el-checkbox img{
    max-width: 200px;
    max-height: 200px;
    border: #dcdfe6 1px dotted;
  }

  ::v-deep
  .el-checkbox__inner {
    //display: none;
  }

  ::v-deep
  .el-radio__inner{
    display: none;
  }

  ::v-deep
  .el-checkbox__label{
    //line-height: 30px;
  }

  ::v-deep
  .el-radio__label{
    line-height: 40px;
    font-size: 12px;
  }

</style>

<!-- R12: 全局修复交卷确认弹窗在移动端的显示问题 -->
<style>
  .el-message-box {
    max-width: 90vw;
    width: auto !important;
    min-width: 280px;
  }
  .el-message-box__btns {
    text-align: center;
    padding: 10px 15px;
  }
  .el-message-box__btns .el-button {
    min-width: 60px;
  }
</style>


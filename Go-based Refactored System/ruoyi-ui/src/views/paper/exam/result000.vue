<template>
  <div class="invitationLetter" style="width: 100%">
<!--    <el-button type="primary" @click="createPdf">导出</el-button>-->
    <main id="meet-pdf-id" :class="{ pdf__output: isPdf }">
      <!-- 页头页眉 -->
      <div id="pdf-header" style="height: 2.2cm" class="pdf-hf"></div>
      <div id="first-page-header" style="height: 0.001cm"></div>

      <!-- 导出区域 -->
      <div class="main one-page">
        <div>
          <div class="report-cover" >
<!--            <div style="position: relative; bottom: 180px; left: 150px; font-size: 15px">-->
<!--              <p>测评编号：{{tester.paperId}}</p>-->
<!--            </div>-->
                      <div class="mt-800 pdf-group-item" style="text-align: center">
                        <div class="flex-column-center">
                          <p class="main__time" style="text-align: center">{{ coverTime }}</p>
                        </div>
                      </div>
          </div>
<!--          <h3 class="main__title">心理特质测评报告</h3>-->
<!--          <h3 class="main__title">Mental Health Assessment Report</h3>-->
<!--          <div class="line-img"><img :src="cover03"></div>-->
<!--          <h3 class="main__title">{{tester.name}}</h3>-->
<!--          <div class="mt-100 pdf-group-item" style="text-align: center">-->
<!--            <div class="flex-column-center">-->
<!--              <p class="main__text" style="text-align: center">{{ coverTime }}</p>-->
<!--            </div>-->
<!--          </div>-->
        </div>
      </div>

      <div class="page-split mt-40" v-show="!isPdf"></div>

      <div class="main page pdf-split-page">
<!--
        <h3 class="main__title mt-40 w-600 pdf-group-item">报告阅读说明</h3>
-->
        <h3 class="main__title mt-40 w-600 pdf-group-item" style="margin: 25px 0 15px;">
          <!--报告阅读说明-->
          <img src="@/assets/report_images/desc.png" width="230">
        </h3>
        <p class="main__text mt-40 text__indent pdf-group-item">
          在社会持续发展与竞争日益激烈的当下，人们面临的心理压力和挑战不断升级，因此，心理健康问题已逐渐上升为职场人士最为关注的焦点。
          心理健康不仅意味着个体不存在心理疾病，更是一种内心的平衡与和谐状态，是成就个人价值和实现幸福生活的根本所在。经调研发现，
          越是在大城市生活、拥有高学历和高收入的人群，即所谓的“高端人群”，越可能面临心理问题的挑战。
        </p>
        <p class="main__text text__indent pdf-group-item">
          本测验旨在全面揭示受测者的心理健康状况，并深入探究其在焦虑、抑郁、心理失衡、敌意、恐惧、身体不适、认知衰退、情绪化、挫折感、
          自我否定、怀疑感以及职业倦怠等十二个关键心理维度上的具体表现。这些心理状态，若未能得到妥善管理和干预，可能转化为心理风险因素，
          进而增加受测者在工作中出现偏差的可能性，从而对个人的工作绩效以及企业的稳定发展产生直接的负面影响。因此，从长远利益考虑，
          企业了解员工的心理状态是非常必要的。
        </p>
        <div class="line-img"><img :src="noteImg"></div>
        <p class="main__text text__indent pdf-group-item">
          特此说明：本报告绝无冒犯或诋毁之意，所提供的测试结果不能单独作为判断个体心理健康状况的唯一标准，
          还需综合考虑受测者的日常表现、其他相关测验结果以及专业医生的诊断意见。因此，本报告仅为辅助参考之用。
        </p>
      </div>

      <!-- 分页 -->
      <div class="page-split mt-40" v-show="!isPdf"></div>
      <div class="main page pdf-split-page">
        <div>
          <h3 class="main__title mt-40 w-600 pdf-group-item">个体报告详细解读</h3>
          <div>
            <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
              <span class="w-600">一、个人信息</span>
            </p>
            <div style="float: left; width: 50%; height: 240px">
              <ul style="margin-left: 60px">
                <li><p class="main__text lh-40 pdf-group-item w-600">姓名：{{tester.name}}</p></li>
                <li><p class="main__text lh-40 pdf-group-item w-600">年龄：{{tester.age}}</p></li>
                <li><p class="main__text lh-40 pdf-group-item w-600">单位：{{tester.affiliation}}</p></li>
                <li><p class="main__text lh-40 pdf-group-item w-600">手机：{{tester.telephone}}</p></li>
                <li><p class="main__text lh-40 pdf-group-item w-600">时间：{{formatDate(tester.createTime)}}</p></li>
              </ul>
            </div>
            <div style="float: right; width: 50%; height: 240px">
              <ul>
                <li><p class="main__text lh-40 pdf-group-item w-600">性别：{{tester.gender}}</p></li>
                <li><p class="main__text lh-40 pdf-group-item w-600">学历：</p></li>
                <li><p class="main__text lh-40 pdf-group-item w-600">岗位：{{tester.post}}</p></li>
                <li><p class="main__text lh-40 pdf-group-item w-600">时长：{{tester.userTime}} 分钟</p></li>
              </ul>
            </div>
            <p class="main__text text__indent pdf-group-item" v-if="tester.userTime > 35">
              在测试过程中，您的答题时间严重超出预设标准时间，因此，所得测试结果可能存在一定的误差。
            </p>
          </div>

          <div>
            <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
              <span class="w-600">二、总体评价</span>
            </p>
            <p class="main__text text__indent pdf-group-item">
              您的心理健康水平：
            </p>
            <TotalChart :count8="count8" :totalAvg="totalAvg"></TotalChart>
            <div class="main__text lh-40">
              <p class="main__text text__indent pdf-group-item">
                经过对本次心理测评结果的深入分析，得出以下结论：目前您的心理状况较好，但在某些极端或特定情景下，
                可能会出现轻微的心理困扰，不过并无大碍，无需过分担忧。（注意：本报告结果参考价值受限于您作答的结果参考性情况。）
              </p>
              <p class="main__text text__indent pdf-group-item">
                与职业群体进行对比分析，您目前所经历的{{lowWord}}、{{midWord}}，均属于普遍存在的心理和情感反应。
                当面临工作中的挑战和压力时，这些心理状态并未显示出异常或困扰的迹象。相比之下，您在{{heightWord}}等心理状态上表现更为突出。这些心理状态若不加以关注和处理，可能会对您的心理健康、生活品质和工作效率造成显著的负面影响。
              </p>
            </div>
            <RadarChart :scores="scores" v-if="isGetData"></RadarChart>
            <div class="main__text lh-40">
              <p class="main__text text__indent pdf-group-item">图示说明：</p>
              <p class="main__text text__indent pdf-group-item">1、常模标准分的参照分数为5.5；</p>
              <p class="main__text text__indent pdf-group-item">
                2、个人各项得分均以标准十分制呈现，是在参考群体常模的基础上，将受测者在12个因素上原始得分按照转化规则转化后得到的分数。</p>
            </div>
          </div>

          <div>
            <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
              <span class="w-600">三、综合评价报告</span>
            </p>
            <p class="main__text text__indent pdf-group-item">
              本报告以您的个人测试得分为依据，从心理状况的高、中、低分段因素出发，进行了详尽的分析与评估，力求全面、客观地揭示您的心理状态。
              具体来说，低分段得分表明您的心理健康状况良好；而得分逐渐升高，进入高分段，则反映出您面临的心理困扰愈发严峻。
            </p>
            <BarChart :scores="scores" v-if="isGetData"></BarChart>
            <div class="main__text lh-40">
              <p class="main__text text__indent pdf-group-item">
                图示说明：
              </p>
              <p class="main__text text__indent pdf-group-item">
                根据常模标准，每个因素若转化后的分数高于7分时，归为高分段因素，以红色部分呈现，表示心理困扰程度相对比较重。相比之下，
                分数位于4.5分至7分之间，归为中分段因素，以橙色部分呈现，表示心理困扰程度比较轻。另外，分数低于4.5分时，归为低分段因素，
                以绿色部分呈现，表示心理状态良好。
              </p>
            </div>
          </div>

          <div>
            <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
              <span class="w-600">四、分层评价报告</span>
            </p>
            <div>
              <el-row  class="whole-node " type="flex" align="middle" justify="center" style="margin-top: 20px"><img class="pdf-group-item" :src="heightImg"></el-row>
              <p class="main__text text__indent pdf-group-item">
                根据您的作答结果显示，您在{{heightWord}}因素的得分较高，反映出您的这些心理状态不够理想。
                考虑到这些因素可能对您的心理健康产生不良影响，您应当给予高度重视，并采取积极有效的措施进行改善。
              </p>
<!--              高分段因素-->
              <div v-for="data in heightScores" >
                <div class="pdf-group-item">
                  <ScoreChart :traits="data.name" :score="data.value"></ScoreChart>
<!--                  <ScoreChart02 :traits="data.name" :score="data.value"></ScoreChart02>
                  <el-image :src="scoreImg"></el-image>-->
                </div>
                <p class="main__text text__indent pdf-group-item">【个人得分】{{data.value}}分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】{{ data.evaluation }}</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】{{data.define}}</p>
              </div>
            </div>

            <div>
              <el-row class="whole-node " type="flex" align="middle" justify="center" style="margin-top: 20px"><el-image class="pdf-group-item" :src="midImg"></el-image></el-row>
<!--              <img :src="midImg">-->
              <p class="main__text text__indent pdf-group-item">
                根据您的作答结果显示，您在{{midWord}}因素的得分适中，反映出您的这些心理状态处于尚可水平。
                考虑到这些因素对您心理健康的潜在威胁，您应当主动关注自己的心理状态，并及时寻求专业的心理疏导。
              </p>
<!--              中分段因素-->
              <div v-for="data in midScores">
                <div class="pdf-group-item">
                  <ScoreChart :traits="data.name" :score="data.value"></ScoreChart>
<!--                  <ScoreChart02 :traits="data.name" :score="data.value"></ScoreChart02>
                  <el-image :src="scoreImg"></el-image>-->
                </div>
<!--                <ScoreChart :traits="data.name" :score="data.value"></ScoreChart>-->
                <p class="main__text text__indent pdf-group-item">【个人得分】{{data.value}}分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】{{ data.evaluation }}</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】{{data.define}}</p>
              </div>
            </div>

            <div >
              <el-row class="whole-node " type="flex" align="middle" justify="center" style="margin-top: 20px"><el-image class="pdf-group-item" :src="lowImg"></el-image></el-row>
<!--              <img :src="lowImg" />-->
              <p class="main__text text__indent pdf-group-item">
                根据您的作答结果显示，您在{{lowWord}}因素的得分较低，
                反映出您在这些方面拥有较为健康的心理状态。您在保持现有良好心理状态的基础上，持续加强心理健康的维护与提升工作，以保持一个健康、稳定的心理状态。
              </p>
<!--              低分段因素-->
              <div v-for="data in lowScores" >
                <div class="pdf-group-item">
                  <ScoreChart :traits="data.name" :score="data.value"></ScoreChart>
<!--                  <ScoreChart02 :traits="data.name" :score="data.value"></ScoreChart02>
                  <el-image :src="scoreImg"></el-image>-->
                </div>
<!--                <ScoreChart :traits="data.name" :score="data.value"></ScoreChart>-->
                <p class="main__text text__indent pdf-group-item">【个人得分】{{data.value}}分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】{{ data.evaluation }}</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】{{data.define}}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div id="pdf-footer" style="height: 2cm" class="pdf-hf">
        <div style="display: flex; justify-content: center; align-items: center; margin-top: 20px">
          第
          <div class="pdf-footer-page"></div>
          页 / 共
          <div class="pdf-footer-page-count"></div>
          页
        </div>
      </div>
      <div id="first-page-footer" style="height: 0.001cm"></div>
    </main>
  </div>
</template>

<script>
import { PdfLoader } from "@/utils/pdfLoader";
import cover02 from '@/assets/report_images/cover02.png'
import cover03 from '@/assets/report_images/cover03.png'
import CircleChart from "@/views/paper/exam/components/CircleChart.vue";
import TotalChart from "@/views/paper/exam/components/TotalChart.vue";
import RadarChart from "@/views/paper/exam/components/RadarChart.vue";
import BarChart from "@/views/paper/exam/components/BarChart.vue";
import heightScore from '@/assets/report_images/heightScore002.png';
import midScore from '@/assets/report_images/midScore002.png';
import lowScore from '@/assets/report_images/lowScore002.png';
import ScoreChart from "@/views/paper/exam/components/ScoreChart.vue";
// import ScoreChart02 from "@/views/paper/exam/components/ScoreChart02.vue";
import noteImg from '@/assets/report_images/note002.png';
import {testerInfo, paperResult, pdfPersistence} from "@/api/candidate/candidate";
import my_copy from "../../user/exam/my_copy.vue";
import {getTesterByIdNumber, paperResult2, pdfPersistence2} from "@/api/tester/tester";
import scoreImg from '@/assets/report_images/score002.png';

export default {
  name: "result",
  computed: {
    my_copy() {
      return my_copy
    }
  },
  components: {CircleChart, TotalChart, RadarChart, BarChart, ScoreChart, /*ScoreChart02*/},
  props: ["msg"],

  data() {
    return {

      pdfName: 'pdf',
      isPdfDownload: true,

      isPdf: false, // 是否生成pdf中

      isGetData: false,

      // 图片变量
      scoreImg: scoreImg,
      cover02: cover02,
      cover03: cover03,
      heightImg: heightScore,
      midImg: midScore,
      lowImg: lowScore,
      noteImg: noteImg,
      myTime: '',
      coverTime: '',

      scores: {},
      heightScores: [],
      midScores: [],
      lowScores: [],
      heightWord:'',
      midWord:'',
      lowWord:'',

      paperId: '',
      testerId: '',

      tester: {
        name: '张三', age: '20', gender:'男', telephone:'19922999908',
        affiliation: '无', post: '无', createTime: '', userTime: 20,
      },

      totalAvg: 0,
      count8: 0,
      healthType: [],
      traitsDefine: null,
      anxiety: null,
      depression: null,
      psyImbalance: null,
      enmity: null,
      fear: null,
      unwell: null,
      cogDecline: null,
      emotional: null,
      frustration: null,
      selfDoubt: null,
      skepticism: null,
      burnout: null,
      evaluation: [],
    };
  },

  created() {

    this.getDictData()

    // this.isPdfDownload = true
    const paperId = this.$route.params.id
    this.examId = this.$route.params.examId
    this.testerId = this.$route.params.testerId
    if (typeof paperId !== 'undefined') {
      this.paperId = paperId
      this.fetchTester(paperId)
      this.fetchScore(paperId)
      console.log(this.paperId)
    }

    this.pdfName = this.getNowTime()
  },

  mounted() {
    // this.isPdfDownload = this.disPdfDownload
    setTimeout(() => {
      // if (this.isPdfDownload) {
        this.createPdf()
      // }

      // this.createPdf()
    }, 4000)

    // const _this = this;
    // this.bodyScale();
    // window.onresize = function() {
    //   _this.bodyScale();
    // }.bind(this);
    //
    // document.addEventListener('gesturechange', this.handlePinch);

  },

  // beforeDestroy() {
  //   document.removeEventListener('gesturechange', this.handlePinch);
  // },

  methods: {

    formatDate(dt) {
      if (!dt) return ''
      return dt.substring(0, 10)
    },

    // bodyScale() {
    //   let devicewidth = document.documentElement.clientWidth;
    //   let scale = devicewidth / 1920;
    //   document.body.style.zoom = scale;
    //
    // },
    // handlePinch(event) {
    //   let scale = event.scale;
    //   document.body.style.transform = `scale(${scale})`;
    // },

    getDictData() {

      this.getDicts('el_health_type').then(res => {
        res.data.forEach(x => {
          this.healthType.push({name: x.dictValue, value: x.dictLabel})
        })
        // console.log(this.healthType)
      })

      this.getDicts('el_traits_define').then(res => {
        this.traitsDefine = res.data
        // console.log(this.traitsDefine)
      })

      this.getDicts('el_anxiety').then(res => {
        let obj = {name: '焦虑'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_depression').then(res => {
        let obj = {name: '抑郁'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_psychological_imbalance').then(res => {
        let obj = {name: '心理失衡'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_enmity').then(res => {
        let obj = {name: '敌意'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_fear').then(res => {
        let obj = {name: '恐惧'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_feeling_unwell').then(res => {
        let obj = {name: '身体不适'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_cognitive_decline').then(res => {
        let obj = {name: '认知衰退'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_emotional').then(res => {
        let obj = {name: '情绪化'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_frustration').then(res => {
        let obj = {name: '挫折感'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_selfdoubt').then(res => {
        let obj = {name: '自我否定'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_skepticism').then(res => {
        let obj = {name: '怀疑感'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

      this.getDicts('el_burnout').then(res => {
        let obj = {name: '职业倦怠'}
        res.data.forEach(x => {
          obj[x.dictValue] = x.dictLabel
        })
        this.evaluation.push(obj)
      })

    },

    getRangeData() {
      let scoresKeys = Object.keys(this.scores)
      scoresKeys.forEach(key => {
        this.totalAvg += this.scores[key]
        let define = this.getDefine(key)

        if(this.scores[key]<0){
          this.scores[key] = 0;
        }else if(this.scores[key]>10){
          this.scores[key] = 10;
        }

        let evaluation = this.getEvaluation(key, this.scores[key])
        if (this.scores[key] <= 4.5) {
          this.lowScores.push({name: key, value: this.scores[key], define: define, evaluation: evaluation})
          this.lowWord += key + '、'
        } else if (this.scores[key] <= 7) {
          this.midScores.push({name: key, value: this.scores[key], define: define, evaluation: evaluation})
          this.midWord += key + '、'
        } else {
          this.heightScores.push({name: key, value: this.scores[key] > 10 ? 10 : this.scores[key], define: define, evaluation: evaluation})
          this.heightWord += key + '、'
          if (this.scores[key] > 8) this.count8++
        }
      })

      this.heightScores.sort((a, b) => {
        return b.value - a.value
      })

      this.midScores.sort((a, b) => {
        return b.value - a.value
      })

      this.lowScores.sort((a, b) => {
        return b.value - a.value
      })

      this.totalAvg = this.totalAvg / 12
      this.lowWord = this.lowWord.substring(0, this.lowWord.length - 1)
      this.midWord = this.midWord.substring(0, this.midWord.length - 1)
      this.heightWord = this.heightWord.substring(0, this.heightWord.length - 1)
      console.log(this.count8)
    },

    getEvaluation(key, scores) {
      // console.log(this.evaluation)
      let evaluation
      this.evaluation.forEach(eva => {
        // console.log(eva.name)
        if (eva.name === key) {
          if (scores <= 2) evaluation =  eva['A']
          else if (scores <= 4) evaluation =  eva['B']
          else if (scores <= 6) evaluation =  eva['C']
          else if (scores <= 8) evaluation =  eva['D']
          else evaluation =  eva['E']
        }
      })

      return evaluation
    },

    getDefine(key) {
      let def
      this.traitsDefine.forEach(define => {
        if (define.dictValue === key) {
          // console.log(define.dictLabel)
          def = define.dictLabel
        }
      })

      return def
    },

    pdfDownload(paperId) {
      // this.isPdfDownload = false
      // this.paperId = paperId
    },

    // setPaperId(paperId) {
    //   this.paperId = paperId
    //   console.log("this.paperId=" + this.paperId)
    // },

    fetchScore(paperId) {
      this.paperId = paperId
      console.log("fetchScore.paperId=" + this.paperId)
      const params = {paperId: paperId}
      if (this.testerId.length === 18) {
        paperResult2(params).then(response => {
          this.scores = response.data
          this.isGetData = true
          this.getRangeData()
          // console.log(this.heightScores)
        })
      } else {
        paperResult(params).then(response => {
          this.scores = response.data
          this.isGetData = true
          this.getRangeData()
          // console.log(this.heightScores)
        })
      }

    },

    fetchTester(paperId) {
      // console.log(this.testerId)
      const params = { paperId: paperId }
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

          let tem = this.formatDate(this.tester.createTime).split('-')
          this.coverTime = tem[0] + ' 年 ' + tem[1] + ' 月 ' + tem[2] + ' 日 '
        })
      } else {
        testerInfo(params).then(response => {

          if (response.data !== null) {
            this.tester = response.data
            if (this.tester.gender === "0") {
              this.tester.gender = "男"
            }

            if (this.tester.gender === "1") {
              this.tester.gender = "女"
            }

          }

          let tem = this.formatDate(this.tester.createTime).split('-')
          this.coverTime = tem[0] + ' 年 ' + tem[1] + ' 月 ' + tem[2] + ' 日 '
        })
      }
    },

    // 生成pdf文件
    async createPdf() {
      return new Promise(async (resolve, reject) => {
        //this.$store.state.pdfStatus.singlePdfFinished = false;
        this.$store.commit('setSinglePdfFinished', false);
        this.isPdf = true;
        await this.$nextTick();

        // const loading = this.$loading({
        //   lock: true,
        //   text: "导出中",
        //   spinner: "el-icon-loading",
        //   background: "rgba(0, 0, 0, 0.7)",
        // });

        const pdfFooter = document.getElementById("pdf-footer");
        const pdfHeader = document.getElementById("pdf-header");
        const firstPageHeader = document.getElementById("first-page-header");
        const firstPageFooter = document.getElementById("first-page-footer");
        setTimeout(() => {
          let pdfDom = document.getElementById("meet-pdf-id");
          let pdfObj = new PdfLoader(pdfDom, {
            fileName: this.pdfName,
            outputType: "datauristring",
            footer: pdfFooter,
            header: pdfHeader,
            headerFirst: firstPageHeader,
            footerFirst: firstPageFooter,
            baseY: 0,
            contentWidth: 595,
            // isPageMessage: true,
          });
          pdfObj.getPdf().then(async (res) => {
            // console.log("[ 导出成功] >", res.pdfResult);
            let dataUri = res.pdfResult;
            await this.UploadPdf(dataUri)
            // await pdfPersistence(pdfBase64)
            // console.log("[ 导出成功] >", base64);
            // loading.close();
            // this.$message.success("导出成功");
            this.isPdf = false;
            this.$store.commit('setSinglePdfFinished', true);
            console.log("pdf finished")
            resolve(res);

          }).catch((error) => {
            this.isPdf = false;
            reject(error);
          });
        });
      }, 500);

    },

    //上传pdf接口
    UploadPdf(res) {
      return new Promise((resolve, reject) => {
        //res拿到base64的pdf
        let pdfBase64Str = res;
        let title ="上传给后端的个人报告"
        let myfile = this.dataURLtoFile(pdfBase64Str, title + ".pdf");//调用一下下面的转文件流函数
        let formdata = new FormData();
        formdata.append("file", myfile); // 文件对象

        if (this.testerId.length === 18) {
          formdata.append("idNumber", this.testerId)
          formdata.append("examId", this.examId)
          pdfPersistence2(formdata).then((res) => {
            console.log("上传pdf接口", res);
            resolve(res)
          }).catch((err) => {
            console.log("上传pdf接口", err);
            reject(err)
          });
        } else {
          console.log("paperId=" + this.paperId)
          formdata.append("paperId", this.paperId)
          //该uploadMy为接口，直接以formdata格式传给后台
          pdfPersistence(formdata).then((res) => {
            console.log("上传pdf接口", res);
            resolve(res)
          }).catch((err) => {
            console.log("上传pdf接口", err);
            reject(err)
          });
        }
      })
    },

    /*
    将base64转换为文件,接收2个参数，第一是base64，第二个是文件名字
    最后返回文件对象
    */
    dataURLtoFile(urlData, filename) {
      let arr = urlData.split('base64,');
      let type = arr[0].match(/:(.*?);/)[1];
      // let fileExt = type.split('/')[1];
      let bstr = atob(arr[1]);
      let n = bstr.length;
      let u8arr = new Uint8Array(n);
      while (n--) {
        u8arr[n] = bstr.charCodeAt(n);
      }
      return new File([u8arr], filename, {type: type});
    },

    getNowTime() {
      let getTime = new Date().getTime(); //获取到当前时间戳
      let time = new Date(getTime); //创建一个日期对象

      let year = time.getFullYear(); // 年
      let month = (time.getMonth() + 1).toString().padStart(2, '0'); // 月
      let date = time.getDate().toString().padStart(2, '0'); // 日
      let hour = time.getHours().toString().padStart(2, '0'); // 时
      let minute = time.getMinutes().toString().padStart(2, '0'); // 分
      let second = time.getSeconds().toString().padStart(2, '0'); // 秒
      return (year + month + date + hour + minute + second);
    }
  },
};
</script>

<style scoped>

.line-img{
  display: flex;
  overflow: hidden;
  align-items: center;
  justify-content: center;
}

.report-cover{
  display: flex;
  flex-direction: column;
  background-image: url('~@/assets/report_images/cover04.jpg');
  background-repeat: no-repeat;

  background-size: cover;
  height: 1122px;
  width: 210mm;
  background-color: #f0f2f5;
  overflow: hidden;
  align-items: center;
  justify-content: center;
}

.report-background{
  display: flex;
  flex-direction: column;
  background-image: url('~@/assets/report_images/background.png');
  background-repeat: no-repeat;

  background-size: cover;
  height: 100%;
  width: 210mm;
  background-color: #f0f2f5;
  overflow: hidden;
  align-items: center;
  justify-content: center;
}

.invitationLetter {
  padding: 20px;
  p,
  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    padding: 0;
    margin: 0;
  }

  .main {
    width: 210mm;
    margin: 0 auto;
    .main__title {
      color: #000;
      font-size: 29px;
      width: 97%;
      margin: 0 auto;
      font-weight: 500;
      text-align: center;
      line-height: 40px;
      font-family: "Times New Roman", "方正小标宋简体";
    }
    .main__text {
      color: #000;
      font-weight: 500;
      line-height: 28px;
      text-align: left;
      font-size: 15px;
      font-family: "Times New Roman", "仿宋_GB2312";
      margin-bottom: 2px;
    }

    .main__time {
      //color: #000;
     /* font-weight: 600;
      line-height: 34px;
      text-align: left;
      font-size: 28px;
      font-family: "Times New Roman", "仿宋_GB2312";
      color: white;*/
      font-size: 12px;
      color: #666666;
      letter-spacing: 0;
      font-weight: 400;
    }

    .main__footer {
      display: flex;
      flex-direction: column;
      align-items: flex-end;
      justify-content: center;
    }
  }
}

.flex-start {
  display: flex;
}
.flex-center {
  display: flex;
  justify-content: center;
  align-items: center;
}
.flex-column-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.w-500 {
  font-weight: 500;
}
.w-600 {
  font-weight: 600;
}
.text__indent {
  text-indent: 2em;
}
.mt-40 {
  margin-top: 40px;
}
.mt-80 {
  margin-top: 80px;
}
.mt-30 {
  margin-top: 30px;
}
.mt-20 {
  margin-top: 20px;
}
.mb-40 {
  margin-bottom: 40px;
}
.mb-30 {
  margin-bottom: 30px;
}
.mb-20 {
  margin-bottom: 20px;
}
.mt-100 {
  margin-top: 80px;
}
.mt-800 {
  /*margin-top: 860px;*/
  margin-top: 950px;
  //margin-left: 55px;
}
.lh-40 {
  line-height: 40px !important;
}

.page {
  min-height: 297mm;
  width: 210mm;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 3.8cm 2.6cm 3.5cm;
}

.one-page {
  min-height: 297mm;
  width: 210mm;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  //padding: 3.8cm 2.6cm 3.5cm;
}

#meet-pdf-id {
  width: 210mm;
  margin: 0 auto;
}

.pdf-hf {
  position: fixed;
  top: -100vh;
  width: 210mm;
}
.pdf__output {
  .page {
    padding: 0 2.6cm;
  }
}

/* 为易截断元素添加保护 */
.pdf-group-item {
  margin-bottom: 15px; /* 安全间距 */
  page-break-inside: avoid; /* 禁止内部拆分 */
}

/* 表格行保护 */
.el-table__row {
  box-shadow: 0 -1px 0 #fff inset; /* 防止底部边框丢失 */
}

.score-t-container{
  position: relative;
  display: flex;
  justify-content: center;
  margin-bottom:15px;
  img{
    display: block; /* 防止图片下方有空隙 */
    height: 60px;
    width: 130px;
  }
  .score-desc{
    position: absolute; /* 绝对定位 */
    font-size: 14px; /* 文字大小 */
    font-weight: 600;
    letter-spacing: 0;
    margin-top: 25px;
  }
  .high{
    color:#F42525;
  }
  .middle{
    color:#FFB900;
  }
  .lower{
    color:#00AE4E;
  }
}
.bg-lred{
  background: rgba(255,0,0,0.05);
  padding: 20px !important;
}
.score-rs-bg{
  background: #FAFBFB;
  border: 1px solid rgba(229,241,233,1);
  padding: 20px;
}
</style>

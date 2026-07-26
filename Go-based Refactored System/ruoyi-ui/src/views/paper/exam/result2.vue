<template>
  <div class="invitationLetter" style="width: 100%">
<!--    <el-button type="primary" @click="createPdf">导出</el-button>-->
    <main id="meet-pdf-id" :class="{ pdf__output: isPdf }">

      <!-- 导出区域 -->
      <div class="main one-page">
        <div style="position: absolute; bottom: 2.5cm ;width: 100%">
          <p class="main__time" style="text-align: center">{{ coverTime }}</p>
        </div>
      </div>

      <div class="page-split mt-40" v-show="!isPdf"></div>

      <div class="main page pdf-split-page" :style="{paddingTop:/*!isPdf ? '10px' :*/'0'}">
        <div v-show="!isPdf">
          <div style="margin-left:-2.6cm;margin-top: 20px;width: 100%;">
            <img src="@/assets/report_images/mng-page-header-left.png" width="21" height="38">
          </div>
          <div style="width: 100%;text-align:right;">
            <span style="font-size: 10px;color: #888888;letter-spacing: 0;font-weight: 400;">管理特质测验报告</span>
          </div>
        </div>
        <h3 class="main__title mt-40 w-600 pdf-group-item" style="margin: 25px 0 15px;">
          <!--报告阅读说明-->
          <img src="@/assets/report_images/mng-desc.png" width="170">
        </h3>
        <p class="main__text mt-40 text__indent pdf-group-item">
          管理特质测验是针对企事业单位高层管理者、中层管理者、基层管理者以及各级储备人员的特质进行测试与评价。
          管理特质测验能够基于组织和岗位需要，对受测者的管理特点、胜任水平、能力倾向进行全面的分析与诊断，
          是组织对各级管理者及储备人员进行招聘选拔、培养开发、评价任用的辅助决策工具。
        </p>
        <p class="main__text text__indent pdf-group-item">
          管理特质测验参考专家学者的研究成果，在上百家企业管理者测评项目实践积累和大量基础理论研究的基础上，
          整合了组织较为重要的13个特质维度：
          社会性、进取性、领导性、计划性、人际敏感性、自信心、责任心、学习力、创新性、情绪稳定性、自律性、决断性、合作性。
          这些特质不仅是对各级管理者及储备人员个人综合个性素质的全面考量，更是确保团队高效运作、推动组织持续健康发展的重要要素。
        </p>
        <div class="line-img"><img :src="noteImg"></div>
        <p class="main__text text__indent pdf-group-item">
          特此说明：本报告绝无冒犯或诋毁之意，所提供的测试结果不能单独作为判断个体管理特质的唯一标准，
          还需综合考虑受测者的日常表现、其他相关测验结果等。因此，本报告仅为辅助参考之用。
        </p>
      </div>

      <!-- 分页 -->
      <div class="page-split mt-40" v-show="!isPdf"></div>
      <div class="main page pdf-split-page" :style="{paddingTop:/*!isPdf ? '10px' :*/'0'}">
        <div>
          <div v-show="!isPdf">
            <div style="margin-left:-2.6cm;margin-top: 20px;width: 100%;">
              <img src="@/assets/report_images/mng-page-header-left.png" width="21" height="38">
            </div>
            <div style="width: 100%;text-align:right;">
              <span style="font-size: 10px;color: #888888;letter-spacing: 0;font-weight: 400;">管理特质测验报告</span>
            </div>
          </div>
          <h3 class="main__title mt-40 w-600 pdf-group-item">报告详细解读</h3>
          <div>
            <!-- FB-041: 节标题 .section-title 在 print CSS 里调大字号+加粗 -->
            <p class="main__text mt-40 lh-40 pdf-group-item section-title-row">
              <span class="w-600 section-title"><img src="@/assets/report_images/mng-no.png" width="20" height="20" style="margin-right: 10px;font-weight: 600;"/>个人信息</span>
            </p>
            <div style="display: flex;flex-direction: column;">
              <div style="display: flex;flex-direction: row;">
                <div class="base-info" style="width: 45%">
                  <ul>
                    <li v-if="showField('name')"><p class="main__text lh-40 pdf-group-item w-600">姓名：{{tester.name}}</p></li>
                    <li v-if="showField('gender')"><p class="main__text lh-40 pdf-group-item w-600">性别：{{tester.gender}}</p></li>
                  </ul>
                </div>
                <div class="base-info" style="width: 55%">
                  <ul>
                    <li v-if="showField('age')"><p class="main__text lh-40 pdf-group-item w-600">年龄：{{tester.age}}岁</p></li>
                    <li v-if="showField('telephone')"><p class="main__text lh-40 pdf-group-item w-600">手机：{{tester.telephone}}</p></li>
                  </ul>
                </div>
              </div>
              <div class="base-info" style="margin-top: -8px;">
                <ul>
                  <li v-if="showField('affiliation')"><p class="main__text lh-40 pdf-group-item w-600">单位：{{tester.affiliation}}</p></li>
                  <li v-if="showField('post')"><p class="main__text lh-40 pdf-group-item w-600">职务：{{tester.post}}</p></li>
                </ul>
              </div>
              <div style="display: flex;flex-direction: row;margin-top: -8px;">
                <div class="base-info" style="width: 45%">
                  <ul>
                    <li><p class="main__text lh-40 pdf-group-item w-600">时间：{{formatDate(tester.createTime)}}</p></li>
                  </ul>
                </div>
                <div class="base-info" style="width: 55%">
                  <ul>
                    <li><p class="main__text lh-40 pdf-group-item w-600">时长：{{tester.userTime}} 分钟</p></li>
                  </ul>
                </div>
              </div>
            </div>
            <p class="main__text text__indent pdf-group-item" v-if="tester.userTime > 35">
              在测试过程中，您的答题时间严重超出预设标准时间，因此，所得测试结果可能存在一定的误差。
            </p>
          </div>

          <div>
            <p class="main__text mt-40 lh-40 pdf-group-item section-title-row">
              <span class="w-600 section-title"><img src="@/assets/report_images/mng-no.png" width="20" height="20" style="margin-right: 10px;font-weight: 600;"/>总体测评结果</span>
            </p>
            <div class="overall-container main__text">
              <div style="height: 350px;">
                <div class="pdf-no-split" style="text-align:center;"><TotalChart :scores="totalScore.toFixed(2)" v-if="isGetData"></TotalChart></div>
              </div>
              <div class="o-item1">
                <div class="o-title">
                  <span class="o-txt">【诊断】</span>
                  <span class="o-level">{{totalResult.mngLevel}}</span>
                </div>
                <div class="o-content">
                  <p>
                    {{totalResult.content}}
                  </p>
                </div>
              </div>
              <div class="o-item2">
                <div class="o-title">
                  <span class="o-txt">【说明】</span>
                </div>
                <div class="o-content">
                  <p>
                    这里的管理特质是指您在所处的具体工作环境与组织氛围中，那些能够显著影响并塑造其领导风格与决策行为的个人内在特征。
                    本测验涵盖13项管理特质，每项特质对整体管理效能均具有特定价值。最终总得分越高，代表您的综合管理潜质与适配性越优秀。
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div class="main__text">
            <p class="mt-40 lh-40 pdf-group-item section-title-row">
              <span class="w-600 section-title"><img src="@/assets/report_images/mng-no.png" width="20" height="20" style="margin-right: 10px;font-weight: 600;"/>测评结果分析</span>
            </p>
            <div class="analysis-container">
              <div class="scores-chart">
                <span class="o-txt">【各维度得分情况】</span>
                <div class="pdf-no-split">
                <div class="pdf-no-split" style="text-align:center;"><RadarChart :scores="scores" v-if="isGetData"></RadarChart></div>
                </div>
                <div class="main__text lh-40">
                  <p class="main__text text__indent pdf-group-item">图示说明：</p>
<!--                  <p class="main__text text__indent pdf-group-item">
                    各维度的评分标准如下：
                  </p>-->
                  <p class="main__text text__indent pdf-group-item">
                    1、得分若处于4＜S≤5的区间内，则被视为高分，表明您在该特征上表现优秀；
                  </p>
                  <p class="main__text text__indent pdf-group-item">
                    2、得分若处于3＜S≤4的区间内，则属于较高分，表示该特征表现良好；
                  </p>
                  <p class="main__text text__indent pdf-group-item">
                    3、得分处于2＜S≤3的区间内，则评定为中等分，表示该特征处于一般水平；
                  </p>
                  <p class="main__text text__indent pdf-group-item">
                    4、得分若处于1＜S≤2的区间内，则为较低分，说明该特征表现有待提高；
                  </p>
                  <p class="main__text text__indent pdf-group-item">
                    5、得分若处于0≤S≤1的区间内，则属于低分，表明该特征表现较差，需要重点关注和改善。
                  </p>
                </div>
              </div>
              <div class="scores-detail">
                <span class="o-txt">【各维度测评结果】</span>
                <div class="detail-item" v-for="item in scoreList">
                  <el-row class="item-title">
                    <el-col :span="12"><img src="@/assets/report_images/star5.png" width="16" height="16" style="margin-right: 10px;"/>{{item.name}}</el-col>
                    <el-col :span="6">得分：{{item.score}}</el-col>
                    <el-col :span="6" style="text-align:right;">等级:{{item.scoreLevel}}</el-col>
                  </el-row>
                  <div class="item-table">
                    <div class="item-tr">
                      <div class="tr-label">说明</div>
                      <div class="tr-content">{{item.define}}</div>
                    </div>
                    <div class="item-tr">
                      <div class="tr-label">测评结果</div>
                      <div class="tr-content">{{item.evaluation}}</div>
                    </div>
                  </div>
                </div>
              </div>

            </div>
          </div>
        </div>
      </div>
    </main>
    <!-- 页头页眉 -->
    <div id="pdf-header" style="height: 2.2cm" class="pdf-hf">
      <div style="width: 100%;">
        <img src="@/assets/report_images/mng-page-header-left.png" width="21" height="38">
      </div>
      <div style="width: 100%;text-align:right;padding-right: 2.6cm">
        <span style="font-size: 10px;color: #888888;letter-spacing: 0;font-weight: 400;">管理特质测验报告</span>
      </div>
    </div>
    <div id="first-page-header" style="height: 0.001cm"></div>
    <div id="pdf-footer" style="height: 2cm" class="pdf-hf">
      <div style="display: flex; justify-content: center; align-items: center; margin-top: 20px;font-size:10px">
        <div style="margin-right: 3px;">第</div>
        <div class="pdf-footer-page"></div>
        <div style="margin-left: 3px;">页</div> <!--/ 共
          <div class="pdf-footer-page-count"></div>
          页-->
      </div>
    </div>
    <div id="first-page-footer" style="height: 0.001cm"></div>
  </div>
</template>

<script>
import { PdfLoader } from "@/utils/pdfLoader";
import cover02 from '@/assets/report_images/cover02.png'
import cover03 from '@/assets/report_images/cover03.png'

import TotalChart from "@/views/paper/exam/components/MngTotalChart.vue";
import RadarChart from "@/views/paper/exam/components/MngRadarChart.vue";

import heightScore from '@/assets/report_images/high-score.png';
import midScore from '@/assets/report_images/middle-score.png';
import lowScore from '@/assets/report_images/lower-score.png';
import noteImg from '@/assets/report_images/mng-note003.png';
import {testerInfo, paperResult, pdfPersistence} from "@/api/candidate/candidate";
import {fetchDetail} from '@/api/exam/exam'
import {getDictsBatch} from '@/api/system/dict/data'
import my_copy from "../../user/exam/my_copy.vue";
import {getTesterByIdNumber, paperResult2, pdfPersistence2} from "@/api/tester/tester";
import scoreImg from '@/assets/report_images/score002.png';

export default {
  name: "result2",
  computed: {
    my_copy() {
      return my_copy
    },
    totalResult(){
      if (this.totalScore > 58) {
        return {
          scoreLevel : '高分' ,
          mngLevel : '卓越管理者',
          content : '具备卓越且全面的管理特质，在复杂多变的领导岗位上能够展现出非凡的领导才能。即使在极具挑战的环境中，也能保持冷静与自信，展现出无与伦比的领导风范。'
        }
      } else if (this.totalScore > 52 && this.totalScore <= 58) {
        return {
          scoreLevel : '较高分' ,
          mngLevel : '优秀管理者',
          content : '具备稳定且可靠的管理特质，在各种场合下都能展现出优秀的管理水平。尽管在个别情况下某些特质可能稍显不足，不过并无大碍，无需过分担忧。'
        }
      } else if (this.totalScore > 45 && this.totalScore <= 52) {
        return {
          scoreLevel : '中等分' ,
          mngLevel : '较好管理者',
          content : '具备基础的管理潜质，能够维持团队的日常运作。但仍有提升空间，需持续努力提升各项管理能力。'
        }
      } else if (this.totalScore > 35 && this.totalScore <= 45) {
        return {
          scoreLevel : '较低分' ,
          mngLevel : '合格管理者',
          content : '当前某些管理特质尚待提升，需进一步挖掘其管理潜力，建议通过提供更多的培训和实践机会，帮助自身快速成长。'
        }
      } else if (this.totalScore <= 35) {
        return {
          scoreLevel : '低分' ,
          mngLevel : '低潜管理者',
          content : '在管理特质方面存在较大的提升空间，需要加大培养和锻炼的力度，建议通过系统的培训和指导，能够逐渐发掘并发挥自己的管理潜能。'
        }
      }else {
        return {
          scoreLevel : '' ,
          mngLevel : '',
          content : ''
        }
      }
    },
    totalContent(){

    }
  },
  components: {TotalChart,RadarChart},
  props: ["msg",'stuFlag','repoCode'],

  data() {
    return {
      r_stuFlag: 0, // 接收从在线考试提交后传递来的参数
      r_repoCode: null, // 接收从在线考试提交后传递来的参数
      requiredFields: [],
      pdfName: 'pdf',
      isPdfDownload: true,

      isPdf: false, // 是否生成pdf中

      isGetData: false,
      _rangeBuilt: false,  // FB-035: 防止 getRangeData 在 evaluation 未齐时跑

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
      totalScore: 0,
      count8: 0,
      // healthType: [],
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
      scoreList:[
        {name: '社会性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '进取性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '领导性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '计划性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '人际敏感性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '自信心', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '责任心', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '学习力', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '创新性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '情绪稳定性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '自律性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '决断性', score : '',scoreLevel : '', define : '', evaluation : ''},
        {name: '合作性', score : '',scoreLevel : '', define : '', evaluation : ''}
      ]
    };
  },

  created() {

    // R7: 必须先赋值 stuFlag/repoCode 再调 getDictData
    // FB-035: 后端 chromedp 路由只进 :id/:testerId，另三个从 query 传进来。
    //         优先读 query，其次兜底读 params (人工跳转场景)
    const params = this.$route.params || {}
    const query = this.$route.query || {}
    const paperId = params.id
    this.examId = query.examId || params.examId
    this.testerId = params.testerId
    this.r_stuFlag = (query.stuFlag !== undefined ? query.stuFlag : params.stuFlag)
    this.r_repoCode = query.repoCode || params.repoCode

    this.getDictData()
    // this.isPdfDownload = true

    // 加载考试配置（勾选信息项）
    if (this.examId) {
      fetchDetail(this.examId).then(res => {
        if (res.data && res.data.requiredFields) {
          this.requiredFields = res.data.requiredFields.split(',')
        }
      }).catch(() => {})
    }

    if (typeof paperId !== 'undefined') {
      this.paperId = paperId
      this.fetchTester(paperId)
      this.fetchScore(paperId)
      console.log(this.paperId)
    }

    this.pdfName = this.getNowTime()
  },

  mounted() {
    const tok = this.$route.query && this.$route.query._internal
    const isInternal = typeof tok === 'string' && /^[a-f0-9]{32}$/i.test(tok)
    if (isInternal) {
      this.isPdf = true
      const style = document.createElement('style')
      style.textContent = `
        /* PDF 模式专用规则（非 @media print，这样 echarts init 时容器已是 600px）*/
        .pdf__output .bar-chart-inner,
        .pdf__output .radar-chart-inner,
        .pdf__output .total-chart-inner,
        .pdf__output .mng-total-chart-inner {
          width: 600px !important;
          margin: 0 auto !important;
        }
        @media print {
          /* 隐藏 jsPDF 专用的页眉页脚调试元素（fixed 会被 Chrome 重复渲染到每页）*/
          #pdf-header, #pdf-footer, #first-page-header, #first-page-footer, .pdf-hf {
            display: none !important;
          }
          /* 强制背景色打印 */
          body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
          /* 去掉浏览器默认 body/html margin（避免封面页四周白边）*/
          html, body { margin: 0 !important; padding: 0 !important; }
          /* .invitationLetter 默认 padding:20px 会让封面背景缩进，在 print 模式去掉 */
          .pdf__output.invitationLetter,
          .invitationLetter:has(.pdf__output) {
            padding: 0 !important;
            margin: 0 !important;
          }
          /* meet-pdf-id 也归零 */
          #meet-pdf-id {
            padding: 0 !important;
            margin: 0 !important;
            width: 210mm !important;
          }
          /* 仅对真正不可分割的小元素：雷达图/单个柱图/图片 */
          .pdf-no-split, .echart-wrap, .score-chart, canvas, svg {
            page-break-inside: avoid !important;
            break-inside: avoid !important;
          }
          .pdf-no-split * {
            page-break-inside: avoid !important;
            break-inside: avoid !important;
          }
          /* FB-035: detail-item (维度卡片) 不可分割，避免 evaluation 文本被分页裁剪丢失 */
          .scores-detail .detail-item,
          .detail-item .item-table,
          .item-table .item-tr {
            page-break-inside: avoid !important;
            break-inside: avoid !important;
          }
          /* 封面页后强制分页 */
          .one-page {
            page-break-after: always !important;
            break-after: page !important;
          }
          /* 内容页 .page padding + 去除阴影边框 */
          .pdf__output .page {
            padding: 0 1.2cm !important;
            box-shadow: none !important;
            border: none !important;
          }
          .pdf__output .one-page,
          .pdf__output .pdf-split-page,
          .pdf__output .main {
            box-shadow: none !important;
            border: none !important;
            outline: none !important;
          }
          /* 封面背景：保持原图比例，不拉伸变形 */
          .pdf__output .one-page {
            height: 297mm !important;
            min-height: 297mm !important;
            max-height: 297mm !important;
            width: 210mm !important;
            margin: 0 !important;
            padding: 0 !important;
            overflow: hidden !important;
            background-size: contain !important;
            background-position: center bottom !important;
            background-repeat: no-repeat !important;
            page-break-after: always !important;
            break-after: page !important;
          }
          /* 字号 / 行高：折中 baseline 视觉密度 + 易读性 */
          /* FB-036: 字体兜底全局应用到 invitationLetter 整树 */
          .pdf__output, .pdf__output *, .pdf__output .one-page, .pdf__output .one-page *,
          .pdf__output .main__time, .pdf__output .main__title {
            font-family: "FangSong", "仿宋", "FangSong_GB2312", "仿宋_GB2312", "STFangsong", "Noto Serif CJK SC", "SimSun", serif !important;
          }
          .pdf__output .page p.main__text,
          .pdf__output .page span,
          .pdf__output .page li {
            font-size: 16px !important;   /* FB-048: 客户要求小四 = 16px */
            line-height: 24px !important; /* FB-048: 客户要求 1.5 倍行距 = 16×1.5 = 24px */
            font-family: "FangSong", "仿宋", "FangSong_GB2312", "仿宋_GB2312", "STFangsong", "Noto Serif CJK SC", "SimSun", serif !important;
          }
          /* FB-048: 全局兑底——.page 下所有后代默认 16px/24px 仿宋，涵盖 .o-txt/.o-level/.tr-label 等元素级 font-size */
          .pdf__output .page,
          .pdf__output .page * {
            font-size: 16px !important;
            line-height: 24px !important;
            font-family: "FangSong", "仿宋", "FangSong_GB2312", "仿宋_GB2312", "STFangsong", "Noto Serif CJK SC", "SimSun", serif !important;
          }
          .pdf__output .page p.main__text {
            margin-bottom: 2px !important;  /* FB-043: 对齐 baseline 2px */
            text-align: justify !important;  /* FB-039: 两端对齐 */
            text-justify: inter-ideograph !important;
          }
          /* FB-041: 仿宋无 Bold 变体，加 text-stroke 模拟粗体 */
          /*   - 节标题（个人信息/总体测评结果/测评结果分析）字号大一号 + 加粗 */
          /*   - .w-600 / .item-title / 各种加粗位置都用 text-stroke 合成粗体 */
          /*   - .o-content > p (诊断/说明文字) 也要两端对齐 */
          /*   - FB-042: 补 .tr-label (维度卡片"说明"/"测评结果") 也加合成粗体 */
          .pdf__output .page .w-600,
          .pdf__output .page .section-title,
          .pdf__output .page .item-title,
          .pdf__output .page .o-txt,
          .pdf__output .page .o-level,
          .pdf__output .page .tr-label {
            font-weight: 600 !important;
            -webkit-text-stroke: 0.4px currentColor !important;
            text-stroke: 0.4px currentColor !important;
          }
          .pdf__output .page .section-title {
            font-size: 21px !important;  /* FB-048: 客户要求三号 = 21px */
            line-height: 22px !important; /* FB-048: 节标题单倍行距 */
          }
          /* FB-048: 节标题所在段落段前段后各 1 行（≈21px）+ 单倍行距 */
          .pdf__output .page p.main__text.section-title-row,
          .pdf__output .page p.section-title-row {
            margin-top: 21px !important;
            margin-bottom: 21px !important;
            line-height: 22px !important;
          }
          /* FB-048: 二级标题（【各维度测评结果】【各维度得分情况】【诊断】【说明】等 .o-txt）
             客户要求：仿宋四号(19px) / 单倍行距 / 段前段后 1 行 */
          .pdf__output .page .o-txt {
            font-size: 19px !important;
            line-height: 19px !important;
            display: inline-block !important;
            margin-top: 19px !important;
            margin-bottom: 19px !important;
          }
          /* .o-level（评级）跟随 .o-txt 同级显示，字号一致 */
          .pdf__output .page .o-level {
            font-size: 19px !important;
            line-height: 19px !important;
          }
          /* 诊断/说明段落两端对齐 */
          .pdf__output .page .o-content p {
            text-align: justify !important;
            text-justify: inter-ideograph !important;
          }
          /* FB-050: h3 主标题——黑体二号(29px)，压缩上方留白 */
          .pdf__output .page h3.main__title,
          .pdf__output .page h3.main__title.mt-40 {
            font-size: 29px !important;
            line-height: 36px !important;
            font-weight: 700 !important;
            margin-top: 0 !important;
            margin-bottom: 12px !important;
            text-align: center !important;
            font-family: "SimHei", "黑体", "Microsoft YaHei", "微软雅黑", "Noto Sans CJK SC", sans-serif !important;
          }
          /* FB-052: 封面日期——四号黑体(19px) */
          .pdf__output .one-page .main__time,
          .pdf__output .one-page p.main__time {
            font-size: 19px !important;
            line-height: 24px !important;
            font-weight: 600 !important;
            font-family: "SimHei", "黑体", "Microsoft YaHei", "微软雅黑", "Noto Sans CJK SC", sans-serif !important;
            color: #000 !important;
          }
          /* FB-050: 个人信息字段（姓名/性别/年龄/手机/单位/职务/时间/时长）不加粗 */
          .pdf__output .page p.main__text.w-600,
          .pdf__output .page li p.main__text {
            font-weight: 400 !important;
            -webkit-text-stroke: 0 !important;
            text-stroke: 0 !important;
          }
          .pdf__output .page .mt-40 {
            margin-top: 0 !important;
          }
        }
      `
      document.head.appendChild(style)
      // 三阶段就绪检测：happy(13齐) / idle(5s无变化) / hard(20s)
      // FB-035: happy 还必须检查 scoreList 最后一个元素的 evaluation 已填充
      //         避免 _tryBuildRange 还在调用中就 ready=true 导致静态模板快照
      const expectedEval = 13
      const start = Date.now()
      let lastEvalLen = 0
      let lastChange = Date.now()
      const timer = setInterval(() => {
        const evalLen = Array.isArray(this.evaluation) ? this.evaluation.length : 0
        if (evalLen !== lastEvalLen) {
          lastEvalLen = evalLen
          lastChange = Date.now()
        }
        const sl = this.scoreList || []
        const slFilled = sl.length === 13 && sl.every(item => item && item.score && item.evaluation)
        const fullReady = this.isGetData && evalLen >= expectedEval && slFilled
        const idleReady = this.isGetData && evalLen > 0 && (Date.now() - lastChange > 5000) && slFilled
        if (fullReady || idleReady) {
          clearInterval(timer)
          if (!fullReady) {
            // eslint-disable-next-line no-console
            console.warn('[pdfgen] partial ready (idle 5s)', { evalLen, expected: expectedEval })
            window.__reportIncomplete = true
          }
          this.$nextTick(() => { setTimeout(() => { window.__reportReady = true }, 200) })
          return
        }
        if (Date.now() - start > 20000) {
          clearInterval(timer)
          // eslint-disable-next-line no-console
          console.warn('[pdfgen] hard timeout, force flush', { isGetData: this.isGetData, evalLen })
          window.__reportIncomplete = true
          window.__reportReady = true
        }
      }, 200)
      return
    }
    setTimeout(() => {
      this.createPdf()
    }, 4000)
  },

  // beforeDestroy() {
  //   document.removeEventListener('gesturechange', this.handlePinch);
  // },

  methods: {

    formatDate(dt) {
      if (!dt) return ''
      return dt.substring(0, 10)
    },

    showField(fieldName) {
      return this.requiredFields.length === 0 || this.requiredFields.includes(fieldName)
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

    // R8: 包装 getDicts，失败时仍 push 占位 obj
    _loadEvalDict(dictName, label) {
      return this.getDicts(dictName).then(res => {
        let obj = {name: label}
        ;(res.data || []).forEach(x => { obj[x.dictValue] = x.dictLabel })
        this.evaluation.push(obj)
        this._tryBuildRange()
      }).catch(err => {
        // eslint-disable-next-line no-console
        console.warn('[pdfgen] getDicts failed', dictName, err)
        this.evaluation.push({name: label, _failed: true})
        this._tryBuildRange()
      })
    },

    _fillEvalFromBatch(batchData, dictName, label) {
      const rows = batchData[dictName] || []
      const obj = {name: label}
      rows.forEach(x => { obj[x.dictValue] = x.dictLabel })
      this.evaluation.push(obj)
      this._tryBuildRange()
    },

    getDictData() {
      // 13 个管理特质评估 dict
      const evalDicts = [
        ['el_sociality', '社会性'],
        ['el_proactiveness', '进取性'],
        ['el_leadership', '领导性'],
        ['el_planning', '计划性'],
        ['el_interpersonal', '人际敏感性'],
        ['el_self_confidence', '自信心'],
        ['el_responsibility', '责任心'],
        ['el_learningability', '学习力'],
        ['el_innovativeness', '创新性'],
        ['el_emotionalstability', '情绪稳定性'],
        ['el_self_discipline', '自律性'],
        ['el_assertiveness', '决断性'],
        ['el_collaboration', '合作性']
      ]
      const allTypes = ['el_mng_traits_define', ...evalDicts.map(d => d[0])]

      // 优先走批量接口（1 次 RTT），失败降级到单个
      getDictsBatch(allTypes).then(res => {
        const data = res.data || {}
        if (data['el_mng_traits_define']) this.traitsDefine = data['el_mng_traits_define']
        evalDicts.forEach(([name, label]) => this._fillEvalFromBatch(data, name, label))
      }).catch(err => {
        // eslint-disable-next-line no-console
        console.warn('[pdfgen] batch dict failed, fallback per-dict', err)
        this.getDicts('el_mng_traits_define').then(res => { this.traitsDefine = res.data }).catch(() => {})
        evalDicts.forEach(([name, label]) => this._loadEvalDict(name, label))
      })
    },

    getRangeData() {
      let scoresKeys = Object.keys(this.scores)
      scoresKeys.forEach(key => {
        this.totalScore += this.scores[key]
        this.scores[key] = this.scores[key].toFixed(2)
        let define = this.getDefine(key)

        if(this.scores[key]<0){
          this.scores[key] = 0;
        }else if(this.scores[key] > 5){
          this.scores[key] = 5;
        }

        // FB-035: 用 findIndex + Vue.set 替代 filter + 直接赋值，确保 Vue 2 reactivity 触发
        const idx = this.scoreList.findIndex(item => item.name == key)
        if (idx < 0) return
        const updated = {
          name: key,
          score: this.scores[key] + '',
          scoreLevel: this.getScoreLevel(this.scores[key]),
          define: this.getDefine(key),
          evaluation: this.getEvaluation(key, this.scores[key])
        }
        this.$set(this.scoreList, idx, updated)
      })
    },

    getEvaluation(key, scores) {
      let evaluation
      this.evaluation.forEach(eva => {
        if (eva.name === key) {
          if (scores > 4 && scores <= 5) evaluation =  eva['A']
          else if (scores > 3 && scores <= 4 ) evaluation =  eva['B']
          else if (scores > 2 && scores <= 3) evaluation =  eva['C']
          else if (scores > 1 && scores <= 2) evaluation =  eva['D']
          else evaluation =  eva['E']
        }
      })

      return evaluation
    },
    getScoreLevel(scores) {
      if (scores > 4 && scores <= 5) {
        return '高分'
      } else if (scores > 3 && scores <= 4) {
        return '较高分'
      } else if (scores > 2 && scores <= 3) {
        return '中等分'
      } else if (scores > 1 && scores <= 2) {
        return '较低分'
      } else {
        return '低分'
      }
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
      const params = {paperId: paperId, repoCode: this.repoCode||this.r_repoCode}
      // 失败仍要设 isGetData=true，让 ready 检测的 idle 兑底能触发
      const onFail = (err) => {
        // eslint-disable-next-line no-console
        console.warn('[pdfgen] fetchScore failed', err)
        this.isGetData = true
      }
      if (this.testerId.length === 18) {
        paperResult2(params).then(response => {
          this.scores = response.data
          console.log('response.data',response.data)
          this.isGetData = true
          this._tryBuildRange()
        }).catch(onFail)
      } else {
        paperResult(params).then(response => {
          this.scores = response.data
          this.isGetData = true
          this._tryBuildRange()
        }).catch(onFail)
      }

    },

    // FB-035: 仅当 evaluation 字典已齐全（>=13）时才调 getRangeData
    _tryBuildRange() {
      if (this._rangeBuilt) return
      const evalReady = Array.isArray(this.evaluation) && this.evaluation.length >= 13
      if (this.isGetData && this.scores && Object.keys(this.scores).length > 0 && evalReady) {
        this._rangeBuilt = true
        this.getRangeData()
      }
    },

    fetchTester(paperId) {
      // console.log(this.testerId)
      const params = { paperId: paperId }
      if (this.testerId.length === 18) {
        getTesterByIdNumber(this.testerId, this.examId).then(response => {

          if (response.data !== null) {
            this.tester = response.data
            // FB-036: gender 兼容 string/number，默认“未知”避免空字段
            const g = String(this.tester.gender == null ? '' : this.tester.gender)
            if (g === '0' || g === '男') this.tester.gender = '男'
            else if (g === '1' || g === '女') this.tester.gender = '女'
            else if (g) this.tester.gender = g
            else this.tester.gender = '未知'

          }

          let tem = this.formatDate(this.tester.createTime).split('-')
          this.coverTime = tem[0] + '年' + tem[1] + '月' + tem[2] + '日'
        })
      } else {
        testerInfo(params).then(response => {

          if (response.data !== null) {
            this.tester = response.data
            const g = String(this.tester.gender == null ? '' : this.tester.gender)
            if (g === '0' || g === '男') this.tester.gender = '男'
            else if (g === '1' || g === '女') this.tester.gender = '女'
            else if (g) this.tester.gender = g
            else this.tester.gender = '未知'

          }

          let tem = this.formatDate(this.tester.createTime).split('-')
          this.coverTime = tem[0] + '年' + tem[1] + '月' + tem[2] + '日'
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
        },1000);
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
  margin: 20px 0;
  display: flex;
  overflow: hidden;
  align-items: center;
  justify-content: center;
  img{
    width:60%;
  }
}

.report-cover{
  //display: flex;
  //flex-direction: column;
  background-image: url('~@/assets/report_images/cover04.jpg');
  background-repeat: no-repeat;

  background-size: cover;
  height: 100%;
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
      color: #1E90FF;
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
      font-size: 16px;
      font-family: "Times New Roman", "仿宋_GB2312";
      margin-bottom: 2px;
    }

    .main__time {
      color: #000;
      font-family: PingFangSC-Regular;
      font-size: 14px;
      //color: #666666;
      letter-spacing: 0;
      font-weight: 400;
    }

    .main__footer {
      display: flex;
      flex-direction: column;
      align-items: flex-end;
      justify-content: center;
    }
    .base-info{
      ul {
        margin: 5px 0;
        li{
          list-style: none;
        }
      }
    }
  }
}

.size12{
  font-size: 12px;
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
  font-size: 20px;
  color: #000000;
  letter-spacing: 0;
}
.text__indent {
  text-indent: 2em;
}
.mt-40 {
  margin: 20px 0 10px !important;
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
  margin-top: 900px;
  margin-left: 55px;
}
.lh-40 {
  line-height: 40px !important;
}

.page {
  min-height: 297mm;
  width: 210mm;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  //padding: 3.8cm 2.6cm 3.5cm;
  padding-left: 2.6cm;
  padding-right: 2.6cm;
  padding-bottom: 3.5cm;
}

.one-page {
  position: relative;
  min-height: 297mm;
  width: 210mm;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  //padding: 3.8cm 2.6cm 3.5cm;
  //background-image: url('~@/assets/report_images/cover04.jpg');
  background-image: url('~@/assets/report_images/mng-cover.jpg');
  background-repeat: no-repeat;

  background-size: cover;
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
.bg-lyellow{
  background: rgba(255,121,0,0.05);
  padding: 20px !important;
}
.bg-lgreen{
  background: rgba(1,141,65,0.05);
  padding: 20px !important;
}
.score-rs-bg{
  background: #FAFBFB;
  border: 1px solid rgba(229,241,233,1);
  padding: 15px;
}
.o-txt{
  font-size: 18px;
  font-weight: 600;
}
.overall-container{
  position: relative;
  background: #F8F8F8;
  border-radius: 8px 8px 8px 8px;
  padding-bottom: 15px;
  .o-content{
    p{
      text-indent: 2em;
      margin: 5px 12px;
    }
  }
  .o-item1{
    margin-top: -110px;
    .o-title{
      .o-level{
        font-size: 20px;
        font-weight: 600;
        color: #1E90FF;
      }
    }
  }
}
.scores-detail{
  margin-top: 20px;
  .detail-item{
    margin-top: 20px;
    .item-title{
      font-weight: 600;
    }
    .item-table{
      display: flex;
      flex-direction: column;
      border-left: 1px solid #DDDDDD;
      border-top: 1px solid #DDDDDD;
      border-right: 1px solid #DDDDDD;
      .item-tr{
        display: flex !important;
        flex-direction: row;
        .tr-label{
          flex: 1;
          padding: 10px;
          border-bottom: 1px solid #DDDDDD;
          border-right: 1px solid #DDDDDD;
          background: #F3F3F3;
          display: flex;
          justify-content: center;
          align-items: center;
        }
        .tr-content{
          width: 75%;
          padding: 10px;
          border-bottom: 1px solid #DDDDDD;

        }
      }
    }
  }
}


</style>

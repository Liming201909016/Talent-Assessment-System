<template>
  <div class="ctn">
    <div class="pdf-ctn">
      <!-- 要导出的部分 -->
      <div class="pdf-panel">
        <div class="pdf-inside-panel">

<!--          封面-->
          <div>
            <div class="report-cover">
              <img :src="cover02">
            </div>

            <h3 class="main__title">心理特质测评报告</h3>
            <h3 class="main__title">Mental Health Assessment Report</h3>
            <div class="line-img"><img :src="cover03"></div>
            <h3 class="main__title">xxx</h3>
            <div class="main__footer mt-100 pdf-group-item">
              <div class="flex-column-center">
                <p class="main__text">2023年6月8号</p>
              </div>
            </div>
          </div>

<!--          报告阅读说明-->
          <div class="pdf-split-page">
            <h3 class="main__title mt-40 w-600 pdf-group-item">报告阅读说明</h3>
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
        </div>

        <div class="pdf-split-page">
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
            <p class="main__text text__indent pdf-group-item">
              时长超过35分钟需说明：在测试过程中，您的答题时间严重超出预设标准时间，因此，所得测试结果可能存在一定的误差。
            </p>
          </div>

          <div>
            <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
              <span class="w-600">二、总体评价</span>
            </p>
            <p class="main__text text__indent pdf-group-item">
              您的心理健康水平：
            </p>
            <TotalChart></TotalChart>
            <div class="main__text lh-40">
              <p class="main__text text__indent pdf-group-item">
                经过对本次心理测评结果的深入分析，得出以下结论：目前您的心理状况较好，但在某些极端或特定情景下，
                可能会出现轻微的心理困扰，不过并无大碍，无需过分担忧。（注意：本报告结果参考价值受限于您作答的结果参考性情况。）
              </p>
              <p class="main__text text__indent pdf-group-item">
                与职业群体进行对比分析，您目前所经历的怀疑感、身体不适、自我否定、认知衰退、敌意、挫折感和焦虑，均属于普遍存在的心理和情感反应。
                当面临工作中的挑战和压力时，这些心理状态并未显示出异常或困扰的迹象。相比之下，您在恐惧、心理失衡、抑郁、情绪化和职业倦怠等心理
                状态上表现更为突出。这些心理状态若不加以关注和处理，可能会对您的心理健康、生活品质和工作效率造成显著的负面影响。
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
                根据常模标准，每个因素若转化后的分数高于7分时，归为高分段因素，以橙色部分呈现，表示心理困扰程度相对比较重。相比之下，
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
              <el-row  class="whole-node" type="flex" align="middle" justify="center" style="margin-top: 20px"><img :src="heightImg"></el-row>
              <p class="main__text text__indent pdf-group-item">
                根据您的作答结果显示，您在恐惧、心理失衡和抑郁因素的得分较高，反映出您的这些心理状态不够理想。
                考虑到这些因素可能对您的心理健康产生不良影响，您应当给予高度重视，并采取积极有效的措施进行改善。
              </p>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】8.00分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】安全感一般，有时会担心出现意外和工作失误。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】恐惧是指个体由于担忧和害怕某些事情或人而产生的显著苦恼情绪。</p>
              </div>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】7.30分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】基本能够接受现状，但当主观愿望与客观现实之间存在较大差距时，会出现一些怨气，觉得自己付出多而得到少。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】心理失衡是指个体感觉到自己的所得与付出不成比例而产生的不满与怨愤。</p>
              </div>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】7.01分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】通常能保持平静，有时可能因工作和生活中的压力感到不安，产生焦躁情绪。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】抑郁是以心境低落为主要特征，个体丧失热忱和乐趣，精力丧失，自我评价过低，消极悲观。</p>
              </div>
            </div>

            <div>
              <el-row class="whole-node" type="flex" align="middle" justify="center" style="margin-top: 20px"><img :src="midImg"></el-row>
              <p class="main__text text__indent pdf-group-item">
                根据您的作答结果显示，您在恐惧、心理失衡和抑郁因素的得分较高，反映出您的这些心理状态不够理想。
                考虑到这些因素可能对您的心理健康产生不良影响，您应当给予高度重视，并采取积极有效的措施进行改善。
              </p>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】8.00分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】安全感一般，有时会担心出现意外和工作失误。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】恐惧是指个体由于担忧和害怕某些事情或人而产生的显著苦恼情绪。</p>
              </div>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】7.30分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】基本能够接受现状，但当主观愿望与客观现实之间存在较大差距时，会出现一些怨气，觉得自己付出多而得到少。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】心理失衡是指个体感觉到自己的所得与付出不成比例而产生的不满与怨愤。</p>
              </div>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】7.01分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】通常能保持平静，有时可能因工作和生活中的压力感到不安，产生焦躁情绪。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】抑郁是以心境低落为主要特征，个体丧失热忱和乐趣，精力丧失，自我评价过低，消极悲观。</p>
              </div>
            </div>

            <div>
              <el-row class="whole-node" type="flex" align="middle" justify="center" style="margin-top: 20px"><img :src="lowImg"></el-row>
              <p class="main__text text__indent pdf-group-item">
                根据您的作答结果显示，您在恐惧、心理失衡和抑郁因素的得分较高，反映出您的这些心理状态不够理想。
                考虑到这些因素可能对您的心理健康产生不良影响，您应当给予高度重视，并采取积极有效的措施进行改善。
              </p>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】8.00分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】安全感一般，有时会担心出现意外和工作失误。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】恐惧是指个体由于担忧和害怕某些事情或人而产生的显著苦恼情绪。</p>
              </div>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】7.30分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】基本能够接受现状，但当主观愿望与客观现实之间存在较大差距时，会出现一些怨气，觉得自己付出多而得到少。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】心理失衡是指个体感觉到自己的所得与付出不成比例而产生的不满与怨愤。</p>
              </div>
              <div>
                <ScoreChart></ScoreChart>
                <p class="main__text text__indent pdf-group-item">【个人得分】7.01分</p>
                <p class="main__text text__indent pdf-group-item">【表现评估】通常能保持平静，有时可能因工作和生活中的压力感到不安，产生焦躁情绪。</p>
                <p class="main__text text__indent pdf-group-item">【概念阐释】抑郁是以心境低落为主要特征，个体丧失热忱和乐趣，精力丧失，自我评价过低，消极悲观。</p>
              </div>
            </div>
          </div>
        </div>


        <div class="pdf-split-page">
          <h3 class="main__title w-600 pdf-group-item">团体报告详细解读</h3>
        </div>

      </div>

      <!-- 页头页尾 -->
      <div
        class="pdf-header"
        style="
          font-weight: bold;
          padding: 30px 8px;
          width: 1200px;
          border-bottom: 1px solid rgba(0, 0, 0, 0.85);
          color: rgba(0, 0, 0, 0.85);
          position: fixed;
          top: -100vh;
        "
      >
      </div>
      <div
        class="pdf-footer"
        style="
          font-weight: bold;
          padding: 30px 8px;
          width: 100%;
          border-top: 1px solid rgba(0, 0, 0, 0.85);
          position: fixed;
          top: -100vh;
        "
      >
<!--        <div style="display: flex; justify-content: center; align-items: center; padding-top: 5px">我是页尾</div>-->
        <div style="display: flex; justify-content: center; align-items: center; margin-top: 20px">
          第
          <div class="pdf-footer-page"></div>
          页 / 共
          <div class="pdf-footer-page-count"></div>
          页
        </div>
      </div>
    </div>
    <div>
      <el-button @click="handleOutput2" type="primary">
        测试导出2
      </el-button>
    </div>
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
import heightScore from '@/assets/report_images/heightScore.png';
import midScore from '@/assets/report_images/midScore.png';
import lowScore from '@/assets/report_images/lowScore.png';
import ScoreChart from "@/views/paper/exam/components/ScoreChart.vue";
import noteImg from '@/assets/report_images/note.png';
import {testerInfo, paperResult} from "@/api/candidate/candidate";

export default {
  name: "HelloWorld",
  components: {CircleChart, TotalChart, RadarChart, BarChart, ScoreChart},

  data() {
    return{

      isGetData: false,

      // 图片变量
      cover02: cover02,
      cover03: cover03,
      heightImg: heightScore,
      midImg: midScore,
      lowImg: lowScore,
      noteImg: noteImg,

      scores: {},

      tester: {
        name: '张三', age: '20', gender:'男', telephone:'19922999908',
        affiliation: '无', post: '无', createTime: '2024-12-01', userTime: 20,
      },

    }
  },

  props: {
    msg: String,
  },

  created() {
    const paperId = this.$route.params.id
    if (typeof paperId !== 'undefined') {
      this.paperId = paperId
      this.fetchTester(paperId)
      this.fetchScore(paperId)
    }
  },

  methods: {

    formatDate(dt) {
      if (!dt) return ''
      return dt.substring(0, 10)
    },

    fetchScore(paperId) {

      const params = {paperId: paperId}
      paperResult(params).then(response => {
        this.scores = response.data
        this.isGetData = true
        console.log(this.scores)
      })

    },


    fetchTester(paperId) {
      const params = { paperId: paperId }
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
      })
    },

    handleOutput2() {
      const loading = this.$loading({
        lock: true,
        text: "导出中",
        spinner: "el-icon-loading",
        background: "rgba(0, 0, 0, 0.7)",
      });

      const element = document.querySelector(".pdf-panel");
      const header = document.querySelector(".pdf-header");
      const footer = document.querySelector(".pdf-footer");

      const pdfLoader = new PdfLoader(element, {
        footer: footer,
        header: header,
        // outputType:'file',
        fileName: "自定义名字",
        direction: "p",
        isPageMessage: true,
        // contentWidth: 595,
      });
      pdfLoader.getPdf().then((res) => {
        console.log("[ 导出成功] >", res);
        loading.close();
        this.$message.success("导出成功");
      });
    },
  },
};
</script>

<style scoped>

.report-cover{
  display: flex;
  flex-direction: column;
  background-image: url('~@/assets/report_images/cover01.png');
  background-repeat: no-repeat;

  background-size: cover;
  height: 50vh;
  background-color: #f0f2f5;
  overflow: hidden;
  align-items: center;
  justify-content: center;
}

.ctn {
  padding: 100px;
  .pdf-ctn {
    width: 1200px;
    margin: 0 auto;
    .pdf-panel {
      position: relative;
    }
  }
}



p, h1, h2, h3, h4, h5, h6 {
  padding: 0;
  margin: 0;
}

.line-img{
  display: flex;
  overflow: hidden;
  align-items: center;
  justify-content: center;
}

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
  line-height: 34px;
  text-align: left;
  font-size: 21px;
  font-family: "Times New Roman", "仿宋_GB2312";
}
.main__footer {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  justify-content: center;
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
.sml_100{
  margin-left: 100px;
}
.pd_100{
  padding-left: 200px;
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
.lh-40 {
  line-height: 40px !important;
}

.page {
  min-height: 297mm;
  width: 210mm;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  padding: 3.8cm 2.6cm 3.5cm; /* 国家标准公文页边距 GB/T 9704-2012 */
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
</style>

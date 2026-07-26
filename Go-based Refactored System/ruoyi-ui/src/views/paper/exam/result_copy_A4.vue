<!--
 * @Descripttion: a4页边距测试 （测试案例我直接拿项目某个代码随便改了一下）
 * @Date: 2023-02-08 15:54:24
 * @LastEditors: 吴为乐/wwl
 * @LastEditTime: 2023-06-14 16:44:36
-->
<template>
  <div class="invitationLetter">
    <el-button type="primary" @click="createPdf">导出</el-button>
    <main id="meet-pdf-id" :class="{ pdf__output: isPdf }">
      <!-- 页头页眉 -->
      <div id="pdf-header" style="height: 3.8cm" class="pdf-hf"></div>
      <div id="pdf-footer" style="height: 3.5cm" class="pdf-hf"></div>

      <!-- 导出区域 -->
      <div class="main page">

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

      <!-- 分页 -->
      <div class="page-split mt-40" v-show="!isPdf"></div>

      <!-- 第二页 -->
      <div class="main page pdf-split-page">
        <h3 class="main__title w-600 pdf-group-item">报告阅读说明</h3>
      </div>

      <!-- 分页 -->
      <div class="page-split mt-40" v-show="!isPdf"></div>
      <!-- 第三页 -->
      <div class="main page pdf-split-page">
        <h3 class="main__title w-600 pdf-group-item">个体报告详细解读</h3>
        <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
          <span class="w-600">一、个人信息</span>
        </p>
        <p class="main__text lh-40 pdf-group-item">
          <span class="w-600 sml_100">姓名：</span>xxx
          <span class="w-600 pd_100" >性别：</span>xxx
        </p>
        <p class="main__text lh-40 pdf-group-item">
          <span class="w-600 sml_100">年龄：</span>xxx
          <span class="w-600 pd_100" >学历：</span>xxx
        </p>
        <p class="main__text lh-40 pdf-group-item">
          <span class="w-600 sml_100">单位：</span>xxx
          <span class="w-600 pd_100" >岗位：</span>xxx
        </p>
        <p class="main__text lh-40 pdf-group-item">
          <span class="w-600 sml_100">手机：</span>xxx
          <span class="w-600 pd_100" >时长：</span>xxx
        </p>
        <p class="main__text lh-40 pdf-group-item">
          <span class="w-600 sml_100">时间：</span>xxx
        </p>
        <p class="main__text text__indent pdf-group-item">
          时长超过35分钟需说明：在测试过程中，您的答题时间严重超出预设标准时间，因此，所得测试结果可能存在一定的误差。
        </p>
        <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
          <span class="w-600">二、总体评价</span>
        </p>
        <p class="main__text text__indent pdf-group-item">
          您的心理健康水平：
        </p>
        <el-row>
          <TotalChart></TotalChart>
        </el-row>
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
      </div>
      <RadarChart></RadarChart>
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
import heightScore from '@/assets/report_images/heightScore.png';
import midScore from '@/assets/report_images/midScore.png';
import lowScore from '@/assets/report_images/lowScore.png';
import ScoreChart from "@/views/paper/exam/components/ScoreChart.vue";

export default {
  name: "invitationLetter",
  components: {CircleChart, TotalChart, RadarChart, BarChart, ScoreChart},

  data() {
    return {

      cover02: cover02,
      cover03: cover03,
      heightImg: heightScore,
      midImg: midScore,
      lowImg: lowScore,

      // pdf名称
      pdfName: "导出的.pdf",
      isPdf: false, // 是否生成pdf中
    };
  },

  created() {},

  methods: {
    // 生成pdf文件
    async createPdf() {
      return new Promise(async (resolve, reject) => {
        this.isPdf = true;
        await this.$nextTick();
        const pdfFooter = document.getElementById("pdf-footer");
        const pdfHeader = document.getElementById("pdf-header");
        setTimeout(() => {
          let pdfDom = document.getElementById("meet-pdf-id");
          let pdfObj = new PdfLoader(pdfDom, {
            fileName: this.pdfName,
            // outputType: "file",
            footer: pdfFooter,
            header: pdfHeader,
            baseY: 0,
            contentWidth: 595,
          });
          pdfObj.getPdf().then(async (res) => {
              this.isPdf = false;
              resolve(res);
            })
            .catch((error) => {
              this.isPdf = false;
              reject(error);
            });
        });
      }, 500);
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

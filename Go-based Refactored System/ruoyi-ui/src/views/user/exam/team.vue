<template>
  <div class="invitationLetter">
<!--    <el-button type="primary" @click="createPdf">导出</el-button>-->
    <main id="meet-pdf-id" :class="{ pdf__output: isPdf }">
      <!-- 页头页眉 -->
      <div id="pdf-header" style="height: 3.8cm" class="pdf-hf"></div>

      <!-- 导出区域 -->

      <div class="main page">
        <h3 class="main__title mt-40 w-600 pdf-group-item">团体报告详细解读</h3>
        <div>
          <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
            <span class="w-600">一、整体情况分析</span>
          </p>
          <p class="main__text text__indent pdf-group-item" v-if="isGetData">
            经过本次心理测评，共有{{teamTotal}}位受测者接受了评估。经过深入分析，该团队的心理状况展现出多元化的特点。具体而言，
            团队中有{{teamHealth[0].value}}位员工的心理状况良好， 不存在明显的心理困扰；另有{{teamHealth[1].value}}位员工的心理状态处于正常范畴，
            没有显著的心理压力或困扰。此外，团队中还有{{teamHealth[2].value}}位员工的心理状况处于一般水平，
            他们偶尔会面临一些心理困扰，可能需要团队提供一定的关注和支持。值得注意的是，团队中有{{teamHealth[3].value}}位员工面临中度心理困扰，这需要团队给予高度重视，
            并及时为他们提供必要的帮助和干预措施。从整体上看，该团队的心理状况以一般状态为主，团队管理者应当根据每位员工的具体情况，灵活调整策略，
            以促进团队的心理健康和稳定发展。
          </p>
          <LineChart :team-health="teamHealth" v-if="isGetData"></LineChart>
          <p class="main__text text__indent pdf-group-item">
            注意：本报告结果参考价值受限于您作答的结果参考性情况。在测试过程中，XXX的答题时间严重超出预设标准时间，因此，所得测试结果可能存在一定的误差。
          </p>
        </div>
        <div>
          <p class="main__text mt-40 lh-40 pdf-group-item text__indent">
            <span class="w-600">二、总体分析报告</span>
          </p>
          <p class="main__text text__indent pdf-group-item">
            与职业团体进行对比分析，该团队目前所经历的职业倦怠、敌意和挫折感，均为普遍存在的心理和情感反应。当面临工作中的挑战和压力时，这些心理状
            态并未显示出异常或困扰的迹象。然而，值得注意的是，该团队在抑郁、怀疑感、心理失衡、焦虑、情绪化、认知衰退、身体不适、恐惧和自我否定等心
            理状态上表现更为突出。这些心理状态若不加以关注和处理，可能会对团队成员的心理健康、生活品质和工作效率造成显著的负面影响。因此，团队管理
            者应当密切关注这些心理状态，并采取积极的措施进行干预和缓解，以确保该团队在心理健康和工作效率方面均能得到有效的保障。
          </p>
          <RadarChart :scores="teamTraits" v-if="isGetData"></RadarChart>
          <div class="main__text lh-40">
            <p class="main__text text__indent pdf-group-item">图示说明：</p>
            <p class="main__text text__indent pdf-group-item">1、团体常模标准分的参照分数为5.5；</p>
            <p class="main__text text__indent pdf-group-item">
              2、团体各项得分均以标准十分制呈现，是在参考群体常模的基础上，将受测者在12个因素上原始得分按照转化规则转化后得到的分数，
              然后将该团体的各项分数取平均数。
            </p>
            <p class="main__text text__indent pdf-group-item">
              本报告以团体测试得分为依据，根据常模标准，每个因素若转化后的分数高于7分时，归为高分段因素，以橙色部分呈现，表示心理困扰程度相对比较重。
              相比之下，分数位于4.5分至7分之间，归为中分段因素，以橙色部分呈现，表示心理困扰程度比较轻。最后，分数低于4.5分时，归为低分段因素，
              以绿色部分呈现，表示心理状态良好。
            </p>
          </div>
          <BarChart :scores="teamTraits" v-if="isGetData"></BarChart>
          <p class="main__text text__indent pdf-group-item">{{teamTotal}}位员工的心理特质测试结果见下表：</p>
          <el-table :data="tableData" border  size="mini" class="pdf-group-item">
            <el-table-column prop="焦虑" label="焦虑" width="47"/>
            <el-table-column prop="抑郁" label="抑郁" width="47"/>
            <el-table-column prop="心理失衡" label="心理失衡" width="50"/>
            <el-table-column prop="敌意" label="敌意" width="50"/>
            <el-table-column prop="恐惧" label="恐惧" width="50"/>
            <el-table-column prop="身体不适" label="身体不适" width="50"/>
            <el-table-column prop="认知衰退" label="认知衰退" width="50"/>
            <el-table-column prop="情绪化" label="情绪化" width="50"/>
            <el-table-column prop="挫折感" label="挫折感" width="50"/>
            <el-table-column prop="自我否定" label="自我否定" width="55"/>
            <el-table-column prop="怀疑感" label="怀疑感" width="50"/>
            <el-table-column prop="职业倦怠" label="职业倦怠" width="47"/>
          </el-table>
        </div>
      </div>

      <!-- 分页 -->

      <div id="pdf-footer" style="height: 3.5cm" class="pdf-hf">
        <div style="display: flex; justify-content: center; align-items: center; margin-top: 20px">
          第
          <div class="pdf-footer-page"></div>
          页 / 共
          <div class="pdf-footer-page-count"></div>
          页
        </div>
      </div>
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
import noteImg from '@/assets/report_images/note.png';
import {pdfTeam} from "@/api/exam/exam";
import LineChart from "@/views/user/components/LineChart.vue";
import {getTeamScore} from "@/api/candidate/candidate";

export default {
  name: "result",
  components: {LineChart, CircleChart, TotalChart, RadarChart, BarChart, ScoreChart},
  props: ["msg", "examId", 'isOpen'],

  data() {
    return {

      pdfName: 'pdf',
      isPdfDownload: true,

      isPdf: false, // 是否生成pdf中

      isGetData: false,

      // 图片变量
      cover02: cover02,
      cover03: cover03,
      heightImg: heightScore,
      midImg: midScore,
      lowImg: lowScore,
      noteImg: noteImg,

      scores: {},

      teamHealth: [],
      teamTraits: {},
      teamTotal: 0,
      tableData: [],

      paperId: '',

      tester: {
        name: '张三', age: '20', gender:'男', telephone:'19922999908',
        affiliation: '无', post: '无', createTime: '2024-12-01', userTime: 20,
      },
    };
  },

  created() {
    this.handleTeamScore()
  },

  mounted() {
    setTimeout(() => {
      this.createPdf()
    }, 5000)
  },

  methods: {

    handleTeamScore() {
      getTeamScore({examId: this.examId, isOpen: this.isOpen}).then(res => {
        console.log(this.isOpen)
        let keys = Object.keys(res.data[0]);
        keys.forEach(key => {
          let avg = 0;
          res.data.forEach(x => {
            avg += x[key]
          })
          avg = avg / res.data.length
          // this.tableData.push({name: key, score: avg})
          this.teamTraits[key] = avg.toFixed(2)
          // this.tableData.push({key: avg})
          // console.log(this.tableData)
          this.tableData = res.data
        })

        console.log(res)

        let h1 = 0;
        let h2 = 0;
        let h3 = 0;
        let h4 = 0;
        let h5 = 0;
        res.data.forEach(x => {
          this.teamTotal++
          let avg = 0;
          let count8 = 0
          keys.forEach(key => {
            avg += x[key]
            if (x[key] > 8) {
              count8++;
            }
          })
          avg /= 12
          // console.log(avg)
          if (avg <= 5 && count8 <= 1) {
            h1++;
            this.topText = '心理状况良好'
          } else if (avg <= 6 && count8 <= 3) {
            h2++;
            this.topText = '心理状态正常'
          } else if (avg <= 7 && count8 <= 5) {
            h3++;
            this.topText = '心理状况一般'
          } else if (avg > 8 && count8 <= 6) {
            h4++;
            this.topText = '中度心理困扰'
          } else if (avg > 8 || count8 > 8) {
            h5++;
            this.topText = '重度心理困扰'
          }

        })
        this.teamHealth.push({name: '心理状况良好', value: h1}, {name: '心理状态正常', value: h2},
          {name: '心理状况一般', value: h3}, {name: '中度心理困扰', value: h4}, {name: '重度心理困扰', value: h5})

        this.isGetData = true

        // console.log(this.teamHealth)
        // console.log(this.teamTraits)
      })
    },

    async createPdf() {
      return new Promise(async (resolve, reject) => {
        this.isPdf = true;
        await this.$nextTick();

        const pdfFooter = document.getElementById("pdf-footer");
        const pdfHeader = document.getElementById("pdf-header");
        // const firstPageHeader = document.getElementById("first-page-header");
        // const firstPageFooter = document.getElementById("first-page-footer");
        setTimeout(() => {
          let pdfDom = document.getElementById("meet-pdf-id");
          let pdfObj = new PdfLoader(pdfDom, {
            fileName: this.pdfName,
            outputType: "datauristring",
            footer: pdfFooter,
            header: pdfHeader,
            // headerFirst: firstPageHeader,
            // footerFirst: firstPageFooter,
            baseY: 0,
            contentWidth: 595,
            // isPageMessage: true,
          });
          pdfObj.getPdf().then(async (res) => {
            let dataUri = res.pdfResult;
            this.UploadPdf(dataUri)
            this.isPdf = false;
            resolve(res);
          }).catch((error) => {
            this.isPdf = false;
            reject(error);
          });
        });
      }, 500);
    },

    UploadPdf(res) {
      //res拿到base64的pdf
      let pdfBase64Str = res;
      let title ="上传给后端的个人报告"
      let myfile = this.dataURLtoFile(pdfBase64Str, title + ".pdf");//调用一下下面的转文件流函数
      let formdata = new FormData();
      formdata.append("file", myfile); // 文件对象
      formdata.append("examId", this.examId)
      pdfTeam(formdata).then((res) => {
        console.log("上传pdf接口", res);
      })
        .catch((err) => {
          console.log("上传pdf接口", err);
        });
    },

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
  padding: 3.8cm 2.6cm 3.5cm;
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

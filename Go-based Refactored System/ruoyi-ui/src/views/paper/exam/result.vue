<template>
  <div class="invitationLetter" style="width: 100%">
<!--    <el-button type="primary" @click="createPdf">导出</el-button>-->
    <main id="meet-pdf-id" :class="{ pdf__output: isPdf }">

      <!-- 导出区域 -->
      <div class="main one-page">
        <div style="position: absolute; bottom: 2.5cm ;width: 100%">
          <p class="main__time" style="text-align: center">{{ coverTime }}</p>
<!--          <div class="report-cover" >-->
<!--            <div style="position: relative; bottom: 180px; left: 150px; font-size: 15px">-->
<!--              <p>测评编号：{{tester.paperId}}</p>-->
<!--            </div>-->
<!--                      <div class="mt-800 pdf-group-item" style="text-align: center">
                        <div class="flex-column-center">
                          <p class="main__time" style="text-align: center">{{ coverTime }}</p>
                        </div>
                      </div>
          </div>-->
<!--          <h3 class="main__title">职业心理测评报告</h3>-->
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

      <div class="main page pdf-split-page" :style="{paddingTop:!isPdf ? '10px' :'0'}">
        <div v-show="!isPdf">
          <div style="margin-left:-2.6cm;margin-top: 20px;width: 100%;">
            <img src="@/assets/report_images/page-header-left.png" width="20" height="22">
          </div>
          <div style="width: 100%;text-align:right;">
            <span style="font-size: 10px;color: #888888;letter-spacing: 0;font-weight: 400;">职业心理测评报告</span>
          </div>
        </div>
        <h3 class="main__title mt-40 w-600 pdf-group-item" style="margin: 25px 0 15px;">
          <!--报告阅读说明-->
          <img src="@/assets/report_images/desc.png" width="230">
        </h3>
        <p class="main__text mt-40 text__indent pdf-group-item">
          在社会持续发展与竞争日益激烈的当下，人们面临的心理压力和挑战不断升级，因此，心理健康问题已逐渐上升为职场人士最为关注的焦点。
          心理健康不仅意味着个体不存在心理疾病，更是一种内心的平衡与和谐状态，是成就个人价值和实现幸福生活的根本所在。经调研发现，
          越是在大城市生活、拥有高学历和高收入的人群，即所谓的“高端人群”，越可能面临心理问题的挑战。
        </p>
        <p class="main__text text__indent pdf-group-item" v-if="stuFlag == 1||r_stuFlag == 1">
          本测验旨在全面揭示受测者的心理健康状况，并深入探究其在焦虑、抑郁、心理失衡、敌意、恐惧、身体不适、认知衰退、情绪化、挫折感、
          自我否定、怀疑感以及倦怠感等十二个关键心理维度上的具体表现。这些心理状态，若未能得到妥善管理和干预，可能转化为心理风险因素，
          进而增加受测者在生活和学习中出现偏差的可能性，从而对个人的学习效果和职业发展产生直接的负面影响。因此，从长远利益考虑，
          学校了解学生的心理状态是非常必要的。
        </p>
        <p class="main__text text__indent pdf-group-item" v-else>
          本测验旨在全面揭示受测者的心理健康状况，并深入探究其在焦虑、抑郁、心理失衡、敌意、恐惧、身体不适、认知衰退、情绪化、挫折感、
          自我否定、怀疑感以及职业倦怠等十二个关键心理维度上的具体表现。这些心理状态，若未能得到妥善管理和干预，可能转化为心理风险因素，
          进而增加受测者在工作中出现偏差的可能性，从而对个人的工作绩效以及企业的稳定发展产生直接的负面影响。因此，从长远利益考虑，
          企业了解员工的心理状态是非常必要的。
        </p>
        <div class="line-img">
          <img :src="noteImg2" v-if="stuFlag == 1||r_stuFlag == 1">
          <img :src="noteImg" v-else>
        </div>
        <p class="main__text text__indent pdf-group-item">
          特此说明：本报告绝无冒犯或诋毁之意，所提供的测试结果不能单独作为判断个体心理健康状况的唯一标准，
          还需综合考虑受测者的日常表现、其他相关测验结果以及专业医生的诊断意见。因此，本报告仅为辅助参考之用。
        </p>
      </div>

      <!-- 分页 -->
      <div class="page-split mt-40" v-show="!isPdf"></div>
      <div class="main page pdf-split-page" :style="{paddingTop:!isPdf ? '10px' :'0'}">
        <div>
          <div v-show="!isPdf">
            <div style="margin-left:-2.6cm;margin-top: 20px;width: 100%;">
              <img src="@/assets/report_images/page-header-left.png" width="20" height="22">
            </div>
            <div style="width: 100%;text-align:right;">
              <span style="font-size: 10px;color: #888888;letter-spacing: 0;font-weight: 400;">职业心理测评报告</span>
            </div>
          </div>
          <h3 class="main__title mt-40 w-600 pdf-group-item">个体报告详细解读</h3>
          <div>
            <p class="main__text mt-40 lh-40 pdf-group-item">
              <span class="w-600 section-title"><img src="@/assets/report_images/no.png" width="17" height="12" style="margin-right: 10px;"/>个人信息</span>
            </p>
            <!-- 学生版 -->
            <div v-if="stuFlag == 1 || r_stuFlag == 1" style="display: flex;flex-direction: column;">
              <div style="display: flex;flex-direction: row;">
                <div class="base-info" style="width: 45%">
                  <ul>
                    <li v-if="showField('name')"><p class="main__text lh-40 pdf-group-item w-600">姓名：{{tester.name}}</p></li>
                    <li v-if="showField('gender')"><p class="main__text lh-40 pdf-group-item w-600">性别：{{tester.gender}}</p></li>
                    <li v-if="showField('degree')"><p class="main__text lh-40 pdf-group-item w-600">学历：{{tester.degree}}</p></li>
                  </ul>
                </div>
                <div class="base-info" style="width: 55%">
                  <ul>
                    <li v-if="showField('age')"><p class="main__text lh-40 pdf-group-item w-600">年龄：{{tester.age}}岁</p></li>
                    <li v-if="showField('telephone')"><p class="main__text lh-40 pdf-group-item w-600">手机：{{tester.telephone}}</p></li>
                    <li v-if="showField('major')"><p class="main__text lh-40 pdf-group-item w-600">专业：{{tester.major}}</p></li>
                  </ul>
                </div>
              </div>
              <div class="base-info" style="margin-top: -8px;">
                <ul>
                  <li v-if="showField('affiliation')"><p class="main__text lh-40 pdf-group-item w-600">学校：{{tester.affiliation}}</p></li>
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
            <!-- 职场版 -->
            <div v-else style="display: flex;flex-direction: column;">
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
                  <li v-if="showField('post')"><p class="main__text lh-40 pdf-group-item w-600">岗位：{{tester.post}}</p></li>
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
            <p class="main__text mt-40 lh-40 pdf-group-item">
              <span class="w-600 section-title"><img src="@/assets/report_images/no.png" width="17" height="12" style="margin-right: 10px;"/>总体评价</span>
            </p>
            <p class="main__text text__indent pdf-group-item">
              您的心理健康水平：
            </p>
            <div class="pdf-no-split" style="text-align:center;"><TotalChart :count8="count8" :totalAvg="totalAvg" v-if="isGetData"></TotalChart></div>
            <div class="main__text lh-40">
              <p class="main__text text__indent pdf-group-item">
                经过对本次心理测评结果的深入分析，得出以下结论：目前您的{{totalContent}} （注意：本报告结果参考价值受限于您作答的结果参考性情况。）
              </p>
              <p class="main__text text__indent pdf-group-item" v-if="stuFlag == 1||r_stuFlag == 1">
                与学生群体进行对比分析，您目前所经历的{{lowWord}}、{{midWord}}，均属于普遍存在的心理和情感反应。
                当面临生活和学习中的挑战和压力时，这些心理状态并未显示出异常或困扰的迹象。相比之下，您在{{heightWord}}等心理状态上表现更为突出。
                这些心理状态若不加以关注和处理，可能会对您的心理健康、生活品质和学习效率造成显著的负面影响。
              </p>
              <p class="main__text text__indent pdf-group-item" v-else>
                与职业群体进行对比分析，您目前所经历的{{lowWord}}、{{midWord}}，均属于普遍存在的心理和情感反应。
                当面临工作中的挑战和压力时，这些心理状态并未显示出异常或困扰的迹象。相比之下，您在{{heightWord}}等心理状态上表现更为突出。这些心理状态若不加以关注和处理，可能会对您的心理健康、生活品质和工作效率造成显著的负面影响。
              </p>
            </div>
            <div class="pdf-no-split" style="text-align:center;">
              <RadarChart :scores="scores" v-if="isGetData" :stuFlag="stuFlag==1||r_stuFlag==1"></RadarChart>
            </div>
            <div class="main__text lh-40">
              <p class="main__text text__indent pdf-group-item">图示说明：</p>
              <p class="main__text text__indent pdf-group-item">1、常模标准分的参照分数为5.5；</p>
              <p class="main__text text__indent pdf-group-item">
                2、个人各项得分均以标准十分制呈现，是在参考群体常模的基础上，将受测者在12个因素上原始得分按照转化规则转化后得到的分数。</p>
            </div>
          </div>

          <div>
            <p class="main__text mt-40 lh-40 pdf-group-item">
              <span class="w-600 section-title"><img src="@/assets/report_images/no.png" width="17" height="12" style="margin-right: 10px;"/>综合评价报告</span>
            </p>
            <p class="main__text text__indent pdf-group-item">
              本报告以您的个人测试得分为依据，从心理状况的高、中、低分段因素出发，进行了详尽的分析与评估，力求全面、客观地揭示您的心理状态。
              具体来说，低分段得分表明您的心理健康状况良好；而得分逐渐升高，进入高分段，则反映出您面临的心理困扰愈发严峻。
            </p>
            <div class="pdf-no-split" style="text-align:center;"><BarChart :scores="scores" v-if="isGetData"></BarChart></div>
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
            <p class="main__text mt-40 lh-40 pdf-group-item">
              <span class="w-600 section-title"><img src="@/assets/report_images/no.png" width="17" height="12" style="margin-right: 10px;"/>分层评价报告</span>
            </p>
            <!-- FB-037: heightScores 为空时整个高分段块隐藏（含标题图标+描述+进度条） -->
            <div v-if="heightScores && heightScores.length > 0">
<!--
              <el-row  class="whole-node " type="flex" align="middle" justify="center" style="margin-top: 20px"><img class="pdf-group-item" :src="heightImg"></el-row>
-->
              <div class="score-t-container">
                <img :src="heightImg">
                <span class="score-desc high">高分段元素</span>
              </div>
              <p class="main__text text__indent pdf-group-item bg-lred">
                根据您的作答结果显示，您在{{heightWord}}因素的得分较高，反映出您的这些心理状态不够理想。
                考虑到这些因素可能对您的心理健康产生不良影响，您应当给予高度重视，并采取积极有效的措施进行改善。
              </p>
<!--              高分段因素 - FB-047: 去掉 pdf-no-split 让评估文本可跨页，仅 ScoreChart 进度条单独保护 -->
              <div v-for="data in heightScores" class="score-detail-item">
                <div class="pdf-group-item score-card-keep">
                  <ScoreChart :traits="data.name" :score="data.value"></ScoreChart>
                </div>
                <div class="score-rs-bg">
                  <p class="main__text text__indent pdf-group-item">【个人得分】{{data.value}}分</p>
                  <p class="main__text text__indent pdf-group-item">【表现评估】{{ data.evaluation }}</p>
                  <p class="main__text text__indent pdf-group-item">【概念阐释】{{data.define}}</p>
                </div>
              </div>
            </div>

            <div>
<!--
              <el-row class="whole-node " type="flex" align="middle" justify="center" style="margin-top: 20px"><el-image class="pdf-group-item" :src="midImg"></el-image></el-row>
-->
              <div class="score-t-container" style=" margin-top: 30px;">
                <img :src="midImg">
                <span class="score-desc middle">中分段元素</span>
              </div>
<!--              <img :src="midImg">-->
              <p class="main__text text__indent pdf-group-item bg-lyellow">
                根据您的作答结果显示，您在{{midWord}}因素的得分适中，反映出您的这些心理状态处于尚可水平。
                考虑到这些因素对您心理健康的潜在威胁，您应当主动关注自己的心理状态，并及时寻求专业的心理疏导。
              </p>
<!--              中分段因素 - FB-047 -->
              <div v-for="data in midScores" class="score-detail-item">
                <div class="pdf-group-item score-card-keep">
                  <ScoreChart :traits="data.name" :score="data.value"></ScoreChart>
                </div>
                <div class="score-rs-bg">
                  <p class="main__text text__indent pdf-group-item">【个人得分】{{data.value}}分</p>
                  <p class="main__text text__indent pdf-group-item">【表现评估】{{ data.evaluation }}</p>
                  <p class="main__text text__indent pdf-group-item">【概念阐释】{{data.define}}</p>
                </div>

              </div>
            </div>

            <div >
<!--
              <el-row class="whole-node " type="flex" align="middle" justify="center" style="margin-top: 20px"><el-image class="pdf-group-item" :src="lowImg"></el-image></el-row>
-->
              <div class="score-t-container" style=" margin-top: 30px;">
                <img :src="lowImg">
                <span class="score-desc lower">低分段元素</span>
              </div>
<!--              <img :src="lowImg" />-->
              <p class="main__text text__indent pdf-group-item bg-lgreen" >
                根据您的作答结果显示，您在{{lowWord}}因素的得分较低，
                反映出您在这些方面拥有较为健康的心理状态。您在保持现有良好心理状态的基础上，持续加强心理健康的维护与提升工作，以保持一个健康、稳定的心理状态。
              </p>
<!--              低分段因素 - FB-047 -->
              <div v-for="data in lowScores" class="score-detail-item">
                <div class="pdf-group-item score-card-keep">
                  <ScoreChart :traits="data.name" :score="data.value"></ScoreChart>
                </div>
                <div class="score-rs-bg">
                  <p class="main__text text__indent pdf-group-item">【个人得分】{{data.value}}分</p>
                  <p class="main__text text__indent pdf-group-item">【表现评估】{{ data.evaluation }}</p>
                  <p class="main__text text__indent pdf-group-item">【概念阐释】{{data.define}}</p>
                </div>

              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
    <!-- 页头页眉 -->
    <div id="pdf-header" style="height: 2.2cm" class="pdf-hf">
      <div style="margin-top: 20px;width: 100%;">
        <img src="@/assets/report_images/page-header-left.png" width="20" height="22">
      </div>
      <div style="width: 100%;text-align:right;padding-right: 2.6cm">
        <span style="font-size: 10px;color: #888888;letter-spacing: 0;font-weight: 400;">职业心理测评报告</span>
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
import CircleChart from "@/views/paper/exam/components/CircleChart.vue";
import TotalChart from "@/views/paper/exam/components/TotalChart.vue";
import RadarChart from "@/views/paper/exam/components/RadarChart.vue";
import BarChart from "@/views/paper/exam/components/BarChart.vue";
// import heightScore from '@/assets/report_images/heightScore002.png';
// import midScore from '@/assets/report_images/midScore002.png';
// import lowScore from '@/assets/report_images/lowScore002.png';
import heightScore from '@/assets/report_images/high-score.png';
import midScore from '@/assets/report_images/middle-score.png';
import lowScore from '@/assets/report_images/lower-score.png';
import ScoreChart from "@/views/paper/exam/components/ScoreChart.vue";
import ScoreChart02 from "@/views/paper/exam/components/ScoreChart02.vue";
import noteImg from '@/assets/report_images/note003.png';
import noteImg2 from '@/assets/report_images/note0032.png';
import {testerInfo, paperResult, pdfPersistence} from "@/api/candidate/candidate";
import {fetchDetail} from '@/api/exam/exam'
import {getDictsBatch} from '@/api/system/dict/data'
import my_copy from "../../user/exam/my_copy.vue";
import {getTesterByIdNumber, paperResult2, pdfPersistence2} from "@/api/tester/tester";
import scoreImg from '@/assets/report_images/score002.png';

export default {
  name: "result",
  computed: {
    my_copy() {
      return my_copy
    },
    totalContent(){
      if (this.totalAvg <= 7 && this.count8 <= 1) {
        return '心理状况良好，未展现出明显的心理困扰，无需额外关注'
      } else if (this.totalAvg <= 7 && this.count8 <= 3) {
        return '心理状况较好，但在某些极端或特定情景下，可能会出现轻微的心理困扰，不过并无大碍，无需过分担忧。'
      } else if (this.totalAvg <= 7 && this.count8 <= 5) {
        return '心理状态尚可，时有出现心理困扰，需要提高警惕，并采取适当措施进行调整，以防止状况进一步加重。'
      } else if (this.totalAvg > 7 && this.count8 <= 6) {
        return '心理状态较差，某些方面存在明显的心理困扰，可能会对生活和工作产生不良影响。建议迅速采取措施，积极调整自身心态，或寻求专业心理咨询师的帮助和建议。'
      } else if (this.totalAvg > 7 || this.count8 > 8) {
        return '心理状态很差，多方面存在明显的心理困扰，可能会对生活和工作造成严重影响。建议立即采取果断行动，积极寻求专业心理医疗卫生机构的帮助和支持。'
      }else{
        return '心理状况较好，但在某些极端或特定情景下可能会出现轻微的心理困扰，不过并无大碍，无需过分担忧。'
      }
    }
  },
  components: {CircleChart, TotalChart, RadarChart, BarChart, ScoreChart, ScoreChart02},
  props: ["msg",'stuFlag','repoCode'],

  data() {
    return {
      r_stuFlag: 0, // 接收从在线考试提交后传递来的参数
      r_repoCode: null, // 接收从在线考试提交后传递来的参数
      requiredFields: [], // 开放模式下勾选的信息项
      pdfName: 'pdf',
      isPdfDownload: true,

      isPdf: false, // 是否生成pdf中

      isGetData: false,
      _rangeBuilt: false,  // FB-035: 防止 getRangeData 在 evaluation 字典加载完之前跑（导致评估全空）

      // 图片变量
      scoreImg: scoreImg,
      cover02: cover02,
      cover03: cover03,
      heightImg: heightScore,
      midImg: midScore,
      lowImg: lowScore,
      noteImg: noteImg,
      noteImg2: noteImg2,
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

    // R7: 必须先赋值 stuFlag 再调 getDictData，否则学生版会拿成职工版字典
    // FB-035: 后端 chromedp 路由只进 :id/:testerId，另三个从 query 传进来。
    //         优先读 query，其次兑底读 params (人工跳转场景)
    const params = this.$route.params || {}
    const query = this.$route.query || {}
    const paperId = params.id
    this.examId = query.examId || params.examId
    this.testerId = params.testerId
    this.r_stuFlag = (query.stuFlag !== undefined ? query.stuFlag : params.stuFlag)
    this.r_repoCode = query.repoCode || params.repoCode

    this.getDictData()
    // this.isPdfDownload = true

    // 加载考试配置（获取勾选信息项）
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
    // 后端生成模式：URL 带 ?_internal=<token>（32位hex）时不调 createPdf，
    // 仅暴露 window.__reportReady = true 让 chromedp 拍 PDF。
    // token 长度校验避免外部用户随便加 ?_internal=1 就能跳过 createPdf。
    const tok = this.$route.query && this.$route.query._internal
    const isInternal = typeof tok === 'string' && /^[a-f0-9]{32}$/i.test(tok)
    if (isInternal) {
      // 设 isPdf=true 让 v-show=!isPdf 的调试元素隐藏
      // + 加 print CSS 隐藏 #pdf-header/#pdf-footer（仅供 jsPDF 抓取的 fixed 元素）
      // Chrome PrintToPDF 会把 fixed 元素重复渲染到每一页，造成"第 页"乱位
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
          /* 仅对真正不可分割的小元素：雷达图/单个柱图/图片
             不能给 .pdf-split-page、.main 等大块加，否则会出空白页 */
          .pdf-no-split, .echart-wrap, .score-chart, canvas, svg {
            page-break-inside: avoid !important;
            break-inside: avoid !important;
          }
          /* RadarChart 内部 echarts 容器（保险）*/
          .pdf-no-split * {
            page-break-inside: avoid !important;
            break-inside: avoid !important;
          }
          /* FB-047: 维度卡片整体允许跨页（消除页底大块空白），仅 ScoreChart 进度条单独保护
             如果当前页放不下"进度条+评估文字"整组，就让进度条独占下一页起，
             评估文字段落允许在卡片内部分页，最大化页面利用率 */
          .pdf__output .page .score-detail-item {
            page-break-inside: auto !important;
            break-inside: auto !important;
            margin-top: 8px !important;
          }
          .pdf__output .page .score-card-keep {
            page-break-inside: avoid !important;
            break-inside: avoid !important;
            page-break-after: avoid !important;
            break-after: avoid !important;  /* 进度条尽量与下方文字同页 */
          }
          .pdf__output .page .score-rs-bg {
            page-break-inside: auto !important;
            break-inside: auto !important;
          }
          /* 封面页后强制分页，避免后面内容拼接到封面 */
          .one-page {
            page-break-after: always !important;
            break-after: page !important;
          }
          /* 内容页 .page 左右 padding：与 baseline 对齐
             baseline 单行 ~35 字，实际表现总左右边距较窄
             chromedp MarginLeft/Right=0.4inch (1.016cm) + .page padding 1.2cm ≈ 总 2.2cm */
          .pdf__output .page {
            padding: 0 1.2cm !important;
            box-shadow: none !important;
            border: none !important;
          }
          /* 同时去掉 .one-page（封面）和 .pdf-split-page 的阴影/边框 */
          .pdf__output .one-page,
          .pdf__output .pdf-split-page,
          .pdf__output .main {
            box-shadow: none !important;
            border: none !important;
            outline: none !important;
          }
          /* 封面 .one-page 背景图限制在第 1 页，避免溢出到第 2 页
             height: 297mm + overflow:hidden 让浏览器把背景裁切在一页内
             background-size:contain 保持原图比例，不拉伸变形（与 baseline 一致）
             background-position:bottom 让树在下方（与 baseline 布局一致）*/
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
          /* FB-036: 字体兜底全局应用到 invitationLetter 整树，避免封面/标题不是仿宋 */
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
          /* FB-048: 全局兑底——.page 下所有后代默认 16px/24px 仿宋，涵盖 .o-txt/.tr-label 等元素级 font-size */
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
          /* FB-045: 补 .score-desc (高分段元素/中分段元素/低分段元素) */
          /* FB-050: 排除 .main__text.w-600（个人信息字段如姓名/性别/年龄）——客户要求不加粗 */
          .pdf__output .page .w-600:not(.main__text):not(p),
          .pdf__output .page .section-title,
          .pdf__output .page .score-desc {
            font-weight: 600 !important;
            -webkit-text-stroke: 0.4px currentColor !important;
            text-stroke: 0.4px currentColor !important;
          }
          /* FB-050: 个人信息字段强制不加粗 */
          .pdf__output .page p.main__text.w-600,
          .pdf__output .page li p.main__text {
            font-weight: 400 !important;
            -webkit-text-stroke: 0 !important;
            text-stroke: 0 !important;
          }
          .pdf__output .page .section-title {
            font-size: 21px !important;  /* FB-048: 客户要求三号 = 21px */
            line-height: 22px !important; /* FB-048: 节标题单倍行距 */
          }
          /* FB-048: 节标题所在段落段前段后各 1 行（≈21px）*/
          .pdf__output .page p.main__text:has(> span.section-title) {
            margin-top: 21px !important;
            margin-bottom: 21px !important;
            line-height: 22px !important;
          }
          /* FB-045: 分段元素标签字号也加大 */
          .pdf__output .page .score-desc {
            font-size: 16px !important;
          }
          /* FB-057: 分段 banner（高/中/低分段元素）必须与紧邻的简介段(.bg-l*)和首个进度条不分页
             避免 banner 落在上页底部、首段内容跑到下一页顶部出现"无标题悬空"现象 */
          .pdf__output .page .score-t-container {
            page-break-inside: avoid !important;
            page-break-after: avoid !important;
            break-inside: avoid !important;
            break-after: avoid !important;
          }
          .pdf__output .page .bg-lred,
          .pdf__output .page .bg-lyellow,
          .pdf__output .page .bg-lgreen {
            page-break-inside: avoid !important;
            page-break-after: avoid !important;
            break-inside: avoid !important;
            break-after: avoid !important;
          }
          /* FB-050: h3 主标题“个体报告详细解读”——黑体二号(29px)，压缩上方留白 */
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
          /* 同时压缩工具类 mt-40 在 PDF 模式的间距 */
          .pdf__output .page .mt-40 {
            margin-top: 0 !important;
          }
        }
      `
      document.head.appendChild(style)
      // 就绪检测三阶段：
      //   1) happy: isGetData=true && evaluation齐 (12) && scoreList 已渲染 → 立即 ready
      //   2) idle:  isGetData=true && evaluation 连续 5s 无变化 → 接受部分结果
      //   3) hard:  20s 硬超时→ 强制起动 (后端 ctx 45s，留 25s buffer 给 print/merge/watermark)
      // FB-035: happy 还必须检查 heightScores+midScores+lowScores 总和 >= 12 (range 已构建)
      //         避免 _tryBuildRange 还在调用中就 ready=true
      const expectedEval = 12
      const start = Date.now()
      let lastEvalLen = 0
      let lastChange = Date.now()
      const timer = setInterval(() => {
        const evalLen = Array.isArray(this.evaluation) ? this.evaluation.length : 0
        if (evalLen !== lastEvalLen) {
          lastEvalLen = evalLen
          lastChange = Date.now()
        }
        const rangeBuilt = ((this.heightScores || []).length + (this.midScores || []).length + (this.lowScores || []).length) >= 12
        const fullReady = this.isGetData && evalLen >= expectedEval && rangeBuilt
        const idleReady = this.isGetData && evalLen > 0 && (Date.now() - lastChange > 5000) && rangeBuilt
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
    // 原前端生成模式（fallback）
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
      // Handle both 'YYYY-MM-DD HH:mm:ss' and RFC3339 'YYYY-MM-DDTHH:mm:ss+08:00'
      return dt.substring(0, 10)
    },

    showField(fieldName) {
      // 封闭模式或未配置时显示全部，开放模式按勾选项显示
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

    // R8: 包装 getDicts，失败时仍 push 占位 obj，保证 evaluation.length 单调递增到 N
    // happy 路径: dict 加载成功 → push 完整 obj
    // failure 路径: dict 失败 → push {name, _failed:true} 占位，让 ready 检测能达标
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

    // 从 batch 响应填充 evaluation（happy path）
    _fillEvalFromBatch(batchData, dictName, label) {
      const rows = batchData[dictName] || []
      const obj = {name: label}
      rows.forEach(x => { obj[x.dictValue] = x.dictLabel })
      this.evaluation.push(obj)
      this._tryBuildRange()
    },

    getDictData() {
      const stu = this.stuFlag == 1 || this.r_stuFlag == 1
      const stuSuffix = stu ? '_stu' : ''
      // 全部 dict 名（14 个）
      const burnoutLabel = stu ? '倦怠感' : '职业倦怠'
      const evalDicts = [
        ['el_anxiety' + stuSuffix, '焦虑'],
        ['el_depression' + stuSuffix, '抑郁'],
        ['el_psychological_imbalance', '心理失衡'],
        ['el_enmity', '敌意'],
        ['el_fear' + stuSuffix, '恐惧'],
        ['el_feeling_unwell' + stuSuffix, '身体不适'],
        ['el_cognitive_decline', '认知衰退'],
        ['el_emotional', '情绪化'],
        ['el_frustration' + stuSuffix, '挫折感'],
        ['el_selfdoubt', '自我否定'],
        ['el_skepticism', '怀疑感'],
        ['el_burnout' + stuSuffix, burnoutLabel]
      ]
      const allTypes = ['el_health_type', 'el_traits_define', ...evalDicts.map(d => d[0])]

      // 优先走批量接口（1 次 RTT），失败降级到单个 getDicts
      getDictsBatch(allTypes).then(res => {
        const data = res.data || {}
        // health type 和 traits define
        ;(data['el_health_type'] || []).forEach(x => {
          this.healthType.push({name: x.dictValue, value: x.dictLabel})
        })
        if (data['el_traits_define']) this.traitsDefine = data['el_traits_define']
        // 12 个评估 dict
        evalDicts.forEach(([name, label]) => this._fillEvalFromBatch(data, name, label))
      }).catch(err => {
        // eslint-disable-next-line no-console
        console.warn('[pdfgen] batch dict failed, fallback per-dict', err)
        this.getDicts('el_health_type').then(res => {
          ;(res.data || []).forEach(x => this.healthType.push({name: x.dictValue, value: x.dictLabel}))
        }).catch(() => {})
        this.getDicts('el_traits_define').then(res => { this.traitsDefine = res.data }).catch(() => {})
        evalDicts.forEach(([name, label]) => this._loadEvalDict(name, label))
      })
    },

    getRangeData() {
      if(this.stuFlag == 1 || this.r_stuFlag ==1){
        this.scores['倦怠感'] = this.scores['职业倦怠']
        delete this.scores['职业倦怠']
      }
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
      if(key === '抑郁' && (this.stuFlag==1||this.r_stuFlag == 1)){
        key = key+'s'
      }
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
      // 失败仍要设 isGetData=true，让 ready 检测的 idle 兑底能触发（避免死等 hard 20s）
      const onFail = (err) => {
        // eslint-disable-next-line no-console
        console.warn('[pdfgen] fetchScore failed', err)
        this.isGetData = true
      }
      if (this.testerId.length === 18) {
        paperResult2(params).then(response => {
          this.scores = response.data
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

    // FB-035: 仅当 evaluation 字典已齐全（>=12）时才调 getRangeData
    // 否则 getEvaluation 遍历空 evaluation 返回 undefined，模板 {{ data.evaluation }} 渲染空白
    _tryBuildRange() {
      if (this._rangeBuilt) return
      const evalReady = Array.isArray(this.evaluation) && this.evaluation.length >= 12
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
            // FB-036: gender 兼容 string/number，默认"未知"避免空字段
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
    width:90%;
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
  background-image: url('~@/assets/report_images/cover04-new.png');
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


</style>

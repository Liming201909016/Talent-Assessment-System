<template>
  <div class="app-container invitationLetter">
    <el-button type="primary" @click="createPdf">导出</el-button>
    <main id="meet-pdf-id" :class="{ pdf__output: isPdf }">
      <!-- 页头页眉 -->
      <div id="pdf-header" style="height: 3.8cm" class="pdf-hf"></div>
      <div id="pdf-footer" style="height: 3.5cm" class="pdf-hf"></div>

      <!-- 导出区域 -->
      <div class="main page">
        <el-card class="whole-node">
          <div class="report-cover">
            <img :src="cover02">
          </div>
          <h1 class="main__title">心理特质测评报告</h1>
          <h2 class="main__title">Mental Health Assessment Report</h2><br><br><br><br><br>
          <span class="line-img"><img :src="cover03"></span><br><br><br><br><br>
          <h2>xxx</h2><br><br><br><br><br><br><br><br>
          <span>2024年03月01日</span>
        </el-card>

        <el-card class="whole-node">
          <div class="header1">
            <span>报告阅读说明</span>
          </div>
        </el-card>

        <el-card>
          <div class="header1">
            <span>个体报告详细解读</span>
          </div>
          <el-row class="perInfo">
            <el-row  class="whole-node"><span class="header2">一、个人信息</span></el-row>
            <el-row>
              <table cellspacing="0" style="width: 80%">
                <tbody>
                <td>
                  <tr>姓名：XXX</tr>
                  <tr>年龄：XX</tr>
                  <tr>所在单位：XXX</tr>
                  <tr>手机：XXX</tr>
                  <tr>时间：2024年03月01日10:00</tr>
                </td>
                <td>
                  <tr>性别：X</tr>
                  <tr>学历：XX</tr>
                  <tr>岗位：XXX</tr>
                  <tr>时长：XX分钟</tr>
                  <tr></tr>
                </td>
                </tbody>
              </table>
            </el-row>
            <el-row  class="whole-node"><p>时长超过35分钟需说明：在测试过程中，您的答题时间严重超出预设标准时间，因此，所得测试结果可能存在一定的误差。</p></el-row>
          </el-row>

          <el-row class="totalEval">
            <el-row  class="whole-node"><span class="header2">二、总体评价</span></el-row>
            <el-row  class="whole-node">您的心理健康水平：</el-row>
            <el-row class="whole-node">
              <TotalChart></TotalChart>
            </el-row>
            <el-row>
              <p  class="whole-node">经过对本次心理测评结果的深入分析，得出以下结论：目前您的心理状况较好，但在某些极端或特定情景下，
                可能会出现轻微的心理困扰，不过并无大碍，无需过分担忧。（注意：本报告结果参考价值受限于您作答的结果参考性情况。）</p>
              <p  class="whole-node">与职业群体进行对比分析，您目前所经历的怀疑感、身体不适、自我否定、认知衰退、敌意、挫折感和焦虑，均属于普遍存在的心理和情感反应。
                当面临工作中的挑战和压力时，这些心理状态并未显示出异常或困扰的迹象。相比之下，您在恐惧、心理失衡、抑郁、情绪化和职业倦怠等心理
                状态上表现更为突出。这些心理状态若不加以关注和处理，可能会对您的心理健康、生活品质和工作效率造成显著的负面影响。</p>
            </el-row>
            <el-row class="whole-node">
              <RadarChart></RadarChart>
            </el-row>
            <el-row>
              <p  class="whole-node">图示说明：</p>
              <p  class="whole-node">1、常模标准分的参照分数为5.5；</p>
              <p  class="whole-node">2、个人各项得分均以标准十分制呈现，是在参考群体常模的基础上，将受测者在12个因素上原始得分按照转化规则转化后得到的分数。</p>
            </el-row>
          </el-row>
          <el-row class="compEval">
            <el-row  class="whole-node"><span class="header2">三、综合评价报告</span></el-row>
            <el-row>
              <p  class="whole-node">本报告以您的个人测试得分为依据，从心理状况的高、中、低分段因素出发，进行了详尽的分析与评估，力求全面、客观地揭示您的心理状态。
                具体来说，低分段得分表明您的心理健康状况良好；而得分逐渐升高，进入高分段，则反映出您面临的心理困扰愈发严峻。</p>
            </el-row>
            <el-row class="whole-node">
              <BarChart></BarChart>
            </el-row>
            <el-row  class="whole-node">图示说明：</el-row>
            <el-row  class="whole-node">
              <p>根据常模标准，每个因素若转化后的分数高于7分时，归为高分段因素，以橙色部分呈现，表示心理困扰程度相对比较重。相比之下，
                分数位于4.5分至7分之间，归为中分段因素，以橙色部分呈现，表示心理困扰程度比较轻。另外，分数低于4.5分时，归为低分段因素，
                以绿色部分呈现，表示心理状态良好。</p>
            </el-row>
          </el-row>
          <el-row class="layerEval">
            <el-row  class="whole-node"><span class="header2">四、分层评价报告</span></el-row>
            <el-row>
              <el-row  class="whole-node" type="flex" align="middle" justify="center" style="margin-top: 20px"><img :src="heightImg"></el-row>
              <el-row  class="whole-node">
                <p>根据您的作答结果显示，您在恐惧、心理失衡和抑郁因素的得分较高，反映出您的这些心理状态不够理想。
                  考虑到这些因素可能对您的心理健康产生不良影响，您应当给予高度重视，并采取积极有效的措施进行改善。</p>
              </el-row>
              <el-row class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row  class="whole-node">【个人得分】8.00分</el-row>
              <el-row  class="whole-node">【表现评估】安全感一般，有时会担心出现意外和工作失误。</el-row>
              <el-row  class="whole-node">【概念阐释】恐惧是指个体由于担忧和害怕某些事情或人而产生的显著苦恼情绪。</el-row>
              <el-row  class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row  class="whole-node">【个人得分】7.30分</el-row>
              <el-row  class="whole-node">【表现评估】基本能够接受现状，但当主观愿望与客观现实之间存在较大差距时，会出现一些怨气，觉得自己付出多而得到少。</el-row>
              <el-row  class="whole-node">【概念阐释】心理失衡是指个体感觉到自己的所得与付出不成比例而产生的不满与怨愤。</el-row>
              <el-row class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row  class="whole-node">【个人得分】7.01分</el-row>
              <el-row  class="whole-node">【表现评估】通常能保持平静，有时可能因工作和生活中的压力感到不安，产生焦躁情绪。</el-row>
              <el-row class="whole-node">【概念阐释】抑郁是以心境低落为主要特征，个体丧失热忱和乐趣，精力丧失，自我评价过低，消极悲观。</el-row>
            </el-row>

            <el-row>
              <el-row class="whole-node" type="flex" align="middle" justify="center" style="margin-top: 20px"><img :src="midImg"></el-row>
              <el-row class="whole-node">
                <p>根据您的作答结果显示，您在恐惧、心理失衡和抑郁因素的得分较高，反映出您的这些心理状态不够理想。
                  考虑到这些因素可能对您的心理健康产生不良影响，您应当给予高度重视，并采取积极有效的措施进行改善。</p>
              </el-row>
              <el-row class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row class="whole-node">【个人得分】8.00分</el-row>
              <el-row class="whole-node">【表现评估】安全感一般，有时会担心出现意外和工作失误。</el-row>
              <el-row class="whole-node">【概念阐释】恐惧是指个体由于担忧和害怕某些事情或人而产生的显著苦恼情绪。</el-row>
              <el-row class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row class="whole-node">【个人得分】7.30分</el-row>
              <el-row class="whole-node">【表现评估】基本能够接受现状，但当主观愿望与客观现实之间存在较大差距时，会出现一些怨气，觉得自己付出多而得到少。</el-row>
              <el-row class="whole-node">【概念阐释】心理失衡是指个体感觉到自己的所得与付出不成比例而产生的不满与怨愤。</el-row>
              <el-row class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row class="whole-node">【个人得分】7.01分</el-row>
              <el-row class="whole-node">【表现评估】通常能保持平静，有时可能因工作和生活中的压力感到不安，产生焦躁情绪。</el-row>
              <el-row class="whole-node">【概念阐释】抑郁是以心境低落为主要特征，个体丧失热忱和乐趣，精力丧失，自我评价过低，消极悲观。</el-row>
            </el-row>

            <el-row>
              <el-row class="whole-node" type="flex" align="middle" justify="center" style="margin-top: 20px"><img :src="lowImg"></el-row>
              <el-row class="whole-node">
                <p>根据您的作答结果显示，您在恐惧、心理失衡和抑郁因素的得分较高，反映出您的这些心理状态不够理想。
                  考虑到这些因素可能对您的心理健康产生不良影响，您应当给予高度重视，并采取积极有效的措施进行改善。</p>
              </el-row>
              <el-row class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row class="whole-node">【个人得分】8.00分</el-row>
              <el-row class="whole-node">【表现评估】安全感一般，有时会担心出现意外和工作失误。</el-row>
              <el-row class="whole-node">【概念阐释】恐惧是指个体由于担忧和害怕某些事情或人而产生的显著苦恼情绪。</el-row>
              <el-row class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row class="whole-node">【个人得分】7.30分</el-row>
              <el-row class="whole-node">【表现评估】基本能够接受现状，但当主观愿望与客观现实之间存在较大差距时，会出现一些怨气，觉得自己付出多而得到少。</el-row>
              <el-row class="whole-node">【概念阐释】心理失衡是指个体感觉到自己的所得与付出不成比例而产生的不满与怨愤。</el-row>
              <el-row class="whole-node">
                <ScoreChart></ScoreChart>
              </el-row>
              <el-row class="whole-node">【个人得分】7.01分</el-row>
              <el-row class="whole-node">【表现评估】通常能保持平静，有时可能因工作和生活中的压力感到不安，产生焦躁情绪。</el-row>
              <el-row class="whole-node">【概念阐释】抑郁是以心境低落为主要特征，个体丧失热忱和乐趣，精力丧失，自我评价过低，消极悲观。</el-row>
            </el-row>

          </el-row>

          <!--        <div class="el-table el-table&#45;&#45;enable-row-hover el-table&#45;&#45;medium">-->
          <!--          <div ref="usedmemory" style="height: 420px" />-->
          <!--        </div>-->
        </el-card>

        <el-card>
          <div class="header1">
            <span>团体报告详细解读</span>
          </div>
          <div class="el-table el-table--enable-row-hover el-table--medium">
            <div ref="usedmemory" style="height: 420px" />
          </div>
        </el-card>
      </div>

    </main>

  </div>
</template>

<script>

import { paperResult } from '@/api/paper/exam'
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
import htmlToPdf from "@/utils/htmlToPdf";
import html2Pdf from "@/utils/jspdf"

export default {
  components: {CircleChart, TotalChart, RadarChart, BarChart, ScoreChart},
  data() {
    return {

      cover02: cover02,
      cover03: cover03,
      heightImg: heightScore,
      midImg: midScore,
      lowImg: lowScore
    }
  },
  created() {

  },
  methods: {



    outPutPdfFn () {
      let vm = this;
      const A4_WIDTH = 592.28;
      const A4_HEIGHT = 841.89;
      // $myLoading 自定义等待动画组件，实现导出事件的异步等待交互
      // this.$myLoading('正在导出pdf，请稍候。。。', true);
      vm.$nextTick(() => {
        // dom的id。
        let target = document.getElementById('content');
        // let target = this.$refs.targetDom
        let pageHeight = target.scrollWidth / A4_WIDTH * A4_HEIGHT;
        // let pageHeight = target.width / A4_WIDTH * A4_HEIGHT;
        // 获取分割dom，此处为class类名为item的dom
        let lableListID = document.getElementsByClassName('whole-node');
        // 进行分割操作，当dom内容已超出a4的高度，则将该dom前插入一个空dom，把他挤下去，分割
        for (let i = 0; i < lableListID.length; i++) {
          let multiple = Math.ceil((lableListID[i].offsetTop + lableListID[i].offsetHeight) / pageHeight);
          if (this.isSplit(lableListID, i, multiple * pageHeight)) {
            let divParent = lableListID[i].parentNode; // 获取该div的父节点
            let newNode = document.createElement('div');
            newNode.className = 'emptyDiv';
            newNode.style.background = '#ffffff';
            let _H = multiple * pageHeight - (lableListID[i].offsetTop + lableListID[i].offsetHeight);
            newNode.style.height = _H + 30 + 'px';
            newNode.style.width = '100%';
            let next = lableListID[i].nextSibling; // 获取div的下一个兄弟节点
            // 判断兄弟节点是否存在
            console.log(next);
            if (next) {
              // 存在则将新节点插入到div的下一个兄弟节点之前，即div之后
              divParent.insertBefore(newNode, next);
            } else {
              // 不存在则直接添加到最后,appendChild默认添加到divParent的最后
              divParent.appendChild(newNode);
            }
          }
        }
        // 传入title和dom标签，此处是 #content
        // 异步函数，导出成功后处理交互
        this.getPdf('巡检报告单-', '#content').then(() => {
          // 自定义等待动画关闭
          // this.$myLoading(false);
          this.$message({
            type: 'success',
            message: '导出成功'
          });
          this.detailSHow = false;
        })
          .catch(() => {
            // this.$myLoading(false);
            this.$message({
              type: 'error',
              message: '导出失败，请重试'
            });
          });
      });
    },

    isSplit (nodes, index, pageHeight) {
      // 计算当前这块dom是否跨越了a4大小，以此分割
      if (nodes[index].offsetTop + nodes[index].offsetHeight < pageHeight && nodes[index + 1]
        && nodes[index + 1].offsetTop + nodes[index + 1].offsetHeight > pageHeight) {
        return true;
      }
      return false;
    },

  }
}
</script>

<style scoped>

  .header1{
    text-align: center;
    font-size: 32px;
    margin-top: 0;
  }

  .header2{
    float: left;
    text-align: left;
    font-size: 22px;
  }

  .report-cover{
    display: flex;
    flex-direction: column;
    background-image: url('~@/assets/report_images/cover01.png');
    background-repeat: no-repeat;
    //background-position: center 100px;

  //background-size: 100%;
    background-size: cover;
    height: 50vh;
    //width: 100vh;
    background-color: #f0f2f5;
    overflow: hidden;
    align-items: center;
    justify-content: center;
  }

  h1{
    text-align: center;
    font-size: 42px;
  }

  h2{
    text-align: center;
    font-size: 32px;
    margin-top: 0;
  }

  span{
    display: flex;
    justify-content: center;
    align-items: center;
  }


</style>


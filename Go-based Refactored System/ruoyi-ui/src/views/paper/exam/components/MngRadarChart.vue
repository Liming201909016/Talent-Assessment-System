<template>
  <div style="text-align:center;">
    <!-- FB-042: 加 mng-radar-chart-inner class + 固定 600px + margin auto 居中 -->
    <div ref="chart" class="pdf-group-item mng-radar-chart-inner" style="width: 600px; height: 450px; margin: 0 auto;"></div>
  </div>
</template>

<script>

import * as echarts from 'echarts';
export default {
  name: "RadarChart",

  props: ['scores'],

  data(){
    return{

    }
  },

  mounted() {

    let myChart = echarts.init(this.$refs.chart, null, { devicePixelRatio: 2 });

    let textList = ['得分']; // 线条类型
    let nameList = [
      {name: '社会性', max: 5},
      {name: '进取性', max: 5},
      {name: '领导性', max: 5},
      {name: '计划性', max: 5},
      {name: '人际敏感性', max: 5},
      {name: '自信心', max: 5},
      {name: '责任心', max: 5},
      {name: '学习力', max: 5},
      {name: '创新性', max: 5},
      {name: '情绪稳定性', max: 5},
      {name: '自律性', max: 5},
      {name: '决断性', max: 5},
      {name: '合作性', max: 5},
    ]; // 选项标题
    let uList = []; // 你的得分

    // if (this.scores !== null && this.scores !== undefined) {
    //   nameList.forEach(c => {
    //     uList.push(this.scores[c.name])
    //   })
    // }

    nameList.forEach((c) => {
      uList.push(this.scores[c.name])
      // console.log(c.name + "====" + uList)
    })

    // // 数据
    // let u = '';
    // this.data.forEach((c) => {
    //   u += c.Value + ',';
    // });
    // uList = u.substring(0, u.length - 1).split(',');

    let option = {
      title: {
        show:false,
      },
      legend: {
        show:false,
      },
      radar: {
        // radius: 100,
        // 设置雷达图中间射线的颜色
        axisLine: {
          lineStyle: {
            color: '#DFDFDF',
          },
          show: true
        },

        splitLine: {
          show: true,
          lineStyle: {
            color: '#DFDFDF'
          }
        },
        indicator: nameList,
        name:{ //修改indicator文字的颜色
          textStyle:{
            color:'#000000',
            fontSize: 14
          }
        },
        // 雷达图背景的颜色，在这儿随便设置了一个颜色，完全不透明度为0，就实现了透明背景
        splitArea: {
          areaStyle: {
            // color: '#ffffff' // 图表背景的颜色
            color: ['#ffffff','rgba(210,219,238,0.2)'],
          }
        }
      },
      series: [
        {
          type: 'radar',
          data: [
            {
              value: uList,
              name: '得分',
              // 设置区域边框和区域的颜色
              itemStyle: {
                color: '#1E90FF',
                borderColor: '#1E90FF',       // 拐点的描边颜色。[ default: '#000' ]
                borderWidth: 4,
              },
              lineStyle: {
                width: 2
              },
              label: {
                show: true,
                fontSize: 14,
                color: '#313131',
                formatter: function (params) {
                  return params.value;
                }
              },
              labelLayout: {
                // 移动标签，使其远离数据点中心
                // dx, dy 可以是固定的像素值，也可以是回调函数
                dx: 20,
                dy: 10
              },
            }
          ],
          areaStyle: {
            color: 'rgba(30,144,255,0.2)' // 数据区域的填充色，和背景区分开
          }
        }
      ]
    };

    myChart.setOption(option);
  }
}
</script>

<style scoped lang="scss">

</style>

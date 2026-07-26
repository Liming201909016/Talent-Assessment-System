<template>
  <div style="text-align:center;">
    <div ref="chart" class="radar-chart-inner pdf-group-item" style="width: 100%; height: 450px;"></div>
  </div>
</template>

<script>

import * as echarts from 'echarts';
export default {
  name: "RadarChart",

  props: ['scores','stuFlag'],

  data(){
    return{

    }
  },

  mounted() {

    let myChart = echarts.init(this.$refs.chart, null, { devicePixelRatio: 2 });

    let textList = ['标准分', '个人得分']; // 线条类型
    let nameList = [
      {name: '焦虑', max: 12},
      {name: '抑郁', max: 12},
      {name: '心理失衡', max: 12},
      {name: '敌意', max: 12},
      {name: '恐惧', max: 12},
      {name: '身体不适', max: 12},
      {name: '认知衰退', max: 12},
      {name: '情绪化', max: 12},
      {name: '挫折感', max: 12},
      {name: '自我否定', max: 12},
      {name: '怀疑感', max: 12},
      {name: this.stuFlag?'倦怠感':'职业倦怠', max: 12},
    ]; // 选项标题
    let standList = ['5.5', '5.5', '5.5', '5.5','5.5', '5.5', '5.5', '5.5', '5.5','5.5', '5.5','5.5']; // 平均线
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
        text: '总体评价',
        left: 'center',
        top :'20px'
      },
      legend: {
        orient: 'vertical',
        data: textList,
        left: 'left',
        top: '60px',
        itemGap: 25,
        textStyle: {
          // 图例文字的样式
          fontSize: 14,
          color:'#000000'
        }
      },
      radar: {
        center: ['50%', '60%'],  // 第二个值从默认50%调整为60%，向下移动
        // radius: 100,
        // 设置雷达图中间射线的颜色
        axisLine: {
          lineStyle: {
            color: '#999',
          },
          show: true
        },

        splitLine: {
          show: true,
          lineStyle: {
            color: '#c8e5b3'
          }
        },
        indicator: nameList,
        name:{ //修改indicator文字的颜色
          textStyle:{
            color:"#999"
          }
        },
        // 雷达图背景的颜色，在这儿随便设置了一个颜色，完全不透明度为0，就实现了透明背景
        splitArea: {
          areaStyle: {
            color: '#ffffff' // 图表背景的颜色
          }
        }
      },
      series: [
        {
          type: 'radar',
          data: [
            {
              value: standList,
              name: '标准分',
              // 设置区域边框和区域的颜色
              itemStyle: {
                color: '#5B8FF9',
                borderColor: '#5B8FF9',       // 拐点的描边颜色。[ default: '#000' ]
                borderWidth: 4,
              },
              lineStyle: {
                width: 2
              },
            },
            {
              value: uList,
              name: '个人得分',
              // 设置区域边框和区域的颜色
              itemStyle: {
                color: '#5AD8A6',
                borderColor: '#5AD8A6',       // 拐点的描边颜色。[ default: '#000' ]
                borderWidth: 4,
              },
              lineStyle: {
                width: 2
              },
              label: {
                show: true,
                fontSize: 16,
                color: '#999',
                formatter: function (params) {
                  return params.value;
                }
              }
            }
          ]
        }
      ]
    };

    myChart.setOption(option);
  }
}
</script>

<style scoped lang="scss">

</style>

<template>
  <div style="text-align:center;">
    <div ref="chart" class="mng-total-chart-inner pdf-group-item" style="width: 100%; height: 350px;"></div>
  </div>
</template>

<script>

import * as echarts from 'echarts';
export default {
  name: "MngTotalChart",

  props: ['scores'],

  data(){
    return{

    }
  },

  mounted() {

    let myChart = echarts.init(this.$refs.chart, null, { devicePixelRatio: 2 });
    /*
    // 获取图表容器尺寸
    let chartWidth = myChart.getWidth();
    let chartHeight = myChart.getHeight();

    // 仪表盘中心点（与配置一致）
    let centerX = chartWidth * 0.5;
    let centerY = chartHeight * 0.5;

    // 仪表盘半径（根据你的radius配置）
    let radius = Math.min(chartWidth, chartHeight) * 0.9 * 0.5;

// 要计算的刻度值
    let targetValue = this.scores;
    let minValue = 0;
    let maxValue = 65;
    let startAngle = 180; // 度
    let endAngle = 0;      // 度

    // 计算对应的角度
    let angleRange = startAngle - endAngle;
    let valueRatio = (targetValue - minValue) / (maxValue - minValue);
    let angle = startAngle - valueRatio * angleRange;

    // 角度转弧度
    let rad = angle * Math.PI / 180;
    // 计算坐标（刻度线位于圆周上）
    let tickX = centerX + (radius - 20) * Math.cos(rad);
    let tickY = centerY - radius * Math.sin(rad); // Y轴向下为正，所以减号

    console.log(`刻度 ${targetValue} 的坐标: (${tickX}, ${tickY})`);
    */
    let option = {
      series: [
        {
          type: 'gauge',
          startAngle: 180,
          endAngle: 0,
          min: 0,
          max: 65,
          radius:'90%',
          itemStyle: {
            color: '#1E90FF',
            shadowColor: 'rgba(0,138,255,0.45)',
            shadowBlur: 10,
            shadowOffsetX: 2,
            shadowOffsetY: 2
          },
          pointer: {
            length: '80%',
            width: 16,
            offsetCenter: [0, '5%']
          },
          axisLine: {
            roundCap:true,
            lineStyle: {
              width: 20,
              color: [
                [this.scores/65, '#1E90FF'],
                [1, '#e5e5e5']
              ],
            }
          },
          axisTick: {
            show:false
          },
          splitLine: {
            show:false,
          },
          axisLabel: {
            show:false
          },
          title: {
            show: false
          },
          detail: {
            show:false,
          },
          data: [
            {
              value: this.scores
            }
          ]
        }
      ],
      graphic: [
        {
          type: 'circle',
          left: 141,
          top: 165,
          shape: { r: 10 },
          style: { fill: '#1E90FF' },
          z: 10
        },
        {
          type: 'circle',
          right: 141,
          top: 165,
          shape: { r: 10 },
          style: { fill: '#e5e5e5' },
          z: 10
        },
        {
          type: 'circle',
          left: 'center',
          top: 60,
          shape: { r: 100 },
          style: {
            fill: new echarts.graphic.LinearGradient(0.5, 0, 0.5, 1, [
              { offset: 0, color: 'rgba(30, 58, 95, 0.05)' },
              { offset: 1, color: 'rgba(255, 255, 255, 0)' }
            ]),
            stroke: '#fff',
            lineWidth: 0
          },
          z: 5
        },
        {
          type: 'circle',
          left: 'center',
          top: 110,
          shape: { r: 50 },
          style: {
            fill: '#fff',
            stroke: '#fff',
            lineWidth: 0,
            // 添加阴影
            shadowBlur: 12,
            shadowColor: 'rgba(200, 200, 200, 0.5)',  // 浅灰色半透明
            shadowOffsetX: 1,
            shadowOffsetY: 1,
          },
          z: 10
        },
        {
          type: 'text',
          left: 'center',
          top: 130,
          style: {
            text: '得分',             // 初始数值
            fill: 'rgb(108,108,108)',
            // font:  '16px Arial',
            fontSize: 16,
            textAlign: 'center',
            textVerticalAlign: 'middle'
          },
          z: 15                     // 确保文字在圆形上方
        },
        {
          type: 'text',
          left: 'center',
          top: 155,
          style: {
            text: this.scores,             // 初始数值
            fill: 'black',
            font: 'bold 25px Arial',
            textAlign: 'center',
            textVerticalAlign: 'middle'
          },
          z: 20                     // 确保文字在圆形上方
        },
       /* {
          type: 'circle',
          // left: tickX - (angle < 90 ? 20 : 0),
          left: tickX,
          // left: tickX - (angle === 90 ? 10 : (angle < 90 ? 15 : 0)),
          // top: tickY - (angle === 90 ? 0 : (angle < 90 ? 0 : 0)),
          top: tickY - 5,
          shape: { r: 10 },
          style: { fill: 'red' },
          z: 100
        }*/
      ]
    };

    myChart.setOption(option);
  }
}
</script>

<style scoped lang="scss">

</style>

<template>
  <div>
    <div ref="chart" class="pdf-group-item" style="width: 100%; height: 400px;"></div>
  </div>
</template>

<script>

import * as echarts from 'echarts';
import el from "element-ui/src/locale/lang/el";
import th from "element-ui/src/locale/lang/th";

export default {

  name: 'LineChart',
  props: ['teamHealth'],

  data() {
    return{

    }
  },

  mounted() {
    this.initCharts();
  },
  methods: {
    initCharts() {

      let name = [];
      let score = [];
      let teamHealth = this.teamHealth
      teamHealth.forEach(x => {
        name.push(x.name)
        score.push(x.value)
      })

      let charts = echarts.init(this.$refs.chart);
      let option = {
        xAxis: {
          type: 'category',
          data: name,
          splitLine: {
            show: true
          },
          axisTick: {
            show: false
          },
          axisLabel: {
            // fontSize: 14
          }
        },
        yAxis: {
          type: 'value',
          max: 30,
          interval: 2,
          axisTick: {
            show: false
          }
        },
        // grid: {
        //   left: '3%',
        //   right: '4%',
        //   bottom: '3%',
        //   containLabel: true,
        //   borderWidth: 1,
        //   borderColor: 'red'
        // },
        series: [{
          data: score,
          type: 'line',
          smooth: true,
          color: '#4874cb',
          label: {
            show: true,
            position: 'top',
            fontSize: 16,
            color: '#000000'
          },
          symbol: 'circle',
          symbolSize: 1,
          // itemStyle: {
          //   color: '#4874cb'
          // },
          lineStyle: {
            color: '#4874cb',
            shadowBlur: 4,
            shadowOffsetX: 4, // 折线的X偏移
            shadowOffsetY: 4, // 折线的Y偏移
            shadowColor: '#cfcfcf', //折线颜色
            width: 4
          }
        }]
      }

      charts.setOption(option);
    },
  },
};
</script>

<style lang="scss" scoped></style>


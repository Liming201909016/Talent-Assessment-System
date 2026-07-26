<template>
  <div>
    <div ref="chart" style="width: 100%; height: 100%;"></div>
  </div>
</template>

<script>

import * as echarts from 'echarts';
import {listTester} from "@/api/tester/tester";
import da from "element-ui/src/locale/lang/da";

export default {

  name: 'PieChart',
  props: ['examId'],

  data() {
    return{

      total: 0,
      // pieData: [{name: "测评", value: 7, itemStyle: {color:'#bd0404'}}],
      pieData: [],

      queryParams: {
        pageNum: 1,
        pageSize: 20,
        examId: undefined
      },
    }
  },

  created() {
    this.queryParams.examId = this.examId
  },

  mounted() {
    this.fetchData()
    // this.initCharts();
  },

  watch: {
    pieData: {
      handler: function (newVal, oldVal) {
        for (let i = 0; i < newVal.length; i++) {
          if (oldVal[i].value !== newVal[i].value) {
            return newVal;
          }
        }
      },
      deep: true
    }
  },

  methods: {

    fetchData() {
      listTester(this.queryParams).then(response => {
        this.testerList = response.rows;
        this.total = response.total;
        let dataList = this.testerList
        // console.log(response)
        if (dataList.length !== 0) {
          let testing = 0;
          let tested = 0;
          dataList.forEach(data => {
            if (data.paperId.length > 5 && data.endTime !== null) {
              tested++;
            } else if (data.paperId.length > 5 && data.endTime === null) {
              testing++;
            }
          })

          this.pieData.push({value:tested, name:'已测评', itemStyle: {color:'#77dc60'}})
          this.pieData.push({value:testing, name:'测评中', itemStyle: {color:'#efce2e'}})

          // this.pieData.push({value:tested, name:'已测评'})
          // this.pieData.push({value:testing, name:'测评中'})
        }

        this.initCharts()
      });
    },

    initCharts() {

      let pieData = this.pieData
      // console.log(pieData)

      let charts = echarts.init(this.$refs.chart);
      let option = {
        // tooltip: {
        //   trigger: 'item'
        // },
        // series : [
        //   {
        //     type: 'pie',
        //     radius: '100%',
        //     data: pieData,
        //     label: {
        //       normal: {
        //         show: true,
        //         position: 'inner', // 数值显示在内部
        //         formatter: '{c}', // 格式化数值百分比输出
        //         fontSize: 20,
        //         color: '#000000FF'
        //       },
        //     },
        //   }
        // ]
        tooltip: {
          trigger: "item",
          formatter: "{a} <br/>{b} : {c} ({d}%)",
        },
        series: [
          {
            name: "命令",
            type: "pie",
            roseType: "radius",
            radius: [15, 95],
            center: ["50%", "38%"],
            data: pieData,
            animationEasing: "cubicInOut",
            animationDuration: 1000,
          },
        ],
      };
      // console.log(option)
      charts.setOption(option);
    },
  },
};
</script>

<style lang="scss" scoped></style>


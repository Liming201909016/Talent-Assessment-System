<template>
  <div class="app-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <div slot="header">
            <span style="font-weight: bold; font-size: 16px;">{{ title }} — 测评统计</span>
            <span v-if="sState" style="margin-left: 16px; color: #999;">状态：{{ sState }}</span>
          </div>
          <div v-loading="loading" element-loading-text="正在加载统计数据，请稍候...">
            <div v-if="noData && !loading" style="text-align:center; padding: 100px 0; color: #999; font-size: 16px;">暂无测评数据</div>
            <div v-show="!noData" ref="chart" style="height: 420px" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import echarts from "echarts";
import { getListTester, listTester } from "@/api/tester/tester";

export default {
  name: "Statistics",
  data() {
    return {
      title: '',
      sState: '',
      isOpen: 0,
      loading: false,
      noData: false,
      queryParams: {
        pageNum: 1,
        pageSize: 1000,
        examId: undefined
      }
    };
  },
  created() {
    this.queryParams.examId = this.$route.params.examId
    this.title = this.$route.params.title
    this.isOpen = parseInt(this.$route.params.isOpen) || 0

    const state = parseInt(this.$route.params.state)
    if (state === 0) this.sState = '进行中'
    else if (state === 2) this.sState = '尚未开放'
    else if (state === 3) this.sState = '已结束'
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    fetchData() {
      this.loading = true
      console.log('statistics fetchData, isOpen:', this.isOpen, 'examId:', this.queryParams.examId)
      const apiCall = this.isOpen === 1
        ? listTester(this.queryParams)
        : getListTester(this.queryParams)

      apiCall.then(response => {
        console.log('statistics response:', response)
        this.loading = false
        const dataList = response.rows || []
        const total = response.total || dataList.length

        if (total === 0) {
          this.noData = true
          return
        }

        let testing = 0
        let tested = 0
        dataList.forEach(data => {
          if (data.paperId != null && data.endTime != null) {
            tested++
          } else if (data.paperId != null && data.endTime == null) {
            testing++
          }
        })

        const notTest = Math.max(0, total - testing - tested)
        const pieData = [
          { value: tested, name: '已测评', itemStyle: { color: '#77dc60' } },
          { value: testing, name: '测评中', itemStyle: { color: '#efce2e' } },
          { value: notTest, name: '未测评', itemStyle: { color: '#bebaba' } }
        ]
        this.$nextTick(() => {
          this.initChart(pieData)
        })
      }).catch(err => {
        console.error('statistics error:', err)
        this.loading = false
        this.noData = true
      })
    },

    initChart(pieData) {
      const chart = echarts.init(this.$refs.chart)
      chart.setOption({
        tooltip: {
          trigger: 'item',
          formatter: '{a} <br/>{b} : {c}人 ({d}%)'
        },
        series: [{
          name: '测评状态',
          type: 'pie',
          roseType: 'radius',
          radius: 95,
          center: ['50%', '38%'],
          data: pieData,
          animationEasing: 'cubicInOut',
          animationDuration: 1000,
          label: {
            show: true,
            formatter: '{b}: {c}人',
            fontSize: 20,
            color: '#000'
          }
        }]
      })
    }
  }
}
</script>

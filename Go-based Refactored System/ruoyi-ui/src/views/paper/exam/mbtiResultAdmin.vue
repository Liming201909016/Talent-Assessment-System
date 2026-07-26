<template>
  <div v-loading="loading" style="padding: 20px;">
    <div v-if="mbtiData" style="text-align: center;">
      <h2 style="color: #409eff;">职业性格测验报告</h2>

      <!-- 个人信息 -->
      <div style="text-align: left; max-width: 600px; margin: 20px auto; background: #f5f7fa; padding: 20px; border-radius: 8px;">
        <h3>个人信息</h3>
        <p v-if="testerInfo.name"><strong>姓名：</strong>{{ testerInfo.name }}</p>
        <p v-if="testerInfo.age"><strong>年龄：</strong>{{ testerInfo.age }}岁</p>
        <p v-if="testerInfo.gender !== undefined"><strong>性别：</strong>{{ testerInfo.gender === '0' || testerInfo.gender === 0 ? '男' : '女' }}</p>
        <p v-if="testerInfo.telephone"><strong>手机：</strong>{{ testerInfo.telephone }}</p>
        <p v-if="testerInfo.affiliation"><strong>单位：</strong>{{ testerInfo.affiliation }}</p>
        <p v-if="testerInfo.post"><strong>岗位：</strong>{{ testerInfo.post }}</p>
        <p v-if="testerInfo.userTime"><strong>用时：</strong>{{ testerInfo.userTime }}分钟</p>
      </div>

      <!-- MBTI 类型 -->
      <div style="margin: 30px 0;">
        <div style="display: inline-block; background: linear-gradient(135deg, #409eff, #67c23a); color: white; font-size: 48px; font-weight: bold; padding: 20px 40px; border-radius: 16px; letter-spacing: 8px;">
          {{ mbtiData.type }}
        </div>
        <div style="font-size: 20px; color: #606266; margin-top: 10px;">{{ typeNames[mbtiData.type] || '' }}</div>
      </div>

      <!-- 8维度得分 -->
      <div style="max-width: 600px; margin: 30px auto;">
        <h3 style="text-align: left;">四维度得分</h3>
        <div v-for="dim in dimensions" :key="dim.left" style="margin-bottom: 16px;">
          <div style="display: flex; align-items: center; margin-bottom: 4px;">
            <span style="width: 80px; text-align: right; font-weight: bold;" :style="{color: mbtiData.scores[dim.left] >= mbtiData.scores[dim.right] ? '#409eff' : '#999'}">
              {{ dim.leftLabel }} {{ dim.left }}
            </span>
            <div style="flex: 1; margin: 0 10px; display: flex; height: 24px; border-radius: 12px; overflow: hidden; background: #eee;">
              <div :style="{width: barWidth(dim.left, dim.right) + '%', background: '#409eff', borderRadius: '12px 0 0 12px', transition: 'width 0.5s'}" />
              <div :style="{width: (100 - barWidth(dim.left, dim.right)) + '%', background: '#e6a23c', borderRadius: '0 12px 12px 0', transition: 'width 0.5s'}" />
            </div>
            <span style="width: 80px; font-weight: bold;" :style="{color: mbtiData.scores[dim.right] > mbtiData.scores[dim.left] ? '#e6a23c' : '#999'}">
              {{ dim.right }} {{ dim.rightLabel }}
            </span>
          </div>
          <div style="display: flex; justify-content: space-between; font-size: 12px; color: #909399; padding: 0 90px;">
            <span>{{ mbtiData.scores[dim.left] }}分</span>
            <span>{{ dim.dimName }}</span>
            <span>{{ mbtiData.scores[dim.right] }}分</span>
          </div>
        </div>
      </div>
    </div>
    <div v-else-if="!loading" style="text-align: center; color: #999; padding: 40px;">
      暂无MBTI测评数据
    </div>
  </div>
</template>

<script>
import { post } from '@/utils/request'

export default {
  name: 'MbtiResultAdmin',
  props: {
    disPdfDownload: { type: Boolean, default: false },
    stuFlag: { type: Number, default: 0 },
    repoCode: { type: String, default: '' }
  },
  data() {
    return {
      loading: false,
      mbtiData: null,
      testerInfo: {},
      typeNames: {
        'ISTJ': '检查员', 'ISFJ': '守护者', 'INFJ': '提倡者', 'INTJ': '建筑师',
        'ISTP': '鉴赏家', 'ISFP': '探险家', 'INFP': '调停者', 'INTP': '逻辑学家',
        'ESTP': '企业家', 'ESFP': '表演者', 'ENFP': '竞选者', 'ENTP': '辩论家',
        'ESTJ': '总经理', 'ESFJ': '执政官', 'ENFJ': '外交家', 'ENTJ': '指挥官'
      },
      dimensions: [
        { left: 'E', right: 'I', leftLabel: '外向', rightLabel: '内向', dimName: '注意力' },
        { left: 'S', right: 'N', leftLabel: '感觉', rightLabel: '直觉', dimName: '认知方式' },
        { left: 'T', right: 'F', leftLabel: '理性', rightLabel: '感性', dimName: '判断方式' },
        { left: 'J', right: 'P', leftLabel: '判断', rightLabel: '感知', dimName: '生活方式' }
      ]
    }
  },
  methods: {
    barWidth(left, right) {
      const l = this.mbtiData?.scores?.[left] || 0
      const r = this.mbtiData?.scores?.[right] || 0
      const total = l + r
      return total > 0 ? Math.round(l / total * 100) : 50
    },
    // 兼容 viewPdf 调用链：fetchTester + fetchScore + pdfDownload
    fetchTester(paperId) {
      // 从 tester-list 获取 tester 信息
      this.paperId = paperId
    },
    fetchScore(paperId) {
      this.loading = true
      post('/exam/api/mbti/score', { paperId }).then(res => {
        if (res.code === 0 && res.data) {
          this.mbtiData = res.data
        }
      }).finally(() => { this.loading = false })
    },
    pdfDownload() {
      // MBTI 暂不支持 PDF 下载，空实现
    },
    fetchData(paperId, testerInfo) {
      this.testerInfo = testerInfo || {}
      this.fetchScore(paperId)
    }
  }
}
</script>

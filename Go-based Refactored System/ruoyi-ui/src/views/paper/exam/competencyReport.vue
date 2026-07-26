<template>
  <div class="competency-report" v-if="data">
    <el-alert class="temporary-alert" :title="data.reportTextMessage" type="error" :closable="false" show-icon />
    <section class="cover">
      <div class="temporary-mark">临时测试报告</div>
      <h1>胜任力测评报告</h1>
      <p>报告版本：{{ audienceLabel }}</p>
      <p>内容版本：{{ data.contentVersion }}</p>
      <p>生成日期：{{ reportDate }}</p>
    </section>
    <section>
      <h2>个人信息</h2>
      <div class="person-grid">
        <span>姓名：{{ data.result.participantName || '—' }}</span>
        <span>手机号：{{ data.result.participantTelephone || '—' }}</span>
        <span>单位：{{ data.result.participantAffiliation || '—' }}</span>
        <span>岗位：{{ data.result.participantPost || '—' }}</span>
      </div>
    </section>
    <section>
      <h2>报告阅读说明</h2>
      <p>题目采用五级量表，正反向题统一换算为 1–5 分。维度得分为题目最终分平均值，整体得分为各维度得分之和。</p>
      <p class="temporary-copy">{{ data.reportTextMessage }}</p>
    </section>
    <section>
      <h2>总体评价</h2>
      <div class="score">{{ format(data.result.overallScore) }}</div>
      <p>评价均值：{{ format(data.result.evaluationAverage) }}（{{ levelLabel(data.result.evaluationLevel) }}）</p>
      <p>{{ reportText.overallText || data.reportTextMessage }}</p>
    </section>
    <section>
      <h2>测评维度分析</h2>
      <div v-for="dimension in data.dimensions" :key="dimension.id" class="dimension">
        <h3>{{ dimension.dimensionCode }} {{ dimension.dimensionName }}</h3>
        <p>得分 {{ format(dimension.dimensionScore) }} · {{ levelLabel(dimension.levelCode) }} · 已答 {{ dimension.answeredQuestionCount }}/{{ dimension.totalQuestionCount }}</p>
        <p>{{ dimensionText(dimension) }}</p>
      </div>
    </section>
  </div>
</template>

<script>
import { fetchCompetencyInternalReportData, fetchCompetencyReportData } from '@/api/competency'

export default {
  name: 'CompetencyReport',
  data() { return { data: null } },
  computed: {
    audienceLabel() { return this.data.result.reportAudience === 'leader' ? '领导人员版' : '基层员工版' },
    reportDate() { return new Date(this.data.result.submittedAt).toLocaleDateString() },
    reportText() { return this.data.reportText || { overallText: '', dimensionTexts: {} } }
  },
  async created() {
    window.__reportReady = false
    window.__reportIncomplete = false
    try {
      const internalToken = this.$route.query._internal || ''
      const response = internalToken
        ? await fetchCompetencyInternalReportData(this.$route.params.paperId, internalToken)
        : await fetchCompetencyReportData(this.$route.params.paperId)
      this.data = response.data
      this.$nextTick(() => {
        window.__reportReady = true
        window.__reportIncomplete = !this.data.reportTextReady
      })
    } catch (error) {
      window.__reportIncomplete = true
    }
  },
  methods: {
    format(value) { return value === null || value === undefined ? '—' : Number(value).toFixed(2) },
    levelLabel(value) { return { low: '较低', average: '一般', good: '良好', high: '较高' }[value] || '—' },
    dimensionText(dimension) { return this.reportText.dimensionTexts[dimension.dimensionId] || '临时维度文案缺失' }
  }
}
</script>

<style scoped>
.competency-report { max-width:900px; margin:auto; padding:40px; color:#303133; font-size:15px; line-height:1.8; }
.temporary-alert { margin-bottom:24px; }
.cover { min-height:500px; display:flex; flex-direction:column; align-items:center; justify-content:center; border-bottom:2px solid #409eff; }
.cover h1 { font-size:40px; margin:20px 0; }
.temporary-mark { padding:8px 18px; border:2px solid #f56c6c; color:#f56c6c; font-size:20px; font-weight:700; transform:rotate(-4deg); }
.competency-report section { page-break-inside:avoid; margin-bottom:36px; }
.person-grid { display:grid; grid-template-columns:1fr 1fr; gap:10px 24px; }
.score { font-size:42px; color:#409eff; font-weight:700; }
.dimension { padding:16px 0; border-bottom:1px solid #ebeef5; }
.temporary-copy { color:#f56c6c; font-weight:600; }
@media (max-width:600px) { .competency-report{padding:18px}.person-grid{grid-template-columns:1fr}.cover h1{font-size:30px} }
@media print { .competency-report{padding:0}.temporary-alert{display:none}.cover{page-break-after:always} }
</style>

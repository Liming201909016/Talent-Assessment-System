<template>
  <div v-if="data" class="competency-report">
    <el-alert class="temporary-alert" :title="data.reportTextMessage" type="error" :closable="false" show-icon />

    <section class="cover report-page">
      <div class="cover-corner" />
      <div class="cover-content">
        <div class="temporary-mark">临时测试报告</div>
        <h1><strong>胜任力</strong> 全景测评报告</h1>
        <p class="english-title">Competency Panoramic Assessment Report</p>
        <p class="cover-slogan">科学评估 · 精准成长</p>
        <div class="growth-visual" aria-hidden="true">
          <span v-for="height in [42, 68, 96, 126, 88]" :key="height" :style="{ height: height + 'px' }" />
        </div>
        <div class="cover-meta">
          <p class="exam-name">{{ data.meta && data.meta.examTitle || '胜任力测评' }}</p>
          <p>{{ data.result.participantName || '—' }} · {{ audienceLabel }}</p>
          <p>{{ reportDate }}</p>
        </div>
      </div>
      <div class="cover-rings" aria-hidden="true" />
    </section>

    <section class="report-page reading-page">
      <header class="page-title">报告阅读说明</header>
      <p>胜任力测评用于了解受测者在本次所选能力维度上的自陈表现。题目采用五级量表，并将正向题与反向题统一换算到 <strong>1.00–5.00</strong> 分范围。</p>
      <p>维度得分为该维度已答题目的最终分平均值；整体得分为各有效维度得分之和；总体评价依据有效维度平均值划分为较低、一般、良好、较高四个等级。</p>
      <p>整体得分会随测量维度数量变化，<strong>不同维度数量的测评结果不可直接横向比较</strong>。比较个人或群体时，应使用相同维度组合、相同计分版本和相近测评情境。</p>
      <div class="reading-flow" aria-label="报告阅读顺序">
        <div><b>01</b><span>查看总体评价</span></div><i />
        <div><b>02</b><span>比较维度得分</span></div><i />
        <div><b>03</b><span>阅读维度提示</span></div>
      </div>
      <div class="notice-box">
        <h3>特别说明</h3>
        <p>{{ data.reportTextMessage }}</p>
        <p>结果可能受到作答状态、自我认知和具体情境影响，不应作为评价个人能力或人才决策的唯一依据，应结合绩效、行为观察和多方评价综合判断。</p>
      </div>
    </section>

    <section class="report-page overview-page">
      <header class="page-title">报告详细解读</header>
      <h2 class="section-title"><i />个人信息</h2>
      <div class="person-grid">
        <div v-for="field in personFields" :key="field.key"><span>{{ field.label }}</span><strong>{{ field.value || '—' }}</strong></div>
        <div><span>完成时间</span><strong>{{ completedAt }}</strong></div>
        <div><span>答题时长</span><strong>{{ data.meta && data.meta.userTime || '—' }} 分钟</strong></div>
      </div>

      <h2 class="section-title"><i />总体评价</h2>
      <div class="overall-summary">
        <div class="overall-score"><span>评价均值</span><strong>{{ format(data.result.evaluationAverage) }}</strong><em>/ 5.00</em></div>
        <div class="overall-copy">
          <p>您的胜任力总体等级为</p><h3>{{ levelLabel(data.result.evaluationLevel) }}</h3>
          <p>维度得分合计：{{ format(data.result.overallScore) }}</p>
        </div>
      </div>
      <div class="overall-scale">
        <div class="scale-segments"><span class="low">较低</span><span class="average">一般</span><span class="good">良好</span><span class="high">较高</span></div>
        <i :style="{ left: overallMarker + '%' }"><b>{{ format(data.result.evaluationAverage) }}</b></i>
      </div>
      <div class="interpretation"><h3>【临时诊断】</h3><p>{{ reportText.overallText || data.reportTextMessage }}</p></div>
    </section>

    <section class="report-page chart-page">
      <header class="page-title">测评结果分析</header>
      <h2 class="section-title"><i />各维度得分情况</h2>
      <div class="dimension-chart" role="img" aria-label="各维度得分柱状图，满分5分">
        <div v-for="dimension in data.dimensions" :key="dimension.id" class="chart-row">
          <span :title="dimension.dimensionName">{{ dimension.dimensionName }}</span>
          <div><i :style="{ width: scoreWidth(dimension.dimensionScore) + '%' }" /></div>
          <strong>{{ format(dimension.dimensionScore) }}</strong>
        </div>
        <div class="chart-axis"><span>1</span><span>2</span><span>3</span><span>4</span><span>5</span></div>
      </div>
      <div class="legend-box">
        <h3>图示说明</h3>
        <p><b>1.00–1.99 较低：</b>该维度表现需要重点关注。</p>
        <p><b>2.00–2.99 一般：</b>该维度已有一定基础，稳定性仍可提升。</p>
        <p><b>3.00–3.99 良好：</b>该维度表现较稳定，可继续巩固。</p>
        <p><b>4.00–5.00 较高：</b>该维度表现突出，可探索经验迁移。</p>
      </div>
    </section>

    <section v-for="(dimension, index) in data.dimensions" :key="dimension.id" class="report-page dimension-page">
      <header class="page-title">维度解读 {{ pad(index + 1) }}</header>
      <div class="dimension-heading">
        <div class="score-ring" :style="{ '--score': scoreWidth(dimension.dimensionScore) + '%' }"><strong>{{ format(dimension.dimensionScore) }}</strong><span>/ 5.00</span></div>
        <div>
          <small>{{ dimension.dimensionCode }}</small><h2>{{ dimension.dimensionName }}</h2>
          <p>能力等级：<b>{{ levelLabel(dimension.levelCode) }}</b></p>
          <p>题目完成：{{ dimension.answeredQuestionCount }}/{{ dimension.totalQuestionCount }} · 得分合计：{{ dimension.scoreSum }}</p>
        </div>
      </div>
      <div class="meaning-box"><h3>【核心含义】</h3><p>{{ dimensionCoreMeaning(dimension) }}</p></div>
      <div class="dimension-copy"><h3>【临时测评解读】</h3><p>{{ dimensionText(dimension) }}</p></div>
      <div class="development-box">
        <h3>【发展提示】</h3>
        <ul>
          <li>结合真实工作案例复盘该能力的触发情境、行为选择和结果。</li>
          <li>选择一个可观察的小行为进行两至四周练习，并记录反馈。</li>
          <li>与上级、同事或导师进行多方校准，避免仅依赖单次自评结果。</li>
        </ul>
        <p>{{ data.reportTextMessage }}</p>
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
    reportDate() { return this.formatDate(this.data.meta && this.data.meta.generatedAt || this.data.result.submittedAt) },
    completedAt() { return this.formatDate(this.data.result.submittedAt, true) },
    reportText() { return this.data.reportText || { overallText: '', dimensionTexts: {} } },
    overallMarker() {
      const score = Number(this.data.result.evaluationAverage)
      return Math.max(0, Math.min(100, ((score - 1) / 4) * 100))
    },
    personFields() {
      const result = this.data.result
      const configured = String(this.data.meta && this.data.meta.requiredFields || '').split(',').map(item => item.trim()).filter(Boolean)
      const all = [
        { key: 'name', label: '姓名', value: result.participantName },
        { key: 'age', label: '年龄', value: result.participantAge },
        { key: 'gender', label: '性别', value: this.genderLabel(result.participantGender) },
        { key: 'telephone', label: '手机号', value: result.participantTelephone },
        { key: 'affiliation', label: '单位', value: result.participantAffiliation },
        { key: 'post', label: '岗位', value: result.participantPost },
        { key: 'degree', label: '学历', value: result.participantDegree },
        { key: 'major', label: '专业', value: result.participantMajor }
      ]
      return configured.length ? all.filter(field => configured.includes(field.key)) : all
    }
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
    formatDate(value, withTime = false) {
      if (!value) return '—'
      const date = new Date(value)
      const parts = [date.getFullYear(), String(date.getMonth() + 1).padStart(2, '0'), String(date.getDate()).padStart(2, '0')]
      if (!withTime) return `${parts[0]}年${Number(parts[1])}月${Number(parts[2])}日`
      return `${parts.join('-')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    },
    genderLabel(value) { return value === '0' ? '男' : value === '1' ? '女' : value || '—' },
    levelLabel(value) { return { low: '较低', average: '一般', good: '良好', high: '较高' }[value] || '—' },
    scoreWidth(value) { return Math.max(0, Math.min(100, Number(value || 0) / 5 * 100)) },
    pad(value) { return String(value).padStart(2, '0') },
    dimensionText(dimension) { return this.reportText.dimensionTexts[dimension.dimensionId] || '临时维度文案缺失' },
    dimensionCoreMeaning(dimension) { return (this.data.meta && this.data.meta.dimensionCoreMeanings || {})[dimension.dimensionId] || '该维度核心含义待配置。' }
  }
}
</script>

<style scoped>
.competency-report { --green:#27c96f; --deep:#186a48; --mint:#e8f7ec; --ink:#20252b; max-width:900px; margin:auto; color:var(--ink); font-family:"Microsoft YaHei","Noto Sans CJK SC",sans-serif; font-size:15px; line-height:1.85; background:#fff; }
.temporary-alert { margin:20px 40px; }
.report-page { box-sizing:border-box; min-height:1040px; padding:64px 76px; page-break-after:always; position:relative; overflow:hidden; }
.report-page:last-child { page-break-after:auto; }
.cover { padding:0; display:flex; justify-content:center; text-align:center; }
.cover-content { width:100%; padding:170px 60px 70px; z-index:1; }
.cover-corner { position:absolute; right:0; top:0; width:230px; height:230px; background:linear-gradient(135deg,transparent 25%,#c9efb7); }
.cover-rings { position:absolute; left:-95px; bottom:-95px; width:290px; height:290px; border-radius:50%; background:radial-gradient(circle,rgba(74,196,86,.13) 0 34%,transparent 35% 48%,rgba(74,196,86,.1) 49% 67%,transparent 68%); }
.temporary-mark { display:inline-block; padding:5px 16px; border:2px solid #f56c6c; color:#f56c6c; font-weight:700; transform:rotate(-3deg); }
.cover h1 { margin:38px 0 8px; font-size:40px; letter-spacing:2px; }.cover h1 strong { color:var(--green); font-size:54px; font-style:italic; }
.english-title { margin:0; font-size:18px; letter-spacing:1px; }.cover-slogan { font-size:20px; font-weight:700; }
.growth-visual { height:170px; margin:80px auto 45px; display:flex; align-items:flex-end; justify-content:center; gap:24px; border-bottom:8px solid #ddd4bd; width:360px; }
.growth-visual span { width:9px; border-radius:9px 9px 0 0; background:linear-gradient(#68d956,#246d42); box-shadow:12px -10px 0 -2px #8ee779,-12px -22px 0 -2px #65cd70; }
.cover-meta { color:#52ae3e; font-weight:700; }.cover-meta p { margin:3px; }.cover-meta .exam-name { color:var(--ink); font-size:18px; }
.page-title { margin-bottom:38px; text-align:center; font-size:28px; font-weight:800; }.page-title span { color:var(--green); }
.reading-page>p { text-align:justify; text-indent:2em; margin:16px 0; }
.reading-flow { display:flex; align-items:center; justify-content:center; margin:58px 0; }.reading-flow div { width:150px; text-align:center; }.reading-flow b { display:block; color:var(--green); font-size:30px; }.reading-flow span { font-weight:700; }.reading-flow i { width:45px; border-top:2px dashed #9bd7ad; }
.notice-box,.meaning-box { padding:18px 24px; background:var(--mint); border-left:5px solid var(--green); }
.notice-box h3,.meaning-box h3,.interpretation h3,.dimension-copy h3,.development-box h3 { margin:0; color:#0e9f4e; }
.section-title { margin:30px 0 22px; font-size:21px; }.section-title i { display:inline-block; width:10px; height:10px; margin-right:12px; border-radius:50% 0 50% 50%; background:var(--green); transform:rotate(25deg); }
.person-grid { display:grid; grid-template-columns:1fr 1fr; gap:0 36px; padding:14px 28px; background:#fafcf9; }.person-grid div { display:flex; padding:8px 0; border-bottom:1px solid #e8eee9; }.person-grid span { width:86px; color:#6d756f; }
.overall-summary { display:flex; align-items:center; justify-content:center; gap:45px; margin:20px 0; }.overall-score { width:150px; height:150px; border:18px solid #8ce0d3; border-radius:50%; display:flex; flex-direction:column; align-items:center; justify-content:center; box-sizing:border-box; }.overall-score span,.overall-score em { font-size:12px; font-style:normal; color:#6d756f; }.overall-score strong { font-size:34px; }.overall-copy h3 { margin:0; color:var(--green); font-size:28px; }.overall-copy p { margin:3px 0; }
.overall-scale { position:relative; padding-top:32px; margin:28px 0 45px; }.scale-segments { display:grid; grid-template-columns:repeat(4,1fr); color:#fff; text-align:center; font-weight:700; }.scale-segments span { padding:9px 0; }.scale-segments .low { background:#e94b35; }.scale-segments .average { background:#e5bd25; }.scale-segments .good { background:#39bfae; }.scale-segments .high { background:#29c74d; }.overall-scale>i { position:absolute; top:2px; width:2px; height:72px; background:#1d7f73; transform:translateX(-1px); }.overall-scale>i:after { content:""; position:absolute; left:-7px; bottom:-2px; width:14px; height:14px; border-radius:50%; background:#1d7f73; border:3px solid #fff; }.overall-scale>i b { position:absolute; top:-25px; left:-20px; width:42px; text-align:center; color:#1d7f73; }
.interpretation,.dimension-copy,.development-box { padding:18px 22px; background:#fafcf9; }
.dimension-chart { margin:36px 0; padding:24px; border:1px solid #dfe9e2; }.chart-row { display:grid; grid-template-columns:145px 1fr 52px; align-items:center; gap:12px; margin:15px 0; }.chart-row>span { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }.chart-row>div { height:16px; background:repeating-linear-gradient(90deg,#eff5f1 0,#eff5f1 calc(20% - 1px),#d9e5dd 20%); }.chart-row i { display:block; height:100%; background:linear-gradient(90deg,#85ddd2,var(--green)); border-radius:0 8px 8px 0; }.chart-row strong { color:var(--deep); }.chart-axis { display:flex; justify-content:space-between; margin:5px 56px 0 157px; color:#909399; font-size:12px; }
.legend-box { padding:18px 26px; background:var(--mint); }.legend-box p { margin:8px 0; }
.dimension-heading { display:flex; align-items:center; gap:45px; margin:55px 0 38px; }.score-ring { --score:50%; width:170px; height:170px; flex:none; border-radius:50%; background:conic-gradient(var(--green) var(--score),#e5e9e7 0); display:flex; flex-direction:column; align-items:center; justify-content:center; position:relative; }.score-ring:before { content:""; position:absolute; inset:18px; border-radius:50%; background:#fff; }.score-ring strong,.score-ring span { z-index:1; }.score-ring strong { font-size:40px; }.score-ring span { color:#7c8580; }.dimension-heading small { color:var(--green); font-weight:700; }.dimension-heading h2 { margin:0; font-size:30px; }.dimension-heading p { margin:3px 0; }
.meaning-box,.dimension-copy,.development-box { margin:24px 0; }.development-box li { margin:10px 0; }.development-box>p { color:#f56c6c; font-size:13px; font-weight:700; }
@media (max-width:600px) { .report-page{min-height:auto;padding:35px 22px}.cover-content{padding:110px 22px 60px}.cover h1{font-size:28px}.cover h1 strong{font-size:38px}.person-grid{grid-template-columns:1fr}.chart-row{grid-template-columns:100px 1fr 45px}.dimension-heading{gap:20px}.score-ring{width:120px;height:120px}.reading-flow{gap:4px}.reading-flow div{width:100px} }
@media print { .competency-report{max-width:none}.temporary-alert{display:none}.report-page{width:100%;min-height:auto;padding:58px 68px}.cover{min-height:850px;padding:0}.cover-content{padding-top:120px}.overview-page{font-size:14px;line-height:1.6;padding:38px 60px}.overview-page .page-title{margin-bottom:18px}.overview-page .section-title{margin:18px 0 12px}.overview-page .person-grid{padding:8px 20px}.overview-page .person-grid div{padding:5px 0}.overview-page .overall-summary{margin:10px 0;gap:32px}.overview-page .overall-score{width:120px;height:120px;border-width:14px}.overview-page .overall-score strong{font-size:30px}.overview-page .overall-scale{margin:18px 0 25px}.overview-page .interpretation{padding:14px 18px}.dimension-page{font-size:14px;line-height:1.6;padding:38px 60px}.dimension-page .page-title{margin-bottom:20px}.dimension-page .dimension-heading{margin:24px 0 18px}.dimension-page .score-ring{width:140px;height:140px}.dimension-page .meaning-box,.dimension-page .dimension-copy,.dimension-page .development-box{margin:14px 0;padding:14px 18px}.dimension-page .development-box li{margin:5px 0} }
</style>

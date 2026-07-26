<template>
  <div class="mbti-report" id="mbti-report-pdf">
    <!-- ==================== 首页 ==================== -->
    <div class="report-cover">
      <div class="cover-title">
        <h1>职业性格测验报告</h1>
        <p class="cover-en">Career Personality Test Report</p>
      </div>
      <div class="cover-info">
        <div class="info-row" v-if="showCoverField('name')"><span class="info-label">姓名：</span><span class="info-value">{{ testerInfo.name || '—' }}</span></div>
        <div class="info-row" v-if="showCoverField('age')"><span class="info-label">年龄：</span><span class="info-value">{{ testerInfo.age || '—' }}</span></div>
        <div class="info-row" v-if="showCoverField('gender')"><span class="info-label">性别：</span><span class="info-value">{{ genderText }}</span></div>
        <div class="info-row" v-if="showCoverField('affiliation')"><span class="info-label">{{ affiliationLabel }}</span><span class="info-value">{{ testerInfo.affiliation || '—' }}</span></div>
        <div class="info-row" v-if="showCoverField('post')"><span class="info-label">{{ repoCode && repoCode.startsWith('002') ? '职务：' : '岗位：' }}</span><span class="info-value">{{ testerInfo.post || '—' }}</span></div>
        <div class="info-row" v-if="showCoverField('telephone')"><span class="info-label">联系方式：</span><span class="info-value">{{ testerInfo.telephone || '—' }}</span></div>
        <div class="info-row"><span class="info-label">测评时长：</span><span class="info-value">{{ testerInfo.userTime ? testerInfo.userTime + '分钟' : '—' }}</span></div>
        <div class="info-row"><span class="info-label">报告日期：</span><span class="info-value">{{ reportDate }}</span></div>
      </div>
    </div>

    <div class="page-break"></div>

    <!-- ==================== 报告阅读说明 ==================== -->
    <div class="report-section" v-if="tpl">
      <h2 class="section-title">报告阅读说明</h2>
      <p v-for="(p, i) in tpl.intro" :key="'intro-'+i" class="text-paragraph">{{ p }}</p>
    </div>

    <div class="page-break"></div>

    <!-- ==================== 第一部分 职业性格类型 ==================== -->
    <div class="report-section" v-if="tpl">
      <h2 class="section-title">第一部分 · 职业性格类型</h2>
      <div class="type-display">
        <div class="type-badge-large">{{ mbtiType }}</div>
        <div class="type-slogan">{{ tpl.typeSlogan }}</div>
      </div>

      <h3 class="sub-title">你的职业性格类型</h3>
      <p class="text-paragraph">{{ tpl.typeDescription }}</p>

      <!-- 四维度图表 -->
      <div class="chart-container">
        <div class="dim-row" v-for="dim in dimensions" :key="dim.left">
          <div class="dim-letter-left" :class="{active: scores[dim.left] >= scores[dim.right]}">{{ dim.left }}<br/><small>{{ dim.leftDesc }}</small></div>
          <div class="dim-bars">
            <div class="bar-track">
              <div class="bar-fill-left" :style="{width: barPct(scores[dim.left]) + '%'}" :class="{active: scores[dim.left] >= scores[dim.right]}">
                <span class="bar-score">{{ scores[dim.left] }}</span>
              </div>
            </div>
            <div class="bar-track">
              <div class="bar-fill-right" :style="{width: barPct(scores[dim.right]) + '%'}" :class="{active: scores[dim.right] > scores[dim.left]}">
                <span class="bar-score">{{ scores[dim.right] }}</span>
              </div>
            </div>
          </div>
          <div class="dim-letter-right" :class="{active: scores[dim.right] > scores[dim.left]}">{{ dim.right }}<br/><small>{{ dim.rightDesc }}</small></div>
        </div>
      </div>
    </div>

    <div class="page-break"></div>

    <!-- ==================== 第二部分 个性基本描述 ==================== -->
    <div class="report-section" v-if="tpl">
      <h2 class="section-title">第二部分 · 个性基本描述</h2>
      <h3 class="sub-title">个性特征描述</h3>
      <p v-for="(p, i) in tpl.personality" :key="'pers-'+i" class="text-paragraph">{{ p }}</p>
      <h3 class="sub-title">可能存在的盲点</h3>
      <p v-for="(p, i) in tpl.blindSpots" :key="'blind-'+i" class="text-paragraph">{{ p }}</p>
    </div>

    <div class="page-break"></div>

    <!-- ==================== 第三部分 职业性格类型分析 ==================== -->
    <div class="report-section" v-if="tpl">
      <h2 class="section-title">第三部分 · 职业性格类型分析</h2>
      <p class="text-paragraph intro-text">{{ tpl.part3Intro }}</p>
      <h3 class="sub-title">工作中的优势</h3>
      <ul class="trait-list advantage">
        <li v-for="(s, i) in tpl.strengths" :key="'str-'+i">{{ s }}</li>
      </ul>
      <h3 class="sub-title">工作中的劣势</h3>
      <ul class="trait-list weakness">
        <li v-for="(w, i) in tpl.weaknesses" :key="'wk-'+i">{{ w }}</li>
      </ul>
    </div>

    <div class="page-break"></div>

    <!-- ==================== 第四部分 职业发展建议 ==================== -->
    <div class="report-section" v-if="tpl">
      <h2 class="section-title">第四部分 · 职业发展建议</h2>
      <p class="text-paragraph intro-text">{{ tpl.part4Intro }}</p>
      <div class="career-card">
        <div class="career-row"><span class="career-label">职场角色：</span><span class="career-value highlight">{{ tpl.role }}</span></div>
        <div class="career-row"><span class="career-label">职场定位：</span><span class="career-value">{{ tpl.positioning }}</span></div>
      </div>
      <h3 class="sub-title">适应的工作氛围</h3>
      <p class="text-paragraph">作为该类型的人，职业满足意味着你应该：</p>
      <ul class="trait-list">
        <li v-for="(a, i) in tpl.atmosphere" :key="'atm-'+i">{{ a }}</li>
      </ul>
      <p class="text-paragraph" v-if="tpl.atmosphere.length">{{ tpl.atmosphere[tpl.atmosphere.length-1].includes('该类型') ? '' : '' }}</p>
      <h3 class="sub-title">适合的一般职业</h3>
      <ul class="career-list">
        <li v-for="(c, i) in tpl.careers" :key="'car-'+i">{{ c }}</li>
      </ul>
    </div>

    <div class="page-break"></div>

    <!-- ==================== 第五部分 团队适配建议 ==================== -->
    <div class="report-section" v-if="tpl">
      <h2 class="section-title">第五部分 · 团队适配建议</h2>
      <p class="text-paragraph intro-text">{{ tpl.part5Intro }}</p>
      <div class="compat-card" v-for="(ct, i) in tpl.compatibleTypes" :key="'ct-'+i">
        <div class="compat-header">
          <span class="compat-num">适配类型{{ i+1 }}</span>
          <span class="compat-type">{{ ct.type }}</span>
        </div>
        <p class="compat-desc">{{ ct.desc }}</p>
      </div>
    </div>
  </div>
</template>

<script>
import request from '@/utils/request'
import reportTemplates from '@/assets/mbti-report-templates.json'

export default {
  name: 'MbtiReportFull',
  props: {
    disPdfDownload: { type: Boolean, default: false },
    stuFlag: { type: [Number, String], default: 0 },
    repoCode: { type: String, default: '' }
  },
  data() {
    return {
      loading: false,
      isPdf: false,
      paperId: '',
      mbtiType: '',
      scores: { E: 0, I: 0, S: 0, N: 0, T: 0, F: 0, J: 0, P: 0 },
      testerInfo: {},
      testerId: '',
      examId: '',
      requiredFields: '',
      dimensions: [
        { left: 'E', right: 'I', leftDesc: '外向', rightDesc: '内向' },
        { left: 'S', right: 'N', leftDesc: '感觉', rightDesc: '直觉' },
        { left: 'T', right: 'F', leftDesc: '理性', rightDesc: '感性' },
        { left: 'J', right: 'P', leftDesc: '判断', rightDesc: '感知' }
      ]
    }
  },
  computed: {
    tpl() { return reportTemplates[this.mbtiType] || null },
    // FB-040 修复：MBTI(003) 的 stuFlag 与"学校/单位"无关，固定显示"单位"。
    // 仅 001 心理特质 + stuFlag==1（学生版）才显示"学校"。
    affiliationLabel() {
      const isStu = (this.stuFlag === 1 || this.stuFlag === '1') && (this.repoCode || '').startsWith('001')
      return isStu ? '学校：' : '单位：'
    },
    genderText() {
      const g = this.testerInfo.gender
      if (g === '0' || g === 0) return '男'
      if (g === '1' || g === 1) return '女'
      return '—'
    },
    reportDate() {
      const d = new Date()
      return `${d.getFullYear()}年${d.getMonth()+1}月${d.getDate()}日`
    }
  },
  mounted() {
    // createPdf 由 fetchScore 设置 paperId 后触发，不在 mounted 中调用
  },
  methods: {
    showCoverField(field) {
      // 姓名始终显示；测评时长和报告日期始终显示
      if (field === 'name') return true
      // 如果有 requiredFields 配置，按配置显示
      if (this.requiredFields) {
        return this.requiredFields.split(',').includes(field)
      }
      // 默认全显示
      return true
    },
    barPct(score) {
      return Math.round((score / 60) * 100)
    },
    // 兼容 viewPdf 调用方式
    fetchTester(paperId) {
      this.paperId = paperId
      this.loading = true
      // 先查 candidate，再查 tester
      request.post('/exam/api/candidate/tester-info', { paperId }).then(res => {
        if (res.data) {
          this.testerInfo = res.data
          this._loadExamConfig()
        }
      }).catch(() => {
        // 兜底查 tester
        if (this.testerId) {
          request.get('/exam/api/tester/idNumber/' + this.testerId).then(res2 => {
            if (res2.data) this.testerInfo = res2.data
            this._loadExamConfig()
          }).catch(() => { this.loading = false })
        }
      })
    },
    fetchScore(paperId) {
      this.paperId = paperId
      request.post('/exam/api/mbti/score', { paperId }).then(res => {
        if (res.data) {
          this.mbtiType = res.data.type || ''
          this.scores = res.data.scores || this.scores
        }
        this.loading = false
        // paperId 已设置，触发报告生成
        this.createPdf()
      }).catch(() => { this.loading = false })
    },
    pdfDownload() {
      // 占位，兼容 viewPdf 调用
    },
    // ========== 报告生成（调后端 docx→PDF API） ==========
    async createPdf() {
      if (!this.paperId) {
        console.log('MBTI report: no paperId, skip createPdf')
        this.$store.commit('setSinglePdfFinished', true)
        return
      }
      this.$store.commit('setSinglePdfFinished', false)
      try {
        const res = await request.post('/exam/api/mbti/generate-report', { paperId: this.paperId })
        if (res.data && res.data.path) {
          console.log('MBTI report generated:', res.data.fileName)
        }
      } catch (err) {
        console.error('createPdf error:', err)
      }
      this.$store.commit('setSinglePdfFinished', true)
    },
    _loadExamConfig() {
      if (this.testerInfo.examId || this.examId) {
        const eid = this.testerInfo.examId || this.examId
        request.post('/exam/api/exam/exam/detail', { id: eid }).then(res => {
          if (res.data) {
            this.requiredFields = res.data.requiredFields || ''
          }
        }).catch(() => {})
      }
    }
  }
}
</script>

<style scoped>
.mbti-report { max-width: 800px; margin: 0 auto; font-family: 'Microsoft YaHei', sans-serif; color: #333; line-height: 1.8; }
.page-break { page-break-after: always; height: 1px; margin: 40px 0; border-bottom: 1px dashed #ddd; }

/* A4 页面模拟 */
.report-cover { text-align: center; padding: 80px 40px; min-height: 900px; display: flex; flex-direction: column; justify-content: center; }
.report-section { min-height: 400px; padding-bottom: 40px; }
.cover-title h1 { font-size: 36px; color: #1a5276; letter-spacing: 6px; margin-bottom: 8px; }
.cover-en { font-size: 16px; color: #7f8c8d; letter-spacing: 2px; margin-bottom: 60px; }
.cover-info { text-align: left; max-width: 400px; margin: 0 auto; }
.info-row { display: flex; margin: 12px 0; font-size: 16px; }
.info-label { width: 100px; color: #2c7a3a; font-weight: bold; flex-shrink: 0; }
.info-value { flex: 1; border-bottom: 1px solid #bbb; padding-bottom: 2px; min-width: 200px; }

/* 章节 */
.section-title { font-size: 22px; color: #1a5276; border-left: 4px solid #409eff; padding-left: 12px; margin: 30px 0 20px; }
.sub-title { font-size: 18px; color: #2c3e50; margin: 24px 0 12px; padding-bottom: 6px; border-bottom: 2px solid #e8e8e8; }
.text-paragraph { text-indent: 2em; margin: 10px 0; font-size: 15px; }
.intro-text { color: #555; font-style: italic; }

/* 类型展示 */
.type-display { text-align: center; margin: 30px 0; }
.type-badge-large { display: inline-block; background: linear-gradient(135deg, #1a5276, #2980b9); color: white; font-size: 56px; font-weight: bold; padding: 20px 50px; border-radius: 16px; letter-spacing: 12px; }
.type-slogan { font-size: 18px; color: #7f8c8d; margin-top: 12px; }

/* 图表 */
.chart-container { margin: 30px 0; padding: 20px; background: #f8f9fa; border-radius: 12px; }
.dim-row { display: flex; align-items: center; margin: 16px 0; }
.dim-letter-left, .dim-letter-right { width: 50px; text-align: center; font-size: 20px; font-weight: bold; color: #bbb; }
.dim-letter-left.active { color: #409eff; }
.dim-letter-right.active { color: #e6a23c; }
.dim-letter-left small, .dim-letter-right small { font-size: 11px; font-weight: normal; }
.dim-bars { flex: 1; padding: 0 10px; }
.bar-track { height: 24px; background: #e8e8e8; border-radius: 12px; margin: 3px 0; overflow: hidden; position: relative; }
.bar-fill-left { height: 100%; background: #d5e8f7; border-radius: 12px; transition: width 0.6s; display: flex; align-items: center; justify-content: flex-end; padding-right: 8px; }
.bar-fill-left.active { background: #409eff; }
.bar-fill-right { height: 100%; background: #fde8cc; border-radius: 12px; transition: width 0.6s; display: flex; align-items: center; justify-content: flex-end; padding-right: 8px; }
.bar-fill-right.active { background: #e6a23c; }
.bar-score { font-size: 12px; color: white; font-weight: bold; }

/* 特质列表 */
.trait-list { padding-left: 20px; }
.trait-list li { margin: 8px 0; font-size: 15px; line-height: 1.6; }
.trait-list.advantage li::marker { color: #67c23a; }
.trait-list.weakness li::marker { color: #f56c6c; }

/* 职业卡片 */
.career-card { background: linear-gradient(135deg, #f0f5ff, #e6f7ff); padding: 20px 24px; border-radius: 12px; margin: 16px 0; }
.career-row { margin: 8px 0; font-size: 16px; }
.career-label { color: #1a5276; font-weight: bold; }
.career-value.highlight { color: #e6a23c; font-size: 20px; font-weight: bold; }
.career-list { padding-left: 20px; }
.career-list li { margin: 8px 0; font-size: 14px; line-height: 1.6; }

/* 适配卡片 */
.compat-card { background: #fff; border: 1px solid #e8e8e8; border-radius: 12px; padding: 16px 20px; margin: 16px 0; }
.compat-header { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
.compat-num { font-size: 13px; color: #909399; }
.compat-type { font-size: 20px; font-weight: bold; color: #409eff; }
.compat-desc { font-size: 14px; color: #606266; line-height: 1.7; }

@media print {
  @page { size: A4; margin: 15mm 20mm; }
  .page-break { page-break-after: always; border: none; margin: 0; height: 0; }
  .mbti-report { max-width: 100%; font-size: 13px; }
  .report-cover { min-height: auto; padding: 60px 20px; }
  .section-title { font-size: 18px; }
  .sub-title { font-size: 15px; }
  .type-badge-large { font-size: 40px; padding: 15px 35px; }
}
</style>

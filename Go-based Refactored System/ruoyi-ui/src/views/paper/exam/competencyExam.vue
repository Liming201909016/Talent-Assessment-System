<template>
  <div class="competency-exam" v-loading="loading">
    <header class="exam-header" v-if="paper" aria-label="答题进度">
      <div class="question-position">
        <span class="exam-kicker">胜任力测评</span>
        <strong>第 {{ currentIndex + 1 }} 题</strong>
        <span class="question-total">共 {{ paper.totalCount }} 题</span>
      </div>
      <div class="progress-panel">
        <div class="progress-copy">
          <span>答题进度</span>
          <strong>{{ progress }}%</strong>
        </div>
        <el-progress :percentage="progress" :show-text="false" :stroke-width="8" :status="paper.unansweredCount === 0 ? 'success' : null" />
      </div>
      <div class="exam-stats">
        <span><i class="el-icon-circle-check" aria-hidden="true" /> 已答 <strong>{{ paper.answeredCount }}</strong></span>
        <span><i class="el-icon-document" aria-hidden="true" /> 未答 <strong>{{ paper.unansweredCount }}</strong></span>
        <span class="timer"><i class="el-icon-time" aria-hidden="true" /> 剩余 <strong>{{ remainingText }}</strong></span>
      </div>
    </header>

    <el-card v-if="currentQuestion" class="question-card" shadow="never">
      <div class="question-content">{{ currentQuestion.content }}</div>
      <div class="scale-hint">请选择最符合自身实际情况的一项</div>
      <el-radio-group class="scale-options" v-model="selectedValue" :disabled="saving || submitting" aria-label="请选择符合程度" @change="handleAnswer">
        <el-radio v-for="option in currentQuestion.options" :key="option.rawValue" :label="option.rawValue" border>
          {{ option.label }}
        </el-radio>
      </el-radio-group>
      <div class="save-state" role="status" aria-live="polite">
        <i v-if="saveState === '已保存'" class="el-icon-circle-check" aria-hidden="true" />
        {{ saveState }}
      </div>
    </el-card>

    <div class="actions" v-if="paper">
      <div class="step-actions">
        <el-button class="previous-button" icon="el-icon-arrow-left" :disabled="currentIndex === 0" @click="currentIndex--">上一题</el-button>
        <el-button v-if="paper.unansweredCount" class="unanswered-button" icon="el-icon-position" @click="locateFirstUnanswered">第一道未答</el-button>
        <el-button class="next-button" type="primary" plain :disabled="currentIndex >= paper.questions.length - 1" @click="currentIndex++">下一题<i class="el-icon-arrow-right el-icon--right" aria-hidden="true" /></el-button>
      </div>
      <el-button class="submit-button" type="primary" icon="el-icon-finished" :loading="submitting" @click="confirmSubmit">交卷</el-button>
    </div>

    <div class="question-overview" v-if="paper">
      <div class="overview-header">
        <strong>题目导航</strong>
        <div class="nav-legend" aria-label="题目状态说明">
          <span><i class="legend-dot current" /> 当前</span>
          <span><i class="legend-dot answered" /> 已答</span>
          <span><i class="legend-dot unanswered" /> 未答</span>
        </div>
      </div>
      <div class="question-nav" aria-label="题目导航">
        <button v-for="(question, index) in paper.questions" type="button" :key="question.id" :class="{ answered: question.answered, active: index === currentIndex }" :aria-label="`第 ${index + 1} 题，${question.answered ? '已答' : '未答'}`" :aria-current="index === currentIndex ? 'true' : null" @click="currentIndex = index">
          {{ index + 1 }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { fetchCompetencyPaper, saveCompetencyAnswer, submitCompetencyPaper } from '@/api/competency'

export default {
  name: 'CompetencyExam',
  data() {
    return { paper: null, currentIndex: 0, loading: true, saving: false, submitting: false, saveState: '', now: Date.now(), timer: null }
  },
  computed: {
    paperId() { return this.$route.params.paperId },
    paperToken() { return sessionStorage.getItem('competencyPaperToken') || '' },
    currentQuestion() { return this.paper && this.paper.questions[this.currentIndex] },
    selectedValue: {
      get() { return this.currentQuestion ? this.currentQuestion.rawValue : null },
      set(value) { if (this.currentQuestion) this.currentQuestion.rawValue = value }
    },
    progress() { return this.paper && this.paper.totalCount ? Math.round(this.paper.answeredCount * 100 / this.paper.totalCount) : 0 },
    remainingSeconds() { return this.paper && this.paper.limitTime ? Math.max(0, Math.floor((new Date(this.paper.limitTime).getTime() - this.now) / 1000)) : 0 },
    remainingText() { const value = this.remainingSeconds; return `${Math.floor(value / 60)}:${String(value % 60).padStart(2, '0')}` }
  },
  watch: {
    currentIndex() { this.saveState = '' }
  },
  created() { this.loadPaper() },
  beforeDestroy() { clearInterval(this.timer) },
  methods: {
    async loadPaper() {
      if (!this.paperToken) { this.$message.error('试卷认证已失效'); this.$router.go(-1); return }
      try {
        const response = await fetchCompetencyPaper(this.paperId, this.paperToken)
        this.paper = response.data
        if (this.paper.state === 2) { this.finish() ; return }
        this.timer = setInterval(() => { this.now = Date.now(); if (this.remainingSeconds === 0) this.submit('timeout') }, 1000)
      } finally { this.loading = false }
    },
    async handleAnswer(value) {
      const answerIndex = this.currentIndex
      const question = this.currentQuestion
      this.saving = true; this.saveState = '保存中…'
      try {
        const response = await saveCompetencyAnswer(this.paperId, question.id, value, this.paperToken)
        if (response.data.expired) { this.finish(); return }
        if (!question.answered) { question.answered = true }
        this.paper.answeredCount = response.data.answeredCount
        this.paper.unansweredCount = this.paper.totalCount - response.data.answeredCount
        this.saveState = '已保存'
        if (this.currentIndex === answerIndex && answerIndex < this.paper.questions.length - 1) {
          this.currentIndex = answerIndex + 1
        }
      } catch (error) { this.saveState = '保存失败，请重新选择'; question.rawValue = null }
      finally { this.saving = false }
    },
    locateFirstUnanswered() { const index = this.paper.questions.findIndex(question => !question.answered); if (index >= 0) this.currentIndex = index },
    confirmSubmit() {
      if (this.paper.unansweredCount > 0) { this.locateFirstUnanswered(); this.$message.warning(`还有 ${this.paper.unansweredCount} 道题未答`); return }
      this.$confirm('交卷后答案不可修改，确认提交吗？', '确认交卷', { type: 'warning' }).then(() => this.submit('manual')).catch(() => {})
    },
    async submit(type) {
      if (this.submitting) return
      this.submitting = true
      try { await submitCompetencyPaper(this.paperId, 'manual', this.paperToken); this.finish() }
      catch (error) { if (type === 'timeout') this.$message.error('到期提交失败，系统将继续重试'); this.submitting = false }
    },
    finish() { clearInterval(this.timer); sessionStorage.removeItem('competencyPaperToken'); this.$router.replace({ name: 'ExamThankYou' }) }
  }
}
</script>

<style scoped>
.competency-exam {
  --primary: #5b5bd6;
  --primary-soft: #f0efff;
  --success: #48a868;
  --success-soft: #edf8f1;
  --text: #262a33;
  --muted: #697180;
  max-width: 1180px;
  min-height: 100vh;
  margin: 0 auto;
  padding: 24px;
  color: var(--text);
  background: #f7f8fc;
}

.exam-header {
  display: grid;
  grid-template-columns: 190px minmax(260px, 1fr) auto;
  gap: 28px;
  align-items: center;
  margin-bottom: 16px;
  padding: 18px 22px;
  border: 1px solid #e8eaf1;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(38, 42, 51, .06);
}

.question-position {
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: baseline;
  gap: 3px 8px;
}

.exam-kicker {
  grid-column: 1 / -1;
  color: var(--primary);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 1px;
}

.question-position strong { font-size: 20px; }
.question-total { color: var(--muted); font-size: 13px; }

.progress-copy {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  color: var(--muted);
  font-size: 13px;
}

.progress-copy strong { color: var(--primary); }

.exam-stats {
  display: flex;
  gap: 16px;
  align-items: center;
  color: var(--muted);
  font-size: 13px;
  white-space: nowrap;
}

.exam-stats span { display: inline-flex; gap: 4px; align-items: center; }
.exam-stats strong { color: var(--text); font-size: 15px; }
.exam-stats i { color: var(--success); font-size: 15px; }
.exam-stats .timer { padding-left: 16px; border-left: 1px solid #e2e5ed; }
.exam-stats .timer i { color: var(--primary); }
.exam-stats .timer strong { min-width: 44px; color: var(--primary); font-variant-numeric: tabular-nums; }

.question-card {
  border: 1px solid #e8eaf1;
  border-radius: 14px;
  box-shadow: 0 12px 32px rgba(38, 42, 51, .07);
}

.question-card ::v-deep .el-card__body { padding: 30px 32px 24px; }

.question-content {
  min-height: 62px;
  margin: 0 0 20px;
  font-size: 22px;
  font-weight: 500;
  line-height: 1.75;
}

.scale-hint {
  margin-bottom: 12px;
  color: var(--muted);
  font-size: 13px;
}

.scale-options {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  width: 100%;
}

.scale-options .el-radio {
  display: flex;
  height: 52px;
  margin: 0;
  padding: 0 16px;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background: #fff;
  transition: border-color .2s ease, background .2s ease, box-shadow .2s ease, transform .2s ease;
}

.scale-options .el-radio:hover {
  border-color: var(--primary);
  background: var(--primary-soft);
  transform: translateY(-1px);
}

.scale-options .el-radio.is-checked {
  border-color: var(--primary);
  background: var(--primary-soft);
  box-shadow: 0 0 0 2px rgba(91, 91, 214, .12);
}

.scale-options ::v-deep .el-radio__label { padding-left: 8px; font-size: 14px; }
.scale-options ::v-deep .el-radio__input.is-checked + .el-radio__label { color: var(--primary); font-weight: 600; }
.scale-options ::v-deep .el-radio__input.is-checked .el-radio__inner { border-color: var(--primary); background: var(--primary); }

.save-state {
  height: 22px;
  margin-top: 16px;
  color: var(--success);
  font-size: 13px;
  text-align: right;
}

.actions {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin: 16px 0;
  padding: 14px 16px;
  border: 1px solid #e8eaf1;
  border-radius: 12px;
  background: #fff;
}

.step-actions { display: flex; gap: 8px; }
.actions .el-button + .el-button { margin-left: 0; }
.submit-button { min-width: 104px; }

.question-overview {
  padding: 18px 20px 20px;
  border: 1px solid #e8eaf1;
  border-radius: 12px;
  background: #fff;
}

.overview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.overview-header strong { font-size: 15px; }
.nav-legend { display: flex; gap: 16px; color: var(--muted); font-size: 12px; }
.nav-legend span { display: inline-flex; gap: 6px; align-items: center; }
.legend-dot { width: 9px; height: 9px; border: 1px solid #cfd3de; border-radius: 50%; background: #fff; }
.legend-dot.current { border-color: var(--primary); background: var(--primary); }
.legend-dot.answered { border-color: var(--success); background: var(--success-soft); }

.question-nav {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(42px, 1fr));
  gap: 8px;
}

.question-nav button {
  min-width: 42px;
  height: 40px;
  border: 1px solid #dfe2ea;
  border-radius: 8px;
  color: #4c5360;
  background: #fff;
  font: inherit;
  font-size: 13px;
  cursor: pointer;
  transition: border-color .2s ease, background .2s ease, color .2s ease, transform .2s ease;
}

.question-nav button:hover { border-color: var(--primary); color: var(--primary); transform: translateY(-1px); }
.question-nav button.answered { border-color: #91ceaa; color: #287a49; background: var(--success-soft); }
.question-nav button.active { border-color: var(--primary); color: #fff; background: var(--primary); box-shadow: 0 4px 10px rgba(91, 91, 214, .22); }
.question-nav button:focus-visible { outline: 3px solid rgba(91, 91, 214, .24); outline-offset: 2px; }

@media (max-width: 900px) {
  .exam-header { grid-template-columns: 160px 1fr; gap: 18px; }
  .exam-stats { grid-column: 1 / -1; justify-content: flex-end; }
  .scale-options { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}

@media (max-width: 600px) {
  .competency-exam { padding: 10px; }
  .exam-header { grid-template-columns: 1fr; gap: 14px; padding: 16px; }
  .exam-stats { grid-column: auto; justify-content: space-between; gap: 8px; }
  .exam-stats .timer { padding-left: 8px; }
  .question-card ::v-deep .el-card__body { padding: 22px 18px 18px; }
  .question-content { min-height: 0; margin: 22px 0 18px; font-size: 19px; }
  .scale-options { grid-template-columns: 1fr; gap: 8px; }
  .scale-options .el-radio { height: 48px; justify-content: flex-start; }
  .actions { align-items: stretch; }
  .step-actions { display: grid; grid-template-columns: 1fr 1fr; flex: 1; }
  .step-actions .el-button { width: 100%; padding: 10px; }
  .previous-button { grid-column: 1; grid-row: 1; }
  .next-button { grid-column: 2; grid-row: 1; }
  .unanswered-button { grid-column: 1 / -1; grid-row: 2; }
  .submit-button { min-width: 80px; }
  .overview-header { align-items: flex-start; gap: 12px; }
  .nav-legend { gap: 8px; }
  .question-nav { grid-template-columns: repeat(5, 1fr); }
  .question-nav button { height: 44px; }
}
</style>

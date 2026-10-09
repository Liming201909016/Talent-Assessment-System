<template>
  <div class="management-traits-exam" v-loading="loading">
    <p class="test-notice"><strong>TEST · 管理特质</strong> — 仅供测试，不可作为人才决策依据</p>
    <p v-if="resumeContext" class="test-notice" role="status">续答凭据本机有效期至 {{ new Date(resumeContext.expiresAt).toLocaleString() }}；不延长原试卷截止时间。到期请联系管理员。</p>
    <div v-if="error" class="error-panel" role="alert">
      <p>{{ error }}</p>
      <el-button v-if="!paper" @click="loadPaper" :disabled="loading">重试加载</el-button>
      <el-button v-else @click="submit" :disabled="busy">重试交卷</el-button>
    </div>
    <template v-if="paper">
      <header class="exam-header" aria-label="答题进度">
        <div><span class="exam-kicker">{{ paper.title }}</span><h1>第 {{ currentIndex + 1 }} 题 <small>/ {{ paper.questions.length }}</small></h1></div>
        <div class="progress-panel"><div>已答 {{ paper.answered }} · 未答 {{ unanswered }} <strong>{{ progress }}%</strong></div><el-progress :percentage="progress" :show-text="false" :stroke-width="8" /></div>
        <div class="timer" aria-label="服务器倒计时"><i class="el-icon-time" aria-hidden="true" /> 剩余 <strong>{{ remainingText }}</strong></div>
      </header>
      <p v-if="reminderVisible" class="reminder" role="status">已到服务器20分钟提示点，请注意剩余时间。到期由服务器决定完成状态。</p>
      <el-card v-if="currentQuestion" class="question-card" shadow="never">
        <h2>{{ currentQuestion.content }}</h2>
        <p class="scale-hint">请选择最符合自身实际情况的一项（也可按数字键 1–5）</p>
        <div class="scale-options" role="group" aria-label="请选择符合程度">
          <button v-for="(option, index) in currentQuestion.options" :key="option.id" type="button" :disabled="busy || expired"
            :aria-pressed="currentQuestion.selectedOptionId === option.id ? 'true' : 'false'"
            :class="{ selected: currentQuestion.selectedOptionId === option.id }" @click="handleAnswer(option.id)">
            <span class="option-number">{{ index + 1 }}</span>{{ option.content }}
          </button>
        </div>
        <div class="save-state" role="status" aria-live="polite">{{ saveState }}
          <el-button v-if="retryOptionId" type="text" :disabled="busy || expired" @click="retryAnswer">重试保存</el-button>
        </div>
      </el-card>
      <div class="actions">
        <div class="step-actions">
          <el-button icon="el-icon-arrow-left" :disabled="busy || currentIndex === 0" @click="navigate(currentIndex - 1)">上一题</el-button>
          <el-button :disabled="busy" @click="locateFirstUnanswered">第一道未答</el-button>
          <el-button type="primary" plain :disabled="busy || currentIndex === 139" @click="navigate(currentIndex + 1)">下一题</el-button>
        </div>
        <el-button type="primary" icon="el-icon-finished" :loading="submitting" :disabled="saving || confirming" @click="confirmSubmit">交卷</el-button>
      </div>
      <section class="question-overview" aria-label="题目导航">
        <div class="overview-header"><strong>题目导航</strong><span>当前 · 已答 · 未答</span></div>
        <div class="question-nav">
          <button v-for="(question, index) in paper.questions" :key="question.id" type="button" :disabled="busy"
            :class="{ answered: question.selectedOptionId !== null, active: index === currentIndex }"
            :aria-label="`第 ${index + 1} 题，${question.selectedOptionId !== null ? '已答' : '未答'}`"
            :aria-current="index === currentIndex ? 'true' : null" @click="navigate(index)">{{ index + 1 }}</button>
        </div>
      </section>
    </template>
  </div>
</template>

<script>
import { fetchManagementTraitsPaper, saveManagementTraitsAnswer, submitManagementTraitsPaper, rememberManagementTraitsPaper, clearManagementTraitsTokens } from '@/api/managementTraits'
import { consumeManagementTraitsResumeRoute, storedManagementTraitsResume, bindManagementTraitsResumePaper, clearManagementTraitsResume } from '@/utils/managementTraitsResume'

function validatePaper(paper, paperId) {
  if (!paper || paper.paperId !== paperId || !paper.examId || ![1, 2].includes(paper.state) || !Array.isArray(paper.questions) || paper.questions.length !== 140) throw new Error('管理特质试卷数据无效，请联系管理员。')
  const times = ['serverTime', 'startedAt', 'deadline', 'reminderAt'].map(key => Date.parse(paper[key]))
  if (times.some(value => !Number.isFinite(value)) || times[2] <= times[1] || times[3] <= times[1] || times[3] >= times[2]) throw new Error('服务器计时数据无效，请重试。')
  const ids = new Set()
  let answered = 0
  paper.questions.forEach((question, index) => {
    if (!question || typeof question.id !== 'string' || !question.id || ids.has(question.id) || question.displayOrder !== index + 1 || typeof question.content !== 'string' || !Array.isArray(question.options) || question.options.length !== 5) throw new Error('题目快照无效，请联系管理员。')
    ids.add(question.id)
    const options = new Set()
    question.options.forEach((option, order) => {
      if (!option || typeof option.id !== 'string' || !option.id || options.has(option.id) || typeof option.content !== 'string' || option.displayOrder !== order + 1) throw new Error('选项快照无效，请联系管理员。')
      options.add(option.id)
    })
    if (question.selectedOptionId !== null) {
      if (!options.has(question.selectedOptionId)) throw new Error('已保存答案无效，请联系管理员。')
      answered++
    }
  })
  if (paper.answered !== answered) throw new Error('试卷答题统计不一致，请重试。')
  return times
}

export default {
  name: 'ManagementTraitsExam',
  data() {
    return { paper: null, currentIndex: 0, loading: false, saving: false, submitting: false, confirming: false, saveState: '', error: '', retryOptionId: '', timer: null, anchor: 0, elapsed: 0, serverTime: 0, deadline: 0, reminderAt: 0, autoAttempted: false, disposed: false, resumeContext: null, resumeAttempted: false, resumeInvalid: false, loadSequence: 0 }
  },
  computed: {
    paperId() { return this.$route.params.paperId },
    paperToken() { return this.resumeContext ? this.resumeContext.token : sessionStorage.getItem('managementTraitsPaperToken') || '' },
    currentQuestion() { return this.paper && this.paper.questions[this.currentIndex] },
    busy() { return this.loading || this.saving || this.submitting || this.confirming },
    unanswered() { return this.paper ? 140 - this.paper.answered : 140 },
    progress() { return this.paper ? Math.round(this.paper.answered * 100 / 140) : 0 },
    remainingSeconds() { return this.paper ? Math.max(0, Math.ceil((this.deadline - this.serverTime - this.elapsed) / 1000)) : 0 },
    remainingText() { const value = this.remainingSeconds; return `${Math.floor(value / 60)}:${String(value % 60).padStart(2, '0')}` },
    expired() { return !!this.paper && this.remainingSeconds === 0 },
    reminderVisible() { return !!this.paper && this.serverTime + this.elapsed >= this.reminderAt }
  },
  watch: {
    $route(to, from) {
      const incoming = to.query && Object.prototype.hasOwnProperty.call(to.query, 'resumeToken')
      if (!incoming && to.params.paperId === from.params.paperId) return
      this.loadSequence++; clearInterval(this.timer)
      this.loading = false; this.saving = false; this.submitting = false; this.confirming = false
      this.paper = null; this.resumeContext = null; this.resumeAttempted = false; this.resumeInvalid = false
      this.retryOptionId = ''; this.saveState = ''
      this.loadPaper()
    }
  },
  created() { this.loadPaper() },
  mounted() { document.addEventListener('keydown', this.onKeydown) },
  beforeDestroy() { this.disposed = true; this.resumeContext = null; clearInterval(this.timer); document.removeEventListener('keydown', this.onKeydown) },
  methods: {
    resumeUsable() {
      if (this.resumeInvalid || (this.resumeContext && Date.now() >= this.resumeContext.expiresAt)) {
        this.resumeInvalid = true; this.resumeContext = null; this.paper = null
        clearManagementTraitsTokens(); clearManagementTraitsResume(); clearInterval(this.timer)
        this.error = '续答凭据无效或已到期，请联系管理员重新签发；不会重开试卷。'
        return false
      }
      return true
    },
    async loadPaper() {
      if (this.loading || this.disposed) return
      if (!this.resumeUsable()) return
      this.loading = true; this.error = ''; clearInterval(this.timer)
      const sequence = ++this.loadSequence
      try {
        if (!this.resumeAttempted && this.$route.query && Object.prototype.hasOwnProperty.call(this.$route.query, 'resumeToken')) {
          this.resumeAttempted = true; clearManagementTraitsTokens(); clearManagementTraitsResume()
          let context
          try { context = await consumeManagementTraitsResumeRoute(this.$route, this.$router) }
          catch (error) { if (sequence === this.loadSequence) this.resumeInvalid = true; throw error }
          if (this.disposed || sequence !== this.loadSequence) return
          this.resumeContext = context
        } else if (!this.resumeContext && sessionStorage.getItem('managementTraitsResumeBinding')) {
          try { this.resumeContext = storedManagementTraitsResume(this.paperId) }
          catch (error) { this.resumeInvalid = true; clearManagementTraitsTokens(); clearManagementTraitsResume(); throw error }
        }
        if (!this.resumeUsable()) return
        if (!this.paperToken) throw new Error('管理特质试卷认证已失效，请返回重新登录。')
        const response = await fetchManagementTraitsPaper(this.paperId, this.paperToken)
        if (this.disposed || sequence !== this.loadSequence) return
        let data = response.data
        const times = validatePaper(data, this.paperId)
        if (this.resumeContext) {
          try { data = bindManagementTraitsResumePaper(data, this.resumeContext) }
          catch (error) { clearManagementTraitsTokens(); clearManagementTraitsResume(); this.resumeInvalid = true; this.resumeContext = null; throw error }
        } else rememberManagementTraitsPaper(data)
        this.paper = data
        if (data.state === 2) { this.finish(); return }
        this.serverTime = times[0]; this.deadline = times[2]; this.reminderAt = times[3]
        this.anchor = performance.now(); this.elapsed = 0; this.autoAttempted = false
        this.locateFirstUnanswered()
        this.timer = setInterval(this.tick, 1000)
      } catch (error) { if (!this.disposed && sequence === this.loadSequence) { this.paper = null; this.error = error.message || '试卷加载失败，请重试。' } }
      finally { if (sequence === this.loadSequence) this.loading = false }
      if (sequence === this.loadSequence) this.tick()
    },
    tick() {
      if (!this.resumeUsable()) return
      if (!this.paper || this.disposed) return
      this.elapsed = Math.max(this.elapsed, performance.now() - this.anchor)
      if (this.expired && !this.busy && !this.autoAttempted) { this.autoAttempted = true; this.submit() }
    },
    navigate(index) {
      if (this.busy || !this.paper || index < 0 || index >= 140) return
      this.currentIndex = index; this.saveState = ''; this.retryOptionId = ''
    },
    locateFirstUnanswered() {
      if (!this.paper || this.saving || this.submitting) return
      const index = this.paper.questions.findIndex(question => question.selectedOptionId === null)
      if (index >= 0) this.currentIndex = index
    },
    async handleAnswer(optionId) {
      if (!this.resumeUsable()) return
      if (this.busy || this.expired || !this.currentQuestion || !this.currentQuestion.options.some(option => option.id === optionId)) return
      const index = this.currentIndex; const question = this.currentQuestion
      const sequence = this.loadSequence
      this.saving = true; this.saveState = '保存中…'; this.retryOptionId = ''
      try {
        const response = await saveManagementTraitsAnswer(this.paperId, question.id, optionId, this.paperToken)
        if (this.disposed || sequence !== this.loadSequence || !this.resumeUsable()) return
        const count = response.data && response.data.answered
        if (!Number.isInteger(count) || count < 0 || count > 140) throw new Error('保存响应无效，请重试确认。')
        question.selectedOptionId = optionId; this.paper.answered = count; this.saveState = '已保存'
        if (this.currentIndex === index && index < 139) this.currentIndex++
      } catch (error) { if (!this.disposed && sequence === this.loadSequence) { this.retryOptionId = optionId; this.saveState = `保存失败，原答案保留：${error.message || '请重试'}` } }
      finally { if (sequence === this.loadSequence) this.saving = false }
    },
    retryAnswer() { return this.handleAnswer(this.retryOptionId) },
    onKeydown(event) {
      if (event.ctrlKey || event.altKey || event.metaKey || ['INPUT', 'TEXTAREA', 'SELECT', 'BUTTON'].includes(event.target && event.target.tagName) || (event.target && event.target.isContentEditable)) return
      if (/^[1-5]$/.test(event.key) && this.currentQuestion) { event.preventDefault(); return this.handleAnswer(this.currentQuestion.options[Number(event.key) - 1].id) }
      if (event.key === 'ArrowLeft') this.navigate(this.currentIndex - 1)
      if (event.key === 'ArrowRight') this.navigate(this.currentIndex + 1)
    },
    async confirmSubmit() {
      if (!this.resumeUsable()) return
      if (this.busy || !this.paper) return
      if (this.expired) return this.submit()
      if (this.unanswered) { this.locateFirstUnanswered(); this.$message.warning(`还有 ${this.unanswered} 道题未答`); return }
      this.confirming = true
      const sequence = this.loadSequence
      let confirmed = false
      try { await this.$confirm('交卷后答案不可修改，确认提交吗？', '确认交卷', { type: 'warning' }); confirmed = true } catch (_) {}
      finally { if (sequence === this.loadSequence) this.confirming = false }
      if (confirmed && !this.disposed && sequence === this.loadSequence) return this.submit()
    },
    async submit() {
      if (!this.resumeUsable()) return
      if (this.busy || !this.paper || this.disposed) return
      const sequence = this.loadSequence
      this.submitting = true; this.error = ''
      try {
        const response = await submitManagementTraitsPaper(this.paperId, this.paperToken)
        if (this.disposed || sequence !== this.loadSequence) return
        const result = response.data
        if (!result || !['completed', 'incomplete'].includes(result.status)) throw new Error('交卷状态未确认，请重试。')
        this.finish()
      } catch (error) { if (!this.disposed && sequence === this.loadSequence) this.error = `交卷未成功：${error.message || '请重试'}。服务器未确认前不会结束测评。` }
      finally { if (sequence === this.loadSequence) this.submitting = false }
    },
    finish() { clearInterval(this.timer); clearManagementTraitsTokens(); if (this.resumeContext) clearManagementTraitsResume(); this.resumeContext = null; this.$router.replace({ name: 'ExamThankYou' }) }
  }
}
</script>

<style scoped>
.management-traits-exam { --primary: #5b5bd6; --success: #287a49; max-width: 1180px; min-height: 100vh; margin: auto; padding: 24px; background: #f7f8fc; color: #262a33; }
.test-notice { margin: 0 0 16px; font-size: 14px; line-height: 1.5; }
.exam-header { display: grid; grid-template-columns: 1fr 1fr auto; gap: 24px; align-items: center; padding: 20px; border-radius: 14px; background: white; }
.exam-kicker, .scale-hint { font-size: 14px; color: #626979; }
h1 { margin: 8px 0 0; font-size: 20px; } small { font-size: 14px; font-weight: normal; }
.progress-panel strong { float: right; } .progress-panel > div { margin-bottom: 8px; }
.timer strong { color: var(--primary); font-variant-numeric: tabular-nums; }
.question-card { margin-top: 16px; border-radius: 14px; }
.question-card h2 { font-size: 22px; line-height: 1.6; overflow-wrap: anywhere; }
.scale-options { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 12px; }
.scale-options button { min-width: 0; min-height: 52px; padding: 12px; border: 1px solid #cfd3de; border-radius: 10px; color: #262a33; background: white; font: inherit; cursor: pointer; overflow-wrap: anywhere; }
.scale-options button.selected { border-color: var(--primary); background: #f0efff; color: var(--primary); }
.option-number { display: inline-block; margin-right: 8px; font-weight: bold; }
.save-state { min-height: 32px; margin-top: 16px; line-height: 1.5; font-size: 14px; }
.actions { display: flex; justify-content: space-between; gap: 16px; margin: 16px 0; padding: 16px; background: white; border-radius: 12px; }
.step-actions { display: flex; gap: 8px; } .actions .el-button + .el-button { margin-left: 0; }
.question-overview { padding: 20px; background: white; border-radius: 12px; } .overview-header { display: flex; justify-content: space-between; margin-bottom: 16px; gap: 12px; font-size: 14px; }
.question-nav { display: grid; grid-template-columns: repeat(auto-fill, minmax(44px, 1fr)); gap: 8px; }
.question-nav button { height: 44px; border: 1px solid #cfd3de; background: white; border-radius: 8px; font: inherit; cursor: pointer; }
.question-nav button.answered { border-color: var(--success); color: var(--success); background: #edf8f1; }
.question-nav button.active { border-color: var(--primary); color: white; background: var(--primary); }
button:focus-visible { outline: 3px solid #5b5bd6; outline-offset: 2px; } button:disabled { cursor: not-allowed; opacity: .65; }
.error-panel { color: #b23c3c; padding: 16px; background: white; border: 1px solid currentColor; border-radius: 12px; overflow-wrap: anywhere; }
.reminder { padding: 12px; background: #f0efff; line-height: 1.5; }
@media (max-width: 900px) { .exam-header { grid-template-columns: 1fr 1fr; } .timer { grid-column: 1 / -1; } .scale-options { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@media (max-width: 600px) { .management-traits-exam { padding: 10px; } .exam-header { grid-template-columns: 1fr; gap: 16px; padding: 16px; } .timer { grid-column: auto; } .scale-options { grid-template-columns: 1fr; gap: 8px; } .scale-options button { min-height: 48px; text-align: left; } .question-card h2 { font-size: 19px; } .actions { flex-direction: column; padding: 12px; } .step-actions { display: grid; grid-template-columns: 1fr 1fr; } .step-actions .el-button:nth-child(2) { grid-column: 1 / -1; grid-row: 2; } .question-nav { grid-template-columns: repeat(5, minmax(0, 1fr)); } .question-overview { padding: 16px; } }
</style>
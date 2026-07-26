<template>
  <div class="mbti-exam">
    <!-- 顶部 -->
    <div class="mbti-header">
      <div class="progress-bar"><div class="progress-fill" :style="{width: progressPct + '%'}"></div></div>
      <div class="header-info">
        <span class="qu-counter">{{ currentIndex + 1 }} / {{ quList.length }}</span>
        <span class="timer" :class="{warning: leftSeconds < 120}">{{ formatTime(leftSeconds) }}</span>
      </div>
    </div>

    <!-- 题目卡片 -->
    <div class="question-card" v-if="currentQu">
      <div class="qu-title">{{ currentQu.title }}</div>

      <!-- 选项文字（上方，醒目） -->
      <div class="opts-top">
        <div class="opt-card opt-a" :class="{hi: selectedCircle >= 0 && selectedCircle <= 2}">
          <span class="opt-tag">A</span>
          <span class="opt-text">{{ currentQu.optionA }}</span>
        </div>
        <div class="opt-card opt-b" :class="{hi: selectedCircle >= 0 && selectedCircle >= 3}">
          <span class="opt-tag">B</span>
          <span class="opt-text">{{ currentQu.optionB }}</span>
        </div>
      </div>

      <!-- 5圈选择器 A ◉◉○◉◉ B -->
      <div class="picker-row">
        <div class="side-label side-a" :class="{win: selectedCircle >= 0 && selectedCircle <= 2}">A</div>
        <div class="circles">
          <div v-for="(c, i) in 6" :key="i"
               class="circle" :class="[sizeClass(i), colorClass(i), {selected: selectedCircle === i}]"
               @click="selectCircle(i)">
            <div class="fill" v-show="selectedCircle === i"></div>
          </div>
        </div>
        <div class="side-label side-b" :class="{win: selectedCircle >= 0 && selectedCircle >= 3}">B</div>
      </div>

      <div class="hint" v-if="selectedCircle < 0">← 偏向A · 选择圆圈 · 偏向B →</div>
     </div>

    <!-- 导航 -->
    <div class="nav-bar">
      <button class="nb" :disabled="currentIndex === 0" @click="prev">‹ 上一题</button>
      <div class="dots">
        <span v-for="(q, i) in quList" :key="q.quId" class="d"
              :class="{ok: q.answered, cur: i === currentIndex}" @click="goTo(i)"></span>
      </div>
      <button class="nb primary" v-if="currentIndex < quList.length - 1" @click="next">下一题 ›</button>
      <button class="nb success" v-else @click="handleSubmit">交卷</button>
    </div>
    <div class="unanswered" v-if="unansweredCount > 0" @click="goToFirstUnanswered">
      定位未答（{{ unansweredCount }} 题）
    </div>
  </div>
</template>

<script>
import request from '@/utils/request'

// 6个圈位置→A分: 大中小小中大 = 5,4,3,2,1,0
const CIRCLE_A = [5, 4, 3, 2, 1, 0]
const CIRCLE_B = [0, 1, 2, 3, 4, 5]

export default {
  name: 'MbtiExam',
  data() {
    return {
      paperId: '', testerId: '', quList: [], currentIndex: 0,
      scoreA: -1, scoreB: -1, selectedCircle: -1,
      leftSeconds: 600, timer: null, loading: false,
      showPdf: true // 是否允许查看报告
    }
  },
  computed: {
    currentQu() { return this.quList[this.currentIndex] || null },
    progressPct() { return this.quList.length ? ((this.currentIndex + 1) / this.quList.length * 100) : 0 },
    unansweredCount() { return this.quList.filter(q => !q.answered).length }
  },
  created() {
    this.paperId = this.$route.params.id
    this.testerId = this.$route.params.testerId
    this.loadPaper()
    // 加载考试配置（showPdf）
    this.loadExamConfig()
  },
  beforeDestroy() { if (this.timer) clearInterval(this.timer) },
  methods: {
    sizeClass(i) { return ['lg','md','sm','sm','md','lg'][i] },
    colorClass(i) { return i <= 2 ? 'ca' : 'cb' },
    selectCircle(i) {
      this.selectedCircle = i
      this.scoreA = CIRCLE_A[i]
      this.scoreB = CIRCLE_B[i]
      this.autoSave()
    },
    clearScore() {
      this.selectedCircle = -1
      this.scoreA = -1
      this.scoreB = -1
      const q = this.quList[this.currentIndex]
      if (q) { q.answered = false; q.scoreA = -1; q.scoreB = -1 }
    },
    scoreToCircle(a) { const i = CIRCLE_A.indexOf(a); return i >= 0 ? i : -1 },
    loadPaper() {
      request.post('/exam/api/mbti/paper-detail', { paperId: this.paperId }).then(res => {
        this.quList = res.data.quList || []
        this.leftSeconds = res.data.leftSeconds || 600
        // 自动定位到第一个未答题
        const firstUnanswered = this.quList.findIndex(q => !q.answered)
        if (firstUnanswered > 0) {
          this.currentIndex = firstUnanswered
        }
        this.syncCurrent()
        this.startTimer()
      })
    },
    syncCurrent() {
      const q = this.quList[this.currentIndex]
      if (q && q.answered) {
        this.scoreA = q.scoreA; this.scoreB = q.scoreB
        this.selectedCircle = this.scoreToCircle(q.scoreA)
      } else {
        this.scoreA = -1; this.scoreB = -1; this.selectedCircle = -1
      }
    },
    autoSave() {
      if (this.scoreA < 0 || this.scoreB < 0 || this.scoreA + this.scoreB !== 5) return
      const q = this.quList[this.currentIndex]
      request.post('/exam/api/mbti/fill-answer', {
        paperId: this.paperId, quId: q.quId, scoreA: this.scoreA, scoreB: this.scoreB
      }).then(() => { q.answered = true; q.scoreA = this.scoreA; q.scoreB = this.scoreB })
    },
    prev() { if (this.currentIndex > 0) { this.currentIndex--; this.syncCurrent() } },
    next() { if (this.currentIndex < this.quList.length - 1) { this.currentIndex++; this.syncCurrent() } },
    goTo(i) { this.currentIndex = i; this.syncCurrent() },
    goToFirstUnanswered() { const i = this.quList.findIndex(q => !q.answered); if (i >= 0) this.goTo(i) },
    handleSubmit() {
      const n = this.unansweredCount
      // FB-040 修复：MBTI 必须答完全部题目才能交卷，与心理特质/管理特质行为一致。
      if (n > 0) {
        this.$confirm(`还有 ${n} 题未作答，请全部作答完毕后再交卷！`, '提示', {
          confirmButtonText: '定位到未作答',
          cancelButtonText: '继续答题',
          type: 'warning'
        }).then(() => this.goToFirstUnanswered()).catch(() => {})
        return
      }
      this.$confirm('确定交卷吗？', '提示', {
        confirmButtonText: '确定交卷', cancelButtonText: '继续答题', type: 'info'
      }).then(() => this.doSubmit()).catch(() => {})
    },
    doSubmit() {
      this.loading = true
      request.post('/exam/api/mbti/submit', { paperId: this.paperId }).then(res => {
        if (this.showPdf) {
          this.$notify({ title: '答题已完成！', message: `你的 MBTI 类型：${res.data.type}`, type: 'success', duration: 5000 })
          this.$router.replace({ name: 'MbtiResult', params: { id: this.paperId, testerId: this.testerId } })
        } else {
          // 跳转到感谢页面
          this.$router.replace({ name: 'ExamThankYou' })
        }
      }).finally(() => { this.loading = false })
    },
    loadExamConfig() {
      // 通过 paper 获取 examId，再查 exam detail 的 showPdf
      request.post('/exam/api/paper/paper/detail', { id: this.paperId }).then(res => {
        if (res.data && res.data.examId) {
          request.post('/exam/api/exam/exam/detail', { id: res.data.examId }).then(r2 => {
            this.showPdf = r2.data?.showPdf !== false && r2.data?.showPdf !== 0
          }).catch(() => {})
        }
      }).catch(() => {})
    },
    startTimer() {
      this.timer = setInterval(() => {
        this.leftSeconds--
        if (this.leftSeconds === 120) this.$notify({ title: '时间提醒', message: '答题时间剩余 2 分钟', type: 'warning', duration: 5000 })
        if (this.leftSeconds <= 0) { clearInterval(this.timer); this.$alert('答题时间已到', '时间到').then(() => this.doSubmit()) }
      }, 1000)
    },
    formatTime(s) { return s <= 0 ? '00:00' : `${String(Math.floor(s/60)).padStart(2,'0')}:${String(s%60).padStart(2,'0')}` }
  }
}
</script>

<style scoped>
.mbti-exam { max-width: 720px; margin: 0 auto; padding: 16px; min-height: 100vh; background: #f5f7fa; display: flex; flex-direction: column; box-sizing: border-box; }
.mbti-header { margin-bottom: 16px; }
.progress-bar { height: 4px; background: #e4e7ed; border-radius: 2px; overflow: hidden; }
.progress-fill { height: 100%; background: #409EFF; transition: width 0.3s; }
.header-info { display: flex; justify-content: space-between; margin-top: 6px; font-size: 13px; color: #909399; }
.timer { font-weight: 700; }
.timer.warning { color: #F56C6C; animation: pulse 1s infinite; }
@keyframes pulse { 50% { opacity: 0.5; } }

.question-card { background: #fff; border-radius: 16px; padding: 28px 24px; box-shadow: 0 2px 20px rgba(0,0,0,0.05); flex: 1; display: flex; flex-direction: column; justify-content: center; margin-bottom: 12px; }
.qu-title { font-size: 22px; font-weight: 600; line-height: 1.7; color: #1a1a2e; text-align: center; margin-bottom: 24px; min-height: 44px; }

/* 选项文字（上方） */
.opts-top { display: flex; gap: 12px; margin-bottom: 32px; min-height: 90px; }
.opt-card { flex: 1; padding: 14px 16px; border-radius: 12px; background: #f9fafb; border: 2px solid #ebeef5; display: flex; align-items: flex-start; gap: 8px; transition: background 0.2s, border-color 0.2s; min-height: 60px; }
.opt-tag { display: inline-flex; align-items: center; justify-content: center; min-width: 28px; height: 28px; border-radius: 6px; font-size: 14px; font-weight: 800; color: #fff; flex-shrink: 0; }
.opt-a .opt-tag { background: #409EFF; }
.opt-b .opt-tag { background: #E6A23C; }
.opt-text { font-size: 15px; line-height: 1.5; color: #303133; }
.opt-card.opt-a.hi { background: #ecf5ff; border-color: #409EFF; }
.opt-card.opt-b.hi { background: #fdf6ec; border-color: #E6A23C; }

.picker-row { display: flex; align-items: center; justify-content: center; gap: 16px; margin-bottom: 12px; }
.side-label { font-size: 28px; font-weight: 900; color: #c0c4cc; transition: color 0.2s; min-width: 32px; text-align: center; }
.side-a.win { color: #409EFF; }
.side-b.win { color: #E6A23C; }
.circles { display: flex; align-items: center; gap: 16px; }
.circle { border-radius: 50%; border: 3px solid #dcdfe6; cursor: pointer; transition: background 0.2s, border-color 0.2s; display: flex; align-items: center; justify-content: center; box-sizing: border-box; }
.circle:hover { border-color: #909399; background: #f5f7fa; }
.circle.lg { width: 56px; height: 56px; }
.circle.md { width: 44px; height: 44px; }
.circle.sm { width: 36px; height: 36px; }
.circle.ca { border-color: #a0cfff; }
.circle.cb { border-color: #f3d19e; }
.circle.cn { border-color: #c0c4cc; }
.circle.ca:hover { border-color: #409EFF; background: #ecf5ff; }
.circle.cb:hover { border-color: #E6A23C; background: #fdf6ec; }
.circle.cn:hover { border-color: #909399; background: #f5f7fa; }
.circle.selected.ca { border-color: #409EFF; background: #409EFF; }
.circle.selected.cb { border-color: #E6A23C; background: #E6A23C; }
.circle.selected.cn { border-color: #909399; background: #909399; }
.fill { width: 40%; height: 40%; border-radius: 50%; background: #fff; position: absolute; }
.circle { position: relative; }

.hint { text-align: center; margin-top: 8px; font-size: 13px; color: #c0c4cc; }
.hint.picked { color: #909399; }

.nav-bar { display: flex; align-items: center; gap: 8px; padding: 8px 0; }
.nb { padding: 10px 20px; border: 1px solid #dcdfe6; background: #fff; border-radius: 8px; font-size: 14px; cursor: pointer; color: #606266; transition: all 0.2s; white-space: nowrap; }
.nb:disabled { opacity: 0.4; cursor: not-allowed; }
.nb:hover:not(:disabled) { border-color: #409EFF; color: #409EFF; }
.nb.primary { background: #409EFF; color: #fff; border-color: #409EFF; }
.nb.success { background: #67C23A; color: #fff; border-color: #67C23A; }
.dots { display: flex; flex-wrap: wrap; gap: 4px; flex: 1; justify-content: center; }
.d { width: 10px; height: 10px; border-radius: 50%; background: #dcdfe6; cursor: pointer; transition: all 0.15s; }
.d.ok { background: #67C23A; }
.d.cur { background: #409EFF; transform: scale(1.4); }
.unanswered { text-align: center; padding: 8px; color: #409EFF; font-size: 14px; cursor: pointer; font-weight: 500; }

/* ===== 手机竖屏 ===== */
@media (max-width: 500px) {
  .mbti-exam { padding: 10px; }
  .question-card { padding: 20px 14px; border-radius: 12px; }
  .qu-title { font-size: 17px; margin-bottom: 16px; }
  .opts-top { flex-direction: column; gap: 8px; margin-bottom: 24px; }
  .opt-card { padding: 12px; }
  .opt-text { font-size: 14px; }
  .picker-row { gap: 8px; }
  .circles { gap: 8px; }
  .circle.lg { width: 46px; height: 46px; }
  .circle.md { width: 36px; height: 36px; }
  .circle.sm { width: 28px; height: 28px; }
  .side-label { font-size: 20px; min-width: 24px; }
  .nb { padding: 8px 14px; font-size: 13px; }
  .dots { gap: 3px; }
  .d { width: 8px; height: 8px; }
}

/* ===== PC 宽屏 ===== */
@media (min-width: 768px) {
  .question-card { min-height: 400px; padding: 40px 36px; }
  .qu-title { font-size: 26px; margin-bottom: 28px; }
  .opts-top { gap: 20px; margin-bottom: 40px; }
  .opt-card { padding: 18px 20px; }
  .opt-text { font-size: 16px; }
  .circles { gap: 28px; }
  .circle.lg { width: 68px; height: 68px; border-width: 4px; }
  .circle.md { width: 52px; height: 52px; }
  .circle.sm { width: 42px; height: 42px; }
  .side-label { font-size: 36px; min-width: 44px; }
}
</style>

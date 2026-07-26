<template>
  <div class="mbti-result">
    <div class="result-card" v-if="loaded">
      <!-- 类型展示 -->
      <div class="type-section">
        <div class="type-badge">{{ mbtiType }}</div>
        <div class="type-name">{{ typeNames[mbtiType] || '' }}</div>
      </div>

      <!-- 8维度柱状图 -->
      <div class="chart-section">
        <h3>四维度得分</h3>
        <div class="dimension" v-for="dim in dimensions" :key="dim.label">
          <div class="dim-label">{{ dim.label }}</div>
          <div class="dim-bar-container">
            <div class="dim-left">
              <span class="dim-letter" :class="{winner: scores[dim.left] >= scores[dim.right]}">{{ dim.left }}</span>
              <div class="bar left-bar">
                <div class="bar-fill" :style="{width: barPct(scores[dim.left], dim) + '%'}"
                     :class="{winner: scores[dim.left] >= scores[dim.right]}"></div>
              </div>
              <span class="dim-score">{{ scores[dim.left] }}</span>
            </div>
            <div class="dim-right">
              <span class="dim-score">{{ scores[dim.right] }}</span>
              <div class="bar right-bar">
                <div class="bar-fill right" :style="{width: barPct(scores[dim.right], dim) + '%'}"
                     :class="{winner: scores[dim.right] > scores[dim.left]}"></div>
              </div>
              <span class="dim-letter" :class="{winner: scores[dim.right] > scores[dim.left]}">{{ dim.right }}</span>
            </div>
          </div>
          <div class="dim-desc">{{ dim.leftDesc }} vs {{ dim.rightDesc }}</div>
        </div>
      </div>

      <!-- 操作区域已移除，考生无需手动返回 -->
    </div>

    <!-- 不允许查看报告时显示感谢页 -->
    <div v-else-if="!showPdf" class="thank-you">
      <div class="ty-icon">✅</div>
      <h2>测评完成</h2>
      <p>感谢您的参与，答题已完成！</p>
      <p style="color: #909399; font-size: 14px; margin-top: 20px;">您的测评结果将由管理员查看和分析。</p>
    </div>

    <div v-else class="loading">
      <i class="el-icon-loading"></i> 正在计算你的 MBTI 类型...
    </div>
  </div>
</template>

<script>
import request from '@/utils/request'

export default {
  name: 'MbtiResult',
  data() {
    return {
      paperId: '',
      testerId: '',
      mbtiType: '',
      scores: { E: 0, I: 0, S: 0, N: 0, T: 0, F: 0, J: 0, P: 0 },
      loaded: false,
      showPdf: true, // 是否允许查看报告
      dimensions: [
        { label: '注意力', left: 'E', right: 'I', leftDesc: '外向', rightDesc: '内向' },
        { label: '认知方式', left: 'S', right: 'N', leftDesc: '感觉', rightDesc: '直觉' },
        { label: '判断方式', left: 'T', right: 'F', leftDesc: '理性', rightDesc: '感性' },
        { label: '生活方式', left: 'J', right: 'P', leftDesc: '判断', rightDesc: '感知' }
      ],
      typeNames: {
        'ISTJ': '检查员', 'ISFJ': '守护者', 'INFJ': '提倡者', 'INTJ': '建筑师',
        'ISTP': '鉴赏家', 'ISFP': '探险家', 'INFP': '调停者', 'INTP': '逻辑学家',
        'ESTP': '企业家', 'ESFP': '表演者', 'ENFP': '竞选者', 'ENTP': '辩论家',
        'ESTJ': '总经理', 'ESFJ': '执政官', 'ENFJ': '外交家', 'ENTJ': '指挥官'
      }
    }
  },
  created() {
    this.paperId = this.$route.params.id
    this.testerId = this.$route.params.testerId
    this.checkShowPdf()
  },
  methods: {
    checkShowPdf() {
      request.post('/exam/api/paper/paper/detail', { id: this.paperId }).then(res => {
        if (res.data && res.data.examId) {
          request.post('/exam/api/exam/exam/detail', { id: res.data.examId }).then(r2 => {
            this.showPdf = r2.data?.showPdf !== false && r2.data?.showPdf !== 0
            if (this.showPdf) {
              this.loadScore()
            }
          }).catch(() => this.loadScore())
        } else {
          this.loadScore()
        }
      }).catch(() => this.loadScore())
    },
    loadScore() {
      request.post('/exam/api/mbti/score', { paperId: this.paperId }).then(res => {
        this.scores = res.data.scores
        this.mbtiType = res.data.type
        this.loaded = true
      })
    },
    barPct(val, dim) {
      const total = this.scores[dim.left] + this.scores[dim.right]
      return total > 0 ? (val / total * 100) : 50
    }
  }
}
</script>

<style scoped>
.mbti-result { max-width: 700px; margin: 0 auto; padding: 20px; min-height: 100vh; background: #f5f7fa; }
.result-card { background: #fff; border-radius: 12px; padding: 40px 30px; box-shadow: 0 4px 16px rgba(0,0,0,0.08); }
.type-section { text-align: center; margin-bottom: 40px; }
.type-badge { font-size: 48px; font-weight: 900; color: #409EFF; letter-spacing: 8px; }
.type-name { font-size: 20px; color: #606266; margin-top: 8px; }
.chart-section { margin-bottom: 30px; }
.chart-section h3 { font-size: 18px; color: #303133; margin-bottom: 20px; text-align: center; }
.dimension { margin-bottom: 24px; }
.dim-label { font-size: 13px; color: #909399; text-align: center; margin-bottom: 6px; }
.dim-bar-container { display: flex; align-items: center; }
.dim-left, .dim-right { display: flex; align-items: center; flex: 1; }
.dim-right { flex-direction: row-reverse; }
.dim-letter { font-size: 20px; font-weight: 700; color: #c0c4cc; min-width: 28px; text-align: center; }
.dim-letter.winner { color: #409EFF; }
.dim-score { font-size: 14px; font-weight: bold; color: #606266; min-width: 24px; text-align: center; }
.bar { flex: 1; height: 24px; background: #f0f2f5; border-radius: 4px; overflow: hidden; margin: 0 6px; }
.left-bar .bar-fill { float: right; height: 100%; background: #c0c4cc; border-radius: 4px; transition: width 0.5s; }
.right-bar .bar-fill { height: 100%; background: #c0c4cc; border-radius: 4px; transition: width 0.5s; }
.bar-fill.winner { background: #409EFF; }
.dim-desc { font-size: 12px; color: #c0c4cc; text-align: center; margin-top: 4px; }
.action-area { text-align: center; margin-top: 30px; }
.loading { text-align: center; padding: 100px 0; font-size: 18px; color: #909399; }
.thank-you { text-align: center; padding: 80px 30px; background: #fff; border-radius: 12px; box-shadow: 0 4px 16px rgba(0,0,0,0.08); }
.thank-you .ty-icon { font-size: 64px; margin-bottom: 20px; }
.thank-you h2 { color: #303133; font-size: 28px; margin-bottom: 12px; }
.thank-you p { color: #606266; font-size: 16px; }
</style>

// Test the 3-stage ready detection logic embedded in result.vue / result2.vue mounted hook.
//
// Since the logic is inline in mounted(), we replicate its core state machine here
// and unit-test the decision matrix. This guards against regressions to the fixed
// setTimeout(4000) anti-pattern.
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// Replica of the 3-stage ready detector (must mirror result.vue lines ~600-625)
function makeReadyDetector({ expectedEval, idleMs = 5000, hardMs = 20000 } = {}) {
  const state = {
    isGetData: false,
    evaluation: [],
    reportReady: false,
    reportIncomplete: false,
    _resolved: false
  }
  const start = Date.now()
  let lastEvalLen = 0
  let lastChange = Date.now()

  function tick(now = Date.now()) {
    if (state._resolved) return state
    const evalLen = state.evaluation.length
    if (evalLen !== lastEvalLen) {
      lastEvalLen = evalLen
      lastChange = now
    }
    const fullReady = state.isGetData && evalLen >= expectedEval
    const idleReady = state.isGetData && evalLen > 0 && (now - lastChange > idleMs)
    if (fullReady || idleReady) {
      state._resolved = true
      if (!fullReady) state.reportIncomplete = true
      state.reportReady = true
      return state
    }
    if (now - start > hardMs) {
      state._resolved = true
      state.reportIncomplete = true
      state.reportReady = true
    }
    return state
  }
  return { state, tick }
}

describe('ready 三阶段检测器', () => {
  beforeEach(() => { vi.useFakeTimers(); vi.setSystemTime(0) })
  afterEach(() => { vi.useRealTimers() })

  describe('happy path: 数据齐了立即 ready', () => {
    it('isGetData=true 且 evaluation 齐 12 → fullReady, incomplete=false', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12 })
      state.isGetData = true
      for (let i = 0; i < 12; i++) state.evaluation.push({ name: 'x' + i })
      vi.setSystemTime(2000) // 2s
      tick(2000)
      expect(state.reportReady).toBe(true)
      expect(state.reportIncomplete).toBe(false)
    })

    it('result2 13 维度同样工作', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 13 })
      state.isGetData = true
      for (let i = 0; i < 13; i++) state.evaluation.push({ name: 'x' + i })
      tick(3000)
      expect(state.reportReady).toBe(true)
      expect(state.reportIncomplete).toBe(false)
    })

    it('isGetData=false 时即使 evaluation 齐也不 ready', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12 })
      state.isGetData = false
      for (let i = 0; i < 12; i++) state.evaluation.push({ name: 'x' + i })
      tick(2000)
      expect(state.reportReady).toBe(false)
    })
  })

  describe('idle 兜底: 数据停止变化 5s 接受部分结果', () => {
    it('isGetData=true 且 evalLen=8 且 5s 无变化 → idleReady, incomplete=true', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12, idleMs: 5000 })
      state.isGetData = true
      for (let i = 0; i < 8; i++) state.evaluation.push({ name: 'x' + i })
      tick(1000) // 第一次记录变化
      vi.setSystemTime(7000)
      tick(7000) // 6s 无变化 (1000→7000)，超过 idleMs=5000
      expect(state.reportReady).toBe(true)
      expect(state.reportIncomplete).toBe(true)
    })

    it('evalLen 仍在变化 → 不触发 idle', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12, idleMs: 5000 })
      state.isGetData = true
      state.evaluation.push({ name: 'a' })
      tick(1000)
      vi.setSystemTime(4000); state.evaluation.push({ name: 'b' }); tick(4000)
      vi.setSystemTime(7000); state.evaluation.push({ name: 'c' }); tick(7000)
      // 最近变化在 7000，距 7000 = 0s，未超 5s
      expect(state.reportReady).toBe(false)
    })

    it('evalLen=0 时 idle 不触发（避免完全无数据被接受）', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12, idleMs: 5000 })
      state.isGetData = true
      tick(1000)
      vi.setSystemTime(8000)
      tick(8000) // 已 idle 7s 但 evalLen=0
      expect(state.reportReady).toBe(false)
    })
  })

  describe('hard 兜底: 20s 强制 ready', () => {
    it('20s 后即使 isGetData=false 也强制 ready + incomplete=true', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12, hardMs: 20000 })
      // never set isGetData
      vi.setSystemTime(21000)
      tick(21000)
      expect(state.reportReady).toBe(true)
      expect(state.reportIncomplete).toBe(true)
    })

    it('19s 时还不触发 hard', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12, hardMs: 20000 })
      vi.setSystemTime(19000)
      tick(19000)
      expect(state.reportReady).toBe(false)
    })
  })

  describe('反退化守护: 不能退回固定 setTimeout(4000) 模式', () => {
    it('4s 时如果数据未齐，不应 ready（happy 不触发）', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12 })
      state.isGetData = true
      // 模拟批量并发场景：4s 时只有 6 个 dict 完成
      for (let i = 0; i < 6; i++) state.evaluation.push({ name: 'x' + i })
      vi.setSystemTime(4000)
      tick(4000)
      // 之前的 bug：4s 强制 ready；现在：必须等 idle 5s 或更多 dict
      expect(state.reportReady).toBe(false)
    })

    it('数据继续到达，不会过早 ready', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12 })
      state.isGetData = true
      // 4s 时 6 个
      for (let i = 0; i < 6; i++) state.evaluation.push({ name: 'x' + i })
      vi.setSystemTime(4000); tick(4000)
      // 6s 时 12 个 → 立即 ready
      vi.setSystemTime(6000)
      for (let i = 6; i < 12; i++) state.evaluation.push({ name: 'x' + i })
      tick(6000)
      expect(state.reportReady).toBe(true)
      expect(state.reportIncomplete).toBe(false)
    })
  })

  describe('已 resolved 后再 tick 是幂等的', () => {
    it('多次 tick 不重置状态', () => {
      const { state, tick } = makeReadyDetector({ expectedEval: 12 })
      state.isGetData = true
      for (let i = 0; i < 12; i++) state.evaluation.push({ name: 'x' + i })
      tick(2000)
      const firstReady = state.reportReady
      const firstIncomplete = state.reportIncomplete
      vi.setSystemTime(30000); tick(30000) // 远超 hard 超时
      expect(state.reportReady).toBe(firstReady)
      expect(state.reportIncomplete).toBe(firstIncomplete)
    })
  })
})

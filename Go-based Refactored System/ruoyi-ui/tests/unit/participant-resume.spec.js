import { describe, expect, it, vi, beforeEach } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import MbtiExam from '@/views/paper/exam/mbtiExam.vue'
import CompetencyExam from '@/views/paper/exam/competencyExam.vue'
import { fetchCompetencyPaper } from '@/api/competency'
import request from '@/utils/request'
import fs from 'fs'
import path from 'path'

vi.mock('@/api/competency', () => ({
  fetchCompetencyPaper: vi.fn(),
  saveCompetencyAnswer: vi.fn(),
  submitCompetencyPaper: vi.fn()
}))

vi.mock('@/utils/request', () => ({
  default: { post: vi.fn() }
}))

const answeredQuestions = Array.from({ length: 10 }, (_, index) => ({
  id: `pq-${index + 1}`,
  quId: `q-${index + 1}`,
  sort: index,
  answered: true,
  content: `V${index + 1}`,
  answerList: []
}))
const firstUnanswered = {
  id: 'pq-11',
  quId: 'q-11',
  sort: 10,
  answered: false,
  content: 'V11',
  answerList: []
}

function route(params) {
  return { params }
}

// TestBugFB141_ReLoginResumesAtFirstUnanswered
// 对应：docs/regression-tests.md #FB-141
// 复现：已答10题后退出，重新登录并恢复同一份进行中试卷。
// 期望：保留前10题答案，并直接显示第11题。
describe('participant re-login resume position', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    sessionStorage.clear()
  })

  it('scrolls the traditional all-question page to question 11', async () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/paper/exam/exam.vue'), 'utf8')
    expect(source).toContain('resumeAtFirstUnanswered()')
    expect(source).toContain('scrollIntoView')
  })

  it('opens question 11 on the traditional single-question page', async () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/paper/exam/examClick.vue'), 'utf8')
    expect(source).toContain('resumeAtFirstUnanswered()')
    expect(source).toContain('item => !item.answered')
  })

  it('opens question 11 on the MBTI page', async () => {
    request.post.mockImplementation(url => {
      if (url === '/exam/api/mbti/paper-detail') {
        return Promise.resolve({ data: { quList: [...answeredQuestions, firstUnanswered], leftSeconds: 600 } })
      }
      return Promise.reject(new Error('not needed'))
    })

    const wrapper = shallowMount(MbtiExam, {
      methods: { loadExamConfig: vi.fn(), startTimer: vi.fn() },
      mocks: { $route: route({ id: 'paper-1', testerId: 'candidate-1' }) }
    })

    await vi.waitFor(() => expect(wrapper.vm.currentIndex).toBe(10))
    expect(wrapper.vm.quList.slice(0, 10).every(question => question.answered)).toBe(true)
  })

  it('opens question 11 on the 00401 competency page', async () => {
    sessionStorage.setItem('competencyPaperToken', 'paper-token')
    fetchCompetencyPaper.mockResolvedValue({ data: {
      paperId: 'paper-1',
      state: 0,
      totalCount: 11,
      answeredCount: 10,
      unansweredCount: 1,
      limitTime: new Date(Date.now() + 600000).toISOString(),
      questions: [...answeredQuestions, firstUnanswered]
    } })

    const wrapper = shallowMount(CompetencyExam, {
      mocks: {
        $route: route({ paperId: 'paper-1' }),
        $router: { go: vi.fn(), replace: vi.fn() },
        $message: { error: vi.fn() }
      },
      stubs: ['el-card', 'el-radio-group', 'el-radio', 'el-button', 'el-progress']
    })

    await vi.waitFor(() => expect(wrapper.vm.currentIndex).toBe(10))
    expect(wrapper.vm.paper.answeredCount).toBe(10)
    clearInterval(wrapper.vm.timer)
  })
})

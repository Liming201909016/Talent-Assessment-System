import { describe, expect, it, vi } from 'vitest'
import Preview from '@/views/paper/exam/preview.vue'
import { createCompetencyPaper } from '@/api/competency'

vi.mock('@/api/competency', () => ({ createCompetencyPaper: vi.fn() }))
vi.mock('@/api/exam/exam', () => ({ fetchDetail: vi.fn() }))
vi.mock('@/api/paper/exam', () => ({ createPaper: vi.fn(), paperDetail: vi.fn() }))
vi.mock('@/api/candidate/candidate', () => ({ updateData: vi.fn() }))
vi.mock('@/api/tester/tester', () => ({ getTesterByIdNumber: vi.fn(() => Promise.resolve({ data: {} })), updateTester: vi.fn() }))

describe('competency start gate', () => {
  it('disables start and shows Chinese guidance while the competency exam is unpublished', async () => {
    createCompetencyPaper.mockClear()
    const warning = vi.fn()
    const vm = {
      ...Preview.data(),
      examId: 'e1', testerId: 'p1',
      detailData: { assessmentType: 'competency', publishStatus: 0, competencyDimensions: [{ questionCount: 8 }, { questionCount: 8 }] },
      $message: { warning },
      $router: { replace: vi.fn() }
    }
    expect(Preview.computed.canStart.call(vm)).toBe(false)
    const reason = Preview.computed.startDisabledReason.call({ ...vm, canStart: false })
    expect(reason).toContain('尚未发布')
    const description = Preview.computed.displayDesc.call(vm)
    expect(description).toContain('本测评旨在全面了解您在各项工作情景中的行为表现与内在倾向')
    expect(description).toContain('作答说明：')
    expect(description).toContain('所有题目无对错、好坏之分')
    expect(description).toContain('请依据您的第一反应如实作答')
    expect(description).toContain('您的作答结果将严格保密')
    expect(description).toContain('本次测验共计90道题目，每道题目都必须作答')
    expect(description).not.toContain('16道题目')
    vm.canStart = false
    vm.startDisabledReason = reason
    await Preview.methods.handleCreate.call(vm)
    expect(createCompetencyPaper).not.toHaveBeenCalled()
    expect(warning).toHaveBeenCalledWith(expect.stringContaining('尚未发布'))
  })
})
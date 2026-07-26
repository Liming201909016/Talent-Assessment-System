import fs from 'fs'
import path from 'path'
import { describe, expect, it } from 'vitest'

const source = fs.readFileSync(path.resolve(__dirname, '../../src/views/exam/exam/index.vue'), 'utf8')

describe('competency dynamic export entry', () => {
  it('shows both existing export actions for competency exams', () => {
    expect(source).toContain('command="competencyResults"')
    expect(source).toContain('command="export"')
    expect(source).toContain('command="exportAnswers"')
    expect(source).not.toMatch(/<template v-else>[\s\S]*command="export"[\s\S]*command="exportAnswers"/)
  })

  it('uses explicit competency download filenames and confirmations', () => {
    expect(source).toContain("row.assessmentType === 'competency'")
    expect(source).toContain('胜任力结果明细.xlsx')
    expect(source).toContain('结果汇总、逐题明细和题目字典')
  })
})

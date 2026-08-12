import { describe, expect, it } from 'vitest'
import fs from 'fs'
import path from 'path'

const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/form.vue'), 'utf8')

describe('competency version form', () => {
  it('round-trips all four version fields and displays the frozen configuration', () => {
    for (const field of [
      'competencyProductVersion',
      'competencyScoringVersion',
      'competencyContentVersion',
      'competencyReportTemplateVersion'
    ]) {
      expect(source).toContain(field)
    }
    expect(source).toContain('version-summary')
    expect(source).toContain('发布后与题目快照一并冻结')
  })

  it('clears competency versions when switching back to legacy', () => {
    const changeStart = source.indexOf('handleAssessmentTypeChange(value)')
    const changeEnd = source.indexOf('\n    loadCompetencyDimensions() {', changeStart)
    const body = source.slice(changeStart, changeEnd)
    for (const field of [
      'competencyProductVersion',
      'competencyScoringVersion',
      'competencyContentVersion',
      'competencyReportTemplateVersion'
    ]) {
      expect(body).toContain(`this.postForm.${field} = ''`)
    }
  })

  it('renders the phase-1 profile as fixed read-only configuration', () => {
    expect(source).toContain('00401 一期固定配置')
    expect(source).toContain('基层员工 · 10个A/B维度 · 90题')
    expect(source).toContain('competency-frontline-phase1-v1')
    expect(source).toContain('competency-phase1-scoring-v1')
    expect(source).toContain('competency-phase1-content-v1')
    expect(source).toContain('competency-phase1-report-v1')
    expect(source).not.toContain('label="leader"')
    expect(source).not.toContain('<competency-dimension-selector')
  })

  it('applies fixed phase-1 fields and enables publish after runtime completion', () => {
    const changeStart = source.indexOf('handleAssessmentTypeChange(value)')
    const changeEnd = source.indexOf('\n    loadCompetencyDimensions() {', changeStart)
    const body = source.slice(changeStart, changeEnd)
    expect(body).toContain('applyPhase1Profile()')
    expect(source).not.toContain('phase1PublishBlocked')
    expect(source).not.toContain('一期五档、一级维度和效度运行时完成后方可发布')
  })
})

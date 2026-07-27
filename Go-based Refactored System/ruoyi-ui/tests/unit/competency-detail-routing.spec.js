import { describe, expect, it } from 'vitest'
import fs from 'fs'
import path from 'path'

const examListSource = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/index.vue'), 'utf8')
const legacyDetailSource = fs.readFileSync(path.resolve(process.cwd(), 'src/views/user/exam/index.vue'), 'utf8')
const dashboardSource = fs.readFileSync(path.resolve(process.cwd(), 'src/views/index.vue'), 'utf8')
const resultButtonsE2E = fs.readFileSync(path.resolve(process.cwd(), '../../scripts/test/staging-competency-result-buttons-e2e.js'), 'utf8')

describe('FB-075 competency detail routing', () => {
  it('opens competency detail in the dedicated result page', () => {
    expect(examListSource).toMatch(/handleExamDetail\(row\)\s*{[\s\S]*row\.assessmentType === ['"]competency['"][\s\S]*name:\s*['"]CompetencyResults['"]/)
  })

  it('keeps legacy detail on the participant/report page', () => {
    expect(examListSource).toMatch(/handleExamDetail\(row\)\s*{[\s\S]*name:\s*['"]ListExamUser['"][\s\S]*examId:\s*row\.id/)
  })

  it('keeps repoCode-dependent legacy controls out of the competency result page', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/views/exam/exam/competencyResults.vue'), 'utf8')
    expect(source).not.toContain('/exam/api/exam/exam/generate-report')
    expect(source).not.toContain('查看团队报告')
    expect(source).not.toContain('删除报告')
    expect(source).toContain('generateCompetencyReport')
    expect(source).toContain('downloadCompetencyReport')
  })

  // FB-076: docs/regression-tests.md
  it('redirects stale competency legacy-detail URLs before loading legacy participants', () => {
    expect(legacyDetailSource).toContain("import {fetchDetail} from '@/api/exam/exam'")
    expect(legacyDetailSource).toMatch(/created\(\)\s*{[\s\S]*redirectCompetencyDetail\(\)/)
    expect(legacyDetailSource).toMatch(/async redirectCompetencyDetail\(\)[\s\S]*fetchDetail\(this\.\$route\.params\.examId\)[\s\S]*assessmentType[^\n]*['"]competency['"][\s\S]*replace\([\s\S]*name:\s*['"]CompetencyResults['"]/)
    expect(legacyDetailSource).toMatch(/if\s*\(await this\.redirectCompetencyDetail\(\)\)\s*return[\s\S]*this\.getList/)
  })

  // FB-076: docs/regression-tests.md
  it('routes recent competency exams from the dashboard to the dedicated result page', () => {
    expect(dashboardSource).toMatch(/goExamDetail\(row\)\s*{[\s\S]*row\.assessmentType === ['"]competency['"][\s\S]*name:\s*['"]CompetencyResults['"][\s\S]*return/)
    expect(dashboardSource).toMatch(/goExamDetail\(row\)[\s\S]*path:\s*`\/exam\/exam\/users/)
  })

  // FB-086: docs/regression-tests.md
  it('reports the actual legacy-control assertion count instead of a stale hard-coded total', () => {
    expect(resultButtonsE2E).toContain('const legacyOnlyControls =')
    expect(resultButtonsE2E).toContain('legacy_controls_hidden=${legacyOnlyControls.length}/${legacyOnlyControls.length}')
    expect(resultButtonsE2E).not.toContain("legacy_controls_hidden=9/9")
  })
})

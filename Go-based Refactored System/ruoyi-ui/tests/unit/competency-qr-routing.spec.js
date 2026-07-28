import { describe, expect, it } from 'vitest'
import fs from 'fs'
import path from 'path'
import { buildExamEntryURL, resolveExamRepoCode } from '@/views/paper/exam/repoCode'

const onlineExamSource = fs.readFileSync(path.resolve(process.cwd(), 'src/views/paper/exam/list.vue'), 'utf8')

describe('FB-102 competency QR routing', () => {
  // FB-102: docs/regression-tests.md
  it('uses the virtual 00401 code when a competency exam has no physical repository', () => {
    expect(onlineExamSource).toContain("import { buildExamEntryURL, resolveExamRepoCode } from './repoCode'")
    expect(onlineExamSource).toMatch(/resolveRepoCode\(row\)\s*{\s*return resolveExamRepoCode\(row\)\s*}/)
    expect(onlineExamSource).toMatch(/creatQrCode\(row\)[\s\S]*QRCodeURL = buildExamEntryURL\(row, origin\)/)
    expect(resolveExamRepoCode({ assessmentType: 'competency', repoCode: '' })).toBe('00401')
    expect(resolveExamRepoCode({ assessmentType: 'legacy', repoCode: '00101' })).toBe('00101')
  })

  // FB-102: docs/regression-tests.md
  it('uses the same resolved code for direct open and closed exam navigation', () => {
    expect(onlineExamSource).toMatch(/handlePre\(row\)[\s\S]*const repoCode = this\.resolveRepoCode\(row\)[\s\S]*name:\s*['"]candidateInfo['"][\s\S]*repoCode/)
    expect(onlineExamSource).toMatch(/handlePre\(row\)[\s\S]*name:\s*['"]tester['"][\s\S]*repoCode/)

    const origin = 'http://localhost:9527'
    expect(buildExamEntryURL({ id: 'open-exam', stuFlag: 0, isOpen: 1, assessmentType: 'competency', repoCode: '' }, origin))
      .toBe('http://localhost:9527/#/my/exam/candidate/open-exam/0/00401')
    expect(buildExamEntryURL({ id: 'closed-exam', stuFlag: 0, isOpen: 2, assessmentType: 'competency', repoCode: '' }, origin))
      .toBe('http://localhost:9527/#/my/exam/tester/closed-exam/00401')
    expect(buildExamEntryURL({ id: 'legacy-open', stuFlag: 1, isOpen: 1, assessmentType: 'legacy', repoCode: '00102' }, origin))
      .toBe('http://localhost:9527/#/my/exam/candidate/legacy-open/1/00102')
    expect(buildExamEntryURL({ id: 'legacy-closed', stuFlag: 0, isOpen: 2, assessmentType: 'legacy', repoCode: '00202' }, origin))
      .toBe('http://localhost:9527/#/my/exam/tester/legacy-closed/00202')
  })

  // FB-102: docs/regression-tests.md
  it('does not call startsWith on an empty repository field in the online list', () => {
    expect(onlineExamSource).not.toContain('scope.row.repoCode.startsWith')
  })
})

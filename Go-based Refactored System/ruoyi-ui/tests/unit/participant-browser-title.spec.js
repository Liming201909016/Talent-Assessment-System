import { describe, expect, it } from 'vitest'
import fs from 'fs'
import path from 'path'
import App from '@/App.vue'

// TestBugFB144_PublicParticipantPagesHideSystemBrowserTitle
// 对应：docs/regression-tests.md #FB-144
// 复现：手机浏览器打开考生信息、准备、答题或完成页面。
// 期望：浏览器顶部不再显示“人才综合素质评估系统”，管理后台标题保持不变。
describe('public participant browser title', () => {
  it('renders an empty browser title only for routes that opt out of the system title', () => {
    const participant = App.metaInfo.call({
      $route: { meta: { hideSystemTitle: true } },
      $store: { state: { settings: { dynamicTitle: false, title: '' } } }
    })
    const administrator = App.metaInfo.call({
      $route: { meta: {} },
      $store: { state: { settings: { dynamicTitle: false, title: '' } } }
    })
    expect(participant.title).toBe('')
    expect(participant.titleTemplate('考生信息')).toBe('')
    expect(administrator.titleTemplate('测评管理')).toContain('测评管理')
    expect(administrator.titleTemplate('测评管理')).not.toBe('')
  })

  it('marks every participant entry, preparation, answering, result and completion route', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/router/index.js'), 'utf8')
    for (const name of [
      'ExamOnline', 'ExamThankYou', 'tester', 'candidateInfo', 'PreExam',
      'MbtiExam', 'CompetencyExam', 'MbtiResult', 'ShowExam', 'ShowMngExam',
      'Finish', 'StartExam', 'StartExamClick'
    ]) {
      const routeStart = source.indexOf(`name: '${name}'`)
      expect(routeStart, `${name} route`).toBeGreaterThan(-1)
      const routeEnd = source.indexOf('\n  },', routeStart)
      expect(source.slice(routeStart, routeEnd), `${name} title flag`).toContain('hideSystemTitle: true')
    }
  })

  it('does not remove the browser title from administrator routes', () => {
    const source = fs.readFileSync(path.resolve(process.cwd(), 'src/router/index.js'), 'utf8')
    const routeStart = source.indexOf("name: 'ListExam'")
    const routeEnd = source.indexOf('\n      },', routeStart)
    expect(source.slice(routeStart, routeEnd)).not.toContain('hideSystemTitle: true')
  })
})

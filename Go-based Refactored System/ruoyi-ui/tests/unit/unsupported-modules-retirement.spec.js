import { describe, expect, it } from 'vitest'
import fs from 'fs'
import path from 'path'

const exists = relative => fs.existsSync(path.resolve(process.cwd(), relative))
const read = relative => fs.readFileSync(path.resolve(process.cwd(), relative), 'utf8')

describe('FB-098/099 unsupported module retirement', () => {
  it('removes unsupported monitor and generator API/pages', () => {
    for (const file of [
      'src/api/monitor/online.js',
      'src/views/monitor/online/index.vue',
      'src/api/monitor/job.js',
      'src/api/monitor/jobLog.js',
      'src/views/monitor/job/index.vue',
      'src/views/monitor/job/log.vue',
      'src/api/monitor/cache.js',
      'src/views/monitor/cache/index.vue',
      'src/api/tool/gen.js',
      'src/views/tool/gen/index.vue',
      'src/views/tool/gen/importTable.vue',
      'src/views/tool/gen/genInfoForm.vue',
      'src/views/tool/gen/editTable.vue',
      'src/views/tool/gen/basicInfoForm.vue'
    ]) {
      expect(exists(file), file).toBe(false)
    }
  })

  it('removes hidden routes and four unconsumed user training wrappers', () => {
    const router = read('src/router/index.js')
    expect(router).not.toContain("path: '/monitor/job-log'")
    expect(router).not.toContain("path: '/tool/gen-edit'")
    for (const file of [
      'src/api/user/repo.js',
      'src/api/user/book.js',
      'src/views/user/repo.js',
      'src/views/user/book.js'
    ]) {
      expect(exists(file), file).toBe(false)
    }
  })

  it('keeps audit list wrappers/pages but removes unsupported mutations and exports', () => {
    for (const module of ['operlog', 'logininfor']) {
      const api = read(`src/api/monitor/${module}.js`)
      const page = read(`src/views/monitor/${module}/index.vue`)
      expect(api).toContain(`/monitor/${module}/list`)
      expect(api).not.toMatch(/export function (del|clean)/)
      expect(page).not.toMatch(/handle(Delete|Clean|Export)/)
      expect(page).not.toContain(`monitor:${module}:remove`)
      expect(page).not.toContain(`monitor:${module}:export`)
    }
  })

  it('keeps real server monitoring, form builder, and swagger pages', () => {
    for (const file of [
      'src/api/monitor/server.js',
      'src/views/monitor/server/index.vue',
      'src/views/tool/build/index.vue',
      'src/views/tool/swagger/index.vue'
    ]) {
      expect(exists(file), file).toBe(true)
    }
  })
})
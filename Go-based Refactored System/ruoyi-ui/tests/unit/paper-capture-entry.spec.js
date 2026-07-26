import { describe, expect, it } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Paper management capture entry', () => {
  // FB-068: 无后端接口、数据表、采集端和展示弹窗时，不应暴露无法工作的考试截图入口。
  it('does not expose the orphaned capture action', () => {
    const file = path.resolve(process.cwd(), 'src/views/paper/paper/index.vue')
    const source = fs.readFileSync(file, 'utf8')
    expect(source).not.toContain('listCaptures')
    expect(source).not.toContain('handleCapture')
    expect(source).not.toContain('考试截图')
  })
})
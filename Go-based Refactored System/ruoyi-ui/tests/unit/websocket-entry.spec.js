import { describe, expect, it } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Optional WebSocket bootstrap', () => {
  // FB-071: Go后端未注册/ws路由时，管理端不应默认发起必然失败且持续重连的连接。
  it('starts the navbar WebSocket only when explicitly enabled', () => {
    const file = path.resolve(process.cwd(), 'src/layout/components/Navbar.vue')
    const source = fs.readFileSync(file, 'utf8')
    expect(source).toContain("process.env.VUE_APP_SOCKET_ENABLED !== 'true'")
    expect(source.indexOf("VUE_APP_SOCKET_ENABLED !== 'true'")).toBeLessThan(source.indexOf('startWebSocket'))
  })
})

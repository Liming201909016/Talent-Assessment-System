import { describe, expect, it } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('User avatar static URL', () => {
  // FB-100: /profile is served by nginx, not by the /prod-api backend proxy.
  it('does not prefix persisted avatar URLs with the API base path', () => {
    const storeSource = fs.readFileSync(
      path.resolve(process.cwd(), 'src/store/modules/user.js'),
      'utf8'
    )
    const uploadSource = fs.readFileSync(
      path.resolve(process.cwd(), 'src/views/system/user/profile/userAvatar.vue'),
      'utf8'
    )

    expect(storeSource).not.toContain('process.env.VUE_APP_BASE_API + user.avatar')
    expect(uploadSource).not.toContain('process.env.VUE_APP_BASE_API + response.imgUrl')
    expect(storeSource).toContain(': user.avatar;')
    expect(uploadSource).toContain('this.options.img = response.imgUrl;')
  })
})

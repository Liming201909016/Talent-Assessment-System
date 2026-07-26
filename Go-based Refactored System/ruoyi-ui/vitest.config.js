// Vitest 配置 — 兼容 Vue 2 (vue 2.6.12 + element-ui 2.x)
//
// 使用：
//   npm install              # 首次安装新增的 devDeps
//   npm test                 # 运行所有测试
//   npm run test:watch       # 监听模式
//   npm run test:ui          # 浏览器 UI 界面
//   npm run test:coverage    # 生成覆盖率报告
//
// 测试文件位置：tests/unit/**/*.spec.js
import { defineConfig } from 'vitest/config'
import { createVuePlugin as vue } from 'vite-plugin-vue2'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/unit/setup.js'],
    include: ['tests/unit/**/*.spec.{js,ts}'],
    exclude: ['node_modules', 'dist', '.git'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'json-summary'],
      reportsDirectory: './coverage',
      include: ['src/utils/**/*.js', 'src/components/**/*.{js,vue}'],
      exclude: [
        'src/utils/jspdf.js',
        'src/utils/pdf*.js',
        'src/utils/jsencrypt.js',
        'src/utils/htmlToPdf.js',
        'src/utils/dict/**',
        'src/utils/generator/**'
      ],
      thresholds: {
        lines: 30,
        functions: 30,
        statements: 30
      }
    }
  }
})

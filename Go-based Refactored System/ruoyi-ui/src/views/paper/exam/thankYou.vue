<template>
  <div class="thank-you">
    <div class="card">
      <i class="el-icon-circle-check icon"></i>
      <h2>测评完成</h2>
      <p>感谢您的参与，答题已完成！</p>
      <p class="sub">您可以关闭此页面。</p>
    </div>
  </div>
</template>

<script>
export default { name: 'ExamThankYou' }
</script>

<style scoped>
.thank-you {
  display: flex; justify-content: center; align-items: center;
  min-height: 100vh; background: #f5f7fa;
}
.card {
  text-align: center; background: #fff; border-radius: 12px;
  padding: 60px 48px; box-shadow: 0 2px 12px rgba(0,0,0,.08);
}
.icon { font-size: 64px; color: #67C23A; }
h2 { margin: 20px 0 12px; color: #303133; }
p { color: #606266; font-size: 16px; margin: 8px 0; }
.sub { color: #909399; font-size: 14px; }
</style>

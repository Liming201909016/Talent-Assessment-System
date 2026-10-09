<template>
  <span v-if="allowed" class="resume-control">
    <el-button size="small" icon="el-icon-link" @click="open">签发续答链接</el-button>
    <el-dialog title="签发考生续答链接" :visible="visible" width="600px" :fullscreen="isMobile" append-to-body :close-on-click-modal="false" @close="close">
      <p class="resume-hint">结果列表仅含已提交记录。请人工核对考生 ID 与原试卷 ID；仅恢复已有冻结试卷，不登记人员、不重开计时。链接最多有效 5 分钟且不晚于原截止时间。</p>
      <el-form ref="form" :model="form" :rules="rules" label-width="100px" size="small">
        <el-form-item label="测评 ID"><span>{{ examId }}</span></el-form-item>
        <el-form-item label="考生 ID" prop="participantId"><el-input v-model="form.participantId" :disabled="loading" placeholder="输入明确 candidate ID" /></el-form-item>
        <el-form-item label="原试卷 ID" prop="paperId"><el-input v-model="form.paperId" :disabled="loading" placeholder="输入明确 paper ID" /></el-form-item>
      </el-form>
      <el-alert v-if="error" :title="error" type="error" :closable="false" show-icon />
      <div v-if="link || linkExpired" class="issued-link" role="status" aria-live="polite">
        <p>本机显示有效期至：{{ expiryText }}。{{ linkExpired ? '已到期，请重新签发。' : '仅交给对应考生，服务器仍会核验签名及状态。' }}</p>
        <el-input :value="linkExpired ? '' : link" type="textarea" readonly aria-label="考生续答链接" :rows="3" />
        <el-button size="small" :disabled="copyDisabled" @click="copyLink">复制续答链接</el-button>
      </div>
      <span slot="footer"><el-button @click="close">关闭</el-button><el-button type="primary" :loading="loading" :disabled="loading || !allowed" @click="confirmIssue">确认签发</el-button></span>
    </el-dialog>
  </span>
</template>

<script>
import { canManageManagementTraits, issueManagementTraitsCandidateResume } from '@/api/managementTraits'
import { buildManagementTraitsResumeLink, validateManagementTraitsResumeToken } from '@/utils/managementTraitsResume'

export default {
  name: 'ManagementTraitsResumeDialog',
  props: { examId: { type: String, required: true } },
  data() {
    return { visible: false, loading: false, error: '', form: { participantId: '', paperId: '' }, link: '', expiresAt: 0, now: Date.now(), timer: null, sequence: 0, disposed: false, expiredNotice: false,
      rules: { participantId: [{ required: true, message: '请输入明确考生 ID', trigger: 'blur' }], paperId: [{ required: true, message: '请输入原试卷 ID', trigger: 'blur' }] } }
  },
  computed: {
    allowed() { return canManageManagementTraits(this.$store) },
    isMobile() { return this.$store.state.app && this.$store.state.app.device === 'mobile' },
    linkExpired() { return this.expiredNotice || (!!this.link && this.now >= this.expiresAt) },
    copyDisabled() { return !this.allowed || !this.link || this.linkExpired || this.loading },
    expiryText() { return this.expiresAt ? new Date(this.expiresAt).toLocaleString() : '—' }
  },
  watch: {
    examId() { this.close() },
    allowed(value) { if (!value) this.close() },
    'form.participantId'() { this.clearLink() },
    'form.paperId'() { this.clearLink() }
  },
  beforeDestroy() { this.disposed = true; this.close() },
  methods: {
    clearLink() { this.link = ''; this.expiresAt = 0; this.expiredNotice = false; clearInterval(this.timer); this.timer = null },
    expireLink() { this.link = ''; this.expiredNotice = true; clearInterval(this.timer); this.timer = null },
    open() { if (!this.allowed || this.disposed || this.visible) return; this.close(); this.visible = true },
    close() { this.sequence++; this.visible = false; this.loading = false; this.error = ''; this.clearLink(); this.form = { participantId: '', paperId: '' }; if (this.$refs.form) this.$refs.form.clearValidate() },
    async confirmIssue() {
      if (!this.allowed || !this.visible || this.loading || this.disposed) return
      this.clearLink(); this.error = ''
      const binding = { examId: this.examId, participantId: this.form.participantId.trim(), paperId: this.form.paperId.trim() }
      if (!binding.examId || !binding.participantId || !binding.paperId) { this.error = '请输入明确考生 ID 和原试卷 ID，再确认签发。'; return }
      this.loading = true
      const sequence = ++this.sequence
      try {
        const valid = await new Promise(resolve => this.$refs.form.validate(resolve))
        if (!valid || sequence !== this.sequence || !this.allowed || this.disposed) return
        const response = await issueManagementTraitsCandidateResume(binding)
        if (sequence !== this.sequence || !this.visible || !this.allowed || this.disposed) return
        const link = buildManagementTraitsResumeLink(this.$router, response, binding)
        const claims = validateManagementTraitsResumeToken(response.data.participantToken, binding)
        this.link = link; this.expiresAt = claims.exp * 1000; this.now = Date.now()
        this.timer = setInterval(() => { this.now = Date.now(); if (this.now >= this.expiresAt) this.expireLink() }, 1000)
      } catch (_) {
        if (sequence === this.sequence && !this.disposed) { this.clearLink(); this.error = '签发失败：试卷已到期、绑定不符或接口暂不可用。请核对原 ID 后重试或联系管理员。' }
      } finally { if (sequence === this.sequence) this.loading = false }
    },
    async copyLink() {
      this.now = Date.now()
      if (this.linkExpired) this.expireLink()
      if (this.copyDisabled) return
      try { await navigator.clipboard.writeText(this.link); if (!this.disposed && this.visible) this.$message.success('续答链接已复制，请仅交给对应考生。') }
      catch (_) { if (!this.disposed) this.$message.warning('浏览器不支持复制，请手动选择上方链接复制。') }
    }
  }
}
</script>

<style scoped>
.resume-control { display: inline-block; margin-right: 8px; }
.resume-hint { color: #606266; font-size: 14px; line-height: 1.5; }
.issued-link { margin-top: 12px; overflow-wrap: anywhere; }
.issued-link .el-button { margin-top: 8px; }
</style>
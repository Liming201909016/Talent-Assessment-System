<template>
  <div class="app-container" style="margin-left: auto; margin-right: auto;">

    <template v-if="!examBlocked">
    <el-card style="margin-top: 20px; ">
      <h3 align="center"  style="margin-bottom: 20px;">考生信息</h3>
      <el-form ref="candidateForm" :model="candidateForm" :rules="rules" label-position="left" label-width="120px"
       >

        <el-form-item label="姓名" prop="name" v-if="showField('name')">
          <el-input v-model="candidateForm.name" :max-width="100" />
        </el-form-item>

        <el-form-item label="手机号" prop="telephone" v-if="showField('telephone')">
          <el-input v-model="candidateForm.telephone" placeholder="请输入手机号" />
        </el-form-item>

        <el-form-item label="身份证号" prop="idNumber" v-if="showField('idNumber')">
          <el-input v-model="candidateForm.idNumber" placeholder="请输入身份证号" />
        </el-form-item>

        <el-form-item label="年龄" prop="age" v-if="showField('age')">
          <el-input v-model="candidateForm.age" />
        </el-form-item>

        <el-form-item label="性别" prop="gender" v-if="showField('gender')">
          <el-select v-model="candidateForm.gender" placeholder="请选择" >
            <el-option value="0" label="男"></el-option>
            <el-option value="1" label="女"></el-option>
          </el-select>
        </el-form-item>

        <el-form-item :label="isStu ? '学校' : '单位'" prop="affiliation" v-if="showField('affiliation')">
          <el-input v-model="candidateForm.affiliation" :placeholder="isStu ? '请输入学校' : '请输入单位'" />
        </el-form-item>
        <el-form-item label="部门" prop="depart" v-if="showField('depart')">
          <el-input v-model="candidateForm.depart" placeholder="请输入部门" />
        </el-form-item>
        <el-form-item :label="repoCode.startsWith('002') ? '职务' : '岗位'" prop="post" v-if="showField('post')">
          <el-input v-model="candidateForm.post" :placeholder="repoCode.startsWith('002') ? '请输入职务' : '请输入岗位'" />
        </el-form-item>
        <el-form-item label="学历" prop="degree" v-if="showField('degree')">
          <el-input v-model="candidateForm.degree" placeholder="请输入学历" />
        </el-form-item>
        <el-form-item label="专业" prop="major" v-if="showField('major')">
          <el-input v-model="candidateForm.major" placeholder="请输入专业" />
        </el-form-item>


      </el-form>
      <div style="margin-top: 20px; text-align: right;" >
        <el-button type="primary" @click="handleSave" style="margin-left: auto; margin-right: auto">保存</el-button>
      </div>
    </el-card>
    </template>


  </div>
</template>

<script>
import {fetchCandidate, saveData} from '@/api/candidate/candidate'
import {fetchDetail} from "@/api/exam/exam";

export default {
  name: 'CandidateInfo',

  data() {
    return {

      candidateForm: {
        id: '',
        // 考生类型
        stuFlag: 0,
        // 姓名
        name: null,
        // 身份证号
        idNumber: null,
        // 年龄
        age: null,
        // 性别
        gender: '',
        // 电话号码
        telephone: '',
        // 所属单位
        affiliation: null,
        // 部门
        depart: null,
        // 岗位
        post: null,
        // 专业
        major: null,
        // 学历
        degree: null,
        examId: '',
      },

      examId: '',
      testerId: '',
      stuFlag: 0,
      examBlocked: false,
      repoCode: '',
      requiredFields: null,
      rules: {
        name: [
          { required: true, message: '姓名不能为空！' },
          {
            validator: function (rule, value, callback) {
              if (value.length > 10) {
                callback(new Error("姓名长度小于10"))
              }

              callback()
            }
          }
        ],

        age: [
          { required: true, message: '年龄不能为空！' },
          {
            validator: function (rule, value, callback) {
              if (isNaN(value)) {
                callback(new Error("请输入数字"));
              }
              if (typeof value === "string" && value.indexOf(".") !== -1) {
                callback(new Error("请输入整数"));
              }

              let age = Number(value)
              if (age < 14 ) {
                callback(new Error("年龄不能小于14"));
              }
              callback();
            },
          },
        ],

        gender: [
          { required: true, message: '性别不能为空！' }
        ],

        telephone: [
          { required: true, message: '电话号码不能为空！' },
          {
            validator: function (rule, value, callback) {
              if (/^1\d{10}$/.test(value) == false) {
                callback(new Error("手机号格式错误"));
              } else {
                callback();
              }
            },
          }
        ],
      }
    }
  },
  computed: {
    // FB-040 修复：仅 001 心理特质 + 学生版（stuFlag==1）显示"学校"。
    // 002 管理特质里 stuFlag==1 含义是"基层员工"，仍属于职场单位，应显示"单位"。
    isStu() {
      return this.stuFlag == 1 && (this.repoCode || '').startsWith('001')
    }
  },
  watch: {
    // R06: 勾选的信息项均为必填
    requiredFields(fields) {
      if (!fields || fields.length === 0) return
      const fieldLabels = {
        name: '姓名', idNumber: '身份证号', age: '年龄', gender: '性别',
        telephone: '电话号码', affiliation: this.isStu ? '学校' : '单位',
        depart: '部门', post: '岗位', degree: '学历', major: '专业'
      }
      fields.forEach(f => {
        if (!this.rules[f]) {
          this.rules[f] = [{ required: true, message: (fieldLabels[f] || f) + '不能为空！' }]
        } else if (!this.rules[f].some(r => r.required)) {
          this.rules[f].unshift({ required: true, message: (fieldLabels[f] || f) + '不能为空！' })
        }
      })
      // 强制 form 重新校验规则
      this.$nextTick(() => {
        if (this.$refs.candidateForm) {
          this.$refs.candidateForm.clearValidate()
        }
      })
    }
  },
  created() {
    this.examId = this.$route.params.examId
    this.stuFlag = this.$route.params.stuFlag
    this.repoCode = this.$route.params.repoCode

    // 加载考试配置（勾选信息项 + 时间校验）
    if (this.examId) {
      fetchDetail(this.examId).then(res => {
        if (res.data) {
          // 考试状态检查 — 同时检查 state 和实际时间
          const state = res.data.state
          const now = new Date()
          const startTime = res.data.startTime ? new Date(res.data.startTime.replace(' ', 'T')) : null
          const endTime = res.data.endTime ? new Date(res.data.endTime.replace(' ', 'T')) : null
          const timeLimit = res.data.timeLimit

          // 时限模式：检查当前时间是否在开放区间内
          if (timeLimit && startTime && now < startTime) {
            this.examBlocked = true
            this.$alert(`考试尚未开始，开放时间：${res.data.startTime}，请在规定时间内参加测评。`, '提示', {
              confirmButtonText: '关闭', type: 'warning', showClose: false, closeOnClickModal: false
            })
            return
          }
          if (timeLimit && endTime && now > endTime) {
            this.examBlocked = true
            this.$alert('考试已结束，无法继续参加测评。', '提示', {
              confirmButtonText: '关闭', type: 'error', showClose: false, closeOnClickModal: false
            })
            return
          }
          if (state === 2) {
            this.examBlocked = true
            this.$alert('考试尚未开始，请在规定时间内参加测评。', '提示', {
              confirmButtonText: '关闭', type: 'warning', showClose: false, closeOnClickModal: false
            })
            return
          }
          if (state === 3) {
            this.examBlocked = true
            this.$alert('考试已结束，无法继续参加测评。', '提示', {
              confirmButtonText: '关闭', type: 'error', showClose: false, closeOnClickModal: false
            })
            return
          }
          // 加载勾选信息项
          if (res.data.requiredFields) {
            this.requiredFields = res.data.requiredFields.split(',')
          } else {
            this.requiredFields = [] // 未配置则显示全部
          }
        }
      }).catch((err) => {
        console.error('fetchDetail failed:', err)
        this.requiredFields = [] // 加载失败则显示全部
      })
    }

    if (this.$route.params.testerId !== undefined && this.$route.params.testerId !== '') {
      this.candidateForm.id = this.$route.params.testerId
      console.log(this.candidateForm.id)
      this.fetchData()
    }
  },

  methods: {

    showField(fieldName) {
      // requiredFields 未加载完成前不显示任何字段
      if (this.requiredFields === null) return false
      return this.requiredFields.length === 0 || this.requiredFields.includes(fieldName)
    },

    fetchData() {
      fetchCandidate(this.examId, this.candidateForm.id).then(response => {
        this.candidateForm = response.data
        console.log(this.candidateForm)
        // this.candidateForm.name = response.data.name
        // this.candidateForm.age = response.data.age
        // this.candidateForm.gender = response.data.gender
        // this.candidateForm.affiliation = response.data.affiliation
        // this.candidateForm.post = response.data.post
      })
    },

    handleSave() {
      this.$refs.candidateForm.validate((valid) => {
        if (!valid) {
          return
        }

        this.candidateForm.examId = this.examId

        this.$confirm('确实要提交保存吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          this.submitForm()
        })
      })
    },

    submitForm() {
      this.candidateForm.stuFlag = this.stuFlag
      saveData(this.candidateForm).then(response => {
        this.candidateForm.id = response.data.id
        if (response.data.participantToken) {
          sessionStorage.setItem('competencyParticipantToken', response.data.participantToken)
          sessionStorage.setItem('competencyParticipantType', 'candidate')
        }
        console.log(this.candidateForm.id)
        this.$notify({
          title: '成功',
          message: '信息保存成功！',
          type: 'success',
          duration: 2000
        })
        // console.log("====================================")
        this.$router.replace({ name: 'PreExam', params:
            { examId: this.examId , id: this.candidateForm.id,stuFlag: this.stuFlag,repoCode: this.repoCode}})
      })

      // this.$router.push({ name: 'PreExam', params: { id: this.examId }})
    },

  }
}
</script>

<style scoped>

</style>



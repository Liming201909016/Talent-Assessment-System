<template>
  <div class="app-container">

    <template v-if="!isCompetency">
    <h3>组卷信息</h3>
    <el-card style="margin-top: 20px">

      <div style="float: right; font-weight: bold; color: #ff0000" v-if="flag">试卷总分：{{ postForm.totalScore }}分</div>

      <div>

<!--        <el-button class="filter-item" size="small" type="primary" icon="el-icon-plus" @click="handleAdd">
          添加题库
        </el-button>-->

        <el-table
          :data="repoList"
          :border="false"
          empty-text="请点击上面的`添加题库`进行设置"
          style="width: 100%; margin-top: 15px"
        >
          <el-table-column
            label="题库"
            width="300"
          >
            <template slot-scope="scope">
              <repo-select v-model="scope.row.repoId" :multi="false" @change="repoChange($event, scope.row,scope.$index)" />
            </template>

          </el-table-column>
          <el-table-column
            label="题目数量"
            align="center"
          >

            <template slot-scope="scope">
              <el-input-number v-model="scope.row.radioCount" :disabled="true" :controls="false" style="width: 100px" />
            </template>

          </el-table-column>

          <el-table-column
            label="题目分数"
            align="center"
            v-if="flag"
          >
            <template slot-scope="scope">
              <el-input-number v-model="scope.row.radioScore" :min="0" :controls="false" style="width: 100%" />
            </template>
          </el-table-column>

          <el-table-column
            label="多选数量"
            align="center"
            v-if="flag"
          >

            <template slot-scope="scope">
              <el-input-number v-model="scope.row.multiCount" :min="0" :max="scope.row.totalMulti" :controls="false" style="width: 100px" /> / {{ scope.row.totalMulti }}
            </template>

          </el-table-column>

          <el-table-column
            label="多选分数"
            align="center"
            v-if="flag"
          >
            <template slot-scope="scope">
              <el-input-number v-model="scope.row.multiScore" :min="0" :controls="false" style="width: 100%" />
            </template>
          </el-table-column>

          <el-table-column
            label="判断题数量"
            align="center"
            v-if="flag"
          >

            <template slot-scope="scope">
              <el-input-number v-model="scope.row.judgeCount" :min="0" :max="scope.row.totalJudge" :controls="false" style="width: 100px" />  / {{ scope.row.totalJudge }}
            </template>

          </el-table-column>

          <el-table-column
            label="判断题分数"
            align="center"
            v-if="flag"
          >
            <template slot-scope="scope">
              <el-input-number v-model="scope.row.judgeScore" :min="0" :controls="false" style="width: 100%" />
            </template>
          </el-table-column>

<!--          <el-table-column
            label="删除"
            align="center"
            width="80px"
          >
            <template slot-scope="scope">
              <el-button type="danger" icon="el-icon-delete" circle @click="removeItem(scope.$index)" />
            </template>
          </el-table-column>-->

        </el-table>

      </div>

    </el-card>
    </template>

    <h3>测评配置</h3>
    <el-card style="margin-top: 20px">

      <el-form ref="postForm" :model="postForm" :rules="rules" label-position="left" label-width="120px">

        <el-form-item label="测评类别" prop="assessmentType">
          <el-radio-group v-model="postForm.assessmentType" :disabled="isPublishedCompetency" @change="handleAssessmentTypeChange">
            <el-radio size="large" border label="legacy">传统测评</el-radio>
            <el-radio size="large" border label="competency">胜任力测评</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="测评名称" prop="title">
          <el-input v-model="postForm.title" />
        </el-form-item>
        <el-form-item v-if="!isCompetency" label="适用版本">
          <el-radio-group v-model="postForm.stuFlag">
            <el-radio size="large" border :label="1">{{repoCode.startsWith('002')?'基层员工':'学生版'}}</el-radio>
            <el-radio size="large" border :label="0">{{repoCode.startsWith('002')?'管理干部':'职场版'}}</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-else>
          <el-form-item label="报告版本" prop="competencyReportAudience">
            <el-radio-group v-model="postForm.competencyReportAudience" :disabled="isPublishedCompetency">
              <el-radio size="large" border label="frontline_employee">基层员工版</el-radio>
              <el-radio size="large" border label="leader">领导人员版</el-radio>
            </el-radio-group>
            <div class="field-hint">两个版本的模板、模块和得分一致，仅总体评价与发展建议文案不同。</div>
          </el-form-item>

          <el-form-item label="测评维度" prop="dimensionIds">
            <competency-dimension-selector
              v-model="postForm.dimensionIds"
              :dimensions="competencyDimensions"
              :loading="dimensionLoading"
              :disabled="isPublishedCompetency"
            />
          </el-form-item>
        </template>

        <el-form-item label="考试描述" prop="content" v-if="flag">
          <el-input v-model="postForm.content" type="textarea" />
        </el-form-item>

        <el-form-item label="总分数" prop="totalScore" v-if="flag">
          <el-input-number :value="postForm.totalScore" disabled />
        </el-form-item>

        <el-form-item label="及格分" prop="qualifyScore" v-if="flag">
          <el-input-number v-model="postForm.qualifyScore" :max="postForm.totalScore" />
        </el-form-item>

        <el-form-item label="测评时长(分钟)" prop="totalTime">
          <el-input-number v-model="postForm.totalTime" :min="isCompetency ? 1 : 0" />
          <div v-if="isCompetency" class="field-hint">胜任力测评必须配置答题时长，到时由系统自动提交。</div>
        </el-form-item>

        <el-form-item label="测评开放类型">
          <el-radio-group v-model="postForm.isOpen">
            <el-radio size="large" border :label="1">开放</el-radio>
            <el-radio size="large" border :label="2">封闭</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="postForm.isOpen === 1" label="考生填写信息">
          <div style="display: grid; grid-template-columns: repeat(5, auto); gap: 8px 16px;">
            <el-checkbox-group v-model="requiredFieldsList" style="display:contents">
              <el-checkbox label="name">姓名</el-checkbox>
              <el-checkbox label="gender">性别</el-checkbox>
              <el-checkbox label="age">年龄</el-checkbox>
              <el-checkbox label="telephone">手机号</el-checkbox>
              <el-checkbox label="idNumber">身份证号</el-checkbox>
              <el-checkbox label="affiliation">单位/学校</el-checkbox>
              <el-checkbox label="post">岗位</el-checkbox>
              <el-checkbox label="depart">部门</el-checkbox>
              <el-checkbox label="degree">学历</el-checkbox>
              <el-checkbox label="major">专业</el-checkbox>
            </el-checkbox-group>
          </div>
          <div style="color: #999; font-size: 12px; margin-top: 4px">勾选后，开放测评时考生需填写对应信息，报告中也只体现已勾选项</div>
        </el-form-item>

        <el-form-item label="测评答题类型">
          <el-radio-group v-model="postForm.answerType">
            <el-radio size="large" border :label="1">滚动</el-radio>
            <el-radio size="large" border :label="2">点击</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="是否查看报告">
          <el-switch v-model="postForm.showPdf" active-text="是" inactive-text="否" />
        </el-form-item>

        <el-form-item label="是否限时">
          <el-switch v-model="postForm.timeLimit" active-text="是" inactive-text="否" />
        </el-form-item>

        <el-form-item v-if="postForm.timeLimit" label="测评时间" prop="totalTime">

          <el-date-picker
            v-model="dateValues"
            format="yyyy-MM-dd HH:mm"
            value-format="yyyy-MM-dd HH:mm"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            :default-time="['08:00:00', '18:00:00']"
          />

        </el-form-item>

      </el-form>

    </el-card>

    <div style="margin-top: 20px">
      <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      <el-button v-if="isCompetency && postForm.id && postForm.publishStatus === 0" type="success" :loading="publishing" @click="handlePublish">
        发布并冻结题目
      </el-button>
    </div>

  </div>
</template>

<script>
import { fetchDetail, saveData } from '@/api/exam/exam'
import { fetchCompetencyDimensions, publishCompetencyExam } from '@/api/competency'
import { fetchTree } from '@/api/sys/depart/depart'
import RepoSelect from '@/components/RepoSelect'
import CompetencyDimensionSelector from './components/CompetencyDimensionSelector'

export default {
  name: 'ExamDetail',
  components: { CompetencyDimensionSelector, RepoSelect },
  data() {
    return {
      repoCode:'',
      flag:false,
      saving: false,
      publishing: false,
      dimensionLoading: false,
      competencyDimensions: [],

      step: 1,
      treeData: [],
      defaultProps: {
        label: 'deptName'
      },
      levels: [
        { value: 0, label: '不限' },
        { value: 1, label: '普通' },
        { value: 2, label: '较难' }
      ],
      filterText: '',
      treeLoading: false,
      dateValues: [],
      requiredFieldsList: ['name', 'gender', 'age', 'telephone'],
      quDialogShow: false,
      quDialogType: 1,
      excludes: [],

      scoreDialog: false,
      scoreBatch: 0,

      // 题库
      repoList: [
        {radioScore: 1}
      ],

      // 题目列表
      quList: [[], [], [], []],
      quEnable: [false, false, false, false],

      postForm: {
        assessmentType: 'legacy',
        scoringMode: 'legacy',
        competencyReportAudience: '',
        dimensionIds: [],
        publishStatus: 1,
        // 总分数
        totalScore: 1,
        // 题库列表
        repoList: [],
        // 题目列表
        quList: [],
        // 组题方式
        joinType: 1,
        // 开放类型
        openType: 1,
        // 考试班级列表
        departIds: [],
        // 测评答题类型
        answerType: 1,
        // 测评开放类型
        isOpen: 1,
        // 是否查看报告
        showPdf: false,
        // 是否限时
        timeLimit: false,
        // 是否学生测评1是0否
        stuFlag: 0,
      },
      rules: {
        assessmentType: [
          { required: true, message: '请选择测评类别', trigger: 'change' }
        ],
        competencyReportAudience: [
          { required: true, message: '请选择基层员工版或领导人员版', trigger: 'change' }
        ],
        dimensionIds: [
          { type: 'array', required: true, min: 1, message: '请至少选择一个测评维度', trigger: 'change' }
        ],
        title: [
          { required: true, message: '测评名称不能为空！' }
        ],

        // content: [
        //   { required: true, message: '考试名称不能为空！' }
        // ],

        open: [
          { required: true, message: '测评权限不能为空！' }
        ],

        // totalScore: [
        //   { required: true, message: '考试分数不能为空！' }
        // ],

        // qualifyScore: [
        //   { required: true, message: '及格分不能为空！' }
        // ],

        totalTime: [
          { required: true, message: '测评时间不能为空！' }
        ],

        // ruleId: [
        //   { required: true, message: '考试规则不能为空' }
        // ],
        password: [
          { required: true, message: '测评口令不能为空！' }
        ]
      }
    }
  },

  computed: {
    isCompetency() {
      return this.postForm.assessmentType === 'competency'
    },
    isPublishedCompetency() {
      return this.isCompetency && Number(this.postForm.publishStatus) === 1
    }
  },

  watch: {

    filterText(val) {
      this.$refs.tree.filter(val)
    },

    dateValues: {
      handler() {
        this.postForm.startTime = this.dateValues[0]
        this.postForm.endTime = this.dateValues[1]
      }
    },

    requiredFieldsList: {
      immediate: true,
      handler(val) {
        this.postForm.requiredFields = val.join(',')
      }
    },

    'postForm.dimensionIds': {
      handler(value) {
        if (this.isCompetency) {
          this.postForm.totalScore = (value || []).length * 5
        }
      },
      deep: true
    },

    // 题库变换
    repoList: {
      handler() {
        const that = this

        if (that.isCompetency) {
          that.postForm.repoList = []
          return
        }

        that.postForm.totalScore = 0
        this.repoList.forEach(function(item) {
          if(that.repoCode.startsWith('002')){
            item.radioScore = 5
          }else{
            item.radioScore = 1
          }
          that.postForm.totalScore += item.radioCount * item.radioScore
          that.postForm.totalScore += item.multiCount * item.multiScore
          that.postForm.totalScore += item.judgeCount * item.judgeScore
        })

        // 赋值
        this.postForm.repoList = this.repoList
      },
      deep: true
    }

  },
  created() {
    const id = this.$route.params.id
    if (typeof id !== 'undefined') {
      this.fetchData(id)
    }

    /*fetchTree({}).then(response => {
      this.treeData = response.data
    })*/
  },
  methods: {

    handlePublish() {
      this.$confirm('发布后测评维度、题目范围和报告版本不可修改，确认发布吗？', '发布胜任力测评', { type: 'warning' }).then(async () => {
        this.publishing = true
        try {
          const response = await publishCompetencyExam(this.postForm.id)
          this.postForm.publishStatus = 1
          this.postForm.publishedAt = response.data.publishedAt
          this.$message.success(`发布成功，共冻结 ${response.data.questionCount} 道题目`)
        } finally {
          this.publishing = false
        }
      }).catch(() => {})
    },

    handleAssessmentTypeChange(value) {
      if (value === 'competency') {
        this.postForm.scoringMode = 'competency_average'
        this.postForm.publishStatus = 0
        this.postForm.repoList = []
        this.postForm.showPdf = false
        this.loadCompetencyDimensions()
      } else {
        this.postForm.scoringMode = 'legacy'
        this.postForm.publishStatus = 1
        this.postForm.competencyReportAudience = ''
        this.postForm.dimensionIds = []
        this.postForm.repoList = this.repoList
      }
      this.$nextTick(() => this.$refs.postForm && this.$refs.postForm.clearValidate())
    },

    loadCompetencyDimensions() {
      if (this.dimensionLoading || this.competencyDimensions.length > 0) return
      this.dimensionLoading = true
      fetchCompetencyDimensions().then(response => {
        this.competencyDimensions = response.data || []
      }).finally(() => {
        this.dimensionLoading = false
      })
    },

    handleSave() {

      this.$refs.postForm.validate((valid) => {
        if (!valid) {
          return
        }

        if (this.isCompetency && !['frontline_employee', 'leader'].includes(this.postForm.competencyReportAudience)) {
          this.$message.warning('请选择基层员工版或领导人员版')
          return
        }

        if (this.isCompetency && (!this.postForm.dimensionIds || this.postForm.dimensionIds.length === 0)) {
          this.$message.warning('请至少选择一个测评维度')
          return
        }

        if (!this.isCompetency && this.postForm.totalScore === 0) {
          this.$notify({
            title: '提示信息',
            message: '测评规则设置不正确，请确认！',
            type: 'warning',
            duration: 2000
          })

          return
        }

        if (!this.isCompetency && this.postForm.joinType === 1) {
          for (let i = 0; i < this.postForm.repoList.length; i++) {
            const repo = this.postForm.repoList[i]

            if (!repo.repoId) {
              this.$notify({
                title: '提示信息',
                message: '测评题库选择不正确！',
                type: 'warning',
                duration: 2000
              })

              return
            }

            if ((repo.radioCount > 0 && repo.radioScore === 0) || (repo.radioCount === 0 && repo.radioScore > 0)) {
              this.$notify({
                title: '提示信息',
                message: '题库第：[' + (i + 1) + ']项存在无效的单选题配置！',
                type: 'warning',
                duration: 2000
              })

              return
            }

            if ((repo.multiCount > 0 && repo.multiScore === 0) || (repo.multiCount === 0 && repo.multiScore > 0)) {
              this.$notify({
                title: '提示信息',
                message: '题库第：[' + (i + 1) + ']项存在无效的多选题配置！',
                type: 'warning',
                duration: 2000
              })

              return
            }

            if ((repo.judgeCount > 0 && repo.judgeScore === 0) || (repo.judgeCount === 0 && repo.judgeScore > 0)) {
              this.$notify({
                title: '提示信息',
                message: '题库第：[' + (i + 1) + ']项存在无效的判断题配置！',
                type: 'warning',
                duration: 2000
              })
              return
            }
          }
        }

        this.$confirm('确实要提交保存吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          this.submitForm()
        })
      })
    },

    handleCheckChange() {
      const that = this
      // 置空
      this.postForm.departIds = []

      const nodes = this.$refs.tree.getCheckedNodes()
      nodes.forEach(function(item) {
        that.postForm.departIds.push(item.id)
      })
    },

    // 添加子项
    handleAdd() {
      this.repoList.push({ rowId: new Date().getTime(), radioCount: 0, radioScore: 0, multiCount: 0, multiScore: 0, judgeCount: 0, judgeScore: 0, saqCount: 0, saqScore: 0 })
    },

    removeItem(index) {
      this.repoList.splice(index, 1)
    },

    fetchData(id) {
      const that = this

      fetchDetail(id).then(response => {
        const data = response.data || {}
        data.assessmentType = data.assessmentType || 'legacy'
        data.scoringMode = data.scoringMode || (data.assessmentType === 'competency' ? 'competency_average' : 'legacy')
        data.competencyReportAudience = data.competencyReportAudience || ''
        data.dimensionIds = data.dimensionIds || []
        this.postForm = data

        if (this.isCompetency) {
          this.loadCompetencyDimensions()
        }

        console.log(this.postForm)
        if (this.postForm.startTime && this.postForm.endTime) {
          this.dateValues = [this.postForm.startTime, this.postForm.endTime]
        }

        // 恢复信息项勾选
        if (this.postForm.requiredFields) {
          this.requiredFieldsList = this.postForm.requiredFields.split(',').filter(f => f)
        } else {
          // DB 为空时用默认值回填，保证保存时会写入
          this.postForm.requiredFields = this.requiredFieldsList.join(',')
        }

        // 按分组填充题目
        if (this.postForm.joinType === 2) {
          this.postForm.quList.forEach(function(item) {
            const index = item.quType - 1
            that.quList[index].push(item)
            that.quEnable[index] = true
          })
        }

        if (this.postForm.joinType === 1) {
          that.repoList = that.postForm.repoList
        }
        const firstRepo = response.data.repoList && response.data.repoList[0]
        this.repoCode = firstRepo ? firstRepo.repoCode : ''
      })
    },

    submitForm() {
      // 校验和处理数据
      this.postForm.repoList = this.isCompetency ? [] : this.repoList
      this.saving = true

      saveData(this.postForm).then(() => {
        this.$notify({
          title: '成功',
          message: '测评保存成功！',
          type: 'success',
          duration: 2000
        })

        this.$router.push({ name: 'ListExam' })
      }).finally(() => {
        this.saving = false
      })
    },

    filterNode(value, data) {
      if (!value) return true
      return data.deptName.indexOf(value) !== -1
    },

     repoChange(e, row,rowIndex) {
      if (e != null) {
        // console.log(e)
        // console.log(row)
        row.radioCount = e.radioCount
        row.totalRadio = e.radioCount
        row.totalMulti = e.multiCount
        row.totalJudge = e.judgeCount
        console.log(e,row,rowIndex)
      } else {
        row.totalRadio = 0
        row.totalMulti = 0
        row.totalJudge = 0
      }
      if(rowIndex === 0){
        this.repoCode = e.code
      }
    }

  }
}
</script>

<style scoped>
.field-hint {
  margin-top: 6px;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}
</style>


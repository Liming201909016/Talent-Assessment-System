<template>
  <div class="app-container">

    <el-form ref="postForm" :model="postForm" :rules="rules" label-position="left" label-width="150px">

      <el-card>

        <el-form-item label="题目类型 " prop="quType" v-if="flag">

          <el-select v-model="postForm.quType" :disabled="quTypeDisabled" class="filter-item" @change="handleTypeChange">
            <el-option
              v-for="item in quTypes"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>

        </el-form-item>

        <el-form-item label="难度等级 " prop="level" v-if="flag">

          <el-select v-model="postForm.level" class="filter-item">
            <el-option
              v-for="item in levels"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>

        </el-form-item>

        <el-form-item label="归属题库" prop="repoId">

          <repo-select v-model="postForm.repoId" :multi="false" />

        </el-form-item>

        <el-form-item label="题目编号" prop="content">
          <el-input v-model="postForm.content" type="textarea" />
        </el-form-item>

        <el-form-item label="题干" prop="title">
          <el-input v-model="postForm.title" type="textarea" placeholder="题目的题干描述" />
        </el-form-item>

        <el-form-item label="整题解析" prop="oriPrice" v-if="flag">
          <el-input v-model="postForm.analysis" type="textarea" :precision="1" :max="999999" />
        </el-form-item>

      </el-card>

      <div v-if="postForm.quType!==4" class="filter-container" style="margin-top: 25px">

        <el-button class="filter-item" type="primary" icon="el-icon-plus" size="small" plain @click="handleAdd" v-if="flag">
          添加
        </el-button>

        <el-table
          :data="postForm.answerList"
          :border="true"
          style="width: 100%;"
        >
          <el-table-column
            label="选项"
            width="60"
            align="center"
          >
            <template slot-scope="scope">
              <span style="color:#909399; font-weight:600;">{{ String.fromCharCode(65 + scope.$index) }}</span>
            </template>
          </el-table-column>

          <el-table-column
            label="题库锚点"
            width="100"
            align="center"
            v-if="showAnchor"
          >
            <template slot-scope="scope">
              <el-radio v-model="anchorIdx" :label="scope.$index" @change="handleAnchorChange">设为锚点</el-radio>
            </template>
          </el-table-column>

          <el-table-column
            label="选项分数"
            width="100"
            align="center"
            v-if="showOptionScore"
          >
            <template slot-scope="scope">
              <el-input-number v-model="scope.row.score" :min="0" :max="10" :precision="0" size="mini" controls-position="right" style="width: 90px;" />
            </template>
          </el-table-column>

          <el-table-column
            label="选项内容"
          >
            <template slot-scope="scope">
              <el-input v-model="scope.row.content" type="textarea" :placeholder="'选项' + String.fromCharCode(65 + scope.$index) + '的内容'"/>
            </template>
          </el-table-column>

          <el-table-column
            label="答案解析"
            v-if="flag"
          >
            <template slot-scope="scope">
              <el-input v-model="scope.row.analysis" type="textarea" />
            </template>
          </el-table-column>

          <el-table-column
            label="操作"
            align="center"
            width="100px"
            v-if="flag"
          >
            <template slot-scope="scope">
              <el-button type="danger" icon="el-icon-delete" circle @click="removeItem(scope.$index)" />
            </template>
          </el-table-column>

        </el-table>

      </div>

      <div style="margin-top: 20px">
        <el-button type="primary" @click="submitForm">保存</el-button>
        <el-button type="info" @click="onCancel">返回</el-button>
      </div>

    </el-form>

  </div>
</template>

<script>
import { fetchDetail, saveData } from '@/api/qu/qu'
import RepoSelect from '@/components/RepoSelect'
import FileUpload from '@/components/FileUpload'

export default {
  name: 'QuDetail',
  components: { FileUpload, RepoSelect },
  data() {
    return {

      flag: false,
      choices: '',
      anchorIdx: 0,

      quTypeDisabled: false,
      itemImage: true,

      levels: [
        { value: 1, label: '普通' },
        { value: 2, label: '较难' }
      ],

      quTypes: [{
        value: 1,
        label: '单选题'
      }, {
        value: 2,
        label: '多选题'
      },
      {
        value: 3,
        label: '判断题'
      }
      ],

      postForm: {
        repoId: '',
        repoIds: [],
        repoCode: '',
        tagList: [],
        answerList: [],
        quType: 1,
        level: 1
      },
      rules: {
        content: [
          { required: true, message: '题目编号不能为空！' },
          {
            validator: function (rule, value, callback) {
              if (isNaN(value)) {
                callback(new Error("请输入数字"));
              }
              if (typeof value === "string" && value.indexOf(".") !== -1) {
                callback(new Error("请输入整数"));
              }

              callback()
            },
          }
        ],

        // quType: [
        //   { required: true, message: '题目类型不能为空！' }
        // ],
        //
        // level: [
        //   { required: true, message: '必须选择难度等级！' }
        // ],

        repoId: [
          { required: true, message: '必须选择一个题库！' }
        ]
      }
    }
  },
  created() {
    const id = this.$route.params.id
    if (typeof id !== 'undefined') {
      this.quTypeDisabled = true
      this.fetchData(id)
    }
  },

  computed: {
    /** 是否显示"题库锚点"列：仅 001 心理特质（A/B 二选一，公式按锚点正/反向计分） */
    showAnchor() {
      const code = this.postForm.repoCode || ''
      return code.startsWith('001')
    },
    /** 是否显示"选项分数"列：仅 002 管理特质（5 选项 likert，每个选项有不同得分 1~5） */
    showOptionScore() {
      const code = this.postForm.repoCode || ''
      return code.startsWith('002')
    },
  },

  mounted() {
    this.postForm.answerList.push({ isRight: true, content: '', analysis: '' })
    this.postForm.answerList.push({ isRight: false, content: '', analysis: '' })
    this.anchorIdx = 0
  },

  methods: {

    /** 锚点选择变化：把选中的那个 isRight=1，其它=0 */
    handleAnchorChange(idx) {
      this.postForm.answerList.forEach((it, i) => {
        it.isRight = (i === idx)
      })
    },

    handleTypeChange(v) {
      this.postForm.answerList = []
      if (v === 3) {
        this.postForm.answerList.push({ isRight: true, content: '正确', analysis: '' })
        this.postForm.answerList.push({ isRight: false, content: '错误', analysis: '' })
      }

      if (v === 1 || v === 2) {
        this.postForm.answerList.push({ isRight: false, content: '', analysis: '' })
        this.postForm.answerList.push({ isRight: false, content: '', analysis: '' })
        this.postForm.answerList.push({ isRight: false, content: '', analysis: '' })
        this.postForm.answerList.push({ isRight: false, content: '', analysis: '' })
      }
    },

    // 添加子项
    handleAdd() {
      this.postForm.answerList.push({ isRight: false, content: '', analysis: '' })
    },

    removeItem(index) {
      this.postForm.answerList.splice(index, 1)
    },

    syncRepoSelection() {
      this.postForm.repoIds = this.postForm.repoId ? [this.postForm.repoId] : []
    },

    fetchData(id) {
      fetchDetail(id).then(response => {
        this.postForm = response.data
        this.postForm.content = response.data.content.substring(1)
        // 回填单选题库
        if (this.postForm.repoIds && this.postForm.repoIds.length > 0) {
          this.postForm.repoId = this.postForm.repoIds[0]
        }
        // 同步锚点 idx
        const idx = (this.postForm.answerList || []).findIndex(a => a.isRight === true || a.isRight === 1)
        this.anchorIdx = idx >= 0 ? idx : 0
        console.log(this.postForm)
      })
    },
    submitForm() {
      console.log(JSON.stringify(this.postForm))

      this.syncRepoSelection()

      let rightCount = 0

      this.postForm.answerList.forEach(function(item) {
        if (item.isRight) {
          rightCount += 1
        }
      })

      if (this.postForm.quType === 1) {
        if (rightCount !== 1) {
          this.$message({
            message: '单选题答案只能有一个',
            type: 'warning'
          })

          return
        }
      }

      if (this.postForm.quType === 2) {
        if (rightCount < 2) {
          this.$message({
            message: '多选题至少要有两个正确答案！',
            type: 'warning'
          })

          return
        }
      }

      if (this.postForm.quType === 3) {
        if (rightCount !== 1) {
          this.$message({
            message: '判断题只能有一个正确项！',
            type: 'warning'
          })

          return
        }
      }

      this.$refs.postForm.validate((valid) => {
        if (!valid) {
          return
        }

        this.postForm.content = 'V' + this.postForm.content

        saveData(this.postForm).then(response => {
          this.postForm = response.data
          this.$notify({
            title: '成功',
            message: '试题保存成功！',
            type: 'success',
            duration: 2000
          })
          // 返回上一页（通常是某题库的题目列表，带 repoId）
          this.$router.back()
        })
      })
    },
    onCancel() {
      this.$router.back()
    }

  }
}
</script>

<style scoped>

</style>


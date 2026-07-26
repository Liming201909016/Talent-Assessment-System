<template>

  <div>

    <div v-if="repoId" style="margin-bottom: 12px; display: flex; align-items: center;">
      <el-button icon="el-icon-back" size="small" @click="$router.back()">返回题库</el-button>
      <h3 style="margin: 0 0 0 16px; font-weight: 600;">{{ repoTitle || '题目管理' }}</h3>
    </div>

    <data-table
      ref="pagingTable"
      :options="options"
      :list-query="listQuery"
      @multi-actions="handleMultiAction"
    >
      <template slot="filter-content">

        <el-row>
          <el-col :span="24">

            <el-select v-model="listQuery.params.quType" class="filter-item" clearable v-if="flag">
              <el-option
                v-for="item in quTypes"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>

            <repo-select v-if="!repoId" v-model="listQuery.params.repoIds" :multi="true" />

            <el-input v-model="listQuery.params.content" placeholder="题目内容" style="width: 200px;" class="filter-item" />


          </el-col>
        </el-row>

      </template>

      <template slot="data-columns">

        <el-table-column
          label="题目类型"
          align="center"
          width="100px"
          v-if="flag"
        >
          <template slot-scope="scope">
            {{ scope.row.quType | quTypeFilter(scope.row.quType) }}
          </template>
        </el-table-column>

        <el-table-column
          label="题目编号"
          show-overflow-tooltip
          width="150px"
          align="center"
        >
          <template slot-scope="scope">
            <router-link :to="{ name: 'UpdateQu', params:{ id: scope.row.id}}">
              {{ formatQuestionNumber(scope.row.content) }}
            </router-link>
          </template>
        </el-table-column>

        <el-table-column
          label="所属题库"
          prop="repoName"
          width="180px"
          show-overflow-tooltip
        />

        <el-table-column
          label="选项内容"
          show-overflow-tooltip
        >
          <template slot-scope="scope">
            <router-link :to="{ name: 'UpdateQu', params:{ id: scope.row.id}}">
              {{ formatAnswerList(scope.row.answerList) }}
            </router-link>
          </template>
        </el-table-column>

        <el-table-column
          label="创建时间"
          align="center"
          width="180px"
        >
          <template slot-scope="scope">{{ parseTime(scope.row.createTime) }}</template>
        </el-table-column>

        <el-table-column
          label="更新时间"
          align="center"
          width="180px"
        >
          <template slot-scope="scope">{{ parseTime(scope.row.updateTime) }}</template>
        </el-table-column>

        <el-table-column
          label="操作"
          align="center"
          width="100px"
          fixed="right"
        >
          <template slot-scope="scope">
            <router-link :to="{ name: 'UpdateQu', params:{ id: scope.row.id}}">
              <el-button type="text" size="mini" icon="el-icon-edit">修改</el-button>
            </router-link>
          </template>
        </el-table-column>

      </template>

    </data-table>

    <el-dialog
      :title="dialogTitle"
      :visible.sync="dialogVisible"
      width="30%"
    >

      <el-form label-position="left" label-width="100px">

        <el-form-item label="操作题库" prop="repoIds">
          <repo-select v-model="dialogRepos" :multi="true" />
        </el-form-item>

        <el-row>
          <el-button type="primary" @click="handlerRepoAction">保存</el-button>
        </el-row>

      </el-form>

    </el-dialog>


  </div>

</template>

<script>
import DataTable from '@/components/DataTable'
import RepoSelect from '@/components/RepoSelect'
import { batchAction } from '@/api/qu/repo'

export default {
  name: 'QuList',
  components: { RepoSelect, DataTable },
  data() {
    return {

      flag: false,
      repoId: this.$route.params.repoId || '',
      repoTitle: this.$route.params.repoTitle || '',

      dialogTitle: '加入题库',
      dialogVisible: false,
      importVisible: false,
      dialogRepos: [],
      dialogQuIds: [],
      dialogFlag: false,

      listQuery: {
        current: 1,
        size: 20,
        params: {
          content: '',
          quType: '',
          repoIds: this.$route.params.repoId ? [this.$route.params.repoId] : []
        }
      },

      quTypes: [
        {
          value: 1,
          label: '单选题'
        },
        {
          value: 2,
          label: '多选题'
        },
        {
          value: 3,
          label: '判断题'
        }
      ],

      options: {

        // 可批量操作
        multi: true,

        // 批量操作列表
        multiActions: [
          {
            value: 'delete',
            label: '删除'
          },
          {
            value: 'add-repo',
            label: '加入题库..'
          },
          {
            value: 'remove-repo',
            label: '从..题库移除'
          }
        ],
        // 列表请求URL
        listUrl: '/exam/api/qu/qu/paging',
        // 删除请求URL
        deleteUrl: '/exam/api/qu/qu/delete',
        // 启用禁用
        stateUrl: '/exam/api/qu/qu//state',
        // 添加数据路由
        addRoute: 'AddQu',
        importFile: true,
      }
    }
  },
  filters: {
    quTypeFilter(value){
      if(value === 1) return '单选题'
      else if(value === 2) return '多选题'
      else if(value === 3) return '判断题'
    }
  },
  methods: {
    formatQuestionNumber(content) {
      const value = String(content || '')
      return value.startsWith('V') ? value.substring(1) : value
    },
    formatAnswerList(answerList) {
      if (!Array.isArray(answerList) || answerList.length === 0) return '—'
      return answerList.map((answer, index) => `${String.fromCharCode(65 + index)}: ${answer.content || ''}`).join('; ')
    },
    handleMultiAction(obj) {
      if (obj.opt === 'add-repo') {
        this.dialogTitle = '加入题库'
        this.dialogFlag = false
      }

      if (obj.opt === 'remove-repo') {
        this.dialogTitle = '从题库移除'
        this.dialogFlag = true
      }

      this.dialogVisible = true
      this.dialogQuIds = obj.ids
    },

    handlerRepoAction() {
      const postForm = { repoIds: this.dialogRepos, quIds: this.dialogQuIds, remove: this.dialogFlag }

      batchAction(postForm).then(() => {
        this.$notify({
          title: '成功',
          message: '批量操作成功！',
          type: 'success',
          duration: 2000
        })

        this.dialogVisible = false
        this.$refs.pagingTable.getList()
      })
    },


    showImport() {
      this.importVisible = true
    },

  }
}
</script>

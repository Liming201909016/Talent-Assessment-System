<template>

  <div>

    <data-table
      ref="pagingTable"
      :options="options"
      :list-query="listQuery"
    >
      <template slot="filter-content" style="display: flex; align-items: flex-start">

        <exam-select v-model="listQuery.params.examId" class="filter-item" />

        <depart-tree-select v-model="listQuery.params.departId" class="el-select filter-item el-select--medium " :options="treeData" :props="defaultProps" width="200px" />
        <el-select v-model="listQuery.params.state" placeholder="考试状态" class="filter-item" clearable>
          <el-option
            v-for="item in paperStates"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>

      </template>

      <template slot="data-columns">

        <el-table-column
          label="考试名称"
          align="center"
          prop="title"
        >

          <template slot-scope="scope">
            <router-link :to="{ name: 'ShowExam', params: {id : scope.row.id}}">
              {{ scope.row.title }}
            </router-link>

          </template>

        </el-table-column>

        <el-table-column
          label="人员"
          align="center"
          prop="userId_dictText"
        />

        <el-table-column
          label="部门"
          align="center"
          prop="departId_dictText"
        />

        <el-table-column
          label="考试时长(分钟)"
          align="center"
          prop="totalTime"
        >

          <template slot-scope="scope">
            {{ scope.row.userTime }} / {{ scope.row.totalTime }}
          </template>

        </el-table-column>

        <el-table-column
          label="考试得分"
          align="center"
        >

          <template slot-scope="scope">
            {{ scope.row.userScore }} / {{ scope.row.totalScore }}
          </template>

        </el-table-column>

        <el-table-column
          label="考试时间"
          align="center"
          prop="createTime"
          width="180px"
        >
          <template slot-scope="scope">
            <span>{{ parseTime(scope.row.createTime, '{y}-{m}-{d} {h}:{i}') }}</span>
          </template>
        </el-table-column>

        <el-table-column
          label="考试结果"
          align="center"
        >

          <template slot-scope="scope">
            <span v-if="scope.row.state===1">待阅卷</span>
            <span v-else-if="scope.row.state===0">待交卷</span>
            <span v-else-if="scope.row.qualifyScore === 0">已完成</span>
            <span v-else>

              <span v-if="scope.row.userScore >= scope.row.qualifyScore" style="color:#00ff00">合格</span>
              <span v-else style="color: #ff0000">不合格</span>
            </span>
          </template>

        </el-table-column>

        <el-table-column
          label="考试状态"
          align="center"
        >

          <template slot-scope="scope">
            <span v-if="scope.row.state===0">考试中</span>
            <span v-else-if="scope.row.state===1">待阅卷</span>
            <span v-else-if="scope.row.state===2">已考完</span>
            <span v-else-if="scope.row.state===3" style="color:#909399">已弃考</span>
            <span v-else>{{ scope.row.state }}</span>
          </template>

        </el-table-column>

        <el-table-column label="操作" align="center" width="100">
          <template slot-scope="scope">
            <el-button type="text" size="mini" icon="el-icon-delete" style="color:#F56C6C" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>

      </template>

    </data-table>


  </div>

</template>

<script>
import DataTable from '@/components/DataTable'
import DepartTreeSelect from '@/components/DepartTreeSelect'
import { fetchTree } from '@/api/sys/depart/depart'
import ExamSelect from '@/components/ExamSelect'
import request from '@/utils/request'

export default {
  components: { ExamSelect, DepartTreeSelect, DataTable },

  data() {
    return {
      paperStates: [
        { value: 0, label: '考试中' },
        { value: 1, label: '待阅卷' },
        { value: 2, label: '已考完' },
        { value: 3, label: '!已弃考' }
      ],
      treeData: [],
      defaultProps: {
        value: 'id',
        label: 'deptName',
        children: 'children'
      },

      listQuery: {
        current: 1,
        size: 20,
        params: {
          examId: ''
        }
      },

      options: {

        // 可批量操作
        multi: false,
        // 列表请求URL
        listUrl: '/exam/api/paper/paper/paging'
      }
    }
  },

  created() {
    const examId = this.$route.params.examId

    if (typeof examId !== 'undefined') {
      this.listQuery.params.examId = examId
    }

    fetchTree({}).then(response => {
      this.treeData = response.data
    })
  },
  methods: {
    handleDelete(row) {
      this.$confirm(`确定删除该测试记录？将级联删除答题记录，且无法恢复。`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        request.post('/exam/api/paper/paper/delete', { ids: [row.id] }).then(res => {
          if (res.code === 0 || res.code === 1 || res.success) {
            this.$message.success('删除成功')
            this.$refs.pagingTable.getList()
          } else {
            this.$message.error(res.msg || '删除失败')
          }
        }).catch(err => {
          this.$message.error('删除失败：' + (err.message || err))
        })
      }).catch(() => {})
    }
  }
}
</script>

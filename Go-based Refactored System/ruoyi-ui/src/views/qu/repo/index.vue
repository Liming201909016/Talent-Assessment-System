<template>

  <data-table
    ref="pagingTable"
    :options="options"
    :list-query="listQuery"
  >
    <template slot="filter-content">

      <el-input v-model="listQuery.params.title" placeholder="查询题库名称" style="width: 200px;" class="filter-item" />

    </template>

    <template slot="data-columns">
      <el-table-column
        label="题目编码"
        prop="code"
        align="center"
      />

      <el-table-column
        label="题库名称"
      >

        <template slot-scope="data">

          <el-button v-if="isCompetencyRepo(data.row)" type="text" @click="handleQuestionManagement(data.row)">
            {{ data.row.title }}
          </el-button>
          <router-link v-else :to="{ name: 'UpdateRepo', params:{id: data.row.id}}">{{ data.row.title }}</router-link>

        </template>

      </el-table-column>

      <el-table-column
        label="题目数量"
        prop="radioCount"
        align="center"
      />

      <el-table-column
        label="多选题数量"
        prop="multiCount"
        align="center"
        v-if="flag"
      />

      <el-table-column
        label="判断题数量"
        prop="judgeCount"
        align="center"
        v-if="flag"
      />

      <el-table-column
        label="创建时间"
        align="center"
        width="180px"
      >
        <template slot-scope="scope">{{ parseTime(scope.row.createTime) }}</template>
      </el-table-column>

      <el-table-column label="操作" align="center" width="120">
        <template slot-scope="scope">
          <el-button
            size="mini"
            type="text"
            icon="el-icon-document"
            @click="handleQuestionManagement(scope.row)"
          >题目管理</el-button>
        </template>
      </el-table-column>

    </template>

  </data-table>

</template>

<script>
import DataTable from '@/components/DataTable'

export default {
  name: 'QuList',
  components: { DataTable },
  data() {
    return {

      flag:false,

      listQuery: {
        current: 1,
        size: 20,
        params: {
          title: ''
        }
      },

      options: {

        // 可批量操作
        multi: true,
        selectable: row => !row.virtual,

        // 批量操作列表
        multiActions: [
          {
            value: 'delete',
            label: '删除'
          }
        ],
        // 列表请求URL
        listUrl: '/exam/api/repo/paging',
        // 删除请求URL
        deleteUrl: '/exam/api/repo/delete',
        // 启用禁用
        stateUrl: '/qu/repo/state',
        // 添加数据路由
        addRoute: 'AddRepo'
      }
    }
  },
  methods: {
    isCompetencyRepo(row) {
      return row && (row.virtual === true || row.code === '00401')
    },
    handleQuestionManagement(row) {
      if (this.isCompetencyRepo(row)) {
        this.$router.push({ name: 'CompetencyQuestionList' })
        return
      }
      this.$router.push({ name: 'RepoQuList', params: { repoId: row.id, repoTitle: row.title } })
    }
  }
}
</script>

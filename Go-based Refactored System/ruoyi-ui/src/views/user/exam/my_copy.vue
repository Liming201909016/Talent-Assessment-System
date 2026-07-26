<template>

  <div>
    <data-table
      ref="pagingTable"
      :options="options"
      :list-query="listQuery"
    >
      <template slot="filter-content">

        <el-input v-model="listQuery.params.title" placeholder="查询考试名称" style="width: 200px;" class="filter-item" />

      </template>

      <template slot="data-columns">

        <el-table-column
          label="考试名称"
          prop="title"
          show-overflow-tooltip
        />

        <el-table-column
          label="考试次数"
          prop="tryCount"
          align="center"
        />

        <el-table-column
          label="最高分"
          prop="maxScore"
          align="center"
          v-if="flag"
        />

        <el-table-column
          label="是否通过"
          align="center"
        >

          <template slot-scope="scope">
            <span v-if="scope.row.passed" style="color: #00ff00;">通过</span>
            <span v-else style="color: #ff0000;">未通过</span>
          </template>

        </el-table-column>

        <el-table-column
          label="最后考试时间"
          prop="updateTime"
          align="center"
        />


      </template>

    </data-table>


  </div>

</template>

<script>
import DataTable from '@/components/DataTable'
import MyPaperList from './paper'
import { mapGetters } from 'vuex'

export default {
  name: 'MyExamList',
  components: { MyPaperList, DataTable },
  data() {
    return {

      flag: false,

      dialogVisible: false,
      examId: '',

      listQuery: {
        current: 1,
        size: 20,
        params: {
          title: ''
        }
      },

      options: {
        // 可批量操作
        multi: false,
        // 列表请求URL
        listUrl: '/exam/api/user/exam/my-paging'
      }
    }
  },
  computed: {
    ...mapGetters([
      'userId'
    ])
  },
  methods: {

    // 开始考试
    handleExamDetail(examId) {
      this.examId = examId
      this.dialogVisible = true
    },


  }
}
</script>

<style scoped>

  .el-dialog-div{
    height: 60vh;
    overflow: auto;
    padding: 10px;
  }

</style>

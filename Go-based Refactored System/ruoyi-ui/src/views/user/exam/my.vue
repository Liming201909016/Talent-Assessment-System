<template>
  <div class="app-container">
    <el-form :model="queryParams" ref="queryForm" size="small" :inline="true" v-show="showSearch" label-width="68px">
      <el-form-item label="姓名" prop="name">
        <el-input
          v-model="queryParams.name"
          placeholder="请输入姓名"
          clearable
          @keyup.enter.native="handleQuery"
        />
      </el-form-item>
      <el-form-item label="电话" prop="telephone">
        <el-input
          v-model="queryParams.telephone"
          placeholder="请输入电话"
          clearable
          @keyup.enter.native="handleQuery"
        />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">查询</el-button>
        <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
      </el-form-item>
    </el-form>

    <el-row :gutter="10" class="mb8">
      <right-toolbar :showSearch.sync="showSearch" @queryTable="getList"></right-toolbar>
    </el-row>

    <el-table border v-loading="loading" :data="testerList">
      <el-table-column width="55" align="center" />
      <el-table-column label="姓名" align="center" prop="name" />
      <el-table-column label="年龄" align="center" prop="age" />
      <el-table-column label="性别" align="center" width="60">
        <template slot-scope="scope">{{ scope.row.gender === '0' || scope.row.gender === 0 ? '男' : scope.row.gender === '1' || scope.row.gender === 1 ? '女' : '' }}</template>
      </el-table-column>
      <el-table-column label="电话号码" align="center" prop="telephone" />
      <el-table-column label="测评开始时间" align="center" width="170">
        <template slot-scope="scope">{{ parseTime(scope.row.createTime) }}</template>
      </el-table-column>
      <el-table-column label="测评结束时间" align="center" width="170">
        <template slot-scope="scope">{{ parseTime(scope.row.endTime) }}</template>
      </el-table-column>
<!--      <el-table-column label="测评时长" align="center" prop="userTime"/>-->
      <el-table-column label="操作" align="center" class-name="small-padding fixed-width">
        <template slot-scope="scope">
          <el-button
            size="mini"
            type="text"
            @click="viewPdf(scope.row)"
          >查看报告</el-button>
          <el-button
            size="mini"
            type="text"
            @click="downloadPdf(scope.row)"
          >下载报告</el-button>
        </template>
      </el-table-column>
    </el-table>

    <pagination
      v-show="total>0"
      :total="total"
      :page.sync="queryParams.pageNum"
      :limit.sync="queryParams.pageSize"
      @pagination="getList"
    />

    <el-dialog style="visibility: hidden" :visible.sync="dialogVisible" width="80%" append-to-body>
<!--      <result v-if="dialogVisible" ref="testResult" :disPdfDownload="disPdfDownload"></result>-->
      <result v-if="dialogVisible" ref="testResult"></result>
    </el-dialog>
  </div>
</template>

<script>
import { listTester } from "@/api/tester/tester";
import result from "@/views/paper/exam/result.vue";
import {pdfDownload} from "@/api/candidate/candidate";

export default {
  name: "Post",
  components: {result},
  data() {
    return {

      disPdfDownload: false,
      dialogVisible: false,
      // 遮罩层
      loading: true,
      // 选中数组
      ids: [],
      // 非单个禁用
      single: true,
      // 非多个禁用
      multiple: true,
      // 显示查询条件
      showSearch: true,
      // 总条数
      total: 0,
      // 测试者表格数据
      testerList: [],
      // 弹出层标题
      title: "",
      // 是否显示弹出层
      open: false,
      // 查询参数
      queryParams: {
        pageNum: 1,
        pageSize: 20,
        name: undefined,
        telephone: undefined,
      },
      // 表单参数
      form: {},
      // 表单校验
      rules: {
        name: [
          { required: true, message: "姓名不能为空", trigger: "blur" }
        ],
        telephone: [
          { required: true, message: "电话不能为空", trigger: "blur" }
        ],
      }
    };
  },
  created() {
    this.getList();
  },
  methods: {
    /** 查询岗位列表 */
    getList() {
      this.loading = true;
      listTester(this.queryParams).then(response => {
        this.testerList = response.rows;
        this.total = response.total;
        this.loading = false;
        // if (this.testerList.userTime === null) {
        //   this.testerList.userTime = 0
        // }

        console.log(this.testerList)
      });
    },
    // 取消按钮
    cancel() {
      this.open = false;
      this.reset();
    },
    // 表单重置
    reset() {
      this.form = {
        id: undefined,
        name: undefined,
        telephone: undefined,
      };
      this.resetForm("form");
    },
    /** 查询按钮操作 */
    handleQuery() {
      this.queryParams.pageNum = 1;
      this.getList();
    },
    /** 重置按钮操作 */
    resetQuery() {
      this.resetForm("queryForm");
      this.handleQuery();
    },
    /** 查看报告操作 */
    viewPdf(row) {
      // console.log(row)
      const paperId = row.paperId
      this.dialogVisible = true;
      this.$nextTick(()=>{
        // this.$refs.testResult.setPaperId(paperId);
        this.$refs.testResult.fetchTester(paperId);
        this.$refs.testResult.fetchScore(paperId);
      })
    },

    /** 下载报告操作 */
    downloadPdf(row) {
      let pdfPath = row.pdfPath
      console.log("==========================", row)
      let file = pdfPath;
      let formData = new FormData();
      formData.append("file", file);
      pdfDownload(formData).then(response => {
        // console.log(response.headers)
        let arr = pdfPath.split("/");
        let filename = arr[arr.length - 1]
        let url = window.URL.createObjectURL(response)
        //创建a标签 并设置属性
        let link = document.createElement('a')
        link.style.display = 'none'
        link.href = url
        link.setAttribute('download', filename)
        //添加a标签
        document.body.appendChild(link)
        //执行下载
        link.click();
        //释放url对象
        URL.revokeObjectURL(link.href);
        //释放a标签
        document.body.removeChild(link);
      })
    },

  }
};
</script>

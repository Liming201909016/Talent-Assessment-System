<template>
  <div class="app-container">
    <el-form :model="queryParams" ref="queryForm" size="small" :inline="true" v-show="showSearch" label-width="80px">
      <el-form-item label="测评名称" prop="examId">
        <el-select
          v-model="queryParams.examId"
          filterable
          clearable
          placeholder="请选择测评"
          style="width: 220px"
          @change="handlerChange"
        >
          <el-option
            v-for="item in examList"
            :key="item.id"
            :label="item.title"
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="姓名" prop="name">
        <el-input
          v-model="queryParams.name"
          placeholder="请输入姓名"
          clearable
          style="width: 140px"
          @keyup.enter.native="handleQuery"
        />
      </el-form-item>
      <el-form-item label="手机号" prop="telephone">
        <el-input
          v-model="queryParams.telephone"
          placeholder="请输入手机号"
          clearable
          style="width: 140px"
          @keyup.enter.native="handleQuery"
        />
      </el-form-item>
      <el-form-item label="是否学生" prop="stuFlag">
        <el-select v-model="queryParams.stuFlag" clearable placeholder="全部" style="width: 100px">
          <el-option label="是" :value="1" />
          <el-option label="否" :value="0" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">查询</el-button>
        <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
      </el-form-item>
    </el-form>

    <el-row :gutter="10" class="mb8">
      <el-col :span="1.5">
        <el-button type="primary" icon="el-icon-plus" size="mini" @click="handleAdd">新增</el-button>
        <el-button type="info" icon="el-icon-upload2" size="mini" @click="handleExcel">导入</el-button>
        <el-button type="warning" icon="el-icon-download" size="mini" @click="handleExportExcel">导出</el-button>
        <el-button type="success" icon="el-icon-download" size="mini" @click="handleExcelTemplate">模板下载</el-button>
      </el-col>
      <right-toolbar :showSearch.sync="showSearch" @queryTable="getList"></right-toolbar>
    </el-row>

    <el-table v-loading="loading" :data="testerList" border @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="50" align="center" />
      <el-table-column label="测评名称" align="center" min-width="220" prop="title" show-overflow-tooltip>
        <template slot-scope="scope">
          {{ scope.row.title || scope.row.examId }}
        </template>
      </el-table-column>
      <el-table-column label="姓名" align="center" width="90" prop="name" />
      <el-table-column label="手机号" align="center" width="120" prop="telephone" show-overflow-tooltip />
      <el-table-column label="密码" align="center" width="100" prop="password" show-overflow-tooltip />
      <el-table-column label="身份证号" align="center" width="175" prop="idNumber" show-overflow-tooltip />
      <el-table-column label="性别" align="center" width="60" prop="gender">
        <template slot-scope="scope">
          {{ scope.row.gender === '0' || scope.row.gender === 0 ? '男' : scope.row.gender === '1' || scope.row.gender === 1 ? '女' : '' }}
        </template>
      </el-table-column>
      <el-table-column label="年龄" align="center" width="60" prop="age" />
      <el-table-column label="是否学生" align="center" width="80" prop="stuFlag">
        <template slot-scope="scope">
          {{ scope.row.stuFlag === 1 ? '是' : '否' }}
        </template>
      </el-table-column>
      <el-table-column label="单位/学校" align="center" min-width="100" prop="affiliation" show-overflow-tooltip />
      <el-table-column label="职务/岗位" align="center" min-width="100" prop="post" show-overflow-tooltip />
      <el-table-column label="操作" align="center" width="120" fixed="right">
        <template slot-scope="scope">
          <el-button size="mini" type="text" icon="el-icon-edit" @click="handleUpdate(scope.row)">修改</el-button>
          <el-button size="mini" type="text" icon="el-icon-delete" @click="handleDelete(scope.row)">删除</el-button>
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

    <el-dialog :title="upload.title" :visible.sync="upload.open" width="400px" append-to-body>
      <el-upload
        ref="upload"
        :limit="upload.fileLimit"
        accept=".xlsx, .xls"
        :headers="upload.headers"
        :action="upload.url"
        :on-progress="handleFileUploadProgress"
        :on-success="handleFileSuccess"
        :file-list="upload.fileList"
        :auto-upload="false"
        :data="{changeExamId:changeExamId}"
        drag
      >
        <i class="el-icon-upload"></i>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        <div class="el-upload__tip text-center" slot="tip">
          <span>仅允许导入xls、xlsx格式文件。</span>
          <el-select
            v-model="changeExamId"
            filterable
            clearable
            placeholder="选择考试"
          >
            <el-option
              v-for="item in examList"
              :key="item.id"
              :label="item.title"
              :value="item.id"
            />
          </el-select>
        </div>
        <!--        <el-button class="btn"><i class="el-icon-paperclip"></i>上传文件</el-button>-->
      </el-upload>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" @click="submitFileForm">确 定</el-button>
        <el-button @click="upload.open = false; upload.isUploading = true;">取 消</el-button>
      </div>
    </el-dialog>

    <!-- 添加或修改用户对话框 -->
    <el-dialog :title="title" :visible.sync="open" width="760px" append-to-body>
      <el-form ref="form" :model="form" :rules="rules" label-width="130px" class="tester-form">
        <el-form-item label="测评名称" prop="examId">
          <el-select
            v-model="form.examId"
            filterable
            clearable
            placeholder="请选择测评"
            style="width: 100%"
          >
            <el-option
              v-for="item in examList"
              :key="item.id"
              :label="item.title"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="手机号" prop="telephone">
          <el-input v-model="form.telephone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="身份证号" prop="idNumber">
          <el-input v-model="form.idNumber" placeholder="请输入身份证号（选填）" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" placeholder="请输入密码（选填，默认手机号后4位）" />
        </el-form-item>
        <el-form-item label="年龄" prop="age">
          <el-input v-model="form.age" placeholder="请输入年龄" />
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-select v-model="form.gender" placeholder="请选择" >
            <el-option value="0" label="男"></el-option>
            <el-option value="1" label="女"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="是否学生" prop="stuFlag" class="tester-flag-item">
          <el-radio-group v-model="form.stuFlag">
            <el-radio size="large" border :label="1">是</el-radio>
            <el-radio size="large" border :label="0">否</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="form.stuFlag === 1 ? '学校' : '单位'" prop="affiliation">
              <el-input
                v-model="form.affiliation"
                :placeholder="form.stuFlag === 1 ? '请输入学校' : '请输入单位'"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="部门" prop="depart">
              <el-input v-model="form.depart" placeholder="请输入部门" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="职务/岗位" prop="post">
              <el-input v-model="form.post" placeholder="请输入职务/岗位" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="学历" prop="degree">
              <el-input v-model="form.degree" placeholder="请输入学历" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="专业" prop="major">
              <el-input v-model="form.major" placeholder="请输入专业" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" @click="submitForm">确 定</el-button>
        <el-button @click="cancel">取 消</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getListTester, getTester, delTester, addTester, updateTester } from "@/api/tester/tester";
import { fetchList } from '@/api/exam/exam'
import ExamSelect from "@/components/ExamSelect/index.vue";
import { getToken } from "@/utils/auth";

const CLOSED_EXAM_TYPE = 2;

export default {
  name: "Tester",
  components: {ExamSelect},
  data() {
    return {
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
      // 用户表格数据
      testerList: [],
      // 弹出层标题
      title: "",
      // 是否显示弹出层
      open: false,

      multi: Boolean,
      value: Array,
      examList:[],
      // dataList: [],
      currentValue: [],
      changeExamId: '',

      upload: {
        fileList: [],
        fileLimit: 1,
        // 是否显示弹出层（用户导入）
        open: false,
        // 弹出层标题（用户导入）
        title: "",
        // 是否禁用上传
        isUploading: false,
        // 是否更新已经存在的用户数据
        updateSupport: 0,
        // 请求头（携带 token）
        headers: { Authorization: "Bearer " + getToken() },
        url: process.env.VUE_APP_BASE_API + "/exam/api/tester/importData"
      },
      // 查询参数
      queryParams: {
        pageNum: 1,
        pageSize: 20,
        examId: '',
        name: null,
        telephone: null,
        stuFlag: null,
      },
      // 表单参数
      form: {},
      // 表单校验
      rules: {
        name: [
          { required: true, message: "姓名不能为空", trigger: "blur" }
        ],
        telephone: [
          { required: true, message: "手机号不能为空", trigger: "blur" }
        ],
        examId: [
          { required: true, message: "请选择测评", trigger: "change" }
        ],
        stuFlag: [
          { required: true, message: "请选择是否学生", trigger: "change" }
        ],
      }
    };
  },
  created() {
    this.getList();
    this.currentValue = this.value
    this.fetchExamData()
  },
  watch: {
    // 检测查询变化
    value: {
      handler() {
        this.currentValue = this.value
      }
    }
  },

  methods: {

    handleExcelTemplate() {
      this.download('/exam/api/tester/importTemplate', {}, `tester_template_${new Date().getTime()}.xlsx`)
    },

    fetchExamData() {
      fetchList().then(response => {
        this.examList = (response.data.records || []).filter(item => item.isOpen === CLOSED_EXAM_TYPE)
      })
    },

    hasClosedExam(examId) {
      return this.examList.some(item => item.id === examId)
    },

    handlerChange(e) {
      console.log(e)

      this.$emit('change', e)
      this.$emit('input', e)
    },

    submitFileForm() {
      if (this.changeExamId !== '' && this.changeExamId !== null) {
        // this.upload.examId = this.changeExamId
        // console.log(this.upload.examId)
        this.$refs.upload.submit();
      } else {
        this.$modal.msgWarning("请选择考试");
      }
    },

    // 文件上传中处理
    handleFileUploadProgress(event, file, fileList) {
      this.upload.isUploading = true;
    },
    // 文件上传成功处理
    handleFileSuccess(response, file, fileList) {
      this.upload.open = false;
      this.upload.isUploading = false;
      this.$refs.upload.clearFiles();
      this.$alert("<div style='overflow: auto;overflow-x: hidden;max-height: 70vh;padding: 10px 20px 0;'>" + response.msg + "</div>", "导入结果", { dangerouslyUseHTMLString: true });
      this.getList()
    },

    handleExcel() {
      this.changeExamId = this.queryParams.examId || '';
      this.upload.title = "Excel文件导入";
      this.upload.open = true;
    },

    /** 查询用户列表 */
    getList() {
      this.loading = true;
      getListTester(this.queryParams).then(response => {
        this.testerList = response.rows;
        this.total = response.total;
        this.loading = false;
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
        examId: this.queryParams.examId || null,
        idNumber: null,
        name: null,
        password: null,
        age: null,
        gender: null,
        telephone: null,
        affiliation: null,
        depart: null,
        post: null,
        degree: null,
        major: null,
        stuFlag: null,
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
    // 多选框选中数据
    handleSelectionChange(selection) {
      this.ids = selection.map(item => item.id)
      this.single = selection.length!==1
      this.multiple = !selection.length
    },
    /** 新增按钮操作 */
    handleAdd() {
      this.reset();
      this.open = true;
      this.title = "添加测评人员";
    },
    /** 修改按钮操作 */
    handleUpdate(row) {
      this.reset();
      const id = row.id || this.ids
      getTester(id).then(response => {
        this.form = response.data;
        if (this.form.examId && !this.hasClosedExam(this.form.examId)) {
          this.form.examId = null;
          this.$modal.msgWarning("当前测评人员所属测评不是封闭测评，请重新选择封闭测评");
        }
        this.open = true;
        this.title = "修改测评人员";
      });
    },
    /** 提交按钮 */
    submitForm() {
      this.$refs["form"].validate(valid => {
        if (valid) {
          if (this.form.id != null) {
            updateTester(this.form).then(response => {
              this.$modal.msgSuccess("修改成功");
              this.open = false;
              this.getList();
            });
          } else {
            addTester(this.form).then(response => {
              this.$modal.msgSuccess("新增成功");
              this.open = false;
              this.getList();
            });
          }
        }
      });
    },
    /** 删除按钮操作 */
    handleDelete(row) {
      const ids = row.id || this.ids;
      this.$modal.confirm('是否确认删除用户编号为"' + ids + '"的数据项？').then(function() {
        return delTester(ids);
      }).then(() => {
        this.getList();
        this.$modal.msgSuccess("删除成功");
      }).catch(() => {});
    },
    /** 导出按钮操作 */
    handleExportExcel() {
      if (!this.queryParams.examId) {
        this.$modal.msgWarning("请先选择要导出的测评");
        return;
      }
      this.download('/exam/api/tester/export', {
        ...this.queryParams
      }, `tester_${this.queryParams.examId}_${new Date().getTime()}.xlsx`)
    }
  }
};
</script>

<style scoped>
.tester-form ::v-deep .el-form-item__label {
  line-height: 20px;
  white-space: normal;
}

.tester-flag-item ::v-deep .el-form-item__content {
  display: flex;
  align-items: flex-start;
}

.tester-flag-item ::v-deep .el-radio-group {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
</style>

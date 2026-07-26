<template>
  <div class="app-container">

    <div class="filter-container">

      <slot name="filter-content" />

      <el-row>
        <el-col>
          <el-button v-if="options.addRoute" type="primary" icon="el-icon-plus" @click="handleAdd">添加</el-button>
          <el-button v-if="options.importFile" type="warning" icon="el-icon-plus" @click="handleExcel">导入题目</el-button>
        </el-col>
      </el-row>

    </div>

    <div v-show="multiShow && options.multiActions" class="filter-container">

      <el-select v-model="multiNow" :placeholder="selectedLabel" class="filter-item" style="width: 130px" @change="handleOption">
        <el-option
          v-for="item in options.multiActions"
          :key="item.value"
          :label="item.label"
          :value="item.value"
        />
      </el-select>

    </div>

    <el-table
      v-loading="listLoading"
      :data="pagedRecords"
      border
      fit
      highlight-current-row
      :header-cell-style="{'background':'#f2f3f4', 'color':'#555', 'font-weight':'bold', 'line-height':'32px'}"
      @selection-change="handleSelection"
    >

      <el-table-column
        v-if="options.multi"
        align="center"
        type="selection"
        :selectable="options.selectable"
        width="55"
      />

      <slot name="data-columns" />

    </el-table>

    <pagination v-show="dataList.total>0" :total="dataList.total" :page.sync="listQuery.current" :limit.sync="listQuery.size" @pagination="getList" />

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
        drag
      >
        <i class="el-icon-upload"></i>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        <div class="el-upload__tip text-center" slot="tip">
          <span>仅允许导入xls、xlsx格式文件。</span>
        </div>
<!--        <el-button class="btn"><i class="el-icon-paperclip"></i>上传文件</el-button>-->
      </el-upload>
      <div slot="footer" class="dialog-footer">
        <el-button type="primary" @click="submitFileForm">确 定</el-button>
        <el-button @click="upload.open = false; upload.isUploading = true;">取 消</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { fetchList, deleteData, changeState } from '@/api/common'
import Pagination from '@/components/Pagination'
import {getToken} from "@/utils/auth";

export default {
  name: 'PagingTable',
  components: { Pagination },
  // 组件入参
  props: {
    options: {
      type: Object,
      default: () => {
        return {
          // 批量操作
          multiActions: [],
          // 列表请求URL
          listUrl: '/exam/api',
          // 删除请求URL
          deleteUrl: '',
          // 启用禁用
          stateUrl: '',
          // 可批量操作
          multi: false
        }
      }
    },

    // 列表查询参数
    listQuery: {
      type: Object,
      default: () => {
        return {
          current: 1,
          size: 20,
          params: {},
          t: 0
        }
      }
    }
  },
  data() {
    return {

      upload: {
        fileList: [],
        fileLimit: 1,
        headers: { Authorization: 'Bearer ' + getToken() },
        // 是否显示弹出层（用户导入）
        open: false,
        // 弹出层标题（用户导入）
        title: "",
        // 是否禁用上传
        isUploading: false,
        // 是否更新已经存在的用户数据
        updateSupport: 0,
        // 设置上传的请求头部
        // headers: { "Content-Type": "multipart/form-data" },
        // 上传地址必须跟随当前构建环境的 API 前缀。
        url: (process.env.VUE_APP_BASE_API || '') + "/exam/api/qu/qu/import-excel"
      },

      // 接口数据返回
      dataList: {
        total: 0
      },
      // 数据加载标识
      listLoading: true,
      // 选定和批量操作
      selectedIds: [],
      selectedObjs: [],
      // 显示已中多少项
      selectedLabel: '',
      // 显示批量操作
      multiShow: false,
      // 批量操作的标识
      multiNow: ''
    }
  },

  computed: {
    pagedRecords() {
      const records = (this.dataList && this.dataList.records) || []
      // 如果后端已正确分页（total > 0 且 records.length <= size），直接返回
      if (this.dataList.total > 0 && records.length <= this.listQuery.size) {
        return records
      }
      // 否则做前端分页
      const start = (this.listQuery.current - 1) * this.listQuery.size
      return records.slice(start, start + this.listQuery.size)
    }
  },

  watch: {

    // 检测查询变化
    listQuery: {
      handler() {
        this.getList()
      },
      deep: true
    }
  },
  created() {
    this.getList()
  },
  methods: {

    handleUpload() {

    },

    submitFileForm() {
      this.$refs.upload.submit();
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
      this.upload.title = "Excel文件导入";
      this.upload.open = true;
    },

    /**
     * 添加数据跳转
     */
    handleAdd() {
      if (this.options.addRoute) {
        this.$router.push({ name: this.options.addRoute, params: {}})
        return
      }
      console.log('未设置添加数据跳转路由！')
    },

    /**
     * 查询数据列表
     */
    getList() {
      this.listLoading = true
      this.listQuery.t = new Date().getTime()
      fetchList(this.options.listUrl, this.listQuery).then(response => {
        this.dataList = response.data
        // 后端分页total可能为0，用records长度补偿
        if (this.dataList && this.dataList.total === 0 && this.dataList.records && this.dataList.records.length > 0) {
          this.dataList.total = this.dataList.records.length
        }
        console.log(this.dataList)
        this.listLoading = false
      })
    },

    /**
     * 查询
     */
    handleFilter() {
      // 重新查询
      this.getList()
    },

    /**
     * 批量操作回调
     */
    handleOption(v) {
      this.multiNow = ''

      // 内部消化的操作
      if (v === 'delete') {
        this.handleDelete()
        return
      }

      if (v === 'enable') {
        this.handleState(0)
        return
      }

      if (v === 'disable') {
        this.handleState(1)
        return
      }

      // 向外回调的操作
      this.$emit('multi-actions', { opt: v, ids: this.selectedIds })
    },

    /**
     * 修改状态，启用禁用
     */
    handleState(state) {
      // 修改状态
      changeState(this.options.stateUrl, this.selectedIds, state).then(response => {
        if (response.code === 0) {
          this.$message({
            type: 'success',
            message: '状态修改成功!'
          })

          // 重新查询
          this.getList()
        }
      })
    },

    /**
     * 删除数据
     */
    handleDelete() {
      if (this.selectedIds.length === 0) {
        this.$message({
          message: '请至少选择一条数据！',
          type: 'warning'
        })
        return
      }

      // 删除
      this.$confirm('确实要删除吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        deleteData(this.options.deleteUrl, this.selectedIds).then(() => {
          this.$message({
            type: 'success',
            message: '删除成功!'
          })
          this.getList()
        })
      })
    },

    /**
     * 列表多选操作
     * @param val
     */
    handleSelection(val) {
      const ids = []
      val.forEach(row => {
        ids.push(row.id)
      })

      this.selectedObjs = val
      this.selectedIds = ids
      this.multiShow = ids.length > 0
      this.selectedLabel = '已选' + ids.length + '项'

      this.$emit('select-changed', { ids: this.selectedIds, objs: this.selectedObjs })
    }

  }
}
</script>

<style>

  .filter-container .filter-item{
    margin-left: 5px;
  }

  .filter-container .filter-item:first-child{
    margin-left: 0px;
  }
</style>

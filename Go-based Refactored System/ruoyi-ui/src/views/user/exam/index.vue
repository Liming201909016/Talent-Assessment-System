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
      <el-form-item label="状态" prop="examStatus">
        <el-select v-model="queryParams.examStatus" clearable placeholder="全部" style="width: 120px;">
          <el-option label="已完成" value="2" />
          <el-option label="进行中" value="1" />
          <el-option label="未测评" value="0" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">查询</el-button>
        <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
      </el-form-item>
    </el-form>

    <div class="mb8" style="display:flex; flex-wrap:wrap; gap:8px; justify-content:flex-end;">
      <div style="display:flex; gap:4px;">
        <el-button type="primary" plain icon="el-icon-document" size="mini" :disabled="multiple" @click="batchCreatePdf">批量生成报告</el-button>
        <el-button type="success" plain icon="el-icon-edit" size="mini" :disabled="multiple" @click="handleBatchDownload">批量下载</el-button>
        <el-button type="warning" plain icon="el-icon-document-copy" size="mini" :disabled="multiple || !isMbtiExam" @click="handleBatchDownloadSimple">批量下载简版</el-button>
        <el-button type="danger" plain icon="el-icon-delete" size="mini" :disabled="multiple" @click="handleDelete">批量删除</el-button>
      </div>
      <div style="display:flex; gap:4px;">
        <el-button type="success" size="mini" @click="deletePdf">删除报告</el-button>
      </div>
      <div style="display:flex; gap:4px;">
        <el-button type="primary" size="mini" @click="viewTeamPdf">查看团队报告</el-button>
        <el-button type="warning" size="mini" @click="downloadTeamPdf">下载团队报告</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="testerList" border @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="50" align="center" />
      <el-table-column label="姓名" align="center" prop="name" min-width="80" show-overflow-tooltip />
      <el-table-column label="手机号" align="center" prop="telephone" min-width="110" />
      <el-table-column v-if="isOpen === 2 || isOpen === '2'" label="身份证" align="center" prop="idNumber" min-width="160" show-overflow-tooltip />
      <el-table-column label="性别" align="center" width="50" prop="gender">
        <template slot-scope="scope">
          {{ scope.row.gender === '0' || scope.row.gender === 0 ? '男' : scope.row.gender === '1' || scope.row.gender === 1 ? '女' : '' }}
        </template>
      </el-table-column>
      <el-table-column label="年龄" align="center" width="50" prop="age" />
      <el-table-column label="状态" align="center" width="70">
        <template slot-scope="scope">
          <el-tag v-if="scope.row.endTime" type="success" size="mini">已完成</el-tag>
          <el-tag v-else-if="scope.row.paperId" type="warning" size="mini">进行中</el-tag>
          <el-tag v-else type="info" size="mini">未测评</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="开始时间" min-width="135" align="center">
        <template slot-scope="scope">
          {{ parseTime(scope.row.createTime, '{y}-{m}-{d} {h}:{i}') }}
        </template>
      </el-table-column>
      <el-table-column label="结束时间" min-width="135" align="center">
        <template slot-scope="scope">
          {{ scope.row.endTime ? parseTime(scope.row.endTime, '{y}-{m}-{d} {h}:{i}') : '' }}
        </template>
      </el-table-column>
      <el-table-column label="时长" align="center" width="60">
        <template slot-scope="scope">
          <span v-if="scope.row.endTime && scope.row.userTime > 0">{{ scope.row.userTime }}分</span>
          <span v-else-if="scope.row.endTime">&lt;1分</span>
          <span v-else style="color:#909399">—</span>
        </template>
      </el-table-column>
      <el-table-column label="答题" align="center" width="50">
        <template slot-scope="scope">
          <span v-if="scope.row.paperId">{{ scope.row.answerNum }}</span>
          <span v-else style="color:#909399">—</span>
        </template>
      </el-table-column>
      <el-table-column label="报告" align="center" width="60">
        <template slot-scope="scope">
          <el-tag v-if="scope.row.pdfFlag === 1" type="success" size="mini" effect="plain">已生成</el-tag>
          <el-tag v-else type="info" size="mini" effect="plain">未生成</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" class-name="small-padding fixed-width" :width="opColumnWidth">
        <template slot-scope="scope">
          <el-button size="mini" type="text" icon="el-icon-view" :disabled="!scope.row.endTime" @click="viewPdf(scope.row)">查看</el-button>
          <el-button size="mini" type="text" icon="el-icon-tickets" :disabled="!scope.row.paperId" @click="viewAnswerDetail(scope.row)">答题详情</el-button>
          <template v-if="scope.row.repoCode && scope.row.repoCode.startsWith('003')">
            <el-button size="mini" type="text" icon="el-icon-download" :disabled="scope.row.pdfFlag !== 1" @click="handleReportCommand('full', scope.row)">完整版</el-button>
            <el-button size="mini" type="text" icon="el-icon-document" :disabled="scope.row.pdfFlag !== 1" @click="handleReportCommand('simple', scope.row)">简版</el-button>
          </template>
          <el-button v-else size="mini" type="text" icon="el-icon-download" :disabled="scope.row.pdfFlag !== 1" @click="downloadPdf(scope.row.pdfPath)">下载</el-button>
        </template>
      </el-table-column>
    </el-table>

    <pagination
      v-show="total>0"
      :total="total"
      :page.sync="queryParams.pageNum"
      :limit.sync="queryParams.pageSize"
      @pagination="getList(isOpen)"
    />

    <el-dialog :visible.sync="dialogVisible" width="80%" append-to-body @close="handleDinpmalogClose">
      <mbti-report-full v-if="dialogVisible && repoCode.startsWith('003')" ref="testResult" :disPdfDownload="disPdfDownload" :stuFlag="stuFlag" :repoCode="repoCode"></mbti-report-full>
      <result2 v-else-if="dialogVisible && repoCode.startsWith('002')" ref="testResult" :disPdfDownload="disPdfDownload" :stuFlag="stuFlag" :repoCode="repoCode"></result2>
      <result v-else-if="dialogVisible" ref="testResult" :disPdfDownload="disPdfDownload" :stuFlag="stuFlag" :repoCode="repoCode"></result>
    </el-dialog>

    <el-dialog  :visible.sync="visibleTeam" width="80%" append-to-body>
      <team v-if="visibleTeam" ref="teamResult" :examId="testerList[0].examId" :isOpen="$route.params.isOpen"></team>
    </el-dialog>

    <el-dialog title="答题详情" :visible.sync="answerDialogVisible" width="80%" append-to-body class="answer-detail-dialog">
      <div v-loading="answerLoading" style="min-height:200px;">
        <div v-if="answerDetail" style="margin-bottom:12px;color:#606266;">
          <strong>{{ answerDetail.name }}</strong>
          <span v-if="answerDetail.telephone">（{{ answerDetail.telephone }}）</span>
          <span style="margin-left:16px;">测评：{{ answerDetail.examTitle }}</span>
          <span style="margin-left:16px;">题库：{{ answerDetail.repoTitle }}</span>
        </div>
        <el-table v-if="answerDetail" :data="answerDetail.questions" border size="mini" max-height="600">
          <el-table-column label="题号" prop="sort" width="60" align="center" />
          <el-table-column label="V编号" prop="vCode" width="70" align="center" />
          <el-table-column label="题干" prop="title" v-if="answerDetail.questions.some(q=>q.title)" />
          <el-table-column label="选项">
            <template slot-scope="scope">
              <div v-for="(o, idx) in scope.row.options" :key="idx"
                   :style="{padding:'2px 4px', background: o.checked === 1 ? '#e1f3d8' : '', fontWeight: o.checked === 1 ? 600 : 400}">
                <span style="display:inline-block;width:18px;color:#909399;">{{ o.abc || (idx===0?'A':'B') }}.</span>
                {{ o.content }}
                <!-- MBTI: 显示考生分配的分数 (0~10) -->
                <el-tag v-if="answerDetail.isMbti" size="mini" type="primary" effect="plain" style="margin-left:8px;">分数：{{ o.score }}</el-tag>
                <!-- 非 MBTI: 已选标记 -->
                <el-tag v-else-if="o.checked === 1" size="mini" type="success" effect="plain" style="margin-left:8px;">已选</el-tag>
                <!-- 仅 001 心理特质显示锚点（002 用 score 计分，003 无锚点） -->
                <el-tag v-if="!answerDetail.isMbti && o.isRight === 1 && answerDetail.repoCode && answerDetail.repoCode.startsWith('001')"
                        size="mini" type="warning" effect="plain" style="margin-left:4px;">题库锚点</el-tag>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>

  </div>
</template>

<script>
import {
    batchDownloadTesterPdf,
    delTester,
    getListTester,
    listTester,
    logisticDelTester,
    updateTester
} from "@/api/tester/tester";
import result from "@/views/paper/exam/result.vue";
import result2 from "@/views/paper/exam/result2.vue";
import MbtiReportFull from "@/views/paper/exam/mbtiReportFull.vue";
import {
  batchDelReport,
  batchDownload,
  logicDeletePdfByIds,
  batchDownloadCandidatePdf, batchLogisticDelCandidate,
  delCandidate, logisticDelCandidate, logisticDelReport,
  pdfDownload,
  updateCandidate
} from "@/api/candidate/candidate";
import team from "@/views/user/exam/team.vue";
import {pdfTeamDownload} from "@/api/exam/exam";
import {Loading, Message} from "element-ui";
import request from '@/utils/request';

export default {
  name: "ListExamUser",
  components: {result,result2, team, MbtiReportFull},
  dicts: ['user_pdf_flag'],
  data() {
    return {
      // singlePdfFinished: false,
      disPdfDownload: false,
      dialogVisible: false,
      visibleTeam: false,
      answerDialogVisible: false,
      answerLoading: false,
      answerDetail: null,
      // 遮罩层
      loading: true,
      // 选中数组
      ids: [],
      // 非单个禁用
      single: true,
      selections: [],
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
        examId: undefined,
        examStatus: undefined,
      },
      isOpen: null,
      stuFlag: null,
      repoCode:null,
      testTitle: '0',
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
    this.queryParams.examId = this.$route.params.examId
    this.stuFlag = this.$route.params.stuFlag
    this.isOpen = this.$route.params.isOpen
    this.testTitle = this.$route.params.title
    this.getList(this.$route.params.isOpen);
  },
  computed:{
    singlePdfFinished :{
      get(){
        return this.$store.state.pdfStatus.singlePdfFinished;
      },
      set(value){
        this.$store.commit('setSinglePdfFinished', value);
      }
    },
    isMbtiExam() {
      return this.testerList.some(r => r.repoCode && r.repoCode.startsWith('003'))
    },
    opColumnWidth() {
      return this.isMbtiExam ? 240 : 160
    }
  },
  methods: {
    handleDinpmalogClose(){

    },
    /** 查看考生答题详情（题目+选项+考生选了哪个） */
    viewAnswerDetail(row) {
      if (!row.paperId) {
        this.$message.warning('该考生未生成试卷')
        return
      }
      this.answerDialogVisible = true
      this.answerLoading = true
      this.answerDetail = null
      request({
        url: '/exam/api/exam/exam/answer-detail',
        method: 'get',
        params: { paperId: row.paperId }
      }).then(res => {
        this.answerDetail = res.data || res
      }).catch(() => {}).finally(() => {
        this.answerLoading = false
      })
    },
    /** 查询岗位列表 */
    getList(isOpen) {

      console.log("isOpen", isOpen)
      this.loading = true;

      // listTester(this.queryParams).then(response => {
      //   // this.testerList = response.rows.filter(tester => tester.delFlag === 0);
      //   this.testerList = response.rows;
      //   // this.total = this.testerList.length;
      //   this.total = response.total;
      //   this.loading = false;
      //   console.log(this.testerList)
      // });

      // 1--开放  2--封闭， 这里貌似没什么区别
      if (isOpen === 1 || isOpen === "1") {
        listTester(this.queryParams).then(response => {
          // this.testerList = response.rows.filter(tester => tester.delFlag === 0);
          this.testerList = response.rows;
          this.total = response.total;
          this.loading = false;
          console.log(this.testerList)
        });
      } else if (isOpen === 2 || isOpen === "2") {
        getListTester(this.queryParams).then(response => {
          // this.testerList = response.rows.filter(tester => tester.delFlag === 0);
          this.testerList = response.rows;
          this.total = response.total;
          this.loading = false;
          console.log(response)
        });
      }
    },

    //删除报告
    deletePdf() {
      console.log("deletePdf")
      let ids = this.ids;
      //如果为空,则提示
      if (ids.length === 0) {
        this.$message.warning("请选择要删除的报告");
        return;
      }
      let data = JSON.stringify(ids)
      logicDeletePdfByIds(ids).then(() => {
        this.getList(this.isOpen);
        this.$modal.msgSuccess("删除报告成功");
      }).catch(() => {});
    },

    // 生成所有报告
    createAllPdf() {

      let downloadLoadingInstance = Loading.service({ text: "正在下载测评报告，请稍候", spinner: "el-icon-loading", background: "rgba(0, 0, 0, 0.7)", })
      let queryParams = this.queryParams
      // console.log("queryParams", queryParams)
      queryParams.pageSize = 10000;
      if (this.isOpen === 1 || this.isOpen === "1") {
        listTester(queryParams).then(async response => {
          let testers = response.rows;
          for (const tester of testers) {
            if (tester.pdfFlag === 1) {
              continue;
            }

            this.$store.state.pdfStatus.singlePdfFinished = false;
            // await this.viewPdf(tester)
            console.log('viewPdf')
            this.viewPdf(tester);
            console.log(tester);
            console.log("before singlePdfFinished", this.$store.state.pdfStatus.singlePdfFinished);
            // console.log('before')
            // console.log('this.dialogVisible')
            // console.log(this.dialogVisible)
            await this.waitForSinglePdf().then(() => {
              if (this.$store.state.pdfStatus.singlePdfFinished) {
                tester.pdfFlag = 1;
                tester.pdfPath = null
                updateCandidate(tester);
              }
            });
            console.log("after singlePdfFinished", this.$store.state.pdfStatus.singlePdfFinished);
            // this.$refs.testResult.visible =  false;
            // this.dialogVisible = false
            // console.log('after')
            // console.log('this.dialogVisible')
            // console.log(this.dialogVisible)
            // while(!this.$store.state.pdfStatus.singlePdfFinished){
            //   // continue;
            //   // console.log('waiting...')
            // }
          }
          // let result = await this.viewPdf(testers[0]);
          // if(result == 1){
          //   this.viewPdf(testers[1]);
          // }

          downloadLoadingInstance.close();
        });
      } else if (this.isOpen === 2 || this.isOpen === "2") {
        getListTester(queryParams).then(async response => {
          let testers = response.rows;
          for (const tester of testers) {

            if (tester.pdfFlag === 1) {
              continue;
            }

            this.$store.state.pdfStatus.singlePdfFinished = false;
            console.log('viewPdf')
            this.viewPdf(tester);
            console.log("before singlePdfFinished", this.$store.state.pdfStatus.singlePdfFinished);
            await this.waitForSinglePdf().then(() => {
              tester.pdfFlag = 1;
              tester.pdfPath = null
              updateTester(tester);
            });
            console.log("after singlePdfFinished", this.$store.state.pdfStatus.singlePdfFinished);
          }
          downloadLoadingInstance.close();
        });
      }
      setTimeout(() => {
        this.getList(this.isOpen)
      }, 2000)
    },

    async waitForSinglePdf(){
      while(this.$store.state.pdfStatus.singlePdfFinished === false){
        console.log('continue waiting...')
        await new Promise(resolve => setTimeout(resolve, 100));
      }
      this.dialogVisible = false
      console.log('pdf waitForSinglePdf')
    },

    handleDialogClose() {
      this.dialogVisible = false;
      this.getList(this.isOpen)
    },

    handleSelectionChange(selection) {
      this.selections = selection
      this.ids = selection.map(item => item.id);
      this.single = selection.length !== 1;
      this.multiple = !selection.length;
    },
    async batchCreatePdf(){
      // F: 上限 20 → 100（与后端 maxBatch 对齐）
      if(this.selections.length > 100){
        this.$message.warning("单次批量生成报告不能超过100份");
        return;
      }
      const filtered = this.selections.filter(t => t.paperId)
      if (filtered.length === 0) {
        this.$message.warning("没有可生成报告的考生");
        return;
      }

      // E: 已生成报告的二次确认
      const alreadyGen = filtered.filter(t => t.pdfFlag === 1)
      if (alreadyGen.length > 0) {
        let action
        try {
          await this.$confirm(
            `选中的 ${filtered.length} 位考生中，已有 ${alreadyGen.length} 位生成过报告。`,
            '确认生成方式',
            { confirmButtonText: '全部重新生成', cancelButtonText: '只生成未生成的', type: 'warning', distinguishCancelAndClose: true }
          )
          // confirm → 全部重新生成
        } catch (e) {
          action = e
          if (action === 'cancel') {
            const remaining = filtered.filter(t => t.pdfFlag !== 1)
            if (remaining.length === 0) {
              this.$message.info('没有需要新生成的报告')
              return
            }
            return this._doBatchGenerate(remaining, false)
          }
          return // close ↔ 不做
        }
      }
      return this._doBatchGenerate(filtered, true)
    },

    /**
     * 真正执行批量生成
     * A: 并发=2（后端 chromedp pool size）
     * C: 实时进度
     * D: 失败重试 1 次
     */
    async _doBatchGenerate(toGenerate, forceRegen = false){
      const isMbti = toGenerate.every(t => t.repoCode && t.repoCode.startsWith('003'))
      const apiUrl = isMbti
        ? '/exam/api/mbti/generate-report'
        : '/exam/api/exam/exam/generate-report'
      const total = toGenerate.length
      const concurrency = 2

      const downloadLoadingInstance = Loading.service({
        text: `正在生成测评报告（0/${total}），请稍候`,
        spinner: 'el-icon-loading',
        background: 'rgba(0, 0, 0, 0.7)'
      })

      let done = 0
      let failed = 0
      let i = 0

      // D: 单个任务含重试
      // MBTI: 完整版 + 简版各一次（顺序，简版一般秒级；后端 generate-report 简版有缓存逻辑，force=true 跳过缓存）
      const genOne = async (tester) => {
        for (let r = 0; r <= 1; r++) {
          try {
            await request.post(apiUrl, { paperId: tester.paperId })
            // MBTI 同时刷新简版（forceRegen=true 时后端会删旧重生）
            if (isMbti) {
              try {
                const body = forceRegen
                  ? { paperId: tester.paperId, type: 'simple', force: true }
                  : { paperId: tester.paperId, type: 'simple' }
                await request.post(apiUrl, body)
              } catch (e2) {
                console.warn('simple regenerate failed for', tester.name, e2 && e2.message)
              }
            }
            return true
          } catch (err) {
            if (r === 1) {
              console.error('batch generate failed for', tester.name, err)
              return false
            }
            await new Promise(resolve => setTimeout(resolve, 1000))
          }
        }
        return false
      }

      // A: worker pool
      const worker = async () => {
        while (i < total) {
          const idx = i++
          const ok = await genOne(toGenerate[idx])
          if (ok) done++
          else failed++
          downloadLoadingInstance.setText(
            `正在生成测评报告（${done + failed}/${total}），请稍候`
          )
        }
      }
      await Promise.all(Array.from({ length: concurrency }, () => worker()))

      downloadLoadingInstance.close()
      if (failed === 0) {
        this.$message.success(`已完成 ${done} 份报告生成`)
      } else {
        this.$message.warning(`完成 ${done} 份，失败 ${failed} 份`)
      }
      this.getList(this.isOpen)
    },
    handleDelete(row) {
      const ids = row.id || this.ids;
      console.log("this.isOpen", this.isOpen)
      // V-008 / D-016 修复 (2026-05-11): 此处实际是软删（del_flag=1，数据保留），
      // 修改提示文案让用户知道是"停用"而非"彻底删除"
      const tip = '此操作将停用所选考生（数据保留可恢复，不会清空答卷与报告记录），是否继续？'
      if (this.isOpen === 1 || this.isOpen === "1") {
        this.$modal.confirm(tip).then(function() {
          return logisticDelCandidate(ids)
        }).then(() => {
          this.getList(this.isOpen);
          this.$modal.msgSuccess("已停用");
        }).catch(() => {});
      } else if (this.isOpen === 2 || this.isOpen === "2"){
        this.$modal.confirm(tip).then(function() {
          return logisticDelTester(ids)
        }).then(() => {
          this.getList(this.isOpen);
          this.$modal.msgSuccess("已停用");
        }).catch(() => {});
      }
    },

    handleBatchDownload() {
      for (const tester of this.selections) {
        if (tester.pdfFlag === 0) {
          Message.error('存在未生成报告，请先生成报告');
          return;
        }
      }

      let downloadLoadingInstance = Loading.service({ text: "正在下载测评报告，请稍候", spinner: "el-icon-loading", background: "rgba(0, 0, 0, 0.7)", })
      const ids = this.ids;
      let data = JSON.stringify(ids)
      if (this.isOpen === 1 || this.isOpen === "1") {
        batchDownloadCandidatePdf(data).then(async (response) => {
          console.log(response.headers)
          let filename = this.testTitle + "_" + this.testerList[0].examId
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
          downloadLoadingInstance.close();
        }).catch((r) => {
          console.error(r)
          Message.error('下载文件出现错误，请联系管理员！')
          downloadLoadingInstance.close();
        })
      } else if (this.isOpen === 2 || this.isOpen === "2"){
        batchDownloadTesterPdf(data).then(async (response) => {
          console.log(response.headers)
          let filename = this.testTitle + "_" + this.testerList[0].examId
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
          downloadLoadingInstance.close();
        }).catch((r) => {
          console.error(r)
          Message.error('下载文件出现错误，请联系管理员！')
          downloadLoadingInstance.close();
        })
      }
    },
    async getPdfByPath(fileFormData) {
      // 可以对这个返回值 res 做些数据处理，再返回
      console.log("fileFormData,", fileFormData)
      return await pdfDownload(fileFormData)
    },

    // 批量下载 MBTI 简版（zip）
    // 流程：1) 校验所选都是 MBTI 完成态  2) 逐个 generate-report type=simple（已有则秒返）
    //       3) 调 batch-download-simple 取 zip
    async handleBatchDownloadSimple() {
      const mbtiSelections = this.selections.filter(t => t.repoCode && t.repoCode.startsWith('003'))
      if (mbtiSelections.length === 0) {
        Message.warning('请勾选 MBTI 测评的考生')
        return
      }
      if (mbtiSelections.length < this.selections.length) {
        Message.warning('已自动忽略 ' + (this.selections.length - mbtiSelections.length) + ' 个非 MBTI 考生')
      }
      const noPaper = mbtiSelections.filter(t => !t.paperId || !t.endTime)
      if (noPaper.length > 0) {
        Message.error('存在未完成测评的考生，请先完成')
        return
      }

      const loading = Loading.service({
        text: '正在生成并打包简版报告（' + mbtiSelections.length + ' 份），请稍候',
        spinner: 'el-icon-loading',
        background: 'rgba(0, 0, 0, 0.7)',
      })
      try {
        // 1) 串行生成（避免 LibreOffice 并发崩溃）
        let okCount = 0
        for (const t of mbtiSelections) {
          try {
            const r = await request.post('/exam/api/mbti/generate-report', { paperId: t.paperId, type: 'simple' })
            if (r && r.success !== false) okCount++
          } catch (err) {
            console.warn('generate simple failed for', t.name, err)
          }
        }
        if (okCount === 0) {
          Message.error('简版报告全部生成失败，请检查模板')
          return
        }
        // 2) 批量打包下载
        const ids = mbtiSelections.map(t => t.paperId)
        const resp = await request.post(
          '/exam/api/mbti/batch-download-simple',
          ids,
          { responseType: 'blob', timeout: 600000 }
        )
        const url = window.URL.createObjectURL(resp)
        const link = document.createElement('a')
        link.style.display = 'none'
        link.href = url
        const fname = (this.testTitle || 'mbti') + '_简版报告_' + (this.testerList[0] && this.testerList[0].examId || '') + '.zip'
        link.setAttribute('download', fname)
        document.body.appendChild(link)
        link.click()
        URL.revokeObjectURL(link.href)
        document.body.removeChild(link)
        Message.success('成功下载 ' + okCount + ' 份简版报告')
      } catch (err) {
        console.error('batch simple download error:', err)
        Message.error('批量下载失败：' + (err.message || '未知错误'))
      } finally {
        loading.close()
      }
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
      this.getList(this.isOpen);
    },
    /** 重置按钮操作 */
    resetQuery() {
      this.resetForm("queryForm");
      this.handleQuery();
    },

    viewTeamPdf() {
      this.visibleTeam = true
    },

    // 下载图队报告
    downloadTeamPdf() {
      let downloadLoadingInstance = Loading.service({ text: "正在下载团队报告，请稍候", spinner: "el-icon-loading", background: "rgba(0, 0, 0, 0.7)", })
      let formData = new FormData();
      formData.append("examId", this.testerList[0].examId);
      pdfTeamDownload(formData).then(async (response) => {
        console.log(response.headers)
        let filename = this.testTitle + "_" + this.testerList[0].examId
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
        downloadLoadingInstance.close();
      }).catch((r) => {
        console.error(r)
        Message.error('下载文件出现错误，请联系管理员！')
        downloadLoadingInstance.close();
      })
    },

    // 查看报告操作
    viewPdf(row) {
      // MBTI 报告：如果已生成 PDF，直接在新窗口打开；否则先生成再打开
      if (row.repoCode && row.repoCode.startsWith('003')) {
        if (row.pdfFlag === 1 && row.pdfPath) {
          // 已有报告，直接下载并在新窗口预览
          this.openPdfPreview(row.pdfPath)
        } else {
          // 未生成，先调 API 生成
          let loadingInstance = Loading.service({ text: "正在生成报告，请稍候", spinner: "el-icon-loading", background: "rgba(0, 0, 0, 0.7)" })
          request.post('/exam/api/mbti/generate-report', { paperId: row.paperId }).then(res => {
            loadingInstance.close()
            if (res.data && res.data.path) {
              this.getList(this.isOpen)
              // 用返回的路径打开预览
              this.openPdfPreview(res.data.path)
            }
          }).catch(err => {
            loadingInstance.close()
            console.error('generate report error:', err)
            Message.error('生成报告失败')
          })
        }
        return
      }
      // 非 MBTI 报告：保持原有对话框模式
      this.$emit('singlePdfFinished');
      const paperId = row.paperId
      this.repoCode = row.repoCode
      this.dialogVisible = true;
      this.$nextTick(() => {
        this.$refs.testResult.testerId = row.idNumber ? row.idNumber : row.id
        this.$refs.testResult.pdfDownload();
        this.$refs.testResult.fetchTester(paperId);
        this.$refs.testResult.fetchScore(paperId);
      })
    },

    // 打开 PDF 预览（通过 blob URL 在新窗口展示）
    openPdfPreview(pdfPath) {
      let formData = new FormData();
      formData.append('file', pdfPath);
      pdfDownload(formData).then(response => {
        const blob = new Blob([response], { type: 'application/pdf' })
        const url = window.URL.createObjectURL(blob)
        window.open(url, '_blank')
      }).catch(err => {
        console.error('open pdf error:', err)
        Message.error('打开报告失败')
      })
    },

    // MBTI 报告下载命令（完整版/简版）
    async handleReportCommand(cmd, row) {
      if (cmd === 'full') {
        this.downloadPdf(row.pdfPath)
      } else if (cmd === 'simple') {
        // 先尝试生成简版（如果未生成）
        // FB-010: 简版可能因模板缺失而无法生成，给出明确错误提示
        let downloadLoadingInstance = Loading.service({ text: "正在准备简版报告，请稍候", spinner: "el-icon-loading", background: "rgba(0, 0, 0, 0.7)" })
        try {
          const genResp = await request.post('/exam/api/mbti/generate-report', { paperId: row.paperId, type: 'simple' })
          // 检查 generate 是否成功（success: false 时后端 message 含具体原因）
          if (genResp && genResp.success === false) {
            Message.error(genResp.msg || '简版报告生成失败')
            return
          }
          // 下载简版
          const resp = await request.post('/exam/api/mbti/download-report', { paperId: row.paperId, type: 'simple' }, { responseType: 'blob' })
          const url = window.URL.createObjectURL(resp)
          const link = document.createElement('a')
          link.style.display = 'none'
          link.href = url
          link.setAttribute('download', row.name + '_简版报告.pdf')
          document.body.appendChild(link)
          link.click()
          URL.revokeObjectURL(link.href)
          document.body.removeChild(link)
        } catch (err) {
          console.error('simple report error:', err)
          // 区分 404（模板缺失）和其他错误
          if (err.response && err.response.status === 404) {
            Message.error('简版报告模板未配置，请联系管理员')
          } else {
            Message.error('简版报告下载失败：' + (err.message || '未知错误'))
          }
        } finally {
          downloadLoadingInstance.close()
        }
      }
    },

    // 下载报告操作
    async downloadPdf(pdfPath) {
      console.log("=====================", pdfPath)
      if (pdfPath === undefined || pdfPath === null || pdfPath === "") {
        Message.error('文件未保存，请先生成报告！');
        return;
      }
      let downloadLoadingInstance = Loading.service({ text: "正在下载测评报告，请稍候", spinner: "el-icon-loading", background: "rgba(0, 0, 0, 0.7)", })
      let file = pdfPath;
      let formData = new FormData();
      formData.append("file", file);
      await pdfDownload(formData).then(async (response) => {
        // console.log(response.headers)
        let arr = pdfPath.replace(/\\/g, '/').split("/");
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
        downloadLoadingInstance.close();
      }).catch((r) => {
        console.error(r)
        Message.error('下载文件出现错误，请联系管理员！')
        downloadLoadingInstance.close();
      })
    },

  }
};
</script>

<style>
.el-loading-mask{
  z-index: 9999!important;
}

</style>

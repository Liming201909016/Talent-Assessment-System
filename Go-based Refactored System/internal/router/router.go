package router

import (
	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/handler"
	"github.com/talent-assessment/refactored/internal/middleware"
	"github.com/talent-assessment/refactored/internal/repository"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
)

func Setup(cfg *config.Config, db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.CORS())

	// 依赖装配
	userRepo := repository.NewSysUserRepo(db)
	menuRepo := repository.NewSysMenuRepo(db)
	authSvc := service.NewAuthService(cfg, userRepo, menuRepo)

	authH := handler.NewAuthHandler(authSvc)
	dictH := handler.NewDictHandler(db)
	repoH := handler.NewRepoHandler(db)
	quH := handler.NewQuHandler(db)
	examH := handler.NewExamHandler(db, cfg)
	testerH := handler.NewTesterHandler(db, cfg)
	userExamH := handler.NewUserExamHandler(db)
	paperH := handler.NewPaperHandler(db)
	candidateH := handler.NewCandidateHandler(db, cfg)
	mbtiH := handler.NewMbtiHandler(db)
	departH := handler.NewSysDepartHandler(db)
	sysRoleH := handler.NewSysRoleHandler(db)
	sysConfigH := handler.NewSysConfigHandler(db)
	sysUserH := handler.NewSysExamUserHandler(db)
	ruoyiSysH := handler.NewRuoYiSystemHandler(db)

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// 应用 JWT 中间件（内部按路径豁免匿名）
	r.Use(middleware.JWT(cfg, authSvc))

	// ============ 认证 / 核心（RuoYi 原生风格） ============
	r.GET("/captchaImage", authH.CaptchaImage)
	r.POST("/login", authH.Login)
	r.POST("/logout", authH.Logout)
	r.GET("/getInfo", authH.GetInfo)
	r.GET("/getRouters", authH.GetRouters)

	// 字典
	r.GET("/system/dict/data/type/:dictType", dictH.DataByType)

	// system.* 占位 — RuoYi admin framework CRUD
	sys := r.Group("/system")
	{
		// user — 对接 RuoYi 前端 system/user 页面
		sys.GET("/user/list", sysUserH.RuoYiUserList)
		sys.GET("/user/", ruoyiSysH.UserDetail)
		sys.GET("/user/:userId", ruoyiSysH.UserDetail)
		sys.POST("/user", ruoyiSysH.UserAdd)
		sys.PUT("/user", ruoyiSysH.UserEdit)
		sys.DELETE("/user/:ids", ruoyiSysH.UserDelete)
		sys.PUT("/user/changeStatus", ruoyiSysH.UserChangeStatus)
		sys.PUT("/user/resetPwd", ruoyiSysH.UserResetPwd)
		sys.GET("/user/profile", ruoyiSysH.UserProfile)
		sys.PUT("/user/profile", handler.AjaxStub("system/user/profile/update"))
		sys.PUT("/user/profile/updatePwd", handler.AjaxStub("system/user/profile/updatePwd"))
		sys.POST("/user/profile/avatar", handler.AjaxStub("system/user/profile/avatar"))
		sys.GET("/user/authRole/:userId", handler.AjaxStub("system/user/authRole"))
		sys.PUT("/user/authRole", handler.AjaxStub("system/user/authRole/update"))

		// role
		sys.GET("/role/list", ruoyiSysH.RoleList)
		sys.GET("/role/:roleId", ruoyiSysH.RoleDetail)
		sys.POST("/role", ruoyiSysH.RoleAdd)
		sys.PUT("/role", ruoyiSysH.RoleEdit)
		sys.DELETE("/role/:roleIds", ruoyiSysH.RoleDelete)
		sys.PUT("/role/changeStatus", handler.AjaxStub("system/role/changeStatus"))
		sys.PUT("/role/dataScope", handler.AjaxStub("system/role/dataScope"))
		sys.GET("/role/authUser/allocatedList", handler.TableStub("system/role/authUser/allocatedList"))
		sys.GET("/role/authUser/unallocatedList", handler.TableStub("system/role/authUser/unallocatedList"))
		sys.PUT("/role/authUser/cancel", handler.AjaxStub("system/role/authUser/cancel"))
		sys.PUT("/role/authUser/cancelAll", handler.AjaxStub("system/role/authUser/cancelAll"))
		sys.PUT("/role/authUser/selectAll", handler.AjaxStub("system/role/authUser/selectAll"))

		// menu
		sys.GET("/menu/list", ruoyiSysH.MenuList)
		sys.GET("/menu/:menuId", ruoyiSysH.MenuDetail)
		sys.POST("/menu", ruoyiSysH.MenuAdd)
		sys.PUT("/menu", ruoyiSysH.MenuEdit)
		sys.DELETE("/menu/:menuId", ruoyiSysH.MenuDelete)
		sys.GET("/menu/treeselect", ruoyiSysH.MenuTreeselect)
		sys.GET("/menu/roleMenuTreeselect/:roleId", ruoyiSysH.RoleMenuTreeselect)

		// dept
		sys.GET("/dept/list", ruoyiSysH.DeptList)
		sys.GET("/dept/list/exclude/:deptId", ruoyiSysH.DeptExcludeChild)
		sys.GET("/dept/:deptId", ruoyiSysH.DeptDetail)
		sys.POST("/dept", ruoyiSysH.DeptAdd)
		sys.PUT("/dept", ruoyiSysH.DeptEdit)
		sys.DELETE("/dept/:deptId", ruoyiSysH.DeptDelete)
		sys.GET("/dept/treeselect", ruoyiSysH.DeptTreeselect)
		sys.GET("/dept/roleDeptTreeselect/:roleId", ruoyiSysH.DeptRoleTreeselect)

		// config
		sys.GET("/config/list", ruoyiSysH.ConfigList)
		sys.GET("/config/:configId", ruoyiSysH.ConfigDetail)
		sys.POST("/config", ruoyiSysH.ConfigAdd)
		sys.PUT("/config", ruoyiSysH.ConfigEdit)
		sys.DELETE("/config/:configIds", ruoyiSysH.ConfigDelete)
		sys.DELETE("/config/refreshCache", handler.AjaxStub("system/config/refreshCache"))
		sys.GET("/config/configKey/:configKey", ruoyiSysH.ConfigByKey)

		// dict/type
		sys.GET("/dict/type/list", ruoyiSysH.DictTypeList)
		sys.GET("/dict/type/:dictId", ruoyiSysH.DictTypeDetail)
		sys.POST("/dict/type", ruoyiSysH.DictTypeAdd)
		sys.PUT("/dict/type", ruoyiSysH.DictTypeEdit)
		sys.DELETE("/dict/type/:dictIds", ruoyiSysH.DictTypeDelete)
		sys.DELETE("/dict/type/refreshCache", handler.AjaxStub("system/dict/type/refreshCache"))
		sys.GET("/dict/type/optionselect", ruoyiSysH.DictTypeOptionselect)

		// dict/data
		sys.GET("/dict/data/list", ruoyiSysH.DictDataList)
		sys.GET("/dict/data/:dictCode", ruoyiSysH.DictDataDetail)
		sys.POST("/dict/data", ruoyiSysH.DictDataAdd)
		sys.PUT("/dict/data", ruoyiSysH.DictDataEdit)
		sys.DELETE("/dict/data/:dictCodes", ruoyiSysH.DictDataDelete)

		// notice
		sys.GET("/notice/list", ruoyiSysH.NoticeList)
		sys.GET("/notice/:noticeId", ruoyiSysH.NoticeDetail)
		sys.POST("/notice", ruoyiSysH.NoticeAdd)
		sys.PUT("/notice", ruoyiSysH.NoticeEdit)
		sys.DELETE("/notice/:noticeIds", ruoyiSysH.NoticeDelete)

		// post
		sys.GET("/post/list", ruoyiSysH.PostList)
		sys.GET("/post/:postId", ruoyiSysH.PostDetail)
		sys.POST("/post", ruoyiSysH.PostAdd)
		sys.PUT("/post", ruoyiSysH.PostEdit)
		sys.DELETE("/post/:postIds", ruoyiSysH.PostDelete)
	}

	// monitor.* 占位
	monitor := r.Group("/monitor")
	{
		monitor.GET("/server", ruoyiSysH.ServerInfo)
		monitor.GET("/cache", ruoyiSysH.CacheInfo)
		monitor.GET("/online/list", ruoyiSysH.OnlineList)
		monitor.DELETE("/online/:tokenId", handler.AjaxStub("monitor/online/forceLogout"))
		monitor.GET("/job/list", ruoyiSysH.JobList)
		monitor.GET("/job/:jobId", handler.AjaxStub("monitor/job/detail"))
		monitor.POST("/job", handler.AjaxStub("monitor/job/add"))
		monitor.PUT("/job", handler.AjaxStub("monitor/job/edit"))
		monitor.DELETE("/job/:jobIds", handler.AjaxStub("monitor/job/delete"))
		monitor.PUT("/job/changeStatus", handler.AjaxStub("monitor/job/changeStatus"))
		monitor.PUT("/job/run", handler.AjaxStub("monitor/job/run"))
		monitor.GET("/jobLog/list", handler.TableStub("monitor/jobLog/list"))
		monitor.DELETE("/jobLog/:jobLogIds", handler.AjaxStub("monitor/jobLog/delete"))
		monitor.DELETE("/jobLog/clean", handler.AjaxStub("monitor/jobLog/clean"))
		monitor.GET("/operlog/list", ruoyiSysH.OperlogList)
		monitor.DELETE("/operlog/:operIds", handler.AjaxStub("monitor/operlog/delete"))
		monitor.DELETE("/operlog/clean", handler.AjaxStub("monitor/operlog/clean"))
		monitor.GET("/logininfor/list", ruoyiSysH.LogininforList)
		monitor.DELETE("/logininfor/:infoIds", handler.AjaxStub("monitor/logininfor/delete"))
		monitor.DELETE("/logininfor/clean", handler.AjaxStub("monitor/logininfor/clean"))
	}

	// tool.* 占位
	tool := r.Group("/tool")
	{
		tool.GET("/gen/list", handler.TableStub("tool/gen/list"))
		tool.GET("/gen/:tableId", handler.AjaxStub("tool/gen/detail"))
		tool.GET("/gen/preview/:tableId", handler.AjaxStub("tool/gen/preview"))
		tool.GET("/gen/db/list", handler.TableStub("tool/gen/db/list"))
		tool.POST("/gen/importTable", handler.AjaxStub("tool/gen/importTable"))
		tool.PUT("/gen", handler.AjaxStub("tool/gen/edit"))
		tool.DELETE("/gen/:tableIds", handler.AjaxStub("tool/gen/delete"))
		tool.GET("/gen/genCode/:tableName", handler.AjaxStub("tool/gen/genCode"))
		tool.GET("/gen/synchDb/:tableName", handler.AjaxStub("tool/gen/synchDb"))
	}

	// register (open)
	r.POST("/register", handler.AjaxStub("register"))

	// ============ /exam/api/* 业务模块（与前端 api/*.js 前缀一致） ============
	api := r.Group("/exam/api")

	// 题库 repo
	repoGrp := api.Group("/repo")
	{
		repoGrp.POST("/paging", repoH.Paging)
		repoGrp.POST("/list", repoH.List)
		repoGrp.GET("/detail", repoH.Detail)
		repoGrp.POST("/detail", repoH.Detail)
		repoGrp.POST("/save", repoH.Save)
		repoGrp.POST("/delete", repoH.Remove)
		repoGrp.POST("/batch-action", repoH.BatchAction)
	}

	// 题目 qu
	quGrp := api.Group("/qu/qu")
	{
		quGrp.POST("/paging", quH.Paging)
		quGrp.POST("/detail", quH.Detail)
		quGrp.POST("/save", quH.Save)
		quGrp.POST("/delete", quH.Delete)
		quGrp.POST("/list", quH.List)
		quGrp.POST("/import-excel", quH.ImportExcel)
		quGrp.POST("/import", quH.ImportExcel)
		quGrp.POST("/export", quH.Export)
		quGrp.POST("/import/template", quH.ImportTemplate)
	}

	// 考试 exam
	examGrp := api.Group("/exam/exam")
	{
		examGrp.POST("/paging", examH.Paging)
		examGrp.POST("/online-paging", examH.OnlinePaging)
		examGrp.POST("/detail", examH.Detail)
		examGrp.POST("/save", examH.Save)
		examGrp.POST("/delete", examH.Delete)
		examGrp.POST("/state", examH.State)
		examGrp.POST("/pdf-team", examH.PdfTeam)
		examGrp.POST("/pdf-upload", examH.PdfUpload)
		examGrp.POST("/export-raw-data", examH.ExportRawData)
		examGrp.GET("/export-raw-data", examH.ExportRawData)
		examGrp.POST("/review-paging", examH.ReviewPaging)
	}

	// tester
	testerGrp := api.Group("/tester")
	{
		testerGrp.POST("/login", testerH.LoginForm)
		testerGrp.GET("", testerH.List)
		testerGrp.GET("/list", testerH.List)
		testerGrp.GET("/:id", testerH.Detail)
		testerGrp.GET("/idNumber/:idNumber", testerH.DetailByIDNumber)
		testerGrp.POST("", testerH.Create)
		testerGrp.PUT("", testerH.Update)
		testerGrp.DELETE("/:ids", testerH.Remove)
		testerGrp.DELETE("/logistic/:ids", testerH.Logistic)
		testerGrp.POST("/end-time", testerH.EndTime)
		testerGrp.POST("/stand-score", testerH.StandScore)
		testerGrp.POST("/batch-download", testerH.BatchDownload)
		testerGrp.POST("/pdf-persistence", testerH.PdfPersistence)
		testerGrp.POST("/importData", testerH.ImportData)
		testerGrp.POST("/importTemplate", testerH.ImportTemplate)
		testerGrp.POST("/export", testerH.Export)
		testerGrp.GET("/tester-list", testerH.TesterList)
		testerGrp.GET("/team-score", testerH.TeamScore)
	}

	// user.exam（我的考试）
	ueGrp := api.Group("/user/exam")
	{
		ueGrp.POST("/paging", userExamH.Paging)
		ueGrp.POST("/my-paging", userExamH.MyPaging)
	}

	// paper
	paperGrp := api.Group("/paper/paper")
	{
		paperGrp.POST("/paging", paperH.Paging)
		paperGrp.POST("/detail", paperH.Detail)
		paperGrp.POST("/save", paperH.Save)
		paperGrp.POST("/delete", paperH.Delete)
		paperGrp.POST("/create-paper", paperH.CreatePaper)
		paperGrp.POST("/paper-detail", paperH.PaperDetail)
		paperGrp.POST("/paperQu-detail", paperH.PaperQuDetail)
		paperGrp.POST("/qu-detail", paperH.QuDetail)
		paperGrp.POST("/fill-answer", paperH.FillAnswer)
		paperGrp.POST("/hand-exam", paperH.HandExam)
		paperGrp.POST("/paper-result", paperH.PaperResult)
		paperGrp.POST("/training", paperH.Training)
		paperGrp.POST("/show_pdf", paperH.ShowPdf)
		paperGrp.POST("/review-paper", paperH.ReviewPaper)
		paperGrp.POST("/stand-score", paperH.PaperStandScore)
	}

	// candidate（开放测评）
	candidateGrp := api.Group("/candidate")
	{
		candidateGrp.PUT("", candidateH.Update)
		candidateGrp.DELETE("/:ids", candidateH.Remove)
		candidateGrp.DELETE("/logistic/:ids", candidateH.Logistic)
		candidateGrp.DELETE("/logicDeletePdfByIds/:ids", candidateH.LogicDeletePdfByIds)
		candidateGrp.PUT("/logicDeletePdfByIds/:ids", candidateH.LogicDeletePdfByIds)
		candidateGrp.POST("/save", candidateH.Save)
		candidateGrp.POST("/update", candidateH.Update)
		candidateGrp.POST("/info", candidateH.Info)
		candidateGrp.POST("/tester-info", candidateH.TesterInfo)
		candidateGrp.POST("/tester-list", candidateH.TesterList)
		candidateGrp.POST("/stand-score", candidateH.StandScoreCandidate)
		candidateGrp.GET("/team-score", candidateH.TeamScore)
		candidateGrp.POST("/end-time", candidateH.EndTimeCandidate)
		candidateGrp.POST("/pdf-persistence", candidateH.PdfPersistence)
		candidateGrp.POST("/pdf-upload", candidateH.PdfUpload)
		candidateGrp.POST("/batch-download", candidateH.BatchDownload)
	}

	// MBTI 职业性格测验
	mbtiGrp := api.Group("/mbti")
	{
		mbtiGrp.POST("/paper-detail", mbtiH.PaperDetail)
		mbtiGrp.POST("/fill-answer", mbtiH.FillAnswer)
		mbtiGrp.POST("/submit", mbtiH.Submit)
		mbtiGrp.POST("/score", mbtiH.Score)

		// MBTI 报告生成（docx 模板替换）
		mbtiReportH := handler.NewMbtiReportHandler(db, cfg.Upload.MbtiTemplates, cfg.Upload.Path)
		mbtiH.SetReportHandler(mbtiReportH) // 注入报告生成器，Submit 后自动生成 PDF
		mbtiGrp.POST("/generate-report", mbtiReportH.GenerateReport)
		mbtiGrp.POST("/download-report", mbtiReportH.DownloadReport)

		// 模板管理
		mbtiGrp.GET("/templates", mbtiReportH.ListTemplates)
		mbtiGrp.GET("/templates/download/:type", mbtiReportH.DownloadTemplate)
		mbtiGrp.POST("/templates/upload", mbtiReportH.UploadTemplate)
	}

	// sys/depart
	departGrp := api.Group("/sys/depart")
	{
		departGrp.POST("/paging", departH.Paging)
		departGrp.POST("/list", departH.List)
		departGrp.GET("/list", departH.List)
		departGrp.POST("/detail", departH.Detail)
		departGrp.POST("/save", departH.Save)
		departGrp.POST("/delete", departH.Delete)
		departGrp.POST("/tree", departH.Tree)
		departGrp.GET("/tree", departH.Tree)
		departGrp.POST("/sort", departH.Sort)
	}

	// sys/role
	sysRoleGrp := api.Group("/sys/role")
	{
		sysRoleGrp.POST("/paging", sysRoleH.Paging)
		sysRoleGrp.POST("/list", sysRoleH.List)
		sysRoleGrp.GET("/list", sysRoleH.List)
	}

	// sys/config
	sysConfigGrp := api.Group("/sys/config")
	{
		sysConfigGrp.POST("/detail", sysConfigH.Detail)
		sysConfigGrp.GET("/detail", sysConfigH.Detail)
		sysConfigGrp.POST("/save", sysConfigH.Save)
	}

	// sys/user（exam 模块用户）
	sysUserGrp := api.Group("/sys/user")
	{
		sysUserGrp.POST("/paging", sysUserH.Paging)
		sysUserGrp.POST("/list", sysUserH.List)
		sysUserGrp.GET("/list", sysUserH.List)
		sysUserGrp.POST("/detail", sysUserH.Detail)
		sysUserGrp.POST("/save", sysUserH.Save)
		sysUserGrp.POST("/delete", sysUserH.Delete)
		sysUserGrp.POST("/state", sysUserH.State)
		sysUserGrp.POST("/reset-pwd", sysUserH.ResetPwd)
		sysUserGrp.POST("/update", sysUserH.Update)
		sysUserGrp.POST("/reg", sysUserH.Reg)
		sysUserGrp.POST("/quick-reg", sysUserH.QuickReg)
	}

	// user/repo 和 user/wrong-book Java 后端无对应实现，保留空 stub
	stubGroup(api, "/user/repo", []string{"paging", "list", "detail", "save", "delete"})
	stubGroup(api, "/user/wrong-book", []string{"paging", "list", "detail", "delete"})

	return r
}

func stubGroup(g *gin.RouterGroup, prefix string, actions []string) {
	sub := g.Group(prefix)
	for _, a := range actions {
		sub.POST("/"+a, handler.Stub(prefix+"/"+a))
		sub.GET("/"+a, handler.Stub(prefix+"/"+a))
	}
}

package model

import "time"

// Qu 题目 el_qu
type Qu struct {
	ID         string     `gorm:"column:id;primaryKey" json:"id"`
	QuType     int        `gorm:"column:qu_type"       json:"quType"`
	Level      int        `gorm:"column:level"         json:"level"`
	Image      string     `gorm:"column:image"         json:"image"`
	Content    string     `gorm:"column:content"       json:"content"`
	CreateTime *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time"   json:"updateTime"`
	Remark     string     `gorm:"column:remark"        json:"remark"`
	Analysis   string     `gorm:"column:analysis"      json:"analysis"`
	Title      string     `gorm:"column:title"         json:"title"`
}

func (Qu) TableName() string { return "el_qu" }

// QuAnswer 备选答案 el_qu_answer
type QuAnswer struct {
	ID       string `gorm:"column:id;primaryKey" json:"id"`
	QuID     string `gorm:"column:qu_id"         json:"quId"`
	IsRight  int8   `gorm:"column:is_right"      json:"isRight"`
	Image    string `gorm:"column:image"         json:"image"`
	Content  string `gorm:"column:content"       json:"content"`
	Analysis string `gorm:"column:analysis"      json:"analysis"`
	Score    int    `gorm:"column:score"         json:"score"`
}

func (QuAnswer) TableName() string { return "el_qu_answer" }

// QuRepo 试题-题库关联 el_qu_repo
type QuRepo struct {
	ID     string `gorm:"column:id;primaryKey" json:"id"`
	QuID   string `gorm:"column:qu_id"         json:"quId"`
	RepoID string `gorm:"column:repo_id"       json:"repoId"`
	QuType int    `gorm:"column:qu_type"       json:"quType"`
	Sort   int    `gorm:"column:sort"          json:"sort"`
}

func (QuRepo) TableName() string { return "el_qu_repo" }

// Repo 题库 el_repo
type Repo struct {
	ID         string     `gorm:"column:id;primaryKey" json:"id"`
	Code       string     `gorm:"column:code"          json:"code"`
	Title      string     `gorm:"column:title"         json:"title"`
	RadioCount int        `gorm:"column:radio_count"   json:"radioCount"`
	MultiCount int        `gorm:"column:multi_count"   json:"multiCount"`
	JudgeCount int        `gorm:"column:judge_count"   json:"judgeCount"`
	Remark     string     `gorm:"column:remark"        json:"remark"`
	CreateTime *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time"   json:"updateTime"`
}

func (Repo) TableName() string { return "el_repo" }

// Exam 考试 el_exam
type Exam struct {
	ID           string     `gorm:"column:id;primaryKey" json:"id"`
	Title        string     `gorm:"column:title"         json:"title"`
	Content      string     `gorm:"column:content"       json:"content"`
	OpenType     int        `gorm:"column:open_type"     json:"openType"`
	JoinType     int        `gorm:"column:join_type"     json:"joinType"`
	IsOpen       int        `gorm:"column:is_open"       json:"isOpen"`
	AnswerType   int        `gorm:"column:answer_type"   json:"answerType"`
	Level        int        `gorm:"column:level"         json:"level"`
	State        int        `gorm:"column:state"         json:"state"`
	TimeLimit    int8       `gorm:"column:time_limit"    json:"timeLimit"`
	ShowPdf      int8       `gorm:"column:show_pdf"      json:"showPdf"`
	StartTime    *time.Time `gorm:"column:start_time"    json:"startTime"`
	EndTime      *time.Time `gorm:"column:end_time"      json:"endTime"`
	CreateTime   *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime   *time.Time `gorm:"column:update_time"   json:"updateTime"`
	TotalScore   int        `gorm:"column:total_score"   json:"totalScore"`
	TotalTime    int        `gorm:"column:total_time"    json:"totalTime"`
	QualifyScore int        `gorm:"column:qualify_score" json:"qualifyScore"`
	PdfPath        string     `gorm:"column:pdf_path"         json:"pdfPath"`
	RequiredFields string     `gorm:"column:required_fields"  json:"requiredFields"`
	StuFlag        int8       `gorm:"column:stu_flag"         json:"stuFlag"`
}

func (Exam) TableName() string { return "el_exam" }

// ExamRepo el_exam_repo
type ExamRepo struct {
	ID         string `gorm:"column:id;primaryKey" json:"id"`
	ExamID     string `gorm:"column:exam_id"       json:"examId"`
	RepoID     string `gorm:"column:repo_id"       json:"repoId"`
	RadioCount int    `gorm:"column:radio_count"   json:"radioCount"`
	RadioScore int    `gorm:"column:radio_score"   json:"radioScore"`
	MultiCount int    `gorm:"column:multi_count"   json:"multiCount"`
	MultiScore int    `gorm:"column:multi_score"   json:"multiScore"`
	JudgeCount int    `gorm:"column:judge_count"   json:"judgeCount"`
	JudgeScore int    `gorm:"column:judge_score"   json:"judgeScore"`
	SaqCount   int    `gorm:"column:saq_count"     json:"saqCount"`
	SaqScore   int    `gorm:"column:saq_score"     json:"saqScore"`
}

func (ExamRepo) TableName() string { return "el_exam_repo" }

// ExamDepart el_exam_depart
type ExamDepart struct {
	ID       string `gorm:"column:id;primaryKey" json:"id"`
	ExamID   string `gorm:"column:exam_id"       json:"examId"`
	DepartID string `gorm:"column:depart_id"     json:"departId"`
}

func (ExamDepart) TableName() string { return "el_exam_depart" }

// Paper el_paper
type Paper struct {
	ID           string     `gorm:"column:id;primaryKey" json:"id"`
	UserID       string     `gorm:"column:user_id"       json:"userId"`
	DepartID     string     `gorm:"column:depart_id"     json:"departId"`
	ExamID       string     `gorm:"column:exam_id"       json:"examId"`
	Title        string     `gorm:"column:title"         json:"title"`
	TotalTime    int        `gorm:"column:total_time"    json:"totalTime"`
	UserTime     int        `gorm:"column:user_time"     json:"userTime"`
	TotalScore   int        `gorm:"column:total_score"   json:"totalScore"`
	QualifyScore int        `gorm:"column:qualify_score" json:"qualifyScore"`
	ObjScore     int        `gorm:"column:obj_score"     json:"objScore"`
	SubjScore    int        `gorm:"column:subj_score"    json:"subjScore"`
	UserScore    int        `gorm:"column:user_score"    json:"userScore"`
	HasSaq       int8       `gorm:"column:has_saq"       json:"hasSaq"`
	State        int        `gorm:"column:state"         json:"state"`
	CreateTime   *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime   *time.Time `gorm:"column:update_time"   json:"updateTime"`
	LimitTime    *time.Time `gorm:"column:limit_time"    json:"limitTime"`
}

func (Paper) TableName() string { return "el_paper" }

// PaperQu el_paper_qu
type PaperQu struct {
	ID          string `gorm:"column:id;primaryKey" json:"id"`
	PaperID     string `gorm:"column:paper_id"      json:"paperId"`
	QuID        string `gorm:"column:qu_id"         json:"quId"`
	QuType      int    `gorm:"column:qu_type"       json:"quType"`
	Answered    int8   `gorm:"column:answered"      json:"answered"`
	Answer      string `gorm:"column:answer"        json:"answer"`
	Sort        int    `gorm:"column:sort"          json:"sort"`
	Score       int    `gorm:"column:score"         json:"score"`
	ActualScore int    `gorm:"column:actual_score"  json:"actualScore"`
	IsRight     int8   `gorm:"column:is_right"      json:"isRight"`
}

func (PaperQu) TableName() string { return "el_paper_qu" }

// PaperQuAnswer el_paper_qu_answer
type PaperQuAnswer struct {
	ID       string `gorm:"column:id;primaryKey" json:"id"`
	PaperID  string `gorm:"column:paper_id"      json:"paperId"`
	QuID     string `gorm:"column:qu_id"         json:"quId"`
	AnswerID string `gorm:"column:answer_id"     json:"answerId"`
	Checked  int8   `gorm:"column:checked"       json:"checked"`
	Sort     int    `gorm:"column:sort"          json:"sort"`
	Abc      string `gorm:"column:abc"           json:"abc"`
	IsRight  int8   `gorm:"column:is_right"      json:"isRight"`
	Score    int    `gorm:"column:score"         json:"score"`
}

func (PaperQuAnswer) TableName() string { return "el_paper_qu_answer" }

// Tester el_tester
type Tester struct {
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	PaperID     *string    `gorm:"column:paper_id"      json:"paperId"`
	ExamID      *string    `gorm:"column:exam_id"       json:"examId"`
	IDNumber    string     `gorm:"column:id_number"     json:"idNumber"`
	Name        string     `gorm:"column:name"          json:"name"`
	Age         *int       `gorm:"column:age"           json:"age"`
	Gender      *string    `gorm:"column:gender"        json:"gender"`
	Password    string     `gorm:"column:password"      json:"password"`
	Telephone   *string    `gorm:"column:telephone"     json:"telephone"`
	Affiliation *string    `gorm:"column:affiliation"   json:"affiliation"`
	Depart      *string    `gorm:"column:depart"        json:"depart"`
	Post        *string    `gorm:"column:post"          json:"post"`
	Degree      *string    `gorm:"column:degree"        json:"degree"`
	Major       *string    `gorm:"column:major"         json:"major"`
	StuFlag     *int       `gorm:"column:stu_flag"      json:"stuFlag"`
	Status      *string    `gorm:"column:status"        json:"status"`
	CreateTime  *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime  *time.Time `gorm:"column:update_time"   json:"updateTime"`
	DelFlag     *int       `gorm:"column:del_flag"      json:"delFlag"`
	EndTime     *time.Time `gorm:"column:end_time"      json:"endTime"`
	PdfPath     *string    `gorm:"column:pdf_path"      json:"pdfPath"`
	PdfFlag     *int       `gorm:"column:pdf_flag"      json:"pdfFlag"`
}

func (Tester) TableName() string { return "el_tester" }

// UserExam el_user_exam
type UserExam struct {
	ID         string     `gorm:"column:id;primaryKey" json:"id"`
	UserID     string     `gorm:"column:user_id"       json:"userId"`
	ExamID     string     `gorm:"column:exam_id"       json:"examId"`
	TryCount   int        `gorm:"column:try_count"     json:"tryCount"`
	MaxScore   int        `gorm:"column:max_score"     json:"maxScore"`
	Passed     int8       `gorm:"column:passed"        json:"passed"`
	CreateTime *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time"   json:"updateTime"`
}

func (UserExam) TableName() string { return "el_user_exam" }

// UserBook el_user_book
type UserBook struct {
	ID         string     `gorm:"column:id;primaryKey" json:"id"`
	ExamID     string     `gorm:"column:exam_id"       json:"examId"`
	UserID     string     `gorm:"column:user_id"       json:"userId"`
	QuID       string     `gorm:"column:qu_id"         json:"quId"`
	CreateTime *time.Time `gorm:"column:create_time"   json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time"   json:"updateTime"`
	WrongCount int        `gorm:"column:wrong_count"   json:"wrongCount"`
	Title      string     `gorm:"column:title"         json:"title"`
	Sort       int        `gorm:"column:sort"          json:"sort"`
}

func (UserBook) TableName() string { return "el_user_book" }

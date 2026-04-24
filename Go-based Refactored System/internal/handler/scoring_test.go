package handler

import (
	"fmt"
	"math"
	"testing"
)

// ===================== standScore1 (心理特质 12维度) =====================

func TestStandScore1_AllZero(t *testing.T) {
	// 91题全答错(isRight=0)
	rows := make([]paperQuContent, 91)
	for i := range rows {
		rows[i] = paperQuContent{Content: "V" + itoa(i+1), IsRight: 0}
	}
	scores := standScore1(rows)
	if len(scores) != 12 {
		t.Fatalf("want 12 dimensions, got %d", len(scores))
	}
	// 所有分数应在 0~10 范围
	for k, v := range scores {
		if v < 0 || v > 10 {
			t.Errorf("%s: score %v out of range [0,10]", k, v)
		}
	}
}

func TestStandScore1_AllCorrect(t *testing.T) {
	// 91题全答对(isRight=1)
	rows := make([]paperQuContent, 91)
	for i := range rows {
		rows[i] = paperQuContent{Content: "V" + itoa(i+1), IsRight: 1}
	}
	scores := standScore1(rows)
	if len(scores) != 12 {
		t.Fatalf("want 12 dimensions, got %d", len(scores))
	}
	for k, v := range scores {
		if v < 0 || v > 10 {
			t.Errorf("%s: score %v out of range [0,10]", k, v)
		}
	}
}

func TestStandScore1_Dimensions(t *testing.T) {
	// 验证12个维度的键名
	rows := make([]paperQuContent, 91)
	for i := range rows {
		rows[i] = paperQuContent{Content: "V" + itoa(i+1), IsRight: 0}
	}
	scores := standScore1(rows)
	expected := []string{"焦虑", "抑郁", "心理失衡", "敌意", "恐惧", "身体不适", "认知衰退", "情绪化", "挫折感", "自我否定", "怀疑感", "职业倦怠"}
	for _, k := range expected {
		if _, ok := scores[k]; !ok {
			t.Errorf("missing dimension: %s", k)
		}
	}
}

func TestStandScore1_EmptyInput(t *testing.T) {
	scores := standScore1(nil)
	if len(scores) != 12 {
		t.Fatalf("empty input should still produce 12 dimensions, got %d", len(scores))
	}
}

// ===================== standScore2 (管理特质 13维度) =====================

func TestStandScore2_AllMiddle(t *testing.T) {
	// 140题全答中间值3
	rows := make([]paperQuContent, 140)
	for i := range rows {
		rows[i] = paperQuContent{Content: "V" + itoa(i+1), Answered: 1, ActualScore: 3}
	}
	scores := standScore2(rows)
	if len(scores) != 13 {
		t.Fatalf("want 13 dimensions, got %d", len(scores))
	}
	for k, v := range scores {
		if v < 0 || v > 5 {
			t.Errorf("%s: score %v out of range [0,5]", k, v)
		}
	}
}

func TestStandScore2_Dimensions(t *testing.T) {
	rows := make([]paperQuContent, 140)
	for i := range rows {
		rows[i] = paperQuContent{Content: "V" + itoa(i+1), Answered: 1, ActualScore: 3}
	}
	scores := standScore2(rows)
	expected := []string{"社会性", "进取性", "领导性", "计划性", "人际敏感性", "自信心", "责任心", "学习力", "创新性", "情绪稳定性", "自律性", "决断性", "合作性"}
	for _, k := range expected {
		if _, ok := scores[k]; !ok {
			t.Errorf("missing dimension: %s", k)
		}
	}
}

func TestStandScore2_ReverseScoring(t *testing.T) {
	// V28 is reverse-scored (6-V28) in 社会性
	// 如果 V28=5, 则 6-5=1; 如果 V28=1, 则 6-1=5
	// 验证反向计分使得高分选项产生低分
	rows1 := make([]paperQuContent, 140)
	rows2 := make([]paperQuContent, 140)
	for i := range rows1 {
		rows1[i] = paperQuContent{Content: "V" + itoa(i+1), Answered: 1, ActualScore: 3}
		rows2[i] = paperQuContent{Content: "V" + itoa(i+1), Answered: 1, ActualScore: 3}
	}
	// V28 (reverse-scored in 社会性): high raw → low score
	rows1[27].ActualScore = 5 // V28=5 → 6-5=1
	rows2[27].ActualScore = 1 // V28=1 → 6-1=5
	s1 := standScore2(rows1)["社会性"]
	s2 := standScore2(rows2)["社会性"]
	if s1 >= s2 {
		t.Errorf("reverse scoring failed: V28=5 → %v should be < V28=1 → %v", s1, s2)
	}
}

// ===================== psyScoreLevel / mngScoreLevel (等级判定) =====================

func TestPsyScoreLevel(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{0, "低分"}, {4.49, "低分"}, {4.5, "中分"}, {6.99, "中分"}, {7.0, "高分"}, {10, "高分"},
	}
	for _, tt := range tests {
		got := psyScoreLevel(tt.score)
		if got != tt.want {
			t.Errorf("psyScoreLevel(%v) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestMngScoreLevel(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{0, "低分"}, {0.99, "低分"}, {1.0, "较低分"}, {1.99, "较低分"},
		{2.0, "中等分"}, {2.99, "中等分"}, {3.0, "较高分"}, {3.99, "较高分"},
		{4.0, "高分"}, {5.0, "高分"},
	}
	for _, tt := range tests {
		got := mngScoreLevel(tt.score)
		if got != tt.want {
			t.Errorf("mngScoreLevel(%v) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestPsyHealthLevel(t *testing.T) {
	tests := []struct {
		scores map[string]float64
		want   string
	}{
		// 全低分 → 良好
		{map[string]float64{"a": 3, "b": 3, "c": 3}, "心理状态良好"},
		// avg≤7 但有2个>8 → 较好
		{map[string]float64{"a": 9, "b": 9, "c": 3}, "心理状况较好"},
		// avg>7 → 较差或很差
		{map[string]float64{"a": 8, "b": 8, "c": 8}, "心理状态较差"},
	}
	for _, tt := range tests {
		got := psyHealthLevel(tt.scores)
		if got != tt.want {
			t.Errorf("psyHealthLevel(%v) = %q, want %q", tt.scores, got, tt.want)
		}
	}
}

func TestMngTotalLevel(t *testing.T) {
	tests := []struct {
		total float64
		want  string
	}{
		{0, "低分"}, {11, "低分"}, {12, "较低分"}, {24, "中等分"}, {36, "较高分"}, {48, "高分"},
	}
	for _, tt := range tests {
		got := mngTotalLevel(tt.total)
		if got != tt.want {
			t.Errorf("mngTotalLevel(%v) = %q, want %q", tt.total, got, tt.want)
		}
	}
}

func TestMngDiagnosis(t *testing.T) {
	tests := []struct {
		total float64
		want  string
	}{
		{0, "非管理型"}, {12, "低潜管理者"}, {24, "中等管理者"}, {36, "较高潜管理者"}, {48, "高潜管理者"},
	}
	for _, tt := range tests {
		got := mngDiagnosis(tt.total)
		if got != tt.want {
			t.Errorf("mngDiagnosis(%v) = %q, want %q", tt.total, got, tt.want)
		}
	}
}

// ===================== capPageSize =====================

func TestCapPageSize(t *testing.T) {
	tests := []struct{ in, want int }{
		{0, 10}, {-1, 10}, {1, 1}, {10, 10}, {100, 100}, {200, 200}, {201, 200}, {99999, 200},
	}
	for _, tt := range tests {
		got := capPageSize(tt.in)
		if got != tt.want {
			t.Errorf("capPageSize(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

// ===================== evalExpr 扩展测试 =====================

func TestEvalExpr_ComplexFormula(t *testing.T) {
	// 焦虑公式: 1-V1+1-V36+V41+V56+1-V61+V81+1-V15
	// V1=1,V36=0,V41=1,V56=0,V61=1,V81=0,V15=1 → 0+1+1+0+0+0+0 = 2
	vars := map[string]float64{
		"V1": 1, "V36": 0, "V41": 1, "V56": 0, "V61": 1, "V81": 0, "V15": 1,
	}
	got, err := evalExpr("1-V1+1-V36+V41+V56+1-V61+V81+1-V15", vars)
	if err != nil {
		t.Fatal(err)
	}
	if got != 2 {
		t.Fatalf("want 2, got %v", got)
	}
}

func TestEvalExpr_StandScore2Formula(t *testing.T) {
	// 社会性公式: (V1+V14+6-V28+6-V42+6-V56+6-V72+6-V87+6-V102+6-V115+6-V128)/10.0
	// 所有V=3 → (3+3+3+3+3+3+3+3+3+3)/10 = 3.0
	vars := map[string]float64{}
	for i := 1; i <= 140; i++ {
		vars["V"+itoa(i)] = 3
	}
	got, err := evalExpr("(V1+V14+6-V28+6-V42+6-V56+6-V72+6-V87+6-V102+6-V115+6-V128)/10.0", vars)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-3.0) > 0.001 {
		t.Fatalf("want 3.0, got %v", got)
	}
}

// helper
func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

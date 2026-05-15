package handler

import (
	"fmt"
	"math"
	"testing"
)

// TestStandScore2_AllFour 验证 002 管理特质公式与客户 Excel 一致。
// 数据来自 docs/管理特质系统测算.xlsx R2（111 考生，140 题全 4 分）
// Excel 输出：社会性=2.4, 进取性=3.6, 自律性=3.2 ...
func TestStandScore2_AllFour(t *testing.T) {
	rows := make([]paperQuContent, 0, 140)
	for i := 1; i <= 140; i++ {
		rows = append(rows, paperQuContent{
			Content:     fmt.Sprintf("V%d", i),
			Answered:    1,
			ActualScore: 4,
		})
	}
	got := standScore2(rows)
	want := map[string]float64{
		"社会性":   2.4,
		"进取性":   3.6,
		"领导性":   3.6,
		"计划性":   3.8333,
		"人际敏感性": 3.5385,
		"自信心":   3.6667,
		"责任心":   3.2,
		"学习力":   3.2727,
		"创新性":   3.2,
		"情绪稳定性": 3.4,
		"自律性":   3.2,
		"决断性":   3.8333,
		"合作性":   3.6,
	}
	for k, exp := range want {
		v, ok := got[k]
		if !ok {
			t.Errorf("维度 %q 缺失", k)
			continue
		}
		if math.Abs(v-exp) > 0.001 {
			t.Errorf("维度 %q: got=%.4f want=%.4f diff=%+.4f", k, v, exp, v-exp)
		}
	}
}

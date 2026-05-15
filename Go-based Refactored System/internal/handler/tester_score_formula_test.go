package handler

import (
	"fmt"
	"math"
	"testing"
)

// TestStandScore1_GaoFenExcelR4 验证用客户 260428测算 Excel "高分"R4 的 90 题输入
// 计算的 12 维度结果与 Excel 公式输出一致（误差 < 0.01）。
//
// 数据来源：docs/260428职业心理测验系统测算.xlsx Sheet1 R4
// - 答题：B4..CM4（V1..V90，0/1）
// - 期望分数：CO4..CZ4（12 维度，由 Excel 公式自动计算）
//
// 此测试同时验证：
//  1. standScore1 算法实现正确
//  2. 公式系数（avg, sd）与 Excel 一致
//  3. paperQuContent.Content 用 "V1".."V90" 字符串（即题库编号 qu_repo.sort）
func TestStandScore1_GaoFenExcelR4(t *testing.T) {
	answers := []int8{
		0, 1, 0, 1, 0, 0, 0, 1, 1, 0, // V1-V10
		1, 0, 0, 0, 0, 0, 0, 0, 1, 1, // V11-V20
		0, 1, 1, 0, 1, 0, 0, 1, 0, 1, // V21-V30
		1, 1, 0, 0, 0, 0, 1, 0, 0, 0, // V31-V40
		0, 0, 0, 0, 0, 0, 1, 1, 0, 0, // V41-V50
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1, // V51-V60
		0, 1, 0, 1, 0, 0, 1, 1, 0, 1, // V61-V70
		1, 1, 0, 1, 0, 1, 1, 0, 0, 1, // V71-V80
		0, 0, 1, 1, 1, 0, 1, 0, 0, 0, // V81-V90
	}

	rows := make([]paperQuContent, 0, 90)
	for i, v := range answers {
		rows = append(rows, paperQuContent{
			Content:  fmt.Sprintf("V%d", i+1),
			IsRight:  v,
			Answered: 1,
		})
	}

	got := standScore1(rows)

	// Excel CO4..CZ4 输出
	want := map[string]float64{
		"焦虑":   7.9945,
		"抑郁":   3.4540,
		"心理失衡": 2.3702,
		"敌意":   1.9662,
		"恐惧":   3.6704,
		"身体不适": 7.8677,
		"认知衰退": 3.6821,
		"情绪化":  8.9773,
		"挫折感":  7.7840,
		"自我否定": 7.7144, // Excel 中名为"自我怀疑"，代码用"自我否定"，是同一维度
		"怀疑感":  7.1288,
		"职业倦怠": 5.5619,
	}

	for k, exp := range want {
		v, ok := got[k]
		if !ok {
			t.Errorf("维度 %q 缺失", k)
			continue
		}
		if math.Abs(v-exp) > 0.01 {
			t.Errorf("维度 %q: got=%.4f want=%.4f diff=%+.4f", k, v, exp, v-exp)
		}
	}
}

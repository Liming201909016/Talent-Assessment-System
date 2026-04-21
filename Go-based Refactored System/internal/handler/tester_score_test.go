package handler

import (
	"math"
	"testing"
)

func TestEvalExpr_ImplicitSubtraction(t *testing.T) {
	// 对齐 Aviator 行为：`1-V1+1-V36+V41` 应按左结合顺序求值：
	// (((1-V1)+1)-V36)+V41
	// 若 V1=0, V36=1, V41=1 → 1-0+1-1+1 = 2
	vars := map[string]float64{"V1": 0, "V36": 1, "V41": 1}
	got, err := evalExpr("1-V1+1-V36+V41", vars)
	if err != nil {
		t.Fatal(err)
	}
	if got != 2 {
		t.Fatalf("want 2, got %v", got)
	}
}

func TestEvalExpr_DivisionWithParens(t *testing.T) {
	// (V1+V2+6-V3)/10.0 ；V1=3, V2=4, V3=2 → (3+4+6-2)/10 = 11/10 = 1.1
	vars := map[string]float64{"V1": 3, "V2": 4, "V3": 2}
	got, err := evalExpr("(V1+V2+6-V3)/10.0", vars)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-1.1) > 1e-9 {
		t.Fatalf("want 1.1, got %v", got)
	}
}

func TestEvalExpr_PrecedenceWithoutParens(t *testing.T) {
	// 2+3*4 = 14
	got, err := evalExpr("2+3*4", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != 14 {
		t.Fatalf("want 14, got %v", got)
	}
}

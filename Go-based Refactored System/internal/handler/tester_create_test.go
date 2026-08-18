package handler

import "testing"

// TestBugFB124_ManualTesterCreateUsesTelephoneWhenIDNumberIsEmpty
// 对应：docs/regression-tests.md #FB-124
// 复现：新增封闭测评人员时身份证号标注选填，但后端强制要求idNumber。
// 期望：身份证为空时按手机号做测评内识别，默认密码为手机号后4位。
func TestBugFB124_ManualTesterCreateUsesTelephoneWhenIDNumberIsEmpty(t *testing.T) {
	telephone := "12341234123"
	identifierColumn, identifier, password, err := prepareTesterCreateIdentity(testerReq{
		ExamID: "exam-closed-1", Telephone: &telephone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if identifierColumn != "telephone" || identifier != telephone || password != "4123" {
		t.Fatalf("identity=%s|%s password=%s", identifierColumn, identifier, password)
	}

	idNumber := "110101199001011234"
	identifierColumn, identifier, password, err = prepareTesterCreateIdentity(testerReq{
		ExamID: "exam-closed-1", IDNumber: idNumber, Telephone: &telephone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if identifierColumn != "id_number" || identifier != idNumber || password != "4123" {
		t.Fatalf("identity=%s|%s password=%s", identifierColumn, identifier, password)
	}
}

func TestBugFB124_ManualTesterCreateRejectsMissingTelephoneAndIDNumber(t *testing.T) {
	if _, _, _, err := prepareTesterCreateIdentity(testerReq{ExamID: "exam-closed-1"}); err == nil {
		t.Fatal("missing telephone and ID number was accepted")
	}
}

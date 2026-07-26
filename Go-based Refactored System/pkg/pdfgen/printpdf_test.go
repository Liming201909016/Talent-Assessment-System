package pdfgen

import "testing"

func TestBugFB070_CompetencyReportContentStartsOnSecondPage(t *testing.T) {
	if got := contentPageRange("competency"); got != "2-" {
		t.Fatalf("competency content range=%q want 2-", got)
	}
	for _, reportType := range []string{"001", "002"} {
		if got := contentPageRange(reportType); got != "3-" {
			t.Fatalf("legacy report %s content range=%q want 3-", reportType, got)
		}
	}
}

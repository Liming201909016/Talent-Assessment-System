package service

import (
	"fmt"
	"testing"
)

// This independent capacity regression uses the public gate's actual GORM
// metadata scans. It does not alter fixtures or the existing signature tests.
func TestBugMTSchemaGeneratedParentCapacity(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for _, parent := range []string{"el_paper", "el_paper_qu"} {
		for length := 1; length <= 65; length++ {
			t.Run(fmt.Sprintf("%s/varchar%d", parent, length), func(t *testing.T) {
				m := managementTraitsSchemaFixture(d)
				for n := range m.Columns {
					if m.Columns[n].Table == parent && m.Columns[n].Name == "id" {
						m.Columns[n].ColumnType = fmt.Sprintf("varchar(%d)", length)
					}
				}
				var want error
				if length < 36 || length > 64 {
					want = ErrManagementTraitsRuntimeInvalid
				}
				managementSchemaCheckTwice(t, d, m, want)
			})
		}
	}
}

func TestManagementTraitsSchemaExternalExamCapacity(t *testing.T) {
	d, err := managementTraitsSchemaContract()
	if err != nil {
		t.Fatal(err)
	}
	for length := 1; length <= 65; length++ {
		t.Run(fmt.Sprintf("varchar%d", length), func(t *testing.T) {
			m := managementTraitsSchemaFixture(d)
			for n := range m.Columns {
				if m.Columns[n].Table == "el_exam" && m.Columns[n].Name == "id" {
					m.Columns[n].ColumnType = fmt.Sprintf("varchar(%d)", length)
				}
			}
			var want error
			if length > 64 {
				want = ErrManagementTraitsRuntimeInvalid
			}
			managementSchemaCheckTwice(t, d, m, want)
		})
	}
}

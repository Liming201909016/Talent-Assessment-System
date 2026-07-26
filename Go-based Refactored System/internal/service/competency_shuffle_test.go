package service

import (
	"errors"
	"reflect"
	"testing"
)

type sequenceReader struct{ values []byte }

func (r *sequenceReader) Read(p []byte) (int, error) {
	if len(r.values) == 0 {
		return 0, errors.New("random source failed")
	}
	for i := range p {
		p[i] = r.values[0]
		r.values = r.values[1:]
		if len(r.values) == 0 && i+1 < len(p) {
			return i + 1, errors.New("random source failed")
		}
	}
	return len(p), nil
}

func TestShuffleCompetencyQuestionIDs_CompleteAndDeterministic(t *testing.T) {
	original := []string{"q1", "q2", "q3", "q4"}
	left := append([]string(nil), original...)
	right := append([]string(nil), original...)
	if err := ShuffleCompetencyQuestionIDs(left, &sequenceReader{values: []byte{1, 2, 3, 4, 5, 6, 7, 8}}); err != nil {
		t.Fatal(err)
	}
	if err := ShuffleCompetencyQuestionIDs(right, &sequenceReader{values: []byte{1, 2, 3, 4, 5, 6, 7, 8}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("same random stream produced different order: %v vs %v", left, right)
	}
	seen := map[string]bool{}
	for _, id := range left {
		seen[id] = true
	}
	if len(seen) != len(original) {
		t.Fatalf("shuffle lost or duplicated IDs: %v", left)
	}
	for _, id := range original {
		if !seen[id] {
			t.Fatalf("shuffle lost %s: %v", id, left)
		}
	}
}

func TestShuffleCompetencyQuestionIDs_RandomFailure(t *testing.T) {
	ids := []string{"q1", "q2", "q3"}
	if err := ShuffleCompetencyQuestionIDs(ids, &sequenceReader{}); err == nil {
		t.Fatal("expected random source error")
	}
}

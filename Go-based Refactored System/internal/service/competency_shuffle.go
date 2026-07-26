package service

import (
	"crypto/rand"
	"io"
	"math/big"
)

// ShuffleCompetencyQuestionIDs applies Fisher-Yates using an injectable secure
// random source. Callers must abort their transaction when it returns an error.
func ShuffleCompetencyQuestionIDs(ids []string, source io.Reader) error {
	if source == nil {
		source = rand.Reader
	}
	for i := len(ids) - 1; i > 0; i-- {
		value, err := rand.Int(source, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		j := int(value.Int64())
		ids[i], ids[j] = ids[j], ids[i]
	}
	return nil
}

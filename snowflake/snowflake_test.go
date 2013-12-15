package snowflake

import (
	"testing"

	"github.com/simonz05/util/assert"
)

func TestSequence(t *testing.T) {
	ast := assert.NewAssert(t)
	sf, _ := New(0)
	prevId, err := sf.Next()
	ast.Nil(err)

	for i := 1; ((int64(i) + 1) & SequenceMax) != 0; i++ {
		curId, err := sf.Next()
		ast.Nil(err)
		ast.NotEqual(prevId, curId)
		prevId = curId
	}
}

func TestUniqueness(t *testing.T) {
	ast := assert.NewAssert(t)
	sf, _ := New(0)
	w0Id, err := sf.Next()
	ast.Nil(err)
	sf2, _ := New(1)
	w1Id, err := sf2.Next()
	ast.Nil(err)
	ast.NotEqual(w0Id, w1Id)
}

func BenchmarkIdGeneration(b *testing.B) {
	sf, _ := New(0)

	for n := 0; n < b.N; n++ {
		sf.Next()
	}
}

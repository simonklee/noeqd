package main

import (
	"bytes"
	"io"
	"testing"

	"github.com/simonz05/util/assert"
)

func TestServeZero(t *testing.T) {
	ast := assert.NewAssert(t)
	i, o := bytes.NewBuffer([]byte{0}), new(bytes.Buffer)
	err := serve(i, o)
	ast.Equal(0, o.Len())
	ast.Equal(ErrInvalidRequest, err)
}

func TestSequence(t *testing.T) {
	ast := assert.NewAssert(t)
	prevId, err := nextId()
	ast.Nil(err)

	for i := 1; ((int64(i) + 1) & sequenceMask) != 0; i++ {
		curId, err := nextId()
		ast.Nil(err)
		ast.NotEqual(prevId, curId)
		prevId = curId
	}
}

func TestUniqueness(t *testing.T) {
	ast := assert.NewAssert(t)
	w0Id, err := nextId()
	ast.Nil(err)
	*wid = 1
	w1Id, err := nextId()
	ast.Nil(err)
	ast.NotEqual(w0Id, w1Id)
}

func TestServeMoreThanZero(t *testing.T) {
	ast := assert.NewAssert(t)
	i, o := bytes.NewBuffer([]byte{1}), new(bytes.Buffer)
	err := serve(i, o)
	ast.Equal(io.EOF, err)
	ast.Equal(8, o.Len())

	i, o = bytes.NewBuffer([]byte{2}), new(bytes.Buffer)
	err = serve(i, o)
	ast.Equal(io.EOF, err)
	ast.Equal(16, o.Len())

	i, o = bytes.NewBuffer([]byte{255}), new(bytes.Buffer)
	err = serve(i, o)
	ast.Equal(io.EOF, err)
	ast.Equal(255*8, o.Len())
}

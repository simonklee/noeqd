package main

import (
	"bytes"
	"io"
	"sync"
	"testing"

	"github.com/simonz05/util/assert"
)

var (
	setupServerOnce sync.Once
)

func TestServeZero(t *testing.T) {
	setupServerOnce.Do(setupServer)
	ast := assert.NewAssert(t)
	i, o := bytes.NewBuffer([]byte{0}), new(bytes.Buffer)
	err := serve(i, o)
	ast.Equal(0, o.Len())
	ast.Equal(ErrInvalidRequest, err)
}

func TestServeMoreThanZero(t *testing.T) {
	setupServerOnce.Do(setupServer)
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

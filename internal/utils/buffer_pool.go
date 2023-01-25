package utils

import (
	"bytes"
	"sync"
)

var bufferPool = sync.Pool{New: func() interface{} {
	return new(bytes.Buffer)
}}

func AllocPooledBuffer() *bytes.Buffer {
	v, ok := bufferPool.Get().(*bytes.Buffer)
	if !ok {
		panic("Invalid Type")
	}

	v.Reset()

	return v
}

func ReleasePooledBuffer(b *bytes.Buffer) {
	bufferPool.Put(b)
}

package project

import (
	"bytes"
	"sync"
)

// ThreadSafeBuffer is a goroutine-safe bytes.Buffer.
type ThreadSafeBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *ThreadSafeBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *ThreadSafeBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

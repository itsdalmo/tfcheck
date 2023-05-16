package tfcheck

import (
	"io"
	"strings"
	"sync"
)

var _ io.Writer = &Buffer{}

// Buffer implements a concurrency-safe io.Writer with convenience methods
// for tailing the contents of the buffer.
type Buffer struct {
	current strings.Builder
	lines   []string
	mu      sync.Mutex
}

// Write implements io.Writer.
func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, c := range p {
		b.current.WriteByte(c)
		if c == '\n' {
			b.lines = append(b.lines, b.current.String())
			b.current.Reset()
		}
	}

	return len(p), nil
}

// Lines returns all lines written (including the current).
func (b *Buffer) Lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	lines := make([]string, len(b.lines))
	copy(lines, b.lines)

	if b.current.Len() > 0 {
		lines = append(lines, b.current.String())
	}

	return lines
}

// Tail returns the last N lines written (including the current).
func (b *Buffer) Tail(n int) []string {
	if n <= 0 {
		return []string{}
	}

	lines := b.Lines()
	count := len(lines)

	if n > count {
		n = count
	}

	return lines[count-n:]
}

// String implements fmt.Stringer.
func (b *Buffer) String() string {
	return strings.Join(b.Lines(), "")
}

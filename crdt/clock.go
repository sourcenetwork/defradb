package crdt

import (
	"time"
)

// Clock interface for generating increasing timestamp values
type Clock interface {
	Apply() int64
}

// DefaultClock implements Clock using increasing time values
// for Last-Writer-Win semantics
type DefaultClock int64

func (clock DefaultClock) apply() {
	max(time.Now().UnixNano(), int64(clock)+1)
}

// Max returns the larger of x or y.
func max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"container/list"
	"context"
	"sync"

	"github.com/sourcenetwork/defradb/errors"
)

// errCARBuildDidNotReturn is what requests waiting on a build receive when the build panicked or
// its goroutine exited, so they see a failure rather than an empty CAR reported as success.
var errCARBuildDidNotReturn = errors.New("building the CAR did not return")

// carCacheMaxBytes bounds the CAR bytes the serving side keeps. Peers ask for a head within
// moments of its announcement, so the cache only has to outlast one burst of requests.
const carCacheMaxBytes = 64 << 20

// carCache holds recently built CARs by head, so a head several peers ask for is built once.
// A CAR is fixed by its head and the blocks it links, none of which change once stored, so an
// entry never goes stale. Only a successful build is kept, and a failed one is retried by the
// next request, since the block it lacked may since have arrived.
//
// Requests for a head that is being built wait on that build rather than starting another, which
// is the case the cache exists for: every subscriber that needs a head asks for it at once.
//
// A cached CAR is shared between replies and must not be modified.
type carCache struct {
	mu       sync.Mutex
	maxBytes int
	bytes    int
	// order holds *carEntry, most recently used at the front.
	order    *list.List
	entries  map[string]*list.Element
	building map[string]*carBuild
}

type carEntry struct {
	key  string
	data []byte
}

// carBuild is a build in progress. data and err are set before done is closed.
type carBuild struct {
	done chan struct{}
	data []byte
	err  error
}

func newCARCache(maxBytes int) *carCache {
	return &carCache{
		maxBytes: maxBytes,
		order:    list.New(),
		entries:  make(map[string]*list.Element),
		building: make(map[string]*carBuild),
	}
}

// getOrBuild returns the CAR for key, calling build only when the CAR is neither cached nor being
// built. shared reports that the result came from the cache or from another request's build. A
// nil cache builds every time.
func (c *carCache) getOrBuild(
	ctx context.Context,
	key string,
	build func() ([]byte, error),
) (data []byte, shared bool, err error) {
	if c == nil {
		data, err = build()
		return data, false, err
	}

	c.mu.Lock()
	if el, ok := c.entries[key]; ok {
		c.order.MoveToFront(el)
		// entries only ever holds elements whose Value is a *carEntry.
		data := el.Value.(*carEntry).data //nolint:forcetypeassert
		c.mu.Unlock()
		return data, true, nil
	}
	if b, ok := c.building[key]; ok {
		c.mu.Unlock()
		select {
		case <-b.done:
			return b.data, true, b.err
		case <-ctx.Done():
			return nil, false, ctx.Err()
		}
	}
	b := &carBuild{done: make(chan struct{})}
	c.building[key] = b
	c.mu.Unlock()

	// Deferred so a build that panics still releases the requests waiting on it. The panic itself
	// carries on up this goroutine; the waiters get errCARBuildDidNotReturn in its place, since
	// otherwise they would read the zero data and error and report an empty CAR as served.
	returned := false
	defer func() {
		if !returned {
			b.data, b.err = nil, errCARBuildDidNotReturn
		}
		c.mu.Lock()
		delete(c.building, key)
		if b.err == nil && len(b.data) > 0 {
			c.add(key, b.data)
		}
		c.mu.Unlock()
		close(b.done)
	}()
	b.data, b.err = build()
	returned = true
	return b.data, false, b.err
}

// add stores data under key and evicts the least recently used entries past the byte budget. A
// CAR larger than the whole budget is not kept. The caller holds mu.
func (c *carCache) add(key string, data []byte) {
	if len(data) > c.maxBytes {
		return
	}
	c.entries[key] = c.order.PushFront(&carEntry{key: key, data: data})
	c.bytes += len(data)
	for c.bytes > c.maxBytes {
		// order only ever holds *carEntry values, pushed by add.
		oldest := c.order.Remove(c.order.Back()).(*carEntry) //nolint:forcetypeassert
		delete(c.entries, oldest.key)
		c.bytes -= len(oldest.data)
	}
}

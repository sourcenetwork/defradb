// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package debug

import (
	"fmt"
	"maps"
	"strings"
	"sync"
)

// ResourceTracker tracks resources (like iterators, connections, etc.) to help
// detect leaks and understand resource lifecycle. It is safe for concurrent use.
//
// The tracker is intentionally not tied to a Tracer; callers that want verbose
// per-operation logging should wrap Track/Untrack calls themselves.
type ResourceTracker struct {
	mu        sync.Mutex
	name      string
	resources map[uintptr]string // addr -> description
}

// NewResourceTracker creates a new tracker for resources of the given name.
func NewResourceTracker(name string) *ResourceTracker {
	return &ResourceTracker{
		name:      name,
		resources: make(map[uintptr]string),
	}
}

// Track registers a resource with an optional description.
// Returns false if the resource was already tracked (duplicate addr).
func (rt *ResourceTracker) Track(resource any, description string) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	addr := Addr(resource)
	if _, exists := rt.resources[addr]; exists {
		return false
	}
	rt.resources[addr] = description
	return true
}

// Untrack removes a resource from tracking.
// Returns false if the resource was not currently tracked.
func (rt *ResourceTracker) Untrack(resource any) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	addr := Addr(resource)
	if _, exists := rt.resources[addr]; !exists {
		return false
	}
	delete(rt.resources, addr)
	return true
}

// Count returns the current number of tracked resources.
func (rt *ResourceTracker) Count() int {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return len(rt.resources)
}

// Snapshot returns a point-in-time copy of all currently tracked resources,
// keyed by address. The returned map is safe to read without holding any lock.
func (rt *ResourceTracker) Snapshot() map[uintptr]string {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return maps.Clone(rt.resources)
}

// Remaining returns the descriptions of all currently tracked resources.
// Useful for human-readable leak reports.
func (rt *ResourceTracker) Remaining() []string {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	result := make([]string, 0, len(rt.resources))
	for addr, desc := range rt.resources {
		result = append(result, fmt.Sprintf("%s (addr: %x)", desc, addr))
	}
	return result
}

// AddedSince returns resources tracked now that were not in snap.
// Useful to isolate leaks introduced by code that ran between Snapshot
// and AddedSince: opens that never got matching closes.
func (rt *ResourceTracker) AddedSince(snap map[uintptr]string) map[uintptr]string {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	added := make(map[uintptr]string)
	for addr, desc := range rt.resources {
		if _, existed := snap[addr]; !existed {
			added[addr] = desc
		}
	}
	return added
}

// StillHeld returns resources from snap that are still tracked now,
// i.e. resources that were alive at snap-time and have not been released.
// Useful to find resources held longer than expected.
func (rt *ResourceTracker) StillHeld(snap map[uintptr]string) map[uintptr]string {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	held := make(map[uintptr]string)
	for addr, desc := range snap {
		if _, still := rt.resources[addr]; still {
			held[addr] = desc
		}
	}
	return held
}

// Clear removes all tracked resources. Useful in test teardown.
func (rt *ResourceTracker) Clear() {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.resources = make(map[uintptr]string)
}

// TB is the subset of testing.TB used by AssertEmpty and AssertCount.
// *testing.T and *testing.B satisfy this interface.
type TB interface {
	Helper()
	Errorf(format string, args ...any)
}

// AssertEmpty fails the test if any resources are still tracked.
// It marks itself as a test helper so failures point to the call site.
func (rt *ResourceTracker) AssertEmpty(t TB) {
	t.Helper()
	remaining := rt.Remaining()
	if len(remaining) > 0 {
		t.Errorf("ResourceTracker[%s]: %d leaked resource(s):\n  %s",
			rt.name, len(remaining), strings.Join(remaining, "\n  "))
	}
}

// AssertCount fails the test if the tracked count does not equal want.
func (rt *ResourceTracker) AssertCount(t TB, want int) {
	t.Helper()
	if got := rt.Count(); got != want {
		t.Errorf("ResourceTracker[%s]: Count() = %d, want %d; remaining: %v",
			rt.name, got, want, rt.Remaining())
	}
}

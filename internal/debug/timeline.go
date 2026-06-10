// Copyright 2026 Democratized Data Foundation
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
	"sync"
	"time"
)

// Timeline is a single-process timeline collector intended for producing
// a visual ordered log of significant events across multiple goroutines.
// Each entry records the elapsed time since the timeline's first call and
// a short actor tag (e.g. node short-peer-id). Entries are appended in
// arrival order; concurrent appends are serialized.
//
// This is purely a debug/diagnostics aid; not for production use.
type Timeline struct {
	mu      sync.Mutex
	start   time.Time
	entries []TimelineEntry
	enabled bool
}

// TimelineEntry is one line in a Timeline.
type TimelineEntry struct {
	OffsetMs int64
	Actor    string
	Message  string
}

// DefaultTimeline is the package-global timeline.
var DefaultTimeline = NewTimeline()

// NewTimeline creates a fresh disabled timeline.
func NewTimeline() *Timeline {
	return &Timeline{}
}

// Enable arms the timeline; the first subsequent Log call sets T+0.
func (tl *Timeline) Enable() {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.enabled = true
	tl.entries = nil
	tl.start = time.Time{}
}

// Disable stops collection but does not clear entries.
func (tl *Timeline) Disable() {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.enabled = false
}

// Log appends an event. Safe to call concurrently. No-op if disabled.
func (tl *Timeline) Log(actor, format string, args ...any) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	if !tl.enabled {
		return
	}
	if tl.start.IsZero() {
		tl.start = time.Now()
	}
	tl.entries = append(tl.entries, TimelineEntry{
		OffsetMs: time.Since(tl.start).Milliseconds(),
		Actor:    actor,
		Message:  fmt.Sprintf(format, args...),
	})
}

// Render returns the entries formatted as an ordered visual log.
// Lines are aligned: T+Xms  ACTOR  MESSAGE.
func (tl *Timeline) Render() string {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	if len(tl.entries) == 0 {
		return "(timeline empty)\n"
	}
	// Find max actor width for alignment.
	maxActor := 0
	for _, e := range tl.entries {
		if len(e.Actor) > maxActor {
			maxActor = len(e.Actor)
		}
	}
	var out string
	for _, e := range tl.entries {
		out += fmt.Sprintf("T+%5dms  %-*s  %s\n", e.OffsetMs, maxActor, e.Actor, e.Message)
	}
	return out
}

// Entries returns a copy of the recorded entries.
func (tl *Timeline) Entries() []TimelineEntry {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	out := make([]TimelineEntry, len(tl.entries))
	copy(out, tl.entries)
	return out
}

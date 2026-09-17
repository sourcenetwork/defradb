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
	"testing"
)

// fakeIter is a stand-in for any pointer-typed resource (iterator, connection, etc.).
type fakeIter struct{ id int }

func TestResourceTracker_BasicLifecycle(t *testing.T) {
	tr := NewResourceTracker("iter")

	a := &fakeIter{1}
	b := &fakeIter{2}
	c := &fakeIter{3}

	tr.Track(a, "iter-a")
	tr.Track(b, "iter-b")
	tr.Track(c, "iter-c")
	tr.AssertCount(t, 3)

	// Close two of three.
	tr.Untrack(a)
	tr.Untrack(b)
	tr.AssertCount(t, 1)

	// One resource should remain.
	remaining := tr.Remaining()
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining resource, got %d: %v", len(remaining), remaining)
	}
	// The description should mention iter-c.
	if len(remaining[0]) == 0 {
		t.Error("remaining description is empty")
	}
}

func TestResourceTracker_DuplicateTrack(t *testing.T) {
	tr := NewResourceTracker("conn")
	r := &fakeIter{42}

	if !tr.Track(r, "first") {
		t.Fatal("first Track should return true")
	}
	if tr.Track(r, "duplicate") {
		t.Fatal("second Track of same pointer should return false")
	}
	tr.AssertCount(t, 1)
}

func TestResourceTracker_UntrackUnknown(t *testing.T) {
	tr := NewResourceTracker("conn")
	r := &fakeIter{99}
	if tr.Untrack(r) {
		t.Fatal("Untrack of unregistered resource should return false")
	}
}

func TestResourceTracker_Snapshot_AddedSince_StillHeld(t *testing.T) {
	tr := NewResourceTracker("iter")

	a := &fakeIter{1}
	b := &fakeIter{2}
	c := &fakeIter{3}

	tr.Track(a, "iter-a")
	tr.Track(b, "iter-b")

	// Checkpoint: two resources live at this point.
	snap := tr.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("snapshot should contain 2 entries, got %d", len(snap))
	}

	// Open a third after the snapshot, close one of the originals.
	tr.Track(c, "iter-c")
	tr.Untrack(a)

	// AddedSince should report only iter-c (opened after snap).
	added := tr.AddedSince(snap)
	if len(added) != 1 {
		t.Fatalf("AddedSince should report 1 new resource, got %d: %v", len(added), added)
	}

	// StillHeld should report only iter-b (in snap, still live).
	held := tr.StillHeld(snap)
	if len(held) != 1 {
		t.Fatalf("StillHeld should report 1 held resource, got %d: %v", len(held), held)
	}
}

func TestResourceTracker_LeakScenario(t *testing.T) {
	// Simulates opening 3 iterators, closing 2, and asserting via AssertEmpty
	// which should report the one still open.
	tr := NewResourceTracker("corekv.Iterator")

	iters := []*fakeIter{{1}, {2}, {3}}
	for i, it := range iters {
		tr.Track(it, "iter-"+string(rune('A'+i)))
	}

	// Close all but the last.
	tr.Untrack(iters[0])
	tr.Untrack(iters[1])

	// AssertEmpty should fail — capture with a spy that satisfies TB.
	spy := &spyT{}
	tr.AssertEmpty(spy)
	if !spy.failed {
		t.Fatal("AssertEmpty should have reported a leak for the unclosed iterator")
	}

	// Cleanup.
	tr.Untrack(iters[2])
	tr.AssertEmpty(t)
}

// spyT satisfies the TB interface and records whether Errorf was called.
type spyT struct {
	failed bool
}

func (s *spyT) Errorf(format string, args ...any) {
	s.failed = true
}

func (s *spyT) Helper() {}

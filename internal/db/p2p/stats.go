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
	"hash/maphash"
	"sort"
	"sync"
)

// Reasons a CAR could not be generated or a DAG walk could not finish. They are a fixed
// set: deriving a reason from an error message would let a remote peer grow the counter
// map without bound.
const (
	reasonRootLink   = "rootLink"
	reasonWalk       = "walk"
	reasonCARWriter  = "carWriter"
	reasonBlockRead  = "blockRead"
	reasonCARPut     = "carPut"
	reasonStoreRoot  = "storeRoot"
	reasonBlockLink  = "blockLink"
	reasonIsMerged   = "isMerged"
	reasonVerifySig  = "verifySig"
	reasonEncKeys    = "encKeys"
	reasonLoadLink   = "loadLink"
	reasonDecodeLink = "decodeLink"
	reasonContext    = "context"
	reasonReader     = "carReader"
	reasonNoRoots    = "carNoRoots"
	reasonNext       = "carNextBlock"
)

// failureReasons counts failures by reason and remembers which reasons it has already
// logged. A repeated failure costs a counter increment rather than a log line; the first
// occurrence of each reason gets one line carrying the underlying error.
type failureReasons struct {
	mu      sync.Mutex
	counts  map[string]int64
	flagged map[string]struct{}
}

// record counts one failure and reports whether this reason has not been logged yet.
func (f *failureReasons) record(reason string) (firstSeen bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.counts == nil {
		f.counts = make(map[string]int64)
		f.flagged = make(map[string]struct{})
	}
	f.counts[reason]++
	if _, seen := f.flagged[reason]; seen {
		return false
	}
	f.flagged[reason] = struct{}{}
	return true
}

// drain returns the counts accumulated since the last call, sorted by reason, and resets
// them. The set of already-logged reasons is deliberately kept: a reason is logged once
// per process, not once per interval.
func (f *failureReasons) drain() []reasonCount {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.counts) == 0 {
		return nil
	}
	out := make([]reasonCount, 0, len(f.counts))
	for reason, n := range f.counts {
		out = append(out, reasonCount{reason: reason, count: n})
		delete(f.counts, reason)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].reason < out[j].reason })
	return out
}

type reasonCount struct {
	reason string
	count  int64
}

// dedupWindowCap bounds the memory one window can hold, so a burst cannot turn the
// instrument into a leak. At the cap the window holds 64k hashes, a little over 1 MiB.
const dedupWindowCap = 1 << 16

// wireDedup counts, over one reporting interval, how many inbound messages carried bytes
// this node had already seen. Gossipsub forwards a published message verbatim, so a
// redelivery of the same publish is byte-identical while the same document published by
// two different indexers is not: this measures redelivery, not agreement.
//
// It hashes the wire bytes rather than the decoded head CID because a message is dropped
// at the door before it is decoded, and the dropped share is the population of interest.
type wireDedup struct {
	seed maphash.Seed

	mu        sync.Mutex
	seen      map[uint64]struct{}
	total     int64
	truncated bool
}

func newWireDedup() *wireDedup {
	return &wireDedup{
		seed: maphash.MakeSeed(),
		seen: make(map[uint64]struct{}),
	}
}

// observe records one inbound message by the hash of its wire bytes. The hash is taken
// outside the lock so a large message does not serialise the pubsub dispatcher.
//
// Nil-safe: this runs on the pubsub dispatcher, where a panic takes the node's inbound
// path with it, and a P2P built by struct literal rather than by New has no window.
func (d *wireDedup) observe(msg []byte) {
	if d == nil {
		return
	}
	h := maphash.Bytes(d.seed, msg)

	d.mu.Lock()
	defer d.mu.Unlock()
	d.total++
	if len(d.seen) >= dedupWindowCap {
		d.truncated = true
		return
	}
	d.seen[h] = struct{}{}
}

// drain returns the totals for the interval and starts a new window. distinct is a floor
// rather than an exact count when truncated is true.
func (d *wireDedup) drain() (total, distinct int64, truncated bool) {
	if d == nil {
		return 0, 0, false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	total, distinct, truncated = d.total, int64(len(d.seen)), d.truncated
	d.total = 0
	d.truncated = false
	d.seen = make(map[uint64]struct{})
	return total, distinct, truncated
}

// dropDoc counts an inbound document that was meant to be stored and was not, naming why.
// The reason set is fixed, so a peer cannot grow the map by varying its errors.
func (p *P2P) dropDoc(reason string) {
	p.statDroppedDocs.Add(1)
	p.docDropReason.record(reason)
}

// skipDoc counts an inbound document deliberately not merged: already held, or excluded
// by access or the replication filter. Kept apart from drops, which are losses.
func (p *P2P) skipDoc(reason string) {
	p.statSkippedDocs.Add(1)
	p.docDropReason.record("skip:" + reason)
}

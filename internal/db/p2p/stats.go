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
	"context"
	"sort"
	"sync"

	"github.com/sourcenetwork/defradb/errors"
)

// Reasons a CAR could not be generated or a DAG walk could not finish. They are a fixed
// set: deriving a reason from an error message would let a remote peer grow the counter
// map without bound.
const (
	reasonRootLink    = "rootLink"
	reasonWalk        = "walk"
	reasonCARWriter   = "carWriter"
	reasonBlockRead   = "blockRead"
	reasonCARPut      = "carPut"
	reasonStoreRoot   = "storeRoot"
	reasonBlockLink   = "blockLink"
	reasonIsMerged    = "isMerged"
	reasonVerifySig   = "verifySig"
	reasonEncKeys     = "encKeys"
	reasonLoadLink    = "loadLink"
	reasonDecodeLink  = "decodeLink"
	reasonContext     = "context"
	reasonReader      = "carReader"
	reasonNoRoots     = "carNoRoots"
	reasonNext        = "carNextBlock"
	reasonNoRootBlock = "carNoRootBlock"

	// An unrecognised failure counts as reasonOther, so that staying non-zero means
	// syncDAGReason has fallen behind the code.
	reasonOther = "other"
)

// syncDAGReason names the step a DAG sync failed on. The specific failures come first: a
// load that failed because the context was cancelled is more usefully reported as a load
// failure than as a cancellation.
func syncDAGReason(err error) string {
	switch {
	case errors.Is(err, ErrStoreBlockDAGSync):
		return reasonStoreRoot
	case errors.Is(err, ErrGenerateBlockLink):
		return reasonBlockLink
	case errors.Is(err, ErrCheckBlockMerged):
		return reasonIsMerged
	case errors.Is(err, ErrVerifyBlockSig):
		return reasonVerifySig
	case errors.Is(err, ErrGetEncKeysForBlock), errors.Is(err, ErrRetrieveEncKey):
		return reasonEncKeys
	case errors.Is(err, ErrLoadLinkedBlock):
		return reasonLoadLink
	case errors.Is(err, ErrDecodeLinkedBlock):
		return reasonDecodeLink
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return reasonContext
	default:
		return reasonOther
	}
}

// Causes a document can be dropped for.
const (
	dropAccessError   = "accessError"
	dropBlockDecode   = "blockDecode"
	dropCIDMismatch   = "cidMismatch"
	dropGenerateLink  = "generateLink"
	dropImportCAR     = "importCAR"
	dropInvalidCID    = "invalidCID"
	dropIsMergedError = "isMergedError"
	dropMergeFailed   = "mergeFailed"
	dropSyncDAG       = "syncDAG"
	dropSyncQueueFull = "syncQueueFull"
)

// Reasons a document was skipped.
const (
	skipAlreadyMerged = "alreadyMerged"
	skipDuplicateHead = "duplicateHead"
	skipFiltered      = "filtered"
	skipInFlight      = "inFlight"
	skipNoAccess      = "noAccess"
)

// failureReasons counts occurrences by reason, drained once per report interval. Callers
// that also log keep the set of reasons already logged, so a repeated occurrence costs a
// counter increment rather than a line.
type failureReasons struct {
	mu      sync.Mutex
	counts  map[string]int64
	flagged map[string]struct{}
}

// record counts one occurrence of reason.
func (f *failureReasons) record(reason string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.counts == nil {
		f.counts = make(map[string]int64)
	}
	f.counts[reason]++
}

// recordFirst counts one occurrence and reports whether this reason has not been seen before,
// so the caller can log the first and count the rest.
func (f *failureReasons) recordFirst(reason string) (firstSeen bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.counts == nil {
		f.counts = make(map[string]int64)
	}
	if f.flagged == nil {
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

// dropDoc counts an inbound document that was meant to be stored and was not, naming why.
// The reason set is fixed, so a peer cannot grow the map by varying its errors.
func (p *P2P) dropDoc(reason string) {
	p.statDroppedDocs.Add(1)
	p.docDropReason.record(reason)
}

// skipDoc counts an inbound document deliberately not merged: already held, already in
// flight, repeated within one document-sync round, or excluded by access or the replication
// filter. Kept apart from drops, which are losses.
func (p *P2P) skipDoc(reason string) {
	p.statSkippedDocs.Add(1)
	p.docSkipReason.record(reason)
}

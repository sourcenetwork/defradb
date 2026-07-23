// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

func TestBatchCIDCollector_LenTracksAddedCIDs(t *testing.T) {
	collector := NewBatchCIDCollector()
	require.Equal(t, 0, collector.Len())

	collector.Add(newTestHeadsCID(t, []byte("a")))
	collector.Add(newTestHeadsCID(t, []byte("b")))
	require.Equal(t, 2, collector.Len())
}

func TestBatchCIDCollector_TruncateRollsBackToMark(t *testing.T) {
	collector := NewBatchCIDCollector()
	a := newTestHeadsCID(t, []byte("a"))
	b := newTestHeadsCID(t, []byte("b"))
	collector.Add(a)
	collector.Add(b)

	mark := collector.Len()
	collector.Add(newTestHeadsCID(t, []byte("c")))
	collector.Add(newTestHeadsCID(t, []byte("d")))
	require.Equal(t, 4, collector.Len())

	collector.Truncate(mark)
	require.Equal(t, []cid.Cid{a, b}, collector.GetCIDs())
}

func TestBatchCIDCollector_TruncatePanicsOutOfRange(t *testing.T) {
	collector := NewBatchCIDCollector()
	a := newTestHeadsCID(t, []byte("a"))
	b := newTestHeadsCID(t, []byte("b"))
	collector.Add(a)
	collector.Add(b)

	require.Panics(t, func() { collector.Truncate(collector.Len() + 1) })
	require.Panics(t, func() { collector.Truncate(-1) })

	// The lock is released before the panic, so GetCIDs stays intact and does not deadlock.
	require.Equal(t, []cid.Cid{a, b}, collector.GetCIDs())
}

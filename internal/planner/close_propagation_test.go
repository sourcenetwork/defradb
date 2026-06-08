// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// Every child of a parallelNode must get Close() called, even when an
// earlier child's Close returned an error. Short-circuiting would
// strand iterators on later children's badger txns.
func TestParallelNode_Close_CallsAllChildrenEvenIfOneErrors(t *testing.T) {
	c1 := &trackingPlanNode{closeErr: errors.New("intentional close failure")}
	c2 := &trackingPlanNode{}
	c3 := &trackingPlanNode{}

	p := &parallelNode{
		children: []planNode{c1, c2, c3},
	}

	err := p.Close()
	require.Error(t, err)
	require.True(t, c1.closeCalled)
	require.True(t, c2.closeCalled)
	require.True(t, c3.closeCalled)
}

// Both sides of an invertibleTypeJoin must get Close() called, even
// when one of them errors.
func TestInvertibleTypeJoin_Close_ClosesBothSidesEvenIfOneErrors(t *testing.T) {
	parent := &trackingPlanNode{closeErr: errors.New("intentional parent close failure")}
	child := &trackingPlanNode{}

	join := &invertibleTypeJoin{
		parentSide: joinSide{plan: parent},
		childSide:  joinSide{plan: child},
	}

	err := join.Close()
	require.Error(t, err)
	require.True(t, parent.closeCalled)
	require.True(t, child.closeCalled)
}

// If parentSide.Init errors after childSide.Init succeeded, childSide
// must be Closed so any resources it opened are released.
func TestInvertibleTypeJoin_Init_ClosesChildSideIfParentInitFails(t *testing.T) {
	parent := &trackingPlanNode{initErr: errors.New("intentional parent init failure")}
	child := &trackingPlanNode{}

	join := &invertibleTypeJoin{
		parentSide: joinSide{plan: parent},
		childSide:  joinSide{plan: child},
	}

	err := join.Init()
	require.Error(t, err)
	require.True(t, child.initCalled)
	require.True(t, parent.initCalled)
	require.True(t, child.closeCalled)
}

// Same as the Init test but on Start.
func TestInvertibleTypeJoin_Start_ClosesChildSideIfParentStartFails(t *testing.T) {
	parent := &trackingPlanNode{startErr: errors.New("intentional parent start failure")}
	child := &trackingPlanNode{}

	join := &invertibleTypeJoin{
		parentSide: joinSide{plan: parent},
		childSide:  joinSide{plan: child},
	}

	err := join.Start()
	require.Error(t, err)
	require.True(t, child.startCalled)
	require.True(t, parent.startCalled)
	require.True(t, child.closeCalled)
}

//  1. Close() after a failed Init() must return its own outcome (nil),
//     not the stuck Init error.
func TestMultiScanNode_Close_ReturnsNilAfterFailedInit(t *testing.T) {
	inner := &trackingPlanNode{initErr: errors.New("init boom")}
	ms := &multiScanNode{planNode: inner, numReaders: 1}

	require.ErrorContains(t, ms.Init(), "init boom")
	require.NoError(t, ms.Close(),
		"Close() must report its own outcome, not a stale Init error")
	require.True(t, inner.closeCalled)
}

// 2. Stuck-error must not leak across reader cycles.
func TestMultiScanNode_Next_ClearsPriorErrorOnNextCycle(t *testing.T) {
	inner := &scriptedPlanNode{
		nextResults: []nextResult{
			{ok: false, err: errors.New("transient")}, // cycle 1
			{ok: true, err: nil},                      // cycle 2
		},
	}
	ms := &multiScanNode{planNode: inner, numReaders: 2}

	ok, err := ms.Next()
	require.ErrorContains(t, err, "transient")
	require.False(t, ok)
	ok, err = ms.Next()
	require.ErrorContains(t, err, "transient")
	require.False(t, ok)

	ok, err = ms.Next()
	require.NoError(t, err, "stale err leaked into next cycle")
	require.True(t, ok)
}

//  3. Buffered-fan-out contract: within a cycle, all readers see the
//     cached result without re-invoking the underlying node.
func TestMultiScanNode_ReadersShareSameResultWithinCycle(t *testing.T) {
	inner := &scriptedPlanNode{
		nextResults: []nextResult{{ok: true, err: nil}},
	}
	ms := &multiScanNode{planNode: inner, numReaders: 3}

	for i := 0; i < 3; i++ {
		ok, err := ms.Next()
		require.NoError(t, err)
		require.True(t, ok, "reader %d should see cached cycle result", i)
	}
	require.Equal(t, 1, inner.nextInvocationCount,
		"underlying Next() should be invoked exactly once per cycle")
}

type trackingPlanNode struct {
	initCalled, startCalled, closeCalled bool
	initErr, startErr, closeErr          error
}

func (n *trackingPlanNode) Init() error {
	n.initCalled = true
	return n.initErr
}

func (n *trackingPlanNode) Start() error {
	n.startCalled = true
	return n.startErr
}

func (n *trackingPlanNode) Close() error {
	n.closeCalled = true
	return n.closeErr
}

func (n *trackingPlanNode) Prefixes([]keys.Walkable) {}
func (n *trackingPlanNode) Next() (bool, error)      { return false, nil }
func (n *trackingPlanNode) Value() core.Doc          { return core.Doc{} }
func (n *trackingPlanNode) Source() planNode         { return nil }
func (n *trackingPlanNode) Kind() string             { return "trackingPlanNode" }
func (n *trackingPlanNode) DocumentMap() *core.DocumentMapping {
	return &core.DocumentMapping{}
}

type nextResult struct {
	ok  bool
	err error
}

// scriptedPlanNode is a test helper whose Next() walks a pre-canned
// sequence of results and counts invocations. Init/Start/Close are
// inherited as no-ops from trackingPlanNode.
type scriptedPlanNode struct {
	trackingPlanNode
	nextResults         []nextResult
	nextInvocationCount int
}

func (n *scriptedPlanNode) Next() (bool, error) {
	idx := n.nextInvocationCount
	n.nextInvocationCount++
	if idx >= len(n.nextResults) {
		return false, nil
	}
	r := n.nextResults[idx]
	return r.ok, r.err
}

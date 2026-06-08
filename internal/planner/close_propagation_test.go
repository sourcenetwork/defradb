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

// Every child of a topLevelNode must get Close() called, even when an
// earlier child's Close returned an error. Short-circuiting would
// strand iterators on later children's shared transaction.
func TestTopLevelNode_Close_CallsAllChildrenEvenIfOneErrors(t *testing.T) {
	c1 := &trackingPlanNode{closeErr: errors.New("intentional close failure")}
	c2 := &trackingPlanNode{}
	c3 := &trackingPlanNode{}

	n := &topLevelNode{
		children: []planNode{c1, c2, c3},
	}

	err := n.Close()
	require.Error(t, err)
	require.ErrorContains(t, err, "intentional close failure")
	require.True(t, c1.closeCalled, "first child Close() should be called")
	require.True(t, c2.closeCalled, "second child Close() must still be called after first errors")
	require.True(t, c3.closeCalled, "third child Close() must still be called after first errors")
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

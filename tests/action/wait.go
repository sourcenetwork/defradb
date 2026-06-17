// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package action

import (
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/immutable"
)

// Wait is an action that will wait for the first not-none property
// to be fulfilled.
type Wait struct {
	stateful

	// Duration is the duration to wait.
	Duration immutable.Option[time.Duration]

	// The action execution status to wait for.
	//
	// Due to implementation limitations, this will pick up events that occured
	// before this test action begins execution.
	Action immutable.Option[client.ActionExecution]
}

var _ Action = (*Wait)(nil)
var _ Stateful = (*Wait)(nil)

func (a *Wait) Execute() {
	switch {
	case a.Duration.HasValue() && a.Action.HasValue():
		for {
			select {
			case <-time.After(a.Duration.Value()):
				return

			case action := <-a.s.Nodes[0].Event.Action.Message():
				expected := a.Action.Value()
				//nolint:forcetypeassert
				actual := action.Data.(event.ActionExecution)
				if expected.Action == actual.Action &&
					expected.CollectionID == actual.CollectionID &&
					expected.Status == actual.Status {
					return
				}
			}
		}

	case a.Duration.HasValue() && !a.Action.HasValue():
		<-time.After(a.Duration.Value())
		return

	case !a.Duration.HasValue() && a.Action.HasValue():
		for {
			action := <-a.s.Nodes[0].Event.Action.Message()
			expected := a.Action.Value()
			//nolint:forcetypeassert
			actual := action.Data.(event.ActionExecution)
			if expected.Action == actual.Action &&
				expected.CollectionID == actual.CollectionID &&
				expected.Status == actual.Status {
				return
			}
		}

	default:
		return
	}
}

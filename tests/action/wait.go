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

// Wait is an action that will wait for the given duration.
type Wait struct {
	stateful

	// Duration is the duration to wait.
	Duration immutable.Option[time.Duration]

	Action immutable.Option[client.ActionExecution]
}

var _ Action = (*Wait)(nil)
var _ Stateful = (*Wait)(nil)

func (a *Wait) Execute() {
	for {
		select {
		case <-time.After(orDefault(a.Duration, time.Hour)):
			return

		case action := <-a.s.Nodes[0].Event.Action.Message():
			if a.Action.HasValue() {
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
	}
}

func orDefault(opt immutable.Option[time.Duration], defaultDuration time.Duration) time.Duration {
	if opt.HasValue() {
		return opt.Value()
	}

	return defaultDuration
}

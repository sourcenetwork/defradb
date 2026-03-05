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
	"github.com/sourcenetwork/testo/action"

	"github.com/sourcenetwork/defradb/tests/state"
)

type Action = action.Action
type Actions = action.Actions
type Stateful = action.Stateful[*state.State]

type stateful struct {
	s *state.State
}

var _ Stateful = (*stateful)(nil)

func (a *stateful) SetState(s *state.State) {
	if a == nil {
		a = &stateful{}
	}
	a.s = s
}

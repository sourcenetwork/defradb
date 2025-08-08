// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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

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
	"time"
)

// CreateTx executes the `client tx new` command and appends the returned transaction id
// to state.Txns.
type Wait struct {
	stateful

	Duration time.Duration
}

var _ Action = (*Wait)(nil)

func (a *Wait) Execute() {
	time.Sleep(a.Duration)
}

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
)

// Wait is an action that will wait for the given duration.
type Wait struct {
	stateful

	// Duration is the duration to wait.
	Duration time.Duration
}

var _ Action = (*Wait)(nil)
var _ Stateful = (*Wait)(nil)

func (a *Wait) Execute() {
	<-time.After(a.Duration)
}

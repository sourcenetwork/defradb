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

package state

import (
	"github.com/sourcenetwork/immutable"
)

func GetNodeAudience(s *State, nodeIndex int) immutable.Option[string] {
	return immutable.None[string]()
}

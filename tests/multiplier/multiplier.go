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

package multiplier

import (
	"github.com/sourcenetwork/corelog"

	m "github.com/sourcenetwork/testo/multiplier"
)

type Multiplier = m.Multiplier
type Name = m.Name

var log = corelog.NewLogger("tests.multiplier")

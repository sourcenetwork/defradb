// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/state"

	"github.com/sourcenetwork/immutable"
)

func setupSourceHub(s *state.State) ([]node.DocumentACPOpt, error) {
	return s.DocumentACPOptions, nil
}

func getNodeAudience(s *state.State, nodeIndex int) immutable.Option[string] {
	return immutable.None[string]()
}

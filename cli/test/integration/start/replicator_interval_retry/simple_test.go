// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package start

import (
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
)

func TestStart_WithReplicatorInterval_NoError(t *testing.T) {
	arguments := []string{"--replicator-retry-intervals=10,20,40"}
	test := &integration.Test{
		Actions: []action.Action{
			action.StartWithArgs(arguments, nil),
		},
	}
	test.Execute(t)
}

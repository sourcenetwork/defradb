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
	"errors"
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
)

func TestStart_WithReplicatorInterval_InvalidIntervalError(t *testing.T) {

	arguments := []string{"--replicator-retry-intervals=garbage"}
	expectedError := "invalid argument \"garbage\" for \"--replicator-retry-intervals\" flag: strconv.Atoi: parsing \"garbage\": invalid syntax"
	test := &integration.Test{
		Actions: []action.Action{
			action.StartWithArgs(arguments, errors.New(expectedError)),
		},
	}

	test.Execute(t)
}

// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/sourcenetwork/testo"
	"github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/defradb/cli/test/action"
	_ "github.com/sourcenetwork/defradb/cli/test/multiplier"
	"github.com/sourcenetwork/defradb/cli/test/state"
)

func init() {
	multiplier.Init("DEFRA_MULTIPLIERS")
}

// Test is a single, self-contained, test.
type Test struct {
	// The test will be skipped if the current active set of multipliers
	// does not contain all of the given multiplier names.
	Includes []multiplier.Name

	// The test will be skipped if the current active set of multipliers
	// contains any of the given multiplier names.
	Excludes []multiplier.Name

	// Actions contains the set of actions that should be
	// executed as part of this test.
	Actions action.Actions
}

func (test *Test) Execute(t testing.TB) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	multiplier.Skip(t, test.Includes, test.Excludes)

	// Prepend a start action if there is not already one present, this saves each test from
	// having to redeclare the same initial action.
	actions := prependStart(test.Actions)

	actions = multiplier.Apply(actions)

	testo.Log(t, actions)

	state := &state.State{
		T:    t,
		Ctx:  ctx,
		Wait: func() {},
	}
	testo.ExecuteS(actions, state)

	state.Wait()
}

func prependStart(actions action.Actions) action.Actions {
	if hasType[*action.StartCli](actions) {
		return actions
	}

	result := make(action.Actions, 1, len(actions)+1)
	result[0] = action.Start()
	result = append(result, actions...)

	return result
}

// hasType returns true if any of the items in the given set are of the given type.
func hasType[TAction any](actions action.Actions) bool {
	for _, action := range actions {
		_, ok := action.(TAction)
		if ok {
			return true
		}
	}

	return false
}

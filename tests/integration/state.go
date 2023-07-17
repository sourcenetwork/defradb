// Copyright 2023 Democratized Data Foundation
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
	"context"
	"testing"
)

type state struct {
	// The test context.
	ctx context.Context

	// The Go Test test state
	t *testing.T

	// The TestCase currently being executed.
	testCase TestCase

	// The type of database currently being tested.
	dbt DatabaseType
}

// newState returns a new fresh state for the given testCase.
func newState(
	ctx context.Context,
	t *testing.T,
	testCase TestCase,
	dbt DatabaseType,
) *state {
	return &state{
		ctx:      ctx,
		t:        t,
		testCase: testCase,
		dbt:      dbt,
	}
}

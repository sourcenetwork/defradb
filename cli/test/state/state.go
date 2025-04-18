// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package state

import (
	"context"
	"testing"
)

type State struct {
	Ctx context.Context
	T   testing.TB
	// Wait must be called at the end of the test execution, to block
	// continuation of the thread until Wait completes.
	//
	// Actions that do stuff that must be waited on before the next test
	// begins (such as mutating global state) should append to this function.
	Wait func()

	// The root directory in which the defra config file should exist.
	RootDir string

	// The base url to the http endpoints.
	Url string

	// The set of available transaction ids.
	Txns []uint64
}

// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

import "testing"

// This file wraps the unit tests from test_node_options.go.
// The purpose is to allow the unit tests to be run by `go test`, which would otherwise have
// problems due to the CGO directives used by the tests.

// Both of the following tests will set each of the possible node options, and then investigate the resulting options
// JSON object, comparing it against the expected values. We do two separate sets of values as tests to ensure that the
// options are actually being applied. With just one set, it's possible that a value not be applied, but that the test not
// catch it due to it also being the default value. This is espeecially possiblee in the case of a default value being
// changed at a later date.
func Test_GetNodeOptions_SetA(t *testing.T) {
	TestGetNodeOptions_SetA(t)
}

func Test_GetNodeOptions_SetB(t *testing.T) {
	TestGetNodeOptions_SetB(t)
}

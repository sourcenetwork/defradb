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
	Ctx     context.Context
	Cancels []context.CancelFunc
	T       testing.TB

	// The root directory in which the defra config file should exist.
	RootDir string

	// The base url to the http endpoints.
	Url string

	// The set of available transaction ids.
	Txns []uint64
}

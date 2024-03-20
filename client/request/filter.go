// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

import "github.com/sourcenetwork/immutable"

const (
	FilterOpOr  = "_or"
	FilterOpAnd = "_and"
	FilterOpNot = "_not"
)

// Filter contains the parsed condition map to be
// run by the Filter Evaluator.
// @todo: Cache filter structure for faster condition
// evaluation.
type Filter struct {
	// parsed filter conditions
	Conditions map[string]any
}

// Filterable is an embeddable struct that hosts a consistent set of properties
// for filtering an aspect of a request.
type Filterable struct {
	// OrderBy is an optional set of conditions used to filter records prior to
	// being processed by the request.
	Filter immutable.Option[Filter]
}

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

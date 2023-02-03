// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

type (
	OrderDirection string

	OrderCondition struct {
		// field may be a compound field statement
		// since the order statement allows ordering on
		// sub objects.
		//
		// Given the statement: {order: {author: {birthday: DESC}}}
		// The field value would be "author.birthday"
		// and the direction would be "DESC"
		Fields    []string
		Direction OrderDirection
	}

	OrderBy struct {
		Conditions []OrderCondition
	}
)

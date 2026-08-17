// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package connor

import (
	"regexp"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// regex is an operator which matches a string against a regular expression.
//
// The pattern is compiled with the standard library, which implements RE2: matching is
// linear in the length of the input and there is no backtracking, so a user-supplied
// pattern cannot make the match run away. The pattern is compiled per call, since the
// filter is evaluated one document at a time and there is nowhere to hold a compiled
// pattern for the length of a query.
func regex(condition, data any) (bool, error) {
	switch d := data.(type) {
	case immutable.Option[string]:
		if !d.HasValue() {
			return condition == nil, nil
		}
		data = d.Value()
	}

	switch cn := condition.(type) {
	case string:
		expr, err := regexp.Compile(cn)
		if err != nil {
			return false, NewErrInvalidRegex(cn, err)
		}
		if d, ok := data.(string); ok {
			return expr.MatchString(d), nil
		}
		return false, nil
	default:
		return false, client.NewErrUnhandledType("condition", cn)
	}
}

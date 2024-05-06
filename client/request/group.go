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

type GroupBy struct {
	Fields []string
}

// Groupable is an embeddable struct that hosts a consistent set of properties
// for grouping an aspect of a request.
type Groupable struct {
	// GroupBy is an optional set of fields for which to group the contents of this
	// request by.
	//
	// If this argument is provided, only fields used to group may be rendered in
	// the immediate child selector.  Additional fields may be selected by using
	// the '_group' selector within the immediate child selector. If an empty set
	// is provided, the restrictions mentioned still apply, although all results
	// will appear within the same group.
	GroupBy immutable.Option[GroupBy]
}

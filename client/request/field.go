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

// Field implements Selection
type Field struct {
	// Name contains the name of the field on it's host object.
	//
	// For example `email` on a `User` collection, or a `_count` aggregate.
	Name string

	// Alias is an optional override for Name, if provided results will be returned
	// from the query using the Alias instead of the Name.
	Alias immutable.Option[string]
}

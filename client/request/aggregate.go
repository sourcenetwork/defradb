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

// Aggregate represents an aggregate operation upon a set of child properties.
//
// Which aggregate this represents (e.g. _count, _avg, etc.) is determined by its
// [Name] property.
type Aggregate struct {
	Field

	// Targets hosts the properties to aggregate.
	//
	// When multiple properties are selected, their values will be gathered into a single set
	// upon which the aggregate will be performed.  For example, if this aggregate represents
	// and average of the Friends.Age and Parents.Age fields, the result will be the average
	// age of all their friends and parents, it will not be an average of their average ages.
	Targets []*AggregateTarget
}

// AggregateTarget represents the target of an [Aggregate].
type AggregateTarget struct {
	Limitable
	Offsetable
	Orderable
	Filterable

	// HostName is the name of the immediate field on the object hosting the aggregate.
	//
	// For example if averaging Friends.Age on the User collection, this property would be
	// "Friends".
	HostName string

	// ChildName is the name of the child field on the object navigated to via [HostName].
	//
	// It is optional, for example when counting the number of Friends on User, or when aggregating
	// scalar arrays, this value will be None.
	//
	// When averaging Friends.Age on the User collection, this property would be
	// "Age".
	ChildName immutable.Option[string]
}

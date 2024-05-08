// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

import "github.com/sourcenetwork/defradb/internal/core"

// An optional child target.
type OptionalChildTarget struct {
	// The field index of this target.
	Index int

	// The name of the target, for example '_sum' or 'Age'.
	Name string

	// If true this child target exists and has been requested.
	//
	// If false, this property is empty and in its default state.
	HasValue bool
}

// The relative target/path from the object hosting an aggregate, to the property to
// be aggregated.
type AggregateTarget struct {
	Targetable

	// The property on the `HostIndex` that this aggregate targets.
	//
	// This may be empty if the aggregate targets a whole collection (e.g. Count),
	// or if `HostIndex` is an inline array.
	ChildTarget OptionalChildTarget
}

// Aggregate represents an aggregate operation definition.
//
// E.g. count, or average. This may have been requested by a consumer, or it may be
// an internal dependency (of for example, another aggregate).
type Aggregate struct {
	Field
	// The mapping of this aggregate's parent/host.
	*core.DocumentMapping

	// The collection of targets that this aggregate will aggregate.
	AggregateTargets []AggregateTarget

	// Any aggregates that this aggregate may dependend on.
	//
	// For example, Average is dependent on a Sum and Count field.
	Dependencies []*Aggregate
}

func (a *Aggregate) CloneTo(index int) Requestable {
	return a.cloneTo(index)
}

func (a *Aggregate) cloneTo(index int) *Aggregate {
	return &Aggregate{
		Field:            *a.Field.cloneTo(index),
		DocumentMapping:  a.DocumentMapping,
		AggregateTargets: a.AggregateTargets,
	}
}

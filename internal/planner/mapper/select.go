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

import (
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/core"
)

// Select represents a request to return data from Defra.
//
// It wraps child Fields belonging to this Select.
type Select struct {

	// The document mapping for this select, describing how items yielded
	// for this select can be accessed and rendered.
	*core.DocumentMapping

	// A commit identifier that can be specified to request data at a given time.
	Cid immutable.Option[string]

	// The name of the collection that this Select selects data from.
	CollectionName string

	// The fields that are to be selected.
	//
	// These can include stuff such as version information, aggregates, and other
	// Selects.
	Fields []Requestable

	// Targeting information used to restrict or format the result.
	Targetable

	// SkipResolve is a flag that indicates that the fields in this Select don't need to be resolved,
	// i.e. it's value doesn't need to be fetched and provided to the user.
	// It is used to avoid resolving related objects if they are used only in a filter and not requested in a response.
	SkipResolve bool
}

func (s *Select) AsTargetable() (*Targetable, bool) {
	return &s.Targetable, true
}

func (s *Select) AsSelect() (*Select, bool) {
	return s, true
}

func (s *Select) CloneTo(index int) Requestable {
	return s.cloneTo(index)
}

func (s *Select) cloneTo(index int) *Select {
	return &Select{
		Targetable:      *s.Targetable.cloneTo(index),
		DocumentMapping: s.DocumentMapping,
		Cid:             s.Cid,
		CollectionName:  s.CollectionName,
		Fields:          s.Fields,
	}
}

func (s *Select) FieldAt(index int) Requestable {
	return fieldAt(s.Fields, index)
}

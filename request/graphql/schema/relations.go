// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/client"
)

// RelationManager keeps track of all the relations that exist
// between schema types
type RelationManager struct {
	relations map[string]*Relation
}

func NewRelationManager() *RelationManager {
	return &RelationManager{
		relations: make(map[string]*Relation),
	}
}

func (rm *RelationManager) GetRelation(name string) (*Relation, error) {
	rel, ok := rm.relations[name]
	if !ok {
		return nil, NewErrRelationNotFound(name)
	}
	return rel, nil
}

// RegisterSingle is used if you only know a single side of the relation
// at a time. It allows you to iteratively, across two calls, build the relation.
// If the relation exists and is finalized, then nothing is done. Returns true
// if nothing is done or the relation is successfully registered.
func (rm *RelationManager) RegisterSingle(
	name string,
	schemaType string,
	schemaField string,
	relType client.RelationType,
) (bool, error) {
	if name == "" {
		return false, client.NewErrUninitializeProperty("RegisterSingle", "name")
	}

	// make sure the relation type is ONLY One or Many, not both
	if relType.IsSet(client.Relation_Type_ONE) == relType.IsSet(client.Relation_Type_MANY) {
		return false, ErrRelationMutlipleTypes
	}

	// make a copy of rel type, one goes to the relation.relType, and the other goes into the []types.
	// We need to clear the Primary bit on the relation.relType so we make a copy
	rt := relType
	rt &^= client.Relation_Type_Primary // clear the primary bit

	rel, ok := rm.relations[name]
	if !ok {
		// If a relation doesn't exist then make one.
		rm.relations[name] = &Relation{
			name:        name,
			relType:     rt,
			types:       []client.RelationType{relType},
			schemaTypes: []string{schemaType},
			fields:      []string{schemaField},
		}
		return true, nil
	}

	if !rel.finalized {
		// If a  relation exists, and is not finalized, then finalizing it.

		// handle relationType, needs to be either One-to-One, One-to-Many, Many-to-Many.
		if rel.relType.IsSet(client.Relation_Type_ONE) {
			if relType.IsSet(client.Relation_Type_ONE) { // One-to-One
				rel.relType = client.Relation_Type_ONEONE
			} else if relType.IsSet(client.Relation_Type_MANY) {
				rel.relType = client.Relation_Type_ONEMANY
			}
		} else { // many
			if relType.IsSet(client.Relation_Type_ONE) {
				rel.relType = client.Relation_Type_ONEMANY
			}
		}

		rel.types = append(rel.types, relType)
		rel.schemaTypes = append(rel.schemaTypes, schemaType)
		rel.fields = append(rel.fields, schemaField)

		if err := rel.finalize(); err != nil {
			return false, err
		}
		rm.relations[name] = rel
	}

	return true, nil
}

type Relation struct {
	name        string
	relType     client.RelationType
	types       []client.RelationType
	schemaTypes []string
	fields      []string

	// finalized indicates if we've properly
	// updated both sides of the relation
	finalized bool
}

func (r *Relation) finalize() error {
	// make sure all the types/fields are set
	if len(r.types) != 2 || len(r.schemaTypes) != 2 || len(r.fields) != 2 {
		return ErrRelationMissingTypes
	}

	// make sure its one of One-to-One, One-to-Many
	if !r.relType.IsSet(client.Relation_Type_ONEONE) &&
		!r.relType.IsSet(client.Relation_Type_ONEMANY) {
		return ErrRelationInvalidType
	}

	// make sure we have a primary set if its a one-to-one
	if IsOneToOne(r.relType) {
		t1, t2 := r.types[0], r.types[1]
		aBit := t1 & t2
		xBit := t1 ^ t2

		// both types have primary set
		if aBit.IsSet(client.Relation_Type_Primary) {
			return ErrMultipleRelationPrimaries
		} else if !xBit.IsSet(client.Relation_Type_Primary) {
			// neither type has primary set, auto add to
			// lexicographically first one by schema type name
			if strings.Compare(r.schemaTypes[0], r.schemaTypes[1]) < 1 {
				r.types[1] = r.types[1] | client.Relation_Type_Primary
			} else {
				r.types[0] = r.types[0] | client.Relation_Type_Primary
			}
		}
	} else if IsOneToMany(r.relType) { // if its a one-to-many, set the one side as primary
		if IsOne(r.types[0]) {
			r.types[0] |= client.Relation_Type_Primary  // set primary on one
			r.types[1] &^= client.Relation_Type_Primary // clear primary on many
		} else {
			r.types[1] |= client.Relation_Type_Primary  // set primary on one
			r.types[0] &^= client.Relation_Type_Primary // clear primary on many
		}
	}

	r.finalized = true
	return nil
}

// Kind returns what type of relation it is
func (r Relation) Kind() client.RelationType {
	return r.relType
}

func (r Relation) GetField(schemaType string, field string) (string, client.RelationType, bool) {
	for i, f := range r.fields {
		if f == field && r.schemaTypes[i] == schemaType {
			return f, r.types[i], true
		}
	}
	return "", client.RelationType(0), false
}

func genRelationName(t1, t2 string) (string, error) {
	if t1 == "" || t2 == "" {
		return "", client.NewErrUninitializeProperty("genRelationName", "relation types")
	}
	t1 = strings.ToLower(t1)
	t2 = strings.ToLower(t2)

	if i := strings.Compare(t1, t2); i < 0 {
		return fmt.Sprintf("%s_%s", t1, t2), nil
	}
	return fmt.Sprintf("%s_%s", t2, t1), nil
}

// IsOne returns true if the Relation_ONE bit is set
func IsOne(fieldmeta client.RelationType) bool {
	return fieldmeta.IsSet(client.Relation_Type_ONE)
}

// IsOneToOne returns true if the Relation_ONEONE bit is set
func IsOneToOne(fieldmeta client.RelationType) bool {
	return fieldmeta.IsSet(client.Relation_Type_ONEONE)
}

// IsOneToMany returns true if the Relation_ONEMANY is set
func IsOneToMany(fieldmeta client.RelationType) bool {
	return fieldmeta.IsSet(client.Relation_Type_ONEMANY)
}

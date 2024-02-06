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

// relationType describes the type of relation between two types.
type relationType uint8

const (
	relation_Type_ONE     relationType = 1   // 0b0000 0001
	relation_Type_MANY    relationType = 2   // 0b0000 0010
	relation_Type_Primary relationType = 128 // 0b1000 0000 Primary reference entity on relation
)

// IsSet returns true if the target relation type is set.
func (m relationType) isSet(target relationType) bool {
	return m&target > 0
}

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
	relType relationType,
) (bool, error) {
	if name == "" {
		return false, client.NewErrUninitializeProperty("RegisterSingle", "name")
	}

	rel, ok := rm.relations[name]
	if !ok {
		// If a relation doesn't exist then make one.
		rm.relations[name] = &Relation{
			name:        name,
			types:       []relationType{relType},
			schemaTypes: []string{schemaType},
			fields:      []string{schemaField},
		}
		return true, nil
	}

	if !rel.finalized {
		// If a  relation exists, and is not finalized, then finalizing it.
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
	types       []relationType
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

	if isOne(r.types[0]) && isMany(r.types[1]) {
		r.types[0] |= relation_Type_Primary  // set primary on one
		r.types[1] &^= relation_Type_Primary // clear primary on many
	} else if isOne(r.types[1]) && isMany(r.types[0]) {
		r.types[1] |= relation_Type_Primary  // set primary on one
		r.types[0] &^= relation_Type_Primary // clear primary on many
	} else if isOne(r.types[1]) && isOne(r.types[0]) {
		t1, t2 := r.types[0], r.types[1]
		aBit := t1 & t2
		xBit := t1 ^ t2

		// both types have primary set
		if aBit.isSet(relation_Type_Primary) {
			return ErrMultipleRelationPrimaries
		} else if !xBit.isSet(relation_Type_Primary) {
			// neither type has primary set, auto add to
			// lexicographically first one by schema type name
			if strings.Compare(r.schemaTypes[0], r.schemaTypes[1]) < 1 {
				r.types[1] = r.types[1] | relation_Type_Primary
			} else {
				r.types[0] = r.types[0] | relation_Type_Primary
			}
		}
	}

	r.finalized = true
	return nil
}

func (r Relation) getField(schemaType string, field string) (string, relationType, bool) {
	for i, f := range r.fields {
		if f == field && r.schemaTypes[i] == schemaType {
			return f, r.types[i], true
		}
	}
	return "", relationType(0), false
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

// isOne returns true if the Relation_ONE bit is set
func isOne(fieldmeta relationType) bool {
	return fieldmeta.isSet(relation_Type_ONE)
}

// isMany returns true if the Relation_ONE bit is set
func isMany(fieldmeta relationType) bool {
	return fieldmeta.isSet(relation_Type_MANY)
}

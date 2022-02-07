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
	"errors"
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/db/base"
)

// type uint8 uint8

// const (
// 	uint8_One uint8 = 1 << iota
// 	uint8_Many
// 	uint8_OneToOne
// 	uint8_OneToMany
// 	uint8_ManyToMany

// 	uint8_Primary = 1 << 7
// )

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

func (rm *RelationManager) GetRelations() {}

func (rm *RelationManager) GetRelation(name string) (*Relation, error) {
	rel, ok := rm.relations[name]
	if !ok {
		return nil, errors.New("No relation found")
	}
	return rel, nil
}

func (rm *RelationManager) GetRelationByDescription(field, schemaType, objectType string) *Relation {
	for _, rel := range rm.relations {
		t1, t2 := rel.schemaTypes[0], rel.schemaTypes[1]
		if (t1 == schemaType && t2 == objectType) ||
			(t1 == objectType && t2 == schemaType) {
			f1, f2 := rel.fields[0], rel.fields[1]
			if field == f1 || field == f2 {
				return rel
			}
		}
	}

	return nil
}

func (rm *RelationManager) NumRelations() int {
	return len(rm.relations)
}

// validate ensures that all the relations are finalized.
// It returns any relations that aren't. Returns true if
// everything is valid
func (rm *RelationManager) validate() ([]*Relation, bool) {
	unfinalized := make([]*Relation, 0)
	for _, rel := range rm.relations {
		if !rel.finalized {
			unfinalized = append(unfinalized, rel)
		}
	}
	if len(unfinalized) > 0 {
		return unfinalized, false
	}
	return nil, true
}

func (rm *RelationManager) Exists(name string) bool {
	_, exists := rm.relations[name]
	return exists
}

// RegisterSingle is used if you only know a single side of the relation
// at a time. It allows you to iteratively, across two calls, build the relation.
// It will fail if you call it again on a relation that has been registered AND finalized
func (rm *RelationManager) RegisterSingle(name, schemaType, schemaField string, relType uint8) (bool, error) {
	if name == "" {
		return false, errors.New("Relation name must be non empty")
	}

	// make sure the relation type is ONLY One or Many, not both
	if base.IsSet(relType, base.Meta_Relation_ONE) == base.IsSet(relType, base.Meta_Relation_MANY) {
		return false, errors.New("Relation type can only be either One or Many, not both")
	}

	// make a copy of rel type, one goes to the relation.relType, and the other goes into the []types.
	// We need to clear the Primary bit on the relation.relType so we make a copy
	rt := relType
	rt &^= base.Meta_Relation_Primary // clear the primary bit

	rel, ok := rm.relations[name]
	if !ok {
		rel = &Relation{
			name:        name,
			relType:     rt,
			types:       []uint8{relType},
			schemaTypes: []string{schemaType},
			fields:      []string{schemaField},
		}
	} else if rel.finalized {
		return false, errors.New("Cannot update a relation that is already finalized")
	} else {
		// relation exists, and is not finalized

		// handle relationType, needs to be either One-to-One, One-to-Many, Many-to-Many
		// one
		if base.IsSet(rel.relType, base.Meta_Relation_ONE) {
			if base.IsSet(relType, base.Meta_Relation_ONE) { // One-to-One
				rel.relType = base.Meta_Relation_ONEONE
			} else if base.IsSet(relType, base.Meta_Relation_MANY) {
				rel.relType = base.Meta_Relation_ONEMANY
			}
		} else { // many
			if base.IsSet(relType, base.Meta_Relation_ONE) {
				rel.relType = base.Meta_Relation_ONEMANY
			} else if base.IsSet(relType, base.Meta_Relation_MANY) {
				rel.relType = base.Meta_Relation_MANYMANY
			}
		}

		rel.types = append(rel.types, relType)
		rel.schemaTypes = append(rel.schemaTypes, schemaType)
		rel.fields = append(rel.fields, schemaField)
		if err := rel.finalize(); err != nil {
			return false, err
		}

	}

	rm.relations[name] = rel
	return true, nil
}

// RegisterRelation adds a new relation to the RelationManager
// if it doesn't already exist.
func (rm *RelationManager) RegisterOneToOne(name, primaryType, primaryField, secondaryType, secondaryField string) (bool, error) {
	return rm.register(nil)
}

func (rm *RelationManager) RegisterOneToMany(name, oneType, oneField, manyType, manyField string) (bool, error) {
	return rm.register(nil)
}

func (rm *RelationManager) RegisterManyToMany(name, type1, type2 string) (bool, error) {
	return rm.register(nil)
}

func (rm *RelationManager) register(rel *Relation) (bool, error) {
	return true, nil
}

type Relation struct {
	name        string
	relType     uint8
	types       []uint8
	schemaTypes []string // []gql.Object??
	fields      []string //

	// finalized indicates if we've properly
	// updated both sides of the relation
	finalized bool
}

func (r *Relation) finalize() error {
	// make sure all the types/fields are set
	if len(r.types) != 2 || len(r.schemaTypes) != 2 || len(r.fields) != 2 {
		return errors.New("Relation is missing its defined types and fields")
	}

	// make sure its one of One-to-One, One-to-Many, Many-to-Many
	if !base.IsSet(r.relType, base.Meta_Relation_ONEONE) &&
		!base.IsSet(r.relType, base.Meta_Relation_ONEMANY) &&
		!base.IsSet(r.relType, base.Meta_Relation_MANYMANY) {
		return errors.New("Relation has an invalid type to be finalize")
	}

	// make sure we have a primary set if its a one-to-one or many-to-many
	if IsOneToOne(r.relType) || IsManyToMany(r.relType) {
		t1, t2 := r.types[0], r.types[1]
		aBit := t1 & t2
		xBit := t1 ^ t2

		// both types have primary set
		if base.IsSet(aBit, base.Meta_Relation_Primary) {
			return errors.New("relation can only have a single field set as primary")
		} else if !base.IsSet(xBit, base.Meta_Relation_Primary) {
			// neither type has primary set, auto add to
			// lexicographically first one by schema type name
			if strings.Compare(r.schemaTypes[0], r.schemaTypes[1]) < 1 {
				r.types[1] = r.types[1] | base.Meta_Relation_Primary
			} else {
				r.types[0] = r.types[0] | base.Meta_Relation_Primary
			}
		}
	} else if IsOneToMany(r.relType) { // if its a one-to-many, set the one side as primary
		if IsOne(r.types[0]) {
			r.types[0] |= base.Meta_Relation_Primary  // set primary on one
			r.types[1] &^= base.Meta_Relation_Primary // clear primary on many
		} else {
			r.types[1] |= base.Meta_Relation_Primary  // set primary on one
			r.types[0] &^= base.Meta_Relation_Primary // clear primary on many
		}
	}

	r.finalized = true
	return nil
}

func (r Relation) GetFields() []string {
	return r.fields
}

// Type returns what kind of relation it is
func (r Relation) Kind() uint8 {
	return r.relType
}

func (r Relation) Valid() bool {
	return r.finalized
}

// SchemaTypeIsPrimary returns true if the provided type of the relation
// is the primary type. Only one-to-one and one-to-many have primaries.
func (r Relation) SchemaTypeIsPrimary(t string) bool {
	i, ok := r.schemaTypeExists(t)
	if !ok {
		return false
	}

	relType := r.types[i]
	return base.IsSet(relType, base.Meta_Relation_Primary)
}

// SchemaTypeIsOne returns true if the provided type of the relation
// is the primary type. Only one-to-one and one-to-many have primaries.
func (r Relation) SchemaTypeIsOne(t string) bool {
	i, ok := r.schemaTypeExists(t)
	if !ok {
		return false
	}

	relType := r.types[i]
	return base.IsSet(relType, base.Meta_Relation_ONE)
}

// SchemaTypeIsMany returns true if the provided type of the relation
// is the primary type. Only one-to-one and one-to-many have primaries.
func (r Relation) SchemaTypeIsMany(t string) bool {
	i, ok := r.schemaTypeExists(t)
	if !ok {
		return false
	}

	relType := r.types[i]
	return base.IsSet(relType, base.Meta_Relation_MANY)
}

func (r Relation) schemaTypeExists(t string) (int, bool) {
	for i, schemaType := range r.schemaTypes {
		if t == schemaType {
			return i, true
		}
	}
	return -1, false
}

func (r Relation) GetField(schemaType string, field string) (string, uint8, bool) {
	for i, f := range r.fields {
		if f == field && r.schemaTypes[i] == schemaType {
			return f, r.types[i], true
		}
	}
	return "", uint8(0), false
}

func (r Relation) GetFieldFromSchemaType(schemaType string) (string, uint8, bool) {
	for i, s := range r.schemaTypes {
		if s == schemaType {
			return r.fields[1-i], r.types[1-i], true
		}
	}
	return "", uint8(0), false
}

func genRelationName(t1, t2 string) (string, error) {
	if t1 == "" || t2 == "" {
		return "", errors.New("relation types cannot be empty")
	}
	t1 = strings.ToLower(t1)
	t2 = strings.ToLower(t2)

	if i := strings.Compare(t1, t2); i < 0 {
		return fmt.Sprintf("%s_%s", t1, t2), nil
	}
	return fmt.Sprintf("%s_%s", t2, t1), nil

}

// IsPrimary returns true if the Relation_Primary bit is set
func IsPrimary(fieldmeta uint8) bool {
	return base.IsSet(fieldmeta, base.Meta_Relation_Primary)
}

// IsOne returns true if the Relation_ONE bit is set
func IsOne(fieldmeta uint8) bool {
	return base.IsSet(fieldmeta, base.Meta_Relation_ONE)
}

// IsOneToOne returns true if the Relation_ONEONE bit is set
func IsOneToOne(fieldmeta uint8) bool {
	return base.IsSet(fieldmeta, base.Meta_Relation_ONEONE)
}

// IsMany returns true if the Relation_MANY bit is set
func IsMany(fieldmeta uint8) bool {
	return base.IsSet(fieldmeta, base.Meta_Relation_MANY)
}

// IsOneToMany returns true if the Relation_ONEMANY is set
func IsOneToMany(fieldmeta uint8) bool {
	return base.IsSet(fieldmeta, base.Meta_Relation_ONEMANY)
}

// IsManyToMany returns true if the Relation_MANYMANY bit is set
func IsManyToMany(fieldmeta uint8) bool {
	return base.IsSet(fieldmeta, base.Meta_Relation_MANYMANY)
}

/* Example usage

rm := NewRelationManager()

type book {
	name: String
	rating: Float
	author: author
}

type author {
	name: String
	age: Int
	verified: Boolean
	published: [book]
}

Relation names are autogenerated. They are the combination of each type
in the relation, sorted alphabetically.

Relations:
name: author_book | related types: author, book | fields: author, published (index same as types) type: one-to-many.

rm.GetRelations(type) returns all the relations containing that type
rel := rm.GetRelation(name) returns the exact relation (if it exists) between those types
rel.IsPrimary(type) => bool, error
rel.IsOne(type) => bool, error
rel.IsMany(type) => bool, error
rel.Type() OneToOne | OneToMany | ManyToOne? | ManyToMany

*/

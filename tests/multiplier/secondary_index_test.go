// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package multiplier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/tests/action"
)

func TestHasIndexActions_WithCreateIndex_ReturnsTrue(t *testing.T) {
	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String }"},
		&action.CreateIndex{CollectionID: 0, FieldName: "name"},
	}

	assert.True(t, hasIndexActions(actions))
}

func TestHasIndexActions_WithDropIndex_ReturnsTrue(t *testing.T) {
	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String }"},
		&action.DropIndex{CollectionID: 0, IndexName: "User_name_idx"},
	}

	assert.True(t, hasIndexActions(actions))
}

func TestHasIndexActions_WithGetIndexes_ReturnsTrue(t *testing.T) {
	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String }"},
		&action.GetIndexes{CollectionID: 0},
	}

	assert.True(t, hasIndexActions(actions))
}

func TestHasIndexActions_WithNoIndexActions_ReturnsFalse(t *testing.T) {
	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String }"},
	}

	assert.False(t, hasIndexActions(actions))
}

func TestHasIndexDirective_WithIndexDirective_ReturnsTrue(t *testing.T) {
	schema := `type User { name: String @index }`
	assert.True(t, hasIndexDirective(schema))
}

func TestHasIndexDirective_WithUniqueIndexDirective_ReturnsTrue(t *testing.T) {
	schema := `type User { email: String @index(unique: true) }`
	assert.True(t, hasIndexDirective(schema))
}

func TestHasIndexDirective_WithNoDirective_ReturnsFalse(t *testing.T) {
	schema := `type User { name: String }`
	assert.False(t, hasIndexDirective(schema))
}

func TestAddIndexesToSchema_WithSimpleField_AddsIndex(t *testing.T) {
	schema := `type User {
	name: String
}`
	expected := `type User {
	name: String @index
}`
	assert.Equal(t, expected, addIndexesToSchema(schema))
}

func TestAddIndexesToSchema_WithMultipleFields_AddsIndexToAll(t *testing.T) {
	schema := `type User {
	name: String
	age: Int
	active: Boolean
}`
	expected := `type User {
	name: String @index
	age: Int @index
	active: Boolean @index
}`
	assert.Equal(t, expected, addIndexesToSchema(schema))
}

func TestAddIndexesToSchema_WithAllScalarTypes_AddsIndexToAll(t *testing.T) {
	schema := `type User {
	name: String
	age: Int
	score: Float
	active: Boolean
	created: DateTime
	docID: ID
	custom: JSON
}`
	expected := `type User {
	name: String @index
	age: Int @index
	score: Float @index
	active: Boolean @index
	created: DateTime @index
	docID: ID @index
	custom: JSON @index
}`
	assert.Equal(t, expected, addIndexesToSchema(schema))
}

func TestAddIndexesToSchema_WithOtherDirectives_AddsIndexBeforeDirective(t *testing.T) {
	schema := `type User {
	name: String @crdt(type: lww)
	points: Float @crdt(type: pcounter)
	active: Boolean @default(bool: true)
}`
	result := addIndexesToSchema(schema)

	assert.Contains(t, result, "name: String @index @crdt(type: lww)")
	assert.Contains(t, result, "points: Float @index @crdt(type: pcounter)")
	assert.Contains(t, result, "active: Boolean @index @default(bool: true)")
}

func TestAddIndexesToSchema_WithNonNullFields_AddsIndex(t *testing.T) {
	schema := `type User {
	name: String!
	age: Int!
	score: Float!
}`
	result := addIndexesToSchema(schema)

	assert.Contains(t, result, "name: String! @index")
	assert.Contains(t, result, "age: Int! @index")
	assert.Contains(t, result, "score: Float! @index")
}

func TestAddIndexesToSchema_WithNonNullAndDirectives_AddsIndex(t *testing.T) {
	schema := `type User {
	name: String! @crdt(type: lww)
	age: Int! @default(int: 0)
}`
	result := addIndexesToSchema(schema)

	assert.Contains(t, result, "name: String! @index @crdt(type: lww)")
	assert.Contains(t, result, "age: Int! @index @default(int: 0)")
}

func TestAddIndexesToSchema_WithArrayFields_AddsIndex(t *testing.T) {
	schema := `type User {
	names: [String]
	numbers: [Int!]
	scores: [Float]!
	flags: [Boolean!]!
}`
	result := addIndexesToSchema(schema)

	assert.Contains(t, result, "names: [String] @index")
	assert.Contains(t, result, "numbers: [Int!] @index")
	assert.Contains(t, result, "scores: [Float]! @index")
	assert.Contains(t, result, "flags: [Boolean!]! @index")
}

func TestAddIndexesToSchema_WithArrayAndDirectives_AddsIndex(t *testing.T) {
	schema := `type User {
	tags: [String] @crdt(type: lww)
	numbers: [Int!] @default(int: [])
}`
	result := addIndexesToSchema(schema)

	assert.Contains(t, result, "tags: [String] @index @crdt(type: lww)")
	assert.Contains(t, result, "numbers: [Int!] @index @default(int: [])")
}

func TestAddIndexesToSchema_WithOneToManyRelation_IndexesManySide(t *testing.T) {
	schema := `type User {
	name: String
	devices: [Device]
}

type Device {
	model: String
	owner: User
}`
	result := addIndexesToSchema(schema)

	// Scalar fields get indexed
	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "model: String @index")

	// Array relation (one side) does not get indexed
	assert.Contains(t, result, "devices: [Device]")
	assert.NotContains(t, result, "[Device] @index")

	// Single relation (many side, holds foreign key) gets indexed
	assert.Contains(t, result, "owner: User @index")
}

func TestAddIndexesToSchema_WithOneToOnePrimaryOnSecondary_IndexesPrimarySide(t *testing.T) {
	// @primary is on Address.user, so Address holds the foreign key (_userID)
	schema := `type User {
	name: String
	address: Address
}

type Address {
	city: String
	user: User @primary
}`
	result := addIndexesToSchema(schema)

	// Scalar fields get indexed
	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "city: String @index")

	// Primary relation (with @primary, holds foreign key) gets indexed
	assert.Contains(t, result, "user: User @index @primary")

	// Secondary relation (no @primary, no foreign key) does not get indexed
	assert.Contains(t, result, "address: Address")
	assert.NotContains(t, result, "address: Address @index")
}

func TestAddIndexesToSchema_WithOneToOnePrimaryOnPrimary_IndexesPrimarySide(t *testing.T) {
	// @primary is on User.address, so User holds the foreign key (_addressID)
	schema := `type User {
	name: String
	address: Address @primary
}

type Address {
	city: String
	user: User
}`
	result := addIndexesToSchema(schema)

	// Scalar fields get indexed
	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "city: String @index")

	// Primary relation (with @primary, holds foreign key) gets indexed
	assert.Contains(t, result, "address: Address @index @primary")

	// Secondary relation (no @primary, no foreign key) does not get indexed
	assert.Contains(t, result, "user: User")
	assert.NotContains(t, result, "user: User @index")
}

func TestAddIndexesToSchema_WithMultipleRelations_IndexesAllManySides(t *testing.T) {
	// Device has two foreign keys: _ownerID and _manufacturerID
	schema := `type User {
	name: String
	devices: [Device]
}

type Device {
	model: String
	owner: User
	manufacturer: Manufacturer
}

type Manufacturer {
	name: String
	devices: [Device]
}`
	result := addIndexesToSchema(schema)

	// Scalar fields get indexed
	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "model: String @index")

	// Array relations (one side, no foreign key) do not get indexed
	assert.NotContains(t, result, "[Device] @index")

	// Both single relations on Device (many side, hold foreign keys) get indexed
	assert.Contains(t, result, "owner: User @index")
	assert.Contains(t, result, "manufacturer: Manufacturer @index")
}

func TestAddIndexesToSchema_WithVariousFormatting_PreservesWhitespace(t *testing.T) {
	schema := `type User {
	name:    String
	age:Int
}`
	result := addIndexesToSchema(schema)

	assert.Contains(t, result, "name:    String @index")
	assert.Contains(t, result, "age:Int @index")
}

func TestApply_WithIndexActions_StillModifiesSchema(t *testing.T) {
	// Note: Apply does not check for index actions - that's ShouldSkip's job.
	// The test framework calls ShouldSkip before Apply.
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String }"},
		&action.CreateIndex{CollectionID: 0, FieldName: "name"},
	}

	result := m.Apply(actions)

	schemaAdd := result[0].(*action.AddSchema)
	assert.Contains(t, schemaAdd.Schema, "@index")

	createIndex := result[1].(*action.CreateIndex)
	assert.Equal(t, 0, createIndex.CollectionID)
	assert.Equal(t, "name", createIndex.FieldName)
}

func TestApply_WithIndexDirective_ReturnsUnchanged(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String @index }"},
	}

	result := m.Apply(actions)

	assert.Equal(t, actions, result)
}

func TestApply_WithoutIndex_ModifiesSchema(t *testing.T) {
	m := &secondaryIndex{}

	original := `type User {
	name: String
	age: Int
}`

	actions := action.Actions{
		&action.AddSchema{Schema: original},
	}

	result := m.Apply(actions)

	assert.NotEqual(t, actions, result)

	schemaAdd := result[0].(*action.AddSchema)
	assert.Contains(t, schemaAdd.Schema, "name: String @index")
	assert.Contains(t, schemaAdd.Schema, "age: Int @index")
}

func TestName_ReturnsSecondaryIndex(t *testing.T) {
	m := &secondaryIndex{}
	assert.Equal(t, SecondaryIndex, m.Name())
	assert.Equal(t, Name("secondary-index"), m.Name())
}

func TestShouldSkip_WithIndexActions_ReturnsTrue(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String }"},
		&action.CreateIndex{CollectionID: 0, FieldName: "name"},
	}

	assert.True(t, m.ShouldSkip(actions))
}

func TestShouldSkip_WithIndexDirective_ReturnsTrue(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String @index }"},
	}

	assert.True(t, m.ShouldSkip(actions))
}

func TestShouldSkip_WithUniqueIndexDirective_ReturnsTrue(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddSchema{Schema: "type User { email: String @index(unique: true) }"},
	}

	assert.True(t, m.ShouldSkip(actions))
}

func TestShouldSkip_WithNoIndex_ReturnsFalse(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddSchema{Schema: "type User { name: String }"},
	}

	assert.False(t, m.ShouldSkip(actions))
}

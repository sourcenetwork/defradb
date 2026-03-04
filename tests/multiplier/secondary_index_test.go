// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package multiplier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/tests/action"
)

func TestHasIndexActions_WithNewIndex_ReturnsTrue(t *testing.T) {
	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.NewIndex{CollectionID: 0, FieldName: "name"},
	}

	assert.True(t, hasIndexActions(actions))
}

func TestHasIndexActions_WithDeleteIndex_ReturnsTrue(t *testing.T) {
	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.DeleteIndex{CollectionID: 0, IndexName: "User_name_idx"},
	}

	assert.True(t, hasIndexActions(actions))
}

func TestHasIndexActions_WithListIndexes_ReturnsTrue(t *testing.T) {
	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.ListIndexes{CollectionID: 0},
	}

	assert.True(t, hasIndexActions(actions))
}

func TestHasIndexActions_WithNoIndexActions_ReturnsFalse(t *testing.T) {
	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
	}

	assert.False(t, hasIndexActions(actions))
}

func TestHasIndexDirective_WithIndexDirective_ReturnsTrue(t *testing.T) {
	sdl := `type User { name: String @index }`
	assert.True(t, hasIndexDirective(sdl))
}

func TestHasIndexDirective_WithUniqueIndexDirective_ReturnsTrue(t *testing.T) {
	sdl := `type User { email: String @index(unique: true) }`
	assert.True(t, hasIndexDirective(sdl))
}

func TestHasIndexDirective_WithNoDirective_ReturnsFalse(t *testing.T) {
	sdl := `type User { name: String }`
	assert.False(t, hasIndexDirective(sdl))
}

func TestNewIndexesToSDL_WithSimpleField_AddsIndex(t *testing.T) {
	sdl := `type User {
	name: String
}`
	expected := `type User {
	name: String @index
}`
	assert.Equal(t, expected, addIndexesToSDL(sdl))
}

func TestNewIndexesToSDL_WithMultipleFields_AddsIndexToAll(t *testing.T) {
	sdl := `type User {
	name: String
	age: Int
	active: Boolean
}`
	expected := `type User {
	name: String @index
	age: Int @index
	active: Boolean @index
}`
	assert.Equal(t, expected, addIndexesToSDL(sdl))
}

func TestNewIndexesToSDL_WithAllScalarTypes_AddsIndexToAll(t *testing.T) {
	sdl := `type User {
	name: String
	age: Int
	score: Float
	points: Float32
	points2: Float64
	active: Boolean
	created: DateTime
	docID: ID
	custom: JSON
}`
	expected := `type User {
	name: String @index
	age: Int @index
	score: Float @index
	points: Float32 @index
	points2: Float64 @index
	active: Boolean @index
	created: DateTime @index
	docID: ID @index
	custom: JSON @index
}`
	assert.Equal(t, expected, addIndexesToSDL(sdl))
}

func TestNewIndexesToSDL_WithOtherDirectives_AddsIndexBeforeDirective(t *testing.T) {
	sdl := `type User {
	name: String @crdt(type: lww)
	points: Float @crdt(type: pcounter)
	active: Boolean @default(bool: true)
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index @crdt(type: lww)")
	assert.Contains(t, result, "points: Float @index @crdt(type: pcounter)")
	assert.Contains(t, result, "active: Boolean @index @default(bool: true)")
}

func TestNewIndexesToSDL_WithNonNullFields_AddsIndex(t *testing.T) {
	sdl := `type User {
	name: String!
	age: Int!
	score: Float!
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String! @index")
	assert.Contains(t, result, "age: Int! @index")
	assert.Contains(t, result, "score: Float! @index")
}

func TestNewIndexesToSDL_WithNonNullAndDirectives_AddsIndex(t *testing.T) {
	sdl := `type User {
	name: String! @crdt(type: lww)
	age: Int! @default(int: 0)
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String! @index @crdt(type: lww)")
	assert.Contains(t, result, "age: Int! @index @default(int: 0)")
}

func TestNewIndexesToSDL_WithArrayFields_AddsIndex(t *testing.T) {
	sdl := `type User {
	names: [String]
	numbers: [Int!]
	scores: [Float]!
	flags: [Boolean!]!
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "names: [String] @index")
	assert.Contains(t, result, "numbers: [Int!] @index")
	assert.Contains(t, result, "scores: [Float]! @index")
	assert.Contains(t, result, "flags: [Boolean!]! @index")
}

func TestNewIndexesToSDL_WithArrayAndDirectives_AddsIndex(t *testing.T) {
	sdl := `type User {
	tags: [String] @crdt(type: lww)
	numbers: [Int!] @default(int: [])
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "tags: [String] @index @crdt(type: lww)")
	assert.Contains(t, result, "numbers: [Int!] @index @default(int: [])")
}

func TestNewIndexesToSDL_WithOneToManyRelation_IndexesManySide(t *testing.T) {
	sdl := `type User {
	name: String
	devices: [Device]
}

type Device {
	model: String
	owner: User
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "model: String @index")

	assert.Contains(t, result, "devices: [Device]")
	assert.NotContains(t, result, "[Device] @index")

	assert.Contains(t, result, "owner: User @index")
}

func TestNewIndexesToSDL_WithNonNullRelation_IndexesManySide(t *testing.T) {
	sdl := `type User {
	name: String
	devices: [Device]
}

type Device {
	model: String
	owner: User!
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "model: String @index")
	assert.Contains(t, result, "owner: User! @index")
}

func TestNewIndexesToSDL_WithOneToOne_DoesNotAddIndex(t *testing.T) {
	// One-to-one relations are NOT indexed because DefraDB automatically
	// creates a unique index to maintain the one-to-one invariant
	sdl := `type User {
	name: String
	address: Address
}

type Address {
	city: String
	user: User @primary
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "city: String @index")

	assert.Contains(t, result, "user: User @primary")
	assert.NotContains(t, result, "user: User @index")
	assert.Contains(t, result, "address: Address")
	assert.NotContains(t, result, "address: Address @index")
}

func TestNewIndexesToSDL_WithNonNullOneToOne_DoesNotAddIndex(t *testing.T) {
	sdl := `type User {
	name: String
	address: Address!
}

type Address {
	city: String
	user: User! @primary
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "city: String @index")

	assert.Contains(t, result, "user: User! @primary")
	assert.NotContains(t, result, "user: User! @index")
	assert.Contains(t, result, "address: Address!")
	assert.NotContains(t, result, "address: Address! @index")
}

func TestNewIndexesToSDL_WithExplicitFKFieldForOneToOne_DoesNotIndex(t *testing.T) {
	sdl := `type Book {
	name: String
	_authorID: Int
	author: Author @primary
}

type Author {
	name: String
	published: Book
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "_authorID: Int")
	assert.NotContains(t, result, "_authorID: Int @index")
	assert.Contains(t, result, "author: Author @primary")
	assert.NotContains(t, result, "author: Author @index")
	assert.Contains(t, result, "published: Book")
	assert.NotContains(t, result, "published: Book @index")
}

func TestNewIndexesToSDL_WithExplicitFKFieldForOneToMany_IndexesFKField(t *testing.T) {
	sdl := `type User {
	name: String
	devices: [Device]
}

type Device {
	model: String
	_ownerID: String
	owner: User
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "model: String @index")
	assert.Contains(t, result, "_ownerID: String @index")
	assert.Contains(t, result, "owner: User @index")
}

func TestNewIndexesToSDL_WithMultipleRelations_IndexesAllManySides(t *testing.T) {
	sdl := `type User {
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
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "model: String @index")

	assert.NotContains(t, result, "[Device] @index")

	assert.Contains(t, result, "owner: User @index")
	assert.Contains(t, result, "manufacturer: Manufacturer @index")
}

func TestNewIndexesToSDL_WithSingleSelfReference_IndexesSelfReference(t *testing.T) {
	sdl := `type User {
	name: String
	boss: User
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "boss: User @index")
}

func TestNewIndexesToSDL_WithOneToOneSelfReference_DoesNotIndex(t *testing.T) {
	sdl := `type User {
	name: String
	boss: User @primary
	underling: User
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "boss: User @primary")
	assert.NotContains(t, result, "boss: User @index")
	assert.Contains(t, result, "underling: User")
	assert.NotContains(t, result, "underling: User @index")
}

func TestNewIndexesToSDL_WithRelationDirective_DoesNotAddIndex(t *testing.T) {
	sdl := `type User {
	hosts: Dog @primary @relation(name:"hosts")
	walks: Dog @relation(name:"walkies")
}

type Dog {
	host: User @relation(name:"hosts")
	walker: User @primary @relation(name:"walkies")
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "hosts: Dog @primary @relation(name:\"hosts\")")
	assert.NotContains(t, result, "hosts: Dog @index")
	assert.Contains(t, result, "walker: User @primary @relation(name:\"walkies\")")
	assert.NotContains(t, result, "walker: User @index")

	assert.Contains(t, result, "walks: Dog @relation(name:\"walkies\")")
	assert.NotContains(t, result, "walks: Dog @index")
	assert.Contains(t, result, "host: User @relation(name:\"hosts\")")
	assert.NotContains(t, result, "host: User @index")
}

func TestNewIndexesToSDL_WithCircularOneToOne_DoesNotAddIndex(t *testing.T) {
	sdl := `type User {
	toleratedBy: Cat @relation(name:"tolerates")
}

type Cat {
	loves: Mouse @primary @relation(name:"loves")
	tolerates: User @primary @relation(name:"tolerates")
}

type Mouse {
	lovedBy: Cat @relation(name:"loves")
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "loves: Mouse @primary")
	assert.NotContains(t, result, "loves: Mouse @index")
	assert.Contains(t, result, "tolerates: User @primary")
	assert.NotContains(t, result, "tolerates: User @index")

	assert.Contains(t, result, "toleratedBy: Cat @relation")
	assert.NotContains(t, result, "toleratedBy: Cat @index")
	assert.Contains(t, result, "lovedBy: Cat @relation")
	assert.NotContains(t, result, "lovedBy: Cat @index")
}

func TestNewIndexesToSDL_WithManyToManyJoinTable_IndexesJoinRelations(t *testing.T) {
	sdl := `type Student {
	name: String
}

type Course {
	name: String
}

type Enrollment {
	student: Student @relation(name: "student_enrollments")
	course: Course @relation(name: "course_enrollments")
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")

	assert.Contains(t, result, "student: Student @index")
	assert.Contains(t, result, "course: Course @index")
}

func TestNewIndexesToSDL_WithSingleRelationNoBackReference_AddsIndex(t *testing.T) {
	sdl := `type Author {
	name: String
}

type Book {
	title: String
	author: Author
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name: String @index")
	assert.Contains(t, result, "title: String @index")

	assert.Contains(t, result, "author: Author @index")
}

func TestNewIndexesToSDL_WithVariousFormatting_PreservesWhitespace(t *testing.T) {
	sdl := `type User {
	name:    String
	age:Int
}`
	result := addIndexesToSDL(sdl)

	assert.Contains(t, result, "name:    String @index")
	assert.Contains(t, result, "age:Int @index")
}

func TestApply_WithIndexActions_StillModifiesSDL(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.NewIndex{CollectionID: 0, FieldName: "name"},
	}

	result := m.Apply(actions)

	collectionAdd, ok := result[0].(*action.AddCollection)
	assert.True(t, ok)
	assert.Contains(t, collectionAdd.SDL, "@index")

	newIndex, ok := result[1].(*action.NewIndex)
	assert.True(t, ok)
	assert.Equal(t, 0, newIndex.CollectionID)
	assert.Equal(t, "name", newIndex.FieldName)
}

func TestApply_WithIndexDirective_ReturnsUnchanged(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String @index }"},
	}

	result := m.Apply(actions)

	assert.Equal(t, actions, result)
}

func TestApply_WithoutIndex_ModifiesSDL(t *testing.T) {
	m := &secondaryIndex{}

	original := `type User {
	name: String
	age: Int
}`

	actions := action.Actions{
		&action.AddCollection{SDL: original},
	}

	result := m.Apply(actions)

	assert.NotEqual(t, actions, result)

	collectionAdd, ok := result[0].(*action.AddCollection)
	assert.True(t, ok)
	assert.Contains(t, collectionAdd.SDL, "name: String @index")
	assert.Contains(t, collectionAdd.SDL, "age: Int @index")
}

func TestName_ReturnsSecondaryIndex(t *testing.T) {
	m := &secondaryIndex{}
	assert.Equal(t, SecondaryIndex, m.Name())
	assert.Equal(t, Name("secondary-index"), m.Name())
}

func TestShouldSkip_WithIndexActions_ReturnsTrue(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.NewIndex{CollectionID: 0, FieldName: "name"},
	}

	assert.True(t, m.ShouldSkip(actions))
}

func TestShouldSkip_WithIndexDirective_ReturnsTrue(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String @index }"},
	}

	assert.True(t, m.ShouldSkip(actions))
}

func TestShouldSkip_WithUniqueIndexDirective_ReturnsTrue(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddCollection{SDL: "type User { email: String @index(unique: true) }"},
	}

	assert.True(t, m.ShouldSkip(actions))
}

func TestShouldSkip_WithExplainRequest_ReturnsTrue(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.ExplainRequest{Request: `query @explain(type: debug) { User { name } }`},
	}

	assert.True(t, m.ShouldSkip(actions))
}

func TestShouldSkip_WithRequestContainingExplainDirective_ReturnsTrue(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
		&action.Request{Request: `query @explain(type: simple) { User { name } }`},
	}

	assert.True(t, m.ShouldSkip(actions))
}

func TestShouldSkip_WithNoIndex_ReturnsFalse(t *testing.T) {
	m := &secondaryIndex{}

	actions := action.Actions{
		&action.AddCollection{SDL: "type User { name: String }"},
	}

	assert.False(t, m.ShouldSkip(actions))
}

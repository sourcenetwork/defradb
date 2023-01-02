// Copyright 2023 Democratized Data Foundation
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
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/client"
)

func TestTypedAddField(t *testing.T) {
	var db client.DB
	var ctx context.Context

	collections, err := db.PatchSchema(
		ctx,
		AppendField("User", "Email", "String"),
	)
}

func TestStringAddField(t *testing.T) {
	var db client.DB
	var ctx context.Context

	collections, err := db.PatchSchemaS(
		ctx,
		// `-` is JSON patch syntax for appending at the end of a list
		`{ "op": "add", "path": "/User/-", "value": "Email: String" }`,
	)
}

func TestTypedRemoveField(t *testing.T) {
	var db client.DB
	var ctx context.Context

	collections, err := db.PatchSchema(
		ctx,
		RemoveField("User", "LastName"),
	)
}

func TestStringRemoveField(t *testing.T) {
	var db client.DB
	var ctx context.Context

	collections, err := db.PatchSchemaS(
		ctx,
		`{ "op": "remove", "path": "/User/LastName" }`,
	)
}

func TestTypedMultiple(t *testing.T) {
	var db client.DB
	var ctx context.Context

	collections, err := db.PatchSchema(
		ctx,
		// If any of these operations fail, the whole set should be reverted/not-commited
		AppendField("User", "LastName", "String"),
		RemoveField("User", "Fax"),
		AppendField("User", "PhoneNumber", "String"),
		AppendSchema(
			`type Dog {
				Name: String
				Owner: User
			}`,
		),
		AppendField("User", "Dogs", "[Dog]"),
	)
}

func TestStringMultiple(t *testing.T) {
	var db client.DB
	var ctx context.Context

	collections, err := db.PatchSchemaS(
		ctx,
		// If any of these operations fail, the whole set should be reverted/not-commited
		`[
			{ "op": "add", "path": "/User/-", "value": "LastName: String" },
			{ "op": "remove", "path": "/User/Fax" },
			{ "op": "add", "path": "/User/-", "value": "PhoneNumber: String" },
			{
				"op": "add",
				"path": "/-",
				"value": "type Dog {
					Name: String
					Owner: User
				}",
			},
			{ "op": "add", "path": "/User/-", "value": "Dog: [Dog]" },
		]`,
	)
}

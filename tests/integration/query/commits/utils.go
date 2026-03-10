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

package commits

import (
	"github.com/sourcenetwork/defradb/tests/action"
)

const userCollectionGQLSchema = (`
	type Users {
		name: String
		age: Int
		verified: Boolean
	}
`)

func updateUserCollectionSchema() *action.AddCollection {
	return &action.AddCollection{
		SDL: userCollectionGQLSchema,
	}
}

// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package state

// Todo - this file is not yet a testo multiplier, and instead contains legacy code.
// There is an issue to convert this here: https://github.com/sourcenetwork/defradb/issues/4393

// The MutationType that tests will run using.
//
// For example if set to [CollectionSaveMutationType], all supporting
// actions (such as [UpdateDoc]) will execute via [Collection.Save].
//
// Defaults to CollectionSaveMutationType.
type MutationType string

const (
	// CollectionSaveMutationType will cause all supporting actions
	// to run their mutations via [Collection.Save].
	CollectionSaveMutationType MutationType = "collection-save"

	// CollectionNamedMutationType will cause all supporting actions
	// to run their mutations via their corresponding named [Collection]
	// call.
	//
	// For example, CreateDoc will call [Collection.Create], and
	// UpdateDoc will call [Collection.Update].
	CollectionNamedMutationType MutationType = "collection-named"

	// GQLRequestMutationType will cause all supporting actions to
	// run their mutations using GQL requests, typically these will
	// include a `id` parameter to target the specified document.
	GQLRequestMutationType MutationType = "gql"
)

var ActiveMutationType MutationType

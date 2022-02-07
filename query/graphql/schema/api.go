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
	gql "github.com/graphql-go/graphql"
	"github.com/sourcenetwork/defradb/query/graphql/schema/types"
)

var queryAllCommits = &gql.Field{
	Name: "allCommits",
	Type: gql.NewList(types.Commit),
	Args: gql.FieldConfigArgument{
		"dockey": newArgConfig(gql.NewNonNull(gql.ID)),
		"field":  newArgConfig(gql.String),
	},
}

var queryLatestCommits = &gql.Field{
	Name: "latestCommits",
	Type: gql.NewList(types.Commit),
	Args: gql.FieldConfigArgument{
		"dockey": newArgConfig(gql.NewNonNull(gql.ID)),
		"field":  newArgConfig(gql.String),
	},
}

var queryCommit = &gql.Field{
	Name: "commit",
	Type: types.Commit,
	Args: gql.FieldConfigArgument{
		"cid": newArgConfig(gql.NewNonNull(gql.ID)),
	},
}

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
)

var queryAllCommits = &gql.Field{
	Name: "allCommits",
	Type: gql.NewList(commit),
	Args: gql.FieldConfigArgument{
		"dockey": newArgConfig(gql.NewNonNull(gql.ID)),
		"field":  newArgConfig(gql.String),
	},
}

var queryLatestCommits = &gql.Field{
	Name: "latestCommits",
	Type: gql.NewList(commit),
	Args: gql.FieldConfigArgument{
		"dockey": newArgConfig(gql.NewNonNull(gql.ID)),
		"field":  newArgConfig(gql.String),
	},
}

var queryCommit = &gql.Field{
	Name: "commit",
	Type: commit,
	Args: gql.FieldConfigArgument{
		"cid": newArgConfig(gql.NewNonNull(gql.ID)),
	},
}

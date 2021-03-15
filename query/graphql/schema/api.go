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

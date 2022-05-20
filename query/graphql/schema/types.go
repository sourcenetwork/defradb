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

var (
	// Helper only for `commit` below.
	commitCountFieldArg = gql.NewEnum(gql.EnumConfig{
		Name: "commitCountFieldArg",
		Values: gql.EnumValueConfigMap{
			"links": &gql.EnumValueConfig{Value: "links"},
		},
	})

	// commit represents an individual commit to a MerkleCRDT
	// type Commit {
	// 	Height: Int
	// 	CID: String
	// 	Delta: String
	// 	Previous: [Commit]
	//  Links: [Commit]
	// }
	//
	// Any self referential type needs to be initalized
	// inside the init() func
	commit = gql.NewObject(gql.ObjectConfig{
		Name: "Commit",
		Fields: gql.Fields{
			"height": &gql.Field{
				Type: gql.Int,
			},
			"cid": &gql.Field{
				Type: gql.String,
			},
			"delta": &gql.Field{
				Type: gql.String,
			},
			// "previous": &gql.Field{
			// 	Type: gql.NewList(Commit),
			// },
			"links": &gql.Field{
				Type: gql.NewList(commitLink),
			},
			"_count": &gql.Field{
				Type: gql.Int,
				Args: gql.FieldConfigArgument{
					"field": &gql.ArgumentConfig{
						Type: commitCountFieldArg,
					},
				},
			},
			// "tests": &gql.Field{
			// 	Type: gql.NewList(gql.String),
			// },
		},
	})

	// Delta represents a Delta State update for a CRDT
	// type Delta {
	// 	Payload: String
	// }
	delta = gql.NewObject(gql.ObjectConfig{
		Name: "Delta",
		Fields: gql.Fields{
			"payload": &gql.Field{
				Type: gql.String,
			},
		},
	})

	// CommitLink is a named DAG link between commits.
	// This is primary used for CompositeDAG CRDTs
	commitLink = gql.NewObject(gql.ObjectConfig{
		Name: "CommitLink",
		Fields: gql.Fields{
			"name": &gql.Field{
				Type: gql.String,
			},
			"cid": &gql.Field{
				Type: gql.String,
			},
		},
	})
)

// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package types

import (
	gql "github.com/sourcenetwork/graphql-go"

	"github.com/sourcenetwork/defradb/client/request"
)

// PageInfoObject creates the PageInfo GraphQL type for cursor pagination metadata.
func PageInfoObject() *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
		Name:        request.PageInfoTypeName,
		Description: "Pagination information for cursor-based queries",
		Fields: gql.Fields{
			request.HasNextFieldName: &gql.Field{
				Type:        gql.Boolean,
				Description: "Whether there are more results after the current page",
			},
			request.HasPrevFieldName: &gql.Field{
				Type:        gql.Boolean,
				Description: "Whether there are results before the current page",
			},
			request.StartCursorFieldName: &gql.Field{
				Type:        gql.String,
				Description: "Opaque cursor for the first item in the current page",
			},
			request.EndCursorFieldName: &gql.Field{
				Type:        gql.String,
				Description: "Opaque cursor for the last item in the current page",
			},
		},
	})
}

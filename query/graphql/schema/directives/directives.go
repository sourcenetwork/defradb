// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package directives

import (
	gql "github.com/graphql-go/graphql"
)

var (
	Explain = gql.NewDirective(gql.DirectiveConfig{
		Name: "explain",
		Args: gql.FieldConfigArgument{
			"simple": &gql.ArgumentConfig{
				Type:         gql.Boolean,
				DefaultValue: true,
			},
			"predict": &gql.ArgumentConfig{
				Type:         gql.Boolean,
				DefaultValue: false,
			},
			"execute": &gql.ArgumentConfig{
				Type:         gql.Boolean,
				DefaultValue: false,
			},
		},
		// A directive is unique to it's location and the location must be provided for directives.
		// We limit @explain directive to only be valid at these two locations: `MUTATION`, `QUERY`.
		Locations: []string{
			gql.DirectiveLocationQuery,
			gql.DirectiveLocationMutation,
		},
	})
)

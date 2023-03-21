// Copyright 2022 Democratized Data Foundation
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
	gql "github.com/graphql-go/graphql"
)

const (
	ExplainLabel  string = "explain"
	PrimaryLabel  string = "primary"
	RelationLabel string = "relation"

	ExplainArgNameType string = "type"
	ExplainArgSimple   string = "simple"
)

var (
	// OrderingEnum is an enum for the Ordering argument.
	OrderingEnum = gql.NewEnum(gql.EnumConfig{
		Name: "Ordering",
		Values: gql.EnumValueConfigMap{
			"ASC": &gql.EnumValueConfig{
				Value: 0,
			},
			"DESC": &gql.EnumValueConfig{
				Value: 1,
			},
		},
	})

	ExplainEnum = gql.NewEnum(gql.EnumConfig{
		Name:        "ExplainType",
		Description: "ExplainType is an enum selecting the type of explanation done by the @explain directive.",
		Values: gql.EnumValueConfigMap{
			ExplainArgSimple: &gql.EnumValueConfig{
				Value:       ExplainArgSimple,
				Description: "Simple explaination - dump of the plan graph.",
			},
		},
	})

	ExplainDirective *gql.Directive = gql.NewDirective(gql.DirectiveConfig{
		Name:        ExplainLabel,
		Description: "@explain is a directive that can be used to explain the query.",
		Args: gql.FieldConfigArgument{
			ExplainArgNameType: &gql.ArgumentConfig{
				Type: ExplainEnum,
			},
		},

		// A directive is unique to it's location and the location must be provided for directives.
		// We limit @explain directive to only be valid at these two locations: `MUTATION`, `QUERY`.
		Locations: []string{
			gql.DirectiveLocationQuery,
			gql.DirectiveLocationMutation,
		},
	})

	// PrimaryDirective @primary is used to indicate the primary
	// side of a one-to-one relationship.
	PrimaryDirective = gql.NewDirective(gql.DirectiveConfig{
		Name: PrimaryLabel,
		Locations: []string{
			gql.DirectiveLocationFieldDefinition,
		},
	})

	// RelationDirective @relation is used to explicitly define
	// the attributes of a relationship, specifically, the name
	// if you don't want to use the default generated relationship
	// name.
	RelationDirective = gql.NewDirective(gql.DirectiveConfig{
		Name: RelationLabel,
		Args: gql.FieldConfigArgument{
			"name": &gql.ArgumentConfig{
				Type: gql.String,
			},
		},
		Locations: []string{
			gql.DirectiveLocationFieldDefinition,
		},
	})
)

func NewArgConfig(t gql.Type) *gql.ArgumentConfig {
	return &gql.ArgumentConfig{
		Type: t,
	}
}

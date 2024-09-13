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
	gql "github.com/sourcenetwork/graphql-go"

	"github.com/sourcenetwork/defradb/client"
)

const (
	ExplainLabel  string = "explain"
	PrimaryLabel  string = "primary"
	RelationLabel string = "relation"

	ExplainArgNameType string = "type"
	ExplainArgSimple   string = "simple"
	ExplainArgExecute  string = "execute"
	ExplainArgDebug    string = "debug"

	CRDTDirectiveLabel    = "crdt"
	CRDTDirectivePropType = "type"

	PolicySchemaDirectiveLabel        = "policy"
	PolicySchemaDirectivePropID       = "id"
	PolicySchemaDirectivePropResource = "resource"

	IndexDirectiveLabel         = "index"
	IndexDirectivePropName      = "name"
	IndexDirectivePropUnique    = "unique"
	IndexDirectivePropDirection = "direction"
	IndexDirectivePropIncludes  = "includes"

	IndexFieldInputName      = "name"
	IndexFieldInputDirection = "direction"

	FieldOrderASC  = "ASC"
	FieldOrderDESC = "DESC"
)

// OrderingEnum is an enum for the Ordering argument.
func OrderingEnum() *gql.Enum {
	return gql.NewEnum(gql.EnumConfig{
		Name: "Ordering",
		Values: gql.EnumValueConfigMap{
			"ASC": &gql.EnumValueConfig{
				Description: ascOrderDescription,
				Value:       0,
			},
			"DESC": &gql.EnumValueConfig{
				Description: descOrderDescription,
				Value:       1,
			},
		},
	})
}

func ExplainEnum() *gql.Enum {
	return gql.NewEnum(gql.EnumConfig{
		Name:        "ExplainType",
		Description: "ExplainType is an enum selecting the type of explanation done by the @explain directive.",
		Values: gql.EnumValueConfigMap{
			ExplainArgSimple: &gql.EnumValueConfig{
				Value:       ExplainArgSimple,
				Description: "Simple explanation - dump of the plan graph.",
			},

			ExplainArgExecute: &gql.EnumValueConfig{
				Value:       ExplainArgExecute,
				Description: "Deeper explanation - insights gathered by executing the plan graph.",
			},

			ExplainArgDebug: &gql.EnumValueConfig{
				Value:       ExplainArgDebug,
				Description: "Like simple explain, but more verbose nodes (no attributes).",
			},
		},
	})
}

func ExplainDirective(explainEnum *gql.Enum) *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name:        ExplainLabel,
		Description: "@explain is a directive that can be used to explain the query.",
		Args: gql.FieldConfigArgument{
			ExplainArgNameType: &gql.ArgumentConfig{
				Type: explainEnum,
			},
		},

		// A directive is unique to it's location and the location must be provided for directives.
		// We limit @explain directive to only be valid at these two locations: `MUTATION`, `QUERY`.
		Locations: []string{
			gql.DirectiveLocationQuery,
			gql.DirectiveLocationMutation,
		},
	})
}

func PolicyDirective() *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name:        PolicySchemaDirectiveLabel,
		Description: "@policy is a directive that can be used to link a policy on a collection type.",
		Args: gql.FieldConfigArgument{
			PolicySchemaDirectivePropID: &gql.ArgumentConfig{
				Type: gql.String,
			},
			PolicySchemaDirectivePropResource: &gql.ArgumentConfig{
				Type: gql.String,
			},
		},
		Locations: []string{
			gql.DirectiveLocationObject,
		},
	})
}

func IndexFieldInputObject(orderingEnum *gql.Enum) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name: "IndexField",
		Fields: gql.InputObjectConfigFieldMap{
			IndexFieldInputName: &gql.InputObjectFieldConfig{
				Type: gql.String,
			},
			IndexFieldInputDirection: &gql.InputObjectFieldConfig{
				Type: orderingEnum,
			},
		},
	})
}

func IndexDirective(orderingEnum *gql.Enum, indexFieldInputObject *gql.InputObject) *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name:        IndexDirectiveLabel,
		Description: "@index is a directive that can be used to create an index on a type or a field.",
		Args: gql.FieldConfigArgument{
			IndexDirectivePropName: &gql.ArgumentConfig{
				Type: gql.String,
			},
			IndexDirectivePropUnique: &gql.ArgumentConfig{
				Type: gql.Boolean,
			},
			IndexDirectivePropDirection: &gql.ArgumentConfig{
				Type: gql.String,
			},
			IndexDirectivePropIncludes: &gql.ArgumentConfig{
				Type: gql.NewList(indexFieldInputObject),
			},
		},
		Locations: []string{
			gql.DirectiveLocationObject,
			gql.DirectiveLocationField,
		},
	})
}

func CRDTEnum() *gql.Enum {
	return gql.NewEnum(gql.EnumConfig{
		Name:        "CRDTType",
		Description: "One of the possible CRDT Types.",
		Values: gql.EnumValueConfigMap{
			client.LWW_REGISTER.String(): &gql.EnumValueConfig{
				Value:       client.LWW_REGISTER,
				Description: "Last Write Wins register",
			},
			client.PN_COUNTER.String(): &gql.EnumValueConfig{
				Value: client.PN_COUNTER,
				Description: `Positive-Negative Counter.
	
	WARNING: Incrementing an integer and causing it to overflow the int64 max value
	will cause the value to roll over to the int64 min value. Incremeting a float and
	causing it to overflow the float64 max value will act like a no-op.`,
			},
			client.P_COUNTER.String(): &gql.EnumValueConfig{
				Value: client.P_COUNTER,
				Description: `Positive Counter.
	
	WARNING: Incrementing an integer and causing it to overflow the int64 max value
	will cause the value to roll over to the int64 min value. Incremeting a float and
	causing it to overflow the float64 max value will act like a no-op.`,
			},
		},
	})
}

// CRDTFieldDirective @crdt is used to define the CRDT type of a field
func CRDTFieldDirective(crdtEnum *gql.Enum) *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name:        CRDTDirectiveLabel,
		Description: crdtDirectiveDescription,
		Args: gql.FieldConfigArgument{
			CRDTDirectivePropType: &gql.ArgumentConfig{
				Type: crdtEnum,
			},
		},
		Locations: []string{
			gql.DirectiveLocationField,
		},
	})
}

// PrimaryDirective @primary is used to indicate the primary
// side of a one-to-one relationship.
func PrimaryDirective() *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name:        PrimaryLabel,
		Description: primaryDirectiveDescription,
		Locations: []string{
			gql.DirectiveLocationFieldDefinition,
		},
	})
}

// RelationDirective @relation is used to explicitly define
// the attributes of a relationship, specifically, the name
// if you don't want to use the default generated relationship
// name.
func RelationDirective() *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name:        RelationLabel,
		Description: relationDirectiveDescription,
		Args: gql.FieldConfigArgument{
			"name": &gql.ArgumentConfig{
				Description: relationDirectiveNameArgDescription,
				Type:        gql.String,
			},
		},
		Locations: []string{
			gql.DirectiveLocationFieldDefinition,
		},
	})
}

func NewArgConfig(t gql.Type, description string) *gql.ArgumentConfig {
	return &gql.ArgumentConfig{
		Type:        t,
		Description: description,
	}
}

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

	ConstraintsDirectiveLabel    = "constraints"
	ConstraintsDirectivePropSize = "size"

	VectorEmbeddingDirectiveLabel        = "embedding"
	VectorEmbeddingDirectivePropProvider = "provider"
	VectorEmbeddingDirectivePropModel    = "model"
	VectorEmbeddingDirectivePropURL      = "url"
	VectorEmbeddingDirectivePropFields   = "fields"
	VectorEmbeddingDirectivePropTemplate = "template"

	PolicySchemaDirectiveLabel        = "policy"
	PolicySchemaDirectivePropID       = "id"
	PolicySchemaDirectivePropResource = "resource"

	IndexDirectiveLabel         = "index"
	IndexDirectivePropName      = "name"
	IndexDirectivePropUnique    = "unique"
	IndexDirectivePropDirection = "direction"
	IndexDirectivePropIncludes  = "includes"

	IncludesPropField     = "field"
	IncludesPropDirection = "direction"

	DefaultDirectiveLabel        = "default"
	DefaultDirectivePropString   = "string"
	DefaultDirectivePropBool     = "bool"
	DefaultDirectivePropInt      = "int"
	DefaultDirectivePropFloat    = "float"
	DefaultDirectivePropFloat32  = "float32"
	DefaultDirectivePropFloat64  = "float64"
	DefaultDirectivePropDateTime = "dateTime"
	DefaultDirectivePropJSON     = "json"
	DefaultDirectivePropBlob     = "blob"

	MaterializedDirectiveLabel  = "materialized"
	MaterializedDirectivePropIf = "if"

	BranchableDirectiveLabel  = "branchable"
	BranchableDirectivePropIf = "if"

	FieldOrderASC  = "ASC"
	FieldOrderDESC = "DESC"

	SimilarityArgVector = "vector"
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

func DefaultDirective() *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name: DefaultDirectiveLabel,
		Description: `@default is a directive that can be used to set a default field value.
		
		Setting a default value on a field within a view has no effect.`,
		Args: gql.FieldConfigArgument{
			DefaultDirectivePropString: &gql.ArgumentConfig{
				Type: gql.String,
			},
			DefaultDirectivePropBool: &gql.ArgumentConfig{
				Type: gql.Boolean,
			},
			DefaultDirectivePropInt: &gql.ArgumentConfig{
				Type: gql.Int,
			},
			DefaultDirectivePropFloat: &gql.ArgumentConfig{
				Type: gql.Float,
			},
			DefaultDirectivePropFloat32: &gql.ArgumentConfig{
				Type: Float32,
			},
			DefaultDirectivePropFloat64: &gql.ArgumentConfig{
				Type: Float64,
			},
			DefaultDirectivePropDateTime: &gql.ArgumentConfig{
				Type: gql.DateTime,
			},
			DefaultDirectivePropJSON: &gql.ArgumentConfig{
				Type: JSONScalarType(),
			},
			DefaultDirectivePropBlob: &gql.ArgumentConfig{
				Type: BlobScalarType(),
			},
		},
		Locations: []string{
			gql.DirectiveLocationFieldDefinition,
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
		Name:        "IndexField",
		Description: "Used to create an index from a field.",
		Fields: gql.InputObjectConfigFieldMap{
			IncludesPropField: &gql.InputObjectFieldConfig{
				Type: gql.String,
			},
			IncludesPropDirection: &gql.InputObjectFieldConfig{
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
				Description: "Sets the index name.",
				Type:        gql.String,
			},
			IndexDirectivePropUnique: &gql.ArgumentConfig{
				Description: "Makes the index unique.",
				Type:        gql.Boolean,
			},
			IndexDirectivePropDirection: &gql.ArgumentConfig{
				Description: `Sets the default index ordering for all fields.
				
	If a field in the includes list does not specify a direction
	the default ordering from this value will be used instead.`,
				Type: orderingEnum,
			},
			IndexDirectivePropIncludes: &gql.ArgumentConfig{
				Description: `Sets the fields the index is created on.

	When used on a field definition and the field is not in the includes list
	it will be implicitly added as the first entry.`,
				Type: gql.NewList(indexFieldInputObject),
			},
		},
		Locations: []string{
			gql.DirectiveLocationObject,
			gql.DirectiveLocationFieldDefinition,
		},
	})
}

func MaterializedDirective() *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name: MaterializedDirectiveLabel,
		Description: `@materialized is a directive that specifies whether a collection is cached or not.
 It will default to true if ommited.  If multiple @materialized directives are provided, they will aggregated
 with OR logic (if any are true, the collection will be cached).`,
		Args: gql.FieldConfigArgument{
			MaterializedDirectivePropIf: &gql.ArgumentConfig{
				Type: gql.Boolean,
			},
		},
		Locations: []string{
			gql.DirectiveLocationObject,
		},
	})
}

func BranchableDirective() *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name: BranchableDirectiveLabel,
		// Todo: This description will need to be changed with:
		// https://github.com/sourcenetwork/defradb/issues/3219
		Description: `@branchable is a directive that defines whether the history of this collection is tracked
 as a single, verifiable entity or not. It will default to false if ommited.

 If multiple @branchable directives are provided, they will aggregated with OR logic (if any are true, the
 collection history will be tracked).

 The history may be queried like a document history can be queried, for example via 'commits'
 GQL queries.

 Currently this property is immutable and can only be set on collection creation, however
 that will change in the future.`,
		Args: gql.FieldConfigArgument{
			BranchableDirectivePropIf: &gql.ArgumentConfig{
				Type: gql.Boolean,
			},
		},
		Locations: []string{
			gql.DirectiveLocationObject,
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
			gql.DirectiveLocationFieldDefinition,
		},
	})
}

// ConstraintsDirective @constraints is used to define various constraints on a field.
func ConstraintsDirective() *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name:        ConstraintsDirectiveLabel,
		Description: constraintsDirectiveDescription,
		Locations: []string{
			gql.DirectiveLocationFieldDefinition,
		},
		Args: gql.FieldConfigArgument{
			ConstraintsDirectivePropSize: &gql.ArgumentConfig{
				Type:        gql.Int,
				Description: "The size constraint for array fields.",
			},
		},
	})
}

// VectorEmbeddingDirective @embedding is used to configure the generation of embedding vectors.
func VectorEmbeddingDirective() *gql.Directive {
	return gql.NewDirective(gql.DirectiveConfig{
		Name:        VectorEmbeddingDirectiveLabel,
		Description: embeddingDirectiveDescription,
		Locations: []string{
			gql.DirectiveLocationFieldDefinition,
		},
		Args: gql.FieldConfigArgument{
			VectorEmbeddingDirectivePropProvider: &gql.ArgumentConfig{
				Type:        gql.String,
				Description: "The provider to use for embedding. (ollama, openAI, etc.)",
			},
			VectorEmbeddingDirectivePropModel: &gql.ArgumentConfig{
				Type:        gql.String,
				Description: "The model to use for embedding. (nomic-embed-text, etc.)",
			},
			VectorEmbeddingDirectivePropURL: &gql.ArgumentConfig{
				Type:        gql.String,
				Description: "The URL of the provider API.",
			},
			VectorEmbeddingDirectivePropFields: &gql.ArgumentConfig{
				Type:        gql.NewList(gql.String),
				Description: "The fields to pass to the model.",
			},
			VectorEmbeddingDirectivePropTemplate: &gql.ArgumentConfig{
				Type:        gql.String,
				Description: "The template to use with the fields to create the content to feed the model.",
			},
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

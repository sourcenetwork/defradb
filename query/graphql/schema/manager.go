package schema

import (
	"github.com/sourcenetwork/defradb/query/graphql/schema/types"

	gql "github.com/graphql-go/graphql"
)

// SchemaManager creates an instanced management point
// for schema intake/outtake, and updates.
type SchemaManager struct {
	schema       gql.Schema
	definedTypes []*gql.Object // base defined types

	Generator *Generator
	Relations *RelationManager
}

// NewSchemaManager returns a new instance of a SchemaManager
// with a new default type map
func NewSchemaManager() (*SchemaManager, error) {
	sm := &SchemaManager{}
	schema, err := gql.NewSchema(gql.SchemaConfig{
		Types:    defaultTypes(),
		Query:    defaultQueryType(),
		Mutation: defaultMutationType(),
	})
	if err != nil {
		return sm, err
	}
	sm.schema = schema

	sm.NewGenerator()
	sm.Relations = NewRelationManager()

	return sm, nil
}

func (s *SchemaManager) Schema() *gql.Schema {
	return &s.schema
}

// ResolveTypes ensures all necessary types are defined, and
// resolves any remaning thunks/closures defined on object
// fields.
// Should be called *after* all dependant types have been added
func (s *SchemaManager) ResolveTypes() error {
	// basically, this function just refreshes the
	// schema.TypeMap, and runs the internal
	// typeMapReducer (https://github.com/graphql-go/graphql/blob/v0.7.9/schema.go#L275)
	// which ensures all the necessary types are defined in the
	// typeMap, and if there are any outstanding Thunks/closures
	// resolve them.

	// ATM, there is no function to easily call the internal
	// typeMapReducer function, so as a hack, we are just
	// going to re-add the Query type.
	query := s.schema.QueryType()
	return s.schema.AppendType(query)
}

// @todo: Use a better default Query type
func defaultQueryType() *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
		Name: "Query",
		Fields: gql.Fields{
			"_": &gql.Field{
				Name: "_",
				Type: gql.Boolean,
			},

			// database API queries
			// "test": queryAllCommits,
			queryLatestCommits.Name: queryLatestCommits,
			// queryCommit.Name:        queryCommit,
		},
	})
}

func defaultMutationType() *gql.Object {
	return gql.NewObject(gql.ObjectConfig{
		Name: "Mutation",
		Fields: gql.Fields{
			"_": &gql.Field{
				Name: "_",
				Type: gql.Boolean,
			},
		},
	})
}

// default type map includes all the native scalar types
func defaultTypes() []gql.Type {
	return []gql.Type{
		// Base Scalar types
		gql.Boolean,
		gql.DateTime,
		gql.Float,
		gql.ID,
		gql.Int,
		gql.String,

		// Base Query types

		// Sort/Order enum
		OrderingEnum,

		// filter scalar blocks
		BooleanOperatorBlock,
		DateTimeOperatorBlock,
		FloatOperatorBlock,
		IDOperatorBlock,
		IntOperatorBlock,
		StringOperatorBlock,

		types.CommitLink,
		// types.CommitLinks,
		types.Commit,
		types.Delta,
	}
}

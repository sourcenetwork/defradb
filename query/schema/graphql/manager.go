package graphql

import (
	gql "github.com/graphql-go/graphql"
)

// SchemaManager creates an instanced management point
// for schema intake/outtake, and updates.
type SchemaManager struct {
	schema       gql.Schema
	definedTypes []*gql.Object // base defined types
}

// NewSchemaManager returns a new instance of a SchemaManager
// with a new default type map
func NewSchemaManager() (*SchemaManager, error) {
	sm := &SchemaManager{}
	schema, err := gql.NewSchema(gql.SchemaConfig{
		Types: defaultTypes(),
	})
	if err != nil {
		return sm, err
	}
	sm.schema = schema
	return sm, nil
}

// default type map includes all the native scalar types
func defaultTypes() []gql.Type {
	return []gql.Type{
		gql.Boolean,
		gql.DateTime,
		gql.Float,
		gql.ID,
		gql.Int,
		gql.String,

		// Root Query Schema types
		// Sort/Order enum
		OrderingEnum,

		// filter scalar blocks
		BooleanOperatorBlock,
		DateTimeOperatorBlock,
		FloatOperatorBlock,
		IntOperatorBlock,
		StringOperatorBlock,
	}
}

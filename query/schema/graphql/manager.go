package graphql

import (
	gql "github.com/graphql-go/graphql"
)

// SchemaManager creates an instanced management point
// for schema intake/outtake, and updates.
type SchemaManager struct {
	schema  gql.Schema
	typeMap gql.TypeMap
	types   []*gql.Object // base defined types
}

// NewSchemaManager returns a new instance of a SchemaManager
// with a new default type map
func NewSchemaManager() *SchemaManager {
	return &SchemaManager{
		typeMap: defaultTypeMap(),
	}
}

// default type map includes all the native scalar types
func defaultTypeMap() gql.TypeMap {
	return gql.TypeMap{
		"Boolean":  gql.Boolean,
		"DateTime": gql.DateTime,
		"Float":    gql.Float,
		"ID":       gql.ID,
		"Int":      gql.Int,
		"String":   gql.String,

		// Root Query Schema types
		// Sort/Order enum
		"Ordering": OrderingEnum,

		// filter scalar blocks
		"BooleanOperatorBlock":  BooleanOperatorBlock,
		"DateTimeOperatorBlock": DateTimeOperatorBlock,
		"FloatOperatorBlock":    FloatOperatorBlock,
		"IntOperatorBlock":      IntOperatorBlock,
		"StringOperatorBlock":   StringOperatorBlock,
	}
}

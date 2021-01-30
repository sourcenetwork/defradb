package schema

import (
	gql "github.com/graphql-go/graphql"
)

// OrderingEnum is an enum for the Ordering argument
var OrderingEnum = gql.NewEnum(gql.EnumConfig{
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

// BooleanOperatorBlock filter block for boolean types
var BooleanOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "BooleanOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.Boolean,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Type: gql.Boolean,
		},
		"_like": &gql.InputObjectFieldConfig{
			Type: gql.Boolean,
		},
		"_in": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.Boolean)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.Boolean)),
		},
	},
})

// DateTimeOperatorBlock filter block for DateTime types
var DateTimeOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "DateTimeOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.DateTime,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Type: gql.DateTime,
		},
		"_gt": &gql.InputObjectFieldConfig{
			Type: gql.DateTime,
		},
		"_ge": &gql.InputObjectFieldConfig{
			Type: gql.DateTime,
		},
		"_lt": &gql.InputObjectFieldConfig{
			Type: gql.DateTime,
		},
		"_le": &gql.InputObjectFieldConfig{
			Type: gql.DateTime,
		},
		"_in": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.DateTime)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.DateTime)),
		},
	},
})

// FloatOperatorBlock filter block for Float types
var FloatOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "FloatOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.Float,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Type: gql.Float,
		},
		"_gt": &gql.InputObjectFieldConfig{
			Type: gql.Float,
		},
		"_ge": &gql.InputObjectFieldConfig{
			Type: gql.Float,
		},
		"_lt": &gql.InputObjectFieldConfig{
			Type: gql.Float,
		},
		"_le": &gql.InputObjectFieldConfig{
			Type: gql.Float,
		},
		"_in": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.Float)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.Float)),
		},
	},
})

// IntOperatorBlock filter block for Int types
var IntOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "IntOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.Int,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Type: gql.Int,
		},
		"_gt": &gql.InputObjectFieldConfig{
			Type: gql.Int,
		},
		"_ge": &gql.InputObjectFieldConfig{
			Type: gql.Int,
		},
		"_lt": &gql.InputObjectFieldConfig{
			Type: gql.Int,
		},
		"_le": &gql.InputObjectFieldConfig{
			Type: gql.Int,
		},
		"_in": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.Int)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.Int)),
		},
	},
})

// StringOperatorBlock filter block for string types
var StringOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "StringOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.String,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Type: gql.String,
		},
		"_like": &gql.InputObjectFieldConfig{
			Type: gql.String,
		},
		"_in": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.String)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.String)),
		},
	},
})

// StringOperatorBlock filter block for ID types
var IDOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "IDOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.ID,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Type: gql.ID,
		},
		"_in": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.ID)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.NewNonNull(gql.ID)),
		},
	},
})

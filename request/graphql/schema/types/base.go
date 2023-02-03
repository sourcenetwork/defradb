// Copyright 2023 Democratized Data Foundation
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

// BooleanOperatorBlock filter block for boolean types.
var BooleanOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "BooleanOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.Boolean,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Type: gql.Boolean,
		},
		"_in": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.Boolean),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.Boolean),
		},
	},
})

// NotNullBooleanOperatorBlock filter block for boolean! types.
var NotNullBooleanOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "NotNullBooleanOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.Boolean,
		},
		"_ne": &gql.InputObjectFieldConfig{
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

// DateTimeOperatorBlock filter block for DateTime types.
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
			Type: gql.NewList(gql.DateTime),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.DateTime),
		},
	},
})

// FloatOperatorBlock filter block for Float types.
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
			Type: gql.NewList(gql.Float),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.Float),
		},
	},
})

// NotNullFloatOperatorBlock filter block for Float! types.
var NotNullFloatOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "NotNullFloatOperatorBlock",
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

// IntOperatorBlock filter block for Int types.
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
			Type: gql.NewList(gql.Int),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.Int),
		},
	},
})

// NotNullIntOperatorBlock filter block for Int! types.
var NotNullIntOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "NotNullIntOperatorBlock",
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

// StringOperatorBlock filter block for string types.
var StringOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "StringOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.String,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Type: gql.String,
		},
		"_in": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.String),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Type: gql.NewList(gql.String),
		},
	},
})

// NotNullstringOperatorBlock filter block for string! types.
var NotNullstringOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name: "NotNullStringOperatorBlock",
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Type: gql.String,
		},
		"_ne": &gql.InputObjectFieldConfig{
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

// IdOperatorBlock filter block for ID types.
var IdOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	gql "github.com/graphql-go/graphql"
)

// orderingEnum is an enum for the Ordering argument.
var orderingEnum = gql.NewEnum(gql.EnumConfig{
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

// booleanOperatorBlock filter block for boolean types.
var booleanOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// notNullBooleanOperatorBlock filter block for boolean! types.
var notNullBooleanOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// dateTimeOperatorBlock filter block for DateTime types.
var dateTimeOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// floatOperatorBlock filter block for Float types.
var floatOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// notNullFloatOperatorBlock filter block for Float! types.
var notNullFloatOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// intOperatorBlock filter block for Int types.
var intOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// notNullIntOperatorBlock filter block for Int! types.
var notNullIntOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// stringOperatorBlock filter block for string types.
var stringOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// notNullstringOperatorBlock filter block for string! types.
var notNullstringOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

// idOperatorBlock filter block for ID types.
var idOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
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

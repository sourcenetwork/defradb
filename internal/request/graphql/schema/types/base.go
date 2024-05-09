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
)

// BooleanOperatorBlock filter block for boolean types.
var BooleanOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "BooleanOperatorBlock",
	Description: booleanOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.Boolean,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.Boolean,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.Boolean),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.Boolean),
		},
	},
})

// NotNullBooleanOperatorBlock filter block for boolean! types.
var NotNullBooleanOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "NotNullBooleanOperatorBlock",
	Description: notNullBooleanOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.Boolean,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.Boolean,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.Boolean)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.Boolean)),
		},
	},
})

// DateTimeOperatorBlock filter block for DateTime types.
var DateTimeOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "DateTimeOperatorBlock",
	Description: dateTimeOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.DateTime,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.DateTime,
		},
		"_gt": &gql.InputObjectFieldConfig{
			Description: gtOperatorDescription,
			Type:        gql.DateTime,
		},
		"_ge": &gql.InputObjectFieldConfig{
			Description: geOperatorDescription,
			Type:        gql.DateTime,
		},
		"_lt": &gql.InputObjectFieldConfig{
			Description: ltOperatorDescription,
			Type:        gql.DateTime,
		},
		"_le": &gql.InputObjectFieldConfig{
			Description: leOperatorDescription,
			Type:        gql.DateTime,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.DateTime),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.DateTime),
		},
	},
})

// FloatOperatorBlock filter block for Float types.
var FloatOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "FloatOperatorBlock",
	Description: floatOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.Float,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.Float,
		},
		"_gt": &gql.InputObjectFieldConfig{
			Description: gtOperatorDescription,
			Type:        gql.Float,
		},
		"_ge": &gql.InputObjectFieldConfig{
			Description: geOperatorDescription,
			Type:        gql.Float,
		},
		"_lt": &gql.InputObjectFieldConfig{
			Description: ltOperatorDescription,
			Type:        gql.Float,
		},
		"_le": &gql.InputObjectFieldConfig{
			Description: leOperatorDescription,
			Type:        gql.Float,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.Float),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.Float),
		},
	},
})

// NotNullFloatOperatorBlock filter block for Float! types.
var NotNullFloatOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "NotNullFloatOperatorBlock",
	Description: notNullFloatOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.Float,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.Float,
		},
		"_gt": &gql.InputObjectFieldConfig{
			Description: gtOperatorDescription,
			Type:        gql.Float,
		},
		"_ge": &gql.InputObjectFieldConfig{
			Description: geOperatorDescription,
			Type:        gql.Float,
		},
		"_lt": &gql.InputObjectFieldConfig{
			Description: ltOperatorDescription,
			Type:        gql.Float,
		},
		"_le": &gql.InputObjectFieldConfig{
			Description: leOperatorDescription,
			Type:        gql.Float,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.Float)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.Float)),
		},
	},
})

// IntOperatorBlock filter block for Int types.
var IntOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "IntOperatorBlock",
	Description: intOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.Int,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.Int,
		},
		"_gt": &gql.InputObjectFieldConfig{
			Description: gtOperatorDescription,
			Type:        gql.Int,
		},
		"_ge": &gql.InputObjectFieldConfig{
			Description: geOperatorDescription,
			Type:        gql.Int,
		},
		"_lt": &gql.InputObjectFieldConfig{
			Description: ltOperatorDescription,
			Type:        gql.Int,
		},
		"_le": &gql.InputObjectFieldConfig{
			Description: leOperatorDescription,
			Type:        gql.Int,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.Int),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.Int),
		},
	},
})

// NotNullIntOperatorBlock filter block for Int! types.
var NotNullIntOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "NotNullIntOperatorBlock",
	Description: notNullIntOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.Int,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.Int,
		},
		"_gt": &gql.InputObjectFieldConfig{
			Description: gtOperatorDescription,
			Type:        gql.Int,
		},
		"_ge": &gql.InputObjectFieldConfig{
			Description: geOperatorDescription,
			Type:        gql.Int,
		},
		"_lt": &gql.InputObjectFieldConfig{
			Description: ltOperatorDescription,
			Type:        gql.Int,
		},
		"_le": &gql.InputObjectFieldConfig{
			Description: leOperatorDescription,
			Type:        gql.Int,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.Int)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.Int)),
		},
	},
})

// StringOperatorBlock filter block for string types.
var StringOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "StringOperatorBlock",
	Description: stringOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.String,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.String,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.String),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.String),
		},
		"_like": &gql.InputObjectFieldConfig{
			Description: likeStringOperatorDescription,
			Type:        gql.String,
		},
		"_nlike": &gql.InputObjectFieldConfig{
			Description: nlikeStringOperatorDescription,
			Type:        gql.String,
		},
		"_ilike": &gql.InputObjectFieldConfig{
			Description: ilikeStringOperatorDescription,
			Type:        gql.String,
		},
		"_nilike": &gql.InputObjectFieldConfig{
			Description: nilikeStringOperatorDescription,
			Type:        gql.String,
		},
	},
})

// NotNullstringOperatorBlock filter block for string! types.
var NotNullstringOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "NotNullStringOperatorBlock",
	Description: notNullStringOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.String,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.String,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.String)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.String)),
		},
		"_like": &gql.InputObjectFieldConfig{
			Description: likeStringOperatorDescription,
			Type:        gql.String,
		},
		"_nlike": &gql.InputObjectFieldConfig{
			Description: nlikeStringOperatorDescription,
			Type:        gql.String,
		},
		"_ilike": &gql.InputObjectFieldConfig{
			Description: ilikeStringOperatorDescription,
			Type:        gql.String,
		},
		"_nilike": &gql.InputObjectFieldConfig{
			Description: nilikeStringOperatorDescription,
			Type:        gql.String,
		},
	},
})

// IdOperatorBlock filter block for ID types.
var IdOperatorBlock = gql.NewInputObject(gql.InputObjectConfig{
	Name:        "IDOperatorBlock",
	Description: idOperatorBlockDescription,
	Fields: gql.InputObjectConfigFieldMap{
		"_eq": &gql.InputObjectFieldConfig{
			Description: eqOperatorDescription,
			Type:        gql.ID,
		},
		"_ne": &gql.InputObjectFieldConfig{
			Description: neOperatorDescription,
			Type:        gql.ID,
		},
		"_in": &gql.InputObjectFieldConfig{
			Description: inOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.ID)),
		},
		"_nin": &gql.InputObjectFieldConfig{
			Description: ninOperatorDescription,
			Type:        gql.NewList(gql.NewNonNull(gql.ID)),
		},
	},
})

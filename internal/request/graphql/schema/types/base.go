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
func BooleanOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// BooleanListOperatorBlock filter block for [Boolean] types.
func BooleanListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "BooleanListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Boolean] values.",
		Fields: gql.InputObjectConfigFieldMap{
			"_any": &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			"_all": &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			"_none": &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
		},
	})
}

// NotNullBooleanOperatorBlock filter block for boolean! types.
func NotNullBooleanOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// NotNullBooleanListOperatorBlock filter block for [Boolean!] types.
func NotNullBooleanListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullBooleanListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Boolean!] values.",
		Fields: gql.InputObjectConfigFieldMap{
			"_any": &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			"_all": &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			"_none": &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
		},
	})
}

// DateTimeOperatorBlock filter block for DateTime types.
func DateTimeOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// FloatOperatorBlock filter block for Float types.
func FloatOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// FloatListOperatorBlock filter block for [Float] types.
func FloatListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "FloatListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Float] values.",
		Fields: gql.InputObjectConfigFieldMap{
			"_any": &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			"_all": &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			"_none": &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
		},
	})
}

// NotNullFloatOperatorBlock filter block for Float! types.
func NotNullFloatOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// NotNullFloatListOperatorBlock filter block for [NotNullFloat] types.
func NotNullFloatListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullFloatListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Float!] values.",
		Fields: gql.InputObjectConfigFieldMap{
			"_any": &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			"_all": &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			"_none": &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
		},
	})
}

// IntOperatorBlock filter block for Int types.
func IntOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// IntListOperatorBlock filter block for [Int] types.
func IntListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "IntListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Int] values.",
		Fields: gql.InputObjectConfigFieldMap{
			"_any": &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			"_all": &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			"_none": &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
		},
	})
}

// NotNullIntOperatorBlock filter block for Int! types.
func NotNullIntOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// NotNullIntListOperatorBlock filter block for [NotNullInt] types.
func NotNullIntListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullIntListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Int!] values.",
		Fields: gql.InputObjectConfigFieldMap{
			"_any": &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			"_all": &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			"_none": &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
		},
	})
}

// StringOperatorBlock filter block for string types.
func StringOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// StringListOperatorBlock filter block for [String] types.
func StringListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "StringListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [String] values.",
		Fields: gql.InputObjectConfigFieldMap{
			"_any": &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			"_all": &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			"_none": &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
		},
	})
}

// NotNullStringOperatorBlock filter block for string! types.
func NotNullStringOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

// NotNullStringListOperatorBlock filter block for [String!] types.
func NotNullStringListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullStringListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [String!] values.",
		Fields: gql.InputObjectConfigFieldMap{
			"_any": &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			"_all": &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			"_none": &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
		},
	})
}

// JSONOperatorBlock filter block for string types.
func JSONOperatorBlock(jsonScalarType *gql.Scalar) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "JSONOperatorBlock",
		Description: stringOperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_eq": &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        jsonScalarType,
			},
			"_ne": &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        jsonScalarType,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(jsonScalarType),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(jsonScalarType),
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
}

// NotNullJSONOperatorBlock filter block for string! types.
func NotNullJSONOperatorBlock(jsonScalarType *gql.Scalar) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullJSONOperatorBlock",
		Description: notNullStringOperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_eq": &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        jsonScalarType,
			},
			"_ne": &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        jsonScalarType,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(jsonScalarType)),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(jsonScalarType)),
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
}

func BlobOperatorBlock(blobScalarType *gql.Scalar) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "BlobOperatorBlock",
		Description: stringOperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_eq": &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        blobScalarType,
			},
			"_ne": &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        blobScalarType,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(blobScalarType),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(blobScalarType),
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
}

// NotNullJSONOperatorBlock filter block for string! types.
func NotNullBlobOperatorBlock(blobScalarType *gql.Scalar) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullBlobOperatorBlock",
		Description: notNullStringOperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_eq": &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        blobScalarType,
			},
			"_ne": &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        blobScalarType,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(blobScalarType)),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(blobScalarType)),
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
}

// IDOperatorBlock filter block for ID types.
func IDOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
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
}

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

// Float64OperatorBlock filter block for Float types.
func Float64OperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "Float64OperatorBlock",
		Description: float64OperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_eq": &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        Float64,
			},
			"_ne": &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        Float64,
			},
			"_gt": &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        Float64,
			},
			"_ge": &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        Float64,
			},
			"_lt": &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        Float64,
			},
			"_le": &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        Float64,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(Float64),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(Float64),
			},
		},
	})
}

// Float64ListOperatorBlock filter block for [Float] types.
func Float64ListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "Float64ListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Float64] values.",
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

// NotNullFloat64OperatorBlock filter block for Float! types.
func NotNullFloat64OperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullFloat64OperatorBlock",
		Description: notNullFloat64OperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_eq": &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        Float64,
			},
			"_ne": &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        Float64,
			},
			"_gt": &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        Float64,
			},
			"_ge": &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        Float64,
			},
			"_lt": &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        Float64,
			},
			"_le": &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        Float64,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float64)),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float64)),
			},
		},
	})
}

// NotNullFloat64ListOperatorBlock filter block for [NotNullFloat] types.
func NotNullFloat64ListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullFloat64ListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Float64!] values.",
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

// Float32OperatorBlock filter block for Float32 types.
func Float32OperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "Float32OperatorBlock",
		Description: float32OperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_eq": &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        Float32,
			},
			"_ne": &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        Float32,
			},
			"_gt": &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        Float32,
			},
			"_ge": &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        Float32,
			},
			"_lt": &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        Float32,
			},
			"_le": &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        Float32,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(Float32),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(Float32),
			},
		},
	})
}

// Float32ListOperatorBlock filter block for [Float32] types.
func Float32ListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "Float32ListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Float32] values.",
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

// NotNullFloat32OperatorBlock filter block for Float32! types.
func NotNullFloat32OperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullFloat32OperatorBlock",
		Description: notNullFloat32OperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_eq": &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        Float32,
			},
			"_ne": &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        Float32,
			},
			"_gt": &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        Float32,
			},
			"_ge": &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        Float32,
			},
			"_lt": &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        Float32,
			},
			"_le": &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        Float32,
			},
			"_in": &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float32)),
			},
			"_nin": &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float32)),
			},
		},
	})
}

// NotNullFloat32ListOperatorBlock filter block for [NotNullFloat32] types.
func NotNullFloat32ListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullFloat32ListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [Float32!] values.",
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

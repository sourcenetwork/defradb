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

	"github.com/sourcenetwork/defradb/internal/connor"
)

// BooleanOperatorBlock filter block for boolean types.
func BooleanOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "BooleanOperatorBlock",
		Description: booleanOperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterBoolean,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterBoolean,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.Boolean),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.Boolean),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.Boolean),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterBoolean,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterBoolean,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.Boolean)),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.Boolean)),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.Boolean)),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.GreaterOp: &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.GreaterOrEqualOp: &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.LesserOp: &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.LesserOrEqualOp: &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.DateTime),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.DateTime),
			},
		},
	})
}

// NotNullDateTimeOperatorBlock filter block for DateTime! types.
func NotNullDateTimeOperatorBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullDateTimeOperatorBlock",
		Description: dateTimeOperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.GreaterOp: &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.GreaterOrEqualOp: &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.LesserOp: &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.LesserOrEqualOp: &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        FilterDateTime,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.DateTime)),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.DateTime)),
			},
		},
	})
}

// DateTimeListOperatorBlock filter block for [DateTime] types.
func DateTimeListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "DateTimeListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [DateTime] values.",
		Fields: gql.InputObjectConfigFieldMap{
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.DateTime),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.DateTime),
			},
		},
	})
}

// NotNullDateTimeListOperatorBlock filter block for [DateTime!] types.
func NotNullDateTimeListOperatorBlock(op *gql.InputObject) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "NotNullDateTimeListOperatorBlock",
		Description: "These are the set of filter operators available for use when filtering on [DateTime!] values.",
		Fields: gql.InputObjectConfigFieldMap{
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.DateTime)),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.DateTime)),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.GreaterOp: &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.GreaterOrEqualOp: &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.LesserOp: &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.LesserOrEqualOp: &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(Float64),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(Float64),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(Float64),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.GreaterOp: &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.GreaterOrEqualOp: &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.LesserOp: &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.LesserOrEqualOp: &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        FilterFloat64,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float64)),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float64)),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float64)),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.GreaterOp: &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.GreaterOrEqualOp: &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.LesserOp: &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.LesserOrEqualOp: &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(Float32),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(Float32),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(Float32),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.GreaterOp: &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.GreaterOrEqualOp: &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.LesserOp: &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.LesserOrEqualOp: &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        FilterFloat32,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float32)),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float32)),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(Float32)),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterInt,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterInt,
			},
			connor.GreaterOp: &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        FilterInt,
			},
			connor.GreaterOrEqualOp: &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        FilterInt,
			},
			connor.LesserOp: &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        FilterInt,
			},
			connor.LesserOrEqualOp: &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        FilterInt,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.Int),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.Int),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.Int),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterInt,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterInt,
			},
			connor.GreaterOp: &gql.InputObjectFieldConfig{
				Description: gtOperatorDescription,
				Type:        FilterInt,
			},
			connor.GreaterOrEqualOp: &gql.InputObjectFieldConfig{
				Description: geOperatorDescription,
				Type:        FilterInt,
			},
			connor.LesserOp: &gql.InputObjectFieldConfig{
				Description: ltOperatorDescription,
				Type:        FilterInt,
			},
			connor.LesserOrEqualOp: &gql.InputObjectFieldConfig{
				Description: leOperatorDescription,
				Type:        FilterInt,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.Int)),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.Int)),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.Int)),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterString,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterString,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.String),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.String),
			},
			connor.LikeOp: &gql.InputObjectFieldConfig{
				Description: likeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.NotLikeOp: &gql.InputObjectFieldConfig{
				Description: nlikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.CaseInsensitiveLikeOp: &gql.InputObjectFieldConfig{
				Description: ilikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.CaseInsensitiveNotLikeOp: &gql.InputObjectFieldConfig{
				Description: nilikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.RegexOp: &gql.InputObjectFieldConfig{
				Description: regexStringOperatorDescription,
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.String),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.String),
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        FilterString,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        FilterString,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.String)),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.String)),
			},
			connor.LikeOp: &gql.InputObjectFieldConfig{
				Description: likeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.NotLikeOp: &gql.InputObjectFieldConfig{
				Description: nlikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.CaseInsensitiveLikeOp: &gql.InputObjectFieldConfig{
				Description: ilikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.CaseInsensitiveNotLikeOp: &gql.InputObjectFieldConfig{
				Description: nilikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.RegexOp: &gql.InputObjectFieldConfig{
				Description: regexStringOperatorDescription,
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
			connor.AnyOp: &gql.InputObjectFieldConfig{
				Description: anyOperatorDescription,
				Type:        op,
			},
			connor.AllOp: &gql.InputObjectFieldConfig{
				Description: allOperatorDescription,
				Type:        op,
			},
			connor.NoneOp: &gql.InputObjectFieldConfig{
				Description: noneOperatorDescription,
				Type:        op,
			},
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.String)),
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.String)),
			},
		},
	})
}

func BlobOperatorBlock(blobScalarType *gql.Scalar) *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "BlobOperatorBlock",
		Description: stringOperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        blobScalarType,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        blobScalarType,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(blobScalarType),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(blobScalarType),
			},
			connor.LikeOp: &gql.InputObjectFieldConfig{
				Description: likeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.NotLikeOp: &gql.InputObjectFieldConfig{
				Description: nlikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.CaseInsensitiveLikeOp: &gql.InputObjectFieldConfig{
				Description: ilikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.CaseInsensitiveNotLikeOp: &gql.InputObjectFieldConfig{
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
		Description: notNullBlobOperatorBlockDescription,
		Fields: gql.InputObjectConfigFieldMap{
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        blobScalarType,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        blobScalarType,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(blobScalarType)),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(blobScalarType)),
			},
			connor.LikeOp: &gql.InputObjectFieldConfig{
				Description: likeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.NotLikeOp: &gql.InputObjectFieldConfig{
				Description: nlikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.CaseInsensitiveLikeOp: &gql.InputObjectFieldConfig{
				Description: ilikeStringOperatorDescription,
				Type:        gql.String,
			},
			connor.CaseInsensitiveNotLikeOp: &gql.InputObjectFieldConfig{
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
			connor.EqualOp: &gql.InputObjectFieldConfig{
				Description: eqOperatorDescription,
				Type:        gql.ID,
			},
			connor.NotEqualOp: &gql.InputObjectFieldConfig{
				Description: neOperatorDescription,
				Type:        gql.ID,
			},
			connor.InOp: &gql.InputObjectFieldConfig{
				Description: inOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.ID)),
			},
			connor.NotInOp: &gql.InputObjectFieldConfig{
				Description: ninOperatorDescription,
				Type:        gql.NewList(gql.NewNonNull(gql.ID)),
			},
		},
	})
}

// ScalarAggregateNumericBlock is the default numeric scalar selector
// for aggregate input arguments
func ScalarAggregateNumericBlock() *gql.InputObject {
	return gql.NewInputObject(gql.InputObjectConfig{
		Name:        "ScalarAggregateNumericBlock",
		Description: scalarAggregateSelectorDescription,
		Fields: gql.InputObjectConfigFieldMap{
			"_": &gql.InputObjectFieldConfig{
				Type: gql.Boolean,
			},
		},
	})
}

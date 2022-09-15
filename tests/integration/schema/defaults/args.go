// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package defaults

var allDefaultGroupArgs = []Field{
	FilterArg,
	GroupByArg,
	LimitArg,
	OffsetArg,
	OrderArg,
}

var allDefaultArgs = []Field{
	FilterArg,
	GroupByArg,
	LimitArg,
	OffsetArg,
	OrderArg,
	CidArg,
	DockeyArg,
	DockeysArg,
}

// DefaultFields contains the list of fields every
// defra schema-object should have.
var DefaultFields = Concat(
	Fields{
		keyField,
		versionField,
		groupField,
	},
	aggregateFields,
)

var keyField = Field{
	"name": "_key",
	"type": map[string]interface{}{
		"kind": "SCALAR",
		"name": "ID",
	},
}

var versionField = Field{
	"name": "_version",
	"type": map[string]interface{}{
		"kind": "LIST",
		"name": nil,
	},
}

var groupField = Field{
	"name": "_group",
	"type": map[string]interface{}{
		"kind": "LIST",
		"name": nil,
	},
}

var aggregateFields = Fields{
	map[string]interface{}{
		"name": "_avg",
		"type": map[string]interface{}{
			"kind": "SCALAR",
			"name": "Float",
		},
	},
	map[string]interface{}{
		"name": "_count",
		"type": map[string]interface{}{
			"kind": "SCALAR",
			"name": "Int",
		},
	},
	map[string]interface{}{
		"name": "_sum",
		"type": map[string]interface{}{
			"kind": "SCALAR",
			"name": "Float",
		},
	},
}

var FilterArg = Field{
	"name": "filter",
	"type": filterArgType,
}

var GroupByArg = Field{
	"name": "groupBy",
	"type": groupByArgType,
}

var LimitArg = Field{
	"name": "limit",
	"type": intArgType,
}

var OffsetArg = Field{
	"name": "offset",
	"type": intArgType,
}

var OrderArg = Field{
	"name": "order",
	"type": orderArgType,
}

var CidArg = Field{
	"name": "cid",
	"type": stringArgType,
}

var DockeyArg = Field{
	"name": "dockey",
	"type": stringArgType,
}

var DockeysArg = Field{
	"name": "dockeys",
	"type": nilArgType,
}

var groupByArgType = Field{
	"name":        nil,
	"inputFields": nil,
	"ofType": Field{
		"kind": "NON_NULL",
		"name": nil,
	},
}

var intArgType = Field{
	"inputFields": nil,
	"name":        "Int",
	"ofType":      nil,
}

var stringArgType = Field{
	"inputFields": nil,
	"name":        "String",
}

var nilArgType = Field{
	"name":        nil,
	"inputFields": nil,
}

var filterArgType = Field{
	"name":   "authorFilterArg",
	"ofType": nil,
	"inputFields": []any{
		makeOuterType(Yes("_and"), Yes(nil), Yes(inputObjectAuthorFilterArg), No()),
		makeOuterType(Yes("_key"), Yes("IDOperatorBlock"), Yes(nil), No()),
		makeOuterType(Yes("_not"), Yes("authorFilterArg"), Yes(nil), No()),
		makeOuterType(Yes("_or"), Yes(nil), Yes(inputObjectAuthorFilterArg), No()),

		// Following can probably reworked to be dynamically added based on the schema.
		makeOuterType(Yes("age"), Yes("IntOperatorBlock"), Yes(nil), No()),
		makeOuterType(Yes("name"), Yes("StringOperatorBlock"), Yes(nil), No()),
		makeOuterType(Yes("verified"), Yes("BooleanOperatorBlock"), Yes(nil), No()),
		makeOuterType(Yes("wrote"), Yes("bookFilterArg"), Yes(nil), No()),
		makeOuterType(Yes("wrote_id"), Yes("IDOperatorBlock"), Yes(nil), No()),
	},
}

var orderArgType = Field{
	"name":   "authorOrderArg",
	"ofType": nil,
	"inputFields": []any{
		makeAuthorOrdering("_key"),

		// Following can probably reworked to be dynamically added based on the schema.
		makeAuthorOrdering("age"),
		makeAuthorOrdering("name"),
		makeAuthorOrdering("verified"),

		// Without the relation type we won't have the following ordering type(s).
		makeOuterType(Yes("wrote"), Yes("bookOrderArg"), Yes(nil), No()),
		makeAuthorOrdering("wrote_id"),
	},
}

func makeAuthorOrdering(name any) Field {
	return makeOuterType(Yes(name), Yes("Ordering"), Yes(nil), No())
}

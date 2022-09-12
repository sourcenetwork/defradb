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

var allDefaultGroupArgs = []any{
	FilterArg,
	GroupByArg,
	LimitArg,
	OffsetArg,
	OrderArg,
	CidArg,
	DockeyArg,
	DockeysArg,
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
	"inputFields": nil,
	"name":        nil,
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
		makeInputObject("_and", nil, inputObjectAuthorFilterArg),
		makeInputObject("_key", "IDOperatorBlock", nil),
		makeInputObject("_not", "authorFilterArg", nil),
		makeInputObject("_or", nil, inputObjectAuthorFilterArg),

		// Following can probably reworked to be dynamically added based on the schema.
		makeInputObject("age", "IntOperatorBlock", nil),
		makeInputObject("name", "StringOperatorBlock", nil),
		makeInputObject("verified", "BooleanOperatorBlock", nil),
		makeInputObject("wrote", "bookFilterBaseArg", nil),
		makeInputObject("wrote_id", "IDOperatorBlock", nil),
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
		makeInputObject("wrote", "bookOrderArg", nil),
		makeAuthorOrdering("wrote_id"),
	},
}

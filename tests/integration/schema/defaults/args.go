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
	filterArg,
	groupByArg,
	limitArg,
	offsetArg,
	orderArg,
	cidArg,
	dockeyArg,
	dockeysArg,
}

var filterArg = field{
	"name": "filter",
	"type": filterArgType,
}

var groupByArg = field{
	"name": "groupBy",
	"type": groupByArgType,
}

var limitArg = field{
	"name": "limit",
	"type": intArgType,
}

var offsetArg = field{
	"name": "offset",
	"type": intArgType,
}

var orderArg = field{
	"name": "order",
	"type": orderArgType,
}

var cidArg = field{
	"name": "cid",
	"type": stringArgType,
}

var dockeyArg = field{
	"name": "dockey",
	"type": stringArgType,
}

var dockeysArg = field{
	"name": "dockeys",
	"type": nilArgType,
}

var groupByArgType = field{
	"inputFields": nil,
	"name":        nil,
	"ofType": field{
		"kind": "NON_NULL",
		"name": nil,
	},
}

var intArgType = field{
	"inputFields": nil,
	"name":        "Int",
	"ofType":      nil,
}

var stringArgType = field{
	"inputFields": nil,
	"name":        "String",
}

var nilArgType = field{
	"name":        nil,
	"inputFields": nil,
}

var filterArgType = field{
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

var orderArgType = field{
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

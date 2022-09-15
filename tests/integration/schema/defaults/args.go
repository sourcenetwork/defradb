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

func makeNameTypeImplicit(name, iType any) Field {
	return makeNameType(Yes(name), Yes(iType))
}

var keyField = makeNameTypeImplicit("_key", scalarIDKind)
var versionField = makeNameTypeImplicit("_version", listKind)
var groupField = makeNameTypeImplicit("_group", listKind)

var aggregateFields = Fields{
	avgField,
	countField,
	sumField,
}

var avgField = makeNameTypeImplicit("_avg", scalarFloatKind)
var countField = makeNameTypeImplicit("_count", scalarIntKind)
var sumField = makeNameTypeImplicit("_sum", scalarFloatKind)

var FilterArg = makeNameTypeImplicit("filter", filterArgType)
var GroupByArg = makeNameTypeImplicit("groupBy", groupByArgType)
var LimitArg = makeNameTypeImplicit("limit", intArgType)
var OffsetArg = makeNameTypeImplicit("offset", intArgType)
var OrderArg = makeNameTypeImplicit("order", orderArgType)
var CidArg = makeNameTypeImplicit("cid", stringArgType)
var DockeyArg = makeNameTypeImplicit("dockey", stringArgType)
var DockeysArg = makeNameTypeImplicit("dockeys", nilArgType)

var intArgType = makeType(Yes("Int"), Yes(nil), Yes(nil))
var stringArgType = makeType(Yes("String"), No(), Yes(nil))
var groupByArgType = makeType(Yes(nil), Yes(nonNullKind), Yes(nil))
var nilArgType = makeType(Yes(nil), No(), Yes(nil))
var filterArgType = makeType(Yes("authorFilterArg"), Yes(nil), Yes(filterInputFieldsStatic))
var orderArgType = makeType(Yes("authorOrderArg"), Yes(nil), Yes(orderInputFieldsStatic))

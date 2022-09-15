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

func makeNameKind(name, kind any) Field {
	return makeNameKindType(Yes(name), No(), Yes(kind))
}

var listKind = makeNameKind(nil, "LIST")
var scalarFloatKind = makeNameKind("Float", "SCALAR")
var scalarIDKind = makeNameKind("ID", "SCALAR")
var scalarIntKind = makeNameKind("Int", "SCALAR")
var nonNullKind = makeNameKind(nil, "NON_NULL")

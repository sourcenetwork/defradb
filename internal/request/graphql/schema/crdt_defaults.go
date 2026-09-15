// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import "github.com/sourcenetwork/defradb/client"

var defaultCRDTForFieldKind = map[client.FieldKind]client.CType{
	client.FieldKind_DocID:                   client.LWW_REGISTER,
	client.FieldKind_BOOL:                    client.LWW_REGISTER,
	client.FieldKind_NILLABLE_BOOL:           client.LWW_REGISTER,
	client.FieldKind_BOOL_ARRAY:              client.LWW_REGISTER,
	client.FieldKind_NILLABLE_BOOL_ARRAY:     client.LWW_REGISTER,
	client.FieldKind_INT:                     client.LWW_REGISTER,
	client.FieldKind_NILLABLE_INT:            client.LWW_REGISTER,
	client.FieldKind_INT_ARRAY:               client.LWW_REGISTER,
	client.FieldKind_NILLABLE_INT_ARRAY:      client.LWW_REGISTER,
	client.FieldKind_FLOAT64:                 client.LWW_REGISTER,
	client.FieldKind_NILLABLE_FLOAT64:        client.LWW_REGISTER,
	client.FieldKind_FLOAT64_ARRAY:           client.LWW_REGISTER,
	client.FieldKind_NILLABLE_FLOAT64_ARRAY:  client.LWW_REGISTER,
	client.FieldKind_FLOAT32:                 client.LWW_REGISTER,
	client.FieldKind_NILLABLE_FLOAT32:        client.LWW_REGISTER,
	client.FieldKind_FLOAT32_ARRAY:           client.LWW_REGISTER,
	client.FieldKind_NILLABLE_FLOAT32_ARRAY:  client.LWW_REGISTER,
	client.FieldKind_DATETIME:                client.LWW_REGISTER,
	client.FieldKind_NILLABLE_DATETIME:       client.LWW_REGISTER,
	client.FieldKind_DATETIME_ARRAY:          client.LWW_REGISTER,
	client.FieldKind_NILLABLE_DATETIME_ARRAY: client.LWW_REGISTER,
	client.FieldKind_STRING:                  client.LWW_REGISTER,
	client.FieldKind_NILLABLE_STRING:         client.LWW_REGISTER,
	client.FieldKind_STRING_ARRAY:            client.LWW_REGISTER,
	client.FieldKind_NILLABLE_STRING_ARRAY:   client.LWW_REGISTER,
	client.FieldKind_BLOB:                    client.LWW_REGISTER,
	client.FieldKind_NILLABLE_BLOB:           client.LWW_REGISTER,
	client.FieldKind_JSON:                    client.LWW_REGISTER,
	client.FieldKind_NILLABLE_JSON:           client.LWW_REGISTER,
}

// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"github.com/sourcenetwork/defradb/internal/encryption"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// we explicitly set LWW CRDT type because we want to test encryption with this specific CRDT type
// and we don't wat to rely on the default behavior
const userCollectionGQLSchema = (`
	type Users {
		name: String
		age: Int @crdt(type: "lww")
		verified: Boolean
	}
`)

func updateUserCollectionSchema() testUtils.SchemaUpdate {
	return testUtils.SchemaUpdate{
		Schema: userCollectionGQLSchema,
	}
}

func encrypt(plaintext []byte) []byte {
	val, _ := encryption.EncryptAES(plaintext, []byte("examplekey1234567890examplekey12"))
	return val
}

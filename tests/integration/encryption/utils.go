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

const (
	john21Doc = `{
		"name":	"John",
		"age":	21
	}`
	islam33Doc = `{
		"name":	"Islam",
		"age":	33
	}`
	john21DocID  = "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3"
	islam33DocID = "bae-d55bd956-1cc4-5d26-aa71-b98807ad49d6"
)

func updateUserCollectionSchema() testUtils.SchemaUpdate {
	return testUtils.SchemaUpdate{
		Schema: userCollectionGQLSchema,
	}
}

// encrypt encrypts the given plain text with a deterministic encryption key.
// We also want to make sure different keys are generated for different docs and fields
// and that's why we use the docID and fieldName to generate the key.
func encrypt(plaintext []byte, docID, fieldName string) []byte {
	const keyLength = 32
	const testEncKey = "examplekey1234567890examplekey12"
	val, _ := encryption.EncryptAES(plaintext, []byte(fieldName + docID + testEncKey)[0:keyLength])
	return val
}

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
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/tests/action"
)

// we explicitly set LWW CRDT type because we want to test encryption with this specific CRDT type
// and we don't wat to rely on the default behavior
const userCollectionGQLSchema = (`
	type Users {
		name: String
		age: Int @crdt(type: lww)
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
	john21DocID  = "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7"
	islam33DocID = "bae-5adc7327-0249-5925-bee7-52b370a4996d"
)

func updateUserCollectionSchema() *action.AddSchema {
	return &action.AddSchema{
		Schema: userCollectionGQLSchema,
	}
}

// encrypt encrypts the given plain text with a deterministic encryption key.
// We also want to make sure different keys are generated for different docs and fields
// and that's why we use the docID and fieldName to generate the key.
func encrypt(plaintext []byte, docID, fieldName string) []byte {
	const keyLength = 32
	const testEncKey = "examplekey1234567890examplekey12"
	val, _, _ := crypto.EncryptAES(plaintext, []byte(fieldName + docID + testEncKey)[0:keyLength], nil, true)
	return val
}

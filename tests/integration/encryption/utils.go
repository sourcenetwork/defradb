// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package encryption

import (
	"bytes"
	"fmt"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// we explicitly set LWW CRDT type because we want to test encryption with this specific CRDT type
// and we don't want to rely on the default behavior
const userCollection = (`
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
	john21DocID  = "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738"
	islam33DocID = "bae-0ee3406d-fe46-59d2-b2ce-618eeb24158f"
	counterDocID = "bae-c60ff298-7222-528f-920f-783ca0caeae1"
)

func addUserCollection() *action.AddCollection {
	return &action.AddCollection{
		SDL: userCollection,
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

type encryptedCBORValueMatcher struct {
	s            state.TestState
	plaintext    []byte
	plaintextDoc *action.DocIndex
	keyDocID     string
	fieldName    string
}

type notPlainCBORValueMatcher struct {
	s            state.TestState
	plaintext    []byte
	plaintextDoc *action.DocIndex
}

func encryptedCBORValueWithKey(
	plaintext []byte,
	keyDocID string,
	fieldName string,
) *encryptedCBORValueMatcher {
	return &encryptedCBORValueMatcher{
		plaintext: plaintext,
		keyDocID:  keyDocID,
		fieldName: fieldName,
	}
}

func genesisDocID(docID string) string {
	return id.NewGenesisDocID(docID)
}

func notPlainCBORValue(plaintext []byte) *notPlainCBORValueMatcher {
	return &notPlainCBORValueMatcher{plaintext: plaintext}
}

func notPlainCBORDocID(plaintextDoc action.DocIndex) *notPlainCBORValueMatcher {
	return &notPlainCBORValueMatcher{plaintextDoc: &plaintextDoc}
}

func (m *encryptedCBORValueMatcher) SetTestState(s state.TestState) {
	m.s = s
}

func (m *encryptedCBORValueMatcher) Match(actual any) (bool, error) {
	actualBytes, ok := actual.([]byte)
	if !ok {
		return false, fmt.Errorf("expected encrypted bytes, got %T", actual)
	}

	plaintext := m.plaintext
	if m.plaintextDoc != nil {
		docID := m.s.GetDocID(m.plaintextDoc.CollectionIndex, m.plaintextDoc.Index).String()
		plaintext = testUtils.CBORValue(docID)
	}

	return bytes.Equal(actualBytes, encrypt(plaintext, m.keyDocID, m.fieldName)), nil
}

func (m *encryptedCBORValueMatcher) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nto match encrypted CBOR value", actual)
}

func (m *encryptedCBORValueMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nnot to match encrypted CBOR value", actual)
}

func (m *notPlainCBORValueMatcher) SetTestState(s state.TestState) {
	m.s = s
}

func (m *notPlainCBORValueMatcher) Match(actual any) (bool, error) {
	actualBytes, ok := actual.([]byte)
	if !ok {
		return false, fmt.Errorf("expected bytes, got %T", actual)
	}

	plaintext := m.plaintext
	if m.plaintextDoc != nil {
		docID := m.s.GetDocID(m.plaintextDoc.CollectionIndex, m.plaintextDoc.Index).String()
		plaintext = testUtils.CBORValue(docID)
	}
	return !bytes.Equal(actualBytes, plaintext), nil
}

func (m *notPlainCBORValueMatcher) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nto differ from plaintext CBOR", actual)
}

func (m *notPlainCBORValueMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nnot to differ from plaintext CBOR", actual)
}

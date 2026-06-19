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
	"encoding/base64"
	"fmt"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/keys"
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

func docRefKey(collectionShortID uint32, docShortID uint32) string {
	return string(keys.EncodeDocRef(collectionShortID, docShortID))
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
	actualBytes, err := commitDeltaBytes(actual)
	if err != nil {
		return false, err
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
	actualBytes, err := commitDeltaBytes(actual)
	if err != nil {
		return false, err
	}

	plaintext := m.plaintext
	if m.plaintextDoc != nil {
		docID := m.s.GetDocID(m.plaintextDoc.CollectionIndex, m.plaintextDoc.Index).String()
		plaintext = testUtils.CBORValue(docID)
	}
	return !bytes.Equal(actualBytes, plaintext), nil
}

func commitDeltaBytes(actual any) ([]byte, error) {
	switch actual := actual.(type) {
	case []byte:
		return actual, nil
	case string:
		return base64.StdEncoding.DecodeString(actual)
	default:
		return nil, fmt.Errorf("expected bytes, got %T", actual)
	}
}

func (m *notPlainCBORValueMatcher) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nto differ from plaintext CBOR", actual)
}

func (m *notPlainCBORValueMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nnot to differ from plaintext CBOR", actual)
}

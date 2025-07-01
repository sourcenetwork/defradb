// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package signature

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/onsi/gomega/types"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

type signatureMatcher struct {
	s                 testUtils.TestState
	block             coreblock.Block
	expectedKeyType   crypto.KeyType
	castFailed        bool
	unexpectedKeyType immutable.Option[crypto.KeyType]
}

func newSignatureMatcher(block coreblock.Block, keyType crypto.KeyType) *signatureMatcher {
	return &signatureMatcher{
		block:           block,
		expectedKeyType: keyType,
	}
}

var _ types.GomegaMatcher = (*signatureMatcher)(nil)

func (matcher *signatureMatcher) SetTestState(s testUtils.TestState) {
	matcher.s = s
}

func (matcher *signatureMatcher) Match(actual any) (bool, error) {
	blockBytes, err := matcher.block.Marshal()
	if err != nil {
		return false, err
	}

	ident := matcher.s.GetIdentity(testUtils.NodeIdentity(matcher.s.GetCurrentNodeID()).Value())
	fullIdent, ok := ident.(acpIdentity.FullIdentity)
	if !ok {
		return false, fmt.Errorf("identity does not implement FullIdentity")
	}

	if fullIdent.PrivateKey().Type() != matcher.expectedKeyType {
		matcher.unexpectedKeyType = immutable.Some(fullIdent.PrivateKey().Type())
		return false, nil
	}

	expectedSigBytes, err := fullIdent.PrivateKey().Sign(blockBytes)
	if err != nil {
		return false, err
	}

	if matcher.s.GetClientType() == testUtils.GoClientType {
		actualSigBytes, ok := actual.([]byte)
		if !ok {
			matcher.castFailed = true
			return false, nil
		}
		return bytes.Equal(expectedSigBytes, actualSigBytes), nil
	} else {
		actualSigString, ok := actual.(string)
		if !ok {
			matcher.castFailed = true
			return false, nil
		}
		// CLI and HTTP clients return json response, so here we should expect a json string
		expectedSigBytes, err = json.Marshal(expectedSigBytes)
		if err != nil {
			return false, err
		}
		expectedSigString := strings.Trim(string(expectedSigBytes), "\"")
		return actualSigString == expectedSigString, nil
	}
}

func (matcher *signatureMatcher) FailureMessage(actual any) string {
	if matcher.castFailed {
		return fmt.Sprintf("Expected actual to be a byte slice, but got %T", actual)
	}
	if matcher.unexpectedKeyType.HasValue() {
		return fmt.Sprintf("Expected key type to be %s, but got %s",
			matcher.expectedKeyType, matcher.unexpectedKeyType.Value())
	}
	return "Expected signature to match"
}

func (matcher *signatureMatcher) NegatedFailureMessage(actual any) string {
	return "Expected signature not to match"
}

// identityMatcher is a matcher that matches an identity.
//
// This is used to match the identity of a node or a client.
//
// The identity is represented as a byte slice, which is the string representation of the public key of the identity.
type identityMatcher struct {
	s        testUtils.TestState
	identity testUtils.Identity
}

// newIdentityMatcher creates a new identity matcher.
//
// This is used to match the identity of a node or a client.
//
// The identity is represented as a byte slice, which is the string representation of the public key of the identity.
func newIdentityMatcher(ident testUtils.Identity) *identityMatcher {
	return &identityMatcher{
		identity: ident,
	}
}

var _ types.GomegaMatcher = (*identityMatcher)(nil)

func (matcher *identityMatcher) SetTestState(s testUtils.TestState) {
	matcher.s = s
}

func (matcher *identityMatcher) Match(actual any) (bool, error) {
	ident := matcher.s.GetIdentity(matcher.identity)

	actualString := ""
	if matcher.s.GetClientType() == testUtils.GoClientType {
		actualBytes, ok := actual.([]byte)
		if !ok {
			return false, fmt.Errorf("expected actual to be a byte slice, but got %T", actual)
		}
		actualString = string(actualBytes)
	} else {
		actualTmpString, ok := actual.(string)
		if !ok {
			return false, fmt.Errorf("expected actual to be a string, but got %T", actual)
		}

		// CLI and HTTP clients return json response, so here we should expect a json string
		var actualBytes []byte
		err := json.Unmarshal([]byte("\""+actualTmpString+"\""), &actualBytes)
		if err != nil {
			return false, err
		}

		actualString = string(actualBytes)
	}

	pubKey, err := crypto.PublicKeyFromString(ident.PublicKey().Type(), actualString)
	if err != nil {
		return false, errors.Wrap("failed to convert actual to public key", err)
	}
	return ident.PublicKey().Equal(pubKey), nil
}

func (matcher *identityMatcher) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected identity to match, but got %s", actual)
}

func (matcher *identityMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected identity not to match, but got %s", actual)
}

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

	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

type signatureMatcher struct {
	s          testUtils.TestState
	block      coreblock.Block
	castFailed bool
}

func newSignatureMatcher(block coreblock.Block) *signatureMatcher {
	return &signatureMatcher{
		block: block,
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

	ident := matcher.s.GetNodeIdentity(matcher.s.GetCurrentNodeID())

	expectedSigBytes, err := crypto.SignECDSA256K(ident.PrivateKey, blockBytes)
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
	return "Expected signature to match"
}

func (matcher *signatureMatcher) NegatedFailureMessage(actual any) string {
	return "Expected signature not to match"
}

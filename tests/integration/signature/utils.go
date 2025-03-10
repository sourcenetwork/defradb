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

	"github.com/onsi/gomega/types"

	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

type signatureMatcher struct {
	s               testUtils.TestState
	block           coreblock.Block
	bytesCastFailed bool
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

	sigBytes, err := crypto.SignECDSA256K(ident.PrivateKey, blockBytes)
	if err != nil {
		return false, err
	}

	actualBytes, ok := actual.([]byte)
	if !ok {
		matcher.bytesCastFailed = true
		return false, nil
	}

	return bytes.Equal(sigBytes, actualBytes), nil
}

func (matcher *signatureMatcher) FailureMessage(actual any) string {
	if matcher.bytesCastFailed {
		return "Expected actual to be a byte slice"
	}
	return "Expected signature to match"
}

func (matcher *signatureMatcher) NegatedFailureMessage(actual any) string {
	return "Expected signature not to match"
}

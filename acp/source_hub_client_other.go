// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package acp

import "github.com/sourcenetwork/sourcehub/sdk"

func NewSourceHubACP(
	chainID string,
	grpcAddress string,
	cometRPCAddress string,
	signer sdk.TxSigner,
) (ACP, error) {
	acpSourceHub, err := NewACPSourceHub(chainID, grpcAddress, cometRPCAddress, signer)
	if err != nil {
		return nil, err
	}

	return &sourceHubBridge{
		client:      acpSourceHub,
		supportsP2P: true,
	}, nil
}

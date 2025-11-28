// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package acp

import (
	"context"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
)

func canDACBypass(
	ctx context.Context,
	nodeACP NACInfo,
	identity string,
) bool {
	// Generally when NAC is not enabled, we allow the gated operations to work by assuming all
	// NAC permissions are granted, however allowing DAC bypass to work for everyone when NAC
	// is not enabled will defeat the purpose of having DAC, so don't bypass DAC in that case.
	if nodeACP.NodeACPDesc.Status != client.NACEnabled ||
		nodeACP.NodeACP == nil {
		return false
	}

	err := CheckNodeOperationAccess(
		ctx,
		identity,
		nodeACP,
		acpTypes.NodeDACBypassPerm,
		acpTypes.NodeACPObject,
	)

	return err == nil
}

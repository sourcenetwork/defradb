// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package acp utilizes the sourcehub acp module to bring the functionality to defradb, this package also helps
avoid the leakage of direct sourcehub references through out the code base, and eases in swapping
between local embedded use case and a more global on sourcehub use case.
*/

package acp

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/logging"
)

var (
	log = logging.MustNewLogger("acp")

	NoACPModule = immutable.None[ACPModule]()
)

type ACPModule interface {
	// Start initializes/starts the acp module, with persistence data
	// stored under the given root folder (most likely defradb root config path).
	//
	// If previously started in the same path, then recovers / reloads,
	// for example upon DB restart will restore the acp module state.
	Start(context.Context, string) error

	// Close closes/frees-up the resources in use by the acp module.
	Close() error

	// AddPolicy attempts to add/load/create the given policy, if isYAML is false, assumes JSON format,
	// upon success a policyID is returned, otherwise returns error.
	AddPolicy(ctx context.Context, creator string, policy string, isYAML bool) (string, error)

	// ValidatePolicyAndResourceExist returns an error if the policyID does not exist on the
	// acp module, or if the resource name is not a valid resource on the target policy,
	// otherwise if all checks out then returns no error (nil).
	ValidatePolicyAndResourceExist(ctx context.Context, policyID, resource string) error

	// RegisterDocCreation registers the document (object) to have access control.
	// No error is returned upon successful registering of a document.
	//
	// Note:
	// - This should be used upon document creation only.
	// - Some documents might be created without an identity signature so they would have public access.
	// - creator here is the actorID, which will be the signature identity if it exists.
	// - resource here is the resource object name (likely collection name).
	// - docID here is the object identifier.
	RegisterDocCreation(ctx context.Context, creator, policyID, resource, docID string) error

	// IsDocRegistered returns true if the document was registered with ACP. Otherwise returns false or an error.
	//
	// Note:
	// - resource here is the resource object name (likely collection name).
	// - docID here is the object identifier we want to see was registered.
	IsDocRegistered(ctx context.Context, policyID, resource, docID string) (bool, error)

	// CheckDocAccess returns true if request has access to the document, otherwise returns false or an error.
	//
	// Note:
	// - permission here is a valid DPI permission we are checking for ("read" or "write").
	// - resource here is the resource object name (likely collection name).
	// - docID here is the object identifier.
	CheckDocAccess(
		ctx context.Context,
		permission DPIPermission,
		actorID string,
		policyID string,
		resource string,
		docID string,
	) (bool, error)
}

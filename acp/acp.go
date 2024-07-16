// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

var (
	log = corelog.NewLogger("acp")

	// NoACP is an empty ACP, this is used to disable access control.
	NoACP = immutable.None[ACP]()
)

// ACP is the interface to all types of access control that might exist.
type ACP interface {
	// Init initializes the acp, with an absolute path. The provided path indicates where the
	// persistent data will be stored for acp.
	//
	// If the path is empty then acp will run in memory.
	Init(ctx context.Context, path string)

	// Start starts the acp, using the initialized path. Will recover acp state
	// from a previous run if under the same path.
	//
	// If the path is empty then acp will run in memory.
	Start(ctx context.Context) error

	// Close closes the resources in use by acp.
	Close() error

	// AddPolicy attempts to add the given policy. Detects the format of the policy automatically
	// by assuming YAML format if JSON validation fails. Upon success a policyID is returned,
	// otherwise returns error.
	//
	// A policy can not be added without a creator identity (sourcehub address).
	AddPolicy(ctx context.Context, creator identity.Identity, policy string) (string, error)

	// ValidateResourceExistsOnValidDPI performs DPI validation of the resource (matching resource name)
	// that is on the policy (matching policyID), returns an error upon validation failure.
	//
	// Learn more about the DefraDB Policy Interface [DPI](/acp/README.md)
	ValidateResourceExistsOnValidDPI(
		ctx context.Context,
		policyID string,
		resourceName string,
	) error

	// RegisterDocObject registers the document (object) to have access control.
	// No error is returned upon successful registering of a document.
	//
	// Note(s):
	// - This function does not check the collection to see if the document actually exists.
	// - Some documents might be created without an identity signature so they would have public access.
	// - actorID here is the identity of the actor registering the document object.
	RegisterDocObject(
		ctx context.Context,
		indentity identity.Identity,
		policyID string,
		resourceName string,
		docID string,
	) error

	// IsDocRegistered returns true if the document was found to be registered, otherwise returns false.
	// If check failed then an error and false will be returned.
	IsDocRegistered(
		ctx context.Context,
		policyID string,
		resourceName string,
		docID string,
	) (bool, error)

	// CheckDocAccess returns true if the check was successfull and the request has access to the document. If
	// the check was successful but the request does not have access to the document, then returns false.
	// Otherwise if check failed then an error is returned (and the boolean result should not be used).
	//
	// Note(s):
	// - permission here is a valid DPI permission we are checking for ("read" or "write").
	CheckDocAccess(
		ctx context.Context,
		permission DPIPermission,
		actorID string,
		policyID string,
		resourceName string,
		docID string,
	) (bool, error)
}

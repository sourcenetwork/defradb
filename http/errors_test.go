// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"net/http"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/db/p2p"
)

func TestHttpStatusFromError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected int
	}{
		// 401
		{"unauthorized", client.ErrNotAuthorizedToPerformOperation, http.StatusUnauthorized},

		// 403
		{"dev mode required", client.NewErrOperationRequiresDeveloperMode("test"), http.StatusForbidden},
		{"NAC op with NAC disabled", client.ErrCanNotDoThisNACOpWithNACIsDisabled, http.StatusForbidden},
		{"missing permission", db.ErrMissingPermission, http.StatusForbidden},
		{"resource missing required permission", acp.ErrResourceIsMissingRequiredPermission, http.StatusForbidden},

		// 404
		{"doc not found or not authorized", client.ErrDocumentNotFoundOrNotAuthorized, http.StatusNotFound},
		{"collection not found", client.ErrCollectionNotFound, http.StatusNotFound},
		{"doc ID not found", db.ErrDocIDNotFound, http.StatusNotFound},
		{"index does not exist", db.ErrIndexWithNameDoesNotExists, http.StatusNotFound},
		{"encrypted index does not exist", db.ErrEncryptedIndexDoesNotExist, http.StatusNotFound},
		{"collection root not found", db.ErrCollectionRootNotFound, http.StatusNotFound},
		{"lens CID not found", db.ErrLensCIDNotFound, http.StatusNotFound},
		{"replicator not found", p2p.ErrReplicatorNotFound, http.StatusNotFound},
		{"policy not found", acp.ErrPolicyDoesNotExistWithACP, http.StatusNotFound},
		{"resource not found on policy", acp.ErrResourceDoesNotExistOnTargetPolicy, http.StatusNotFound},

		// 409
		{"collection already exists", db.ErrCollectionAlreadyExists, http.StatusConflict},
		{"document already exists", db.ErrDocumentAlreadyExists, http.StatusConflict},
		{"index already exists", db.ErrIndexWithNameAlreadyExists, http.StatusConflict},
		{"encrypted index already exists", db.ErrEncryptedIndexAlreadyExists, http.StatusConflict},
		{"replicator exists", db.ErrReplicatorExists, http.StatusConflict},
		{"multiple active collection versions", db.ErrMultipleActiveCollectionVersions, http.StatusConflict},
		{"txn conflict", corekv.ErrTxnConflict, http.StatusConflict},

		// 422
		{"no policy without ACP", db.ErrCanNotHavePolicyWithoutACP, http.StatusUnprocessableEntity},
		{"materialized view and ACP", db.ErrMaterializedViewAndACPNotSupported, http.StatusUnprocessableEntity},
		{"col not materialized", db.ErrColNotMaterialized, http.StatusUnprocessableEntity},
		{"col mutating is branchable", db.ErrColMutatingIsBranchable, http.StatusUnprocessableEntity},
		{"p2p col has policy", db.ErrP2PColHasPolicy, http.StatusUnprocessableEntity},
		{"replicator col has policy", db.ErrReplicatorColHasPolicy, http.StatusUnprocessableEntity},
		{"collection name mutated", db.ErrCollectionNameMutated, http.StatusUnprocessableEntity},
		{"cannot delete old version", db.ErrCannotDeleteOldVersion, http.StatusUnprocessableEntity},
		{"migration between non-adjacent versions", db.ErrMigrationBetweenNonAdjacentVersions, http.StatusUnprocessableEntity},
		{"NAC already disabled", db.ErrNACIsAlreadyDisabled, http.StatusUnprocessableEntity},
		{"NAC already enabled", db.ErrNACIsAlreadyEnabled, http.StatusUnprocessableEntity},
		{"ACP not available", client.ErrACPOperationButACPNotAvailable, http.StatusUnprocessableEntity},

		// 503
		{"p2p disabled", ErrP2PDisabled, http.StatusServiceUnavailable},

		// 400 default
		{"unknown error", errors.New("something unknown"), http.StatusBadRequest},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, httpStatusFromError(c.err))
		})
	}
}

// TestHttpStatusFromError_ConstructedErrors verifies that errors created via NewErr constructors
// from internal/db match the exported sentinel vars. This guards against message string drift.
func TestHttpStatusFromError_ConstructedErrors(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected int
	}{
		{"document already exists", db.NewErrDocumentAlreadyExists("doc123"), http.StatusConflict},
		{"index already exists", db.NewErrIndexWithNameAlreadyExists("idx1"), http.StatusConflict},
		{"index does not exist", db.NewErrIndexWithNameDoesNotExists("idx1"), http.StatusNotFound},
		{"encrypted index already exists", db.NewErrEncryptedIndexAlreadyExists("field1"), http.StatusConflict},
		{"encrypted index does not exist", db.NewErrEncryptedIndexDoesNotExist("field1"), http.StatusNotFound},
		{"replicator exists", db.NewErrReplicatorExists("col1", peer.ID("peer1")), http.StatusConflict},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, httpStatusFromError(c.err),
				"sentinel may have drifted from internal/db — update the exported sentinel in internal/db/errors.go")
		})
	}
}

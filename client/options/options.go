// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package options provides option types for DefraDB client operations.
//
// This package provides struct-based options with fluent setter methods
// for configuring operations.
//
// Example usage:
//
//	// Store operation with identity
//	result, err := store.AddDACPolicy(ctx, policy,
//	    options.AddDACPolicy().SetIdentity(myIdentity))
//
//	// Collection operation with identity and encryption
//	err := collection.Create(ctx, doc,
//	    options.CollectionCreate().
//	        SetIdentity(myIdentity).
//	        SetEncryptDoc(true))
package options

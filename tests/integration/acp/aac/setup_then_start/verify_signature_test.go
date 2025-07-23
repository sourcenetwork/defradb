// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac_setup_then_start

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_GatesVerifySignaturePreSetup_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "admin acp correctly gates verify signature operation (setup before aac), allow if authorized, otherwise error",
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since the setup steps will be done before
				// the node is re-started with aac enabled (if it's in-memory it will loose setup state).
				testUtils.BadgerFileType,
			},
		),
		SupportedClientTypes: immutable.Some(
			[]testUtils.ClientType{
				// Default signer can be only tested with HTTP and CLI clients, because with Go client
				// when providing an identity, it includes the private key.
				testUtils.HTTPClientType,
				testUtils.CLIClientType,
			},
		),
		EnableSigning: true,
		Actions: []any{
			// Note: Since this is not an in-memory test, we can do the setup steps before aac is enabled.
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},

			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.NoIdentity(),
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreicwhd5s762awsrx6eowwqkkfpq7r5nnjosiru7blgaxo32wx6enp4",
				ExpectedError:  "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(2),
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreicwhd5s762awsrx6eowwqkkfpq7r5nnjosiru7blgaxo32wx6enp4",
				ExpectedError:  "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(1),
				SignerIdentity: testUtils.NodeIdentity(0).Value(),
				Cid:            "bafyreicwhd5s762awsrx6eowwqkkfpq7r5nnjosiru7blgaxo32wx6enp4",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_GatesVerifySignatureGo_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description:   "admin acp correctly gates verify signature operation (go client & setup before aac), allow if authorized, otherwise error",
		EnableSigning: true,
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since the setup steps will be done before
				// the node is re-started with aac enabled (if it's in-memory it will loose setup state).
				testUtils.BadgerFileType,
			},
		),
		SupportedClientTypes: immutable.Some(
			[]testUtils.ClientType{
				// Creating of signed documents over HTTP is not supported yet, because signing
				// requires a private key which we do not pass over HTTP.
				testUtils.GoClientType,
			},
		),
		Actions: []any{
			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			// Note: Doing setup steps after starting with aac enabled, otherwise the in-memory tests
			// will loose setup state when the restart happens (i.e. the restart that started aac).
			testUtils.SchemaUpdate{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  21,
				},
			},
			// Note: Setup is now done, the test code that follows is what we want to assert.

			// We haven't authorized non-identities. So, this should error.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.NoIdentity(),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreicwhd5s762awsrx6eowwqkkfpq7r5nnjosiru7blgaxo32wx6enp4",
				ExpectedError:  "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(2),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreicwhd5s762awsrx6eowwqkkfpq7r5nnjosiru7blgaxo32wx6enp4",
				ExpectedError:  "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			testUtils.VerifyBlockSignature{
				Identity:       testUtils.ClientIdentity(1),
				SignerIdentity: testUtils.ClientIdentity(1).Value(),
				Cid:            "bafyreicwhd5s762awsrx6eowwqkkfpq7r5nnjosiru7blgaxo32wx6enp4",
				ExpectedError:  "could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package wizard

const (
	// Keyboard keys
	KEY_CONTROL_C = "ctrl+c"
	KEY_ENTER     = "enter"
	KEY_SPACE     = " "
	KEY_UP        = "up"
	KEY_DOWN      = "down"
	KEY_K         = "k"
	KEY_J         = "j"

	TerminalClearANSICode = "\033[H\033[2J"

	// Step IDs
	BlankStepID                                   = "_blank_"
	stepWizardStartID                             = "stepWizardStart"
	stepConfigGeneratorID                         = "stepConfigGenerator"
	stepConfigGeneratedID                         = "stepConfigGenerated"
	stepKeyringStorageLocationID                  = "stepKeyringStorageLocation"
	stepKeyringStorageLocationBrancherID          = "stepKeyringStorageLocationBrancher"
	stepQueryGeneratingEnvironmentVariableID      = "stepQueryGeneratingEnvironmentVariable"
	stepGetDefraKeyringSecretInputID              = "stepGetDefraKeyringSecretInput"
	stepEnvironmentVariableGeneratedID            = "stepEnvironmentVariableGenerated"
	stepWizardExitMissingDefraKeyringSecretID     = "stepWizardExitMissingDefraKeyringSecret"
	stepQueryImportingKeysID                      = "stepQueryImportingKeys"
	stepQueryImportingIdentityKeyID               = "stepQueryImportingIdentityKey"
	stepQueryImportingIdentityKeyTypeID           = "stepQueryImportingIdentityKeyType"
	stepGettingIdentityKeyForImportID             = "stepGettingIdentityKeyForImport"
	stepImportingIdentityKeyID                    = "stepImportingIdentityKey"
	stepImportedIdentityKeyID                     = "stepImportedIdentityKey"
	stepGeneratedIdentityKeyID                    = "stepGeneratedIdentityKey"
	stepQueryImportingPeerKeyID                   = "stepQueryImportingPeerKey"
	stepGettingPeerKeyForImportID                 = "stepGettingPeerKeyForImport"
	stepImportedPeerKeyID                         = "stepImportedPeerKey"
	stepGeneratedPeerKeyID                        = "stepGeneratedPeerKey"
	stepQueryImportingEncryptionKeyID             = "stepQueryImportingEncryptionKey"
	stepGettingEncryptionKeyForImportID           = "stepGettingEncryptionKeyForImport"
	stepImportedEncryptionKeyID                   = "stepImportedEncryptionKey"
	stepGeneratedEncryptionKeyID                  = "stepGeneratedEncryptionKey"
	stepQueryImportingSearchableEncryptionKeyID   = "stepQueryImportingSearchableEncryptionKey"
	stepGettingSearchableEncryptionKeyForImportID = "stepGettingSearchableEncryptionKeyForImport"
	stepImportedSearchableEncryptionKeyID         = "stepImportedSearchableEncryptionKey"
	stepGeneratedSearchableEncryptionKeyID        = "stepGeneratedSearchableEncryptionKey"
	stepSelectKeyTypesID                          = "stepSelectKeyTypes"
	stepConfirmKeyringFilesGeneratedID            = "stepConfirmKeyringFilesGenerated"
	stepConfirmSystemKeyringKeysGeneratedID       = "stepConfirmSystemKeyringKeysGenerated"
)

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

	// The maximum amount of time to allow for the health check to complete
	HealthCheckTimeoutTimeInSeconds = 10

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
	BlankStepID                                = "_blank_"
	stepWizardStartID                          = "stepWizardStart"
	stepConfigGeneratorID                      = "stepConfigGenerator"
	stepConfigGeneratedID                      = "stepConfigGenerated"
	stepKeyringStorageLocationID               = "stepKeyringStorageLocation"
	stepKeyringStorageLocationBrancherID       = "stepKeyringStorageLocationBrancher"
	stepQueryGeneratingEnvironmentVariableID   = "stepQueryGeneratingEnvironmentVariable"
	stepGetDefraKeyringSecretInputID           = "stepGetDefraKeyringSecretInput"
	stepEnvironmentVariableGeneratedID         = "stepEnvironmentVariableGenerated"
	stepWizardExitMissingDefraKeyringSecretID  = "stepWizardExitMissingDefraKeyringSecret"
	stepQueryAddingKeysID                      = "stepQueryAddingKeys"
	stepQueryAddingIdentityKeyID               = "stepQueryAddingIdentityKey"
	stepQueryAddingIdentityKeyTypeID           = "stepQueryAddingIdentityKeyType"
	stepGettingIdentityKeyForAddID             = "stepGettingIdentityKeyForAdd"
	stepAddingIdentityKeyID                    = "stepAddingIdentityKey"
	stepAddedIdentityKeyID                     = "stepAddedIdentityKey"
	stepGeneratedIdentityKeyID                 = "stepGeneratedIdentityKey"
	stepQueryAddingPeerKeyID                   = "stepQueryAddingPeerKey"
	stepGettingPeerKeyForAddID                 = "stepGettingPeerKeyForAdd"
	stepAddedPeerKeyID                         = "stepAddedPeerKey"
	stepGeneratedPeerKeyID                     = "stepGeneratedPeerKey"
	stepQueryAddingEncryptionKeyID             = "stepQueryAddingEncryptionKey"
	stepGettingEncryptionKeyForAddID           = "stepGettingEncryptionKeyForAdd"
	stepAddedEncryptionKeyID                   = "stepAddedEncryptionKey"
	stepGeneratedEncryptionKeyID               = "stepGeneratedEncryptionKey"
	stepQueryAddingSearchableEncryptionKeyID   = "stepQueryAddingSearchableEncryptionKey"
	stepGettingSearchableEncryptionKeyForAddID = "stepGettingSearchableEncryptionKeyForAdd"
	stepAddedSearchableEncryptionKeyID         = "stepAddedSearchableEncryptionKey"
	stepGeneratedSearchableEncryptionKeyID     = "stepGeneratedSearchableEncryptionKey"
	stepSelectKeyTypesID                       = "stepSelectKeyTypes"
	stepConfirmKeyringFilesGeneratedID         = "stepConfirmKeyringFilesGenerated"
	stepConfirmSystemKeyringKeysGeneratedID    = "stepConfirmSystemKeyringKeysGenerated"
	stepQueryPerformingHealthCheckID           = "stepQueryPerformingHealthCheck"
	stepWillRunHealthcheckID                   = "stepWillRunHealthcheck"
	stepPerformHealthcheckID                   = "stepPerformHealthcheck"
	stepHealthcheckGoodID                      = "stepHealthcheckGood"
	stepSetupCompleteID                        = "stepSetupComplete"
)

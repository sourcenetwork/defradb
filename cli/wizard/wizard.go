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

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Context is the context for the wizard, and is used by the main model, and is passed
// to the callback functions so they have access to any information they need. This can be
// expanded as needed, but should be kept minimal.
type WizardContext struct {
	// Results is a map of step IDs to the results of the step. This can be accessed to retrieve
	// any of the results of previous steps that have occurred so far.
	Results map[string][]any

	// RootDir is the root directory of the defradb installation. This is part of the context
	// to allow easier integration with the unit tests, so that they can set a different, temporary
	// root directory for the test if needed.
	RootDir string
}

// Main is the entry point of the wizard, and is wired into the CLI's MakeWizardCommand() function.
func Main() {
	ctx := &WizardContext{
		Results: map[string][]any{},
		RootDir: getRootDir(),
	}

	// Define the steps
	stepWizardStart := initialModelMultipleChoice(
		stepWizardStartID,
		"You are about to run the DefraDB setup wizard. Do you wish to continue?",
		[]string{"Yes", "No"},
	)

	stepConfigGenerator := initialModelText(
		stepConfigGeneratorID,
		"A config.yaml file will be generated.",
	)

	stepConfigGenerated := initialModelText(
		stepConfigGeneratedID,
		"The config.yaml file was generated successfully.",
	)

	stepKeyringStorageLocation := initialModelMultipleChoice(
		stepKeyringStorageLocationID,
		"DefraDB protects the storage and transmission of data with a keypair that\n"+
			"will be generated now. You have the choice of where to store these generated keys.\n\n"+
			"Where do you want to store your keypair?",
		[]string{"Filesystem (~/.defradb/keys)", "OS (KeyChain)"},
	)

	stepKeyringStorageLocationBrancher := initialModelBrancher()

	stepQueryGeneratingEnvironmentVariable := initialModelMultipleChoice(
		stepQueryGeneratingEnvironmentVariableID,
		"To proceed, the DEFRA_KEYRING_SECRET environment variable must first be set.\n\n"+
			"Do you wish to generate a .env file containing it now?",
		[]string{"Yes", "No"},
	)

	stepGetDefraKeyringSecretInput := initialModelTextInput(
		stepGetDefraKeyringSecretInputID,
		"Please enter the DEFRA_KEYRING_SECRET value:",
		"my-secret-password",
	)

	stepEnvironmentVariableGenerated := initialModelText(
		stepEnvironmentVariableGeneratedID,
		"DEFRA_KEYRING_SECRET value was set in the .env file.",
	)

	stepWizardExitMissingDefraKeyringSecret := initialModelText(
		stepWizardExitMissingDefraKeyringSecretID,
		"Environment variable DEFRA_KEYRING_SECRET must be set to continue.\n\n"+
			"Please set the environment variable first and run the wizard again.\n\n"+
			"To set the environment variable, you can use the command: DEFRA_KEYRING_SECRET=my-secret-password\n\n"+
			"To run the wizard again you can use the command: defradb wizard",
	)

	stepQueryAddingKeys := initialModelMultipleChoice(
		stepQueryAddingKeysID,
		"Do you want to add any existing keys into the keyring?",
		[]string{"Yes", "No"},
	)

	stepQueryAddingIdentityKey := initialModelMultipleChoice(
		stepQueryAddingIdentityKeyID,
		"An identity key is required to be added or generated.\n"+
			"Do you want to add an existing identity key into the keyring?",
		[]string{"Yes, add an existing identity key", "No, generate a new identity key"},
	)

	stepQueryAddingIdentityKeyType := initialModelMultipleChoice(
		stepQueryAddingIdentityKeyTypeID,
		"What type of identity key do you want to add?",
		[]string{"Ed25519", "Secp256k1", "Secp256r1"},
	)

	stepGettingIdentityKeyForAdd := initialModelTextInput(
		stepGettingIdentityKeyForAddID,
		"Please enter the identity key you want to add:",
		"",
	)

	stepAddingIdentityKey := initialModelBlank()

	stepAddedIdentityKey := initialModelText(
		stepAddedIdentityKeyID,
		"Identity key added.",
	)

	stepGeneratingIdentityKey := initialModelBlank()

	stepGeneratedIdentityKey := initialModelText(
		stepGeneratedIdentityKeyID,
		"Identity key generated successfully.",
	)

	stepQueryAddingPeerKey := initialModelMultipleChoice(
		stepQueryAddingPeerKeyID,
		"Do you want to add an existing peer key into the keyring?",
		[]string{"Yes", "No"},
	)

	stepGettingPeerKeyForAdd := initialModelTextInput(
		stepGettingPeerKeyForAddID,
		"Please enter the peer key you want to add:",
		"",
	)

	stepAddingPeerKey := initialModelBlank()

	stepAddedPeerKey := initialModelText(
		stepAddedPeerKeyID,
		"Peer key added.",
	)

	stepGeneratingPeerKey := initialModelBlank()

	stepGeneratedPeerKey := initialModelText(
		stepGeneratedPeerKeyID,
		"Peer key generated successfully.",
	)

	stepQueryAddingEncryptionKey := initialModelMultipleChoice(
		stepQueryAddingEncryptionKeyID,
		"Do you want to add an existing encryption key into the keyring?",
		[]string{"Yes", "No"},
	)

	stepGettingEncryptionKeyForAdd := initialModelTextInput(
		stepGettingEncryptionKeyForAddID,
		"Please enter the encryption key you want to add:",
		"",
	)

	stepAddingEncryptionKey := initialModelBlank()

	stepAddedEncryptionKey := initialModelText(
		stepAddedEncryptionKeyID,
		"Encryption key added.",
	)

	stepGeneratingEncryptionKey := initialModelBlank()

	stepGeneratedEncryptionKey := initialModelText(
		stepGeneratedEncryptionKeyID,
		"Encryption key generated successfully.",
	)

	stepQueryAddingSearchableEncryptionKey := initialModelMultipleChoice(
		stepQueryAddingSearchableEncryptionKeyID,
		"Do you want to add an existing searchable encryption key into the keyring?",
		[]string{"Yes", "No"},
	)

	stepGettingSearchableEncryptionKeyForAdd := initialModelTextInput(
		stepGettingSearchableEncryptionKeyForAddID,
		"Please enter the searchable encryption key you want to add:",
		"",
	)

	stepAddingSearchableEncryptionKey := initialModelBlank()

	stepAddedSearchableEncryptionKey := initialModelText(
		stepAddedSearchableEncryptionKeyID,
		"Searchable encryption key added.",
	)

	stepGeneratingSearchableEncryptionKey := initialModelBlank()

	stepGeneratedSearchableEncryptionKey := initialModelText(
		stepGeneratedSearchableEncryptionKeyID,
		"Searchable encryption key generated successfully.",
	)

	stepSelectKeyTypes := initialModelToggleChoice(
		stepSelectKeyTypesID,
		"An identity key will be generated. Additionally, you may have this wizard generate the following"+
			" additional key types:",
		[]string{"Peer Key", "Encryption Key", "Searchable Encryption Key"},
	)

	stepGenerateKeyringFiles := initialModelBlank()
	stepGenerateSystemKeyringKeys := initialModelBlank()

	stepKeyringGenerationBrancher := initialModelBrancher()

	stepConfirmKeyringFilesGenerated := initialModelText(
		stepConfirmKeyringFilesGeneratedID,
		"Keyring file(s) generated successfully.",
	)

	stepConfirmSystemKeyringKeysGenerated := initialModelText(
		stepConfirmSystemKeyringKeysGeneratedID,
		"Key(s) generated in system keyring successfully.",
	)

	stepQueryPerformingHealthCheck := initialModelMultipleChoice(
		stepQueryPerformingHealthCheckID,
		"Do you want to test that DefraDB is configured correctly by "+
			"performing a health check?",
		[]string{"Yes", "No"},
	)

	stepWillRunHealthcheck := initialModelText(
		stepWillRunHealthcheckID,
		"A health check will be performed. This may take up to "+
			fmt.Sprintf("%d seconds", HealthCheckTimeoutTimeInSeconds)+
			" to complete.",
	)

	stepPerformHealthcheck := initialModelBlank()

	stepHealthcheckGood := initialModelText(
		stepHealthcheckGoodID,
		"DefraDB is configured and ready for use.",
	)

	stepSetupComplete := initialModelText(
		stepSetupCompleteID,
		"Setup is complete.",
	)

	// Chain the steps together
	stepWizardStart.nextSteps = []step{stepConfigGenerator, nil}
	stepConfigGenerator.nextStep = stepConfigGenerated
	stepConfigGenerated.nextStep = stepKeyringStorageLocation
	stepKeyringStorageLocation.nextSteps = []step{stepKeyringStorageLocationBrancher, stepQueryAddingKeys}
	stepKeyringStorageLocationBrancher.nextSteps = []step{
		stepQueryGeneratingEnvironmentVariable,
		stepQueryAddingKeys,
	}
	stepQueryGeneratingEnvironmentVariable.nextSteps = []step{
		stepGetDefraKeyringSecretInput,
		stepWizardExitMissingDefraKeyringSecret,
	}
	stepQueryAddingKeys.nextSteps = []step{stepQueryAddingIdentityKey, stepSelectKeyTypes}
	stepQueryAddingIdentityKey.nextSteps = []step{stepQueryAddingIdentityKeyType, stepGeneratingIdentityKey}
	stepQueryAddingIdentityKeyType.nextSteps = []step{
		stepGettingIdentityKeyForAdd,
		stepGettingIdentityKeyForAdd,
		stepGettingIdentityKeyForAdd,
	}
	stepGettingIdentityKeyForAdd.nextStep = stepAddingIdentityKey
	stepAddingIdentityKey.nextStep = stepAddedIdentityKey
	stepAddedIdentityKey.nextStep = stepQueryAddingPeerKey
	stepGeneratingIdentityKey.nextStep = stepGeneratedIdentityKey
	stepGeneratedIdentityKey.nextStep = stepQueryAddingPeerKey
	stepQueryAddingPeerKey.nextSteps = []step{stepGettingPeerKeyForAdd, stepQueryAddingEncryptionKey}
	stepGettingPeerKeyForAdd.nextStep = stepAddingPeerKey
	stepAddingPeerKey.nextStep = stepAddedPeerKey
	stepAddedPeerKey.nextStep = stepQueryAddingEncryptionKey
	stepGeneratingPeerKey.nextStep = stepGeneratedPeerKey
	stepGeneratedPeerKey.nextStep = stepQueryAddingEncryptionKey
	stepQueryAddingEncryptionKey.nextSteps = []step{
		stepGettingEncryptionKeyForAdd,
		stepQueryAddingSearchableEncryptionKey,
	}
	stepGettingEncryptionKeyForAdd.nextStep = stepAddingEncryptionKey
	stepAddingEncryptionKey.nextStep = stepAddedEncryptionKey
	stepAddedEncryptionKey.nextStep = stepQueryAddingSearchableEncryptionKey
	stepGeneratingEncryptionKey.nextStep = stepGeneratedEncryptionKey
	stepGeneratedEncryptionKey.nextStep = stepQueryAddingSearchableEncryptionKey
	stepQueryAddingSearchableEncryptionKey.nextSteps = []step{
		stepGettingSearchableEncryptionKeyForAdd,
		stepConfirmKeyringFilesGenerated,
	}
	stepGettingSearchableEncryptionKeyForAdd.nextStep = stepAddingSearchableEncryptionKey
	stepAddingSearchableEncryptionKey.nextStep = stepAddedSearchableEncryptionKey
	stepAddedSearchableEncryptionKey.nextStep = stepConfirmKeyringFilesGenerated
	stepGeneratingSearchableEncryptionKey.nextStep = stepGeneratedSearchableEncryptionKey
	stepGeneratedSearchableEncryptionKey.nextStep = stepConfirmKeyringFilesGenerated
	stepGenerateKeyringFiles.nextStep = stepConfirmKeyringFilesGenerated
	stepGenerateSystemKeyringKeys.nextStep = stepConfirmSystemKeyringKeysGenerated
	stepGetDefraKeyringSecretInput.nextStep = stepEnvironmentVariableGenerated
	stepEnvironmentVariableGenerated.nextStep = stepKeyringStorageLocationBrancher
	stepSelectKeyTypes.nextStep = stepKeyringGenerationBrancher
	stepKeyringGenerationBrancher.nextSteps = []step{stepGenerateKeyringFiles, stepGenerateSystemKeyringKeys}
	stepConfirmKeyringFilesGenerated.nextStep = stepQueryPerformingHealthCheck
	stepConfirmSystemKeyringKeysGenerated.nextStep = stepQueryPerformingHealthCheck
	stepQueryPerformingHealthCheck.nextSteps = []step{stepWillRunHealthcheck, stepSetupComplete}
	stepWillRunHealthcheck.nextStep = stepPerformHealthcheck
	stepPerformHealthcheck.nextStep = stepHealthcheckGood
	stepHealthcheckGood.nextStep = stepSetupComplete

	// Setup the callbacks
	stepKeyringStorageLocation.callback = callback_SetKeyringBackend
	stepConfigGenerator.callback = callback_GenerateConfigYAMLFile
	stepGenerateKeyringFiles.callback = callback_GenerateKeyringFiles
	stepGenerateSystemKeyringKeys.callback = callback_GenerateKeysInSystemKeyring
	stepGetDefraKeyringSecretInput.callback = callback_SetAndReloadDefraKeyringSecretEnvironmentVariable
	stepGeneratingIdentityKey.callback = callback_GenerateIdentityKey
	stepAddingIdentityKey.callback = callback_AddIdentityKey
	stepGeneratingPeerKey.callback = callback_GeneratePeerKey
	stepAddingPeerKey.callback = callback_AddPeerKey
	stepGeneratingEncryptionKey.callback = callback_GenerateEncryptionKey
	stepAddingEncryptionKey.callback = callback_AddEncryptionKey
	stepGeneratingSearchableEncryptionKey.callback = callback_GenerateSearchableEncryptionKey
	stepAddingSearchableEncryptionKey.callback = callback_AddSearchableEncryptionKey
	stepPerformHealthcheck.callback = callback_PerformHealthcheck

	// Setup the evaluators
	stepKeyringStorageLocationBrancher.evaluator = evaluator_IsEnvironmentVariableDefraKeyringSecretSet
	stepKeyringGenerationBrancher.evaluator = evaluator_ResultOfStepKeyringStorageLocation

	// Run the Bubbletea program
	program := tea.NewProgram(&mainModel{currentStep: stepWizardStart, ctx: ctx})
	if _, err := program.Run(); err != nil {
		os.Exit(1)
	}
}

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

	stepQueryImportingKeys := initialModelMultipleChoice(
		stepQueryImportingKeysID,
		"Do you want to import any existing keys into the keyring?",
		[]string{"Yes", "No"},
	)

	stepQueryImportingIdentityKey := initialModelMultipleChoice(
		stepQueryImportingIdentityKeyID,
		"An identity key is required to be imported or generated.\n"+
			"Do you want to import an existing identity key into the keyring?",
		[]string{"Yes, import an existing identity key", "No, generate a new identity key"},
	)

	stepQueryImportingIdentityKeyType := initialModelMultipleChoice(
		stepQueryImportingIdentityKeyTypeID,
		"What type of identity key do you want to import?",
		[]string{"Ed25519", "Secp256k1", "Secp256r1"},
	)

	stepGettingIdentityKeyForImport := initialModelTextInput(
		stepGettingIdentityKeyForImportID,
		"Please enter the identity key you want to import:",
		"",
	)

	stepImportingIdentityKey := initialModelBlank()

	stepImportedIdentityKey := initialModelText(
		stepImportedIdentityKeyID,
		"Identity key imported.",
	)

	stepGeneratingIdentityKey := initialModelBlank()

	stepGeneratedIdentityKey := initialModelText(
		stepGeneratedIdentityKeyID,
		"Identity key generated successfully.",
	)

	stepQueryImportingPeerKey := initialModelMultipleChoice(
		stepQueryImportingPeerKeyID,
		"Do you want to import an existing peer key into the keyring?",
		[]string{"Yes", "No"},
	)

	stepGettingPeerKeyForImport := initialModelTextInput(
		stepGettingPeerKeyForImportID,
		"Please enter the peer key you want to import:",
		"",
	)

	stepImportingPeerKey := initialModelBlank()

	stepImportedPeerKey := initialModelText(
		stepImportedPeerKeyID,
		"Peer key imported.",
	)

	stepGeneratingPeerKey := initialModelBlank()

	stepGeneratedPeerKey := initialModelText(
		stepGeneratedPeerKeyID,
		"Peer key generated successfully.",
	)

	stepQueryImportingEncryptionKey := initialModelMultipleChoice(
		stepQueryImportingEncryptionKeyID,
		"Do you want to import an existing encryption key into the keyring?",
		[]string{"Yes", "No"},
	)

	stepGettingEncryptionKeyForImport := initialModelTextInput(
		stepGettingEncryptionKeyForImportID,
		"Please enter the encryption key you want to import:",
		"",
	)

	stepImportingEncryptionKey := initialModelBlank()

	stepImportedEncryptionKey := initialModelText(
		stepImportedEncryptionKeyID,
		"Encryption key imported.",
	)

	stepGeneratingEncryptionKey := initialModelBlank()

	stepGeneratedEncryptionKey := initialModelText(
		stepGeneratedEncryptionKeyID,
		"Encryption key generated successfully.",
	)

	stepQueryImportingSearchableEncryptionKey := initialModelMultipleChoice(
		stepQueryImportingSearchableEncryptionKeyID,
		"Do you want to import an existing searchable encryption key into the keyring?",
		[]string{"Yes", "No"},
	)

	stepGettingSearchableEncryptionKeyForImport := initialModelTextInput(
		stepGettingSearchableEncryptionKeyForImportID,
		"Please enter the searchable encryption key you want to import:",
		"",
	)

	stepImportingSearchableEncryptionKey := initialModelBlank()

	stepImportedSearchableEncryptionKey := initialModelText(
		stepImportedSearchableEncryptionKeyID,
		"Searchable encryption key imported.",
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
	stepKeyringStorageLocation.nextSteps = []step{stepKeyringStorageLocationBrancher, stepQueryImportingKeys}
	stepKeyringStorageLocationBrancher.nextSteps = []step{
		stepQueryGeneratingEnvironmentVariable,
		stepQueryImportingKeys,
	}
	stepQueryGeneratingEnvironmentVariable.nextSteps = []step{
		stepGetDefraKeyringSecretInput,
		stepWizardExitMissingDefraKeyringSecret,
	}
	stepQueryImportingKeys.nextSteps = []step{stepQueryImportingIdentityKey, stepSelectKeyTypes}
	stepQueryImportingIdentityKey.nextSteps = []step{stepQueryImportingIdentityKeyType, stepGeneratingIdentityKey}
	stepQueryImportingIdentityKeyType.nextSteps = []step{
		stepGettingIdentityKeyForImport,
		stepGettingIdentityKeyForImport,
		stepGettingIdentityKeyForImport,
	}
	stepGettingIdentityKeyForImport.nextStep = stepImportingIdentityKey
	stepImportingIdentityKey.nextStep = stepImportedIdentityKey
	stepImportedIdentityKey.nextStep = stepQueryImportingPeerKey
	stepGeneratingIdentityKey.nextStep = stepGeneratedIdentityKey
	stepGeneratedIdentityKey.nextStep = stepQueryImportingPeerKey
	stepQueryImportingPeerKey.nextSteps = []step{stepGettingPeerKeyForImport, stepQueryImportingEncryptionKey}
	stepGettingPeerKeyForImport.nextStep = stepImportingPeerKey
	stepImportingPeerKey.nextStep = stepImportedPeerKey
	stepImportedPeerKey.nextStep = stepQueryImportingEncryptionKey
	stepGeneratingPeerKey.nextStep = stepGeneratedPeerKey
	stepGeneratedPeerKey.nextStep = stepQueryImportingEncryptionKey
	stepQueryImportingEncryptionKey.nextSteps = []step{
		stepGettingEncryptionKeyForImport,
		stepQueryImportingSearchableEncryptionKey,
	}
	stepGettingEncryptionKeyForImport.nextStep = stepImportingEncryptionKey
	stepImportingEncryptionKey.nextStep = stepImportedEncryptionKey
	stepImportedEncryptionKey.nextStep = stepQueryImportingSearchableEncryptionKey
	stepGeneratingEncryptionKey.nextStep = stepGeneratedEncryptionKey
	stepGeneratedEncryptionKey.nextStep = stepQueryImportingSearchableEncryptionKey
	stepQueryImportingSearchableEncryptionKey.nextSteps = []step{
		stepGettingSearchableEncryptionKeyForImport,
		stepConfirmKeyringFilesGenerated,
	}
	stepGettingSearchableEncryptionKeyForImport.nextStep = stepImportingSearchableEncryptionKey
	stepImportingSearchableEncryptionKey.nextStep = stepImportedSearchableEncryptionKey
	stepImportedSearchableEncryptionKey.nextStep = stepConfirmKeyringFilesGenerated
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
	stepImportingIdentityKey.callback = callback_ImportIdentityKey
	stepGeneratingPeerKey.callback = callback_GeneratePeerKey
	stepImportingPeerKey.callback = callback_ImportPeerKey
	stepGeneratingEncryptionKey.callback = callback_GenerateEncryptionKey
	stepImportingEncryptionKey.callback = callback_ImportEncryptionKey
	stepGeneratingSearchableEncryptionKey.callback = callback_GenerateSearchableEncryptionKey
	stepImportingSearchableEncryptionKey.callback = callback_ImportSearchableEncryptionKey
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

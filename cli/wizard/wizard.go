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
}

// Main is the entry point of the wizard, and is wired into the CLI's MakeWizardCommand() function.
func Main() {
	ctx := &WizardContext{
		Results: map[string][]any{},
	}

	// Define the steps
	stepWizardStart := initialModelMultipleChoice(
		"stepWizardStart",
		"You are about to run the DefraDB setup wizard. Do you wish to continue?",
		[]string{"Yes", "No"},
	)

	stepConfigGenerator := initialModelText(
		"stepConfigGenerator",
		"A config.yaml file will be generated.",
	)

	stepConfigGenerated := initialModelText(
		"stepConfigGenerated",
		"Config.yaml file generated successfully",
	)

	stepKeyringStorageLocation := initialModelMultipleChoice(
		"stepKeyringStorageLocation",
		"DefraDB protects the storage and transmission of data with a keypair that\n"+
			"will be generated now. You have the choice of where to store these generated keys.\n\n"+
			"Where do you want to store your keypair?",
		[]string{"Filesystem (~/.defradb/keys)", "OS (KeyChain)"},
	)

	stepKeyringStorageLocationBrancher := initialModelBrancher()

	stepQueryGeneratingEnvironmentVariable := initialModelMultipleChoice(
		"stepQueryGeneratingEnvironmentVariable",
		"To proceed, the DEFRA_KEYRING_SECRET environment variable must first be set.\n\n"+
			"Do you wish to generate a .env file containing it now?",
		[]string{"Yes", "No"},
	)

	stepGetDefraKeyringSecretInput := initialModelTextInput(
		"stepGetDefraKeyringSecretInput",
		"Please enter the DEFRA_KEYRING_SECRET value:",
		"my-secret-password",
	)

	stepEnvironmentVariableGenerated := initialModelText(
		"stepEnvironmentVariableGenerated",
		"DEFRA_KEYRING_SECRET value was set in the .env file.",
	)

	stepWizardExitMissingDefraKeyringSecret := initialModelText(
		"stepWizardExitMissingDefraKeyringSecret",
		"Environment variable DEFRA_KEYRING_SECRET must be set to continue.\n\n"+
			"Please set the environment variable first and run the wizard again.\n\n"+
			"To set the environment variable, you can use the command: DEFRA_KEYRING_SECRET=my-secret-password\n\n"+
			"To run the wizard again you can use the command: defradb wizard",
	)

	stepGenerateKeyringFiles := initialModelBlank()
	stepGenerateSystemKeyringKeys := initialModelBlank()

	stepConfirmKeyringFilesGenerated := initialModelText(
		"stepConfirmKeyringFilesGenerated",
		"Keyring files generated successfully.",
	)

	stepConfirmSystemKeyringKeysGenerated := initialModelText(
		"stepConfirmSystemKeyringKeysGenerated",
		"Keys generated in system keyring successfully.",
	)

	// Chain the steps together
	stepWizardStart.nextSteps = []step{stepConfigGenerator, nil}
	stepConfigGenerator.nextStep = stepConfigGenerated
	stepConfigGenerated.nextStep = stepKeyringStorageLocation
	stepKeyringStorageLocation.nextSteps = []step{stepKeyringStorageLocationBrancher, stepGenerateSystemKeyringKeys}
	stepKeyringStorageLocationBrancher.nextSteps = []step{
		stepQueryGeneratingEnvironmentVariable,
		stepGenerateKeyringFiles,
	}
	stepQueryGeneratingEnvironmentVariable.nextSteps = []step{
		stepGetDefraKeyringSecretInput,
		stepWizardExitMissingDefraKeyringSecret,
	}
	stepGenerateKeyringFiles.nextStep = stepConfirmKeyringFilesGenerated
	stepGenerateSystemKeyringKeys.nextStep = stepConfirmSystemKeyringKeysGenerated
	stepGetDefraKeyringSecretInput.nextStep = stepEnvironmentVariableGenerated
	stepEnvironmentVariableGenerated.nextStep = stepGenerateKeyringFiles

	// Setup the callbacks
	stepKeyringStorageLocation.callback = callback_SetKeyringBackend
	stepConfigGenerator.callback = callback_GenerateConfigYAMLFile
	stepGenerateKeyringFiles.callback = callback_GenerateKeyringFiles
	stepGenerateSystemKeyringKeys.callback = callback_GenerateKeyringFilesInSystemKeyring
	stepGetDefraKeyringSecretInput.callback = callback_SetAndReloadDefraKeyringSecretEnvironmentVariable

	// Setup the evaluators
	stepKeyringStorageLocationBrancher.evaluator = evaluator_IsEnvironmentVariableDefraKeyringSecretSet

	// Run the Bubbletea program
	program := tea.NewProgram(&mainModel{currentStep: stepWizardStart, ctx: ctx})
	if _, err := program.Run(); err != nil {
		os.Exit(1)
	}
}

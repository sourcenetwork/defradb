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

// This callback will set keyring.backend to either "file" or "system"
func callback_SetKeyringBackend(s step) {
	mm := s.(*modelMultipleChoice)

	choice := "file"
	if mm.cursor == 1 {
		choice = "system"
	}

	err := setYAMLValueInFile(getConfigFile(), []string{"keyring", "backend"}, choice)
	if err != nil {
		println("error setting YAML value:", err.Error())
		return
	}
}

// This callback prints a message saying that the setup is complete
func callback_PrintSetupComplete(_ step) {
	println("Setup complete")
}

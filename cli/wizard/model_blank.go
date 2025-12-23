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
	tea "github.com/charmbracelet/bubbletea"
)

// The purpose of a blank model, is that we can have a step that does not, itself,
// do anything, but which can still have a callback function. This can be useful to
// chain arbitraryfunction invocations into the step chain.
type modelBlank struct {
	// nextStep can be assigned dynamically. It should be set to the next step
	// to be executed after this step. nil is valid, and will be treated as the
	// end of the chain.
	nextStep step

	// callback is a function that will be called when this step is done.
	callback func(s step, ctx *WizardContext) error
}

// initialModelBlank should be called instead of manually constructing the struct
func initialModelBlank() *modelBlank {
	return &modelBlank{
		nextStep: nil,
		callback: nil,
	}
}

// Init() should not be called except by the main model
func (m *modelBlank) Init() tea.Cmd {
	return nil
}

// Update() should not be called except by the main model
func (m *modelBlank) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View() should not be called except by the main model
func (m *modelBlank) View() string {
	return ""
}

// Next() will return the next step, which may be nil
func (m *modelBlank) Next(_ *WizardContext) (step, error) {
	return m.nextStep, nil
}

// There is no result for this step
func (m *modelBlank) Result() any {
	return nil
}

// Because this step does not do anything, it is always done
func (m *modelBlank) Done() bool {
	return true
}

// Returns a dummy string
func (m *modelBlank) ID() string {
	return BlankStepID
}

// Callback() should not be called except by the main model
func (m *modelBlank) Callback(ctx *WizardContext) error {
	if m.callback == nil {
		return nil
	}
	return m.callback(m, ctx)
}

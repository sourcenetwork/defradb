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
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// For interface conformance.
// In the future, this can be removed, as this model type will be used in the main model.
var _ tea.Model = (*modelTextInput)(nil)

type modelTextInput struct {
	id     string         // Used to access the result of this step from the main model's results map
	prompt string         // The prompt to display to the user
	done   bool           // Whether the step is done
	input  textarea.Model // The text input model will manage input

	// nextStep can be assigned dynamically. It should be set to the next step
	// to be executed after this step. nil is valid, and will be treated as the
	// end of the chain.
	nextStep step

	// callback is a function that will be called when this step is done.
	callback func(s step, ctx *WizardContext) error
}

// initialModelTextInput should be called instead of manually constructing the struct
// This will be used in the main model in the future.
//
//nolint:unused
func initialModelTextInput(id string, prompt string, placeholder string) *modelTextInput {
	ti := textarea.New()
	ti.SetWidth(70)
	ti.SetHeight(1)
	ti.Placeholder = placeholder
	ti.ShowLineNumbers = false
	ti.Cursor.Blink = true
	ti.Focus()

	return &modelTextInput{
		id:     id,
		prompt: prompt,
		done:   false,
		input:  ti,
	}
}

// Init() should not be called except by the main model
func (m *modelTextInput) Init() tea.Cmd {
	return nil
}

// Update() should not be called except by the main model
func (m *modelTextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Check for a quick-quit
		case KEY_CONTROL_C:
			return m, tea.Quit
		// Enter submits the input
		case KEY_ENTER:
			m.done = true
		}
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View() should not be called except by the main model
func (m *modelTextInput) View() string {
	var s string

	// Prompt
	s += promptStyle.Render(m.prompt)
	s += "\n"

	ta := m.input.View()
	s += singleLineInputBox.Render(ta)

	s += "\n"

	// Hint
	s += hintStyle.Render("Enter to continue • Ctrl+C to quit")

	return s
}

// Next() will return the next step associated with the current cursor selection.
func (m *modelTextInput) Next() step {
	return m.nextStep
}

// The result is the text input by the user
func (m *modelTextInput) Result() any {
	return m.input.Value()
}

// This model tracks its own done state, which should at some point be set to true
// by the Update method.
func (m *modelTextInput) Done() bool {
	return m.done
}

// Returns the ID assigned during construction
func (m *modelTextInput) ID() string {
	return m.id
}

// Callback() should not be called except by the main model
func (m *modelTextInput) Callback(ctx *WizardContext) error {
	if m.callback == nil {
		return nil
	}
	return m.callback(m, ctx)
}

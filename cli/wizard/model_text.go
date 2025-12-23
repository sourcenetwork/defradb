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
	"github.com/charmbracelet/lipgloss"
)

type modelText struct {
	id   string // Used to access the result of this step from the main model's results map
	text string // The text to display to the user
	done bool   // Whether the step is done

	// nextStep can be assigned dynamically. It should be set to the next step
	// to be executed after this step. nil is valid, and will be treated as the
	// end of the chain.
	nextStep step

	// callback is a function that will be called when this step is done.
	callback func(s step, ctx *WizardContext) error

	// color refers to the lipgloss color attached to the text. A blank value indicates
	// no change to the default color.
	color string
}

// initialModelText should be called instead of manually constructing the struct
func initialModelText(id string, text string) *modelText {
	return &modelText{
		id:       id,
		text:     text,
		done:     false,
		nextStep: nil,
		callback: nil,
		color:    "",
	}
}

// Init() should not be called except by the main model
func (m *modelText) Init() tea.Cmd {
	return nil
}

// Update() should not be called except by the main model
func (m *modelText) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Check for a quick-quit
		case KEY_CONTROL_C:
			return m, tea.Quit

		// Enter and space proceed to the next step
		case KEY_ENTER, KEY_SPACE:
			m.done = true
		}
	}
	return m, nil
}

// View() should not be called except by the main model
func (m *modelText) View() string {
	var s string
	style := bodyStyle
	if m.color != "" {
		style = style.Foreground(lipgloss.Color(m.color))
	}
	s += style.Render(m.text)
	s += "\n"
	s += hintStyle.Render("Enter/Space to continue • Ctrl+C to quit")
	return s
}

// Next() will return the next step, which may be nil
func (m *modelText) Next(_ *WizardContext) (step, error) {
	return m.nextStep, nil
}

// There is no result for this step
func (m *modelText) Result() any {
	return ""
}

// This model tracks its own done state, which should at some point be set to true
// by the Update method.
func (m *modelText) Done() bool {
	return m.done
}

// Returns the ID assigned during construction
func (m *modelText) ID() string {
	return m.id
}

// Callback() should not be called except by the main model
func (m *modelText) Callback(ctx *WizardContext) error {
	if m.callback == nil {
		return nil
	}
	return m.callback(m, ctx)
}

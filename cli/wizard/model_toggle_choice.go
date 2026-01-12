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

type modelToggleChoice struct {
	id      string   // Used to access the result of this step from the main model's results map
	prompt  string   // The prompt to display to the user
	choices []string // The choices given to the user
	results []bool   // Whether each choice is selected or not
	cursor  int      // The current cursor position
	done    bool     // Whether the step is done

	// nextStep can be assigned dynamically. It should be set to the next step
	// to be executed after this step. nil is valid, and will be treated as the
	// end of the chain.
	nextStep step

	// callback is a function that will be called when this step is done.
	callback func(s step, ctx *WizardContext) error
}

// initialModelToggleChoice should be called instead of manually constructing the struct
func initialModelToggleChoice(id string, prompt string, choices []string) *modelToggleChoice {
	return &modelToggleChoice{
		id:       id,
		prompt:   prompt,
		choices:  choices,
		results:  make([]bool, len(choices)),
		cursor:   0,
		done:     false,
		nextStep: nil,
		callback: nil,
	}
}

// Init() should not be called except by the main model
func (m *modelToggleChoice) Init() tea.Cmd {
	return nil
}

// Update() should not be called except by the main model
func (m *modelToggleChoice) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Check for a quick-quit
		case KEY_CONTROL_C:
			return m, tea.Quit

		// Move the cursor up
		case KEY_UP, KEY_K:
			if m.cursor > 0 {
				m.cursor--
			}

		// Move the cursor down
		case KEY_DOWN, KEY_J:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// Space toggle the selection of the current choice
		case KEY_SPACE:
			m.results[m.cursor] = !m.results[m.cursor]

		// Enter select the current choice
		case KEY_ENTER:
			m.done = true
		}
	}

	return m, nil
}

// View() should not be called except by the main model
func (m *modelToggleChoice) View() string {
	var s string

	s += promptStyle.Render(m.prompt) + "\n"

	// Iterate over the choices
	for i, choice := range m.choices {
		cursor := " "
		rowStyle := choiceStyle
		openBracket := defaultStyle.Render("[")
		closeBracket := defaultStyle.Render("]")

		// Draw the cursor in front of the selected choice
		if m.cursor == i {
			cursor = ">"
			rowStyle = selectedChoiceStyle
			// Colorize the brackets of the selected choice
			openBracket = defraBlueStyle.Render("[")
			closeBracket = defraBlueStyle.Render("]")
		}

		// Create the toggle marker for the choice
		selectionToggle := " "
		if m.results[i] {
			selectionToggle = "●"
		}

		// Render each part of the row into the full line
		cursorRendered := rowStyle.Render(cursor)
		choiceRendered := rowStyle.Render(choice)
		toggleRendered := toggleSelectionMarker.Render(selectionToggle)
		line := cursorRendered + " " + openBracket + toggleRendered + closeBracket + choiceRendered

		s += line + "\n"
	}
	s += hintStyle.Render("↑/↓ to move cursor • Space to toggle selection • Enter to continue • Ctrl+C to quit")
	return s
}

// Next() will return the next step associated with the current cursor selection.
func (m *modelToggleChoice) Next(_ *WizardContext) (step, error) {
	return m.nextStep, nil
}

// The result is the model's results array
func (m *modelToggleChoice) Result() any {
	return m.results
}

// This model tracks its own done state, which should at some point be set to true
// by the Update method.
func (m *modelToggleChoice) Done() bool {
	return m.done
}

// Returns the ID assigned during construction
func (m *modelToggleChoice) ID() string {
	return m.id
}

// Callback() should not be called except by the main model
func (m *modelToggleChoice) Callback(ctx *WizardContext) error {
	if m.callback == nil {
		return nil
	}
	return m.callback(m, ctx)
}

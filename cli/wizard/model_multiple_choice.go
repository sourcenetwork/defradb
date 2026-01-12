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

	tea "github.com/charmbracelet/bubbletea"
)

// The purpose of a multiple choice model, is that we can have a step that allows the user to
// select one out of a list of choices. The next step will be determined by the choice made,
// and the result will be available for retrieval later.
type modelMultipleChoice struct {
	id      string   // Used to access the result of this step from the main model's results map
	prompt  string   // The prompt to display to the user
	choices []string // The choices given to the user
	cursor  int      // The current cursor position
	done    bool     // Whether the step is done

	// nextSteps can be assigned dynamically. It should be set to a slice of steps,
	// in order of the choices. For example, if the choices are ["Yes", "No"], then
	// the nextSteps value should be set to []step{step2, step2}, corresponding to
	// how the next choice should be branched to. nil is valid, and will be treated
	// as the end of the chain.
	nextSteps []step

	// callback is a function that will be called when this step is done.
	callback func(s step, ctx *WizardContext) error
}

// initialModelMultipleChoice should be called instead of manually constructing the struct
func initialModelMultipleChoice(id string, prompt string, choices []string) *modelMultipleChoice {
	return &modelMultipleChoice{
		id:        id,
		prompt:    prompt,
		choices:   choices,
		cursor:    0,
		done:      false,
		nextSteps: nil,
		callback:  nil,
	}
}

// Init() should not be called except by the main model
func (m *modelMultipleChoice) Init() tea.Cmd {
	return nil
}

// Update() should not be called except by the main model
func (m *modelMultipleChoice) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		// Enter and space select the current choice
		case KEY_ENTER, KEY_SPACE:
			m.done = true
		}
	}

	return m, nil
}

// View() should not be called except by the main model
func (m *modelMultipleChoice) View() string {
	var s string

	s += promptStyle.Render(m.prompt) + "\n"

	// Iterate over our choices
	for i, choice := range m.choices {
		// Draw the cursor in front of the selected choice
		cursor := " "
		style := choiceStyle
		if m.cursor == i {
			cursor = ">"
			style = selectedChoiceStyle
		}

		// Render the row
		line := fmt.Sprintf("%s %s", cursor, choice)
		s += style.Render(line) + "\n"
	}

	s += hintStyle.Render("↑/↓ to move cursor • Enter/Space to select • Ctrl+C to quit")
	return s
}

// Next() will return the next step associated with the current cursor selection.
func (m *modelMultipleChoice) Next(_ *WizardContext) (step, error) {
	if len(m.nextSteps) == 0 {
		return nil, nil
	}
	if m.cursor < len(m.nextSteps) {
		return m.nextSteps[m.cursor], nil
	}
	return nil, nil
}

// The result is the current cursor selection
func (m *modelMultipleChoice) Result() any {
	return m.cursor
}

// This model tracks its own done state, which should at some point be set to true
// by the Update method.
func (m *modelMultipleChoice) Done() bool {
	return m.done
}

// Returns the ID assigned during construction
func (m *modelMultipleChoice) ID() string {
	return m.id
}

// Callback() should not be called except by the main model
func (m *modelMultipleChoice) Callback(ctx *WizardContext) error {
	if m.callback == nil {
		return nil
	}
	return m.callback(m, ctx)
}

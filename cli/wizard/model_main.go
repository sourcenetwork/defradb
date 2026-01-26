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

// A "step" is a single step in the wizard, each of which is its own
// sub-model responsible for its UI and logic.
type step interface {
	tea.Model

	// ID() must return the ID of the step. This will be used to store the result
	// of the step in the main model's results map.
	ID() string

	// Done() must return whether the step is complete. It will be checked
	// by the main model to determine when to move to the next step.
	Done() bool

	// Next() must return which step to move to. It can be dynamic, reflecting
	// internal logic of the step. But what must be true, is that at the time of
	// the Done() method resolving to true, the Next() method must resolve to the
	// next step in the chain, or to nil.
	Next(ctx *WizardContext) (step, error)

	// Result() must return the result of the step. Like Next(), it can be
	// dynamic, but what must be true, is that at the time of the Done() method
	// resolving to true, the Result() method must be resolvable to a final value.
	Result() any

	// Callback() must return a function that will be called when the step is done.
	// It can be nil, in which case the step will not call any callback when it is done.
	Callback(ctx *WizardContext) error
}

// The main model is the top-level model that runs the wizard. It will track the
// current step that it is on, as well as store the results of each step as they
// are completed.
type mainModel struct {
	currentStep step
	done        bool
	ctx         *WizardContext
}

// Delegate responsibility of Init to the current step
func (m *mainModel) Init() tea.Cmd {
	return m.currentStep.Init()
}

// Update method handles the logic of the wizard application
func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var updatedModel tea.Model

	// In the case of the window size changing, clear the terminal to
	// prevent visual artifacts, or garbled text.
	switch msg.(type) {
	case tea.WindowSizeMsg:
		// Clear the terminal by sending ANSI escape code
		printToTerminal(TerminalClearANSICode)
		return m, nil
	}

	// Update the current step

	// If the current step has become nil, we are done and must quit
	if m.currentStep == nil {
		return m, tea.Quit
	}

	updatedModel, cmd = m.currentStep.Update(msg)
	m.currentStep = updatedModel.(step) //nolint:forcetypeassert
	// If the step is done...
	for m.currentStep != nil && m.currentStep.Done() {
		// ... call its callback, store the results, and move onto the next step
		m.ctx.Results[m.currentStep.ID()] = append(m.ctx.Results[m.currentStep.ID()], m.currentStep.Result())
		err := m.currentStep.Callback(m.ctx)
		// If the callback returns an error, we must gracefully proceed to an error-step exit
		if err != nil {
			errorStep := createErrorStep(err)
			m.currentStep = errorStep
			return m, cmd
		}
		// If getting the next step fails, generate an error step
		next, err := m.currentStep.Next(m.ctx)
		if err != nil {
			errorStep := createErrorStep(err)
			m.currentStep = errorStep
			return m, cmd
		}

		// Movethrough blank steps, calling their callbacks, until we reach a non-blank step
		for next != nil && next.ID() == BlankStepID {
			m.ctx.Results[next.ID()] = append(m.ctx.Results[next.ID()], next.Result())
			err := next.Callback(m.ctx)
			// If any of the callbacks return an error, we must gracefully proceed to an error-step exit
			if err != nil {
				errorStep := createErrorStep(err)
				m.currentStep = errorStep
				return m, cmd
			}
			// Get the next step, erroring out if it fails
			next, err = next.Next(m.ctx)
			if err != nil {
				errorStep := createErrorStep(err)
				m.currentStep = errorStep
				return m, cmd
			}
		}
		m.currentStep = next
	}

	// If there are no more steps, we are done, and can quit
	if m.currentStep == nil {
		m.done = true
		return m, tea.Quit
	}

	return m, cmd
}

// Delegate responsibility of View to the current step, with a bit of simple logic to check if
// we are done, and to wrap a quit message below it.
func (m *mainModel) View() string {
	if m.done {
		return "\n"
	}
	return "\n" + m.currentStep.View()
}

// createErrorStep creates a new step that displays an error message and allows the user to exit
// the wizard. This can be called by the main model upon encountering an error from a callback.
func createErrorStep(err error) step {
	errStep := initialModelText(
		"errorStep",
		"An error occurred: "+err.Error(),
	)
	errStep.color = defraRed
	return errStep
}

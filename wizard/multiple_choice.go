package wizard

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

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
}

// initialModelMultipleChoice should be called instead of manually constructing the struct
func initialModelMultipleChoice(id string, prompt string, choices []string) modelMultipleChoice {
	return modelMultipleChoice{
		id:      id,
		prompt:  prompt,
		choices: choices,
		cursor:  0,
		done:    false,
	}
}

// Init should not be called except by the main model
func (m modelMultipleChoice) Init() tea.Cmd {
	return nil
}

// Update should not be called except by the main model
func (m modelMultipleChoice) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		// Check for a quick-quit
		case "ctrl+c", "q":
			return m, tea.Quit

		// Move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// Move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// Enter and space select the current choice
		case "enter", " ":
			m.done = true
			return m, nil
		}
	}

	return m, nil
}

// View should not be called except by the main model
func (m modelMultipleChoice) View() string {
	s := m.prompt + "\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Draw the cursor in front of the selected choice
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	return s
}

// Next() will return the next step associated with the current cursor selection.
func (m modelMultipleChoice) Next() step {
	if m.nextSteps == nil || len(m.nextSteps) == 0 {
		return nil
	}
	if m.cursor < len(m.nextSteps) {
		return m.nextSteps[m.cursor]
	}
	return nil
}

// The result is the current cursor selection
func (m modelMultipleChoice) Result() any {
	return m.cursor
}

// This model tracks its own done state, which should at some point be set to true
// by the Update method.
func (m modelMultipleChoice) Done() bool {
	return m.done
}

// Returns the ID assigned during construction
func (m modelMultipleChoice) ID() string {
	return m.id
}

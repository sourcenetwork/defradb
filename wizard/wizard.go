package wizard

import (
	"fmt"
	"os"

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

	// Next() must return which step to move to. It can be dynaamic, reflecting
	// internal logic of the step. But what must be true, is that at the time of
	// the Done() method resolving to true, the Next() method must resolve to the
	// next step in the chain, or to nil.
	Next() step

	// Result() must return the result of the step. Like Next(), it can be
	// dynamic, but what must be true, is that at the time of the Done() method
	// resolving to true, the Result() method must be resolvable to a final value.
	Result() any
}

// The main model is the top-level model that runs the wizard. It will track the
// current step that it is on, as well as store the results of each step as they
// are completed.
type mainModel struct {
	currentStep step
	done        bool
	results     map[string][]any // Map ID of the step to the result of it
}

// Delegate responsibility of Init to the current step
func (m mainModel) Init() tea.Cmd {
	return m.currentStep.Init()
}

// Update method handles the logic of the wizard application
func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	var cmd tea.Cmd
	var updatedModel tea.Model

	// Update the current step
	updatedModel, cmd = m.currentStep.Update(msg)
	m.currentStep = updatedModel.(step)

	// Check if the step is done, and if so, store the result and move on to the next step
	if m.currentStep.Done() {
		next := m.currentStep.Next()
		m.results[m.currentStep.ID()] = append(m.results[m.currentStep.ID()], m.currentStep.Result())
		if next != nil {
			m.currentStep = next
		} else {
			// If there is no next step, the wizard is complete
			m.done = true
		}
	}

	return m, cmd
}

// Delegate responsibility of View to the current step, with a bit of simple logic to check if
// we are done, and to wrap a quit message below it.
func (m mainModel) View() string {
	if m.done {
		s := "Wizard complete\n"
		return s
	}
	return m.currentStep.View() + "\nYou can press Q at any time to exit this wizard.\n"
}

// Main is the entry point of the wizard, and is wired into the CLI's MakeWizardCommand() function.
func Main() {
	// Define the steps
	step1 := initialModelMultipleChoice(
		"step1",
		"You are about to run the DefraDB setup wizard. Do you wish to continue?:",
		[]string{"Yes", "No"},
	)

	step2 := initialModelMultipleChoice(
		"step2",
		"DefraDB protects the storage and transmission of data with a keypair that "+
			" will be generated now. You have the choice of where to store these generated keys.\n\n"+
			" Where do you want to store your keypair?",
		[]string{"Filesystem (~/.defradb/keys)", "OS (KeyChain, etc)"},
	)

	// Chain the steps together
	step1.nextSteps = []step{step2, nil}
	step2.nextSteps = []step{nil, nil}

	// Run the Bubbletea program
	program := tea.NewProgram(mainModel{currentStep: step1, results: make(map[string][]any)})
	if _, err := program.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

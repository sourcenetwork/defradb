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
	"github.com/charmbracelet/lipgloss"
)

var (
	sourceGreen = "#39E265"
	defraBlue   = "#10CBFF"
	defraRed    = "#E25647"

	defaultStyle = lipgloss.NewStyle()

	defraBlueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(defraBlue))

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(sourceGreen)).
			MarginBottom(1).
			Width(80)

	choiceStyle = lipgloss.NewStyle().
			PaddingLeft(1)

	selectedChoiceStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Bold(true).
				Foreground(lipgloss.Color(defraBlue))

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(sourceGreen)).
			PaddingLeft(1).
			Width(80)

	hintStyle = lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color("240")).
			MarginTop(1)

	singleLineInputBox = lipgloss.NewStyle().
				Height(1).
				Width(80).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(sourceGreen)).
				Padding(0, 1)

	toggleSelectionMarker = lipgloss.NewStyle().
				Foreground(lipgloss.Color(sourceGreen))
)

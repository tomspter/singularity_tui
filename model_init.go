package main

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

func newModel(ds dataSource) model {
	keys := keyMap{
		PrevPartition: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "prev partition"),
		),
		NextPartition: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "next partition"),
		),
		SortState: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort state"),
		),
		SortCPU: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "sort cpu"),
		),
		SortMem: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "sort mem"),
		),
		SortGPU: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "sort gpu"),
		),
		NodeDetail: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "node detail"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}

	h := help.New()
	h.ShowAll = false

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "Node", Width: 24},
			{Title: "State", Width: 14},
			{Title: "CPU(A/T)", Width: 12},
			{Title: "MEM(A/T)", Width: 16},
			{Title: "GPU(A/T)", Width: 14},
		}),
		table.WithRows([]table.Row{}),
		table.WithHeight(12),
		table.WithWidth(100),
		table.WithFocused(true),
	)
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		Bold(true).
		Foreground(lipgloss.Color(colorFgPrimary))
	styles.Cell = styles.Cell.
		Foreground(lipgloss.Color(colorFgSecondary))
	styles.Selected = styles.Cell
	t.SetStyles(styles)

	cpuBar := progress.New(progress.WithColors(lipgloss.Color(colorProgressCPU)))
	memBar := progress.New(progress.WithColors(lipgloss.Color(colorProgressMEM)))
	gpuBar := progress.New(progress.WithColors(lipgloss.Color(colorProgressGPU)))
	cpuBar.ShowPercentage = false
	memBar.ShowPercentage = false
	gpuBar.ShowPercentage = false
	cpuBar.EmptyColor = lipgloss.Color(colorProgressBase)
	memBar.EmptyColor = lipgloss.Color(colorProgressBase)
	gpuBar.EmptyColor = lipgloss.Color(colorProgressBase)
	cpuBar.SetWidth(progressWidth)
	memBar.SetWidth(progressWidth)
	gpuBar.SetWidth(progressWidth)

	return model{
		ds:          ds,
		table:       t,
		help:        h,
		keys:        keys,
		loading:     true,
		activeTab:   0,
		sortMode:    sortByState,
		tableHeight: 12,
		nodeColW:    24,
		stateColW:   14,
		cpuColW:     12,
		memColW:     16,
		gpuColW:     14,
		cpuBar:      cpuBar,
		memBar:      memBar,
		gpuBar:      gpuBar,
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorAccent)),
		subtitleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgMuted)),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorStateDrain)),
		mutedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgMuted)),
		separatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSeparator)),
		detailStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgPrimary)),
		popupStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorAccent)).
			Background(lipgloss.Color(colorPanel)).
			Padding(1, 2),
		activeTabStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgPrimary)).
			Background(lipgloss.Color(colorAccent)).
			Padding(0, 1),
		tabStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgSecondary)).
			Background(lipgloss.Color(colorPanelAlt)).
			Padding(0, 1),
	}
}

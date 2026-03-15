package main

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchDataCmd(m.ds), tickCmd())
}

func fetchDataCmd(ds dataSource) tea.Cmd {
	return func() tea.Msg {
		summaries, err := ds.Fetch(context.Background())
		if err != nil {
			return errMsg{err: err}
		}
		user := fetchUserSummary(context.Background())
		return dataMsg{
			partitions: summaries,
			user:       user,
			loadedAt:   time.Now(),
		}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg {
		return refreshMsg{}
	})
}

func fetchNodeDetailCmd(node string) tea.Cmd {
	return func() tea.Msg {
		out, err := runCommand(context.Background(), "scontrol", "show", "node", node)
		if err != nil {
			return nodeDetailErrMsg{node: node, err: err}
		}
		return nodeDetailMsg{
			node:   node,
			detail: strings.TrimSpace(out),
		}
	}
}

func cancelJobCmd(jobID string) tea.Cmd {
	return func() tea.Msg {
		_, err := runCommand(context.Background(), "scancel", jobID)
		return userCancelResultMsg{jobID: jobID, err: err}
	}
}

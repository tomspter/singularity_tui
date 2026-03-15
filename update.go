package main

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableColumns()
		m.refreshTableRows(false)
		m.updateProgressWidth()
		m.updateTableHeight()
		return m, nil
	case tea.KeyPressMsg:
		if m.userActionOpen {
			switch msg.String() {
			case "esc", "n":
				m.userActionOpen = false
				m.userActionBusy = false
				return m, nil
			case "enter", "y":
				if m.userActionCanCancel && !m.userActionBusy {
					m.userActionBusy = true
					m.userActionMsg = "Running scancel ..."
					return m, cancelJobCmd(m.userActionJob.JobID)
				}
				m.userActionOpen = false
				m.userActionBusy = false
				return m, nil
			default:
				return m, nil
			}
		}

		if m.detailOpen {
			switch {
			case msg.String() == "t", msg.String() == "esc", key.Matches(msg, m.keys.NodeDetail):
				m.detailOpen = false
				m.detailBusy = false
				m.detailErr = nil
				return m, nil
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			return m, fetchDataCmd(m.ds)
		case key.Matches(msg, m.keys.PrevPartition):
			m.switchTab(-1)
			return m, nil
		case key.Matches(msg, m.keys.NextPartition):
			m.switchTab(1)
			return m, nil
		case key.Matches(msg, m.keys.SortState):
			m.sortMode = sortByState
			m.updateTableColumns()
			m.refreshTableRows(true)
			return m, nil
		case key.Matches(msg, m.keys.SortCPU):
			m.sortMode = sortByCPU
			m.updateTableColumns()
			m.refreshTableRows(true)
			return m, nil
		case key.Matches(msg, m.keys.SortMem):
			m.sortMode = sortByMEM
			m.updateTableColumns()
			m.refreshTableRows(true)
			return m, nil
		case key.Matches(msg, m.keys.SortGPU):
			m.sortMode = sortByGPU
			m.updateTableColumns()
			m.refreshTableRows(true)
			return m, nil
		case key.Matches(msg, m.keys.UserCancel):
			if m.isUserTab() {
				m.openUserActionDialog()
				return m, nil
			}
			return m, nil
		case msg.String() == "t" || key.Matches(msg, m.keys.NodeDetail):
			return m, m.toggleNodeDetail()
		case key.Matches(msg, m.keys.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll
			m.updateTableHeight()
			return m, nil
		}
	case refreshMsg:
		m.loading = true
		return m, tea.Batch(fetchDataCmd(m.ds), tickCmd())
	case dataMsg:
		prevPartitionName := ""
		prevIsUserTab := m.isUserTab()
		if p := m.selectedPartition(); p != nil {
			prevPartitionName = p.Name
		}
		m.loading = false
		m.lastErr = nil
		m.partitions = msg.partitions
		m.userSummary = msg.user
		m.refreshUserList(true)
		m.lastUpdated = msg.loadedAt
		if prevIsUserTab {
			m.activeTab = len(m.partitions)
		} else if prevPartitionName != "" {
			found := -1
			for i := range m.partitions {
				if m.partitions[i].Name == prevPartitionName {
					found = i
					break
				}
			}
			if found >= 0 {
				m.activeTab = found
			} else if len(m.partitions) > 0 && m.activeTab >= len(m.partitions) {
				m.activeTab = len(m.partitions) - 1
			}
		} else if len(m.partitions) > 0 && m.activeTab >= len(m.partitions) {
			m.activeTab = len(m.partitions) - 1
		} else if len(m.partitions) == 0 {
			m.activeTab = 0 // User tab index when no partitions.
		}
		m.updateTableColumns()
		m.refreshTableRows(true)
		m.updateProgressWidth()
		m.updateTableHeight()
		return m, nil
	case errMsg:
		m.loading = false
		m.lastErr = msg.err
		return m, nil
	case nodeDetailMsg:
		if m.detailNode == msg.node {
			m.detailBusy = false
			m.detailErr = nil
			m.detailBody = msg.detail
		}
		m.updateTableHeight()
		return m, nil
	case nodeDetailErrMsg:
		if m.detailNode == msg.node {
			m.detailBusy = false
			m.detailErr = msg.err
			m.detailBody = ""
		}
		m.updateTableHeight()
		return m, nil
	case userCancelResultMsg:
		m.userActionBusy = false
		if msg.err != nil {
			m.userActionCanCancel = false
			m.userActionMsg = "scancel failed: " + msg.err.Error()
			return m, nil
		}
		m.userActionCanCancel = false
		m.userActionMsg = "scancel success."
		m.loading = true
		return m, fetchDataCmd(m.ds)
	}

	if m.isUserTab() {
		m.userList, cmd = m.userList.Update(msg)
		return m, cmd
	}
	m.table, cmd = m.table.Update(msg)
	m.refreshTableRows(false)
	return m, cmd
}

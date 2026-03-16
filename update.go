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
		if m.modalOpen {
			if key.Matches(msg, m.keys.Quit) {
				return m, tea.Quit
			}
			return m, m.updateModal(msg)
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
		case msg.String() == "enter":
			if m.isUserTab() {
				return m, nil
			}
			return m, m.openNodeSrunDialog()
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
			m.activeTab = 0
		} else if prevPartitionName != "" {
			found := -1
			for i := range m.partitions {
				if m.partitions[i].Name == prevPartitionName {
					found = i
					break
				}
			}
			if found >= 0 {
				m.activeTab = found + 1 // tab 0 is User
			} else if len(m.partitions) == 0 {
				m.activeTab = 0
			} else if m.activeTab > len(m.partitions) {
				m.activeTab = len(m.partitions)
			}
		} else if len(m.partitions) == 0 {
			m.activeTab = 0
		} else if m.activeTab > len(m.partitions) {
			m.activeTab = len(m.partitions)
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
			if m.modalOpen && m.modalKind == modalNodeDetail {
				m.modalTitle = nodeDetailTitle(msg.node, msg.detail)
				m.setModalBody(formatNodeDetailDisplay(msg.node, msg.detail))
			}
		}
		m.updateTableHeight()
		return m, nil
	case nodeDetailErrMsg:
		if m.detailNode == msg.node {
			m.detailBusy = false
			m.detailErr = msg.err
			m.detailBody = ""
			if m.modalOpen && m.modalKind == modalNodeDetail {
				m.modalTitle = firstNonEmpty(m.detailNode, "Node Detail")
				m.setModalBody("Error: " + msg.err.Error())
			}
		}
		m.updateTableHeight()
		return m, nil
	case userCancelResultMsg:
		m.modalBusy = false
		if msg.err != nil {
			if m.modalOpen && m.modalKind == modalUserAction {
				m.setModalBody("scancel failed: " + msg.err.Error())
				m.modalButtons = []modalButton{{Label: "Close", Action: modalActionClose}}
				m.modalFocus = 0
			}
			return m, nil
		}
		if m.modalOpen && m.modalKind == modalUserAction {
			m.setModalBody("scancel success.")
			m.modalButtons = []modalButton{{Label: "Close", Action: modalActionClose}}
			m.modalFocus = 0
		}
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

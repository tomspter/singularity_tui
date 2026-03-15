package main

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) updateTableColumns() {
	w := m.width
	if w <= 0 {
		w = 120
	}
	if m.isUserTab() {
		jobIDW := 10
		partW := 10
		nameW := 18
		userW := 10
		stateW := 5
		timeW := 10
		nodesW := 7
		nodeListW := w - (jobIDW + partW + nameW + userW + stateW + timeW + nodesW + 10)
		if nodeListW < 18 {
			nodeListW = 18
		}
		m.table.SetColumns([]table.Column{
			{Title: "JOBID", Width: jobIDW},
			{Title: "PARTITION", Width: partW},
			{Title: "NAME", Width: nameW},
			{Title: "USER", Width: userW},
			{Title: "ST", Width: stateW},
			{Title: "TIME", Width: timeW},
			{Title: "NODES", Width: nodesW},
			{Title: "NODELIST(REASON)", Width: nodeListW},
		})
		tw := w - 2
		if tw < 40 {
			tw = 40
		}
		m.table.SetWidth(tw)
		return
	}
	p := m.selectedPartition()
	nodeNeed := len("Node") + 2
	stateNeed := len("State") + 2
	cpuNeed := len("CPU(A/T)") + 2
	memNeed := len("MEM(A/T)") + 2
	gpuNeed := len("GPU(A/T)") + 2

	if p != nil {
		for _, n := range p.Nodes {
			nodeNeed = maxInt(nodeNeed, len(n.Name)+2)
			stateNeed = maxInt(stateNeed, lipgloss.Width(renderStateCell(n.State))+2)
			cpuNeed = maxInt(cpuNeed, lipgloss.Width(m.formatCPUCell(n.CPUAlloc, n.CPUTotal))+2)
			memNeed = maxInt(memNeed, lipgloss.Width(m.formatMEMCell(n.MemAllocMB, n.MemTotalMB))+2)
			gpuNeed = maxInt(gpuNeed, lipgloss.Width(formatGPUCell(n.GPUAlloc, n.GPUTotal))+2)
		}
	}

	nodeMin, stateMin, cpuMin, memMin, gpuMin := 14, 18, 10, 14, 10
	nodeMax, stateMax, cpuMax, memMax, gpuMax := 26, 22, 14, 18, 18

	nodeW := clampInt(nodeNeed, nodeMin, nodeMax)
	stateW := clampInt(stateNeed, stateMin, stateMax)
	cpuW := clampInt(cpuNeed, cpuMin, cpuMax)
	memW := clampInt(memNeed, memMin, memMax)
	gpuW := clampInt(gpuNeed, gpuMin, gpuMax)

	total := nodeW + stateW + cpuW + memW + gpuW
	available := w - 2
	if available < nodeMin+stateMin+cpuMin+memMin+gpuMin {
		available = nodeMin + stateMin + cpuMin + memMin + gpuMin
	}

	if total > available {
		over := total - available
		// Shrink node first to avoid overly wide node column.
		over = shrinkWidth(&nodeW, nodeMin, over)
		over = shrinkWidth(&memW, memMin, over)
		over = shrinkWidth(&stateW, stateMin, over)
		over = shrinkWidth(&cpuW, cpuMin, over)
		_ = shrinkWidth(&gpuW, gpuMin, over)
	} else if total < available {
		extra := available - total
		// Grow metrics columns first; keep node moderate.
		extra = growWidth(&memW, memMax, extra)
		extra = growWidth(&cpuW, cpuMax, extra)
		extra = growWidth(&stateW, stateMax, extra)
		extra = growWidth(&gpuW, gpuMax, extra)
		_ = growWidth(&nodeW, nodeMax, extra)
	}

	m.nodeColW = nodeW
	m.stateColW = stateW
	m.cpuColW = cpuW
	m.memColW = memW
	m.gpuColW = gpuW
	m.table.SetColumns([]table.Column{
		{Title: "Node", Width: nodeW},
		{Title: "State", Width: stateW},
		{Title: "CPU(A/T)", Width: cpuW},
		{Title: "MEM(A/T)", Width: memW},
		{Title: "GPU(A/T)", Width: gpuW},
	})
	tw := m.width - 2
	if tw < 40 {
		tw = 40
	}
	m.table.SetWidth(tw)
}

func (m *model) updateProgressWidth() {
	w := progressWidth
	if m.width > 0 {
		w = m.width / 4
	}
	if w < 16 {
		w = 16
	}
	if w > 44 {
		w = 44
	}
	m.cpuBar.SetWidth(w)
	m.memBar.SetWidth(w)
	m.gpuBar.SetWidth(w)
}

func (m *model) updateTableHeight() {
	if m.height <= 0 {
		return
	}
	nonTableHeight := 12
	h := m.height - nonTableHeight
	if h < 3 {
		h = 3
	}
	m.tableHeight = h
	m.table.SetHeight(h)
}

func (m *model) selectedPartition() *partitionSummary {
	if m.isUserTab() {
		return nil
	}
	if len(m.partitions) == 0 || m.activeTab < 0 || m.activeTab >= len(m.partitions) {
		return nil
	}
	return &m.partitions[m.activeTab]
}

func (m *model) isUserTab() bool {
	return m.activeTab == len(m.partitions)
}

func (m *model) tabCount() int {
	return len(m.partitions) + 1 // + User
}

func (m *model) selectedNodeName() string {
	if len(m.visibleNodes) == 0 {
		return ""
	}
	cursor := m.table.Cursor()
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(m.visibleNodes) {
		cursor = len(m.visibleNodes) - 1
	}
	return strings.TrimSpace(m.visibleNodes[cursor].Name)
}

func (m *model) refreshTableRows(resetCursor bool) {
	if m.isUserTab() {
		m.visibleNodes = nil
		m.table.SetRows([]table.Row{})
		m.table.SetCursor(0)
		return
	}

	p := m.selectedPartition()
	if p == nil {
		m.table.SetRows([]table.Row{})
		m.visibleNodes = nil
		return
	}

	cursor := m.table.Cursor()
	if resetCursor {
		cursor = 0
	}

	nodes := append([]nodeInfo(nil), p.Nodes...)
	sortNodes(nodes, m.sortMode)
	m.visibleNodes = nodes

	rows := make([]table.Row, 0, len(nodes))
	for i, n := range nodes {
		nodeCell := n.Name
		if i == cursor {
			nodeCell = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorFgPrimary)).
				Background(lipgloss.Color(colorSelectionBg)).
				Bold(true).
				Render(n.Name)
		}
		rows = append(rows, table.Row{
			nodeCell,
			renderStateCell(n.State),
			m.formatCPUCell(n.CPUAlloc, n.CPUTotal),
			m.formatMEMCell(n.MemAllocMB, n.MemTotalMB),
			formatGPUCell(n.GPUAlloc, n.GPUTotal),
		})
	}
	m.table.SetRows(m.normalizeRowsForTable(rows))
	if len(rows) == 0 {
		m.table.SetCursor(0)
		return
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(rows) {
		cursor = len(rows) - 1
	}
	m.table.SetCursor(cursor)
}

func (m *model) switchTab(delta int) {
	if m.tabCount() == 0 {
		return
	}
	m.activeTab = (m.activeTab + delta + m.tabCount()) % m.tabCount()
	m.detailOpen = false
	m.detailBusy = false
	m.detailErr = nil
	m.detailBody = ""
	m.detailNode = ""
	m.table.SetCursor(0)
	m.updateTableColumns()
	m.refreshTableRows(true)
	m.updateTableHeight()
}

func (m *model) toggleNodeDetail() tea.Cmd {
	if m.isUserTab() {
		return nil
	}
	node := m.selectedNodeName()
	if node == "" {
		m.detailOpen = true
		m.detailBusy = false
		m.detailErr = errors.New("no node selected")
		m.detailBody = ""
		return nil
	}
	if m.detailOpen && m.detailNode == node {
		m.detailOpen = false
		m.detailBusy = false
		m.detailErr = nil
		m.updateTableHeight()
		return nil
	}
	m.detailOpen = true
	m.detailNode = node
	m.detailBody = ""
	m.detailErr = nil
	m.detailBusy = true
	m.updateTableHeight()
	return fetchNodeDetailCmd(node)
}

func (m *model) buildUserRows() []table.Row {
	return nil
}

func defaultUserLabel(user string) string {
	if strings.TrimSpace(user) == "" {
		return "current"
	}
	return user
}

func (m *model) normalizeRowsForTable(rows []table.Row) []table.Row {
	cols := m.table.Columns()
	colN := len(cols)
	if colN == 0 {
		return rows
	}
	out := make([]table.Row, 0, len(rows))
	for _, r := range rows {
		row := make(table.Row, colN)
		copyN := len(r)
		if copyN > colN {
			copyN = colN
		}
		copy(row[:copyN], r[:copyN])
		for i := copyN; i < colN; i++ {
			row[i] = ""
		}
		out = append(out, row)
	}
	return out
}

func (m *model) formatCPUCell(alloc, total int) string {
	return fmt.Sprintf("%d/%d", alloc, total)
}

func (m *model) formatMEMCell(allocMB, totalMB int) string {
	return fmt.Sprintf("%s/%s", formatMemMB(allocMB), formatMemMB(totalMB))
}

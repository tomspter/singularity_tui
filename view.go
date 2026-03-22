package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) separator() string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	return m.separatorStyle.Render(strings.Repeat("-", w))
}

func (m model) renderPopupBox() string {
	w := m.width
	h := m.height
	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 30
	}

	popupW := int(float64(w) * 0.78)
	if popupW < 64 {
		popupW = 64
	}
	if popupW > 130 {
		popupW = 130
	}
	maxPopupH := h - 4
	if maxPopupH < 10 {
		maxPopupH = 10
	}

	bodyW := popupW - 6
	if bodyW < 20 {
		bodyW = 20
	}

	var lines []string
	title := m.detailStyle.Render(fmt.Sprintf("Node Detail: %s", m.detailNode))
	if m.detailBusy {
		lines = []string{m.mutedStyle.Render("loading ...")}
	} else if m.detailErr != nil {
		lines = []string{m.errorStyle.Render(m.detailErr.Error())}
	} else if strings.TrimSpace(m.detailBody) == "" {
		lines = []string{m.mutedStyle.Render("no detail")}
	} else {
		lines = strings.Split(strings.TrimSpace(m.detailBody), "\n")
	}

	maxBodyLines := maxPopupH - 6
	if maxBodyLines < 3 {
		maxBodyLines = 3
	}
	if len(lines) > maxBodyLines {
		lines = lines[:maxBodyLines]
		lines[len(lines)-1] = lines[len(lines)-1] + " ..."
	}

	body := lipgloss.NewStyle().Width(bodyW).Render(strings.Join(lines, "\n"))
	hint := m.mutedStyle.Render("t/esc: close")
	popupSep := m.separatorStyle.Render(strings.Repeat("-", bodyW))
	content := strings.Join([]string{title, popupSep, body, hint}, "\n")
	return m.popupStyle.Width(popupW).Render(content)
}

func (m model) overlayPopup(base, popup string) string {
	if popup == "" {
		return base
	}
	w := m.width
	h := m.height
	if w <= 0 || h <= 0 {
		return base + "\n" + popup
	}

	// Base frame.
	frame := lipgloss.Place(w, h, lipgloss.Left, lipgloss.Top, base)
	// Lipgloss mask: dim and mute everything behind the popup.
	masked := m.ui.OverlayMask.Render(frame)

	pw := lipgloss.Width(popup)
	ph := lipgloss.Height(popup)
	x := (w - pw) / 2
	y := (h - ph) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	bg := lipgloss.NewLayer(masked).X(0).Y(0).Z(0)
	fg := lipgloss.NewLayer(popup).X(x).Y(y).Z(1)
	return lipgloss.NewCompositor(bg, fg).Render()
}

func (m model) View() tea.View {
	content := m.renderMainView()
	if m.modalOpen {
		content = m.overlayPopup(content, m.renderModal())
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) renderMainView() string {
	status := "loading..."
	if m.lastErr != nil {
		status = "error"
	} else if !m.loading {
		status = "ready"
	}
	lastSync := "-"
	if !m.lastUpdated.IsZero() {
		lastSync = m.lastUpdated.Format("2006-01-02 15:04:05")
	}
	sections := []string{m.renderTabs(), ""}

	p := m.selectedPartition()
	if m.lastErr != nil {
		sections = append(sections, m.errorStyle.Render("Error: "+m.lastErr.Error()))
	} else if m.isUserTab() {
		sections = append(sections,
			m.renderUserOverview(),
			m.separator(),
			m.renderUserTables(),
			m.renderStatusBar(status, lastSync),
		)
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	} else if p == nil {
		sections = append(sections, m.mutedStyle.Render("No partition data from sinfo."))
	} else {
		sections = append(sections,
			m.renderUsageLine("CPU", p.CPUAlloc, p.CPUTotal, false),
			"",
			m.renderUsageLine("MEM", p.MemAllocMB, p.MemTotalMB, true),
			"",
			m.renderUsageLine("GPU", p.GPUAlloc, p.GPUTotal, false),
			"",
			"States: "+renderStateSummary(p.StateCount),
		)
	}
	sections = append(sections, m.separator(), m.table.View(), m.renderStatusBar(status, lastSync))
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) renderStatusBar(status, lastSync string) string {
	w := m.width
	if w <= 0 {
		w = 100
	}
	leftText := strings.ToUpper(status)
	rightText := lastSync
	centerText := m.renderStatusHelp()

	availableText := w - 4 // left/right each have horizontal padding(0,1)
	if availableText < 3 {
		availableText = 3
	}
	leftW := lipgloss.Width(leftText)
	rightW := lipgloss.Width(rightText)
	centerW := lipgloss.Width(centerText)
	minLeft := 4
	minRight := 8

	total := leftW + rightW + centerW
	if total > availableText {
		overflow := total - availableText
		dec := minInt(centerW, overflow)
		centerW -= dec
		overflow -= dec

		if overflow > 0 {
			can := maxInt(0, rightW-minRight)
			dec = minInt(can, overflow)
			rightW -= dec
			overflow -= dec
		}
		if overflow > 0 {
			can := maxInt(0, leftW-minLeft)
			dec = minInt(can, overflow)
			leftW -= dec
			overflow -= dec
		}
		if overflow > 0 {
			dec = minInt(maxInt(1, rightW)-1, overflow)
			rightW -= dec
			overflow -= dec
		}
		if overflow > 0 {
			dec = minInt(maxInt(1, leftW)-1, overflow)
			leftW -= dec
			overflow -= dec
		}
	}

	leftText = truncateToWidth(leftText, leftW)
	rightText = truncateToWidth(rightText, rightW)
	centerText = truncateToWidth(centerText, centerW)

	left := m.ui.statusLeft(status).Render(leftText)
	right := m.ui.StatusRight.Render(rightText)

	fillWidth := w - lipgloss.Width(left) - lipgloss.Width(right)
	if fillWidth < 0 {
		fillWidth = 0
	}
	center := m.ui.StatusCenter.Width(fillWidth).Render(centerText)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, center, right)
}

func (m model) renderStatusHelp() string {
	if m.modalOpen {
		if m.modalKind == modalNodeDetail {
			return "r raw on/off  ↑/↓ or j/k scroll raw  ←/→ button  Enter select  esc close"
		}
		if m.modalKind == modalNodeSrun && m.srunForm != nil {
			return "↑/↓ next field  ←/→ change option  Enter next/run  esc close  q quit"
		}
		return "←/→ focus button  Enter/y select  n/esc close  q quit"
	}
	if m.isUserTab() {
		return "↑/↓ select  pgup/pgdn page  k cancel job  ←/→ tab  r refresh  q quit"
	}
	return "↑/↓ select node  Enter srun form  t detail  s state  c cpu  m mem  g gpu  ←/→ tab  r refresh  q quit"
}

func (m model) renderUserOverview() string {
	u := m.userSummary
	var b strings.Builder
	if u.QuotaErr != "" {
		b.WriteString(m.errorStyle.Render("squota: " + u.QuotaErr))
		return b.String()
	}
	if len(u.QuotaEntries) == 0 {
		b.WriteString(m.mutedStyle.Render("squota: no data"))
		return b.String()
	}
	maxRows := 4
	if len(u.QuotaEntries) < maxRows {
		maxRows = len(u.QuotaEntries)
	}

	labelW := 0
	capTextW := 0
	fileTextW := 0
	for i := 0; i < maxRows; i++ {
		q := u.QuotaEntries[i]
		label := fmt.Sprintf("%s %s", q.Account, q.Filesystem)
		if w := lipgloss.Width(label); w > labelW {
			labelW = w
		}
		capRatio := ratioFromInts(q.UsedBytes, q.HardBytes)
		filesRatio := ratioFromInts(q.FilesUsed, q.FilesHard)
		capText := fmt.Sprintf("%s/%s (%d%%)", formatMemMB(q.UsedBytes), formatMemMB(q.HardBytes), int(capRatio*100))
		fileText := fmt.Sprintf("%d/%d (%d%%)", q.FilesUsed, q.FilesHard, int(filesRatio*100))
		if w := lipgloss.Width(capText); w > capTextW {
			capTextW = w
		}
		if w := lipgloss.Width(fileText); w > fileTextW {
			fileTextW = w
		}
	}
	if labelW < 18 {
		labelW = 18
	}

	barW := 26
	if m.width > 0 {
		fixed := labelW + 2 + 3 + 1 + capTextW + 2 + 4 + 1 + fileTextW
		spaceForBars := m.width - fixed - 6
		if spaceForBars > 16 {
			barW = spaceForBars / 2
		}
	}
	if barW < 12 {
		barW = 12
	}
	if barW > 36 {
		barW = 36
	}

	labelStyle := m.ui.QuotaLabel.Width(labelW)
	tagStyle := m.ui.QuotaTag
	capTextStyle := m.ui.QuotaValue.Width(capTextW)
	fileTextStyle := m.ui.QuotaValue.Width(fileTextW)

	for i := 0; i < maxRows; i++ {
		q := u.QuotaEntries[i]
		label := labelStyle.Render(fmt.Sprintf("%s %s", q.Account, q.Filesystem))
		capRatio := ratioFromInts(q.UsedBytes, q.HardBytes)
		filesRatio := ratioFromInts(q.FilesUsed, q.FilesHard)
		capBarModel := m.memBar
		capBarModel.SetWidth(barW)
		filesBarModel := m.cpuBar
		filesBarModel.SetWidth(barW)
		capBar := capBarModel.ViewAs(capRatio)
		filesBar := filesBarModel.ViewAs(filesRatio)
		capPct := int(capRatio * 100)
		filesPct := int(filesRatio * 100)
		capText := fmt.Sprintf("%s/%s (%d%%)", formatMemMB(q.UsedBytes), formatMemMB(q.HardBytes), capPct)
		filesText := fmt.Sprintf("%d/%d (%d%%)", q.FilesUsed, q.FilesHard, filesPct)
		row := strings.Join([]string{
			label,
			tagStyle.Render("CAP"),
			capBar,
			capTextStyle.Render(capText),
			tagStyle.Render("FILE"),
			filesBar,
			fileTextStyle.Render(filesText),
		}, " ")
		b.WriteString(row)
		if i < maxRows-1 {
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

func (m model) renderUserTables() string {
	return m.userList.View()
}

func ratioFromInts(used, total int) float64 {
	if total <= 0 {
		return 0
	}
	r := float64(used) / float64(total)
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}
	runes := []rune(s)
	out := make([]rune, 0, len(runes))
	curr := 0
	limit := width - 3
	for _, r := range runes {
		rw := lipgloss.Width(string(r))
		if curr+rw > limit {
			break
		}
		out = append(out, r)
		curr += rw
	}
	return string(out) + "..."
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m model) renderTabs() string {
	parts := make([]string, 0, len(m.partitions)+1)
	userLabel := "User"
	if m.activeTab == 0 {
		parts = append(parts, m.activeTabStyle.Render(userLabel))
	} else {
		parts = append(parts, m.tabStyle.Render(userLabel))
	}
	if len(m.partitions) > 0 {
		parts = append(parts, "  ")
	}
	for i, p := range m.partitions {
		label := fmt.Sprintf("%s (%d)", p.Name, len(p.Nodes))
		tabIdx := i + 1 // tab 0 is User
		if tabIdx == m.activeTab {
			parts = append(parts, m.activeTabStyle.Render(label))
		} else {
			parts = append(parts, m.tabStyle.Render(label))
		}
	}
	if len(parts) == 0 {
		return m.mutedStyle.Render("Tabs: -")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m model) renderUsageLine(label string, alloc, total int, memory bool) string {
	if total <= 0 {
		return fmt.Sprintf("%-3s n/a", label)
	}

	ratio := float64(alloc) / float64(total)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	bar := m.cpuBar.ViewAs(ratio)
	switch label {
	case "MEM":
		bar = m.memBar.ViewAs(ratio)
	case "GPU":
		bar = m.gpuBar.ViewAs(ratio)
	}
	pct := int(ratio * 100)

	var allocLabel, totalLabel string
	if memory {
		allocLabel = formatMemMB(alloc)
		totalLabel = formatMemMB(total)
	} else {
		allocLabel = fmt.Sprintf("%d", alloc)
		totalLabel = fmt.Sprintf("%d", total)
	}

	labelColor := lipgloss.Color(colorProgressCPU)
	switch label {
	case "MEM":
		labelColor = lipgloss.Color(colorProgressMEM)
	case "GPU":
		labelColor = lipgloss.Color(colorProgressGPU)
	}
	labelText := lipgloss.NewStyle().Bold(true).Foreground(labelColor).Width(5).Render(label)
	metrics := lipgloss.NewStyle().Foreground(lipgloss.Color(colorFgPrimary)).Render(fmt.Sprintf("%s/%s", allocLabel, totalLabel))
	percent := lipgloss.NewStyle().Bold(true).Foreground(labelColor).Width(4).Align(lipgloss.Right).Render(fmt.Sprintf("%d%%", pct))
	return fmt.Sprintf("%s %s  %s  %s", labelText, bar, metrics, percent)
}

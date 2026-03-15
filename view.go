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
	masked := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMaskFg)).
		Faint(true).
		Render(frame)

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
	if m.detailOpen {
		content = m.overlayPopup(content, m.renderPopupBox())
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) renderMainView() string {
	var b strings.Builder

	//b.WriteString(m.titleStyle.Render("Slurm Visual TUI"))
	//b.WriteString("\n")

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
	//b.WriteString(m.separator())
	//b.WriteString("\n")

	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	p := m.selectedPartition()
	b.WriteString("\n")
	if m.lastErr != nil {
		b.WriteString(m.errorStyle.Render("Error: " + m.lastErr.Error()))
		b.WriteString("\n")
	} else if p == nil {
		b.WriteString(m.mutedStyle.Render("No partition data from sinfo."))
		b.WriteString("\n")
	} else {
		b.WriteString(m.renderUsageLine("CPU", p.CPUAlloc, p.CPUTotal, false))
		b.WriteString("\n\n")
		b.WriteString(m.renderUsageLine("MEM", p.MemAllocMB, p.MemTotalMB, true))
		b.WriteString("\n\n")
		b.WriteString(m.renderUsageLine("GPU", p.GPUAlloc, p.GPUTotal, false))
		b.WriteString("\n\n")
		b.WriteString("States: " + renderStateSummary(p.StateCount))
		b.WriteString("\n")
	}
	b.WriteString(m.separator())
	b.WriteString("\n")

	b.WriteString(m.table.View())
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar(status, lastSync))

	return b.String()
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

	leftStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Background(lipgloss.Color(colorTNBrightBlack)).
		Foreground(lipgloss.Color(colorFgPrimary))
	switch strings.ToLower(status) {
	case "ready":
		leftStyle = leftStyle.Foreground(lipgloss.Color(colorStateIdle))
	case "loading...":
		leftStyle = leftStyle.Foreground(lipgloss.Color(colorStateMixed))
	case "error":
		leftStyle = leftStyle.Foreground(lipgloss.Color(colorStateDrain))
	}
	left := leftStyle.Render(leftText)

	right := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colorFgSecondary)).
		Background(lipgloss.Color(colorTNBrightBlack)).
		Padding(0, 1).
		Render(rightText)

	fillWidth := w - lipgloss.Width(left) - lipgloss.Width(right)
	if fillWidth < 0 {
		fillWidth = 0
	}
	center := lipgloss.NewStyle().
		Background(lipgloss.Color(colorPanel)).
		Foreground(lipgloss.Color(colorFgMuted)).
		Width(fillWidth).
		Align(lipgloss.Center).
		Render(centerText)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, center, right)
}

func (m model) renderStatusHelp() string {
	return "←/→ partition  s state  c cpu  m mem  g gpu  t detail  r refresh  q quit"
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
	if len(m.partitions) == 0 {
		return m.mutedStyle.Render("Partitions: -")
	}
	parts := make([]string, 0, len(m.partitions))
	for i, p := range m.partitions {
		label := fmt.Sprintf("%s (%d)", p.Name, len(p.Nodes))
		if i == m.activeTab {
			parts = append(parts, m.activeTabStyle.Render(label))
		} else {
			parts = append(parts, m.tabStyle.Render(label))
		}
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

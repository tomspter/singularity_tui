package main

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
)

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func renderStateSummary(count map[string]int) string {
	if len(count) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(count))
	for k := range count {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		oi := stateOrder(keys[i])
		oj := stateOrder(keys[j])
		if oi == oj {
			return keys[i] < keys[j]
		}
		return oi < oj
	})
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", renderStateLabel(k), count[k]))
	}
	return strings.Join(parts, " | ")
}

func renderStateLabel(state string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(stateColor(state))).Render(state)
}

func renderStateCell(state string) string {
	bullet := lipgloss.NewStyle().Foreground(lipgloss.Color(stateColor(state))).Render("●")
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(colorFgPrimary)).Render(" " + state)
	return bullet + text
}

func shortBlockBar(alloc, total, width int) string {
	if width <= 0 {
		width = 6
	}
	if total <= 0 {
		return strings.Repeat("░", width)
	}
	ratio := float64(alloc) / float64(total)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("▮", filled) + strings.Repeat("▯", width-filled)
}

func formatGPUCell(alloc, total int) string {
	if total <= 0 {
		return fmt.Sprintf("%d/%d", alloc, total)
	}
	blocks := 8
	if total <= 4 {
		blocks = 4
	}
	return fmt.Sprintf("%d/%d %s", alloc, total, shortBlockBar(alloc, total, blocks))
}

func stateColor(state string) string {
	s := strings.ToUpper(strings.TrimSpace(state))
	switch {
	case strings.Contains(s, "IDLE"):
		return colorStateIdle
	case strings.Contains(s, "MIX"):
		return colorStateMixed
	case strings.Contains(s, "ALLOC"), strings.Contains(s, "COMP"), strings.Contains(s, "RUN"):
		return colorStateAlloc
	case strings.Contains(s, "DRAIN"), strings.Contains(s, "DOWN"), strings.Contains(s, "FAIL"):
		return colorStateDrain
	default:
		return colorStateDefault
	}
}

func sortNodes(nodes []nodeInfo, mode sortMode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		ni := nodes[i]
		nj := nodes[j]
		switch mode {
		case sortByCPU:
			ri := usageForAscSort(ni.CPUAlloc, ni.CPUTotal)
			rj := usageForAscSort(nj.CPUAlloc, nj.CPUTotal)
			if ri != rj {
				return ri < rj
			}
		case sortByMEM:
			ri := usageForAscSort(ni.MemAllocMB, ni.MemTotalMB)
			rj := usageForAscSort(nj.MemAllocMB, nj.MemTotalMB)
			if ri != rj {
				return ri < rj
			}
		case sortByGPU:
			ri := usageForAscSort(ni.GPUAlloc, ni.GPUTotal)
			rj := usageForAscSort(nj.GPUAlloc, nj.GPUTotal)
			if ri != rj {
				return ri < rj
			}
		default:
			oi := stateOrder(ni.State)
			oj := stateOrder(nj.State)
			if oi != oj {
				return oi < oj
			}

			// If GPU exists, sort by remaining GPU (descending).
			gi := remainingResource(ni.GPUAlloc, ni.GPUTotal)
			gj := remainingResource(nj.GPUAlloc, nj.GPUTotal)
			if gi != gj {
				return gi > gj
			}

			// Then remaining CPU (descending).
			ciRem := remainingResource(ni.CPUAlloc, ni.CPUTotal)
			cjRem := remainingResource(nj.CPUAlloc, nj.CPUTotal)
			if ciRem != cjRem {
				return ciRem > cjRem
			}

			// Then remaining MEM (descending).
			miRem := remainingResource(ni.MemAllocMB, ni.MemTotalMB)
			mjRem := remainingResource(nj.MemAllocMB, nj.MemTotalMB)
			if miRem != mjRem {
				return miRem > mjRem
			}
		}

		ci := usageRatio(ni.CPUAlloc, ni.CPUTotal)
		cj := usageRatio(nj.CPUAlloc, nj.CPUTotal)
		if ci != cj {
			return ci > cj
		}
		return ni.Name < nj.Name
	})
}

func remainingResource(alloc, total int) int {
	if total <= 0 {
		return -1
	}
	return total - alloc
}

func usageRatio(alloc, total int) float64 {
	if total <= 0 {
		return -1
	}
	return float64(alloc) / float64(total)
}

func usageForAscSort(alloc, total int) float64 {
	if total <= 0 {
		return 2
	}
	return float64(alloc) / float64(total)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func shrinkWidth(target *int, min, need int) int {
	if need <= 0 {
		return 0
	}
	can := *target - min
	if can <= 0 {
		return need
	}
	use := can
	if use > need {
		use = need
	}
	*target -= use
	return need - use
}

func growWidth(target *int, max, extra int) int {
	if extra <= 0 {
		return 0
	}
	can := max - *target
	if can <= 0 {
		return extra
	}
	use := can
	if use > extra {
		use = extra
	}
	*target += use
	return extra - use
}

func stateOrder(s string) int {
	s = strings.ToUpper(s)
	switch {
	case strings.Contains(s, "IDLE"):
		return 1
	case strings.Contains(s, "MIX"):
		return 2
	case strings.Contains(s, "ALLOC"):
		return 3
	case strings.Contains(s, "DRAIN"), strings.Contains(s, "DOWN"), strings.Contains(s, "FAIL"):
		return 4
	default:
		return 100
	}
}

func formatMemMB(mb int) string {
	if mb <= 0 {
		return "0M"
	}
	if mb >= 1024*1024 {
		return fmt.Sprintf("%.1fT", float64(mb)/1024.0/1024.0)
	}
	if mb >= 1024 {
		return fmt.Sprintf("%.1fG", float64(mb)/1024.0)
	}
	return fmt.Sprintf("%dM", mb)
}

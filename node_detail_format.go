package main

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

func formatNodeDetailDisplay(node, raw string) string {
	fields := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		for k, v := range parseKeyValueLine(line) {
			fields[k] = v
		}
	}
	if len(fields) == 0 {
		return strings.TrimSpace(raw)
	}

	part := firstNonEmpty(fields["Partitions"], fields["Partition"], "-")
	arch := firstNonEmpty(fields["Arch"], "-")

	cfgCPU, cfgMem, cfgGPU := parseTRES(fields["CfgTRES"])
	if cfgCPU <= 0 {
		cfgCPU = parseInt(fields["CPUTot"])
	}
	if cfgMem <= 0 {
		cfgMem = parseInt(fields["RealMemory"])
	}
	if cfgGPU <= 0 {
		cfgGPU = parseGPUFromGRES(firstNonEmpty(fields["Gres"], fields["CfgGRES"]))
	}

	head := []string{
		fmt.Sprintf("%s partition · %s · %s CPU · %s MEM · %s",
			part,
			arch,
			hl(fmt.Sprintf("%d", cfgCPU)),
			hl(formatMemMB(cfgMem)),
			gpuSummary(cfgGPU),
		),
		"",
	}

	cpuAlloc := parseInt(fields["CPUAlloc"])
	cpuTot := parseInt(fields["CPUTot"])
	if cpuTot <= 0 {
		cpuTot = cfgCPU
	}
	memAlloc := parseInt(fields["AllocMem"])
	memTot := parseInt(fields["RealMemory"])
	if memTot <= 0 {
		memTot = cfgMem
	}
	allocCPU, allocMem, allocGPU := parseTRES(fields["AllocTRES"])
	if allocCPU > 0 {
		cpuAlloc = allocCPU
	}
	if allocMem > 0 {
		memAlloc = allocMem
	}
	gpuAlloc := allocGPU
	if gpuAlloc <= 0 {
		gpuAlloc = parseGPUFromGRES(fields["GresUsed"])
	}
	gpuTot := cfgGPU
	if gpuTot <= 0 {
		gpuTot = parseGPUFromGRES(fields["Gres"])
	}

	left := sectionLines("Resources", [][2]string{
		{"CPU", fmt.Sprintf("%s / %s", hl(fmt.Sprintf("%d", cpuAlloc)), hl(fmt.Sprintf("%d", maxInt(cpuTot, 0))))},
		{"Memory", fmt.Sprintf("%s / %s", hl(formatMemMB(memAlloc)), hl(formatMemMB(maxInt(memTot, 0))))},
		{"GPU", gpuPair(gpuAlloc, gpuTot)},
		{"Load", firstNonEmpty(fields["CPULoad"], "-")},
	})

	right := sectionLines("Scheduling", [][2]string{
		{"State", normalizeState(firstNonEmpty(fields["State"], "-"))},
		{"Partition", part},
		{"Weight", firstNonEmpty(fields["Weight"], "-")},
		{"Owner", firstNonEmpty(fields["Owner"], "N/A")},
	})

	left2 := sectionLines("Topology", [][2]string{
		{"Sockets", hl(firstNonEmpty(fields["Sockets"], "-"))},
		{"Cores/Sock", hl(firstNonEmpty(fields["CoresPerSocket"], "-"))},
		{"Threads/Core", hl(firstNonEmpty(fields["ThreadsPerCore"], "-"))},
		{"Board", hl(firstNonEmpty(fields["Boards"], "-"))},
	})

	right2 := sectionLines("Runtime", [][2]string{
		{"BootTime", shortTime(fields["BootTime"])},
		{"SlurmdStart", shortTime(firstNonEmpty(fields["SlurmdStartTime"], fields["SlurmdStart"]))},
		{"LastBusy", shortTime(firstNonEmpty(fields["LastBusyTime"], fields["LastBusy"]))},
		{"ResumeAfter", firstNonEmpty(fields["ResumeAfterTime"], fields["ResumeAfter"], "None")},
	})

	body := []string{}
	body = append(body, joinTwoColumns(left, right, 38)...)
	body = append(body, "")
	body = append(body, joinTwoColumns(left2, right2, 38)...)
	body = append(body, "")
	body = append(body, sectionLines("TRES", [][2]string{
		{"CfgTRES", firstNonEmpty(fields["CfgTRES"], "-")},
		{"AllocTRES", firstNonEmpty(fields["AllocTRES"], "-")},
	})...)

	return strings.Join(append(head, body...), "\n")
}

func nodeDetailTitle(node, raw string) string {
	fields := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		for k, v := range parseKeyValueLine(line) {
			fields[k] = v
		}
	}
	name := firstNonEmpty(fields["NodeName"], node, "-")
	state := normalizeState(firstNonEmpty(fields["State"], "UNKNOWN"))
	return fmt.Sprintf("%s [%s]", name, state)
}

func sectionLines(title string, kv [][2]string) []string {
	lines := []string{title}
	for _, p := range kv {
		lines = append(lines, fmt.Sprintf("%-13s %s", p[0], p[1]))
	}
	return lines
}

func joinTwoColumns(left, right []string, leftWidth int) []string {
	n := len(left)
	if len(right) > n {
		n = len(right)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		l := ""
		r := ""
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		out = append(out, padRight(l, leftWidth)+"  "+r)
	}
	return out
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func hl(v string) string {
	return v
}

func gpuSummary(total int) string {
	if total <= 0 {
		return "no GPU"
	}
	return fmt.Sprintf("%s GPU", hl(fmt.Sprintf("%d", total)))
}

func gpuPair(alloc, total int) string {
	if total <= 0 {
		return "none"
	}
	return fmt.Sprintf("%s / %s", hl(fmt.Sprintf("%d", alloc)), hl(fmt.Sprintf("%d", total)))
}

func shortTime(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "N/A" || v == "(null)" {
		return "-"
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02-15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}
	if len(v) >= 19 {
		return strings.ReplaceAll(v[:19], "T", " ")
	}
	return v
}

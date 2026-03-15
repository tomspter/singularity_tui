package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type slurmDataSource struct{}

func (d slurmDataSource) Fetch(ctx context.Context) ([]partitionSummary, error) {
	nodes, err := d.fetchSinfoNodes(ctx)
	if err != nil {
		return nil, err
	}

	allocMap, _ := d.fetchScontrolNodeAlloc(ctx)
	for i := range nodes {
		if alloc, ok := allocMap[nodes[i].Name]; ok {
			if alloc.CPUAlloc > 0 {
				nodes[i].CPUAlloc = alloc.CPUAlloc
			}
			if alloc.CPUTotal > 0 {
				nodes[i].CPUTotal = alloc.CPUTotal
			}
			if alloc.MemAllocMB > 0 {
				nodes[i].MemAllocMB = alloc.MemAllocMB
			}
			if alloc.MemTotalMB > 0 {
				nodes[i].MemTotalMB = alloc.MemTotalMB
			}
			if alloc.GPUAlloc >= 0 {
				nodes[i].GPUAlloc = alloc.GPUAlloc
			}
			if alloc.GPUTotal > 0 {
				nodes[i].GPUTotal = alloc.GPUTotal
			}
		}
	}

	return summarizeByPartition(nodes), nil
}

func (d slurmDataSource) fetchSinfoNodes(ctx context.Context) ([]nodeInfo, error) {
	out, err := runCommand(ctx, "sinfo", "-Nh", "-o", "%P|%N|%T|%c|%C|%m|%G")
	if err != nil {
		return nil, fmt.Errorf("无法读取 sinfo: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	nodes := make([]nodeInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}

		partition := strings.TrimSuffix(strings.TrimSpace(fields[0]), "*")
		nodeName := strings.TrimSpace(fields[1])
		state := normalizeState(fields[2])
		cpuTotal := parseInt(fields[3])
		cpuAlloc, cpuTotalFromState := parseCPUState(fields[4])
		if cpuTotalFromState > 0 {
			cpuTotal = cpuTotalFromState
		}
		memTotal := parseInt(fields[5])
		gpuTotal := parseGPUFromGRES(fields[6])

		if partition == "" || nodeName == "" {
			continue
		}

		parts := strings.Split(partition, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			nodes = append(nodes, nodeInfo{
				Partition:  part,
				Name:       nodeName,
				State:      state,
				CPUAlloc:   cpuAlloc,
				CPUTotal:   cpuTotal,
				MemAllocMB: 0,
				MemTotalMB: memTotal,
				GPUAlloc:   0,
				GPUTotal:   gpuTotal,
			})
		}
	}

	if len(nodes) == 0 {
		return nil, errors.New("sinfo 没有返回节点数据")
	}
	return nodes, nil
}

func (d slurmDataSource) fetchScontrolNodeAlloc(ctx context.Context) (map[string]nodeAllocInfo, error) {
	out, err := runCommand(ctx, "scontrol", "show", "node", "-o")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make(map[string]nodeAllocInfo, len(lines))

	for _, line := range lines {
		fields := parseKeyValueLine(line)
		name := fields["NodeName"]
		if name == "" {
			continue
		}

		info := nodeAllocInfo{
			CPUAlloc:   parseInt(fields["CPUAlloc"]),
			CPUTotal:   parseInt(fields["CPUTot"]),
			MemAllocMB: parseInt(fields["AllocMem"]),
			MemTotalMB: parseInt(fields["RealMemory"]),
			GPUAlloc:   parseGPUFromGRES(fields["GresUsed"]),
			GPUTotal:   parseGPUFromGRES(fields["Gres"]),
		}

		if cfgTRES := fields["CfgTRES"]; cfgTRES != "" {
			cpu, mem, gpu := parseTRES(cfgTRES)
			if cpu > 0 {
				info.CPUTotal = cpu
			}
			if mem > 0 {
				info.MemTotalMB = mem
			}
			if gpu > 0 {
				info.GPUTotal = gpu
			}
		}
		if allocTRES := fields["AllocTRES"]; allocTRES != "" {
			cpu, mem, gpu := parseTRES(allocTRES)
			if cpu > 0 {
				info.CPUAlloc = cpu
			}
			if mem > 0 {
				info.MemAllocMB = mem
			}
			if gpu >= 0 {
				info.GPUAlloc = gpu
			}
		}
		result[name] = info
	}

	return result, nil
}

func parseKeyValueLine(line string) map[string]string {
	out := map[string]string{}
	for _, token := range strings.Fields(line) {
		i := strings.Index(token, "=")
		if i <= 0 {
			continue
		}
		out[token[:i]] = token[i+1:]
	}
	return out
}

func parseCPUState(s string) (alloc, total int) {
	parts := strings.Split(strings.TrimSpace(s), "/")
	if len(parts) != 4 {
		return 0, 0
	}
	return parseInt(parts[0]), parseInt(parts[3])
}

func parseGPUFromGRES(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "(null)" || raw == "N/A" {
		return 0
	}
	total := 0
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" || !strings.Contains(part, "gpu") {
			continue
		}
		base := part
		if idx := strings.Index(base, "("); idx >= 0 {
			base = base[:idx]
		}
		tokens := strings.Split(base, ":")
		found := false
		for i := len(tokens) - 1; i >= 0; i-- {
			v := parseInt(tokens[i])
			if v > 0 {
				total += v
				found = true
				break
			}
		}
		if !found && strings.HasPrefix(base, "gpu") {
			total++
		}
	}
	return total
}

func parseTRES(raw string) (cpu, memMB, gpu int) {
	gpu = 0
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := parts[1]
		switch {
		case key == "cpu":
			cpu = parseInt(val)
		case key == "mem":
			memMB = parseMemToMB(val)
		case strings.HasPrefix(key, "gres/gpu"):
			gpu = parseInt(val)
		}
	}
	return cpu, memMB, gpu
}

func parseMemToMB(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	scale := 1.0
	last := raw[len(raw)-1]
	switch last {
	case 'K', 'k':
		scale = 1.0 / 1024.0
		raw = raw[:len(raw)-1]
	case 'M', 'm':
		scale = 1.0
		raw = raw[:len(raw)-1]
	case 'G', 'g':
		scale = 1024.0
		raw = raw[:len(raw)-1]
	case 'T', 't':
		scale = 1024.0 * 1024.0
		raw = raw[:len(raw)-1]
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return parseInt(raw)
	}
	return int(val * scale)
}

func parseInt(raw string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(raw))
	return n
}

func normalizeState(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "UNKNOWN"
	}
	raw = strings.Split(raw, "+")[0]
	raw = strings.Split(raw, "~")[0]
	return strings.ToUpper(raw)
}

func summarizeByPartition(nodes []nodeInfo) []partitionSummary {
	grouped := map[string]*partitionSummary{}
	for _, n := range nodes {
		p, ok := grouped[n.Partition]
		if !ok {
			p = &partitionSummary{
				Name:       n.Partition,
				StateCount: map[string]int{},
			}
			grouped[n.Partition] = p
		}
		p.Nodes = append(p.Nodes, n)
		p.CPUAlloc += n.CPUAlloc
		p.CPUTotal += n.CPUTotal
		p.MemAllocMB += n.MemAllocMB
		p.MemTotalMB += n.MemTotalMB
		p.GPUAlloc += n.GPUAlloc
		p.GPUTotal += n.GPUTotal
		p.StateCount[n.State]++
	}

	out := make([]partitionSummary, 0, len(grouped))
	for _, summary := range grouped {
		sort.Slice(summary.Nodes, func(i, j int) bool {
			return summary.Nodes[i].Name < summary.Nodes[j].Name
		})
		out = append(out, *summary)
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i].Nodes) == len(out[j].Nodes) {
			return out[i].Name < out[j].Name
		}
		return len(out[i].Nodes) > len(out[j].Nodes)
	})
	return out
}

func runCommand(parent context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %v: %w (%s)", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func fetchUserSummary(ctx context.Context) userSummary {
	user := strings.TrimSpace(os.Getenv("USER"))
	out := userSummary{
		User:       user,
		StateCount: map[string]int{},
		Jobs:       []userJob{},
	}

	// squeue aggregation
	sqOut, sqErr := runCommand(ctx, "squeue", "-h", "-o", "%i|%P|%j|%u|%t|%M|%D|%R")
	if sqErr != nil {
		out.QueueErr = sqErr.Error()
	} else {
		lines := strings.Split(strings.TrimSpace(sqOut), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			f := strings.Split(line, "|")
			if len(f) < 8 {
				continue
			}
			job := userJob{
				JobID:     strings.TrimSpace(f[0]),
				Partition: strings.TrimSpace(f[1]),
				Name:      strings.TrimSpace(f[2]),
				User:      strings.TrimSpace(f[3]),
				State:     strings.TrimSpace(f[4]),
				RunTime:   strings.TrimSpace(f[5]),
				Nodes:     strings.TrimSpace(f[6]),
				NodeList:  strings.TrimSpace(f[7]),
			}
			state := normalizeState(job.State)
			out.TotalJobs++
			out.StateCount[state]++
			out.Jobs = append(out.Jobs, job)
		}
	}

	// squota summary lines
	quotaOut, quotaErr := runCommand(ctx, "squota")
	if quotaErr != nil {
		out.QuotaErr = quotaErr.Error()
	} else {
		for _, line := range strings.Split(strings.TrimSpace(quotaOut), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "-") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 13 {
				continue
			}
			if strings.EqualFold(fields[0], "account") {
				continue
			}
			usedMB, idx, ok := parseQuotaSize(fields, 2)
			if !ok || idx >= len(fields) {
				continue
			}
			softMB, idx, ok := parseQuotaSize(fields, idx)
			if !ok || idx >= len(fields) {
				continue
			}
			hardMB, idx, ok := parseQuotaSize(fields, idx)
			if !ok || idx+4 >= len(fields) {
				continue
			}
			entry := quotaEntry{
				Account:    fields[0],
				Filesystem: fields[1],
				UsedBytes:  usedMB,
				SoftBytes:  softMB,
				HardBytes:  hardMB,
				GraceBytes: fields[idx],
				FilesUsed:  parseInt(fields[idx+1]),
				FilesSoft:  parseInt(fields[idx+2]),
				FilesHard:  parseInt(fields[idx+3]),
				GraceFiles: fields[idx+4],
			}
			out.QuotaEntries = append(out.QuotaEntries, entry)
			if len(out.QuotaEntries) >= 10 {
				break
			}
		}
	}

	return out
}

func parseQuotaSize(fields []string, idx int) (valueMB int, next int, ok bool) {
	if idx >= len(fields) {
		return 0, idx, false
	}
	// Case 1: split tokens, e.g. "8.4" "GB"
	if idx+1 < len(fields) {
		unit := strings.ToUpper(strings.TrimSpace(fields[idx+1]))
		if unit == "KB" || unit == "MB" || unit == "GB" || unit == "TB" {
			return parseMemToMB(fields[idx] + unit[:1]), idx + 2, true
		}
	}

	// Case 2: combined token, e.g. "8.4GB"
	token := strings.TrimSpace(fields[idx])
	if token == "" {
		return 0, idx, false
	}
	if len(token) > 1 {
		suffix := strings.ToUpper(token[len(token)-2:])
		if suffix == "KB" || suffix == "MB" || suffix == "GB" || suffix == "TB" {
			token = token[:len(token)-1] // drop trailing 'B', keep K/M/G/T
		}
	}
	last := token[len(token)-1]
	switch last {
	case 'K', 'k', 'M', 'm', 'G', 'g', 'T', 't':
		return parseMemToMB(token), idx + 1, true
	}
	return 0, idx, false
}

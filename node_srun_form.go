package main

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	huh "charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) openNodeSrunDialog() tea.Cmd {
	if m.isUserTab() {
		return nil
	}
	node := m.selectedNodeName()
	if node == "" {
		m.openModal(modalNodeSrun, "Generate srun", "No node selected.", []modalButton{
			{Label: "Close", Action: modalActionClose},
		})
		return nil
	}

	part := ""
	if p := m.selectedPartition(); p != nil {
		part = strings.TrimSpace(p.Name)
	}
	if part == "" {
		part = "default"
	}

	gpuDefault := "0"
	if n := m.selectedNode(); n != nil && n.GPUTotal > 0 {
		gpuDefault = "1"
	}

	m.srunFormCfg = nodeSrunFormConfig{
		Partition: part,
		Node:      node,
		JobName:   "interactive",
		CPUs:      "4",
		Memory:    "16G",
		GPUs:      gpuDefault,
		TimeLimit: "01:00:00",
		Command:   "bash",
	}
	m.srunCommand = ""
	m.srunForm = newNodeSrunForm(&m.srunFormCfg, m.partitionNames(), m.width)
	m.openModal(modalNodeSrun, fmt.Sprintf("Generate srun · %s", node), "", []modalButton{
		{Label: "Close", Action: modalActionClose},
	})
	if m.srunForm == nil {
		return nil
	}
	return m.srunForm.Init()
}

func (m *model) updateNodeSrunForm(msg tea.Msg) tea.Cmd {
	if m.srunForm == nil {
		return nil
	}
	updated, cmd := m.srunForm.Update(msg)
	if form, ok := updated.(*huh.Form); ok {
		m.srunForm = form
	}
	switch m.srunForm.State {
	case huh.StateCompleted:
		m.srunCommand = buildSrunCommand(m.srunFormCfg)
		m.srunForm = nil
		m.modalTitle = "srun Command"
		m.setModalBody(m.srunCommand)
		m.modalButtons = []modalButton{{Label: "Close", Action: modalActionClose}}
		m.modalFocus = 0
		return nil
	case huh.StateAborted:
		m.closeModal()
		return nil
	default:
		return cmd
	}
}

func newNodeSrunForm(cfg *nodeSrunFormConfig, partitions []string, width int) *huh.Form {
	if cfg == nil {
		return nil
	}
	if strings.TrimSpace(cfg.Command) == "" {
		cfg.Command = "bash"
	}
	options := make([]huh.Option[string], 0, len(partitions))
	for _, p := range partitions {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		options = append(options, huh.NewOption(p, p))
	}
	if len(options) == 0 {
		options = append(options, huh.NewOption("default", "default"))
	}
	if strings.TrimSpace(cfg.Partition) == "" {
		cfg.Partition = options[0].Value
	}

	formW := (width * 2) / 3
	if formW < 52 {
		formW = 52
	}
	if formW > 94 {
		formW = 94
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Partition").
				Description("Select target partition").
				Options(options...).
				Value(&cfg.Partition),
			huh.NewInput().
				Title("Node").
				Description("Target node name").
				Value(&cfg.Node).
				Validate(requiredField("node")),
			huh.NewInput().
				Title("Job Name").
				Description("SLURM job-name label").
				Value(&cfg.JobName).
				Validate(requiredField("job name")),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("CPUs").
				Description("cpus-per-task").
				Value(&cfg.CPUs).
				Validate(minIntField("CPUs", 1)),
			huh.NewInput().
				Title("Memory").
				Description("Examples: 16G, 64000").
				Value(&cfg.Memory).
				Validate(requiredField("memory")),
			huh.NewInput().
				Title("GPUs").
				Description("0 means no GPU request").
				Value(&cfg.GPUs).
				Validate(minIntField("GPUs", 0)),
			huh.NewInput().
				Title("Time Limit").
				Description("Examples: 01:00:00, 2:00:00").
				Value(&cfg.TimeLimit).
				Validate(requiredField("time limit")),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("WorkDir").
				Description("Optional: empty = current directory").
				Value(&cfg.WorkDir),
			huh.NewInput().
				Title("Extra Args").
				Description("Optional: e.g. --pty").
				Value(&cfg.ExtraArgs),
			huh.NewInput().
				Title("Command").
				Description("Command after '--', e.g. bash -lc 'nvidia-smi'").
				Value(&cfg.Command).
				Validate(requiredField("command")),
		),
	).
		WithTheme(newNodeSrunTheme()).
		WithWidth(formW).
		WithShowHelp(false)

	return form
}

func newNodeSrunTheme() huh.ThemeFunc {
	return func(bool) *huh.Styles {
		styles := huh.ThemeCharm(true)
		accent := lipgloss.Color(colorAccent)
		panel := lipgloss.Color(colorPanel)
		muted := lipgloss.Color(colorFgMuted)
		primary := lipgloss.Color(colorFgPrimary)
		secondary := lipgloss.Color(colorFgSecondary)
		separator := lipgloss.Color(colorSeparator)

		styles.Form.Base = styles.Form.Base.
			Background(panel).
			Foreground(primary)
		styles.Group.Base = styles.Group.Base.
			Background(panel).
			Foreground(primary)
		styles.Group.Title = styles.Group.Title.
			Bold(true).
			Foreground(accent)
		styles.Group.Description = styles.Group.Description.
			Foreground(muted)
		styles.FieldSeparator = styles.FieldSeparator.Foreground(separator)
		styles.Focused.Base = styles.Focused.Base.Foreground(primary)
		styles.Focused.Title = styles.Focused.Title.Foreground(accent).Bold(true)
		styles.Focused.Description = styles.Focused.Description.Foreground(secondary)
		styles.Focused.Next = styles.Focused.Next.Foreground(muted)
		styles.Focused.Card = styles.Focused.Card.BorderForeground(accent)
		styles.Focused.NoteTitle = styles.Focused.NoteTitle.Foreground(accent)
		styles.Blurred.Base = styles.Blurred.Base.Foreground(secondary)
		styles.Blurred.Title = styles.Blurred.Title.Foreground(secondary)
		styles.Blurred.Description = styles.Blurred.Description.Foreground(muted)
		styles.Blurred.Next = styles.Blurred.Next.Foreground(muted)
		styles.Blurred.Card = styles.Blurred.Card.BorderForeground(separator)
		styles.Blurred.NoteTitle = styles.Blurred.NoteTitle.Foreground(secondary)
		return styles
	}
}

func (m *model) selectedNode() *nodeInfo {
	if len(m.visibleNodes) == 0 {
		return nil
	}
	cursor := m.table.Cursor()
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(m.visibleNodes) {
		cursor = len(m.visibleNodes) - 1
	}
	return &m.visibleNodes[cursor]
}

func (m *model) partitionNames() []string {
	if len(m.partitions) == 0 {
		return nil
	}
	out := make([]string, 0, len(m.partitions))
	for _, p := range m.partitions {
		name := strings.TrimSpace(p.Name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func requiredField(field string) func(string) error {
	return func(v string) error {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}

func minIntField(field string, min int) func(string) error {
	return func(v string) error {
		raw := strings.TrimSpace(v)
		if raw == "" {
			return fmt.Errorf("%s is required", field)
		}
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s must be an integer", field)
		}
		if n < min {
			return fmt.Errorf("%s must be >= %d", field, min)
		}
		return nil
	}
}

func buildSrunCommand(cfg nodeSrunFormConfig) string {
	args := []string{"srun"}
	if v := strings.TrimSpace(cfg.Partition); v != "" {
		args = append(args, "--partition="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.Node); v != "" {
		args = append(args, "--nodelist="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.JobName); v != "" {
		args = append(args, "--job-name="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.CPUs); v != "" {
		args = append(args, "--cpus-per-task="+v)
	}
	if v := strings.TrimSpace(cfg.Memory); v != "" {
		args = append(args, "--mem="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.GPUs); v != "" && v != "0" {
		args = append(args, "--gres=gpu:"+v)
	}
	if v := strings.TrimSpace(cfg.TimeLimit); v != "" {
		args = append(args, "--time="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.WorkDir); v != "" {
		args = append(args, "--chdir="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.ExtraArgs); v != "" {
		args = append(args, v)
	}
	cmd := strings.TrimSpace(cfg.Command)
	if cmd == "" {
		cmd = "bash"
	}
	args = append(args, "--", cmd)
	return strings.Join(args, " ")
}

func shellQuote(v string) string {
	if v == "" {
		return "''"
	}
	if !strings.ContainsAny(v, " \t\n'\"`$&|;<>(){}[]*?!\\") {
		return v
	}
	return "'" + strings.ReplaceAll(v, "'", `'"'"'`) + "'"
}

package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	huh "charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

const memoryModeCustom = "__custom__"

type conditionalField struct {
	field  huh.Field
	hidden func() bool
}

func wrapConditionalField(field huh.Field, hidden func() bool) huh.Field {
	return &conditionalField{field: field, hidden: hidden}
}

func (c *conditionalField) isHidden() bool {
	return c.hidden != nil && c.hidden()
}

func (c *conditionalField) Init() tea.Cmd {
	if c.isHidden() {
		return nil
	}
	return c.field.Init()
}

func (c *conditionalField) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	if c.isHidden() {
		return c, nil
	}
	updated, cmd := c.field.Update(msg)
	if f, ok := updated.(huh.Field); ok {
		c.field = f
	}
	return c, cmd
}

func (c *conditionalField) View() string {
	if c.isHidden() {
		return ""
	}
	return c.field.View()
}

func (c *conditionalField) Blur() tea.Cmd {
	if c.isHidden() {
		return nil
	}
	return c.field.Blur()
}

func (c *conditionalField) Focus() tea.Cmd {
	if c.isHidden() {
		return nil
	}
	return c.field.Focus()
}

func (c *conditionalField) Error() error {
	if c.isHidden() {
		return nil
	}
	return c.field.Error()
}

func (c *conditionalField) Run() error {
	if c.isHidden() {
		return nil
	}
	return c.field.Run()
}

func (c *conditionalField) RunAccessible(w io.Writer, r io.Reader) error {
	if c.isHidden() {
		return nil
	}
	return c.field.RunAccessible(w, r)
}

func (c *conditionalField) Skip() bool {
	if c.isHidden() {
		return true
	}
	return c.field.Skip()
}

func (c *conditionalField) Zoom() bool {
	if c.isHidden() {
		return false
	}
	return c.field.Zoom()
}

func (c *conditionalField) KeyBinds() []key.Binding {
	if c.isHidden() {
		return nil
	}
	return c.field.KeyBinds()
}

func (c *conditionalField) WithTheme(t huh.Theme) huh.Field {
	c.field = c.field.WithTheme(t)
	return c
}

func (c *conditionalField) WithKeyMap(k *huh.KeyMap) huh.Field {
	c.field = c.field.WithKeyMap(k)
	return c
}

func (c *conditionalField) WithWidth(w int) huh.Field {
	c.field = c.field.WithWidth(w)
	return c
}

func (c *conditionalField) WithHeight(h int) huh.Field {
	c.field = c.field.WithHeight(h)
	return c
}

func (c *conditionalField) WithPosition(p huh.FieldPosition) huh.Field {
	c.field = c.field.WithPosition(p)
	return c
}

func (c *conditionalField) GetKey() string {
	return c.field.GetKey()
}

func (c *conditionalField) GetValue() any {
	if c.isHidden() {
		return nil
	}
	return c.field.GetValue()
}

func (m *model) openNodeSrunDialog() tea.Cmd {
	if m.isUserTab() {
		return nil
	}
	node := m.selectedNodeName()
	if node == "" {
		m.openModal(modalNodeSrun, "Launch srun", "No node selected.", []modalButton{
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

	gpuFree := 0
	cpuFree := 0
	memFreeMB := 0
	if n := m.selectedNode(); n != nil {
		gpuFree = maxInt(0, n.GPUTotal-n.GPUAlloc)
		cpuFree = maxInt(0, n.CPUTotal-n.CPUAlloc)
		memFreeMB = maxInt(0, n.MemTotalMB-n.MemAllocMB)
	}

	gpuDefault := "0"
	if gpuFree > 0 {
		gpuDefault = "1"
	}

	m.srunFormCfg = nodeSrunFormConfig{
		Partition:    part,
		Node:         node,
		Placement:    "Selected node only",
		JobName:      "interactive",
		CPUs:         "8",
		Memory:       "16G",
		MemoryMode:   "16G",
		MemoryCustom: "",
		GPUs:         gpuDefault,
		TimeLimit:    "02:00:00",
		QOS:          "",
		Constraint:   "",
		ExtraArgs:    "",
		Command:      "bash",
		AdvancedOpen: false,
		Confirmed:    false,
	}

	m.srunCommand = buildSrunCommand(m.srunFormCfg)
	m.srunForm = newNodeSrunForm(&m.srunFormCfg, m.partitionNames(), gpuFree, cpuFree, memFreeMB, m.width)
	m.openModal(modalNodeSrun, "Launch srun", "", []modalButton{
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

	m.srunCommand = buildSrunCommand(m.srunFormCfg)

	switch m.srunForm.State {
	case huh.StateCompleted:
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

func newNodeSrunForm(cfg *nodeSrunFormConfig, partitions []string, gpuFree, cpuFree, memFreeMB, width int) *huh.Form {
	if cfg == nil {
		return nil
	}
	if strings.TrimSpace(cfg.Command) == "" {
		cfg.Command = "bash"
	}
	if strings.TrimSpace(cfg.CPUs) == "" {
		cfg.CPUs = "8"
	}
	if strings.TrimSpace(cfg.MemoryMode) == "" {
		cfg.MemoryMode = "16G"
	}
	if strings.TrimSpace(cfg.TimeLimit) == "" {
		cfg.TimeLimit = "02:00:00"
	}

	partitionOptions := buildPartitionOptions(partitions)
	if strings.TrimSpace(cfg.Partition) == "" {
		cfg.Partition = partitionOptions[0].Value
	}

	gpuOptions := buildGPUOptions(gpuFree)
	if !optionValuesContain(gpuOptions, strings.TrimSpace(cfg.GPUs)) {
		cfg.GPUs = gpuOptions[0].Value
	}

	memoryOptions := []huh.Option[string]{
		huh.NewOption("8G", "8G"),
		huh.NewOption("16G", "16G"),
		huh.NewOption("32G", "32G"),
		huh.NewOption("64G", "64G"),
		huh.NewOption("128G", "128G"),
		huh.NewOption("Custom", memoryModeCustom),
	}
	if !optionValuesContain(memoryOptions, strings.TrimSpace(cfg.MemoryMode)) {
		cfg.MemoryMode = "16G"
	}

	timeOptions := []huh.Option[string]{
		huh.NewOption("00:30:00", "00:30:00"),
		huh.NewOption("01:00:00", "01:00:00"),
		huh.NewOption("02:00:00", "02:00:00"),
		huh.NewOption("04:00:00", "04:00:00"),
		huh.NewOption("08:00:00", "08:00:00"),
		huh.NewOption("24:00:00", "24:00:00"),
		huh.NewOption("48:00:00", "48:00:00"),
	}
	if !optionValuesContain(timeOptions, strings.TrimSpace(cfg.TimeLimit)) {
		cfg.TimeLimit = "02:00:00"
	}

	formW := (width * 2) / 3
	if formW < 70 {
		formW = 70
	}
	if formW > 110 {
		formW = 110
	}

	nodeSummary := fmt.Sprintf("Node: %s\nFree: %d GPU · %d CPU · %s MEM",
		strings.TrimSpace(cfg.Node), gpuFree, cpuFree, formatMemMB(memFreeMB))

	fields := []huh.Field{
		huh.NewNote().
			Title("Launch srun").
			Description(nodeSummary),

		huh.NewSelect[string]().
			Title("Placement").
			Options(
				huh.NewOption("Selected node only", "Selected node only"),
				huh.NewOption("Partition only", "Partition only"),
			).
			Value(&cfg.Placement).
			Inline(true),

		huh.NewSelect[string]().
			Title("Partition").
			Options(partitionOptions...).
			Value(&cfg.Partition).
			Inline(true),

		huh.NewInput().
			Title("Run on").
			Value(&cfg.Node).
			Inline(true).
			Validate(func(v string) error {
				if strings.TrimSpace(cfg.Placement) != "Selected node only" {
					return nil
				}
				if strings.TrimSpace(v) == "" {
					return errors.New("run on node is required")
				}
				return nil
			}),

		huh.NewSelect[string]().
			Title("GPU").
			Options(gpuOptions...).
			Value(&cfg.GPUs).
			Inline(true),

		huh.NewInput().
			Title("CPU").
			Value(&cfg.CPUs).
			Inline(true).
			Validate(minIntField("CPU count", 1)),

		huh.NewSelect[string]().
			Title("Memory").
			Options(memoryOptions...).
			Value(&cfg.MemoryMode).
			Inline(true),

		wrapConditionalField(
			huh.NewInput().
				Title("Memory Custom").
				Value(&cfg.MemoryCustom).
				Inline(true).
				Validate(requiredField("custom memory")),
			func() bool {
				return strings.TrimSpace(cfg.MemoryMode) != memoryModeCustom
			},
		),

		huh.NewSelect[string]().
			Title("Time").
			Options(timeOptions...).
			Value(&cfg.TimeLimit).
			Inline(true),

		huh.NewInput().
			Title("Command").
			Value(&cfg.Command).
			Inline(true).
			Validate(requiredField("command")),

		huh.NewConfirm().
			TitleFunc(func() string {
				if cfg.AdvancedOpen {
					return "Advanced ▾"
				}
				return "Advanced ▸"
			}, cfg).
			Affirmative("Expanded").
			Negative("Collapsed").
			Value(&cfg.AdvancedOpen).
			Inline(true),

		wrapConditionalField(
			huh.NewInput().
				Title("Job name").
				Value(&cfg.JobName).
				Inline(true),
			func() bool { return !cfg.AdvancedOpen },
		),
		wrapConditionalField(
			huh.NewInput().
				Title("QOS").
				Value(&cfg.QOS).
				Inline(true),
			func() bool { return !cfg.AdvancedOpen },
		),
		wrapConditionalField(
			huh.NewInput().
				Title("Constraint").
				Value(&cfg.Constraint).
				Inline(true),
			func() bool { return !cfg.AdvancedOpen },
		),
		wrapConditionalField(
			huh.NewInput().
				Title("Extra args").
				Value(&cfg.ExtraArgs).
				Inline(true),
			func() bool { return !cfg.AdvancedOpen },
		),

		huh.NewNote().
			Title("Preview").
			DescriptionFunc(func() string {
				return buildSrunCommand(*cfg)
			}, cfg),

		huh.NewConfirm().
			Title("Run").
			Description("Press Enter to confirm and generate").
			Affirmative("Run").
			Negative("Cancel").
			Value(&cfg.Confirmed).
			Inline(true).
			Validate(func(v bool) error {
				if !v {
					return errors.New("select Run to continue")
				}
				return nil
			}),
	}

	form := huh.NewForm(
		huh.NewGroup(fields...),
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

func buildPartitionOptions(partitions []string) []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(partitions))
	for _, p := range partitions {
		p = strings.TrimSpace(p)
		if p != "" {
			options = append(options, huh.NewOption(p, p))
		}
	}
	if len(options) == 0 {
		options = append(options, huh.NewOption("default", "default"))
	}
	return options
}

func buildGPUOptions(gpuFree int) []huh.Option[string] {
	if gpuFree < 0 {
		gpuFree = 0
	}
	options := make([]huh.Option[string], 0, gpuFree+1)
	for i := 0; i <= gpuFree; i++ {
		v := strconv.Itoa(i)
		options = append(options, huh.NewOption(v, v))
	}
	if len(options) == 0 {
		options = append(options, huh.NewOption("0", "0"))
	}
	return options
}

func optionValuesContain(options []huh.Option[string], value string) bool {
	for _, op := range options {
		if op.Value == value {
			return true
		}
	}
	return false
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

func resolvedMemory(cfg nodeSrunFormConfig) string {
	if strings.TrimSpace(cfg.MemoryMode) == memoryModeCustom {
		return strings.TrimSpace(cfg.MemoryCustom)
	}
	if v := strings.TrimSpace(cfg.MemoryMode); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Memory)
}

func buildSrunCommand(cfg nodeSrunFormConfig) string {
	args := []string{"srun"}
	if v := strings.TrimSpace(cfg.Partition); v != "" {
		args = append(args, "--partition="+shellQuote(v))
	}
	if strings.TrimSpace(cfg.Placement) == "Selected node only" {
		if v := strings.TrimSpace(cfg.Node); v != "" {
			args = append(args, "--nodelist="+shellQuote(v))
		}
	}
	if v := strings.TrimSpace(cfg.JobName); v != "" {
		args = append(args, "--job-name="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.QOS); v != "" {
		args = append(args, "--qos="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.Constraint); v != "" {
		args = append(args, "--constraint="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.CPUs); v != "" {
		args = append(args, "--cpus-per-task="+v)
	}
	if v := resolvedMemory(cfg); v != "" {
		args = append(args, "--mem="+shellQuote(v))
	}
	if v := strings.TrimSpace(cfg.GPUs); v != "" && v != "0" {
		args = append(args, "--gres=gpu:"+v)
	}
	if v := strings.TrimSpace(cfg.TimeLimit); v != "" {
		args = append(args, "--time="+shellQuote(v))
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

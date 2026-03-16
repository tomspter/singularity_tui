package main

import (
	"context"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/table"
	huh "charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

const (
	refreshInterval = 30 * time.Second
	commandTimeout  = 8 * time.Second
	progressWidth   = 28
)

type nodeInfo struct {
	Partition  string
	Name       string
	State      string
	CPUAlloc   int
	CPUTotal   int
	MemAllocMB int
	MemTotalMB int
	GPUAlloc   int
	GPUTotal   int
}

type partitionSummary struct {
	Name       string
	Nodes      []nodeInfo
	CPUAlloc   int
	CPUTotal   int
	MemAllocMB int
	MemTotalMB int
	GPUAlloc   int
	GPUTotal   int
	StateCount map[string]int
}

type nodeAllocInfo struct {
	CPUAlloc   int
	CPUTotal   int
	MemAllocMB int
	MemTotalMB int
	GPUAlloc   int
	GPUTotal   int
}

type dataSource interface {
	Fetch(context.Context) ([]partitionSummary, error)
}

type dataMsg struct {
	partitions []partitionSummary
	user       userSummary
	loadedAt   time.Time
}

type errMsg struct {
	err error
}

type refreshMsg struct{}

type nodeDetailMsg struct {
	node   string
	detail string
}

type nodeDetailErrMsg struct {
	node string
	err  error
}

type userCancelResultMsg struct {
	jobID string
	err   error
}

type userSummary struct {
	User         string
	TotalJobs    int
	StateCount   map[string]int
	CPUReq       int
	MemReqMB     int
	GPUReq       int
	Jobs         []userJob
	QuotaEntries []quotaEntry
	QueueErr     string
	QuotaErr     string
}

type userJob struct {
	JobID     string
	Partition string
	Name      string
	User      string
	State     string
	RunTime   string
	Nodes     string
	NodeList  string
}

type quotaEntry struct {
	Account    string
	Filesystem string
	UsedBytes  int
	SoftBytes  int
	HardBytes  int
	GraceBytes string
	FilesUsed  int
	FilesSoft  int
	FilesHard  int
	GraceFiles string
}

type nodeSrunFormConfig struct {
	Partition string
	Node      string
	JobName   string
	CPUs      string
	Memory    string
	GPUs      string
	TimeLimit string
	WorkDir   string
	ExtraArgs string
	Command   string
}

type keyMap struct {
	PrevPartition key.Binding
	NextPartition key.Binding
	SortState     key.Binding
	SortCPU       key.Binding
	SortMem       key.Binding
	SortGPU       key.Binding
	UserCancel    key.Binding
	NodeDetail    key.Binding
	Refresh       key.Binding
	ToggleHelp    key.Binding
	Quit          key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PrevPartition, k.NextPartition, k.SortState, k.SortCPU, k.UserCancel, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PrevPartition, k.NextPartition, k.SortState, k.SortCPU},
		{k.SortMem, k.SortGPU, k.UserCancel, k.Refresh},
		{k.ToggleHelp, k.Quit},
	}
}

type sortMode int

const (
	sortByState sortMode = iota
	sortByCPU
	sortByMEM
	sortByGPU
)

type modalKind int

const (
	modalNone modalKind = iota
	modalNodeDetail
	modalUserAction
	modalNodeSrun
)

type modalAction int

const (
	modalActionClose modalAction = iota
	modalActionCancelJob
	modalActionToggleRaw
)

type modalButton struct {
	Label  string
	Action modalAction
}

type model struct {
	ds            dataSource
	table         table.Model
	userList      list.Model
	help          help.Model
	keys          keyMap
	width         int
	height        int
	tableHeight   int
	loading       bool
	lastErr       error
	lastUpdated   time.Time
	partitions    []partitionSummary
	userSummary   userSummary
	visibleNodes  []nodeInfo
	activeTab     int
	sortMode      sortMode
	nodeColW      int
	stateColW     int
	cpuColW       int
	memColW       int
	gpuColW       int
	cpuBar        progress.Model
	memBar        progress.Model
	gpuBar        progress.Model
	detailNode    string
	detailBody    string
	detailErr     error
	detailBusy    bool
	userActionJob userJob
	modalOpen     bool
	modalKind     modalKind
	modalTitle    string
	modalBody     string
	modalButtons  []modalButton
	modalFocus    int
	modalBusy     bool
	modalRawOpen  bool
	modalRawOff   int
	srunForm      *huh.Form
	srunFormCfg   nodeSrunFormConfig
	srunCommand   string
	ui            uiStyleSet

	titleStyle     lipgloss.Style
	subtitleStyle  lipgloss.Style
	errorStyle     lipgloss.Style
	mutedStyle     lipgloss.Style
	separatorStyle lipgloss.Style
	detailStyle    lipgloss.Style
	popupStyle     lipgloss.Style
	activeTabStyle lipgloss.Style
	tabStyle       lipgloss.Style
}

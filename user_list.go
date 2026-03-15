package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type userJobListItem struct {
	job userJob
}

func (i userJobListItem) Title() string {
	return fmt.Sprintf("%s  %s  %s  %s", i.job.JobID, i.job.Partition, i.job.State, i.job.Name)
}

func (i userJobListItem) Description() string {
	return fmt.Sprintf("user=%s  time=%s  nodes=%s  nodelist=%s", i.job.User, i.job.RunTime, i.job.Nodes, i.job.NodeList)
}

func (i userJobListItem) FilterValue() string {
	return strings.Join([]string{i.job.JobID, i.job.Partition, i.job.Name, i.job.User, i.job.State, i.job.NodeList}, " ")
}

type userJobDelegate struct {
	idStyle       lipgloss.Style
	nameStyle     lipgloss.Style
	subtleStyle   lipgloss.Style
	timeStyle     lipgloss.Style
	nodeCount     lipgloss.Style
	nodeListStyle lipgloss.Style
	selectedBg    lipgloss.Style
	indicatorOn   string
	indicatorOff  string
	stateBadge    lipgloss.Style
	partBadge     lipgloss.Style
}

func newUserJobDelegate() userJobDelegate {
	return userJobDelegate{
		idStyle:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorTNOrange)),
		nameStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorFgPrimary)),
		subtleStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color(colorFgSecondary)),
		timeStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorTNOrange)),
		nodeCount:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorTNMagenta)),
		nodeListStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorTNCyan)),
		selectedBg:    lipgloss.NewStyle().Background(lipgloss.Color(colorTNBrightBlack)),
		indicatorOn:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorAccent)).Render(">"),
		indicatorOff:  " ",
		stateBadge: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgSecondary)),
		partBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgSecondary)).
			Faint(true),
	}
}

func (d userJobDelegate) Height() int  { return 2 }
func (d userJobDelegate) Spacing() int { return 0 }
func (d userJobDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

func (d userJobDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	v, ok := item.(userJobListItem)
	if !ok || m.Width() <= 0 {
		return
	}
	job := v.job
	selected := index == m.Index() && m.FilterState() != list.Filtering

	indicator := d.indicatorOff
	if selected {
		indicator = d.indicatorOn
	}

	totalW := m.Width()
	contentW := totalW - 2 // indicator + gap
	if contentW < 20 {
		contentW = 20
	}

	state := strings.TrimSpace(job.State)
	if state == "" {
		state = "-"
	}
	partition := strings.TrimSpace(job.Partition)
	if partition == "" {
		partition = "-"
	}
	idW := 10
	if contentW < 28 {
		idW = 8
	}
	if idW > contentW-4 {
		idW = contentW - 4
	}
	if idW < 4 {
		idW = 4
	}

	stateRaw := "[" + state + "]"
	partRaw := "[" + partition + "]"
	rightRaw := stateRaw + " " + partRaw
	rightW := lipgloss.Width(rightRaw)

	nameRaw := strings.TrimSpace(job.Name)
	if nameRaw == "" {
		nameRaw = "-"
	}
	nameW := contentW - idW - rightW - 4
	if nameW < 4 {
		rightRaw = stateRaw
		rightW = lipgloss.Width(rightRaw)
		nameW = contentW - idW - rightW - 4
	}
	if nameW < 4 {
		rightRaw = ""
		rightW = 0
		nameW = contentW - idW - 2
	}
	if nameW < 1 {
		nameW = 1
	}

	idVal := d.idStyle.Width(idW).Render(shortenField(job.JobID, idW))
	nameTrimmed := truncateToWidth(nameRaw, nameW)
	nameVal := d.nameStyle.Render(nameTrimmed)
	line1Content := idVal + "  " + nameVal
	if rightRaw != "" {
		rightVal := d.stateBadge.Render(stateRaw)
		if strings.Contains(rightRaw, partRaw) {
			rightVal += " " + d.partBadge.Render(partRaw)
		}
		line1Content += "  " + rightVal
	}
	line1Content = padToDisplayWidth(line1Content, contentW)
	firstLine := indicator + " " + line1Content
	if selected {
		plain := shortenField(job.JobID, idW) + "  " + nameTrimmed
		if rightRaw != "" {
			plain += "  " + rightRaw
		}
		plain = truncateToWidth(plain, contentW)
		plain = padToDisplayWidth(plain, contentW)
		firstLine = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgPrimary)).
			Background(lipgloss.Color(colorTNBrightBlack)).
			Width(totalW).
			Render("> " + plain)
	}

	nodesRaw := strings.TrimSpace(job.Nodes)
	if nodesRaw == "" {
		nodesRaw = "0"
	}
	nodeLabel := "node"
	if nodesRaw != "1" {
		nodeLabel = "nodes"
	}
	userRaw := strings.TrimSpace(job.User)
	if userRaw == "" {
		userRaw = "-"
	}
	timeRaw := strings.TrimSpace(job.RunTime)
	if timeRaw == "" {
		timeRaw = "-"
	}
	nodesRaw = nodesRaw + " " + nodeLabel

	leftMetaRaw := fmt.Sprintf("%s  %s  %s", userRaw, timeRaw, nodesRaw)
	nodelistRaw := strings.TrimSpace(job.NodeList)
	if nodelistRaw == "" {
		nodelistRaw = "-"
	}

	left2Max := (contentW * 2) / 3
	if left2Max < 16 {
		left2Max = 16
	}
	if left2Max > contentW-8 {
		left2Max = contentW - 8
	}
	left2 := d.subtleStyle.Render(shortenField(leftMetaRaw, left2Max))
	left2W := lipgloss.Width(left2)
	nodeAvail := contentW - left2W - 2

	if nodeAvail < 8 {
		line2Raw := shortenField(leftMetaRaw+"  "+nodelistRaw, contentW)
		line2Content := d.subtleStyle.Render(line2Raw)
		_, _ = fmt.Fprintf(w, "%s\n%s   %s", firstLine, d.indicatorOff, line2Content)
		return
	}

	right2 := d.nodeListStyle.Render(shortenField(nodelistRaw, nodeAvail))
	line2Content := padToDisplayWidth(left2+"  "+right2, contentW)

	_, _ = fmt.Fprintf(w, "%s\n%s   %s", firstLine, d.indicatorOff, line2Content)
}

func (m *model) refreshUserList(reset bool) {
	jobs := append([]userJob(nil), m.userSummary.Jobs...)
	me := strings.TrimSpace(m.userSummary.User)
	sort.SliceStable(jobs, func(i, j int) bool {
		mi := me != "" && strings.TrimSpace(jobs[i].User) == me
		mj := me != "" && strings.TrimSpace(jobs[j].User) == me
		if mi != mj {
			return mi
		}
		return jobs[i].JobID > jobs[j].JobID
	})

	items := make([]list.Item, 0, len(jobs))
	for _, j := range jobs {
		items = append(items, userJobListItem{job: j})
	}
	m.userList.Title = fmt.Sprintf("SQUEUE · %d jobs", len(jobs))
	if len(items) == 0 {
		if m.userSummary.QueueErr != "" {
			items = append(items, userJobListItem{job: userJob{JobID: "ERR", Partition: "-", Name: m.userSummary.QueueErr, User: "-", State: "ERR", RunTime: "-", Nodes: "-", NodeList: "-"}})
		} else {
			items = append(items, userJobListItem{job: userJob{JobID: "-", Partition: "-", Name: "no jobs", User: defaultUserLabel(m.userSummary.User), State: "-", RunTime: "-", Nodes: "-", NodeList: "-"}})
		}
	}
	m.userList.SetItems(items)
	if reset {
		m.userList.ResetSelected()
	}
}

func shortenField(s string, width int) string {
	s = strings.TrimSpace(s)
	if width <= 0 || len(s) <= width {
		return s
	}
	if width <= 3 {
		return s[:width]
	}
	return s[:width-3] + "..."
}

func padToDisplayWidth(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func (m *model) selectedUserJob() (userJob, bool) {
	item := m.userList.SelectedItem()
	v, ok := item.(userJobListItem)
	if !ok {
		return userJob{}, false
	}
	return v.job, true
}

func (m *model) openUserActionDialog() {
	job, ok := m.selectedUserJob()
	if !ok {
		m.openModal(modalUserAction, "Cancel Job", "No job selected.", []modalButton{
			{Label: "OK", Action: modalActionClose},
		})
		return
	}
	m.userActionJob = job
	me := strings.TrimSpace(m.userSummary.User)
	if me != "" && strings.TrimSpace(job.User) == me && job.JobID != "-" && job.JobID != "ERR" {
		m.openModal(modalUserAction, "Cancel Job",
			fmt.Sprintf("Cancel job %s (%s)?", job.JobID, job.Name),
			[]modalButton{
				{Label: "Confirm", Action: modalActionCancelJob},
				{Label: "Cancel", Action: modalActionClose},
			},
		)
		return
	}
	m.openModal(modalUserAction, "Cancel Job", "No permission to scancel this job.", []modalButton{
		{Label: "OK", Action: modalActionClose},
	})
}

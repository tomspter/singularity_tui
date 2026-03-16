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
	subtleSel     lipgloss.Style
	timeStyle     lipgloss.Style
	nodeCount     lipgloss.Style
	nodeListStyle lipgloss.Style
	selectedLine  lipgloss.Style
	indicatorOn   lipgloss.Style
	indicatorOff  lipgloss.Style
	stateBadge    lipgloss.Style
	partBadge     lipgloss.Style
}

func newUserJobDelegate(ui uiStyleSet) userJobDelegate {
	return userJobDelegate{
		idStyle:       ui.UserJobID,
		nameStyle:     ui.UserJobName,
		subtleStyle:   ui.UserMeta,
		subtleSel:     ui.UserMetaSelected,
		timeStyle:     ui.UserMetaTime,
		nodeCount:     ui.UserMetaNodes,
		nodeListStyle: ui.UserNodeList,
		selectedLine:  ui.UserSelectedLine,
		indicatorOn:   ui.UserIndicatorOn,
		indicatorOff:  ui.UserIndicatorOff,
		stateBadge:    ui.UserStateBadge,
		partBadge:     ui.UserPartBadge,
	}
}

func (d userJobDelegate) Height() int  { return 2 }
func (d userJobDelegate) Spacing() int { return 1 }
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
	totalW := m.Width()
	indicatorW := 2
	contentW := totalW - indicatorW
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
	stateW := 6

	nameRaw := strings.TrimSpace(job.Name)
	if nameRaw == "" {
		nameRaw = "-"
	}
	nameW := contentW - stateW - idW - 4
	if nameW < 4 {
		nameW = contentW - idW - 2
	}
	if nameW < 1 {
		nameW = 1
	}

	stateVal := d.stateBadge.Foreground(lipgloss.Color(jobStateColor(state))).Width(stateW).MaxWidth(stateW).Align(lipgloss.Left).Render(truncateToWidth(stateRaw, stateW))
	idVal := d.idStyle.Width(idW).MaxWidth(idW).Align(lipgloss.Left).Render(shortenField(job.JobID, idW))
	nameTrimmed := truncateToWidth(nameRaw, nameW)
	nameVal := d.nameStyle.Render(nameTrimmed)
	line1Content := lipgloss.JoinHorizontal(lipgloss.Top, stateVal, "  ", idVal, "  ", nameVal)
	line1Content = padToDisplayWidth(line1Content, contentW)
	firstLine := line1Content
	indicator := d.indicatorOff.Render(" ")
	if selected {
		idSel := d.idStyle.Width(idW).MaxWidth(idW).Align(lipgloss.Left).Render(shortenField(job.JobID, idW))
		nameSel := d.nameStyle.Foreground(lipgloss.Color(colorTNBrightWhite)).Render(nameTrimmed)
		stateSel := d.stateBadge.Foreground(lipgloss.Color(jobStateColor(state))).Width(stateW).MaxWidth(stateW).Align(lipgloss.Left).Render(truncateToWidth(stateRaw, stateW))
		line1Sel := lipgloss.JoinHorizontal(lipgloss.Top, stateSel, "  ", idSel, "  ", nameSel)
		firstLine = d.selectedLine.Width(contentW).Render(padToDisplayWidth(line1Sel, contentW))
		indicator = d.indicatorOn.Render("│")
	}
	userRaw := strings.TrimSpace(job.User)
	if userRaw == "" {
		userRaw = "-"
	}
	timeRaw := strings.TrimSpace(job.RunTime)
	if timeRaw == "" {
		timeRaw = "-"
	}

	nodelistRaw := strings.TrimSpace(job.NodeList)
	if nodelistRaw == "" {
		nodelistRaw = "-"
	}

	userW := 12
	timeW := 10
	partW := 8
	minNodeListW := 8
	if contentW < userW+timeW+partW+minNodeListW+6 {
		userW = 10
		timeW = 9
		partW = 7
	}
	if contentW < userW+timeW+partW+minNodeListW+6 {
		userW = 8
		timeW = 8
		partW = 6
	}
	nodelistW := contentW - userW - timeW - partW - 6
	if nodelistW < minNodeListW {
		nodelistW = minNodeListW
	}
	metaStyle := d.partBadge
	if selected {
		metaStyle = d.partBadge.Foreground(lipgloss.Color(colorTNBrightWhite))
	}
	userCell := metaStyle.Width(userW).MaxWidth(userW).Align(lipgloss.Left).Render(shortenField(userRaw, userW))
	timeCell := metaStyle.Width(timeW).MaxWidth(timeW).Align(lipgloss.Left).Render(shortenField(timeRaw, timeW))
	partCell := metaStyle.Width(partW).MaxWidth(partW).Align(lipgloss.Left).Render(shortenField(partition, partW))
	nodeCell := metaStyle.Width(nodelistW).MaxWidth(nodelistW).Align(lipgloss.Left).Render(shortenField(nodelistRaw, nodelistW))
	line2Content := lipgloss.JoinHorizontal(lipgloss.Top, userCell, "  ", timeCell, "  ", partCell, "  ", nodeCell)
	line2Content = padToDisplayWidth(line2Content, contentW)

	firstLineWithIndicator := lipgloss.JoinHorizontal(lipgloss.Top, indicator, " ", firstLine)
	secondIndicator := d.indicatorOff.Render(" ")
	if selected {
		secondIndicator = d.indicatorOn.Render("│")
	}
	secondLineWithIndicator := lipgloss.JoinHorizontal(lipgloss.Top, secondIndicator, " ", line2Content)
	_, _ = fmt.Fprintf(w, "%s\n%s", firstLineWithIndicator, secondLineWithIndicator)
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

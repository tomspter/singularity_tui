package main

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/list"
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

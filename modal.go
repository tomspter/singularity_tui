package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) openModal(kind modalKind, title, body string, buttons []modalButton) {
	if len(buttons) == 0 {
		buttons = []modalButton{{Label: "OK", Action: modalActionClose}}
	}
	m.modalOpen = true
	m.modalKind = kind
	m.modalTitle = title
	m.modalBody = body
	m.modalButtons = buttons
	m.modalFocus = 0
	m.modalBusy = false
	m.modalRawOpen = false
	m.modalRawOff = 0
}

func (m *model) closeModal() {
	m.modalOpen = false
	m.modalKind = modalNone
	m.modalTitle = ""
	m.modalBody = ""
	m.modalButtons = nil
	m.modalFocus = 0
	m.modalBusy = false
	m.modalRawOpen = false
	m.modalRawOff = 0
	m.userActionJob = userJob{}
	m.srunForm = nil
	m.srunFormCfg = nodeSrunFormConfig{}
	m.srunCommand = ""
}

func (m *model) setModalBody(body string) {
	m.modalBody = body
}

func (m model) renderModal() string {
	const rawWindowHeight = 8

	w := m.width
	if w <= 0 {
		w = 100
	}
	popupW := int(float64(w) * 0.72)
	if popupW < 56 {
		popupW = 56
	}
	if popupW > 116 {
		popupW = 116
	}
	bodyW := popupW - 6
	if bodyW < 20 {
		bodyW = 20
	}

	title := m.ui.ModalTitle.Width(bodyW).Render(m.modalTitle)
	if m.modalKind == modalNodeSrun && m.srunForm != nil {
		formW := bodyW
		if formW < 52 {
			formW = 52
		}
		m.srunForm.WithWidth(formW)
		formView := strings.TrimSpace(m.srunForm.View())
		if formView == "" {
			formView = "-"
		}
		body := m.ui.ModalBodyLeft.Width(bodyW).Render(formView)
		hint := m.ui.ModalRawMeta.Width(bodyW).Render("Enter next/submit  Tab next field  Esc close")
		content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", hint)
		return m.ui.ModalBox.Width(popupW).Render(content)
	}

	bodyStyle := m.ui.ModalBodyLeft.Width(bodyW)
	body := ""
	if m.modalKind == modalNodeDetail {
		body = bodyStyle.Render(strings.TrimSpace(m.modalBody))
	} else {
		body = m.ui.ModalBodyCenter.Width(bodyW).Render(strings.TrimSpace(m.modalBody))
	}

	buttons := m.renderModalButtons(bodyW)
	if m.modalKind == modalNodeDetail && m.modalRawOpen {
		rawHeader := m.ui.ModalRawHeader.Render("Raw Output")
		rawBodyText, rawMeta := m.renderRawOutputWindow(rawWindowHeight)
		rawBody := m.ui.ModalRawBody.Width(bodyW).Render(rawBodyText)
		rawInfo := m.ui.ModalRawMeta.Width(bodyW).Render(rawMeta)
		rawBox := m.ui.ModalRawBox.Width(bodyW).Render(lipgloss.JoinVertical(lipgloss.Left, rawHeader, rawBody, rawInfo))
		content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", rawBox, "", buttons)
		return m.ui.ModalBox.Width(popupW).Render(content)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", buttons)
	return m.ui.ModalBox.Width(popupW).Render(content)
}

func (m model) renderModalButtons(width int) string {
	if len(m.modalButtons) == 0 {
		return ""
	}
	maxLabelW := 0
	for _, b := range m.modalButtons {
		lbl := strings.ReplaceAll(b.Label, "\n", " ")
		if w := lipgloss.Width(lbl); w > maxLabelW {
			maxLabelW = w
		}
	}
	if maxLabelW < 5 {
		maxLabelW = 5
	}

	parts := make([]string, 0, len(m.modalButtons))
	for i, b := range m.modalButtons {
		label := strings.ReplaceAll(b.Label, "\n", " ")
		if m.modalBusy && i == m.modalFocus {
			label = fmt.Sprintf("%s...", b.Label)
		}
		label = centerText(label, maxLabelW)
		if i == m.modalFocus {
			parts = append(parts, m.ui.ModalBtnActive.Render(label))
		} else {
			parts = append(parts, m.ui.ModalBtnIdle.Render(label))
		}
	}
	row := lipgloss.NewStyle().Padding(0, 1).Render(strings.Join(parts, "   "))
	if width <= 0 {
		width = lipgloss.Width(row)
	}
	return m.ui.ModalBtnWrap.Width(width).Render(row)
}

func centerText(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	left := (width - w) / 2
	right := width - w - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func (m *model) updateModal(msg tea.KeyPressMsg) tea.Cmd {
	if !m.modalOpen {
		return nil
	}
	if m.modalKind == modalNodeSrun && m.srunForm != nil {
		if msg.String() == "esc" {
			m.closeModal()
			return nil
		}
		return m.updateNodeSrunForm(msg)
	}
	if m.modalBusy {
		if msg.String() == "esc" {
			m.closeModal()
		}
		return nil
	}
	switch msg.String() {
	case "r":
		if m.modalKind == modalNodeDetail {
			m.modalRawOpen = !m.modalRawOpen
			if !m.modalRawOpen {
				m.modalRawOff = 0
			}
		}
		return nil
	case "up", "k":
		if m.modalKind == modalNodeDetail && m.modalRawOpen {
			if m.modalRawOff > 0 {
				m.modalRawOff--
			}
			return nil
		}
		if len(m.modalButtons) > 1 {
			m.modalFocus = (m.modalFocus - 1 + len(m.modalButtons)) % len(m.modalButtons)
		}
		return nil
	case "down", "j":
		if m.modalKind == modalNodeDetail && m.modalRawOpen {
			lines := strings.Split(strings.TrimSpace(m.detailBody), "\n")
			maxOff := len(lines) - 8
			if maxOff < 0 {
				maxOff = 0
			}
			if m.modalRawOff < maxOff {
				m.modalRawOff++
			}
			return nil
		}
		if len(m.modalButtons) > 1 {
			m.modalFocus = (m.modalFocus + 1) % len(m.modalButtons)
		}
		return nil
	case "pgup":
		if m.modalKind == modalNodeDetail && m.modalRawOpen {
			m.modalRawOff -= 8
			if m.modalRawOff < 0 {
				m.modalRawOff = 0
			}
		}
		return nil
	case "pgdown":
		if m.modalKind == modalNodeDetail && m.modalRawOpen {
			lines := strings.Split(strings.TrimSpace(m.detailBody), "\n")
			maxOff := len(lines) - 8
			if maxOff < 0 {
				maxOff = 0
			}
			m.modalRawOff += 8
			if m.modalRawOff > maxOff {
				m.modalRawOff = maxOff
			}
		}
		return nil
	case "left", "h":
		if len(m.modalButtons) > 1 {
			m.modalFocus = (m.modalFocus - 1 + len(m.modalButtons)) % len(m.modalButtons)
		}
		return nil
	case "right", "l", "tab":
		if len(m.modalButtons) > 1 {
			m.modalFocus = (m.modalFocus + 1) % len(m.modalButtons)
		}
		return nil
	case "esc", "n":
		m.closeModal()
		return nil
	case "enter", "y", " ":
		if m.modalFocus < 0 || m.modalFocus >= len(m.modalButtons) {
			m.closeModal()
			return nil
		}
		return m.executeModalAction(m.modalButtons[m.modalFocus].Action)
	default:
		return nil
	}
}

func (m *model) executeModalAction(action modalAction) tea.Cmd {
	switch action {
	case modalActionCancelJob:
		if strings.TrimSpace(m.userActionJob.JobID) == "" {
			m.closeModal()
			return nil
		}
		m.modalBusy = true
		m.setModalBody("Running scancel ...")
		return cancelJobCmd(m.userActionJob.JobID)
	case modalActionToggleRaw:
		m.modalRawOpen = !m.modalRawOpen
		return nil
	default:
		m.closeModal()
		return nil
	}
}

func (m model) renderRawOutputWindow(height int) (string, string) {
	lines := strings.Split(strings.TrimSpace(m.detailBody), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		lines = nil
	}
	if height <= 0 {
		height = 8
	}
	total := len(lines)
	if total == 0 {
		return "-", "0/0"
	}
	off := m.modalRawOff
	if off < 0 {
		off = 0
	}
	maxOff := total - height
	if maxOff < 0 {
		maxOff = 0
	}
	if off > maxOff {
		off = maxOff
	}
	end := off + height
	if end > total {
		end = total
	}
	window := strings.Join(lines[off:end], "\n")
	meta := fmt.Sprintf("%d-%d/%d", off+1, end, total)
	return window, meta
}

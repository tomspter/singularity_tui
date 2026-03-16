package main

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type uiStyleSet struct {
	StatusLeftBase    lipgloss.Style
	StatusLeftReady   lipgloss.Style
	StatusLeftLoading lipgloss.Style
	StatusLeftError   lipgloss.Style
	StatusCenter      lipgloss.Style
	StatusRight       lipgloss.Style

	QuotaLabel lipgloss.Style
	QuotaTag   lipgloss.Style
	QuotaValue lipgloss.Style

	TableHeader       lipgloss.Style
	TableCell         lipgloss.Style
	TableSelectedCell lipgloss.Style

	UserJobID        lipgloss.Style
	UserJobName      lipgloss.Style
	UserMeta         lipgloss.Style
	UserMetaSelected lipgloss.Style
	UserMetaTime     lipgloss.Style
	UserMetaNodes    lipgloss.Style
	UserNodeList     lipgloss.Style
	UserSelectedLine lipgloss.Style
	UserIndicatorOn  lipgloss.Style
	UserIndicatorOff lipgloss.Style
	UserStateBadge   lipgloss.Style
	UserPartBadge    lipgloss.Style

	ModalBox        lipgloss.Style
	ModalTitle      lipgloss.Style
	ModalBodyLeft   lipgloss.Style
	ModalBodyCenter lipgloss.Style
	ModalRawHeader  lipgloss.Style
	ModalRawBody    lipgloss.Style
	ModalRawMeta    lipgloss.Style
	ModalRawBox     lipgloss.Style
	ModalBtnActive  lipgloss.Style
	ModalBtnIdle    lipgloss.Style
	ModalBtnWrap    lipgloss.Style

	OverlayMask lipgloss.Style
}

func newUIStyleSet() uiStyleSet {
	return uiStyleSet{
		StatusLeftBase: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color(colorTNBrightBlack)).
			Foreground(lipgloss.Color(colorFgPrimary)),
		StatusLeftReady: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color(colorTNBrightBlack)).
			Foreground(lipgloss.Color(colorStateIdle)),
		StatusLeftLoading: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color(colorTNBrightBlack)).
			Foreground(lipgloss.Color(colorStateMixed)),
		StatusLeftError: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color(colorTNBrightBlack)).
			Foreground(lipgloss.Color(colorStateDrain)),
		StatusCenter: lipgloss.NewStyle().
			Background(lipgloss.Color(colorPanel)).
			Foreground(lipgloss.Color(colorFgMuted)).
			Align(lipgloss.Center),
		StatusRight: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgSecondary)).
			Background(lipgloss.Color(colorTNBrightBlack)).
			Padding(0, 1),

		QuotaLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgSecondary)).
			Align(lipgloss.Left),
		QuotaTag: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgPrimary)),
		QuotaValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgPrimary)).
			Align(lipgloss.Right),

		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgPrimary)),
		TableCell: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgSecondary)),
		TableSelectedCell: lipgloss.NewStyle().
			Background(lipgloss.Color(colorSelectionBg)),

		UserJobID: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorTNBlue)),
		UserJobName: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgPrimary)),
		UserMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgSecondary)),
		UserMetaSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTNBrightWhite)),
		UserMetaTime: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorTNOrange)),
		UserMetaNodes: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTNMagenta)),
		UserNodeList: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorTNCyan)),
		UserSelectedLine: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorTNBrightWhite)),
		UserIndicatorOn: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true),
		UserIndicatorOff: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPanel)),
		UserStateBadge: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgSecondary)),
		UserPartBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgSecondary)).
			Faint(true),

		ModalBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorAccent)).
			Background(lipgloss.Color(colorPanel)).
			Padding(1, 2),
		ModalTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgPrimary)).
			Background(lipgloss.Color(colorPanel)).
			Align(lipgloss.Center),
		ModalBodyLeft: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgPrimary)).
			Background(lipgloss.Color(colorPanel)).
			Align(lipgloss.Left),
		ModalBodyCenter: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgPrimary)).
			Background(lipgloss.Color(colorPanel)).
			Align(lipgloss.Center),
		ModalRawHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgPrimary)).
			Background(lipgloss.Color(colorPanel)),
		ModalRawBody: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgMuted)).
			Background(lipgloss.Color(colorPanel)).
			Align(lipgloss.Left),
		ModalRawMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgMuted)).
			Background(lipgloss.Color(colorPanel)).
			Align(lipgloss.Right),
		ModalRawBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorSeparator)).
			Background(lipgloss.Color(colorPanel)).
			Padding(0, 1),
		ModalBtnActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorFgPrimary)).
			Background(lipgloss.Color(colorAccent)).
			Padding(0, 2).
			Align(lipgloss.Center),
		ModalBtnIdle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFgPrimary)).
			Background(lipgloss.Color(colorPanelAlt)).
			Padding(0, 2).
			Align(lipgloss.Center),
		ModalBtnWrap: lipgloss.NewStyle().
			Background(lipgloss.Color(colorPanel)).
			Align(lipgloss.Center).
			PaddingTop(1),

		OverlayMask: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMaskFg)).
			Faint(true),
	}
}

func (s uiStyleSet) statusLeft(status string) lipgloss.Style {
	switch strings.ToLower(status) {
	case "ready":
		return s.StatusLeftReady
	case "loading...":
		return s.StatusLeftLoading
	case "error":
		return s.StatusLeftError
	default:
		return s.StatusLeftBase
	}
}

func (s uiStyleSet) selectedTableCell(content string, width int) string {
	bg := s.TableSelectedCell
	if width <= 0 {
		return bg.Render(content)
	}
	c := content
	if w := lipgloss.Width(c); w < width {
		c += strings.Repeat(" ", width-w)
	}
	return bg.Render(c)
}

func (s uiStyleSet) selectedTableStateCell(state string, width int) string {
	st := strings.TrimSpace(state)
	if st == "" {
		st = "-"
	}
	fg := lipgloss.Color(stateColor(st))
	bgColor := lipgloss.Color(colorSelectionBg)
	dot := lipgloss.NewStyle().Foreground(fg).Background(bgColor).Render("●")
	txt := lipgloss.NewStyle().Foreground(fg).Background(bgColor).Render(" " + st)
	base := "● " + st
	if width > lipgloss.Width(base) {
		pad := strings.Repeat(" ", width-lipgloss.Width(base))
		return dot + txt + lipgloss.NewStyle().Background(bgColor).Render(pad)
	}
	return dot + txt
}

package main

const (
	// Tokyo Night official palette
	colorTNBg            = "#1a1b26"
	colorTNFg            = "#c0caf5"
	colorTNBlack         = "#15161e"
	colorTNRed           = "#f7768e"
	colorTNGreen         = "#9ece6a"
	colorTNYellow        = "#e0af68"
	colorTNBlue          = "#7aa2f7"
	colorTNMagenta       = "#bb9af7"
	colorTNCyan          = "#7dcfff"
	colorTNWhite         = "#a9b1d6"
	colorTNBrightBlack   = "#414868"
	colorTNBrightRed     = "#ff899d"
	colorTNBrightGreen   = "#9fe044"
	colorTNBrightYellow  = "#faba4a"
	colorTNBrightBlue    = "#8db0ff"
	colorTNBrightMagenta = "#c7a9ff"
	colorTNBrightCyan    = "#a4daff"
	colorTNBrightWhite   = "#c0caf5"
	colorTNOrange        = "#ff9e64"
	colorTNRed2          = "#db4b4b"

	// App semantic mapping (based on official palette above)
	colorFgPrimary    = colorTNFg
	colorFgSecondary  = colorTNWhite
	colorFgMuted      = colorTNBrightBlack
	colorSeparator    = colorTNBrightBlack
	colorPanel        = colorTNBg
	colorPanelAlt     = colorTNBlack
	colorAccent       = colorTNBlue
	colorSelectionBg  = colorTNBrightBlack
	colorMaskFg       = colorTNBrightBlack
	colorProgressCPU  = colorTNBlue
	colorProgressMEM  = colorTNCyan
	colorProgressGPU  = colorTNGreen
	colorProgressBase = colorTNBlack

	colorStatusReadyBg   = colorTNBlack
	colorStatusLoadingBg = colorTNBlack
	colorStatusErrorBg   = colorTNBlack
	colorStatusNeutralBg = colorTNBlack
	colorStatusRightBg   = colorTNBrightBlack

	colorStateIdle    = colorTNGreen
	colorStateAlloc   = colorTNBlue
	colorStateMixed   = colorTNYellow
	colorStateDrain   = colorTNRed
	colorStateDefault = colorTNFg
)

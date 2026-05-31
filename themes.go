package main

import "github.com/charmbracelet/lipgloss"

type theme struct {
	name    string
	bg      lipgloss.Color
	surface lipgloss.Color
	border  lipgloss.Color
	text    lipgloss.Color
	muted   lipgloss.Color
	primary lipgloss.Color
	accent  lipgloss.Color
	warn    lipgloss.Color
}

type colorScheme struct {
	name    string
	primary lipgloss.Color
	accent  lipgloss.Color
}

var themes = []theme{
	{
		name:    "catppuccin",
		bg:      lipgloss.Color("#1E1E2E"),
		surface: lipgloss.Color("#313244"),
		border:  lipgloss.Color("#45475A"),
		text:    lipgloss.Color("#CDD6F4"),
		muted:   lipgloss.Color("#6C7086"),
		primary: lipgloss.Color("#CBA6F7"),
		accent:  lipgloss.Color("#F5C2E7"),
		warn:    lipgloss.Color("#F9E2AF"),
	},
	{
		name:    "default",
		bg:      lipgloss.Color("#0F0F23"),
		surface: lipgloss.Color("#1E1E38"),
		border:  lipgloss.Color("#3A3A5C"),
		text:    lipgloss.Color("#F0F0F5"),
		muted:   lipgloss.Color("#8888A0"),
		primary: lipgloss.Color("#5B9BF5"),
		accent:  lipgloss.Color("#6EDFF7"),
		warn:    lipgloss.Color("#FFD700"),
	},
	{
		name:    "dracula",
		bg:      lipgloss.Color("#282A36"),
		surface: lipgloss.Color("#343746"),
		border:  lipgloss.Color("#44475A"),
		text:    lipgloss.Color("#F8F8F2"),
		muted:   lipgloss.Color("#6272A4"),
		primary: lipgloss.Color("#BD93F9"),
		accent:  lipgloss.Color("#FF79C6"),
		warn:    lipgloss.Color("#F1FA8C"),
	},
	{
		name:    "gruvbox",
		bg:      lipgloss.Color("#282828"),
		surface: lipgloss.Color("#3C3836"),
		border:  lipgloss.Color("#504945"),
		text:    lipgloss.Color("#EBDBB2"),
		muted:   lipgloss.Color("#A89984"),
		primary: lipgloss.Color("#FE8019"),
		accent:  lipgloss.Color("#B8BB26"),
		warn:    lipgloss.Color("#FABD2F"),
	},
	{
		name:    "monokai",
		bg:      lipgloss.Color("#272822"),
		surface: lipgloss.Color("#3E3D32"),
		border:  lipgloss.Color("#49483E"),
		text:    lipgloss.Color("#F8F8F2"),
		muted:   lipgloss.Color("#75715E"),
		primary: lipgloss.Color("#F92672"),
		accent:  lipgloss.Color("#A6E22E"),
		warn:    lipgloss.Color("#E6DB74"),
	},
	{
		name:    "nord",
		bg:      lipgloss.Color("#2E3440"),
		surface: lipgloss.Color("#3B4252"),
		border:  lipgloss.Color("#434C5E"),
		text:    lipgloss.Color("#ECEFF4"),
		muted:   lipgloss.Color("#7B88A1"),
		primary: lipgloss.Color("#88C0D0"),
		accent:  lipgloss.Color("#81A1C1"),
		warn:    lipgloss.Color("#EBCB8B"),
	},
	{
		name:    "one-dark",
		bg:      lipgloss.Color("#282C34"),
		surface: lipgloss.Color("#2C313A"),
		border:  lipgloss.Color("#3E4451"),
		text:    lipgloss.Color("#ABB2BF"),
		muted:   lipgloss.Color("#5C6370"),
		primary: lipgloss.Color("#61AFEF"),
		accent:  lipgloss.Color("#C678DD"),
		warn:    lipgloss.Color("#E5C07B"),
	},
	{
		name:    "solarized-dark",
		bg:      lipgloss.Color("#002B36"),
		surface: lipgloss.Color("#073642"),
		border:  lipgloss.Color("#586E75"),
		text:    lipgloss.Color("#839496"),
		muted:   lipgloss.Color("#586E75"),
		primary: lipgloss.Color("#268BD2"),
		accent:  lipgloss.Color("#2AA198"),
		warn:    lipgloss.Color("#B58900"),
	},
	{
		name:    "terminal",
		bg:      lipgloss.Color(""),
		surface: lipgloss.Color(""),
		border:  lipgloss.Color(""),
		text:    lipgloss.Color(""),
		muted:   lipgloss.Color(""),
		primary: lipgloss.Color(""),
		accent:  lipgloss.Color(""),
		warn:    lipgloss.Color(""),
	},
}

var colorSchemes = []colorScheme{
	{"blue", lipgloss.Color("#5B9BF5"), lipgloss.Color("#6EDFF7")},
	{"coral", lipgloss.Color("#FF6B6B"), lipgloss.Color("#FF8A8A")},
	{"cyan", lipgloss.Color("#5BF5E8"), lipgloss.Color("#7FFFFF")},
	{"emerald", lipgloss.Color("#10B981"), lipgloss.Color("#34D399")},
	{"gold", lipgloss.Color("#FBBF24"), lipgloss.Color("#FCD34D")},
	{"green", lipgloss.Color("#5BF5A0"), lipgloss.Color("#6EFFBF")},
	{"indigo", lipgloss.Color("#818CF8"), lipgloss.Color("#A5B4FC")},
	{"lavender", lipgloss.Color("#A78BFA"), lipgloss.Color("#C4B5FD")},
	{"lime", lipgloss.Color("#A3E635"), lipgloss.Color("#BEF264")},
	{"magenta", lipgloss.Color("#D946EF"), lipgloss.Color("#E879F9")},
	{"mint", lipgloss.Color("#34D399"), lipgloss.Color("#6EE7B7")},
	{"orange", lipgloss.Color("#F5A05B"), lipgloss.Color("#FFBF7F")},
	{"peach", lipgloss.Color("#FB923C"), lipgloss.Color("#FDBA74")},
	{"pink", lipgloss.Color("#F55BB8"), lipgloss.Color("#FF7FD0")},
	{"purple", lipgloss.Color("#B4A7F5"), lipgloss.Color("#D4CFFF")},
	{"red", lipgloss.Color("#F55B6E"), lipgloss.Color("#FF7F8E")},
	{"rose", lipgloss.Color("#FB7185"), lipgloss.Color("#FDA4AF")},
	{"sky", lipgloss.Color("#38BDF8"), lipgloss.Color("#7DD3FC")},
	{"teal", lipgloss.Color("#3CBFB4"), lipgloss.Color("#5CDFCA")},
	{"yellow", lipgloss.Color("#F5E85B"), lipgloss.Color("#FFFF7F")},
}

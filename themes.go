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

var themes = []theme{
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

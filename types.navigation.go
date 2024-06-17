package main

import "fyne.io/fyne/v2"

// Screen defines the data structure for the navigation pages
type Screen struct {
	Title, Intro string
	View         func(w fyne.Window) fyne.CanvasObject
	SupportWeb   bool
}

var (
	// Windows defines the metadata for each windowPane
	Windows = map[string]Screen{
		"welcome": {"Welcome", "", welcomeScreen, true},
	}

	// WindowIndex  defines how our windows should be laid out in the navigation index tree
	WindowIndex = map[string][]string{
		"": {"welcome"},
		//"":            {"welcome", "canvas", "animations", "icons", "widgets", "collections", "containers", "dialogs", "windows", "binding", "advanced"},
		//"collections": {"list", "table", "tree", "gridwrap"},
		//"containers":  {"apptabs", "border", "box", "center", "doctabs", "grid", "scroll", "split"},
		//"widgets":     {"accordion", "button", "card", "entry", "form", "input", "progress", "text", "toolbar"},
	}
)

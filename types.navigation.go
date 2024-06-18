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
		"vmware": {"VMWare Migration",
			"In this section you can connect to a VMWare vCenter server and build OpenShift compatible YAML manifests to migrate your network configuration.",
			vmwareScreen,
			true,
		},
		"vmwareconnection": {"Connection",
			"Connect to a VMWare vCenter server to begin building OpenShift compatible YAML manifests to migrate your network configuration.",
			makeVMWareConnectionTab,
			true,
		},
		"vmwareinventory": {"Inventory",
			"Select vCenter Inventory Objects",
			makeVMWareInventoryTab,
			true,
		},
		"vmwarenetworks": {"Networks",
			"List the available networks in vCenter",
			makeVMWareNetworksTab,
			true,
		},
	}

	// WindowIndex  defines how our windows should be laid out in the navigation index tree
	WindowIndex = map[string][]string{
		"":       {"welcome", "vmware"},
		"vmware": {"vmwareconnection", "vmwareinventory", "vmwarenetworks"},
		//"":            {"welcome", "canvas", "animations", "icons", "widgets", "collections", "containers", "dialogs", "windows", "binding", "advanced"},
		//"collections": {"list", "table", "tree", "gridwrap"},
		//"containers":  {"apptabs", "border", "box", "center", "doctabs", "grid", "scroll", "split"},
		//"widgets":     {"accordion", "button", "card", "entry", "form", "input", "progress", "text", "toolbar"},
	}
)

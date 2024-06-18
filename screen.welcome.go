package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func welcomeScreen(_ fyne.Window) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(OpenShiftLogoTransparent)
	logo.FillMode = canvas.ImageFillContain
	if fyne.CurrentDevice().IsMobile() {
		logo.SetMinSize(fyne.NewSize(192, 192))
	} else {
		logo.SetMinSize(fyne.NewSize(256, 256))
	}

	return container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle("Welcome to the OpenShift Network Wizard app", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		logo,
		container.NewHBox(
			widget.NewHyperlink("GitHub", parseURL("https://github.com/kenmoini/openshiftnetwizard/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("NMState Documentation", parseURL("https://nmstate.io/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("CNI Plugin Documentation", parseURL("https://www.cni.dev/plugins/current/main/")),
		),
		container.NewHBox(
			widget.NewLabel("Note: This app is not affiliated with Red Hat, Inc. or the OpenShift project."),
		),
		widget.NewLabel(""), // balance the header on the main screen we leave blank on this content
	))
}

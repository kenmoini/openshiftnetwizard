package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var topWindow fyne.Window
var screen fyne.Container
var split *container.Split
var application *fyne.App

func main() {
	application := app.NewWithID("io.kemo.openshiftnetwizard")
	logLifecycle(application)

	w := application.NewWindow("OpenShift Network Wizard")
	topWindow = w
	w.SetMaster()

	content := container.NewStack()
	title := widget.NewLabel("Component name")
	intro := widget.NewLabel("An introduction would probably go\nhere, as well as a")
	intro.Wrapping = fyne.TextWrapWord
	setScreen := func(s Screen) {
		if fyne.CurrentDevice().IsMobile() {
			child := application.NewWindow(s.Title)
			topWindow = child
			child.SetContent(s.View(topWindow))
			child.Show()
			child.SetOnClosed(func() {
				topWindow = w
			})
			return
		}

		title.SetText(s.Title)
		intro.SetText(s.Intro)

		content.Objects = []fyne.CanvasObject{s.View(w)}
		content.Refresh()
	}

	screen := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), intro), nil, nil, nil, content)
	if fyne.CurrentDevice().IsMobile() {
		w.SetContent(makeNav(setScreen, false))
	} else {
		split := container.NewHSplit(makeNav(setScreen, true), screen)
		split.Offset = 0.2
		w.SetContent(split)
	}

	w.Resize(fyne.NewSize(1024, 768))
	w.ShowAndRun()
}

package main

import (
	"net/url"

	"fyne.io/fyne/v2"
)

const preferenceCurrentScreen = "currentScreen"

func unsupportedScreen(s Screen) bool {
	return !s.SupportWeb && fyne.CurrentDevice().IsBrowser()
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

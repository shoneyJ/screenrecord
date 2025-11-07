package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func StartApp() {
	a := app.New()
	w := a.NewWindow("Screen Recorder")

	status := widget.NewLabel("Ready")
	startBtn := widget.NewButton("▶ Start", func() {
		status.SetText("Recording...")
	})
	stopBtn := widget.NewButton("⏹ Stop", func() {
		status.SetText("Stopped")
	})

	w.SetContent(container.NewVBox(
		status,
		startBtn,
		stopBtn,
	))

	w.Resize(fyne.NewSize(160, 130))
	w.CenterOnScreen()
	w.ShowAndRun()
}

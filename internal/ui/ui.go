package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/shoneyj/screenrecord/internal/config"
	"github.com/shoneyj/screenrecord/internal/ffmpeg"
)

var (
	a           fyne.App
	w           fyne.Window
	statusLabel *widget.Label
	timerLabel  *widget.Label
	startBtn    *widget.Button
	stopBtn     *widget.Button
	stopTimer   chan struct{}
	outputDir   string
)

func StartApp() {
	cfg, err := config.Load()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		outputDir = filepath.Join(homeDir, "Videos")
	} else {
		outputDir = cfg.OutputDirectory
	}

	a = app.New()
	w = a.NewWindow("Screen Recorder")

	statusLabel = widget.NewLabel("Ready")
	statusLabel.Alignment = fyne.TextAlignCenter

	timerLabel = widget.NewLabel("")
	timerLabel.Alignment = fyne.TextAlignCenter

	startBtn = widget.NewButton("▶ Start", func() {
		err := ffmpeg.RecordScreen(outputDir)
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Error: %v", err))
			return
		}

		w.SetTitle("Screen Recorder • Recording...")
		statusLabel.SetText("● Recording")
		startTimer()
		updateButtonStates(true)
	})

	stopBtn = widget.NewButton("⏹ Stop", func() {
		err := ffmpeg.StopRecording()
		if err != nil {
			statusLabel.SetText(fmt.Sprintf("Error: %v", err))
			return
		}

		w.SetTitle("Screen Recorder")
		statusLabel.SetText("Saved")
		stopTimerChan()
		updateButtonStates(false)

		cfg := &config.Config{OutputDirectory: outputDir}
		config.Save(cfg)
	})

	stopBtn.Disable()

	w.SetContent(container.NewVBox(
		widget.NewLabel("Screen Recorder"),
		statusLabel,
		container.NewHBox(startBtn, stopBtn),
		timerLabel,
	))

	w.Resize(fyne.NewSize(180, 140))
	w.CenterOnScreen()
	w.ShowAndRun()
}

func startTimer() {
	stopTimer = make(chan struct{})
	go func() {
		start := time.Now()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(start)
				mins := int(elapsed.Minutes())
				secs := int(elapsed.Seconds()) % 60
				timerLabel.SetText(fmt.Sprintf("%02d:%02d", mins, secs))
			case <-stopTimer:
				return
			}
		}
	}()
}

func stopTimerChan() {
	if stopTimer != nil {
		close(stopTimer)
		stopTimer = nil
	}
	timerLabel.SetText("")
}

func updateButtonStates(recording bool) {
	if recording {
		startBtn.Disable()
		stopBtn.Enable()
	} else {
		startBtn.Enable()
		stopBtn.Disable()
	}
}

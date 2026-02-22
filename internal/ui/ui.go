package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/shoneyj/screenrecord/internal/config"
	"github.com/shoneyj/screenrecord/internal/ffmpeg"
)

var (
	a             fyne.App
	w             fyne.Window
	timerLabel    *widget.Label
	toggleBtn     *widget.Button
	closeBtn      *widget.Button
	displaySelect *widget.Select
	stopTimer     chan struct{}
	outputDir     string
	displays      []ffmpeg.Display
	selectedDisp  ffmpeg.Display
	recordIcon    fyne.Resource
	stopIcon      fyne.Resource
)

func createRecordIcon() fyne.Resource {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#dc3545"><circle cx="12" cy="12" r="10"/></svg>`
	return fyne.NewStaticResource("record", []byte(svg))
}

func createStopIcon() fyne.Resource {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#6c757d"><rect x="4" y="4" width="16" height="16" rx="2"/></svg>`
	return fyne.NewStaticResource("stop", []byte(svg))
}

func StartApp() {
	cfg, err := config.Load()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		outputDir = filepath.Join(homeDir, "Videos")
	} else {
		outputDir = cfg.OutputDirectory
	}

	displays = ffmpeg.GetDisplays()
	if len(displays) > 0 {
		selectedDisp = displays[0]
	}

	recordIcon = createRecordIcon()
	stopIcon = createStopIcon()

	a = app.New()

	drv := fyne.CurrentApp().Driver()
	if d, ok := drv.(desktop.Driver); ok {
		w = d.CreateSplashWindow()
	} else {
		w = a.NewWindow("")
	}

	timerLabel = widget.NewLabel("00:00")
	timerLabel.Alignment = fyne.TextAlignCenter

	displayNames := make([]string, len(displays))
	for i, d := range displays {
		displayNames[i] = d.Name
	}

	displaySelect = widget.NewSelect(displayNames, func(s string) {
		for _, d := range displays {
			if d.Name == s {
				selectedDisp = d
				break
			}
		}
	})
	if len(displayNames) > 0 {
		displaySelect.SetSelected(displayNames[0])
	}

	toggleBtn = widget.NewButtonWithIcon("", recordIcon, toggleRecording)
	toggleBtn.Importance = widget.LowImportance

	closeBtn = widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		a.Quit()
	})
	closeBtn.Importance = widget.LowImportance

	displayOrTimer := container.NewStack(displaySelect, timerLabel)
	timerLabel.Hide()

	topBar := container.NewBorder(nil, nil, toggleBtn, closeBtn, container.NewCenter(displayOrTimer))

	w.SetContent(topBar)
	w.Resize(fyne.NewSize(220, 45))
	w.SetFixedSize(true)
	w.CenterOnScreen()
	w.ShowAndRun()
}

func toggleRecording() {
	if ffmpeg.IsRecording() {
		stopRecording()
	} else {
		startRecording()
	}
}

func startRecording() {
	err := ffmpeg.RecordScreen(outputDir, selectedDisp)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	toggleBtn.SetIcon(stopIcon)
	displaySelect.Hide()
	timerLabel.Show()
	startTimer()
}

func stopRecording() {
	err := ffmpeg.StopRecording()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	toggleBtn.SetIcon(recordIcon)
	timerLabel.Hide()
	displaySelect.Show()
	stopTimerChan()

	cfg := &config.Config{OutputDirectory: outputDir}
	config.Save(cfg)
}

func startTimer() {
	stopTimer = make(chan struct{})
	timerLabel.SetText("00:00")
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
				text := fmt.Sprintf("%02d:%02d", mins, secs)
				fyne.Do(func() {
					timerLabel.SetText(text)
				})
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

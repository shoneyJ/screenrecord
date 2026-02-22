package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

const (
	windowWidth  = 120
	windowHeight = 45
	margin       = 10
)

var (
	a                   fyne.App
	w                   fyne.Window
	timerLabel          *widget.Label
	toggleBtn           *widget.Button
	closeBtn            *widget.Button
	stopTimer           chan struct{}
	outputDir           string
	recordIcon          fyne.Resource
	stopIcon            fyne.Resource
	displays            []ffmpeg.Display
	currentDisplayIndex int
	cfg                 *config.Config
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
	var err error
	cfg, err = config.Load()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		outputDir = filepath.Join(homeDir, "Videos")
	} else {
		outputDir = cfg.OutputDirectory
	}

	displays = ffmpeg.GetDisplays()

	primary := ffmpeg.GetPrimaryDisplay()
	currentDisplayIndex = ffmpeg.GetDisplayIndex(primary.Name)

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
	timerLabel.Hide()

	toggleBtn = widget.NewButtonWithIcon("", recordIcon, toggleRecording)
	toggleBtn.Importance = widget.LowImportance

	closeBtn = widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		a.Quit()
	})
	closeBtn.Importance = widget.LowImportance

	content := container.NewBorder(nil, nil, toggleBtn, closeBtn, container.NewCenter(timerLabel))

	w.SetContent(content)
	w.Resize(fyne.NewSize(windowWidth, windowHeight))
	w.SetFixedSize(true)

	setupKeyboardShortcuts()

	go func() {
		time.Sleep(100 * time.Millisecond)
		positionWindowAtDisplay(primary)
	}()

	w.ShowAndRun()
}

func setupKeyboardShortcuts() {
	toggleShortcut := parseShortcut(cfg.Shortcuts.ToggleRecording)
	w.Canvas().AddShortcut(toggleShortcut, func(_ fyne.Shortcut) {
		toggleRecording()
	})

	cycleShortcut := parseShortcut(cfg.Shortcuts.CycleDisplay)
	w.Canvas().AddShortcut(cycleShortcut, func(_ fyne.Shortcut) {
		cycleDisplay()
	})
}

func parseShortcut(shortcut string) *desktop.CustomShortcut {
	parts := strings.Split(shortcut, "+")
	var modifier fyne.KeyModifier
	var key fyne.KeyName

	for _, p := range parts {
		switch strings.ToLower(p) {
		case "ctrl":
			modifier |= fyne.KeyModifierControl
		case "shift":
			modifier |= fyne.KeyModifierShift
		case "alt":
			modifier |= fyne.KeyModifierAlt
		case "super":
			modifier |= fyne.KeyModifierSuper
		default:
			key = fyne.KeyName(strings.ToUpper(p))
		}
	}

	return &desktop.CustomShortcut{KeyName: key, Modifier: modifier}
}

func positionWindowAtDisplay(display ffmpeg.Display) {
	x := display.X + display.Width - windowWidth - margin
	y := display.Y + margin
	moveWindowTo(x, y)
}

func moveWindowTo(x, y int) {
	winID, err := getActiveWindowID()
	if err != nil {
		fmt.Printf("⚠️  Could not get window ID: %v\n", err)
		return
	}
	exec.Command("xdotool", "windowmove", winID, strconv.Itoa(x), strconv.Itoa(y)).Run()
}

func getActiveWindowID() (string, error) {
	cmd := exec.Command("xdotool", "getwindowfocus")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func cycleDisplay() {
	if ffmpeg.IsRecording() {
		return
	}

	displayCount := len(displays)
	if displayCount <= 1 {
		return
	}

	currentDisplayIndex = (currentDisplayIndex + 1) % displayCount
	display := displays[currentDisplayIndex]
	positionWindowAtDisplay(display)
	fmt.Printf("🖥️  Switched to display: %s\n", display.Name)
}

func toggleRecording() {
	if ffmpeg.IsRecording() {
		stopRecording()
	} else {
		startRecording()
	}
}

func startRecording() {
	selectedDisplay := displays[currentDisplayIndex]

	err := ffmpeg.RecordScreen(outputDir, selectedDisplay)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	toggleBtn.SetIcon(stopIcon)
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
	stopTimerChan()

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

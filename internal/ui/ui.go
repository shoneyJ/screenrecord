package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/shoneyj/screenrecord/internal/config"
	"github.com/shoneyj/screenrecord/internal/ffmpeg"
)

const (
	windowWidth    = 40
	windowHeight   = 40
	menuWidth      = 100
	menuItemHeight = 30
	menuGap        = 5
	margin         = 10
)

var (
	a                   fyne.App
	w                   fyne.Window
	toggleIcon          *ClickableIcon
	outputDir           string
	recordIcon          fyne.Resource
	stopIcon            fyne.Resource
	gifIcon             fyne.Resource
	displays            []ffmpeg.Display
	currentDisplayIndex int
	cfg                 *config.Config
	isGifMode           bool
	autoStopTimer       *time.Timer
	timerMutex          sync.Mutex
	menuVisible         bool
	menuWindow          fyne.Window
)

func createRecordIcon() fyne.Resource {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#dc3545"><circle cx="12" cy="12" r="10"/></svg>`
	return fyne.NewStaticResource("record", []byte(svg))
}

func createStopIcon() fyne.Resource {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#6c757d"><rect x="4" y="4" width="16" height="16" rx="2"/></svg>`
	return fyne.NewStaticResource("stop", []byte(svg))
}

func createGifIcon() fyne.Resource {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#fd7e14"><circle cx="12" cy="12" r="10"/><text x="12" y="16" text-anchor="middle" font-size="10" font-weight="bold" fill="white">G</text></svg>`
	return fyne.NewStaticResource("gif", []byte(svg))
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

	isGifMode = cfg.RecordingMode == "gif"

	recordIcon = createRecordIcon()
	stopIcon = createStopIcon()
	gifIcon = createGifIcon()

	initialIcon := recordIcon
	if isGifMode {
		initialIcon = gifIcon
	}

	a = app.New()

	drv := fyne.CurrentApp().Driver()
	if d, ok := drv.(desktop.Driver); ok {
		w = d.CreateSplashWindow()
	} else {
		w = a.NewWindow("")
	}

	toggleIcon = NewClickableIcon(initialIcon, toggleRecording, showRightClickMenu)

	content := fyne.NewContainerWithoutLayout()
	content.Resize(fyne.NewSize(windowWidth, windowHeight))

	toggleIcon.Resize(fyne.NewSize(30, 30))
	toggleIcon.Move(fyne.NewPos(5, 5))

	content.Add(toggleIcon)

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

func showRightClickMenu(e *fyne.PointEvent) {
	if menuVisible {
		hideMenu()
		return
	}
	showMenu()
}

func getWindowPosition() (int, int, error) {
	cmd := exec.Command("xdotool", "getwindowfocus", "getwindowgeometry", "--shell")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	var x, y int
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "X=") {
			x, _ = strconv.Atoi(strings.TrimPrefix(line, "X="))
		} else if strings.HasPrefix(line, "Y=") {
			y, _ = strconv.Atoi(strings.TrimPrefix(line, "Y="))
		}
	}
	return x, y, nil
}

func showMenu() {
	mainX, mainY, err := getWindowPosition()
	if err != nil {
		fmt.Printf("Could not get window position: %v\n", err)
		return
	}

	menuVisible = true

	modeLabel := "GIF Mode"
	if isGifMode {
		modeLabel = "Video Mode"
	}

	drv := fyne.CurrentApp().Driver()
	if d, ok := drv.(desktop.Driver); ok {
		menuWindow = d.CreateSplashWindow()
	} else {
		menuWindow = a.NewWindow("")
	}

	modeBtn := widget.NewButton(modeLabel, func() {
		hideMenu()
		toggleMode()
	})
	modeBtn.Importance = widget.LowImportance

	quitBtn := widget.NewButton("Quit", func() {
		a.Quit()
	})
	quitBtn.Importance = widget.LowImportance

	menuHeight := float32(2 * menuItemHeight)
	content := fyne.NewContainerWithoutLayout()
	content.Resize(fyne.NewSize(menuWidth, menuHeight))
	modeBtn.Resize(fyne.NewSize(menuWidth, menuItemHeight))
	modeBtn.Move(fyne.NewPos(0, 0))
	quitBtn.Resize(fyne.NewSize(menuWidth, menuItemHeight))
	quitBtn.Move(fyne.NewPos(0, menuItemHeight))
	content.Add(modeBtn)
	content.Add(quitBtn)

	menuWindow.SetContent(content)
	menuWindow.Resize(fyne.NewSize(menuWidth, menuHeight))
	menuWindow.SetFixedSize(true)
	menuWindow.Show()

	go func() {
		time.Sleep(50 * time.Millisecond)
		menuX := mainX
		menuY := mainY + windowHeight + menuGap
		winID, err := getActiveWindowID()
		if err == nil {
			exec.Command("xdotool", "windowmove", winID, strconv.Itoa(menuX), strconv.Itoa(menuY)).Run()
		}
	}()
}

func hideMenu() {
	menuVisible = false
	if menuWindow != nil {
		menuWindow.Close()
		menuWindow = nil
	}
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

	modeShortcut := parseShortcut(cfg.Shortcuts.ToggleMode)
	w.Canvas().AddShortcut(modeShortcut, func(_ fyne.Shortcut) {
		toggleMode()
	})

	quitShortcut := parseShortcut(cfg.Shortcuts.QuitApp)
	w.Canvas().AddShortcut(quitShortcut, func(_ fyne.Shortcut) {
		a.Quit()
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
	if menuVisible {
		hideMenu()
		return
	}
	if ffmpeg.IsRecording() {
		stopRecording()
	} else {
		startRecording()
	}
}

func toggleMode() {
	if ffmpeg.IsRecording() {
		return
	}

	isGifMode = !isGifMode
	if isGifMode {
		cfg.RecordingMode = "gif"
		toggleIcon.SetIcon(gifIcon)
		fmt.Println("📷 Switched to GIF mode")
	} else {
		cfg.RecordingMode = "video"
		toggleIcon.SetIcon(recordIcon)
		fmt.Println("📹 Switched to Video mode")
	}
	config.Save(cfg)
}

func startRecording() {
	selectedDisplay := displays[currentDisplayIndex]

	err := ffmpeg.RecordScreen(outputDir, selectedDisplay)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	toggleIcon.SetIcon(stopIcon)

	if isGifMode {
		startAutoStopTimer()
	}
}

func startAutoStopTimer() {
	timerMutex.Lock()
	defer timerMutex.Unlock()

	autoStopTimer = time.AfterFunc(time.Duration(cfg.Gif.MaxDuration)*time.Second, func() {
		if ffmpeg.IsRecording() {
			fmt.Printf("⏱️ Auto-stopping GIF recording at %d seconds\n", cfg.Gif.MaxDuration)
			fyne.Do(func() {
				stopRecording()
			})
		}
	})
}

func stopAutoStopTimer() {
	timerMutex.Lock()
	defer timerMutex.Unlock()

	if autoStopTimer != nil {
		autoStopTimer.Stop()
		autoStopTimer = nil
	}
}

func stopRecording() {
	stopAutoStopTimer()

	wasGifMode := isGifMode
	outputPath := ffmpeg.GetOutputPath()

	err := ffmpeg.StopRecording()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if wasGifMode && outputPath != "" {
		fmt.Println("🔄 Converting to GIF...")
		go func() {
			gifPath, err := ffmpeg.ConvertToGif(outputPath, cfg.Gif.MaxWidth, cfg.Gif.Fps)
			if err != nil {
				fmt.Printf("Error converting to GIF: %v\n", err)
			} else {
				fmt.Printf("✅ GIF saved: %s\n", gifPath)
			}
		}()
	}

	if isGifMode {
		toggleIcon.SetIcon(gifIcon)
	} else {
		toggleIcon.SetIcon(recordIcon)
	}

	config.Save(cfg)
}


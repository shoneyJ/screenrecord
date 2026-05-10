package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	recordingMutex    sync.Mutex
	recordingCmd      *exec.Cmd
	isRecording       bool
	currentOutputPath string
)

type Display struct {
	Name        string
	Resolution  string
	DisplayName string
	X           int
	Y           int
	Width       int
	Height      int
	IsPrimary   bool
}

func getBaseDisplay() string {
	display := os.Getenv("DISPLAY")
	if display == "" {
		return ":0"
	}
	return strings.Split(display, ".")[0]
}

func GetDisplays() []Display {
	displays := []Display{}
	baseDisplay := getBaseDisplay()

	cmd := exec.Command("xrandr", "--query")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  Could not detect displays via xrandr: %v\n", err)
		return []Display{{Name: "Default", Resolution: getScreenResolution(), DisplayName: baseDisplay + ".0+0,0", IsPrimary: true}}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, " connected") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				name := fields[0]
				res := ""
				displayInput := ""
				x, y := 0, 0
				width, height := 0, 0
				isPrimary := strings.Contains(line, " primary ")

				for i, f := range fields {
					if strings.Contains(f, "x") && strings.Contains(f, "+") {
						parts := strings.Split(f, "+")
						res = parts[0]
						if len(parts) >= 3 {
							x, _ = strconv.Atoi(parts[1])
							y, _ = strconv.Atoi(parts[2])
						}
						dimParts := strings.Split(res, "x")
						if len(dimParts) == 2 {
							width, _ = strconv.Atoi(dimParts[0])
							height, _ = strconv.Atoi(dimParts[1])
						}
						break
					}
					if f == "primary" && i+1 < len(fields) {
						next := fields[i+1]
						if strings.Contains(next, "x") && strings.Contains(next, "+") {
							parts := strings.Split(next, "+")
							res = parts[0]
							if len(parts) >= 3 {
								x, _ = strconv.Atoi(parts[1])
								y, _ = strconv.Atoi(parts[2])
							}
							dimParts := strings.Split(res, "x")
							if len(dimParts) == 2 {
								width, _ = strconv.Atoi(dimParts[0])
								height, _ = strconv.Atoi(dimParts[1])
							}
						}
					}
				}

				if res == "" {
					res = getScreenResolution()
					dimParts := strings.Split(res, "x")
					if len(dimParts) == 2 {
						width, _ = strconv.Atoi(dimParts[0])
						height, _ = strconv.Atoi(dimParts[1])
					}
				}
				displayInput = fmt.Sprintf("%s+%d,%d", baseDisplay, x, y)

				displays = append(displays, Display{
					Name:        name,
					Resolution:  res,
					DisplayName: displayInput,
					X:           x,
					Y:           y,
					Width:       width,
					Height:      height,
					IsPrimary:   isPrimary,
				})
			}
		}
	}

	if len(displays) == 0 {
		displays = append(displays, Display{
			Name:        "Default",
			Resolution:  getScreenResolution(),
			DisplayName: baseDisplay + "+0,0",
			X:           0,
			Y:           0,
			Width:       1600,
			Height:      900,
			IsPrimary:   true,
		})
	}

	return displays
}

func getScreenResolution() string {
	fmt.Printf("🔍 Detecting screen resolution...\n")
	fmt.Printf("   XDG_SESSION_TYPE: %s\n", os.Getenv("XDG_SESSION_TYPE"))
	fmt.Printf("   WAYLAND_DISPLAY: %s\n", os.Getenv("WAYLAND_DISPLAY"))

	resolution := "1600x900"

	modes, err := os.ReadDir("/sys/class/drm")
	if err == nil {
		for _, m := range modes {
			if m.IsDir() && (strings.HasSuffix(m.Name(), "-eDP-1") || strings.HasSuffix(m.Name(), "-HDMI-A-1")) {
				modeData, err := os.ReadFile("/sys/class/drm/" + m.Name() + "/modes")
				if err == nil && len(modeData) > 0 {
					resolution = strings.TrimSpace(string(modeData))
					break
				}
			}
		}
	}

	fmt.Printf("📺 Using screen resolution: %s\n", resolution)
	return resolution
}

func getDisplayInput() string {
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}
	baseDisplay := strings.Split(display, ".")[0]
	return baseDisplay + ".0+0,0"
}

func isWayland() bool {
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	isWl := sessionType == "wayland" || waylandDisplay != ""
	fmt.Printf("🔍 Is Wayland: %v (session=%s, display=%s)\n", isWl, sessionType, waylandDisplay)
	return isWl
}

func isNVIDIAAvailable() bool {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("🔍 No NVIDIA GPU detected\n")
		return false
	}
	gpuName := strings.TrimSpace(string(output))
	fmt.Printf("🔍 NVIDIA GPU detected: %s\n", gpuName)
	return true
}

func RecordScreen(outputPath string, display Display) error {
	recordingMutex.Lock()
	if isRecording {
		recordingMutex.Unlock()
		return fmt.Errorf("already recording")
	}
	recordingMutex.Unlock()

	outputFile := outputPath
	if filepath.Ext(outputFile) == "" {
		outputFile = filepath.Join(outputPath, fmt.Sprintf("recording_%s.mp4", time.Now().Format("2006-01-02_15-04-05")))
	}

	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	currentOutputPath = outputFile

	screenRes := display.Resolution
	displayInput := display.DisplayName
	if displayInput == "" {
		displayInput = getDisplayInput()
	}
	if screenRes == "" {
		screenRes = getScreenResolution()
	}

	fmt.Printf("📺 Display: %s (%s)\n", display.Name, screenRes)

	var cmd *exec.Cmd
	if isNVIDIAAvailable() {
		fmt.Printf("📹 Recording with NVIDIA NVENC\n")
		cmd = exec.Command(
			"ffmpeg",
			"-f", "x11grab",
			"-framerate", "30",
			"-draw_mouse", "1",
			"-s", screenRes,
			"-i", displayInput,
			"-c:v", "h264_nvenc",
			"-preset", "fast",
			"-pix_fmt", "yuv420p",
			"-movflags", "+faststart",
			"-y",
			outputFile,
		)
	} else {
		fmt.Printf("📹 Recording with CPU (libx264)\n")
		cmd = exec.Command(
			"ffmpeg",
			"-f", "x11grab",
			"-framerate", "30",
			"-draw_mouse", "1",
			"-s", screenRes,
			"-i", displayInput,
			"-c:v", "libx264",
			"-preset", "ultrafast",
			"-tune", "zerolatency",
			"-pix_fmt", "yuv420p",
			"-movflags", "+faststart",
			"-y",
			outputFile,
		)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	recordingMutex.Lock()
	recordingCmd = cmd
	isRecording = true
	recordingMutex.Unlock()

	fmt.Printf("🔴 Recording started: %s\n", outputFile)
	return nil
}

func StopRecording() error {
	recordingMutex.Lock()
	defer recordingMutex.Unlock()

	if !isRecording || recordingCmd == nil {
		return fmt.Errorf("not currently recording")
	}

	if recordingCmd.Process != nil {
		if err := recordingCmd.Process.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("failed to stop recording: %w", err)
		}
		recordingCmd.Wait()
	}

	isRecording = false
	recordingCmd = nil

	fmt.Println("⏹ Recording stopped")
	return nil
}

func IsRecording() bool {
	recordingMutex.Lock()
	defer recordingMutex.Unlock()
	return isRecording
}

func GetDisplayFromPosition(winX, winY, winWidth, winHeight int) Display {
	displays := GetDisplays()

	centerX := winX + winWidth/2
	centerY := winY + winHeight/2

	for _, d := range displays {
		if centerX >= d.X && centerX < d.X+d.Width &&
			centerY >= d.Y && centerY < d.Y+d.Height {
			fmt.Printf("🖥️  Window on display: %s (%s)\n", d.Name, d.Resolution)
			return d
		}
	}

	if len(displays) > 0 {
		fmt.Printf("🖥️  Using default display: %s\n", displays[0].Name)
		return displays[0]
	}

	return Display{
		Name:        "Default",
		Resolution:  "1600x900",
		DisplayName: ":0+0,0",
		X:           0,
		Y:           0,
		Width:       1600,
		Height:      900,
		IsPrimary:   true,
	}
}

func GetPrimaryDisplay() Display {
	displays := GetDisplays()
	for _, d := range displays {
		if d.IsPrimary {
			return d
		}
	}
	if len(displays) > 0 {
		return displays[0]
	}
	return Display{
		Name:        "Default",
		Resolution:  "1600x900",
		DisplayName: ":0+0,0",
		X:           0,
		Y:           0,
		Width:       1600,
		Height:      900,
		IsPrimary:   true,
	}
}

func GetDisplayByIndex(index int) Display {
	displays := GetDisplays()
	if index >= 0 && index < len(displays) {
		return displays[index]
	}
	return GetPrimaryDisplay()
}

func GetDisplayCount() int {
	return len(GetDisplays())
}

func GetDisplayIndex(name string) int {
	displays := GetDisplays()
	for i, d := range displays {
		if d.Name == name {
			return i
		}
	}
	return 0
}

func GetOutputPath() string {
	recordingMutex.Lock()
	defer recordingMutex.Unlock()
	return currentOutputPath
}

func ConvertToGif(inputPath string, maxWidth, fps int) (string, error) {
	outputPath := strings.TrimSuffix(inputPath, ".mp4") + ".gif"
	palettePath := strings.TrimSuffix(inputPath, ".mp4") + "_palette.png"

	scaleFilter := fmt.Sprintf("fps=%d,scale='min(%d,iw)':-2:flags=lanczos", fps, maxWidth)

	// Pass 1: generate palette with stats_mode=full for accurate colors across all frames
	fmt.Printf("🔄 Generating palette: %s\n", palettePath)
	pass1 := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-vf", scaleFilter+",palettegen=max_colors=256:stats_mode=full",
		"-y",
		palettePath,
	)
	pass1.Stdout = os.Stdout
	pass1.Stderr = os.Stderr

	if err := pass1.Run(); err != nil {
		return "", fmt.Errorf("failed to generate palette: %w", err)
	}

	// Pass 2: apply palette with dithering optimized for screen content
	fmt.Printf("🔄 Converting to GIF: %s\n", outputPath)
	pass2 := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-i", palettePath,
		"-lavfi", scaleFilter+"[x];[x][1:v]paletteuse=dither=sierra2_4a",
		"-y",
		outputPath,
	)
	pass2.Stdout = os.Stdout
	pass2.Stderr = os.Stderr

	if err := pass2.Run(); err != nil {
		os.Remove(palettePath)
		return "", fmt.Errorf("failed to convert to GIF: %w", err)
	}

	os.Remove(palettePath)
	fmt.Printf("✅ GIF created: %s\n", outputPath)
	return outputPath, nil
}

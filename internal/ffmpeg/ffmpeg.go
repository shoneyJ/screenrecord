package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	recordingMutex sync.Mutex
	recordingCmd   *exec.Cmd
	isRecording    bool
)

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

func RecordScreen(outputPath string) error {
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

	screenRes := getScreenResolution()

	var cmd *exec.Cmd
	if isNVIDIAAvailable() {
		fmt.Printf("📹 Recording with NVIDIA NVENC\n")
		cmd = exec.Command(
			"ffmpeg",
			"-f", "x11grab",
			"-framerate", "30",
			"-draw_mouse", "1",
			"-s", screenRes,
			"-i", ":0.0+0,0",
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
			"-i", ":0.0+0,0",
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

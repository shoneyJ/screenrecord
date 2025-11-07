package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
)

type fileInfo struct {
	path string
	mod  time.Time
}

func getFileList(sourceDir string) ([]string, error) {

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []fileInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Optional: only include certain file extensions
		ext := filepath.Ext(entry.Name())
		switch ext {
		case ".mp4", ".mov", ".mkv", ".avi":
			info, err := entry.Info()
			if err != nil {
				return nil, err
			}
			files = append(files, fileInfo{
				path: filepath.Join(sourceDir, entry.Name()),
				mod:  info.ModTime(),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].mod.Before(files[j].mod)
	})

	sortedPaths := make([]string, len(files))
	for i, f := range files {
		sortedPaths[i] = f.path
	}

	return sortedPaths, nil

}

func JoinListOfVideos(sourceDir string, outputFile string) error {

	files, err := getFileList(sourceDir)

	if len(files) == 0 {
		return fmt.Errorf("no input files provided")
	}

	listFile := "list.txt"
	f, err := os.Create(listFile)

	defer os.Remove(listFile)
	defer f.Close()

	for _, file := range files {
		_, err := fmt.Fprintf(f, "file '%s'\n", file)
		if err != nil {
			return fmt.Errorf("failed to write to list file: %w", err)
		}
	}
	f.Sync()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading files: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command(
		"ffmpeg",
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c:v", "h264_nvenc",
		"-preset", "fast",
		"-b:v", "5M",
		"-c:a", "aac",
		"-b:a", "192k",
		outputFile,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("🔄 Merging videos from list.txt...")

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w", err)
	}

	fmt.Println("✅ Videos merged successfully into output.mp4")
	return nil
}

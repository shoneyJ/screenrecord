package main

import (
	"github.com/shoneyj/screenrecord/internal/ffmpeg"
)

func main() {
	// ui.StartApp()

	ffmpeg.JoinListOfVideos("/home/shoney/Downloads/screenrecord/review1", "/home/shoney/Downloads/screenrecord/review1/outputnew.mp4")
}

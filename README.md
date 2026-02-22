# Screen Recorder

A minimalist screen recording application for Linux.

## Features

- One-click start/stop recording
- Automatic GPU detection (NVIDIA NVENC hardware encoding)
- CPU fallback when no GPU available
- Recording timer
- Configurable output directory
- Minimalist widget-like UI

## Requirements

- **FFmpeg** - for screen capture and encoding
- **NVIDIA GPU** (optional) - for hardware encoding
- **X11 or XWayland** session

### Install Dependencies

#### Arch Linux
```bash
sudo pacman -S ffmpeg
```

#### Debian/Ubuntu
```bash
sudo apt install ffmpeg
```

#### Fedora
```bash
sudo dnf install ffmpeg
```

### For NVIDIA GPU Support (Optional)
```bash
# Arch
sudo pacman -S nvidia-utils

# Debian/Ubuntu
sudo apt install nvidia-driver-535
```

## Installation

### Quick Install
```bash
chmod +x install.sh
./install.sh
```

### Manual Install
```bash
# Build the application
go build -o screenrecord ./cmd/screenrecorder

# Install to /usr/local/bin
sudo cp screenrecord /usr/local/bin/
```

## Usage

```bash
screenrecord
```

### Controls
- Click **Start** to begin recording
- Click **Stop** to end recording
- Video is saved to `~/Videos/recording_<timestamp>.mp4`

### Configuration
Config file is stored at: `~/.config/screenrecord/config.json`

## Building from Source

### Prerequisites
- Go 1.21+
- FFmpeg

### Build
```bash
go build -o screenrecord ./cmd/screenrecorder
```

## Technical Details

### Encoding Options

| System | Encoder | FFmpeg Flag |
|--------|---------|-------------|
| NVIDIA GPU | NVENC (Hardware) | `-c:v h264_nvenc` |
| No GPU | CPU | `-c:v libx264` |

### Screen Capture

- **X11**: `ffmpeg -f x11grab`
- **XWayland**: `ffmpeg -f x11grab` (via XWayland)

## License

MIT

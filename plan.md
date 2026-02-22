# Screen Recording App - Implementation Plan

## Use Case

As a software developer, I often need to record screen for short duration. Using tools such as OBS studio is overkill. I need a quick start record, stop record functionality with widget like UI.

## Criterias

1. On start recording, the screen should start recording.
2. Only ffmpeg should be used as video encoder.
3. Detect GPU, if yes then encode with gpu. Focus on nvidia first.

## Out of Scope

- Audio recording is not required.
- Windows and MacOS support is not required.

---

# Implementation Plan

## Phase 1: Core Recording (Priority: Critical) ✅

### Goals

Get basic screen recording working reliably.

### Tasks

- [x] Basic UI with Start/Stop buttons
- [x] Output directory selector
- [x] Status display
- [x] Auto-detect screen resolution
- [x] Fix Wayland recording (wf-recorder codec issue)
- [x] Test on X11 (ffmpeg x11grab)

### Technical Details

- **X11**: `ffmpeg -f x11grab -framerate 30 -s RES -i :0.0 -c:v libx264 -preset ultrafast -pix_fmt yuv420p`
- **Wayland**: `wf-recorder -c libx264 -p preset=ultrafast -r 30 -f output.mp4`

---

## Phase 2: UI Enhancements (Priority: High)

### Goals

Improve user experience with visual feedback.

### Tasks

- [x] Add recording duration timer (shows elapsed time)
- [x] Add recording indicator (red blinking dot)
- [x] Update window title during recording: "Screen Recorder • Recording..."
- [x] Config file for output directory

### UI Mockup (Phase 2)

```
───────────────────────
│  DP1                    │
│  [▶]       00:15   X │
└───────────────────────
```

- [] window should not be resizable.
- [] No text is required.
- [] Record and stop button should be togleable.
- [] In a multi monitor setup, option to choose the display must be available.
- [] Add draggable functionality to the window
- [] Add logic to detect which monitor the window is on based on window position
- [] Store current display selection internally (not shown to user)

---

## Phase 3: GPU Detection & Encoding (Priority: High) ✅

### Goals

Detect NVIDIA GPU and use hardware encoder (NVENC) for better performance.

### Tasks

- [ ] Add detectNVIDIA() function in ffmpeg.go
- [ ] Add isNVIDIAAvailable() function
- [ ] Modify RecordScreen() to use h264_nvenc when GPU available
- [ ] Fallback to libx264 when no GPU

### Technical Details

| System                      | Encoder | FFmpeg Flag                    |
| --------------------------- | ------- | ------------------------------ |
| NVIDIA GPU                  | NVENC   | `-c:v h264_nvenc -preset fast` |
| No GPU / Fall `-c:v libback | CPU     | x264 -preset ultrafast`        |

### Detection Method

```bash
# Check for NVIDIA GPU
nvidia-smi --query-gpu=name --format=csv,noheader
```

### Implementation

- Silent detection (no UI changes)
- Check for NVIDIA GPU on app start
- Use h264_nvenc encoder when available
- Fallback to libx264 CPU encoding otherwise

---

## Phase 4: Keyboard Shortcuts & Polish (Priority: Medium)

### Goals

Quick access and better UX.

### Tasks

- [ ] Add keyboard shortcut: Ctrl+R to toggle recording
- [ ] Show desktop notification when recording saves

---

## Phase 5: Future Enhancements (Nice to Have)

- System tray icon
- Custom resolution selection
- Frame rate options
- Output format selection (mp4, avi, mkv)
- AMD GPU support (VAAPI)
- Intel GPU support (QuickSync)

---

## Technical Architecture

```
┌─────────────────────────────────────────┐
│            cmd/screenrecorder            │
│                  main.go                 │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│            internal/ui/                  │
│                  ui.go                  │
│  - Fyne UI components                   │
│  - Button handlers                      │
│  - Status updates                       │
│  - Timer                                │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│            internal/ffmpeg/              │
│                ffmpeg.go                │
│  - RecordScreen()                      │
│  - StopRecording()                     │
│  - isWayland() detection               │
│  - getScreenResolution()               │
│  - isNVIDIAAvailable()                 │
│  - detectNVIDIA()                       │
└─────────────────────────────────────────┘
                  │
                  ▼
         ┌────────┴────────┐
         │   External     │
         │   Tools        │
         ├────────────────┤
         │ ffmpeg x11grab │
         │ (CPU or NVENC) │
         └────────────────┘
```

---

## Testing Checklist

- [x] Start recording - screen capture begins
- [x] Stop recording - file saves successfully
- [x] Video plays in mpv
- [x] Duration timer updates correctly
- [x] Window title shows recording status
- [x] Config file saves output directory
- [ ] GPU encoding works when NVIDIA available
- [ ] CPU fallback works when no GPU

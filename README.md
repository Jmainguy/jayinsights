# JayInsights

JayInsights is a graphical system information and sensor dashboard for Linux.

## Features

- Displays CPU, RAM, GPU, motherboard, drives, fans, and temperature information.
- Shows per-core CPU speeds, RAM bank details, GPU VBIOS version, and more.
- Customizable fan labels via YAML config.
- Modern, compact UI using Fyne.

## Requirements

- **Linux only**: This tool is designed for Linux systems.
- **Root/Sudo access**: Some hardware information (e.g., DMI, sensors, drives) requires elevated permissions. Run with `sudo` for full data.
- **Graphical environment**: Requires X11, Wayland, or similar desktop environment.
- **OpenGL support**: For GPU info, OpenGL libraries and drivers must be present.

## Usage

1. Run: `sudo jayinsights`
2. Optionally, customize fan labels in `~/.config/jayinsights/config.yaml`.
3. See config.yaml in thisd repo as an example of how to label fans.

## Notes

- Some data may be missing if run without sudo.
- Only works on Linux; not compatible with Windows or macOS.
- Requires a graphical session (cannot run in pure SSH or TTY).

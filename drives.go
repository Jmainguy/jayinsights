package main

import (
	"fmt"
	"os"
	"strings"
)

func GetDrives() []string {
	drives := []string{}
	files, _ := os.ReadDir("/sys/block/")
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "sd") || strings.HasPrefix(f.Name(), "nvme") {
			drives = append(drives, f.Name())
		}
	}
	return drives
}

func GetDriveModel(dev string) string {
	path := fmt.Sprintf("/sys/block/%s/device/model", dev)
	data, err := os.ReadFile(path)
	if err != nil {
		return "N/A"
	}
	return strings.TrimSpace(string(data))
}

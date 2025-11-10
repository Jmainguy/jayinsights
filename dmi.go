package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetDMIInfo() (ramInfo, boardInfo, biosInfo string, coreCount, threadCount string) {
	ramInfo = ""
	memSize, _ := os.ReadFile("/sys/class/dmi/id/memory_size")
	memType, _ := os.ReadFile("/sys/class/dmi/id/memory_type")
	memSpeed, _ := os.ReadFile("/sys/class/dmi/id/memory_speed")
	memBank, _ := os.ReadFile("/sys/class/dmi/id/memory_bank_locator")
	if len(memSize) > 0 || len(memType) > 0 || len(memSpeed) > 0 {
		ramInfo = fmt.Sprintf("Size: %s\nType: %s\nSpeed: %s\nBank: %s", strings.TrimSpace(string(memSize)), strings.TrimSpace(string(memType)), strings.TrimSpace(string(memSpeed)), strings.TrimSpace(string(memBank)))
	} else {
		ramInfo = "N/A"
	}

	boardVendor, _ := os.ReadFile("/sys/class/dmi/id/board_vendor")
	boardName, _ := os.ReadFile("/sys/class/dmi/id/board_name")
	boardVersion, _ := os.ReadFile("/sys/class/dmi/id/board_version")
	boardSerial, _ := os.ReadFile("/sys/class/dmi/id/board_serial")
	boardInfo = fmt.Sprintf("Vendor: %s\nName: %s\nVersion: %s\nSerial: %s", strings.TrimSpace(string(boardVendor)), strings.TrimSpace(string(boardName)), strings.TrimSpace(string(boardVersion)), strings.TrimSpace(string(boardSerial)))

	biosVendor, _ := os.ReadFile("/sys/class/dmi/id/bios_vendor")
	biosVersion, _ := os.ReadFile("/sys/class/dmi/id/bios_version")
	biosDate, _ := os.ReadFile("/sys/class/dmi/id/bios_date")
	biosRelease, _ := os.ReadFile("/sys/class/dmi/id/bios_release")
	biosInfo = fmt.Sprintf("Vendor: %s\nVersion: %s\nDate: %s\nRevision: %s", strings.TrimSpace(string(biosVendor)), strings.TrimSpace(string(biosVersion)), strings.TrimSpace(string(biosDate)), strings.TrimSpace(string(biosRelease)))

	cpuPath := "/sys/devices/system/cpu/"
	cpus, _ := os.ReadDir(cpuPath)
	coreCountInt := 0
	threadCountInt := 0
	for _, cpu := range cpus {
		if strings.HasPrefix(cpu.Name(), "cpu") && len(cpu.Name()) > 3 {
			threadCountInt++
			onlinePath := filepath.Join(cpuPath, cpu.Name(), "online")
			online, err := os.ReadFile(onlinePath)
			if err == nil && strings.TrimSpace(string(online)) == "1" {
				coreCountInt++
			}
		}
	}
	coreCount = fmt.Sprintf("%d", coreCountInt)
	threadCount = fmt.Sprintf("%d", threadCountInt)
	return ramInfo, boardInfo, biosInfo, coreCount, threadCount
}

package main

import (
	"fmt"
	"os"
	"strings"
)

func GetCPUInfo() (model string, mhz string, cores string, threads string) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return model, mhz, "N/A", "N/A"
	}
	blocks := strings.Split(string(data), "\n\n")
	model = "N/A"
	coreSpeedMap := make(map[string]string)
	coreIDSet := make(map[string]struct{})
	threadCount := 0
	physIDPresent := false
	for _, block := range blocks {
		physID := ""
		coreID := ""
		speed := ""
		lines := strings.Split(block, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "physical id") {
				physID = strings.TrimSpace(strings.Split(line, ":")[1])
				physIDPresent = true
			}
			if strings.HasPrefix(line, "core id") {
				coreID = strings.TrimSpace(strings.Split(line, ":")[1])
			}
			if strings.HasPrefix(line, "processor") {
				threadCount++
			}
			if strings.HasPrefix(line, "model name") {
				model = strings.TrimSpace(strings.Split(line, ":")[1])
			}
			if strings.HasPrefix(line, "cpu MHz") {
				speed = strings.TrimSpace(strings.Split(line, ":")[1])
			}
		}
		if physID != "" && coreID != "" && speed != "" {
			key := physID + "-" + coreID
			coreSpeedMap[key] = speed
		}
		if coreID != "" {
			coreIDSet[coreID] = struct{}{}
		}
	}
	var speeds []string
	var coreCount int
	if physIDPresent {
		// Sort keys for display
		type coreKey struct{ phys, core int }
		var keys []coreKey
		for k := range coreSpeedMap {
			var phys, core int
			fmt.Sscanf(k, "%d-%d", &phys, &core)
			keys = append(keys, coreKey{phys, core})
		}
		// Sort by phys then core
		for i := 0; i < len(keys)-1; i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i].phys > keys[j].phys || (keys[i].phys == keys[j].phys && keys[i].core > keys[j].core) {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
		for _, k := range keys {
			speeds = append(speeds, coreSpeedMap[fmt.Sprintf("%d-%d", k.phys, k.core)])
		}
		coreCount = len(coreSpeedMap)
	} else if len(coreIDSet) > 0 {
		// Fallback: use unique core ids
		seen := make(map[string]bool)
		for _, block := range blocks {
			coreID := ""
			speed := ""
			lines := strings.Split(block, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "core id") {
					coreID = strings.TrimSpace(strings.Split(line, ":")[1])
				}
				if strings.HasPrefix(line, "cpu MHz") {
					speed = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
			if coreID != "" && speed != "" && !seen[coreID] {
				speeds = append(speeds, speed)
				seen[coreID] = true
			}
		}
		coreCount = len(coreIDSet)
	} else {
		// Fallback: show all thread speeds
		for _, block := range blocks {
			speed := ""
			lines := strings.Split(block, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "cpu MHz") {
					speed = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
			if speed != "" {
				speeds = append(speeds, speed)
			}
		}
		coreCount = threadCount
	}
	return model, strings.Join(speeds, ","), fmt.Sprintf("%d", coreCount), fmt.Sprintf("%d", threadCount)
}

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readSensors() SensorData {
	temps := map[string]float64{}
	fans := map[string]int{}

	// Read from /sys/class/hwmon/hwmon*/
	hwmonBase := "/sys/class/hwmon/"
	if hwmons, err := os.ReadDir(hwmonBase); err == nil {
		for _, hw := range hwmons {
			hwPath := hwmonBase + hw.Name() + "/"
			// Try to get sensor name
			name := hw.Name()
			if nBytes, err := os.ReadFile(hwPath + "name"); err == nil {
				name = strings.TrimSpace(string(nBytes))
			}
			// Find temp sensors and use temp*_label if available
			for i := 1; i <= 10; i++ {
				tPath := fmt.Sprintf("%stemp%d_input", hwPath, i)
				labelPath := fmt.Sprintf("%stemp%d_label", hwPath, i)
				var label string
				if lBytes, err := os.ReadFile(labelPath); err == nil {
					label = strings.TrimSpace(string(lBytes))
				} else {
					label = fmt.Sprintf("%s Temp%d", name, i)
				}
				if tBytes, err := os.ReadFile(tPath); err == nil {
					tVal, err := strconv.ParseFloat(strings.TrimSpace(string(tBytes)), 64)
					if err == nil {
						// hwmon reports in millidegrees C
						temps[label] = tVal / 1000.0
					}
				}
			}
			// Find fan sensors
			for i := 1; i <= 10; i++ {
				fPath := fmt.Sprintf("%sfan%d_input", hwPath, i)
				if fBytes, err := os.ReadFile(fPath); err == nil {
					fVal, err := strconv.Atoi(strings.TrimSpace(string(fBytes)))
					if err == nil {
						fans[fmt.Sprintf("%s Fan%d", name, i)] = fVal
					}
				}
			}
		}
	}

	// Also try /sys/class/thermal/thermal_zone*/temp for generic temps
	thermalBase := "/sys/class/thermal/"
	if thermals, err := os.ReadDir(thermalBase); err == nil {
		for _, th := range thermals {
			if strings.HasPrefix(th.Name(), "thermal_zone") {
				tPath := thermalBase + th.Name() + "/temp"
				if tBytes, err := os.ReadFile(tPath); err == nil {
					tVal, err := strconv.ParseFloat(strings.TrimSpace(string(tBytes)), 64)
					if err == nil {
						temps[th.Name()] = tVal / 1000.0
					}
				}
			}
		}
	}

	return SensorData{
		Temperatures: temps,
		FanSpeeds:    fans,
	}
}

type SensorData struct {
	Temperatures map[string]float64
	FanSpeeds    map[string]int
}

func categorizeSensors(sensor SensorData) (moboTemps, cpuTemps, gpuTemps, hdTemps map[string]float64, cpuFans, gpuFans, moboFans map[string]int) {
	moboTemps = map[string]float64{}
	cpuTemps = map[string]float64{}
	gpuTemps = map[string]float64{}
	hdTemps = map[string]float64{}
	cpuFans = map[string]int{}
	gpuFans = map[string]int{}
	moboFans = map[string]int{}
	for k, v := range sensor.Temperatures {
		lk := strings.ToLower(k)
		// Only include 'Core N' temps in cpuTemps, ignore 'Package id' and other package temps
		if strings.HasPrefix(lk, "core ") {
			cpuTemps[k] = v
			continue
		}
		switch {
		case strings.Contains(lk, "gpu"):
			gpuTemps[k] = v
		case strings.Contains(lk, "hd"), strings.Contains(lk, "nvme"), strings.Contains(lk, "disk"):
			hdTemps[k] = v
		case strings.Contains(lk, "mobo"), strings.Contains(lk, "board"), strings.Contains(lk, "pch"):
			moboTemps[k] = v
		}
	}
	for k, v := range sensor.FanSpeeds {
		lk := strings.ToLower(k)
		switch {
		case strings.Contains(lk, "cpu"):
			cpuFans[k] = v
		case strings.Contains(lk, "gpu"):
			gpuFans[k] = v
		case strings.Contains(lk, "mobo"), strings.Contains(lk, "board"), strings.Contains(lk, "pch"):
			moboFans[k] = v
		}
	}
	return
}

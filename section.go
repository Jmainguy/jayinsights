package main

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func MakeSection(title string, sensors map[string]float64, fans map[string]int, tempThreshold float64, showNoData bool, fanLabelMap map[string]string) fyne.CanvasObject {
	if title == "CPU Temp" {
		// Dynamically show all detected core temps, sorted, each on its own line, with color
		coreKeys := []int{}
		coreMap := map[int]float64{}
		for k, v := range sensors {
			lk := strings.ToLower(k)
			// Match "core N" or "coreN"
			var coreNum int
			if _, err := fmt.Sscanf(lk, "core %d", &coreNum); err == nil {
				coreKeys = append(coreKeys, coreNum)
				coreMap[coreNum] = v
			} else if _, err := fmt.Sscanf(lk, "core%d", &coreNum); err == nil {
				coreKeys = append(coreKeys, coreNum)
				coreMap[coreNum] = v
			}
		}
		// Remove duplicates
		unique := map[int]struct{}{}
		filteredKeys := []int{}
		for _, n := range coreKeys {
			if _, ok := unique[n]; !ok {
				unique[n] = struct{}{}
				filteredKeys = append(filteredKeys, n)
			}
		}
		sort.Ints(filteredKeys)
		var coreRows []fyne.CanvasObject
		for _, n := range filteredKeys {
			temp := coreMap[n]
			col := colorTemp(temp, tempThreshold)
			var tempColor color.Color
			if col == "red" {
				tempColor = color.RGBA{220, 0, 0, 255}
			} else {
				tempColor = color.RGBA{0, 180, 0, 255}
			}
			label := widget.NewLabelWithStyle(fmt.Sprintf("Core %d:", n), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			value := canvas.NewText(fmt.Sprintf("%.1f°C", temp), tempColor)
			value.TextStyle = fyne.TextStyle{Bold: true}
			value.Alignment = fyne.TextAlignLeading
			row := container.NewHBox(label, value)
			coreRows = append(coreRows, row)
		}
		return container.NewVBox(
			widget.NewLabelWithStyle("CPU Temp", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true}),
			container.NewVBox(coreRows...),
		)
	}
	if title == "RAM Info" {
		// Expect sensors to contain keys: "Speed", "Banks", "BankSizes" (comma-separated sizes in GB)
		speed := "N/A"
		banks := 0
		bankSizes := ""
		for k, v := range sensors {
			lk := strings.ToLower(k)
			switch lk {
			case "speed":
				speed = fmt.Sprintf("%.0f MHz", v)
			case "banks":
				banks = int(v)
			case "banksizes":
				bankSizes = fmt.Sprintf("%.0f", v)
			}
		}
		// If bankSizes is a comma-separated string, split and show each as a separate card
		var bankCards []fyne.CanvasObject
		if bankSizes != "" {
			sizes := strings.Split(bankSizes, ",")
			for i, s := range sizes {
				card := widget.NewCard(
					fmt.Sprintf("RAM Bank #%d", i+1),
					fmt.Sprintf("Speed: %s", speed),
					widget.NewLabel(fmt.Sprintf("Size: %s GB", strings.TrimSpace(s))),
				)
				bankCards = append(bankCards, card)
			}
		}
		// Optionally add a summary card for total banks
		summaryCard := widget.NewCard(
			"RAM Info",
			fmt.Sprintf("Banks: %d", banks),
			widget.NewLabel(fmt.Sprintf("Speed: %s", speed)),
		)
		return container.NewVBox(
			summaryCard,
			container.NewVBox(bankCards...),
		)
	}
	rows := []fyne.CanvasObject{}
	for k, v := range sensors {
		lk := strings.ToLower(k)
		var labelText string
		// Map coretemp TempN to Core N-1 Temp
		if strings.Contains(lk, "coretemp") && strings.Contains(lk, "temp") {
			// Extract TempN
			idx := strings.Index(lk, "temp")
			tempNum := ""
			for i := idx + 4; i < len(lk); i++ {
				if lk[i] >= '0' && lk[i] <= '9' {
					tempNum += string(lk[i])
				} else {
					break
				}
			}
			if tempNum != "" {
				n := 0
				fmt.Sscanf(tempNum, "%d", &n)
				labelText = fmt.Sprintf("Core %d Temp", n-1)
			} else {
				labelText = k
			}
		} else {
			switch {
			case strings.Contains(lk, "cpu"):
				labelText = "CPU"
			case strings.Contains(lk, "pch"):
				labelText = "PCH"
			case strings.Contains(lk, "mobo") || strings.Contains(lk, "board"):
				labelText = "Motherboard"
			case strings.Contains(lk, "gpu"):
				if strings.Contains(lk, "amdgpu") {
					labelText = "GPU: amdgpu"
				} else if strings.Contains(lk, "nvidia") {
					labelText = "GPU: nvidia"
				} else {
					labelText = "GPU"
				}
			default:
				labelText = k
			}
		}
		col := colorTemp(v, tempThreshold)
		var tempColor color.Color
		if col == "red" {
			tempColor = color.RGBA{220, 0, 0, 255}
		} else {
			tempColor = color.RGBA{0, 180, 0, 255}
		}
		label := widget.NewLabelWithStyle(labelText, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		value := canvas.NewText(fmt.Sprintf("%.1f°C", v), tempColor)
		value.TextStyle = fyne.TextStyle{Bold: true}
		value.Alignment = fyne.TextAlignLeading
		row := container.NewHBox(label, value)
		rows = append(rows, row)
	}

	if len(sensors) == 0 && showNoData {
		rows = append(rows, widget.NewLabel("No temperature data found."))
	}
	if fans != nil {
		if len(fans) == 0 && showNoData {
			rows = append(rows, widget.NewLabel("No fan data found."))
		}
		// Sort fan keys alphabetically
		fanKeys := make([]string, 0, len(fans))
		for k := range fans {
			fanKeys = append(fanKeys, k)
		}
		sort.Strings(fanKeys)
		for _, k := range fanKeys {
			v := fans[k]
			lk := strings.ToLower(k)
			var labelText string
			switch {
			case strings.Contains(lk, "gpu"):
				if strings.Contains(lk, "amdgpu") {
					labelText = "GPU Fan (amdgpu):"
				} else if strings.Contains(lk, "nvidia") {
					labelText = "GPU Fan (nvidia):"
				} else {
					labelText = "GPU Fan:"
				}
			case strings.Contains(lk, "cpu"):
				labelText = "CPU Fan:"
			case strings.Contains(lk, "fan"):
				// Use getFanLabel for custom names
				labelText = getFanLabel(k, fanLabelMap) + ":"
			case strings.Contains(lk, "mobo") || strings.Contains(lk, "board"):
				labelText = "Motherboard Fan:"
			default:
				labelText = k + ":"
			}
			// Make RPM value bold
			value := widget.NewLabelWithStyle(fmt.Sprintf("%d rpm", v), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			row := container.NewHBox(widget.NewLabelWithStyle(labelText, fyne.TextAlignLeading, fyne.TextStyle{Bold: false}), value)
			rows = append(rows, row)
		}
	}
	return container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true}),
		container.NewVBox(rows...),
	)
}

func colorTemp(temp float64, threshold float64) string {
	if temp >= threshold {
		return "red"
	}
	return "green"
}

func getFanLabel(key string, fanLabelMap map[string]string) string {
	if val, ok := fanLabelMap[key]; ok && val != "" {
		return val
	}
	return key
}

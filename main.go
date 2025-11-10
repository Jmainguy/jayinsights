package main

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	fanLabelMap := loadConfig()

	// Gather GPU info first, before any Fyne code
	gpuInfo := GetGPUInfo()

	a := app.New()
	w := a.NewWindow("JayInsight")
	w.Resize(fyne.NewSize(800, 600))

	infoContainer := container.NewVBox()
	scroll := container.NewScroll(infoContainer)
	scroll.SetMinSize(w.Canvas().Size())

	// Use a goroutine to periodically update scroll area size to match window size
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			size := w.Canvas().Size()
			// Set scroll min size to window size, not to scroll's current min size
			fyne.Do(func() {
				scroll.Resize(size)
			})
		}
	}()

	// Set a custom background color for the main window
	bg := canvas.NewRectangle(&color.RGBA{R: 30, G: 30, B: 40, A: 255}) // dark blue-gray
	bg.Resize(fyne.NewSize(800, 600))

	w.SetContent(container.NewStack(
		bg,
		container.NewVBox(
			widget.NewSeparator(),
			scroll,
		),
	))

	refresh := func() {
		sensor := readSensors()
		moboTemps, cpuTemps, gpuTemps, _, _, gpuFans, moboFans := categorizeSensors(sensor)

		cpuModel, cpuMHz, _, _ := GetCPUInfo()
		// Count threads from /proc/cpuinfo
		threadCountDisplay := "N/A"
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			lines := strings.Split(string(data), "\n")
			count := 0
			for _, line := range lines {
				if strings.HasPrefix(line, "processor") {
					count++
				}
			}
			threadCountDisplay = fmt.Sprintf("%d", count)
		}
		drives := GetDrives()
		_, boardDMI, biosDMI, _, _ := GetDMIInfo()

		driveCards := []fyne.CanvasObject{}
		for _, d := range drives {
			model := GetDriveModel(d)
			treeRows := BuildPartitionTree(d, "")
			driveCards = append(driveCards, widget.NewCard(
				fmt.Sprintf("Drive: %s", d),
				fmt.Sprintf("Model: %s", model),
				widget.NewLabel(strings.Join(treeRows, "\n")),
			))
		}

		// Show per-core speeds only
		coreSpeeds := strings.Split(cpuMHz, ",")
		var speedRows []string
		for i, spd := range coreSpeeds {
			speedRows = append(speedRows, fmt.Sprintf("Core %d: %s MHz", i, spd))
		}
		coreCountDisplay := fmt.Sprintf("%d", len(coreSpeeds))
		// RAM Info Card
		ramBanks, ramErr := GetRAMBanks()
		var ramRows []fyne.CanvasObject
		var totalBankSize uint32
		if ramErr != nil {
			ramRows = append(ramRows, widget.NewLabel("Error reading RAM info"))
		} else if len(ramBanks) == 0 {
			ramRows = append(ramRows, widget.NewLabel("No RAM banks found"))
		} else {
			for i, bank := range ramBanks {
				ramInfoLines := []string{
					fmt.Sprintf("Bank #%d", i+1),
					fmt.Sprintf("  Locator: %s", bank.Locator),
					fmt.Sprintf("  Size: %d MB ", bank.SizeMB),
					fmt.Sprintf("  Speed: %d MHz", bank.SpeedMHz),
					fmt.Sprintf("  Type: %s", bank.MemoryType),
					fmt.Sprintf("  Manufacturer: %s", bank.Manufacturer),
				}
				totalBankSize += bank.SizeMB
				// Use canvas.Text for compact, non-padded rendering
				var bankFields []fyne.CanvasObject
				for idx, line := range ramInfoLines {
					txt := canvas.NewText(line, color.White)
					txt.TextSize = 13
					txt.TextStyle = fyne.TextStyle{Monospace: true}
					if idx == 0 {
						txt.TextStyle.Bold = true
					}
					bankFields = append(bankFields, txt)
				}
				ramRows = append(ramRows, container.NewVBox(bankFields...))
			}
		}
		totalRamLine := fmt.Sprintf("Total RAM: %d MB", totalBankSize)
		ramRows = append([]fyne.CanvasObject{widget.NewLabel(totalRamLine)}, ramRows...)
		ramCard := widget.NewCard("RAM Info", "", container.NewVBox(ramRows...))

		// Gather GPU VBIOS version (only one per GPU)
		gpuVbiosVersion := ""
		drmBase := "/sys/class/drm/"
		if cards, err := os.ReadDir(drmBase); err == nil {
			for _, card := range cards {
				if strings.HasPrefix(card.Name(), "card") && !strings.Contains(card.Name(), "-") {
					vbiosPath := drmBase + card.Name() + "/device/vbios_version"
					if vbiosBytes, err := os.ReadFile(vbiosPath); err == nil {
						vbios := strings.TrimSpace(string(vbiosBytes))
						if vbios != "" {
							gpuVbiosVersion = vbios
						}
					}
					break // only one per GPU
				}
			}
		}

		sysCards := []fyne.CanvasObject{
			widget.NewCard("CPU Info", "", container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Model: %s", cpuModel)),
				widget.NewLabel(strings.Join(speedRows, "\n")),
				widget.NewLabel(fmt.Sprintf("Cores: %s  Threads: %s", coreCountDisplay, threadCountDisplay)),
			)),
			widget.NewCard("Motherboard Info (DMI)", "", widget.NewLabel(boardDMI)),
			widget.NewCard("BIOS Info (DMI)", "", widget.NewLabel(biosDMI)),
			widget.NewCard("GPU Info", "", widget.NewLabel(
				fmt.Sprintf("%s\nVBIOS: %s", gpuInfo, gpuVbiosVersion),
			)),
		}

		caseFans := map[string]int{}
		for k, v := range sensor.FanSpeeds {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "cpu") || strings.Contains(lk, "gpu") || strings.Contains(lk, "mobo") || strings.Contains(lk, "board") {
				continue
			}

			// Normalize to canonical key like "Fan1"
			normalized := normalizeFanKey(k)

			// Use config label if available, otherwise fallback to normalized key
			label := normalized
			if custom, ok := fanLabelMap[normalized]; ok && custom != "" {
				label = custom
			}
			caseFans[label] = v
		}

		// Filter out 0.0°C motherboard temps
		filteredMoboTemps := map[string]float64{}
		for k, v := range moboTemps {
			if v != 0.0 {
				filteredMoboTemps[k] = v
			}
		}

		// Ensure GPU temp is present, including from /sys/class/drm/card*/device/hwmon/hwmon*/temp*_input
		filteredGPUTemps := map[string]float64{}
		for k, v := range gpuTemps {
			filteredGPUTemps[k] = v
		}
		if len(filteredGPUTemps) == 0 {
			// Try to find a GPU temp from all sensors
			for k, v := range sensor.Temperatures {
				lk := strings.ToLower(k)
				if strings.Contains(lk, "gpu") && v != 0.0 {
					filteredGPUTemps[k] = v
				}
			}
		}
		// Scan /sys/class/drm/card*/device/hwmon/hwmon*/temp*_input for additional GPU temps
		drmBase = "/sys/class/drm/"
		if cards, err := os.ReadDir(drmBase); err == nil {
			for _, card := range cards {
				if strings.HasPrefix(card.Name(), "card") && !strings.Contains(card.Name(), "-") {
					hwmonPath := drmBase + card.Name() + "/device/hwmon/"
					if hwmons, err := os.ReadDir(hwmonPath); err == nil {
						for _, hw := range hwmons {
							tempBase := hwmonPath + hw.Name() + "/"
							for i := 1; i <= 10; i++ {
								tPath := fmt.Sprintf("%stemp%d_input", tempBase, i)
								labelPath := fmt.Sprintf("%stemp%d_label", tempBase, i)
								var label string
								if lBytes, err := os.ReadFile(labelPath); err == nil {
									label = strings.TrimSpace(string(lBytes))
								} else {
									label = fmt.Sprintf("%s Temp%d", card.Name(), i)
								}
								if tBytes, err := os.ReadFile(tPath); err == nil {
									tVal, err := strconv.ParseFloat(strings.TrimSpace(string(tBytes)), 64)
									if err == nil {
										tempC := tVal / 1000.0
										if tempC != 0.0 {
											filteredGPUTemps[label] = tempC
										}
									}
								}
							}
						}
					}
				}
			}
		}

		cardColor := &color.RGBA{R: 60, G: 60, B: 80, A: 255} // lighter blue-gray for cards

		// Wrap each card in a colored background
		wrapCard := func(card fyne.CanvasObject) fyne.CanvasObject {
			rect := canvas.NewRectangle(cardColor)
			rect.SetMinSize(card.MinSize())
			return container.NewStack(rect, card)
		}

		driveCardsWrapped := []fyne.CanvasObject{}
		for _, c := range driveCards {
			driveCardsWrapped = append(driveCardsWrapped, wrapCard(c))
		}

		sysCardsWrapped := []fyne.CanvasObject{}
		for _, c := range sysCards {
			sysCardsWrapped = append(sysCardsWrapped, wrapCard(c))
		}

		ramCardWrapped := wrapCard(ramCard)

		sysInfoCard := container.NewGridWithColumns(3,
			container.NewVBox(sysCardsWrapped...),
			container.NewVBox(
				container.NewVBox(driveCardsWrapped...),
				ramCardWrapped,
			),
			container.NewVBox(
				wrapCard(MakeSection("Motherboard Temp", filteredMoboTemps, moboFans, 60, false, fanLabelMap)),
				wrapCard(MakeSection("CPU Temp", cpuTemps, nil, 80, false, fanLabelMap)),
				wrapCard(MakeSection("Fans", nil, caseFans, 0, false, fanLabelMap)),
				wrapCard(MakeSection("GPU Temp & Fan", filteredGPUTemps, gpuFans, 80, false, fanLabelMap)),
			),
		)

		// Instead of replacing infoContainer.Objects, update its content in place
		if len(infoContainer.Objects) == 0 {
			infoContainer.Objects = []fyne.CanvasObject{sysInfoCard}
		} else {
			infoContainer.Objects[0] = sysInfoCard
		}
		infoContainer.Refresh()
	}

	refresh()

	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			fyne.Do(func() {
				refresh()
			})
		}
	}()
	w.ShowAndRun()
}

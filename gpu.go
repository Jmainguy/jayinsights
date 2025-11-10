package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/jaypipes/pcidb"
)

// Get OpenGL GPU Model (Renderer) string
func GetOpenGLModel() (string, string) {
	if err := glfw.Init(); err != nil {
		return "N/A", "N/A"
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.Visible, glfw.False)
	window, err := glfw.CreateWindow(100, 100, "Hidden", nil, nil)
	if err != nil {
		return "N/A", "N/A"
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		return "N/A", "N/A"
	}
	renderer := gl.GoStr(gl.GetString(gl.RENDERER))
	vendor := gl.GoStr(gl.GetString(gl.VENDOR))
	return renderer, vendor
}

func GetGPUInfo() (info string) {
	vendor := "N/A"
	subsystemVendor := "N/A"
	subsystemVendorName := "N/A"
	// subsystemDevice removed (was unused)
	vram := "N/A"
	drmPath := "/sys/class/drm/"
	entries, _ := os.ReadDir(drmPath)
	// displays removed
	pci, err := pcidb.New()
	if err != nil {
		// fallback: no PCI DB
		pci = nil
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "card") && !strings.Contains(entry.Name(), "-") {
			subsystemVendorPath := drmPath + entry.Name() + "/device/subsystem_vendor"
			subsystemVendorBytes, _ := os.ReadFile(subsystemVendorPath)
			if len(subsystemVendorBytes) > 0 {
				subsystemVendor = strings.TrimPrefix(strings.TrimSpace(string(subsystemVendorBytes)), "0x")
				if pci != nil {
					if v, ok := pci.Vendors[subsystemVendor]; ok {
						subsystemVendorName = v.Name
					}
				}
			}
			// modes and displays logic removed
			vramPath := drmPath + entry.Name() + "/device/mem_info_vram_total"
			vramBytes, err := os.ReadFile(vramPath)
			if err == nil && len(vramBytes) > 0 {
				vramInt, err := strconv.ParseInt(strings.TrimSpace(string(vramBytes)), 10, 64)
				if err == nil {
					vram = fmt.Sprintf("%.2f GB", float64(vramInt)/1024.0/1024.0/1024.0)
				}
			}
		}
	}

	model, vendor := GetOpenGLModel()
	// Remove everything inside parentheses from model string
	if idx := strings.Index(model, "("); idx != -1 {
		model = strings.TrimSpace(model[:idx])
	}
	info = fmt.Sprintf("Vendor: %s \nSubsystem Vendor: %s\nModel: %s\nVRAM: %s", vendor, subsystemVendorName, model, vram)
	return info
}

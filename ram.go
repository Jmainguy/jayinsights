package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
)

type RAMBank struct {
	Locator      string
	BankLocator  string
	SizeMB       uint32
	SpeedMHz     uint16
	MemoryType   string
	Manufacturer string
}

// Map some known manufacturer IDs to friendly names (add more as needed)
var manufacturerMap = map[string]string{
	"029E": "Corsair",
	"80CE": "Kingston",
	"04CD": "Samsung",
	"049F": "Micron",
	"02C0": "Crucial",
	"0417": "ADATA",
}

var memoryTypeMap = map[byte]string{
	0x01: "Other",
	0x02: "DRAM",
	0x03: "Synchronous DRAM",
	0x04: "Cache DRAM",
	0x05: "EDO",
	0x06: "EDRAM",
	0x07: "VRAM",
	0x08: "SRAM",
	0x09: "RAM",
	0x0A: "ROM",
	0x0B: "Flash",
	0x0C: "EEPROM",
	0x0D: "FEPROM",
	0x0E: "EPROM",
	0x0F: "CDRAM",
	0x10: "3DRAM",
	0x11: "SDRAM",
	0x12: "SGRAM",
	0x13: "RDRAM",
	0x14: "DDR",
	0x15: "DDR2",
	0x18: "DDR3",
	0x1A: "DDR4",
	0x1C: "DDR5",
}

func GetRAMBanks() ([]RAMBank, error) {
	rawFiles, err := filepath.Glob("/sys/firmware/dmi/entries/17-*/raw")
	if err != nil {
		return nil, err
	}

	var banks []RAMBank

	for _, file := range rawFiles {
		raw, err := os.ReadFile(file)
		if err != nil || len(raw) < 0x1C {
			continue
		}

		structLen := int(raw[1])
		if len(raw) < structLen {
			continue
		}

		strSection := raw[structLen:]
		strs := parseDMIStrings(strSection)

		getIndex := func(offset int) int {
			if offset < structLen && offset < len(raw) {
				return int(raw[offset])
			}
			return 0
		}

		size := binary.LittleEndian.Uint16(raw[0x0C:0x0E])
		if size == 0 || size == 0xFFFF {
			continue
		}

		speed := binary.LittleEndian.Uint16(raw[0x15:0x17])
		memType := memoryTypeMap[raw[0x12]]

		rawManufacturer := strings.TrimSpace(safeString(strs, getIndex(0x17)))
		prettyManufacturer := mapManufacturer(rawManufacturer)

		bank := RAMBank{
			SizeMB:       uint32(size),
			SpeedMHz:     speed,
			MemoryType:   memType,
			Locator:      strings.TrimSpace(safeString(strs, getIndex(0x10))),
			BankLocator:  strings.TrimSpace(safeString(strs, getIndex(0x11))),
			Manufacturer: prettyManufacturer,
		}

		banks = append(banks, bank)
	}

	return banks, nil
}

func mapManufacturer(code string) string {
	if name, ok := manufacturerMap[code]; ok {
		return name
	}
	if code != "" {
		return code
	}
	return "Unknown"
}

func safeString(strings []string, index int) string {
	if index <= 0 || index > len(strings) {
		return ""
	}
	return strings[index-1]
}

func parseDMIStrings(data []byte) []string {
	var out []string
	start := 0
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0 && data[i+1] == 0 {
			break
		}
		if data[i] == 0 {
			out = append(out, string(data[start:i]))
			start = i + 1
		}
	}
	return out
}

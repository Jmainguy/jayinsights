package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func BuildPartitionTree(dev string, indent string) []string {
	sizePath := fmt.Sprintf("/sys/block/%s/size", dev)
	sizeGB := ""
	sizeRaw, err := os.ReadFile(sizePath)
	if err == nil {
		sectors, err := strconv.ParseInt(strings.TrimSpace(string(sizeRaw)), 10, 64)
		if err == nil {
			gb := float64(sectors*512) / 1024.0 / 1024.0 / 1024.0
			sizeGB = fmt.Sprintf("%.2f GB", gb)
		}
	}
	rows := []string{fmt.Sprintf("%s%s (%s)", indent, dev, sizeGB)}
	parts := GetPartitions(dev)
	for _, p := range parts {
		partSize, partMount := GetPartitionInfo(p)
		partRow := fmt.Sprintf("%s├─%s (%s) %s", indent, p, partSize, partMount)
		rows = append(rows, partRow)
		mapperDir := "/dev/mapper/"
		if files, err := os.ReadDir(mapperDir); err == nil {
			for _, f := range files {
				linkPath := mapperDir + f.Name()
				linkTarget, err := os.Readlink(linkPath)
				if err == nil && (strings.Contains(linkTarget, p) || strings.Contains(linkTarget, dev)) {
					luksMount := ""
					mounts, err := os.ReadFile("/proc/mounts")
					if err == nil {
						for _, line := range strings.Split(string(mounts), "\n") {
							fields := strings.Fields(line)
							if len(fields) >= 2 && strings.Contains(fields[0], f.Name()) {
								luksMount = fields[1]
								break
							}
						}
					}
					cryptRow := fmt.Sprintf("%s└─%s (LUKS) %s", indent+"  ", f.Name(), luksMount)
					rows = append(rows, cryptRow)
					break
				}
			}
		}
	}
	return rows
}

func GetPartitions(dev string) []string {
	parts := []string{}
	files, _ := os.ReadDir(fmt.Sprintf("/sys/block/%s/", dev))
	for _, f := range files {
		if strings.HasPrefix(f.Name(), dev) {
			parts = append(parts, f.Name())
		}
	}
	return parts
}

func GetPartitionInfo(part string) (sizeGB, mount string) {
	sizePath := ""
	if strings.HasPrefix(part, "nvme") {
		sizePath = fmt.Sprintf("/sys/block/%s/%s/size", part[:6], part)
	} else {
		sizePath = fmt.Sprintf("/sys/block/%s/%s/size", part[:3], part)
	}
	sizeBytes := ""
	sizeRaw, err := os.ReadFile(sizePath)
	if err == nil {
		sectors, err := strconv.ParseInt(strings.TrimSpace(string(sizeRaw)), 10, 64)
		if err == nil {
			gb := float64(sectors*512) / 1024.0 / 1024.0 / 1024.0
			sizeBytes = fmt.Sprintf("%.2f GB", gb)
		}
	}
	mount = ""
	mounts, err := os.ReadFile("/proc/mounts")
	if err == nil {
		for _, line := range strings.Split(string(mounts), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && strings.Contains(fields[0], part) {
				mount = fields[1]
				break
			}
		}
	}
	return sizeBytes, mount
}

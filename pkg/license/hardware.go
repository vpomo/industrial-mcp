package license

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type HardwareInfo struct {
	CPUID       string
	MACAddr     string
	VolumeID    string
	Motherboard string
}

func GetHardwareInfo() (HardwareInfo, error) {
	info := HardwareInfo{
		CPUID:       getCPUID(),
		MACAddr:     getMACAddress(),
		VolumeID:    getVolumeID(),
		Motherboard: getMotherboardSerial(),
	}
	return info, nil
}

func (h HardwareInfo) Hash() string {
	data := h.CPUID + h.MACAddr + h.VolumeID + h.Motherboard
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func getCPUID() string {
	switch runtime.GOOS {
	case "windows":
		return runPowerShell(`(Get-WmiObject Win32_Processor | Select-Object -First 1).ProcessorId`)
	case "linux":
		cpuInfo := readFile("/proc/cpuinfo")
		for _, line := range strings.Split(cpuInfo, "\n") {
			if strings.HasPrefix(line, "model identifier") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
		return getLinuxMachineID()
	}
	return generateFallbackID("cpu")
}

func getMACAddress() string {
	switch runtime.GOOS {
	case "windows":
		return runPowerShell(`(Get-NetAdapter -Physical | Where-Object { $_.Status -eq "Up" } | Select-Object -First 1).MacAddress`)
	case "linux":
		output, _ := exec.Command("ip", "link", "show").Output()
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "ether") && !strings.Contains(line, "00:00:00:00:00:00") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return parts[1]
				}
			}
		}
		return getLinuxMachineID()
	}
	return generateFallbackID("mac")
}

func getVolumeID() string {
	switch runtime.GOOS {
	case "windows":
		return runPowerShell(`(Get-WmiObject Win32_Volume -Filter "DriveLetter='C:'" | Select-Object -First 1).SerialNumber`)
	case "linux":
		output, _ := exec.Command("blkid", "-s", "UUID", "-o", "value", "/dev/sda1").Output()
		return strings.TrimSpace(string(output))
	}
	return generateFallbackID("vol")
}

func getMotherboardSerial() string {
	switch runtime.GOOS {
	case "windows":
		return runPowerShell(`(Get-WmiObject Win32_BaseBoard | Select-Object -First 1).SerialNumber`)
	case "linux":
		output, _ := exec.Command("dmidecode", "-s", "baseboard-serial-number").Output()
		result := strings.TrimSpace(string(output))
		if result == "" || strings.Contains(result, "Not") {
			return getLinuxMachineID()
		}
		return result
	}
	return generateFallbackID("mb")
}

func getLinuxMachineID() string {
	id := readFile("/etc/machine-id")
	id = strings.TrimSpace(id)
	if id != "" {
		return id
	}
	return generateFallbackID("mid")
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func runPowerShell(script string) string {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func generateFallbackID(prefix string) string {
	hostname, _ := os.Hostname()
	data := prefix + hostname + runtime.GOOS
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

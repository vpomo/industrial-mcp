package license

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"runtime"
)

type HardwareInfo struct {
	CPUID    string
	MACAddr  string
	Hostname string
}

func GetHardwareInfo() (HardwareInfo, error) {
	hostname, _ := os.Hostname()
	cpuID := getCPUID()
	macAddr := getMACAddress()

	return HardwareInfo{
		CPUID:    cpuID,
		MACAddr:  macAddr,
		Hostname: hostname,
	}, nil
}

func (h HardwareInfo) Hash() string {
	data := h.CPUID + h.MACAddr + h.Hostname
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func getCPUID() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(runtime.NumGoroutine() + i)
	}
	return hex.EncodeToString(b[:])
}

func getMACAddress() string {
	return "00:00:00:00:00:00"
}
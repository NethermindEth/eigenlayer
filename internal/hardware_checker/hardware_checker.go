package hardwarechecker

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
)

// HardwareMetrics represents hardware metrics such as CPU, RAM, and disk space.
type HardwareMetrics struct {
	CPU       float64 `json:"cpu"`        // Cores
	RAM       float64 `json:"ram"`        // Mb
	DiskSpace float64 `json:"disk_space"` // Mb
}

// Meets checks if the current HardwareMetrics instance meets the specified hardware metrics.
func (h *HardwareMetrics) Meets(hm HardwareMetrics) bool {
	return h.CPU >= hm.CPU && h.RAM >= hm.RAM && h.DiskSpace >= hm.DiskSpace
}

func (h *HardwareMetrics) String() string {
	return fmt.Sprintf("CPU: %.2f Cores, RAM: %.2f Mb, Disk Space: %.2f Mb", h.CPU, h.RAM, h.DiskSpace)
}

// GetHardwareMetrics retrieves hardware metrics from a Prometheus server using the provided address.
// func GetHardwareMetrics(address string) (hardwareMetrics HardwareMetrics, err error) {
// 	cpuQuery := "count(count(node_cpu_seconds_total) by (cpu))"
// 	ramQuery := "node_memory_MemTotal_bytes/1024/1024"                        // Mb
// 	diskSpaceQuery := "node_filesystem_avail_bytes{mountpoint='/'}/1024/1024" // Mb

// 	hardwareMetrics.CPU, err = QueryNodeExporter(address, cpuQuery)
// 	if err != nil {
// 		return HardwareMetrics{}, err
// 	}

// 	hardwareMetrics.RAM, err = QueryNodeExporter(address, ramQuery)
// 	if err != nil {
// 		return HardwareMetrics{}, err
// 	}

// 	hardwareMetrics.DiskSpace, err = QueryNodeExporter(address, diskSpaceQuery)
// 	if err != nil {
// 		return HardwareMetrics{}, err
// 	}

// 	return hardwareMetrics, nil
// }

// GetHardwareMetrics retrieves hardware metrics from a Linux host
func GetMetrics() (hardwareMetrics HardwareMetrics, err error) {
	// CPU Cores
	cpuCores := runtime.NumCPU()
	hardwareMetrics.CPU = float64(cpuCores)

	// Total Memory RAM
	memInfo := &syscall.Sysinfo_t{}
	err = syscall.Sysinfo(memInfo)
	if err != nil {
		return hardwareMetrics, fmt.Errorf("failed to get memory info: %w", err)
	}
	totalMemory := float64(memInfo.Totalram*uint64(memInfo.Unit)) / (1024 * 1024) // Convert to Mb
	hardwareMetrics.RAM = totalMemory

	// Disk Free Space
	wd, err := os.Getwd()
	if err != nil {
		return hardwareMetrics, fmt.Errorf("failed to get current working directory: %w", err)
	}
	var stat syscall.Statfs_t
	err = syscall.Statfs(wd, &stat)
	if err != nil {
		return hardwareMetrics, fmt.Errorf("failed to get disk free space: %w", err)
	}
	freeSpace := float64(stat.Bavail*uint64(stat.Bsize)) / (1024 * 1024) // Convert to Mb
	hardwareMetrics.DiskSpace = freeSpace

	return hardwareMetrics, nil
}

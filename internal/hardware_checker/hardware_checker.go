package hardwarechecker

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// HardwareMetrics represents hardware metrics such as CPU, RAM, and disk space.
type HardwareMetrics struct {
	CPU       float64 `json:"cpu"`        // Cores
	RAM       float64 `json:"ram"`        // Mb
	DiskSpace float64 `json:"disk_space"` // Mb
}

// Meets checks if the current HardwareMetrics instance meets the specified hardware metrics.
func (h *HardwareMetrics) Meets(hm HardwareMetrics) bool {
	return h.CPU <= hm.CPU && h.RAM <= hm.RAM && h.DiskSpace <= hm.DiskSpace
}

// GetHardwareMetrics retrieves hardware metrics from a Prometheus server using the provided address.
func GetHardwareMetrics(address string) (hardwareMetrics HardwareMetrics, err error) {
	cpuQuery := "count(count(node_cpu_seconds_total) by (cpu))"
	ramQuery := "node_memory_MemTotal_bytes/1024/1024"                        // Mb
	diskSpaceQuery := "node_filesystem_avail_bytes{mountpoint='/'}/1024/1024" // Mb

	hardwareMetrics.CPU, err = QueryNodeExporter(address, cpuQuery)
	if err != nil {
		return HardwareMetrics{}, err
	}

	hardwareMetrics.RAM, err = QueryNodeExporter(address, ramQuery)
	if err != nil {
		return HardwareMetrics{}, err
	}

	hardwareMetrics.DiskSpace, err = QueryNodeExporter(address, diskSpaceQuery)
	if err != nil {
		return HardwareMetrics{}, err
	}

	return hardwareMetrics, nil
}

// QueryNodeExporter queries the Prometheus server at the specified address with the given query.
func QueryNodeExporter(address, query string) (float64, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return 0, fmt.Errorf("error creating client: %v", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, _, err := v1api.Query(ctx, query, time.Now(), v1.WithTimeout(5*time.Second))
	if err != nil {
		return 0, fmt.Errorf("error querying Prometheus: %v", err)
	}

	vectorResult, ok := result.(model.Vector)
	if !ok || len(vectorResult) == 0 {
		return 0, fmt.Errorf("no data found for query: %s", query)
	}

	// Return the first value
	return float64(vectorResult[0].Value), nil
}

package system

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bumbacea/go-mktxp/collector"
	"github.com/bumbacea/go-mktxp/config"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	collector.RegisterAvailableCollector(&ResourcesCollector{})
}

type ResourcesCollector struct {
	freeMemory    *prometheus.GaugeVec
	totalMemory   *prometheus.GaugeVec
	freeHddSpace  *prometheus.GaugeVec
	totalHddSpace *prometheus.GaugeVec
	uptime        *prometheus.GaugeVec
	cpuLoad       *prometheus.GaugeVec
	cpuCount      *prometheus.GaugeVec
	cpuFrequency  *prometheus.GaugeVec
}

// Collect retrieves system resource information and sets Prometheus metrics.
func (r *ResourcesCollector) Collect(ctx context.Context, router *collector.RouterEntry) error {
	// Run the "/system/resource/print" command to fetch system details
	rply, err := router.Conn.RunContext(ctx, "/system/resource/print", "proplist=uptime,free-memory,total-memory,free-hdd-space,total-hdd-space,cpu-load,cpu-count,cpu-frequency,architecture-name,board-name,cpu,version")
	if err != nil {
		return fmt.Errorf("failed to run /system/resource/print command: %w", err)
	}

	for _, sentence := range rply.Re {
		// Extract relevant fields
		labels := prometheus.Labels{
			"architecture_name": sentence.Map["architecture-name"],
			"board_name":        sentence.Map["board-name"],
			"cpu":               sentence.Map["cpu"],
			"version":           sentence.Map["version"],
		}

		// Helper function to set metric values
		setMetricValue := func(metric *prometheus.GaugeVec, key string, sentenceMap map[string]string) {
			value, err := strconv.ParseFloat(sentenceMap[key], 64)
			if err == nil {
				metric.With(labels).Set(value)
			}
		}

		// Set all the metrics
		setMetricValue(r.freeMemory, "free-memory", sentence.Map)
		setMetricValue(r.totalMemory, "total-memory", sentence.Map)
		setMetricValue(r.freeHddSpace, "free-hdd-space", sentence.Map)
		setMetricValue(r.totalHddSpace, "total-hdd-space", sentence.Map)
		setMetricValue(r.cpuLoad, "cpu-load", sentence.Map)
		setMetricValue(r.cpuCount, "cpu-count", sentence.Map)
		setMetricValue(r.cpuFrequency, "cpu-frequency", sentence.Map)

		// Parse uptime from string format to seconds (if needed support uptime later parsing)
		uptimeSeconds, err := parseDurationToSeconds(sentence.Map["uptime"])
		if err == nil {
			r.uptime.With(labels).Set(uptimeSeconds)
		}
	}

	return nil
}

// IsEnabled determines if this collector is enabled for the current router.
func (r *ResourcesCollector) IsEnabled(entry config.RouterConfig) bool {
	return true // Always enabled
}

// Declare initializes the Prometheus gauges and registers them with Prometheus.
func (r *ResourcesCollector) Declare(registry prometheus.Registerer, address string, routerName string) error {
	commonLabels := []string{"architecture_name", "board_name", "cpu", "version"}

	r.freeMemory = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "system_free_memory",
			Help:        "Free memory available on the router (in bytes).",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	r.totalMemory = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "system_total_memory",
			Help:        "Total memory on the router (in bytes).",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	r.freeHddSpace = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "system_free_hdd_space",
			Help:        "Free HDD space available on the router (in bytes).",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	r.totalHddSpace = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "system_total_hdd_space",
			Help:        "Total HDD space on the router (in bytes).",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	r.uptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "system_uptime",
			Help:        "System uptime in seconds.",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	r.cpuLoad = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "system_cpu_load",
			Help:        "CPU load on the router (percentage).",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	r.cpuCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "system_cpu_count",
			Help:        "Number of available CPU cores.",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	r.cpuFrequency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "system_cpu_frequency",
			Help:        "CPU frequency in MHz.",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)

	// Register all metrics with Prometheus
	for _, metric := range []*prometheus.GaugeVec{r.freeMemory, r.totalMemory, r.freeHddSpace, r.totalHddSpace, r.uptime, r.cpuLoad, r.cpuCount, r.cpuFrequency} {
		if err := registry.Register(metric); err != nil {
			return fmt.Errorf("failed to register metric: %w", err)
		}
	}

	return nil
}

// parseDurationToSeconds converts a RouterOS duration string (e.g., "1d2h3m4s") to seconds.
func parseDurationToSeconds(duration string) (float64, error) {
	// Add logic for parsing Mikrotik uptime format into seconds
	return strconv.ParseFloat(duration, 64) // Simple placeholder
}

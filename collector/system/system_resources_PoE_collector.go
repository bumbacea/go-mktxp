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
	collector.RegisterAvailableCollector(&POECollector{})
}

type POECollector struct {
	poeVoltage *prometheus.GaugeVec
	poeCurrent *prometheus.GaugeVec
	poePower   *prometheus.GaugeVec
	poeInfo    *prometheus.GaugeVec
}

// Collect retrieves PoE metrics from the router and sets Prometheus metrics.
func (p *POECollector) Collect(ctx context.Context, router *collector.RouterEntry) error {
	// Run the "/interface/ethernet/print" command to fetch PoE details
	rply, err := router.Conn.RunContext(ctx, "/interface/ethernet/poe/print", "proplist=name")
	if err != nil {
		return fmt.Errorf("failed to run /interface/ethernet/poe/print command: %w", err)
	}

	for idx := range rply.Re {
		monitorReply, err := router.Conn.RunContext(ctx, "/interface/ethernet/poe/monitor", "=once=", fmt.Sprintf("=numbers=%d", idx))
		if err != nil {
			return fmt.Errorf("failed to run /interface/ethernet/poe/monitor command: %w", err)
		}

		for _, sentence := range monitorReply.Re {
			// Extract fields for labels and metrics
			name := sentence.Map["name"]
			poeOut := sentence.Map["poe-out"]
			poePriority := sentence.Map["poe-priority"]
			poeOutStatus := sentence.Map["poe-out-status"]

			// Use these labels for all metrics
			labels := prometheus.Labels{
				"name":           name,
				"poe_out":        poeOut,
				"poe_priority":   poePriority,
				"poe_out_status": poeOutStatus,
			}

			// Function to safely convert values and set metrics
			setMetricValue := func(metric *prometheus.GaugeVec, key string, sentenceMap map[string]string) {
				value, err := strconv.ParseFloat(sentenceMap[key], 64)
				if err == nil {
					metric.With(labels).Set(value)
				}
			}

			// Collect numeric metrics (voltage, current, power)
			setMetricValue(p.poeVoltage, "poe-out-voltage", sentence.Map)
			setMetricValue(p.poeCurrent, "poe-out-current", sentence.Map)
			setMetricValue(p.poePower, "poe-out-power", sentence.Map)

			// Set the PoE info metric
			p.poeInfo.With(labels).Set(1) // Info metric is a "presence" metric, always set to 1
		}

	}

	return nil
}

// IsEnabled determines if this collector is enabled for the current router.
func (p *POECollector) IsEnabled(entry config.RouterConfig) bool {
	// Check if PoE is enabled in the router's config; return false if not
	if entry.POE == nil {
		return false
	}
	return *entry.POE
}

// Declare initializes the Prometheus gauges and registers them.
func (p *POECollector) Declare(registry prometheus.Registerer, address string, routerName string) error {
	commonLabels := []string{"name", "poe_out", "poe_priority", "poe_out_status"}

	p.poeVoltage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "poe_out_voltage",
			Help:        "Output voltage of PoE interfaces (in Volts).",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	p.poeCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "poe_out_current",
			Help:        "Output current of PoE interfaces (in Amperes).",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	p.poePower = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "poe_out_power",
			Help:        "Output power of PoE interfaces (in Watts).",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)
	p.poeInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "mktxp",
			Name:        "poe_info",
			Help:        "Information about PoE interfaces.",
			ConstLabels: prometheus.Labels{"routerboard_address": address, "routerboard_name": routerName},
		},
		commonLabels,
	)

	// Register all metrics
	for _, metric := range []*prometheus.GaugeVec{p.poeVoltage, p.poeCurrent, p.poePower, p.poeInfo} {
		if err := registry.Register(metric); err != nil {
			return fmt.Errorf("failed to register metric: %w", err)
		}
	}

	return nil
}

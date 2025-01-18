package system

import (
	"context"
	"fmt"

	"github.com/bumbacea/go-mktxp/collector"
	"github.com/bumbacea/go-mktxp/config"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	collector.RegisterAvailableCollector(&IdentityCollector{})
}

type IdentityCollector struct {
	gauge *prometheus.GaugeVec
}

func (c *IdentityCollector) Collect(ctx context.Context, router *collector.RouterEntry) error {
	// Run the RouterOS command to get the system identity
	reply, err := router.Conn.RunContext(ctx, "/system/identity/print")
	if err != nil {
		return fmt.Errorf("failed to run /system/identity/print command: %w", err)
	}

	// Process the response
	for _, sentence := range reply.Re {
		identityName := sentence.Map["name"]
		if identityName == "" {
			return fmt.Errorf("missing 'name' field in /system/identity/print response")
		}
		// Set the prometheus gauge with the name label
		c.gauge.WithLabelValues(identityName).Set(1)
	}

	return nil
}

func (c *IdentityCollector) IsEnabled(_ config.RouterConfig) bool {
	// Always enabled
	return true
}

func (c *IdentityCollector) Declare(registry prometheus.Registerer, address string, routerName string) error {
	// Define the Prometheus gauge metric
	c.gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "mktxp",
			Name:      "system_identity_info",
			Help:      "Information about the system identity of the router",
			ConstLabels: map[string]string{
				"routerboard_address": address,
				"routerboard_name":    routerName,
			},
		},
		[]string{"name"},
	)

	// Register the gauge metric
	if err := registry.Register(c.gauge); err != nil {
		return err
	}

	return nil
}

package system

import (
	"context"
	"fmt"

	"github.com/bumbacea/go-mktxp/collector"
	"github.com/bumbacea/go-mktxp/config"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	collector.RegisterAvailableCollector(&PackagesCollector{})
}

type PackagesCollector struct {
	gauge *prometheus.GaugeVec
}

func (p *PackagesCollector) Collect(ctx context.Context, router *collector.RouterEntry) error {
	rply, err := router.Conn.RunContext(ctx, "/system/package/print", "proplist=name,version,build-time,disabled")
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	for _, sentence := range rply.Re {
		p.gauge.WithLabelValues(sentence.Map["name"], sentence.Map["version"], sentence.Map["build-time"], sentence.Map["disabled"]).Set(1)
	}
	return nil
}

func (p *PackagesCollector) IsEnabled(entry config.RouterConfig) bool {
	if entry.InstalledPackages == nil {
		return true
	}
	if *entry.InstalledPackages {
		return true
	}
	return false
}

func (p *PackagesCollector) Declare(registry prometheus.Registerer, address string, routerName string) error {
	p.gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "mktxp",
			Name:      "installed_packages_info",
			Help:      "Information about installed packages on the router",
			ConstLabels: map[string]string{
				"routerboard_address": address,
				"routerboard_name":    routerName,
			},
		},
		[]string{"name", "version", "build_time", "disabled"},
	)

	if err := registry.Register(p.gauge); err != nil {
		return err
	}

	return nil
}

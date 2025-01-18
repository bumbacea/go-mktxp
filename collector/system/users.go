package system

import (
	"context"
	"fmt"

	"github.com/bumbacea/go-mktxp/collector"
	"github.com/bumbacea/go-mktxp/config"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	collector.RegisterAvailableCollector(&ActiveUsersCollector{})
}

type ActiveUsersCollector struct {
	gauge *prometheus.GaugeVec
}

func (a *ActiveUsersCollector) Collect(ctx context.Context, router *collector.RouterEntry) error {
	rply, err := router.Conn.RunContext(ctx, "/user/active/print", "proplist=name,when,address,via,group")
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	for _, sentence := range rply.Re {
		name := sentence.Map["name"]
		when := sentence.Map["when"]
		address := sentence.Map["address"]
		via := sentence.Map["via"]
		group := sentence.Map["group"]

		a.gauge.WithLabelValues(name, when, address, via, group, router.ConfigEntry.Hostname, router.ConfigEntry.Name).Set(1)
	}
	return nil
}

func (a *ActiveUsersCollector) IsEnabled(entry config.RouterConfig) bool {
	if entry.User == nil {
		return true
	}
	if *entry.User {
		return true
	}
	return false
}

func (a *ActiveUsersCollector) Declare(registry prometheus.Registerer) error {
	a.gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "mktxp",
			Name:      "active_users_info",
			Help:      "Information about active users on the router",
		},
		[]string{"name", "when", "address", "via", "group", "routerboard_address", "routerboard_name"},
	)

	if err := registry.Register(a.gauge); err != nil {
		return err
	}

	return nil
}

package collector

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bumbacea/go-mktxp/config"
	"github.com/go-routeros/routeros/v3"
	"github.com/prometheus/client_golang/prometheus"
)

var availableConnectors = make([]Collector, 0)

func RegisterAvailableCollector(collector Collector) {
	availableConnectors = append(availableConnectors, collector)
}

type Collector interface {
	Collect(router *RouterEntry) error
	IsEnabled(entry config.RouterConfig) bool
	Declare(registry prometheus.Registerer, address string, routerName string) error
}

// RouterEntry holds configuration entry details.
type RouterEntry struct {
	ConfigEntry config.RouterConfig
	Conn        *routeros.Client
	Collectors  []Collector
}

func NewRouterEntry(cfg config.RouterConfig) (*RouterEntry, error) {
	dial, err := routeros.DialTLSTimeout(
		fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port),
		cfg.Username,
		cfg.Password,
		&tls.Config{
			InsecureSkipVerify: !(cfg.SSLCertificateVerify == nil || *cfg.SSLCertificateVerify),
		},
		time.Minute,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	dial.SetLogHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))
	return &RouterEntry{
		ConfigEntry: cfg,
		Conn:        dial,
		Collectors:  availableConnectors,
	}, nil
}

func (e *RouterEntry) Declare(registry prometheus.Registerer) error {
	for _, collector := range e.Collectors {
		if !collector.IsEnabled(e.ConfigEntry) {
			continue
		}
		err := collector.Declare(registry, e.ConfigEntry.Hostname, e.ConfigEntry.Name)
		if err != nil {
			return fmt.Errorf("failed to declare collector: %w", err)
		}
	}
	return nil
}
func (e *RouterEntry) Collect() error {
	for _, collector := range e.Collectors {
		if !collector.IsEnabled(e.ConfigEntry) {
			continue
		}
		err := collector.Collect(e)
		if err != nil {
			return fmt.Errorf("failed to collect metrics for %s of type %T: %w", e.ConfigEntry.Hostname, collector, err)
		}
	}
	return nil
}

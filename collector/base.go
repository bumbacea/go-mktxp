package collector

import (
	"context"
	"crypto/tls"
	"fmt"
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
	Collect(ctx context.Context, router *RouterEntry) error
	IsEnabled(entry config.RouterConfig) bool
	Declare(registry prometheus.Registerer) error
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
	//dial.SetLogHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	//	AddSource:   true,
	//	Level:       slog.LevelDebug,
	//	ReplaceAttr: nil,
	//}))
	return &RouterEntry{
		ConfigEntry: cfg,
		Conn:        dial,
		Collectors:  availableConnectors,
	}, nil
}

func (e *RouterEntry) Collect(ctx context.Context) error {
	for _, collector := range e.Collectors {
		if !collector.IsEnabled(e.ConfigEntry) {
			continue
		}
		err := collector.Collect(ctx, e)
		if err != nil {
			return fmt.Errorf("failed to collect metrics for %s of type %T: %w", e.ConfigEntry.Hostname, collector, err)
		}
	}
	return nil
}

func DeclareAll(registry *prometheus.Registry) error {
	for _, connector := range availableConnectors {
		err := connector.Declare(registry)
		if err != nil {
			return fmt.Errorf("failed to declare collector %T: %w", connector, err)
		}
	}
	return nil
}

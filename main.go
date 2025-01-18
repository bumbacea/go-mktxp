package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/bumbacea/go-mktxp/collector"
	_ "github.com/bumbacea/go-mktxp/collector/system"
	"github.com/bumbacea/go-mktxp/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const defaultKey = "default"

func main() {
	var cfgdir string

	rootCmd := &cobra.Command{
		Use:   "metrics-server",
		Short: "Starts a metrics server",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := initializeRegistry()
			globalConfig, err := config.LoadConfig(path.Join(cfgdir, "_mktxp.conf"))
			if err != nil {
				return fmt.Errorf("error parsing config file: %w", err)
			}
			instances, err := config.ParseConfig(path.Join(cfgdir, "mktxp.conf"))
			if err != nil {
				return fmt.Errorf("error parsing config file: %w", err)
			}

			defaultInstance, ok := instances[defaultKey]
			if ok {
				delete(instances, defaultKey)
			}

			for instance := range instances {
				mergedConfig := instances[instance]
				if ok {
					mergedConfig = config.MergeDefaults(defaultInstance, instances[instance])
				}

				if mergedConfig.Enabled != nil && !*mergedConfig.Enabled {
					continue
				}

				mergedConfig.Name = instance

				log.Printf("Going to start collector for router %s", instance)

				err := startCollector(registry, mergedConfig)
				if err != nil {
					return fmt.Errorf("failed to start collector for router %s: %w", instance, err)
				}
			}
			startHTTPServer(globalConfig.Listen, registry)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&cfgdir, "cfg-dir", "", "./", "MKTXP config files directory (optional)")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func startCollector(registry prometheus.Registerer, conf config.RouterConfig) error {
	router, err := collector.NewRouterEntry(conf)
	if err != nil {
		return fmt.Errorf("failed to create router entry: %w", err)
	}
	err = router.Declare(registry)
	if err != nil {
		return fmt.Errorf("failed to declare collector: %w", err)
	}
	err = router.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}
	go func() {
		select {
		case <-time.Tick(time.Second * 30):
			err = router.Collect()
			if err != nil {
				log.Printf("failed to collect metrics: %v", err)
			}
		}
	}()
	return nil
}

func initializeRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func startHTTPServer(port string, registry *prometheus.Registry) {
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	fmt.Printf("Starting server on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

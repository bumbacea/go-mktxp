package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
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
	var configDir string

	rootCmd := &cobra.Command{
		Use:   "metrics-server",
		Short: "Starts a metrics server",
		RunE: func(cmd *cobra.Command, args []string) error {
			wg := &sync.WaitGroup{}
			defer wg.Wait()
			// Create a context that is cancelable
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			// Set up signal handling for graceful shutdown
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-signals
				log.Printf("Received signal: %s, initiating shutdown...", sig)
				cancel()
			}()

			// Initialize registry and load configuration
			registry := prometheus.NewRegistry()
			globalConfig, err := config.LoadConfig(path.Join(configDir, "_mktxp.conf"))
			if err != nil {
				return fmt.Errorf("error parsing global config file: %w", err)
			}

			instances, err := config.ParseConfig(path.Join(configDir, "mktxp.conf"))
			if err != nil {
				return fmt.Errorf("error parsing instances config file: %w", err)
			}

			defaultInstance, ok := instances[defaultKey]
			if ok {
				delete(instances, defaultKey)
			}

			// Start collectors
			for instanceName, instanceConfig := range instances {
				mergedConfig := instanceConfig
				if ok {
					mergedConfig = config.MergeDefaults(defaultInstance, instanceConfig)
				}

				if mergedConfig.Enabled != nil && !*mergedConfig.Enabled {
					continue
				}

				mergedConfig.Name = instanceName
				log.Printf("Starting collector for router: %s", instanceName)

				if err := startCollector(registry, mergedConfig, ctx, wg); err != nil {
					return fmt.Errorf("failed to start collector for router %s: %w", instanceName, err)
				}
			}

			// Start HTTP server and serve until context is canceled
			serverErr := make(chan error, 1)
			go func() {
				serverErr <- startHTTPServer(globalConfig.Listen, registry, ctx)
			}()
			select {
			case <-ctx.Done():
				// Shutdown initiated via signal
				log.Println("Shutting down metrics server...")
				return nil
			case err := <-serverErr:
				// HTTP server error
				return fmt.Errorf("server error: %w", err)
			}
		},
	}

	rootCmd.Flags().StringVarP(&configDir, "cfg-dir", "", "./", "MKTXP config files directory (optional)")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func startCollector(registry prometheus.Registerer, conf config.RouterConfig, ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	router, err := collector.NewRouterEntry(conf)
	if err != nil {
		return fmt.Errorf("failed to create router entry: %w", err)
	}

	if err := router.Declare(registry); err != nil {
		return fmt.Errorf("failed to declare collector: %w", err)
	}

	if err := router.Collect(nil); err != nil {
		return fmt.Errorf("failed to collect initial metrics: %w", err)
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		defer wg.Done()

		for {
			select {
			case <-ticker.C:
				if err := router.Collect(ctx); err != nil {
					log.Printf("failed to collect metrics: %v", err)
				}
			case <-ctx.Done():
				router.Conn.Close()
				log.Printf("Stopping collector for router: %s", conf.Name)
				return
			}
		}
	}()

	return nil
}

func startHTTPServer(port string, registry *prometheus.Registry, ctx context.Context) error {
	server := &http.Server{
		Addr:    port,
		Handler: nil, // Default handler
	}

	// Set up handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf("called %s", r.URL.Path)))
	})
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Run server in a goroutine to allow graceful shutdown
	go func() {
		log.Printf("Starting HTTP server on %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done() // Wait for context cancellation

	// Shutdown server gracefully
	log.Println("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

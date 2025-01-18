package config

import (
	"gopkg.in/ini.v1"
)

type MKTXPConfig struct {
	Listen                   string
	SocketTimeout            int
	InitialDelayOnFailure    int
	MaxDelayOnFailure        int
	DelayIncDiv              int
	Bandwidth                bool
	BandwidthTestInterval    int
	MinimalCollectInterval   int
	VerboseMode              bool
	FetchRoutersInParallel   bool
	MaxWorkerThreads         int
	MaxScrapeDuration        int
	TotalMaxScrapeDuration   int
	CompactDefaultConfValues bool
}

func LoadConfig(filename string) (*MKTXPConfig, error) {
	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}

	config := &MKTXPConfig{}

	section := cfg.Section("MKTXP")
	config.Listen = section.Key("listen").MustString("0.0.0.0:49090")
	config.SocketTimeout = section.Key("socket_timeout").MustInt(5)
	config.InitialDelayOnFailure = section.Key("initial_delay_on_failure").MustInt(120)
	config.MaxDelayOnFailure = section.Key("max_delay_on_failure").MustInt(900)
	config.DelayIncDiv = section.Key("delay_inc_div").MustInt(5)
	config.Bandwidth = section.Key("bandwidth").MustBool(false)
	config.BandwidthTestInterval = section.Key("bandwidth_test_interval").MustInt(600)
	config.MinimalCollectInterval = section.Key("minimal_collect_interval").MustInt(5)
	config.VerboseMode = section.Key("verbose_mode").MustBool(false)
	config.FetchRoutersInParallel = section.Key("fetch_routers_in_parallel").MustBool(false)
	config.MaxWorkerThreads = section.Key("max_worker_threads").MustInt(5)
	config.MaxScrapeDuration = section.Key("max_scrape_duration").MustInt(30)
	config.TotalMaxScrapeDuration = section.Key("total_max_scrape_duration").MustInt(90)
	config.CompactDefaultConfValues = section.Key("compact_default_conf_values").MustBool(false)

	return config, nil
}

package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// RouterConfig represents the configuration for a single router.
type RouterConfig struct {
	Hostname string
	Port     int
	Username string
	Password string
	Enabled  *bool

	UseSSL               *bool `ini:"use_ssl"`
	NoSSLCertificate     *bool
	SSLCertificateVerify *bool `ini:"ssl_certificate_verify"`
	PlaintextLogin       *bool

	// Metrics settings
	InstalledPackages  *bool
	DHCP               *bool `ini:"dhcp"`
	DHCPLease          *bool `ini:"dhcp_lease"`
	Connections        *bool
	ConnectionStats    *bool
	Interface          *bool
	Route              *bool
	Pool               *bool
	Firewall           *bool
	Neighbor           *bool
	DNS                *bool `ini:"dns"`
	IPv6Route          *bool
	IPv6Pool           *bool
	IPv6Firewall       *bool
	IPv6Neighbor       *bool
	POE                *bool `ini:"poe"`
	Monitor            *bool
	Netwatch           *bool
	PublicIP           *bool `ini:"public_ip"`
	Wireless           *bool
	WirelessClients    *bool
	CAPsMAN            *bool `ini:"capsman"`
	CAPsMANClients     *bool `ini:"capsman_clients"`
	EoIP               *bool `ini:"eoip"`
	GRE                *bool `ini:"gre"`
	IPIP               *bool `ini:"ipip"`
	LTE                *bool `ini:"lte"`
	IPSec              *bool `ini:"ipsec"`
	SwitchPort         *bool
	KidControlAssigned *bool
	KidControlDynamic  *bool
	User               *bool
	Queue              *bool
	BGP                *bool `ini:"bgp"`
	RoutingStats       *bool
	Certificate        *bool
	RemoteDHCPEntry    string
	RemoteCAPsMANEntry string

	UseCommentsOverNames *bool
	CheckForUpdates      *bool
	Name                 string
}

// ParseConfig parses the configuration file into a map of RouterConfigs.
func ParseConfig(filePath string) (map[string]RouterConfig, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		Insensitive:                 false,
		InsensitiveSections:         false,
		InsensitiveKeys:             true,
		IgnoreContinuation:          false,
		IgnoreInlineComment:         false,
		SkipUnrecognizableLines:     false,
		ShortCircuit:                false,
		AllowBooleanKeys:            false,
		AllowShadows:                false,
		AllowNestedValues:           false,
		AllowPythonMultilineValues:  false,
		SpaceBeforeInlineComment:    false,
		UnescapeValueDoubleQuotes:   false,
		UnescapeValueCommentSymbols: false,
		UnparseableSections:         nil,
		KeyValueDelimiters:          "",
		KeyValueDelimiterOnWrite:    "",
		ChildSectionDelimiter:       "",
		PreserveSurroundedQuote:     false,
		DebugFunc:                   nil,
		ReaderBufferSize:            0,
		AllowNonUniqueSections:      false,
		AllowDuplicateShadowValues:  false,
	}, filePath)
	if err != nil {
		return nil, err
	}

	routers := make(map[string]RouterConfig)

	// Iterate through all sections in the file
	for _, section := range cfg.Sections() {
		name := section.Name()

		// Skip the default/global section (handled separately if needed)
		if name == ini.DefaultSection {
			continue
		}

		// Load RouterConfig from the section
		router := RouterConfig{}
		err := section.MapTo(&router)
		if err != nil {
			return nil, fmt.Errorf("error mapping section '%s': %w", name, err)
		}

		routers[name] = router
	}

	return routers, nil
}

func MergeDefaults(defaultConfig RouterConfig, instanceConfig RouterConfig) RouterConfig {
	if instanceConfig.Hostname == "" {
		instanceConfig.Hostname = defaultConfig.Hostname
	}
	if instanceConfig.Port == 0 {
		instanceConfig.Port = defaultConfig.Port
	}
	if instanceConfig.Username == "" {
		instanceConfig.Username = defaultConfig.Username
	}
	if instanceConfig.Password == "" {
		instanceConfig.Password = defaultConfig.Password
	}
	if instanceConfig.Enabled == nil {
		instanceConfig.Enabled = defaultConfig.Enabled
	}
	if instanceConfig.UseSSL == nil {
		instanceConfig.UseSSL = defaultConfig.UseSSL
	}
	if instanceConfig.NoSSLCertificate == nil {
		instanceConfig.NoSSLCertificate = defaultConfig.NoSSLCertificate
	}
	if instanceConfig.SSLCertificateVerify == nil {
		instanceConfig.SSLCertificateVerify = defaultConfig.SSLCertificateVerify
	}
	if instanceConfig.PlaintextLogin == nil {
		instanceConfig.PlaintextLogin = defaultConfig.PlaintextLogin
	}

	// Metrics
	if instanceConfig.InstalledPackages == nil {
		instanceConfig.InstalledPackages = defaultConfig.InstalledPackages
	}
	if instanceConfig.DHCP == nil {
		instanceConfig.DHCP = defaultConfig.DHCP
	}
	if instanceConfig.DHCPLease == nil {
		instanceConfig.DHCPLease = defaultConfig.DHCPLease
	}
	if instanceConfig.Connections == nil {
		instanceConfig.Connections = defaultConfig.Connections
	}
	if instanceConfig.ConnectionStats == nil {
		instanceConfig.ConnectionStats = defaultConfig.ConnectionStats
	}
	if instanceConfig.Interface == nil {
		instanceConfig.Interface = defaultConfig.Interface
	}
	if instanceConfig.Route == nil {
		instanceConfig.Route = defaultConfig.Route
	}
	if instanceConfig.Pool == nil {
		instanceConfig.Pool = defaultConfig.Pool
	}
	if instanceConfig.Firewall == nil {
		instanceConfig.Firewall = defaultConfig.Firewall
	}
	if instanceConfig.Neighbor == nil {
		instanceConfig.Neighbor = defaultConfig.Neighbor
	}
	if instanceConfig.DNS == nil {
		instanceConfig.DNS = defaultConfig.DNS
	}
	if instanceConfig.IPv6Route == nil {
		instanceConfig.IPv6Route = defaultConfig.IPv6Route
	}
	if instanceConfig.IPv6Pool == nil {
		instanceConfig.IPv6Pool = defaultConfig.IPv6Pool
	}
	if instanceConfig.IPv6Firewall == nil {
		instanceConfig.IPv6Firewall = defaultConfig.IPv6Firewall
	}
	if instanceConfig.IPv6Neighbor == nil {
		instanceConfig.IPv6Neighbor = defaultConfig.IPv6Neighbor
	}
	if instanceConfig.POE == nil {
		instanceConfig.POE = defaultConfig.POE
	}
	if instanceConfig.Monitor == nil {
		instanceConfig.Monitor = defaultConfig.Monitor
	}
	if instanceConfig.Netwatch == nil {
		instanceConfig.Netwatch = defaultConfig.Netwatch
	}
	if instanceConfig.PublicIP == nil {
		instanceConfig.PublicIP = defaultConfig.PublicIP
	}
	if instanceConfig.Wireless == nil {
		instanceConfig.Wireless = defaultConfig.Wireless
	}
	if instanceConfig.WirelessClients == nil {
		instanceConfig.WirelessClients = defaultConfig.WirelessClients
	}
	if instanceConfig.CAPsMAN == nil {
		instanceConfig.CAPsMAN = defaultConfig.CAPsMAN
	}
	if instanceConfig.CAPsMANClients == nil {
		instanceConfig.CAPsMANClients = defaultConfig.CAPsMANClients
	}
	if instanceConfig.EoIP == nil {
		instanceConfig.EoIP = defaultConfig.EoIP
	}
	if instanceConfig.GRE == nil {
		instanceConfig.GRE = defaultConfig.GRE
	}
	if instanceConfig.IPIP == nil {
		instanceConfig.IPIP = defaultConfig.IPIP
	}
	if instanceConfig.LTE == nil {
		instanceConfig.LTE = defaultConfig.LTE
	}
	if instanceConfig.IPSec == nil {
		instanceConfig.IPSec = defaultConfig.IPSec
	}
	if instanceConfig.SwitchPort == nil {
		instanceConfig.SwitchPort = defaultConfig.SwitchPort
	}
	if instanceConfig.KidControlAssigned == nil {
		instanceConfig.KidControlAssigned = defaultConfig.KidControlAssigned
	}
	if instanceConfig.KidControlDynamic == nil {
		instanceConfig.KidControlDynamic = defaultConfig.KidControlDynamic
	}
	if instanceConfig.User == nil {
		instanceConfig.User = defaultConfig.User
	}
	if instanceConfig.Queue == nil {
		instanceConfig.Queue = defaultConfig.Queue
	}
	if instanceConfig.BGP == nil {
		instanceConfig.BGP = defaultConfig.BGP
	}
	if instanceConfig.RoutingStats == nil {
		instanceConfig.RoutingStats = defaultConfig.RoutingStats
	}
	if instanceConfig.Certificate == nil {
		instanceConfig.Certificate = defaultConfig.Certificate
	}
	if instanceConfig.RemoteDHCPEntry == "" {
		instanceConfig.RemoteDHCPEntry = defaultConfig.RemoteDHCPEntry
	}
	if instanceConfig.RemoteCAPsMANEntry == "" {
		instanceConfig.RemoteCAPsMANEntry = defaultConfig.RemoteCAPsMANEntry
	}
	if instanceConfig.UseCommentsOverNames == nil {
		instanceConfig.UseCommentsOverNames = defaultConfig.UseCommentsOverNames
	}
	if instanceConfig.CheckForUpdates == nil {
		instanceConfig.CheckForUpdates = defaultConfig.CheckForUpdates
	}

	return instanceConfig

}

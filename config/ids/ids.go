package ids

const (
	// Shared
	MaxPeersID              string = "maxPeers"
	ContainerTagID          string = "containerTag"
	AdditionalFlagsID       string = "additionalFlags"
	HttpPortID              string = "httpPort"
	OpenHttpPortsID         string = "openHttpPort"
	P2pPortID               string = "p2pPort"
	PortID                  string = "port"
	OpenPortID              string = "openPort"
	HttpUrlID               string = "httpUrl"
	FastTimeoutID           string = "fastTimeout"
	SlowTimeoutID           string = "slowTimeout"
	EcID                    string = "executionClient"
	BnID                    string = "beaconNode"
	GraffitiID              string = "graffiti"
	DoppelgangerDetectionID string = "doppelgangerDetection"
	MetricsPortID           string = "metricsPort"
	CacheSizeID             string = "cacheSize"

	// Logger
	LoggerLevelID             string = "level"
	LoggerFormatID            string = "format"
	LoggerAddSourceID         string = "addSource"
	LoggerMaxSizeID           string = "maxSize"
	LoggerMaxBackupsID        string = "maxBackups"
	LoggerMaxAgeID            string = "maxAge"
	LoggerLocalTimeID         string = "localTime"
	LoggerCompressID          string = "compress"
	LoggerEnableHttpTracingID string = "enableHttpTracing"

	// Besu
	BesuJvmHeapSizeID   string = "jvmHeapSize"
	BesuMaxBackLayersID string = "maxBackLayers"
	BesuArchiveModeID   string = "archiveMode"

	// Bitfly
	BitflySecretID      string = "bitflySecret"
	BitflyEndpointID    string = "bitflyEndpoint"
	BitflyMachineNameID string = "bitflyMachineName"

	// Exporter
	ExporterEnableRootFsID string = "enableRootFs"

	// External Execution
	ExternalEcWebsocketUrlID string = "wsUrl"

	// Fallback
	FallbackUseFallbackClientsID string = "useFallbackClients"
	FallbackEcHttpUrlID          string = "ecHttpUrl"
	FallbackBnHttpUrlID          string = "bnHttpUrl"
	FallbackReconnectDelayID     string = "reconnectDelay"

	// Geth
	GethEvmTimeoutID  string = "evmTimeout"
	GethArchiveModeID string = "archiveMode"

	// Lighthouse
	LighthouseQuicPortID string = "p2pQuicPort"

	// Local Beacon Node
	LocalBnCheckpointSyncUrlID string = "checkpointSyncUrl"
	LocalBnLighthouseID        string = "lighthouse"
	LocalBnLodestarID          string = "lodestar"
	LocalBnNimbusID            string = "nimbus"
	LocalBnPrysmID             string = "prysm"
	LocalBnTekuID              string = "teku"

	// Local Execution Client
	LocalEcWebsocketPortID string = "wsPort"
	LocalEcEnginePortID    string = "enginePort"
	LocalEcOpenApiPortsID  string = "openApiPorts"
	LocalEcBesuID          string = "besu"
	LocalEcGethID          string = "geth"
	LocalEcNethermindID    string = "nethermind"
	LocalEcRethID          string = "reth"

	// Metrics
	MetricsEnableID       string = "enableMetrics"
	MetricsEnableBitflyID string = "enableBitflyNodeMetrics"
	MetricsEcPortID       string = "ecMetricsPort"
	MetricsBnPortID       string = "bnMetricsPort"
	MetricsDaemonPortID   string = "daemonMetricsPort"
	MetricsExporterPortID string = "exporterMetricsPort"
	MetricsGrafanaID      string = "grafana"
	MetricsPrometheusID   string = "prometheus"
	MetricsExporterID     string = "exporter"
	MetricsBitflyID       string = "bitfly"

	// Nethermind
	NethermindPruneMemSizeID           string = "pruneMemSize"
	NethermindAdditionalModulesID      string = "additionalModules"
	NethermindAdditionalUrlsID         string = "additionalUrls"
	NethermindFullPruneMemoryBudgetID  string = "fullPruneMemoryBudget"
	NethermindFullPruningThresholdMbID string = "fullPruningThresholdMb"

	// Nimbus
	NimbusPruningModeID string = "pruningMode"

	// Prysm
	PrysmRpcPortID     string = "rpcPort"
	PrysmOpenRpcPortID string = "openRpcPort"
	PrysmRpcUrlID      string = "prysmRpcUrl"

	// Reth
	RethMaxInboundPeersID  string = "maxInboundPeers"
	RethMaxOutboundPeersID string = "maxOutboundPeers"

	// Teku
	TekuJvmHeapSizeID           string = "jvmHeapSize"
	TekuArchiveModeID           string = "archiveMode"
	TekuUseSlashingProtectionID string = "useSlashingProtection"
)

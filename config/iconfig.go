package config

import (
	"time"

	"github.com/rocket-pool/node-manager-core/log"
)

// Timeout settings for the client
type ClientTimeouts struct {
	// The timeout for requests that are expected to be fast
	FastTimeout time.Duration

	// The timeout for requests that are expected to be slow and require either significant processing or a large return size from the server
	SlowTimeout time.Duration

	// The delay before rechecking the primary client, if fallbacks support is enabled
	RecheckDelay time.Duration
}

// NMC servers typically provide some kind of persistent configuration; it must implement this interface.
type IConfig interface {
	IConfigSection

	// The path to use for the API log file
	GetApiLogFilePath() string

	// The path to use for the tasks log file
	GetTasksLogFilePath() string

	// The path to use for the node address file
	GetNodeAddressFilePath() string

	// The path to use for the wallet keystore file
	GetWalletFilePath() string

	// The path to use for the wallet keystore's password file
	GetPasswordFilePath() string

	// The resources for the selected network
	GetNetworkResources() *NetworkResources

	// The URLs for the Execution clients to use
	GetExecutionClientUrls() (string, string)

	// The timeouts for the Execution clients and manager to use
	GetExecutionClientTimeouts() ClientTimeouts

	// The URLs for the Beacon nodes to use
	GetBeaconNodeUrls() (string, string)

	// The timeouts for the Beacon nodes and manager to use
	GetBeaconNodeTimeouts() ClientTimeouts

	// The configuration for the daemon loggers
	GetLoggerOptions() log.LoggerOptions
}

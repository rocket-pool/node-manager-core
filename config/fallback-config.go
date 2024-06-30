package config

import "github.com/rocket-pool/node-manager-core/config/ids"

// Fallback configuration
type FallbackConfig struct {
	// Flag for enabling fallback clients
	UseFallbackClients Parameter[bool]

	// The URL of the Execution Client HTTP endpoint
	EcHttpUrl Parameter[string]

	// The URL of the Beacon Node HTTP endpoint
	BnHttpUrl Parameter[string]

	// The URL of the Prysm gRPC endpoint (only needed if using Prysm VCs)
	PrysmRpcUrl Parameter[string]

	// The delay in milliseconds when checking a client again after it disconnects during a request
	ReconnectDelayMs Parameter[uint64]
}

// Generates a new FallbackConfig configuration
func NewFallbackConfig() *FallbackConfig {
	return &FallbackConfig{
		UseFallbackClients: Parameter[bool]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.FallbackUseFallbackClientsID,
				Name:               "Use Fallback Clients",
				Description:        "Enable this if you would like to specify a fallback Execution and Beacon Node, which will temporarily be used by your node and Validator Client(s) if your primary Execution / Beacon Node pair ever go offline (e.g. if you switch, prune, or resync your clients).",
				AffectsContainers:  []ContainerID{ContainerID_Daemon, ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]bool{
				Network_All: false,
			},
		},

		EcHttpUrl: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.FallbackEcHttpUrlID,
				Name:               "Execution Client URL",
				Description:        "The URL of the HTTP API endpoint for your fallback Execution client.\n\nNOTE: If you are running it on the same machine as your node, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},

		BnHttpUrl: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.FallbackBnHttpUrlID,
				Name:               "Beacon Node URL",
				Description:        "The URL of the HTTP Beacon API endpoint for your fallback Beacon Node.\n\nNOTE: If you are running it on the same machine as your node, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon, ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},

		PrysmRpcUrl: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.PrysmRpcUrlID,
				Name:               "RPC URL (Prysm Only)",
				Description:        "**Only used if you have a Prysm Validator Client.**\n\nThe URL of Prysm's gRPC API endpoint for your fallback Beacon Node. Prysm's Validator Client will need this in order to connect to it.\nNOTE: If you are running it on the same machine as your node, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},

		ReconnectDelayMs: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.FallbackReconnectDelayID,
				Name:               "Reconnect Delay",
				Description:        "The delay, in milliseconds, to wait after the primary Execution Client or primary Beacon Node disconnects during a request before trying it again.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint64{
				Network_All: 60000,
			},
		},
	}
}

// The title for the config
func (cfg *FallbackConfig) GetTitle() string {
	return "Fallback Clients"
}

// Get the Parameters for this config
func (cfg *FallbackConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.UseFallbackClients,
		&cfg.EcHttpUrl,
		&cfg.BnHttpUrl,
		&cfg.PrysmRpcUrl,
		&cfg.ReconnectDelayMs,
	}
}

// Get the sections underneath this one
func (cfg *FallbackConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

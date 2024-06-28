package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

// Configuration for external Execution clients
type ExternalExecutionConfig struct {
	// The selected EC
	ExecutionClient Parameter[ExecutionClient]

	// The URL of the HTTP endpoint
	HttpUrl Parameter[string]

	// The URL of the Websocket endpoint
	WebsocketUrl Parameter[string]

	// Number of seconds to wait for a fast request to complete
	FastTimeout Parameter[uint64]

	// Number of seconds to wait for a slow request to complete
	SlowTimeout Parameter[uint64]
}

// Generates a new ExternalExecutionConfig configuration
func NewExternalExecutionConfig() *ExternalExecutionConfig {
	return &ExternalExecutionConfig{
		ExecutionClient: Parameter[ExecutionClient]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.EcID,
				Name:               "Execution Client",
				Description:        "Select which Execution client your external client is.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Options: []*ParameterOption[ExecutionClient]{
				{
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Geth",
						Description: "Select if your external client is Geth.",
					},
					Value: ExecutionClient_Geth,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Nethermind",
						Description: "Select if your external client is Nethermind.",
					},
					Value: ExecutionClient_Nethermind,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Besu",
						Description: "Select if your external client is Besu.",
					},
					Value: ExecutionClient_Besu,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Reth",
						Description: "Select if your external client is Reth.",
					},
					Value: ExecutionClient_Reth,
				}},
			Default: map[Network]ExecutionClient{
				Network_All: ExecutionClient_Geth},
		},

		HttpUrl: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.HttpUrlID,
				Name:               "HTTP URL",
				Description:        "The URL of the HTTP RPC endpoint for your external Execution client.\nNOTE: If you are running it on the same machine as this node, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead, for example 'http://192.168.1.100:8545'.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},

		WebsocketUrl: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.ExternalEcWebsocketUrlID,
				Name:               "Websocket URL",
				Description:        "The URL of the Websocket RPC endpoint for your external Execution client.\nNOTE: If you are running it on the same machine as this node, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead, for example 'http://192.168.1.100:8545'.",
				AffectsContainers:  []ContainerID{},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},

		FastTimeout: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.FastTimeoutID,
				Name:               "Fast Timeout",
				Description:        "Number of seconds to wait for a request to complete that is expected to be fast and light before timing out the request.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint64{
				Network_All: 5,
			},
		},

		SlowTimeout: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.SlowTimeoutID,
				Name:               "Slow Timeout",
				Description:        "Number of seconds to wait for a request to complete that is expected to be slow and heavy, either taking a long time to process or returning a large amount of data, before timing out the request. Examples include filtering through Ethereum event logs.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint64{
				Network_All: 30,
			},
		},
	}
}

// The title for the config
func (cfg *ExternalExecutionConfig) GetTitle() string {
	return "External Execution Client"
}

// Get the parameters for this config
func (cfg *ExternalExecutionConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.ExecutionClient,
		&cfg.HttpUrl,
		&cfg.WebsocketUrl,
		&cfg.FastTimeout,
		&cfg.SlowTimeout,
	}
}

// Get the sections underneath this one
func (cfg *ExternalExecutionConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

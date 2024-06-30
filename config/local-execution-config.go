package config

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/config/ids"
)

// Configuration for the Execution client
type LocalExecutionConfig struct {
	// The selected EC
	ExecutionClient Parameter[ExecutionClient]

	// The HTTP API port
	HttpPort Parameter[uint16]

	// The Websocket API port
	WebsocketPort Parameter[uint16]

	// The Engine API port
	EnginePort Parameter[uint16]

	// Toggle for forwarding the HTTP API port outside of Docker
	OpenApiPorts Parameter[RpcPortMode]

	// P2P traffic port
	P2pPort Parameter[uint16]

	// Number of milliseconds to wait for a fast request to complete
	FastTimeoutMs Parameter[uint64]

	// Number of milliseconds to wait for a slow request to complete
	SlowTimeoutMs Parameter[uint64]

	// Subconfigs
	Geth       *GethConfig
	Nethermind *NethermindConfig
	Besu       *BesuConfig
	Reth       *RethConfig
}

// Create a new LocalExecutionConfig struct
func NewLocalExecutionConfig() *LocalExecutionConfig {
	cfg := &LocalExecutionConfig{
		ExecutionClient: Parameter[ExecutionClient]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.EcID,
				Name:               "Execution Client",
				Description:        "Select which Execution client you would like to run.",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient, ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Options: []*ParameterOption[ExecutionClient]{
				{
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "*Geth",
						Description: "Geth is one of the three original implementations of the Ethereum protocol. It is written in Go, fully open source and licensed under the GNU LGPL v3.\n\n[orange]NOTE: Geth is currently overrepresented on the Ethereum network (a \"supermajority\" clients). We recommend choosing a different client for the health of the network. Please see https://clientdiversity.org/ to learn more.",
					},
					Value: ExecutionClient_Geth,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Nethermind",
						Description: "Nethermind is a high-performance full Ethereum protocol client with very fast sync speeds. Nethermind is built with proven industrial technologies such as .NET 6 and the Kestrel web server. It is fully open source.",
					},
					Value: ExecutionClient_Nethermind,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Besu",
						Description: "Hyperledger Besu is a robust full Ethereum protocol client. It uses a novel system called \"Bonsai Trees\" to store its chain data efficiently, which allows it to access block states from the past and does not require pruning. Besu is fully open source and written in Java.",
					},
					Value: ExecutionClient_Besu,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Reth",
						Description: "Reth is a new Ethereum full node implementation that is focused on being user-friendly, highly modular, as well as being fast and efficient. Reth is fully open source and written in Rust.",
					},
					Value: ExecutionClient_Reth,
				}},
			Default: map[Network]ExecutionClient{
				Network_All: ExecutionClient_Geth,
			},
		},

		HttpPort: Parameter[uint16]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.HttpPortID,
				Name:               "HTTP API Port",
				Description:        "The port your Execution client should use for its HTTP API endpoint (also known as HTTP RPC API endpoint).",
				AffectsContainers:  []ContainerID{ContainerID_Daemon, ContainerID_ExecutionClient, ContainerID_BeaconNode},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint16{
				Network_All: 8545,
			},
		},

		WebsocketPort: Parameter[uint16]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.LocalEcWebsocketPortID,
				Name:               "Websocket API Port",
				Description:        "The port your Execution client should use for its Websocket API endpoint (also known as Websocket RPC API endpoint).",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint16{
				Network_All: 8546,
			},
		},

		EnginePort: Parameter[uint16]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.LocalEcEnginePortID,
				Name:               "Engine API Port",
				Description:        "The port your Execution client should use for its Engine API endpoint (the endpoint the Beacon Node will connect to post-merge).",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient, ContainerID_BeaconNode},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint16{
				Network_All: 8551,
			},
		},

		OpenApiPorts: Parameter[RpcPortMode]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.LocalEcOpenApiPortsID,
				Name:               "Expose API Ports",
				Description:        "Expose the HTTP and Websocket API ports to other processes on your machine, or to your local network so other machines can access your Execution Client's API endpoints.",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Options: GetPortModes(""),
			Default: map[Network]RpcPortMode{
				Network_All: RpcPortMode_Closed,
			},
		},

		P2pPort: Parameter[uint16]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.P2pPortID,
				Name:               "P2P Port",
				Description:        "The port the Execution Client should use for P2P (blockchain) traffic to communicate with other nodes.",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint16{
				Network_All: 30303,
			},
		},

		FastTimeoutMs: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.FastTimeoutID,
				Name:               "Fast Timeout",
				Description:        "Number of milliseconds to wait for a request to complete that is expected to be fast and light before timing out the request.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint64{
				Network_All: 5000,
			},
		},

		SlowTimeoutMs: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.SlowTimeoutID,
				Name:               "Slow Timeout",
				Description:        "Number of milliseconds to wait for a request to complete that is expected to be slow and heavy, either taking a long time to process or returning a large amount of data, before timing out the request. Examples include filtering through Ethereum event logs.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint64{
				Network_All: 30000,
			},
		},
	}

	// Create the subconfigs
	cfg.Geth = NewGethConfig()
	cfg.Nethermind = NewNethermindConfig()
	cfg.Besu = NewBesuConfig()
	cfg.Reth = NewRethConfig()

	return cfg
}

// Get the title for the config
func (cfg *LocalExecutionConfig) GetTitle() string {
	return "Local Execution Client"
}

// Get the parameters for this config
func (cfg *LocalExecutionConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.ExecutionClient,
		&cfg.HttpPort,
		&cfg.WebsocketPort,
		&cfg.EnginePort,
		&cfg.OpenApiPorts,
		&cfg.P2pPort,
		&cfg.FastTimeoutMs,
		&cfg.SlowTimeoutMs,
	}
}

// Get the sections underneath this one
func (cfg *LocalExecutionConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{
		ids.LocalEcBesuID:       cfg.Besu,
		ids.LocalEcGethID:       cfg.Geth,
		ids.LocalEcNethermindID: cfg.Nethermind,
		ids.LocalEcRethID:       cfg.Reth,
	}
}

// ==================
// === Templating ===
// ==================

// Get the Docker mapping for the selected API port mode
func (cfg *LocalExecutionConfig) GetOpenApiPortMapping() string {
	rpcMode := cfg.OpenApiPorts.Value
	if !rpcMode.IsOpen() {
		return ""
	}
	httpMapping := rpcMode.DockerPortMapping(cfg.HttpPort.Value)
	wsMapping := rpcMode.DockerPortMapping(cfg.WebsocketPort.Value)
	return fmt.Sprintf(", \"%s\", \"%s\"", httpMapping, wsMapping)
}

// Gets the max peers of the selected EC
// Note that Reth treats the max peer count specially
func (cfg *LocalExecutionConfig) GetMaxPeers() uint16 {
	switch cfg.ExecutionClient.Value {
	case ExecutionClient_Geth:
		return cfg.Geth.MaxPeers.Value
	case ExecutionClient_Nethermind:
		return cfg.Nethermind.MaxPeers.Value
	case ExecutionClient_Besu:
		return cfg.Besu.MaxPeers.Value
	case ExecutionClient_Reth:
		return cfg.Reth.MaxInboundPeers.Value + cfg.Reth.MaxOutboundPeers.Value
	default:
		panic(fmt.Sprintf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value)))
	}
}

// Get the container tag of the selected EC
func (cfg *LocalExecutionConfig) GetContainerTag() string {
	switch cfg.ExecutionClient.Value {
	case ExecutionClient_Geth:
		return cfg.Geth.ContainerTag.Value
	case ExecutionClient_Nethermind:
		return cfg.Nethermind.ContainerTag.Value
	case ExecutionClient_Besu:
		return cfg.Besu.ContainerTag.Value
	case ExecutionClient_Reth:
		return cfg.Reth.ContainerTag.Value
	default:
		panic(fmt.Sprintf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value)))
	}
}

// Gets the additional flags of the selected EC
func (cfg *LocalExecutionConfig) GetAdditionalFlags() string {
	switch cfg.ExecutionClient.Value {
	case ExecutionClient_Geth:
		return cfg.Geth.AdditionalFlags.Value
	case ExecutionClient_Nethermind:
		return cfg.Nethermind.AdditionalFlags.Value
	case ExecutionClient_Besu:
		return cfg.Besu.AdditionalFlags.Value
	case ExecutionClient_Reth:
		return cfg.Reth.AdditionalFlags.Value
	default:
		panic(fmt.Sprintf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value)))
	}
}

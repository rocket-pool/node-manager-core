package config

import "fmt"

const (
	// The standard name of the Ethereum mainnet network, used for EL / CL flags
	EthNetwork_Mainnet string = "mainnet"

	// The standard name of the Ethereum holesky network, used for EL / CL flags
	EthNetwork_Holesky string = "holesky"
)

// Deprecated: Switch to using dynamically-loaded network settings instead
type Network string

// Enum to describe well-known network keys for convenience, which may or may not be used by the application.
const (
	// Unknown
	Network_Unknown Network = ""

	// All networks (used for parameter defaults)
	Network_All Network = "all"

	// The Ethereum mainnet
	Network_Mainnet Network = Network(EthNetwork_Mainnet)

	// The Holesky test network
	Network_Holesky Network = Network(EthNetwork_Holesky)
)

// A Docker container name
type ContainerID string

// Enum to describe the names / IDs of various containers controlled by NMC
const (
	// Unknown
	ContainerID_Unknown ContainerID = ""

	// The daemon
	ContainerID_Daemon ContainerID = "daemon"

	// The Execution client
	ContainerID_ExecutionClient ContainerID = "ec"

	// The Beacon node (Beacon Node)
	ContainerID_BeaconNode ContainerID = "bn"

	// The Validator client
	ContainerID_ValidatorClient ContainerID = "vc"

	// MEV-Boost
	ContainerID_MevBoost ContainerID = "mev-boost"

	// The Node Exporter
	ContainerID_Exporter ContainerID = "exporter"

	// Prometheus
	ContainerID_Prometheus ContainerID = "prometheus"

	// Grafana
	ContainerID_Grafana ContainerID = "grafana"
)

// An Execution client
type ExecutionClient string

// Enum to describe the Execution clients
const (
	// Unknown
	ExecutionClient_Unknown ExecutionClient = ""

	// Geth
	ExecutionClient_Geth ExecutionClient = "geth"

	// Nethermind
	ExecutionClient_Nethermind ExecutionClient = "nethermind"

	// Besu
	ExecutionClient_Besu ExecutionClient = "besu"

	// Reth
	ExecutionClient_Reth ExecutionClient = "reth"
)

// A Beacon Node (Beacon Node)
type BeaconNode string

// Enum to describe the Beacon Nodes
const (
	// Unknown
	BeaconNode_Unknown BeaconNode = ""

	// Lighthouse
	BeaconNode_Lighthouse BeaconNode = "lighthouse"

	// Lodestar
	BeaconNode_Lodestar BeaconNode = "lodestar"

	// Nimbus
	BeaconNode_Nimbus BeaconNode = "nimbus"

	// Prysm
	BeaconNode_Prysm BeaconNode = "prysm"

	// Teku
	BeaconNode_Teku BeaconNode = "teku"
)

// A client ownership mode
type ClientMode string

// Enum to describe client modes
const (
	// Unknown
	ClientMode_Unknown ClientMode = ""

	// Locally-owned clients (managed by the NMC service)
	ClientMode_Local ClientMode = "local"

	// Externally-managed clients (managed by the user)
	ClientMode_External ClientMode = "external"
)

// How to expose the RPC ports
type RpcPortMode string

// Enum to describe the mode for the RPC port exposure setting
const (
	// Do not allow any connections to the RPC port
	RpcPortMode_Closed RpcPortMode = "closed"

	// Allow connections from the same host
	RpcPortMode_OpenLocalhost RpcPortMode = "localhost"

	// Allow connections from external hosts
	RpcPortMode_OpenExternal RpcPortMode = "external"
)

// True if the port is open locally or externally
func (m RpcPortMode) IsOpen() bool {
	return m == RpcPortMode_OpenLocalhost || m == RpcPortMode_OpenExternal
}

// Creates the appropriate Docker config string for the provided port, based on the port mode
func (m RpcPortMode) DockerPortMapping(port uint16) string {
	ports := fmt.Sprintf("%d:%d/tcp", port, port)

	switch m {
	case RpcPortMode_OpenExternal:
		return ports
	case RpcPortMode_OpenLocalhost:
		return fmt.Sprintf("127.0.0.1:%s", ports)
	default:
		return ""
	}
}

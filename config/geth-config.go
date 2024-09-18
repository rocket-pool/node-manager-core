package config

import (
	"fmt"
	"runtime"

	"github.com/rocket-pool/node-manager-core/config/ids"
)

// Constants
const (
	// Tags
	gethTag string = "ethereum/client-go:v1.14.9"
)

// Configuration for Geth
type GethConfig struct {
	// Max number of P2P peers to connect to
	MaxPeers Parameter[uint16]

	// Number of seconds EVM calls can run before timing out
	EvmTimeout Parameter[uint64]

	// The archive mode flag
	ArchiveMode Parameter[bool]

	// The Docker Hub tag for Geth
	ContainerTag Parameter[string]

	// Custom command line flags
	AdditionalFlags Parameter[string]
}

// Generates a new Geth configuration
func NewGethConfig() *GethConfig {
	return &GethConfig{
		MaxPeers: Parameter[uint16]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.MaxPeersID,
				Name:               "Max Peers",
				Description:        "The maximum number of peers Geth should connect to. This can be lowered to improve performance on low-power systems or constrained Networks. We recommend keeping it at 12 or higher.",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint16{
				Network_All: calculateGethPeers(),
			},
		},

		EvmTimeout: Parameter[uint64]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.GethEvmTimeoutID,
				Name:               "EVM Timeout",
				Description:        "The number of seconds an Execution Client API call is allowed to run before Geth times out and aborts it. Increase this if you see a lot of timeout errors in your logs.",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint64{
				Network_All: 5,
			},
		},

		ArchiveMode: Parameter[bool]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.GethArchiveModeID,
				Name:               "Enable Archive Mode",
				Description:        "When enabled, Geth will run in \"archive\" mode which means it can recreate the state of the chain for a previous block. This is required for manually generating the Merkle rewards tree.\n\nArchive mode takes several TB of disk space, so only enable it if you need it and can support it.",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]bool{
				Network_All: false,
			},
		},

		ContainerTag: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.ContainerTagID,
				Name:               "Container Tag",
				Description:        "The tag name of the Geth container you want to use on Docker Hub.",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[Network]string{
				Network_All: gethTag,
			},
		},

		AdditionalFlags: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.AdditionalFlagsID,
				Name:               "Additional Flags",
				Description:        "Additional custom command line flags you want to pass to Geth, to take advantage of other settings that aren't covered here.",
				AffectsContainers:  []ContainerID{ContainerID_ExecutionClient},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},
	}
}

// Get the title for the config
func (cfg *GethConfig) GetTitle() string {
	return "Geth"
}

// Get the parameters for this config
func (cfg *GethConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.MaxPeers,
		&cfg.EvmTimeout,
		&cfg.ArchiveMode,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *GethConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

// Calculate the default number of Geth peers
func calculateGethPeers() uint16 {
	switch runtime.GOARCH {
	case "arm64":
		return 25
	case "amd64":
		return 50
	default:
		panic(fmt.Sprintf("unsupported architecture %s", runtime.GOARCH))
	}
}

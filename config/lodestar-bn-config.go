package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

const (
	lodestarBnTag string = "chainsafe/lodestar:v1.22.0"
)

// Configuration for the Lodestar BN
type LodestarBnConfig struct {
	// The max number of P2P peers to connect to
	MaxPeers Parameter[uint16]

	// The Docker Hub tag for Lodestar BN
	ContainerTag Parameter[string]

	// Custom command line flags for the BN
	AdditionalFlags Parameter[string]
}

// Generates a new Lodestar BN configuration
func NewLodestarBnConfig() *LodestarBnConfig {
	return &LodestarBnConfig{
		MaxPeers: Parameter[uint16]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.MaxPeersID,
				Name:               "Max Peers",
				Description:        "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
				AffectsContainers:  []ContainerID{ContainerID_BeaconNode},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint16{
				Network_All: 100,
			},
		},

		ContainerTag: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.ContainerTagID,
				Name:               "Container Tag",
				Description:        "The tag name of the Lodestar container from Docker Hub you want to use for the Beacon Node.",
				AffectsContainers:  []ContainerID{ContainerID_BeaconNode},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[Network]string{
				Network_All: lodestarBnTag,
			},
		},

		AdditionalFlags: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.AdditionalFlagsID,
				Name:               "Additional Flags",
				Description:        "Additional custom command line flags you want to pass Lodestar's Beacon Client, to take advantage of other settings that aren't covered here.",
				AffectsContainers:  []ContainerID{ContainerID_BeaconNode},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},
	}
}

// The title for the config
func (cfg *LodestarBnConfig) GetTitle() string {
	return "Lodestar Beacon Node"
}

// Get the parameters for this config
func (cfg *LodestarBnConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.MaxPeers,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *LodestarBnConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

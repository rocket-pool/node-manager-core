package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

const (
	// Tags
	lighthouseBnTagProd string = "sigp/lighthouse:v5.2.1"
	lighthouseBnTagTest string = "sigp/lighthouse:v5.2.1"
)

// Configuration for the Lighthouse BN
type LighthouseBnConfig struct {
	// The port to use for gossip traffic using the QUIC protocol
	P2pQuicPort Parameter[uint16]

	// The max number of P2P peers to connect to
	MaxPeers Parameter[uint16]

	// The Docker Hub tag for Lighthouse BN
	ContainerTag Parameter[string]

	// Custom command line flags for the BN
	AdditionalFlags Parameter[string]
}

// Generates a new Lighthouse BN configuration
func NewLighthouseBnConfig() *LighthouseBnConfig {
	return &LighthouseBnConfig{
		P2pQuicPort: Parameter[uint16]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.LighthouseQuicPortID,
				Name:               "P2P QUIC Port",
				Description:        "The port to use for P2P (blockchain) traffic using the QUIC protocol.",
				AffectsContainers:  []ContainerID{ContainerID_BeaconNode},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]uint16{
				Network_All: 8001,
			},
		},

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
				Description:        "The tag name of the Lighthouse container from Docker Hub you want to use for the Beacon Node.",
				AffectsContainers:  []ContainerID{ContainerID_BeaconNode},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[Network]string{
				Network_Mainnet: lighthouseBnTagProd,
				Network_Holesky: lighthouseBnTagTest,
			},
		},

		AdditionalFlags: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.AdditionalFlagsID,
				Name:               "Additional Flags",
				Description:        "Additional custom command line flags you want to pass Lighthouse's Beacon Node, to take advantage of other settings that aren't covered here.",
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
func (cfg *LighthouseBnConfig) GetTitle() string {
	return "Lighthouse Beacon Node"
}

// Get the parameters for this config
func (cfg *LighthouseBnConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.MaxPeers,
		&cfg.P2pQuicPort,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *LighthouseBnConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

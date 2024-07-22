package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

const (
	// Tags
	prysmVcTag string = prysmBnTag
)

// Configuration for the Prysm VC
type PrysmVcConfig struct {
	// The Docker Hub tag for the Prysm BN
	ContainerTag Parameter[string]

	// Custom command line flags for the BN
	AdditionalFlags Parameter[string]
}

// Generates a new Prysm VC configuration
func NewPrysmVcConfig() *PrysmVcConfig {
	return &PrysmVcConfig{
		ContainerTag: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.ContainerTagID,
				Name:               "Validator Client Container Tag",
				Description:        "The tag name of the Prysm container on Docker Hub you want to use for the Validator Client.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[Network]string{
				Network_All: prysmVcTag,
			},
		},

		AdditionalFlags: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.AdditionalFlagsID,
				Name:               "Additional Validator Client Flags",
				Description:        "Additional custom command line flags you want to pass the Prysm Validator Client, to take advantage of other settings that aren't covered here.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
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
func (cfg *PrysmVcConfig) GetTitle() string {
	return "Prysm Validator Client"
}

// Get the parameters for this config
func (cfg *PrysmVcConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *PrysmVcConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

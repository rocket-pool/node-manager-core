package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

const (
	lodestarVcTag string = lodestarBnTag
)

// Configuration for the Lodestar VC
type LodestarVcConfig struct {
	// The Docker Hub tag for Lodestar VC
	ContainerTag Parameter[string]

	// Custom command line flags for the VC
	AdditionalFlags Parameter[string]
}

// Generates a new Lodestar VC configuration
func NewLodestarVcConfig() *LodestarVcConfig {
	return &LodestarVcConfig{
		ContainerTag: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.ContainerTagID,
				Name:               "Validator Client Container Tag",
				Description:        "The tag name of the Lodestar container from Docker Hub you want to use for the Validator Client.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[Network]string{
				Network_All: lodestarVcTag,
			},
		},

		AdditionalFlags: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.AdditionalFlagsID,
				Name:               "Additional Validator Client Flags",
				Description:        "Additional custom command line flags you want to pass the Lodestar Validator Client, to take advantage of other settings that aren't covered here.",
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
func (cfg *LodestarVcConfig) GetTitle() string {
	return "Lodestar Validator Client"
}

// Get the parameters for this config
func (cfg *LodestarVcConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *LodestarVcConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

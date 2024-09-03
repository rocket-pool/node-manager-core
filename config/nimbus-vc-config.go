package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

const (
	// Tags
	nimbusVcTag string = "statusim/nimbus-validator-client:multiarch-v24.8.0"
)

// Configuration for Nimbus
type NimbusVcConfig struct {
	// The Docker Hub tag for the VC
	ContainerTag Parameter[string]

	// Custom command line flags for the VC
	AdditionalFlags Parameter[string]
}

// Generates a new Nimbus VC configuration
func NewNimbusVcConfig() *NimbusVcConfig {
	return &NimbusVcConfig{
		ContainerTag: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.ContainerTagID,
				Name:               "Validator Client Container Tag",
				Description:        "The tag name of the Nimbus Validator Client container you want to use on Docker Hub.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[Network]string{
				Network_All: nimbusVcTag,
			},
		},

		AdditionalFlags: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.AdditionalFlagsID,
				Name:               "Additional Validator Client Flags",
				Description:        "Additional custom command line flags you want to pass the Nimbus Validator Client, to take advantage of other settings that aren't covered here.",
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

// Get the title for the config
func (cfg *NimbusVcConfig) GetTitle() string {
	return "Nimbus Validator Client"
}

// Get the parameters for this config
func (cfg *NimbusVcConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *NimbusVcConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

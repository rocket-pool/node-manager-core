package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

const (
	// Tags
	lighthouseVcTag string = lighthouseBnTag
)

// Configuration for the Lighthouse VC
type LighthouseVcConfig struct {
	// The Docker Hub tag for Lighthouse VC
	ContainerTag Parameter[string]

	// Custom command line flags for the VC
	AdditionalFlags Parameter[string]
}

// Generates a new Lighthouse VC configuration
func NewLighthouseVcConfig() *LighthouseVcConfig {
	return &LighthouseVcConfig{
		ContainerTag: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.ContainerTagID,
				Name:               "Validator Client Container Tag",
				Description:        "The tag name of the Lighthouse container from Docker Hub you want to use for the Validator Client.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[Network]string{
				Network_All: lighthouseVcTag,
			},
		},

		AdditionalFlags: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.AdditionalFlagsID,
				Name:               "Additional Validator Client Flags",
				Description:        "Additional custom command line flags you want to pass the Lighthouse Validator Client, to take advantage of other settings that aren't covered here.",
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
func (cfg *LighthouseVcConfig) GetTitle() string {
	return "Lighthouse Validator Client"
}

// Get the parameters for this config
func (cfg *LighthouseVcConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *LighthouseVcConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

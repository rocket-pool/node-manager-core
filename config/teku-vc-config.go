package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

const (
	// Tags
	tekuVcTag string = tekuBnTag
)

// Configuration for Teku
type TekuVcConfig struct {
	// The use slashing protection flag
	UseSlashingProtection Parameter[bool]

	// The Docker Hub tag for the Teku VC
	ContainerTag Parameter[string]

	// Custom command line flags for the VC
	AdditionalFlags Parameter[string]
}

// Generates a new Teku VC configuration
func NewTekuVcConfig() *TekuVcConfig {
	return &TekuVcConfig{
		UseSlashingProtection: Parameter[bool]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.TekuUseSlashingProtectionID,
				Name:               "Use Validator Slashing Protection",
				Description:        "When enabled, Teku will use the Validator Slashing Protection feature. See https://docs.teku.consensys.io/how-to/prevent-slashing/detect-slashing for details.",
				AffectsContainers:  []ContainerID{ContainerID_BeaconNode, ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]bool{
				Network_All: true,
			},
		},

		ContainerTag: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.ContainerTagID,
				Name:               "Validator Client Container Tag",
				Description:        "The tag name of the Teku container on Docker Hub you want to use for the Validator Client.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[Network]string{
				Network_All: tekuVcTag,
			},
		},

		AdditionalFlags: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.AdditionalFlagsID,
				Name:               "Additional Validator Client Flags",
				Description:        "Additional custom command line flags you want to pass the Teku Validator Client, to take advantage of other settings that aren't covered here.",
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
func (cfg *TekuVcConfig) GetTitle() string {
	return "Teku Validator Client"
}

// Get the parameters for this config
func (cfg *TekuVcConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.UseSlashingProtection,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the sections underneath this one
func (cfg *TekuVcConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

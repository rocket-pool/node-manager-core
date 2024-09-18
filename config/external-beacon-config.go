package config

import (
	"github.com/rocket-pool/node-manager-core/config/ids"
)

// Configuration for external Beacon Nodes
type ExternalBeaconConfig struct {
	// The selected BN
	BeaconNode Parameter[BeaconNode]

	// The URL of the HTTP endpoint
	HttpUrl Parameter[string]

	// The URL of the Prysm gRPC endpoint (only needed if using Prysm VCs)
	PrysmRpcUrl Parameter[string]
}

// Generates a new ExternalBeaconConfig configuration
func NewExternalBeaconConfig() *ExternalBeaconConfig {
	return &ExternalBeaconConfig{
		BeaconNode: Parameter[BeaconNode]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.BnID,
				Name:               "Beacon Node",
				Description:        "Select which Beacon Node your external client is.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Options: []*ParameterOption[BeaconNode]{
				{
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Lighthouse",
						Description: "Select if your external client is Lighthouse.",
					},
					Value: BeaconNode_Lighthouse,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Lodestar",
						Description: "Select if your external client is Lodestar.",
					},
					Value: BeaconNode_Lodestar,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Nimbus",
						Description: "Select if your external client is Nimbus.",
					},
					Value: BeaconNode_Nimbus,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Prysm",
						Description: "Select if your external client is Prysm.",
					},
					Value: BeaconNode_Prysm,
				}, {
					ParameterOptionCommon: &ParameterOptionCommon{
						Name:        "Teku",
						Description: "Select if your external client is Teku.",
					},
					Value: BeaconNode_Teku,
				}},
			Default: map[Network]BeaconNode{
				Network_All: BeaconNode_Nimbus,
			},
		},

		HttpUrl: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.HttpUrlID,
				Name:               "HTTP URL",
				Description:        "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as this node, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead, for example 'http://192.168.1.100:5052'.",
				AffectsContainers:  []ContainerID{ContainerID_Daemon, ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},

		PrysmRpcUrl: Parameter[string]{
			ParameterCommon: &ParameterCommon{
				ID:                 ids.PrysmRpcUrlID,
				Name:               "Prysm RPC URL",
				Description:        "The URL of Prysm's gRPC API endpoint for your external Beacon Node. Prysm's Validator Client will need this in order to connect to it.\nNOTE: If you are running it on the same machine as this node, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead, for example 'http://192.168.1.100:5053'.",
				AffectsContainers:  []ContainerID{ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[Network]string{
				Network_All: "",
			},
		},
	}
}

// The title for the config
func (cfg *ExternalBeaconConfig) GetTitle() string {
	return "External Beacon Node"
}

// Get the parameters for this config
func (cfg *ExternalBeaconConfig) GetParameters() []IParameter {
	return []IParameter{
		&cfg.BeaconNode,
		&cfg.HttpUrl,
		&cfg.PrysmRpcUrl,
	}
}

// Get the sections underneath this one
func (cfg *ExternalBeaconConfig) GetSubconfigs() map[string]IConfigSection {
	return map[string]IConfigSection{}
}

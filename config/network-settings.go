package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v3"
)

const ()

var (
	DefaultMainnetSettings = &NetworkSettings{
		Key:         "mainnet",
		Name:        "Ethereum Mainnet",
		Description: "The Ethereum Mainnet network",
		NetworkResources: &NetworkResources{
			EthNetworkName:        "mainnet",
			ChainID:               1,
			GenesisForkVersion:    common.FromHex("0x00000000"), // https://github.com/eth-clients/eth2-networks/tree/master/shared/mainnet#genesis-information
			MulticallAddress:      common.HexToAddress("0x5BA1e12693Dc8F9c48aAD8770482f4739bEeD696"),
			BalanceBatcherAddress: common.HexToAddress("0xb1f8e55c7f64d203c1400b9d8555d050f94adf39"),
			TxWatchUrl:            "https://etherscan.io/tx",
			FlashbotsProtectUrl:   "https://rpc.flashbots.net/",
		},
		DefaultConfigSettings: map[string]any{},
	}

	DefaultHoleskySettings = &NetworkSettings{
		Key:         "holesky",
		Name:        "Holesky Testnet",
		Description: "The Ethereum holesky public test network",
		NetworkResources: &NetworkResources{
			EthNetworkName:        "holesky",
			ChainID:               17000,
			GenesisForkVersion:    common.FromHex("0x01017000"), // https://github.com/eth-clients/holesky
			MulticallAddress:      common.HexToAddress("0x0540b786f03c9491f3a2ab4b0e3ae4ecd4f63ce7"),
			BalanceBatcherAddress: common.HexToAddress("0xfAa2e7C84eD801dd9D27Ac1ed957274530796140"),
			TxWatchUrl:            "https://holesky.etherscan.io/tx",
			FlashbotsProtectUrl:   "https://rpc-holesky.flashbots.net",
		},
		DefaultConfigSettings: map[string]any{},
	}
)

// A collection of network-specific resources and getters for them
type NetworkResources struct {
	// The actual name of the underlying Ethereum network, passed into the clients
	EthNetworkName string `yaml:"ethNetworkName" json:"ethNetworkName"`

	// The chain ID for the current network
	ChainID uint `yaml:"chainID" json:"chainID"`

	// The genesis fork version for the network according to the Beacon config for the network
	GenesisForkVersion []byte `yaml:"genesisForkVersion" json:"genesisForkVersion"`

	// The address of the multicall contract
	MulticallAddress common.Address `yaml:"multicallAddress" json:"multicallAddress"`

	// The BalanceChecker contract address
	BalanceBatcherAddress common.Address `yaml:"balanceBatcherAddress" json:"balanceBatcherAddress"`

	// The URL for transaction monitoring on the network's chain explorer
	TxWatchUrl string `yaml:"txWatchUrl" json:"txWatchUrl"`

	// The FlashBots Protect RPC endpoint
	FlashbotsProtectUrl string `yaml:"flashbotsProtectUrl" json:"flashbotsProtectUrl"`
}

// NetworkSettings contains all of the settings for a given Ethereum network
type NetworkSettings struct {
	// The unique key used to identify the network in the configuration
	Key Network `yaml:"key" json:"key"`

	// Human-readable name of the network
	Name string `yaml:"name" json:"name"`

	// A brief description of the network
	Description string `yaml:"description" json:"description"`

	// The list of resources for the network
	NetworkResources *NetworkResources `yaml:"networkResources" json:"networkResources"`

	// A collection of default configuration settings to use for the network, which will override
	// the standard "general-purpose" default value for the setting
	DefaultConfigSettings map[string]any `yaml:"defaultConfigSettings" json:"defaultConfigSettings"`
}

// Load network settings from a file
func LoadSettingsFile(path string) (*NetworkSettings, error) {
	// Make sure the file exists
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("network settings file [%s] does not exist", path)
	}

	// Load the file
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading network settings file [%s]: %w", path, err)
	}

	// Unmarshal the settings
	settings := new(NetworkSettings)
	err = yaml.Unmarshal(bytes, settings)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling network settings file [%s]: %w", path, err)
	}

	return settings, nil
}

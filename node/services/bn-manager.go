package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
)

// This is a proxy for multiple Beacon clients, providing natural fallback support if one of them fails.
type BeaconClientManager struct {
	primaryBc        beacon.IBeaconClient
	fallbackBc       beacon.IBeaconClient
	primaryReady     bool
	fallbackReady    bool
	expectedChainID  uint
	fallbackEnabled  bool
	primaryFailTime  time.Time
	fallbackFailTime time.Time
	recheckTime      time.Duration
}

// Creates a new BeaconClientManager instance
func NewBeaconClientManager(primaryBc beacon.IBeaconClient, chainID uint) *BeaconClientManager {
	return &BeaconClientManager{
		primaryBc:       primaryBc,
		primaryReady:    true,
		fallbackReady:   false,
		expectedChainID: chainID,
		fallbackEnabled: false,
	}
}

// Creates a new BeaconClientManager instance with a fallback client
func NewBeaconClientManagerWithFallback(primaryBc beacon.IBeaconClient, fallbackBc beacon.IBeaconClient, chainID uint, clientRecheckTime time.Duration) *BeaconClientManager {
	return &BeaconClientManager{
		primaryBc:       primaryBc,
		fallbackBc:      fallbackBc,
		primaryReady:    true,
		fallbackReady:   true,
		expectedChainID: chainID,
		fallbackEnabled: true,
		recheckTime:     clientRecheckTime,
	}
}

/// ========================
/// IClientManager Functions
/// ========================

func (m *BeaconClientManager) GetPrimaryClient() beacon.IBeaconClient {
	return m.primaryBc
}

func (m *BeaconClientManager) GetFallbackClient() beacon.IBeaconClient {
	return m.fallbackBc
}

func (m *BeaconClientManager) IsPrimaryReady() bool {
	return m.primaryReady
}

func (m *BeaconClientManager) IsFallbackReady() bool {
	return m.fallbackReady
}

func (m *BeaconClientManager) IsFallbackEnabled() bool {
	return m.fallbackEnabled
}

func (m *BeaconClientManager) GetClientTypeName() string {
	return "Beacon Node"
}

func (m *BeaconClientManager) SetPrimaryReady(ready bool) {
	m.primaryReady = ready
	if ready {
		m.primaryFailTime = time.Time{}
	} else {
		m.primaryFailTime = time.Now()
	}
}

func (m *BeaconClientManager) SetFallbackReady(ready bool) {
	m.fallbackReady = ready
	if ready {
		m.fallbackFailTime = time.Time{}
	} else {
		m.fallbackFailTime = time.Now()
	}
}

func (m *BeaconClientManager) RecheckFailTimes(logger *log.Logger) {
	if !m.fallbackEnabled {
		return
	}

	if !m.primaryReady {
		timeSincePrimaryFail := time.Since(m.primaryFailTime)
		if time.Since(m.primaryFailTime) > m.recheckTime {
			if logger != nil {
				logger.Info("Reconnecting primary Beacon client",
					slog.Duration("elapsed", timeSincePrimaryFail),
				)
			}
			m.SetPrimaryReady(true)
		}
	}
	if !m.fallbackReady {
		timeSinceFallbackFail := time.Since(m.fallbackFailTime)
		if time.Since(m.fallbackFailTime) > m.recheckTime {
			if logger != nil {
				logger.Info("Reconnecting fallback Beacon client",
					slog.Duration("elapsed", timeSinceFallbackFail),
				)
			}
			m.SetFallbackReady(true)
		}
	}
}

/// =======================
/// IBeaconClient Functions
/// =======================

// Get the client's sync status
func (m *BeaconClientManager) GetSyncStatus(ctx context.Context) (beacon.SyncStatus, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (beacon.SyncStatus, error) {
		return client.GetSyncStatus(ctx)
	})
}

// Get the Beacon configuration
func (m *BeaconClientManager) GetEth2Config(ctx context.Context) (beacon.Eth2Config, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (beacon.Eth2Config, error) {
		return client.GetEth2Config(ctx)
	})
}

// Get the Beacon configuration
func (m *BeaconClientManager) GetEth2DepositContract(ctx context.Context) (beacon.Eth2DepositContract, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (beacon.Eth2DepositContract, error) {
		return client.GetEth2DepositContract(ctx)
	})
}

// Get the attestations in a Beacon chain block
func (m *BeaconClientManager) GetAttestations(ctx context.Context, blockId string) ([]beacon.AttestationInfo, bool, error) {
	return runFunction2(m, ctx, func(client beacon.IBeaconClient) ([]beacon.AttestationInfo, bool, error) {
		return client.GetAttestations(ctx, blockId)
	})
}

// Get a Beacon chain block
func (m *BeaconClientManager) GetBeaconBlock(ctx context.Context, blockId string) (beacon.BeaconBlock, bool, error) {
	return runFunction2(m, ctx, func(client beacon.IBeaconClient) (beacon.BeaconBlock, bool, error) {
		return client.GetBeaconBlock(ctx, blockId)
	})
}

// Get the header of a Beacon chain block
func (m *BeaconClientManager) GetBeaconBlockHeader(ctx context.Context, blockId string) (beacon.BeaconBlockHeader, bool, error) {
	return runFunction2(m, ctx, func(client beacon.IBeaconClient) (beacon.BeaconBlockHeader, bool, error) {
		return client.GetBeaconBlockHeader(ctx, blockId)
	})
}

// Get the Beacon chain's head information
func (m *BeaconClientManager) GetBeaconHead(ctx context.Context) (beacon.BeaconHead, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (beacon.BeaconHead, error) {
		return client.GetBeaconHead(ctx)
	})
}

// Get a validator's status by its index
func (m *BeaconClientManager) GetValidatorStatusByIndex(ctx context.Context, index string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (beacon.ValidatorStatus, error) {
		return client.GetValidatorStatusByIndex(ctx, index, opts)
	})
}

// Get a validator's status by its pubkey
func (m *BeaconClientManager) GetValidatorStatus(ctx context.Context, pubkey beacon.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (beacon.ValidatorStatus, error) {
		return client.GetValidatorStatus(ctx, pubkey, opts)
	})
}

// Get the statuses of multiple validators by their pubkeys
func (m *BeaconClientManager) GetValidatorStatuses(ctx context.Context, pubkeys []beacon.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (map[beacon.ValidatorPubkey]beacon.ValidatorStatus, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (map[beacon.ValidatorPubkey]beacon.ValidatorStatus, error) {
		return client.GetValidatorStatuses(ctx, pubkeys, opts)
	})
}

// Get a validator's index
func (m *BeaconClientManager) GetValidatorIndex(ctx context.Context, pubkey beacon.ValidatorPubkey) (string, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (string, error) {
		return client.GetValidatorIndex(ctx, pubkey)
	})
}

// Get a validator's sync duties
func (m *BeaconClientManager) GetValidatorSyncDuties(ctx context.Context, indices []string, epoch uint64) (map[string]bool, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (map[string]bool, error) {
		return client.GetValidatorSyncDuties(ctx, indices, epoch)
	})
}

// Get a validator's proposer duties
func (m *BeaconClientManager) GetValidatorProposerDuties(ctx context.Context, indices []string, epoch uint64) (map[string]uint64, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (map[string]uint64, error) {
		return client.GetValidatorProposerDuties(ctx, indices, epoch)
	})
}

// Get the Beacon chain's domain data
func (m *BeaconClientManager) GetDomainData(ctx context.Context, domainType []byte, epoch uint64, useGenesisFork bool) ([]byte, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) ([]byte, error) {
		return client.GetDomainData(ctx, domainType, epoch, useGenesisFork)
	})
}

// Voluntarily exit a validator
func (m *BeaconClientManager) ExitValidator(ctx context.Context, validatorIndex string, epoch uint64, signature beacon.ValidatorSignature) error {
	return runFunction0(m, ctx, func(client beacon.IBeaconClient) error {
		return client.ExitValidator(ctx, validatorIndex, epoch, signature)
	})
}

// Close the connection to the Beacon client
func (m *BeaconClientManager) Close(ctx context.Context) error {
	return runFunction0(m, ctx, func(client beacon.IBeaconClient) error {
		return client.Close(ctx)
	})
}

// Get the EL data for a CL block
func (m *BeaconClientManager) GetEth1DataForEth2Block(ctx context.Context, blockId string) (beacon.Eth1Data, bool, error) {
	return runFunction2(m, ctx, func(client beacon.IBeaconClient) (beacon.Eth1Data, bool, error) {
		return client.GetEth1DataForEth2Block(ctx, blockId)
	})
}

// Get the attestation committees for an epoch
func (m *BeaconClientManager) GetCommitteesForEpoch(ctx context.Context, epoch *uint64) (beacon.Committees, error) {
	return runFunction1(m, ctx, func(client beacon.IBeaconClient) (beacon.Committees, error) {
		return client.GetCommitteesForEpoch(ctx, epoch)
	})
}

// Change the withdrawal credentials for a validator
func (m *BeaconClientManager) ChangeWithdrawalCredentials(ctx context.Context, validatorIndex string, fromBlsPubkey beacon.ValidatorPubkey, toExecutionAddress common.Address, signature beacon.ValidatorSignature) error {
	return runFunction0(m, ctx, func(client beacon.IBeaconClient) error {
		return client.ChangeWithdrawalCredentials(ctx, validatorIndex, fromBlsPubkey, toExecutionAddress, signature)
	})
}

/// =================
/// Manager Functions
/// =================

// Get the status of the primary and fallback clients
func (m *BeaconClientManager) CheckStatus(ctx context.Context, checkChainIDs bool) *types.ClientManagerStatus {
	status := &types.ClientManagerStatus{
		FallbackEnabled: m.fallbackEnabled,
	}

	// Get the primary BC status
	status.PrimaryClientStatus = checkBcStatus(ctx, m.primaryBc, checkChainIDs)
	if checkChainIDs && status.PrimaryClientStatus.Error == "" && status.PrimaryClientStatus.ChainId != m.expectedChainID {
		m.primaryReady = false
		status.PrimaryClientStatus.Error = fmt.Sprintf("The primary client is using a different chain (%d) than what your node is configured for (%d)", status.PrimaryClientStatus.ChainId, m.expectedChainID)
	} else {
		// Flag if primary client is ready
		m.primaryReady = (status.PrimaryClientStatus.IsWorking && status.PrimaryClientStatus.IsSynced)
	}

	// Get the fallback BC status if applicable
	if status.FallbackEnabled {
		status.FallbackClientStatus = checkBcStatus(ctx, m.fallbackBc, checkChainIDs)
		// Check if fallback is using the expected network
		if checkChainIDs && status.FallbackClientStatus.Error == "" && status.FallbackClientStatus.ChainId != m.expectedChainID {
			m.fallbackReady = false
			status.FallbackClientStatus.Error = fmt.Sprintf("The fallback client is using a different chain (%d) than what your node is configured for (%d)", status.FallbackClientStatus.ChainId, m.expectedChainID)
			return status
		}
	}

	m.fallbackReady = (status.FallbackEnabled && status.FallbackClientStatus.IsWorking && status.FallbackClientStatus.IsSynced)

	return status
}

// Check the client status
func checkBcStatus(ctx context.Context, client beacon.IBeaconClient, checkChainIDs bool) types.ClientStatus {
	status := types.ClientStatus{}

	if checkChainIDs {
		// Get the Chain ID
		contractInfo, err := client.GetEth2DepositContract(ctx)
		if err != nil {
			status.Error = fmt.Sprintf("Chain ID check failed with [%s]", err.Error())
			status.IsSynced = false
			status.IsWorking = false
			return status
		}

		status.ChainId = uint(contractInfo.ChainID)
	}

	// Get the client's sync progress
	syncStatus, err := client.GetSyncStatus(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("Sync progress check failed with [%s]", err.Error())
		status.IsSynced = false
		status.IsWorking = false
		return status
	}

	// Return the sync status
	if !syncStatus.Syncing {
		status.IsWorking = true
		status.IsSynced = true
		status.SyncProgress = 1
	} else {
		status.IsWorking = true
		status.IsSynced = false
		status.SyncProgress = syncStatus.Progress
	}
	return status
}

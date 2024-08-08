package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	"golang.org/x/sync/errgroup"
)

// Beacon client using the standard Beacon HTTP REST API (https://ethereum.github.io/beacon-APIs/)
type StandardClient struct {
	provider IBeaconApiProvider
}

// Create a new client instance
func NewStandardClient(provider IBeaconApiProvider) *StandardClient {
	return &StandardClient{
		provider: provider,
	}
}

// Close the client connection
func (c *StandardClient) Close(ctx context.Context) error {
	return nil
}

// Get the client's underlying Beacon HTTP API provider
func (c *StandardClient) GetProvider() IBeaconApiProvider {
	return c.provider
}

// Get the node's sync status
func (c *StandardClient) GetSyncStatus(ctx context.Context) (beacon.SyncStatus, error) {
	// Get sync status
	syncStatus, err := c.provider.Node_Syncing(ctx)
	if err != nil {
		return beacon.SyncStatus{}, err
	}

	// Calculate the progress
	progress := float64(syncStatus.Data.HeadSlot) / float64(syncStatus.Data.HeadSlot+syncStatus.Data.SyncDistance)

	// Return response
	return beacon.SyncStatus{
		Syncing:  syncStatus.Data.IsSyncing,
		Progress: progress,
	}, nil
}

// Get the eth2 config
func (c *StandardClient) GetEth2Config(ctx context.Context) (beacon.Eth2Config, error) {
	// Data
	var wg errgroup.Group
	var eth2Config Eth2ConfigResponse
	var genesis GenesisResponse

	// Get eth2 config
	wg.Go(func() error {
		var err error
		eth2Config, err = c.provider.Config_Spec(ctx)
		return err
	})

	// Get genesis
	wg.Go(func() error {
		var err error
		genesis, err = c.provider.Beacon_Genesis(ctx)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return beacon.Eth2Config{}, err
	}

	// Return response
	return beacon.Eth2Config{
		GenesisForkVersion:           genesis.Data.GenesisForkVersion,
		GenesisValidatorsRoot:        genesis.Data.GenesisValidatorsRoot,
		GenesisEpoch:                 0,
		GenesisTime:                  uint64(genesis.Data.GenesisTime),
		SecondsPerSlot:               uint64(eth2Config.Data.SecondsPerSlot),
		SlotsPerEpoch:                uint64(eth2Config.Data.SlotsPerEpoch),
		SecondsPerEpoch:              uint64(eth2Config.Data.SecondsPerSlot * eth2Config.Data.SlotsPerEpoch),
		EpochsPerSyncCommitteePeriod: uint64(eth2Config.Data.EpochsPerSyncCommitteePeriod),
		ShardCommitteePeriod:         uint64(eth2Config.Data.ShardCommitteePeriod),
	}, nil
}

// Get the eth2 deposit contract info
func (c *StandardClient) GetEth2DepositContract(ctx context.Context) (beacon.Eth2DepositContract, error) {
	// Get the deposit contract
	depositContract, err := c.provider.Config_DepositContract(ctx)
	if err != nil {
		return beacon.Eth2DepositContract{}, err
	}

	// Return response
	return beacon.Eth2DepositContract{
		ChainID: uint64(depositContract.Data.ChainID),
		Address: depositContract.Data.Address,
	}, nil
}

// Get the beacon head
func (c *StandardClient) GetBeaconHead(ctx context.Context) (beacon.BeaconHead, error) {
	// Data
	var wg errgroup.Group
	var eth2Config beacon.Eth2Config
	var finalityCheckpoints FinalityCheckpointsResponse

	// Get eth2 config
	wg.Go(func() error {
		var err error
		eth2Config, err = c.GetEth2Config(ctx)
		return err
	})

	// Get finality checkpoints
	wg.Go(func() error {
		var err error
		finalityCheckpoints, err = c.provider.Beacon_FinalityCheckpoints(ctx, "head")
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return beacon.BeaconHead{}, err
	}

	// Return response
	return beacon.BeaconHead{
		Epoch:                  epochAt(eth2Config, uint64(time.Now().Unix())),
		FinalizedEpoch:         uint64(finalityCheckpoints.Data.Finalized.Epoch),
		JustifiedEpoch:         uint64(finalityCheckpoints.Data.CurrentJustified.Epoch),
		PreviousJustifiedEpoch: uint64(finalityCheckpoints.Data.PreviousJustified.Epoch),
	}, nil
}

// Get a validator's status
func (c *StandardClient) GetValidatorStatus(ctx context.Context, pubkey beacon.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	return c.getValidatorStatus(ctx, pubkey.HexWithPrefix(), opts)
}

func (c *StandardClient) GetValidatorStatusByIndex(ctx context.Context, index string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	return c.getValidatorStatus(ctx, index, opts)
}

func (c *StandardClient) getValidatorStatus(ctx context.Context, pubkeyOrIndex string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	// Return zero status for null pubkeyOrIndex
	if pubkeyOrIndex == "" {
		return beacon.ValidatorStatus{}, nil
	}

	// Get validator
	validators, err := c.getValidatorsByOpts(ctx, []string{pubkeyOrIndex}, opts)
	if err != nil {
		return beacon.ValidatorStatus{}, err
	}
	if len(validators.Data) == 0 {
		return beacon.ValidatorStatus{}, nil
	}
	validator := validators.Data[0]

	// Return response
	return beacon.ValidatorStatus{
		Pubkey:                     beacon.ValidatorPubkey(validator.Validator.Pubkey),
		Index:                      validator.Index,
		WithdrawalCredentials:      common.BytesToHash(validator.Validator.WithdrawalCredentials),
		Balance:                    uint64(validator.Balance),
		EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
		Status:                     beacon.ValidatorState(validator.Status),
		Slashed:                    validator.Validator.Slashed,
		ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
		ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
		ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
		WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
		Exists:                     true,
	}, nil

}

// Get multiple validators' statuses
func (c *StandardClient) GetValidatorStatuses(ctx context.Context, pubkeys []beacon.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (map[beacon.ValidatorPubkey]beacon.ValidatorStatus, error) {
	// The null validator pubkey
	nullPubkey := beacon.ValidatorPubkey{}

	// Filter out null, invalid and duplicate pubkeys
	realPubkeys := []beacon.ValidatorPubkey{}
	for _, pubkey := range pubkeys {
		if bytes.Equal(pubkey[:], nullPubkey[:]) {
			continue
		}
		isDuplicate := false
		for _, pk := range realPubkeys {
			if bytes.Equal(pubkey[:], pk[:]) {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}

		// Teku doesn't like invalid pubkeys, so filter them out to make it consistent with other clients
		_, err := bls.PublicKeyFromBytes(pubkey[:])
		if err != nil {
			return nil, fmt.Errorf("error creating pubkey from %s: %w", pubkey.HexWithPrefix(), err)
		}
		realPubkeys = append(realPubkeys, pubkey)
	}
	// Convert pubkeys into hex strings
	pubkeysHex := make([]string, len(realPubkeys))
	for vi := 0; vi < len(realPubkeys); vi++ {
		pubkeysHex[vi] = realPubkeys[vi].HexWithPrefix()
	}
	// Get validators
	validators, err := c.getValidatorsByOpts(ctx, pubkeysHex, opts)
	if err != nil {
		return nil, err
	}

	// Build validator status map
	statuses := make(map[beacon.ValidatorPubkey]beacon.ValidatorStatus)
	for _, validator := range validators.Data {

		// Ignore empty pubkeys
		if bytes.Equal(validator.Validator.Pubkey, nullPubkey[:]) {
			continue
		}

		// Get validator pubkey
		pubkey := beacon.ValidatorPubkey(validator.Validator.Pubkey)

		// Add status
		statuses[pubkey] = beacon.ValidatorStatus{
			Pubkey:                     beacon.ValidatorPubkey(validator.Validator.Pubkey),
			Index:                      validator.Index,
			WithdrawalCredentials:      common.BytesToHash(validator.Validator.WithdrawalCredentials),
			Balance:                    uint64(validator.Balance),
			EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
			Status:                     beacon.ValidatorState(validator.Status),
			Slashed:                    validator.Validator.Slashed,
			ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
			ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
			ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
			WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
			Exists:                     true,
		}

	}

	// Put an empty status in for null pubkeys
	statuses[nullPubkey] = beacon.ValidatorStatus{}

	// Return
	return statuses, nil

}

// Get whether validators have sync duties to perform at given epoch
func (c *StandardClient) GetValidatorSyncDuties(ctx context.Context, indices []string, epoch uint64) (map[string]bool, error) {
	// Perform the post request
	response, err := c.provider.Validator_DutiesSync_Post(ctx, indices, epoch)
	if err != nil {
		return nil, err
	}

	// Map the results
	validatorMap := make(map[string]bool)

	for _, index := range indices {
		validatorMap[index] = false
		for _, duty := range response.Data {
			if duty.ValidatorIndex == index {
				validatorMap[index] = true
				break
			}
		}
	}

	return validatorMap, nil
}

// Sums proposer duties per validators for a given epoch
func (c *StandardClient) GetValidatorProposerDuties(ctx context.Context, indices []string, epoch uint64) (map[string]uint64, error) {
	// Perform the post request
	response, err := c.provider.Validator_DutiesProposer(ctx, indices, epoch)
	if err != nil {
		return nil, err
	}

	// Map the results
	proposerMap := make(map[string]uint64)

	for _, index := range indices {
		proposerMap[index] = 0
		for _, duty := range response.Data {
			if duty.ValidatorIndex == index {
				proposerMap[index]++
				break
			}
		}
	}

	return proposerMap, nil
}

// Get a validator's index
func (c *StandardClient) GetValidatorIndex(ctx context.Context, pubkey beacon.ValidatorPubkey) (string, error) {
	// Get validator
	pubkeyString := pubkey.HexWithPrefix()
	validators, err := c.getValidatorsByOpts(ctx, []string{pubkeyString}, nil)
	if err != nil {
		return "", err
	}
	if len(validators.Data) == 0 {
		return "", fmt.Errorf("validator %s index not found", pubkeyString)
	}
	validator := validators.Data[0]

	// Return validator index
	return validator.Index, nil
}

// Get domain data for a domain type at a given epoch
func (c *StandardClient) GetDomainData(ctx context.Context, domainType []byte, epoch uint64, useGenesisFork bool) ([]byte, error) {
	// Data
	var wg errgroup.Group
	var genesis GenesisResponse
	var eth2Config Eth2ConfigResponse

	// Get genesis
	wg.Go(func() error {
		var err error
		genesis, err = c.provider.Beacon_Genesis(ctx)
		return err
	})

	// Get the BN spec as we need the CAPELLA_FORK_VERSION
	wg.Go(func() error {
		var err error
		eth2Config, err = c.provider.Config_Spec(ctx)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return []byte{}, err
	}

	// Get fork version
	var forkVersion []byte
	if useGenesisFork {
		// Used to compute the domain for credential changes
		forkVersion = genesis.Data.GenesisForkVersion
	} else {
		// According to EIP-7044 (https://eips.ethereum.org/EIPS/eip-7044) the CAPELLA_FORK_VERSION should always be used to compute the domain for voluntary exits signatures.
		forkVersion = eth2Config.Data.CapellaForkVersion
	}

	// Compute & return domain
	var dt [4]byte
	copy(dt[:], domainType[:])
	return eth2types.ComputeDomain(dt, forkVersion, genesis.Data.GenesisValidatorsRoot)
}

// Perform a voluntary exit on a validator
func (c *StandardClient) ExitValidator(ctx context.Context, validatorIndex string, epoch uint64, signature beacon.ValidatorSignature) error {
	return c.provider.Beacon_VoluntaryExits_Post(ctx, VoluntaryExitRequest{
		Message: VoluntaryExitMessage{
			Epoch:          utils.Uinteger(epoch),
			ValidatorIndex: validatorIndex,
		},
		Signature: signature[:],
	})
}

// Get the ETH1 data for the target beacon block
func (c *StandardClient) GetEth1DataForEth2Block(ctx context.Context, blockId string) (beacon.Eth1Data, bool, error) {
	// Get the Beacon block
	block, exists, err := c.provider.Beacon_Block(ctx, blockId)
	if err != nil {
		return beacon.Eth1Data{}, false, err
	}
	if !exists {
		return beacon.Eth1Data{}, false, nil
	}

	// Convert the response to the eth1 data struct
	return beacon.Eth1Data{
		DepositRoot:  common.BytesToHash(block.Data.Message.Body.Eth1Data.DepositRoot),
		DepositCount: uint64(block.Data.Message.Body.Eth1Data.DepositCount),
		BlockHash:    common.BytesToHash(block.Data.Message.Body.Eth1Data.BlockHash),
	}, true, nil
}

func (c *StandardClient) GetAttestations(ctx context.Context, blockId string) ([]beacon.AttestationInfo, bool, error) {
	attestations, exists, err := c.provider.Beacon_Attestations(ctx, blockId)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	// Add attestation info
	attestationInfo := make([]beacon.AttestationInfo, len(attestations.Data))
	for i, attestation := range attestations.Data {
		bitString := utils.RemovePrefix(attestation.AggregationBits)
		attestationInfo[i].SlotIndex = uint64(attestation.Data.Slot)
		attestationInfo[i].CommitteeIndex = uint64(attestation.Data.Index)
		attestationInfo[i].AggregationBits, err = hex.DecodeString(bitString)
		if err != nil {
			return nil, false, fmt.Errorf("error decoding aggregation bits for attestation %d of block %s: %w", i, blockId, err)
		}
	}

	return attestationInfo, true, nil
}

func (c *StandardClient) GetBeaconBlock(ctx context.Context, blockId string) (beacon.BeaconBlock, bool, error) {
	block, exists, err := c.provider.Beacon_Block(ctx, blockId)
	if err != nil {
		return beacon.BeaconBlock{}, false, err
	}
	if !exists {
		return beacon.BeaconBlock{}, false, nil
	}

	beaconBlock := beacon.BeaconBlock{
		Header: beacon.BeaconBlockHeader{
			Slot:          uint64(block.Data.Message.Slot),
			ProposerIndex: block.Data.Message.ProposerIndex,
		},
	}

	// Execution payload only exists after the merge, so check for its existence
	if block.Data.Message.Body.ExecutionPayload == nil {
		beaconBlock.HasExecutionPayload = false
	} else {
		beaconBlock.HasExecutionPayload = true
		beaconBlock.FeeRecipient = common.BytesToAddress(block.Data.Message.Body.ExecutionPayload.FeeRecipient)
		beaconBlock.ExecutionBlockNumber = uint64(block.Data.Message.Body.ExecutionPayload.BlockNumber)
	}

	// Add attestation info
	for i, attestation := range block.Data.Message.Body.Attestations {
		bitString := utils.RemovePrefix(attestation.AggregationBits)
		info := beacon.AttestationInfo{
			SlotIndex:      uint64(attestation.Data.Slot),
			CommitteeIndex: uint64(attestation.Data.Index),
		}
		info.AggregationBits, err = hex.DecodeString(bitString)
		if err != nil {
			return beacon.BeaconBlock{}, false, fmt.Errorf("error decoding aggregation bits for attestation %d of block %s: %w", i, blockId, err)
		}
		beaconBlock.Attestations = append(beaconBlock.Attestations, info)
	}

	return beaconBlock, true, nil
}

func (c *StandardClient) GetBeaconBlockHeader(ctx context.Context, blockId string) (beacon.BeaconBlockHeader, bool, error) {
	block, exists, err := c.provider.Beacon_Header(ctx, blockId)
	if err != nil {
		fmt.Printf("Error getting beacon block header: %s\n", err.Error())
		return beacon.BeaconBlockHeader{}, false, err
	}
	if !exists {
		return beacon.BeaconBlockHeader{}, false, nil
	}
	header := beacon.BeaconBlockHeader{
		Slot:          uint64(block.Data.Header.Message.Slot),
		ProposerIndex: block.Data.Header.Message.ProposerIndex,
	}
	return header, true, nil
}

// Get the attestation committees for the given epoch, or the current epoch if nil
func (c *StandardClient) GetCommitteesForEpoch(ctx context.Context, epoch *uint64) (beacon.Committees, error) {
	response, err := c.provider.Beacon_Committees(ctx, "head", epoch)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Perform a withdrawal credentials change on a validator
func (c *StandardClient) ChangeWithdrawalCredentials(ctx context.Context, validatorIndex string, fromBlsPubkey beacon.ValidatorPubkey, toExecutionAddress common.Address, signature beacon.ValidatorSignature) error {
	return c.provider.Beacon_BlsToExecutionChanges_Post(ctx, BLSToExecutionChangeRequest{
		Message: BLSToExecutionChangeMessage{
			ValidatorIndex:     validatorIndex,
			FromBLSPubkey:      fromBlsPubkey[:],
			ToExecutionAddress: toExecutionAddress[:],
		},
		Signature: signature[:],
	})
}

// Get fork
/*
func (c *StandardClient) getFork(ctx context.Context, stateId string) (ForkResponse, error) {
	responseBody, status, err := c.getRequest(ctx, fmt.Sprintf(RequestForkPath, stateId))
	if err != nil {
		return ForkResponse{}, fmt.Errorf("error getting fork data: %w", err)
	}
	if status != http.StatusOK {
		return ForkResponse{}, fmt.Errorf("error getting fork data: HTTP status %d; response body: '%s'", status, string(responseBody))
	}
	var fork ForkResponse
	if err := json.Unmarshal(responseBody, &fork); err != nil {
		return ForkResponse{}, fmt.Errorf("error decoding fork data: %w", err)
	}
	return fork, nil
}
*/

// Get validators by pubkeys and status options
func (c *StandardClient) getValidatorsByOpts(ctx context.Context, pubkeysOrIndices []string, opts *beacon.ValidatorStatusOptions) (ValidatorsResponse, error) {
	// Get state ID
	var stateId string
	if opts == nil {
		stateId = "head"
	} else if opts.Slot != nil {
		stateId = strconv.FormatInt(int64(*opts.Slot), 10)
	} else if opts.Epoch != nil {

		// Get eth2 config
		eth2Config, err := c.provider.Config_Spec(ctx)
		if err != nil {
			return ValidatorsResponse{}, err
		}

		// Get slot nuimber
		slot := *opts.Epoch * uint64(eth2Config.Data.SlotsPerEpoch)
		stateId = strconv.FormatInt(int64(slot), 10)

	} else {
		return ValidatorsResponse{}, fmt.Errorf("must specify a slot or epoch when calling getValidatorsByOpts")
	}

	count := len(pubkeysOrIndices)
	data := make([]Validator, count)
	validFlags := make([]bool, count)
	var wg errgroup.Group
	wg.SetLimit(runtime.NumCPU() / 2)
	for i := 0; i < count; i += MaxRequestValidatorsCount {
		i := i
		max := i + MaxRequestValidatorsCount
		if max > count {
			max = count
		}

		wg.Go(func() error {
			// Get & add validators
			batch := pubkeysOrIndices[i:max]
			validators, err := c.provider.Beacon_Validators(ctx, stateId, batch)
			if err != nil {
				return fmt.Errorf("error getting validator statuses: %w", err)
			}
			for j, responseData := range validators.Data {
				data[i+j] = responseData
				validFlags[i+j] = true
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return ValidatorsResponse{}, fmt.Errorf("error getting validators by opts: %w", err)
	}

	// Clip all of the empty responses so only the valid pubkeys get returned
	trueData := make([]Validator, 0, count)
	for i, valid := range validFlags {
		if valid {
			trueData = append(trueData, data[i])
		}
	}

	return ValidatorsResponse{Data: trueData}, nil
}

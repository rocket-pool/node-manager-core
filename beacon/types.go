package beacon

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/go-bitfield"
)

// API request options
type ValidatorStatusOptions struct {
	Epoch *uint64
	Slot  *uint64
}

// API response types
type SyncStatus struct {
	Syncing  bool
	Progress float64
}
type Eth2Config struct {
	GenesisForkVersion           []byte
	GenesisValidatorsRoot        []byte
	GenesisEpoch                 uint64
	GenesisTime                  uint64
	SecondsPerSlot               uint64
	SlotsPerEpoch                uint64
	SecondsPerEpoch              uint64
	EpochsPerSyncCommitteePeriod uint64
	ShardCommitteePeriod         uint64
}
type Eth2DepositContract struct {
	ChainID uint64
	Address common.Address
}
type BeaconHead struct {
	Epoch                  uint64
	FinalizedEpoch         uint64
	JustifiedEpoch         uint64
	PreviousJustifiedEpoch uint64
}
type ValidatorStatus struct {
	Pubkey                     ValidatorPubkey
	Index                      string
	WithdrawalCredentials      common.Hash
	Balance                    uint64
	Status                     ValidatorState
	EffectiveBalance           uint64
	Slashed                    bool
	ActivationEligibilityEpoch uint64
	ActivationEpoch            uint64
	ExitEpoch                  uint64
	WithdrawableEpoch          uint64
	Exists                     bool
}
type Eth1Data struct {
	DepositRoot  common.Hash
	DepositCount uint64
	BlockHash    common.Hash
}
type BeaconBlock struct {
	Header               BeaconBlockHeader
	HasExecutionPayload  bool
	Attestations         []AttestationInfo
	FeeRecipient         common.Address
	ExecutionBlockNumber uint64
}
type BeaconBlockHeader struct {
	Slot          uint64
	ProposerIndex string
}

// Committees is an interface as an optimization- since committees responses
// are quite large, there's a decent cpu/memory improvement to removing the
// translation to an intermediate storage class.
//
// Instead, the interface provides the access pattern that utilities want,
// and the underlying format is just the format of the Beacon Node response.
type Committees interface {
	// Index returns the index of the committee at the provided offset
	Index(int) uint64

	// Slot returns the slot of the committee at the provided offset
	Slot(int) uint64

	// Validators returns the list of validators of the committee at
	// the provided offset
	Validators(int) []string

	// Count returns the number of committees in the response
	Count() int

	// Release returns the reused validators slice buffer to the pool for
	// further reuse, and must be called when the user is done with this
	// committees instance
	Release()
}

type AttestationInfo struct {
	AggregationBits bitfield.Bitlist
	SlotIndex       uint64
	CommitteeIndex  uint64
}

type ValidatorState string

const (
	ValidatorState_PendingInitialized ValidatorState = "pending_initialized"
	ValidatorState_PendingQueued      ValidatorState = "pending_queued"
	ValidatorState_ActiveOngoing      ValidatorState = "active_ongoing"
	ValidatorState_ActiveExiting      ValidatorState = "active_exiting"
	ValidatorState_ActiveSlashed      ValidatorState = "active_slashed"
	ValidatorState_ExitedUnslashed    ValidatorState = "exited_unslashed"
	ValidatorState_ExitedSlashed      ValidatorState = "exited_slashed"
	ValidatorState_WithdrawalPossible ValidatorState = "withdrawal_possible"
	ValidatorState_WithdrawalDone     ValidatorState = "withdrawal_done"
)

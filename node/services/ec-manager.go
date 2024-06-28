package services

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	apitypes "github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
)

// This is a proxy for multiple ETH clients, providing natural fallback support if one of them fails.
type ExecutionClientManager struct {
	primaryEc        eth.IExecutionClient
	fallbackEc       eth.IExecutionClient
	primaryReady     bool
	fallbackReady    bool
	expectedChainID  uint
	fallbackEnabled  bool
	primaryFailTime  time.Time
	fallbackFailTime time.Time
	recheckTime      time.Duration
}

// Creates a new ExecutionClientManager instance
func NewExecutionClientManager(primaryEc eth.IExecutionClient, chainID uint) *ExecutionClientManager {
	return &ExecutionClientManager{
		primaryEc:       primaryEc,
		primaryReady:    true,
		fallbackReady:   false,
		expectedChainID: chainID,
		fallbackEnabled: false,
	}
}

// Creates a new ExecutionClientManager instance that includes a fallback client
func NewExecutionClientManagerWithFallback(primaryEc eth.IExecutionClient, fallbackEc eth.IExecutionClient, chainID uint, clientRecheckTime time.Duration) *ExecutionClientManager {
	return &ExecutionClientManager{
		primaryEc:       primaryEc,
		fallbackEc:      fallbackEc,
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

func (m *ExecutionClientManager) GetPrimaryClient() eth.IExecutionClient {
	return m.primaryEc
}

func (m *ExecutionClientManager) GetFallbackClient() eth.IExecutionClient {
	return m.fallbackEc
}

func (m *ExecutionClientManager) IsPrimaryReady() bool {
	return m.primaryReady
}

func (m *ExecutionClientManager) IsFallbackReady() bool {
	return m.fallbackReady
}

func (m *ExecutionClientManager) IsFallbackEnabled() bool {
	return m.fallbackEnabled
}

func (m *ExecutionClientManager) GetClientTypeName() string {
	return "Execution Client"
}

func (m *ExecutionClientManager) SetPrimaryReady(ready bool) {
	m.primaryReady = ready
	if ready {
		m.primaryFailTime = time.Time{}
	} else {
		m.primaryFailTime = time.Now()
	}
}

func (m *ExecutionClientManager) SetFallbackReady(ready bool) {
	m.fallbackReady = ready
	if ready {
		m.fallbackFailTime = time.Time{}
	} else {
		m.fallbackFailTime = time.Now()
	}
}

func (m *ExecutionClientManager) RecheckFailTimes(logger *log.Logger) {
	if !m.fallbackEnabled {
		return
	}

	if !m.primaryReady {
		timeSincePrimaryFail := time.Since(m.primaryFailTime)
		if time.Since(m.primaryFailTime) > m.recheckTime {
			if logger != nil {
				logger.Info("Reconnecting primary Execution client",
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
				logger.Info("Reconnecting fallback Execution client",
					slog.Duration("elapsed", timeSinceFallbackFail),
				)
			}
			m.SetFallbackReady(true)
		}
	}
}

/// ========================
/// ContractCaller Functions
/// ========================

// CodeAt returns the code of the given account. This is needed to differentiate
// between contract internal errors and the local chain being out of sync.
func (m *ExecutionClientManager) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) ([]byte, error) {
		return client.CodeAt(ctx, contract, blockNumber)
	})
}

// CallContract executes an Ethereum contract call with the specified data as the
// input.
func (m *ExecutionClientManager) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) ([]byte, error) {
		return client.CallContract(ctx, call, blockNumber)
	})
}

/// ============================
/// ContractTransactor Functions
/// ============================

// HeaderByHash returns the block header with the given hash.
func (m *ExecutionClientManager) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (*types.Header, error) {
		return client.HeaderByHash(ctx, hash)
	})
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (m *ExecutionClientManager) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (*types.Header, error) {
		return client.HeaderByNumber(ctx, number)
	})
}

// PendingCodeAt returns the code of the given account in the pending state.
func (m *ExecutionClientManager) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) ([]byte, error) {
		return client.PendingCodeAt(ctx, account)
	})
}

// PendingNonceAt retrieves the current pending nonce associated with an account.
func (m *ExecutionClientManager) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (uint64, error) {
		return client.PendingNonceAt(ctx, account)
	})
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (m *ExecutionClientManager) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (*big.Int, error) {
		return client.SuggestGasPrice(ctx)
	})
}

// SuggestGasTipCap retrieves the currently suggested 1559 priority fee to allow
// a timely execution of a transaction.
func (m *ExecutionClientManager) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (*big.Int, error) {
		return client.SuggestGasTipCap(ctx)
	})
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
// There is no guarantee that this is the true gas limit requirement as other
// transactions may be added or removed by miners, but it should provide a basis
// for setting a reasonable default.
func (m *ExecutionClientManager) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (uint64, error) {
		return client.EstimateGas(ctx, call)
	})
}

// SendTransaction injects the transaction into the pending pool for execution.
func (m *ExecutionClientManager) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return runFunction0(m, ctx, func(client eth.IExecutionClient) error {
		return client.SendTransaction(ctx, tx)
	})
}

/// ==========================
/// ContractFilterer Functions
/// ==========================

// FilterLogs executes a log filter operation, blocking during execution and
// returning all the results in one batch.
//
// TODO(karalabe): Deprecate when the subscription one can return past data too.
func (m *ExecutionClientManager) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) ([]types.Log, error) {
		return client.FilterLogs(ctx, query)
	})
}

// SubscribeFilterLogs creates a background log filtering operation, returning
// a subscription immediately, which can be used to stream the found events.
func (m *ExecutionClientManager) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (ethereum.Subscription, error) {
		return client.SubscribeFilterLogs(ctx, query, ch)
	})
}

/// =======================
/// DeployBackend Functions
/// =======================

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (m *ExecutionClientManager) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (*types.Receipt, error) {
		return client.TransactionReceipt(ctx, txHash)
	})
}

/// ================
/// Client functions
/// ================

// BlockNumber returns the most recent block number
func (m *ExecutionClientManager) BlockNumber(ctx context.Context) (uint64, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (uint64, error) {
		return client.BlockNumber(ctx)
	})
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (m *ExecutionClientManager) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (*big.Int, error) {
		return client.BalanceAt(ctx, account, blockNumber)
	})
}

// TransactionByHash returns the transaction with the given hash.
func (m *ExecutionClientManager) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	return runFunction2(m, ctx, func(client eth.IExecutionClient) (*types.Transaction, bool, error) {
		return client.TransactionByHash(ctx, hash)
	})
}

// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (m *ExecutionClientManager) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (uint64, error) {
		return client.NonceAt(ctx, account, blockNumber)
	})
}

// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (m *ExecutionClientManager) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (*ethereum.SyncProgress, error) {
		return client.SyncProgress(ctx)
	})
}

/// =======================
/// ChainIDReader Functions
/// =======================

func (m *ExecutionClientManager) ChainID(ctx context.Context) (*big.Int, error) {
	return runFunction1(m, ctx, func(client eth.IExecutionClient) (*big.Int, error) {
		return client.ChainID(ctx)
	})
}

/// =================
/// Manager Functions
/// =================

// Get the status of the primary and fallback clients
func (m *ExecutionClientManager) CheckStatus(ctx context.Context, checkChainIDs bool) *apitypes.ClientManagerStatus {
	status := &apitypes.ClientManagerStatus{
		FallbackEnabled: m.fallbackEnabled,
	}

	// Get the primary EC status
	status.PrimaryClientStatus = checkEcStatus(ctx, m.primaryEc, checkChainIDs)

	// Check if primary is using the expected network
	if checkChainIDs && status.PrimaryClientStatus.Error == "" && status.PrimaryClientStatus.ChainId != m.expectedChainID {
		m.primaryReady = false
		status.PrimaryClientStatus.Error = fmt.Sprintf("The primary client is using a different chain (%d) than what your node is configured for (%d)", status.PrimaryClientStatus.ChainId, m.expectedChainID)
	} else {
		// Flag if primary client is ready
		m.primaryReady = (status.PrimaryClientStatus.IsWorking && status.PrimaryClientStatus.IsSynced)
	}

	// Get the fallback EC status if applicable
	if status.FallbackEnabled {
		status.FallbackClientStatus = checkEcStatus(ctx, m.fallbackEc, checkChainIDs)
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
func checkEcStatus(ctx context.Context, client eth.IExecutionClient, checkChainIDs bool) apitypes.ClientStatus {
	status := apitypes.ClientStatus{}

	if checkChainIDs {
		// Get the Chain ID
		chainId, err := client.ChainID(ctx)
		if err != nil {
			status.Error = fmt.Sprintf("Chain ID check failed with [%s]", err.Error())
			status.IsSynced = false
			status.IsWorking = false
			return status
		}

		if chainId != nil {
			status.ChainId = uint(chainId.Uint64())
		}
	}

	// Get the client's sync progress
	progress, err := client.SyncProgress(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("Sync progress check failed with [%s]", err.Error())
		status.IsSynced = false
		status.IsWorking = false
		return status
	}

	// Make sure it's up to date
	if progress == nil {

		isUpToDate, blockTime, err := IsSyncWithinThreshold(client)
		if err != nil {
			status.Error = fmt.Sprintf("Error checking if client's sync progress is up to date: [%s]", err.Error())
			status.IsSynced = false
			status.IsWorking = false
			return status
		}

		status.IsWorking = true
		if !isUpToDate {
			status.Error = fmt.Sprintf("Client claims to have finished syncing, but its last block was from %s ago. It likely doesn't have enough peers", time.Since(blockTime))
			status.IsSynced = false
			status.SyncProgress = 0
			return status
		}

		// It's synced and it works!
		status.IsSynced = true
		status.SyncProgress = 1
		return status

	}

	// It's not synced yet, print the progress
	status.IsWorking = true
	status.IsSynced = false

	status.SyncProgress = float64(progress.CurrentBlock) / float64(progress.HighestBlock)
	if status.SyncProgress > 1 {
		status.SyncProgress = 1
	}
	if math.IsNaN(status.SyncProgress) {
		status.SyncProgress = 0
	}

	return status
}

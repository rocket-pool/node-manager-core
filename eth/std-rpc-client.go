package eth

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"net/http/httptrace"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/node-manager-core/log"
)

// Standard RPC-based Execution Client binding with logging support, using Geth as the backing client implementation.
type StandardRpcClient struct {
	client      *ethclient.Client
	fastTimeout time.Duration
	slowTimeout time.Duration
}

// Creates a new StandardRpcClient instance
func NewStandardRpcClient(address string, fastTimeout time.Duration, slowTimeout time.Duration) (*StandardRpcClient, error) {
	client, err := ethclient.Dial(address)
	if err != nil {
		return nil, fmt.Errorf("error creating EC binding for [%s]: %w", address, err)
	}
	return &StandardRpcClient{
		client:      client,
		fastTimeout: fastTimeout,
		slowTimeout: slowTimeout,
	}, nil
}

// CodeAt returns the code of the given account. This is needed to differentiate
// between contract internal errors and the local chain being out of sync.
func (m *StandardRpcClient) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "CodeAt")
	defer cancel()
	return m.client.CodeAt(ctx, contract, blockNumber)
}

// CallContract executes an Ethereum contract call with the specified data as the
// input.
func (m *StandardRpcClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "CallContract")
	defer cancel()
	return m.client.CallContract(ctx, call, blockNumber)
}

// HeaderByHash returns the block header with the given hash.
func (m *StandardRpcClient) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "HeaderByHash")
	defer cancel()
	return m.client.HeaderByHash(ctx, hash)
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (m *StandardRpcClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "HeaderByNumber")
	defer cancel()
	return m.client.HeaderByNumber(ctx, number)
}

// PendingCodeAt returns the code of the given account in the pending state.
func (m *StandardRpcClient) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "PendingCodeAt")
	defer cancel()
	return m.client.PendingCodeAt(ctx, account)
}

// PendingNonceAt retrieves the current pending nonce associated with an account.
func (m *StandardRpcClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "PendingNonceAt")
	defer cancel()
	return m.client.PendingNonceAt(ctx, account)
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (m *StandardRpcClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "SuggestGasPrice")
	defer cancel()
	return m.client.SuggestGasPrice(ctx)
}

// SuggestGasTipCap retrieves the currently suggested 1559 priority fee to allow
// a timely execution of a transaction.
func (m *StandardRpcClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "SuggestGasTipCap")
	defer cancel()
	return m.client.SuggestGasTipCap(ctx)
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
// There is no guarantee that this is the true gas limit requirement as other
// transactions may be added or removed by miners, but it should provide a basis
// for setting a reasonable default.
func (m *StandardRpcClient) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "EstimateGas")
	defer cancel()
	return m.client.EstimateGas(ctx, call)
}

// SendTransaction injects the transaction into the pending pool for execution.
func (m *StandardRpcClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "SendTransaction")
	defer cancel()
	return m.client.SendTransaction(ctx, tx)
}

// FilterLogs executes a log filter operation, blocking during execution and
// returning all the results in one batch.
func (m *StandardRpcClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.slowTimeout, "FilterLogs")
	defer cancel()
	return m.client.FilterLogs(ctx, query)
}

// SubscribeFilterLogs creates a background log filtering operation, returning
// a subscription immediately, which can be used to stream the found events.
func (m *StandardRpcClient) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "SubscribeFilterLogs")
	defer cancel()
	return m.client.SubscribeFilterLogs(ctx, query, ch)
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (m *StandardRpcClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "TransactionReceipt")
	defer cancel()
	return m.client.TransactionReceipt(ctx, txHash)
}

// BlockNumber returns the most recent block number
func (m *StandardRpcClient) BlockNumber(ctx context.Context) (uint64, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "BlockNumber")
	defer cancel()
	return m.client.BlockNumber(ctx)
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (m *StandardRpcClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "BalanceAt")
	defer cancel()
	return m.client.BalanceAt(ctx, account, blockNumber)
}

// TransactionByHash returns the transaction with the given hash.
func (m *StandardRpcClient) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "TransactionByHash")
	defer cancel()
	return m.client.TransactionByHash(ctx, hash)
}

// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (m *StandardRpcClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "NonceAt")
	defer cancel()
	return m.client.NonceAt(ctx, account, blockNumber)
}

// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (m *StandardRpcClient) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "SyncProgress")
	defer cancel()
	return m.client.SyncProgress(ctx)
}

func (m *StandardRpcClient) ChainID(ctx context.Context) (*big.Int, error) {
	ctx, cancel := logRequestAndCreateContext(ctx, m.fastTimeout, "ChainID")
	defer cancel()
	return m.client.ChainID(ctx)
}

/// ========================
/// == Internal Functions ==
/// ========================

// Logs the request and returns a context with the provided timeout and HTTP tracing enabled if requested
func logRequestAndCreateContext(ctx context.Context, timeout time.Duration, methodName string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	logger, _ := log.FromContext(ctx)
	if logger != nil {
		logger.Debug("Running EC request",
			slog.String(log.MethodKey, methodName),
		)
		tracer := logger.GetHttpTracer()
		if tracer != nil {
			ctx = httptrace.WithClientTrace(ctx, tracer)
		}
	}
	return ctx, cancel
}

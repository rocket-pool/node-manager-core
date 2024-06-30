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

const (
	// Default timeout for fast requests
	DefaultFastTimeout = 5 * time.Second

	// Default timeout for slow requests
	DefaultSlowTimeout = 30 * time.Second
)

// Options for creating a new StandardRpcClient
type StandardRpcClientOptions struct {
	// Timeout to use for requests that should return quickly
	FastTimeout time.Duration

	// Timeout to use for requests that are expected to take longer to process
	SlowTimeout time.Duration
}

// Standard RPC-based Execution Client binding with logging support, using Geth as the backing client implementation.
type StandardRpcClient struct {
	client             *ethclient.Client
	defaultFastTimeout time.Duration
	defaultSlowTimeout time.Duration
}

// Creates a new StandardRpcClient instance
func NewStandardRpcClient(address string, opts *StandardRpcClientOptions) (*StandardRpcClient, error) {
	client, err := ethclient.Dial(address)
	if err != nil {
		return nil, fmt.Errorf("error creating EC binding for [%s]: %w", address, err)
	}
	wrapper := &StandardRpcClient{
		client: client,
	}
	if opts != nil {
		wrapper.defaultFastTimeout = opts.FastTimeout
		wrapper.defaultSlowTimeout = opts.SlowTimeout
	} else {
		wrapper.defaultFastTimeout = DefaultFastTimeout
		wrapper.defaultSlowTimeout = DefaultSlowTimeout
	}
	return wrapper, nil
}

// CodeAt returns the code of the given account. This is needed to differentiate
// between contract internal errors and the local chain being out of sync.
func (c *StandardRpcClient) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "CodeAt")
	return c.client.CodeAt(ctx, contract, blockNumber)
}

// CallContract executes an Ethereum contract call with the specified data as the
// input.
func (c *StandardRpcClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "CallContract")
	return c.client.CallContract(ctx, call, blockNumber)
}

// HeaderByHash returns the block header with the given hash.
func (c *StandardRpcClient) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "HeaderByHash")
	return c.client.HeaderByHash(ctx, hash)
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (c *StandardRpcClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "HeaderByNumber")
	return c.client.HeaderByNumber(ctx, number)
}

// PendingCodeAt returns the code of the given account in the pending state.
func (c *StandardRpcClient) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "PendingCodeAt")
	return c.client.PendingCodeAt(ctx, account)
}

// PendingNonceAt retrieves the current pending nonce associated with an account.
func (c *StandardRpcClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "PendingNonceAt")
	return c.client.PendingNonceAt(ctx, account)
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (c *StandardRpcClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "SuggestGasPrice")
	return c.client.SuggestGasPrice(ctx)
}

// SuggestGasTipCap retrieves the currently suggested 1559 priority fee to allow
// a timely execution of a transaction.
func (c *StandardRpcClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "SuggestGasTipCap")
	return c.client.SuggestGasTipCap(ctx)
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
// There is no guarantee that this is the true gas limit requirement as other
// transactions may be added or removed by miners, but it should provide a basis
// for setting a reasonable default.
func (c *StandardRpcClient) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "EstimateGas")
	return c.client.EstimateGas(ctx, call)
}

// SendTransaction injects the transaction into the pending pool for execution.
func (c *StandardRpcClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "SendTransaction")
	return c.client.SendTransaction(ctx, tx)
}

// FilterLogs executes a log filter operation, blocking during execution and
// returning all the results in one batch.
func (c *StandardRpcClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultSlowTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "FilterLogs")
	return c.client.FilterLogs(ctx, query)
}

// SubscribeFilterLogs creates a background log filtering operation, returning
// a subscription immediately, which can be used to stream the found events.
func (c *StandardRpcClient) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "SubscribeFilterLogs")
	return c.client.SubscribeFilterLogs(ctx, query, ch)
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (c *StandardRpcClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "TransactionReceipt")
	return c.client.TransactionReceipt(ctx, txHash)
}

// BlockNumber returns the most recent block number
func (c *StandardRpcClient) BlockNumber(ctx context.Context) (uint64, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "BlockNumber")
	return c.client.BlockNumber(ctx)
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (c *StandardRpcClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "BalanceAt")
	return c.client.BalanceAt(ctx, account, blockNumber)
}

// TransactionByHash returns the transaction with the given hash.
func (c *StandardRpcClient) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "TransactionByHash")
	return c.client.TransactionByHash(ctx, hash)
}

// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (c *StandardRpcClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "NonceAt")
	return c.client.NonceAt(ctx, account, blockNumber)
}

// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (c *StandardRpcClient) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "SyncProgress")
	return c.client.SyncProgress(ctx)
}

func (c *StandardRpcClient) ChainID(ctx context.Context) (*big.Int, error) {
	// Prep the context
	ctx, cancel := c.prepareContext(ctx, c.defaultFastTimeout)
	defer cancel()

	ctx = c.logRequest(ctx, "ChainID")
	return c.client.ChainID(ctx)
}

/// ========================
/// == Internal Functions ==
/// ========================

// Adds a timeout to the context if one didn't already exist
func (c *StandardRpcClient) prepareContext(ctx context.Context, defaultTimeout time.Duration) (context.Context, context.CancelFunc) {
	// Make a new context if it wasn't provided
	if ctx == nil {
		ctx = context.Background()
	}

	// Return if there was already a deadline
	_, hasDeadline := ctx.Deadline()
	if hasDeadline {
		return ctx, func() {}
	}

	// Add a default timeout if there isn't one
	return context.WithTimeout(ctx, defaultTimeout)
}

// Logs the request and returns a context with the provided timeout and HTTP tracing enabled if requested
func (c *StandardRpcClient) logRequest(ctx context.Context, methodName string) context.Context {
	logger, _ := log.FromContext(ctx)
	if logger == nil {
		return ctx
	}

	// Log the request
	args := []any{
		slog.String(log.MethodKey, methodName),
	}
	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		args = append(args, slog.Time("deadline", deadline.UTC()))
	}
	logger.Debug("Running EC request", args...)
	tracer := logger.GetHttpTracer()
	if tracer != nil {
		// Enable HTTP tracing if requested
		ctx = httptrace.WithClientTrace(ctx, tracer)
	}
	return ctx
}

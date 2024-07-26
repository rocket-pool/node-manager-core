package eth

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/sync/errgroup"
)

const (
	// The block gas limit (gwei)
	GasLimit uint64 = 30000000

	// Default value for the safe gas limit buffer, in gwei
	DefaultSafeGasBuffer uint64 = 0

	// Default value for the safe gas limit multiplier
	DefaultSafeGasMultiplier float64 = 1.5

	// Prefix for errors caused by gas estimation
	gasSimErrorPrefix string = "error estimating gas needed"
)

// A simple calculator to bolster gas estimates to safe values, checking against the Ethereum gas block limit.
type TransactionManager struct {
	// Gwei ammount added to estimated gas limits, as a safety buffer
	buffer uint64

	// Estimated gas limits are multiplied by this value as part of safety buffer calculation
	multiplier float64

	// The client to use for running transaction simulations
	client IExecutionClient
}

// Creates a new transaction manager, which can simulate and execute transactions.
// The simulator determines if transactions will complete without reversion and provides a safe gas limit suggestion.
// The formula for safe gas calculation is estimate * multiplier + buffer, where the buffer is in gwei.
func NewTransactionManager(client IExecutionClient, safeGasBuffer uint64, safeGasMultiplier float64) (*TransactionManager, error) {
	if safeGasMultiplier != 0 && safeGasMultiplier < 1 {
		return nil, fmt.Errorf("multiplier cannot be less than 1")
	}

	return &TransactionManager{
		client:     client,
		buffer:     safeGasBuffer,
		multiplier: safeGasMultiplier,
	}, nil
}

// ==================
// === Simulation ===
// ==================

// Calculates a gas limit for a gas estimate with the provided safety buffer: estimate * multiplier + buffer.
// Returns an error if the calculate safe gas limit is higher than the Ethereum block limit.
func (t *TransactionManager) GetSafeGasLimit(estimate uint64) (uint64, error) {
	if estimate > GasLimit {
		return 0, fmt.Errorf("estimated gas usage of %d is greater than the block gas limit of %d", estimate, GasLimit)
	}

	safeLimit := uint64(math.Ceil(float64(estimate)*t.multiplier)) + t.buffer
	if safeLimit > GasLimit {
		return 0, fmt.Errorf("safe gas limit of %d is greater than the block gas limit of %d", safeLimit, GasLimit)
	}
	return safeLimit, nil
}

// Simulates the transaction, getting the expected and safe gas limits in gwei.
func (t *TransactionManager) SimulateTransaction(client IExecutionClient, to common.Address, opts *bind.TransactOpts, input []byte) SimulationResult {
	// Handle requests without opts
	if opts == nil {
		return SimulationResult{
			IsSimulated:       false,
			EstimatedGasLimit: 0,
			SafeGasLimit:      0,
			SimulationError:   "",
		}
	}

	// Estimate gas limit
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:      opts.From,
		To:        &to,
		GasFeeCap: big.NewInt(0),
		GasTipCap: big.NewInt(0),
		Value:     opts.Value,
		Data:      input,
	})
	if err != nil {
		return SimulationResult{
			IsSimulated:       true,
			EstimatedGasLimit: 0,
			SafeGasLimit:      0,
			SimulationError:   fmt.Sprintf("%s: %s", gasSimErrorPrefix, normalizeRevertMessage(err).Error())}
	}

	// Get a safe gas limit
	safeLimit, err := t.GetSafeGasLimit(gasLimit)
	if err != nil {
		return SimulationResult{
			IsSimulated:       true,
			EstimatedGasLimit: 0,
			SafeGasLimit:      0,
			SimulationError:   fmt.Sprintf("error estimating gas limit: %s", err.Error()),
		}
	}
	return SimulationResult{
		IsSimulated:       true,
		EstimatedGasLimit: gasLimit,
		SafeGasLimit:      safeLimit,
		SimulationError:   "",
	}
}

// ===================
// === TX Creation ===
// ===================

// Create a new TransactionInfo  binding for a contract method and simulate its execution
func (t *TransactionManager) CreateTransactionInfo(contract *Contract, method string, opts *bind.TransactOpts, parameters ...any) (*TransactionInfo, error) {
	// Create the data data
	data, err := contract.ABI.Pack(method, parameters...)
	if err != nil {
		return nil, fmt.Errorf("error packing input data: %w", err)
	}

	// Simulate the TX
	simResult := t.SimulateTransaction(t.client, contract.Address, opts, data)

	// Create the info wrapper
	var value *big.Int
	if opts != nil {
		value = opts.Value
	}
	txInfo := &TransactionInfo{
		Data:             data,
		To:               contract.Address,
		Value:            value,
		SimulationResult: simResult,
	}
	return txInfo, nil
}

// Create a new serializable TransactionInfo from raw data and simulate its execution
func (t *TransactionManager) CreateTransactionInfoRaw(to common.Address, data []byte, opts *bind.TransactOpts) *TransactionInfo {
	// Simulate the TX
	simResult := t.SimulateTransaction(t.client, to, opts, data)

	// Create the info wrapper
	var value *big.Int
	if opts != nil {
		value = opts.Value
	}
	txInfo := &TransactionInfo{
		Data:             data,
		To:               to,
		Value:            value,
		SimulationResult: simResult,
	}
	return txInfo
}

// =================
// === Execution ===
// =================

// Signs a transaction but does not submit it to the network. Use this if you want to sign something offline and submit it later,
// or submit it as part of a bundle.
func (t *TransactionManager) SignTransaction(txInfo *TransactionInfo, opts *bind.TransactOpts) (*types.Transaction, error) {
	opts.NoSend = true
	return t.ExecuteTransactionRaw(txInfo.To, txInfo.Data, txInfo.Value, opts)
}

// Signs and submits a transaction to the network.
// The nonce and gas fee info in the provided opts will be used.
// The value will come from the provided txInfo. It will *not* use the value in the provided opts.
func (t *TransactionManager) ExecuteTransaction(txInfo *TransactionInfo, opts *bind.TransactOpts) (*types.Transaction, error) {
	return t.ExecuteTransactionRaw(txInfo.To, txInfo.Data, txInfo.Value, opts)
}

// Create a transaction from serialized info, signs it, and submits it to the network if requested in opts.
// Note the value in opts is not used; set it in the value argument instead.
func (t *TransactionManager) ExecuteTransactionRaw(to common.Address, data []byte, value *big.Int, opts *bind.TransactOpts) (*types.Transaction, error) {
	// Create a "dummy" contract for the Geth API with no ABI since we don't need it for this
	contract := bind.NewBoundContract(to, abi.ABI{}, t.client, t.client, t.client)

	newOpts := &bind.TransactOpts{
		// Copy the original fields
		From:      opts.From,
		Nonce:     opts.Nonce,
		Signer:    opts.Signer,
		GasPrice:  opts.GasPrice,
		GasFeeCap: opts.GasFeeCap,
		GasTipCap: opts.GasTipCap,
		GasLimit:  opts.GasLimit,
		Context:   opts.Context,
		NoSend:    opts.NoSend,

		// Overwrite the value
		Value: value,
	}

	return contract.RawTransact(newOpts, data)
}

// Signs and submits a bundle of transactions to the network that are all sent from the same address.
// The values for each TX will be in each TX info; the value specified in the opts argument is not used.
// The GasFeeCap and GasTipCap from opts will be used for all transactions.
// NOTE: this assumes the bundle is meant to be submitted sequentially, so the nonce of each one will be incremented.
// Assign the Nonce in the opts tto the nonce you want to use for the first transaction.
func (t *TransactionManager) BatchExecuteTransactions(txSubmissions []*TransactionSubmission, opts *bind.TransactOpts) ([]*types.Transaction, error) {
	if opts.Nonce == nil {
		// Get the latest nonce and use that as the nonce for the first TX
		nonce, err := t.client.NonceAt(context.Background(), opts.From, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting latest nonce for node: %w", err)
		}
		opts.Nonce = big.NewInt(0).SetUint64(nonce)
	}

	txs := make([]*types.Transaction, len(txSubmissions))
	for i, txSubmission := range txSubmissions {
		txInfo := txSubmission.TxInfo
		opts.GasLimit = txSubmission.GasLimit
		tx, err := t.ExecuteTransactionRaw(txInfo.To, txInfo.Data, txInfo.Value, opts)
		if err != nil {
			return nil, fmt.Errorf("error creating transaction %d in bundle: %w", i, err)
		}
		txs[i] = tx

		// Increment the nonce for the next TX
		opts.Nonce.Add(opts.Nonce, common.Big1)
	}
	return txs, nil
}

// ===============
// === Waiting ===
// ===============

// Wait for a transaction to get included in blocks
func (t *TransactionManager) WaitForTransaction(tx *types.Transaction) error {
	// Wait for transaction to be included
	txReceipt, err := bind.WaitMined(context.Background(), t.client, tx)
	if err != nil {
		return fmt.Errorf("error running transaction %s: %w", tx.Hash().Hex(), err)
	}

	// Check transaction status
	if txReceipt.Status == 0 {
		return fmt.Errorf("transaction %s failed with status 0", tx.Hash().Hex())
	}

	// Return
	return nil
}

// Wait for a set of transactions to get included in blocks
func (t *TransactionManager) WaitForTransactions(txs []*types.Transaction) error {
	var wg errgroup.Group
	for _, tx := range txs {
		tx := tx
		wg.Go(func() error {
			return t.WaitForTransaction(tx)
		})
	}

	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("error waiting for transactions: %w", err)
	}

	return nil
}

// Wait for a transaction to get included in blocks
func (t *TransactionManager) WaitForTransactionByHash(hash common.Hash) error {
	// Get the TX
	tx, err := t.getTransactionFromHash(hash)
	if err != nil {
		return fmt.Errorf("error getting transaction %s: %w", hash.Hex(), err)
	}

	// Wait for transaction to be included
	return t.WaitForTransaction(tx)
}

// Wait for a set of transactions to get included in blocks
func (t *TransactionManager) WaitForTransactionsByHash(hashes []common.Hash) error {
	var wg errgroup.Group

	// Get the TXs from the hashes
	for _, hash := range hashes {
		hash := hash
		wg.Go(func() error {
			return t.WaitForTransactionByHash(hash)
		})
	}
	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("error waiting for transactions: %w", err)
	}

	// Wait for the TXs
	return nil
}

// Get a TX from its hash
func (t *TransactionManager) getTransactionFromHash(hash common.Hash) (*types.Transaction, error) {
	// Retry for 30 sec if the TX wasn't found
	for i := 0; i < 30; i++ {
		tx, _, err := t.client.TransactionByHash(context.Background(), hash)
		if err != nil {
			if err.Error() == "not found" {
				time.Sleep(1 * time.Second)
				continue
			}
			return nil, err
		}

		return tx, nil
	}

	return nil, fmt.Errorf("transaction not found after 30 seconds")
}

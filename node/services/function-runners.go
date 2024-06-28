package services

import (
	"context"
	"fmt"

	"github.com/rocket-pool/node-manager-core/log"
)

// This is a signature for a wrapped function that only returns an error
type function0[ClientType any] func(ClientType) error

// This is a signature for a wrapped function that returns 1 var and an error
type function1[ClientType any, ReturnType any] func(ClientType) (ReturnType, error)

// This is a signature for a wrapped function that returns 2 vars and an error
type function2[ClientType any, ReturnType1 any, ReturnType2 any] func(ClientType) (ReturnType1, ReturnType2, error)

// Attempts to run a function progressively through each client until one succeeds or they all fail.
// Expects functions with 1 output and an error; for functions with other signatures, see the other runFunctionX functions.
func runFunction1[ClientType any, ReturnType any](m iClientManagerImpl[ClientType], ctx context.Context, function function1[ClientType, ReturnType]) (ReturnType, error) {
	// If there's no fallback, just run the function on the primary
	if !m.IsFallbackEnabled() {
		return function(m.GetPrimaryClient())
	}

	var blank ReturnType
	logger, _ := log.FromContext(ctx)
	typeName := m.GetClientTypeName()

	// Check the clients for recovery
	m.RecheckFailTimes(logger)

	// Check if we can use the primary
	if m.IsPrimaryReady() {
		// Try to run the function on the primary
		result, err := function(m.GetPrimaryClient())
		if err == nil {
			// If there's no error, return the result
			return result, nil
		}

		// If it's not a disconnect error, just return it
		if !isDisconnected(err) {
			return blank, err
		}

		// Log the disconnect and try the fallback if available
		m.SetPrimaryReady(false)
		if logger != nil {
			logger.Warn("Primary "+typeName+" client disconnected, using fallback...", log.Err(err))
		}
		return runFunction1[ClientType, ReturnType](m, ctx, function)
	}

	// Check if we can use the fallback
	if m.IsFallbackReady() {
		// Try to run the function on the fallback
		result, err := function(m.GetFallbackClient())
		if err == nil {
			// If there's no error, return the result
			return result, nil
		}

		// If it's not a disconnect error, just return it
		if !isDisconnected(err) {
			return blank, err
		}

		// If Log the disconnect and return an error
		if logger != nil {
			logger.Warn("Fallback "+typeName+" disconnected", log.Err(err))
		}
		m.SetFallbackReady(false)
		return blank, fmt.Errorf("all " + typeName + "s failed")
	}

	// If neither client is ready, just run the primary
	if logger != nil {
		logger.Warn("No " + typeName + "s are ready, forcing use of primary...")
	}
	return function(m.GetPrimaryClient())
}

// Run a function with 0 outputs and an error
func runFunction0[ClientType any](m iClientManagerImpl[ClientType], ctx context.Context, function function0[ClientType]) error {
	_, err := runFunction1(m, ctx, func(client ClientType) (any, error) {
		return nil, function(client)
	})
	return err
}

// Run a function with 2 outputs and an error
func runFunction2[ClientType any, ReturnType1 any, ReturnType2 any](m iClientManagerImpl[ClientType], ctx context.Context, function function2[ClientType, ReturnType1, ReturnType2]) (ReturnType1, ReturnType2, error) {
	type out struct {
		arg1 ReturnType1
		arg2 ReturnType2
	}
	result, err := runFunction1(m, ctx, func(client ClientType) (out, error) {
		arg1, arg2, err := function(client)
		return out{
			arg1: arg1,
			arg2: arg2,
		}, err
	})
	return result.arg1, result.arg2, err
}

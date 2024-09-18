package services

import (
	"context"
	"errors"

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
	logger, _ := log.FromContext(ctx)
	var blank ReturnType
	typeName := m.GetClientTypeName()

	// Check if we can use the primary
	if m.IsPrimaryReady() {
		// Try to run the function on the primary
		result, err := function(m.GetPrimaryClient())
		if err != nil {
			if isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				m.SetPrimaryReady(false)
				if m.IsFallbackEnabled() {
					logger.Warn("Primary "+typeName+" client disconnected, using fallback...", log.Err(err))
					return runFunction1[ClientType, ReturnType](m, ctx, function)
				} else {
					logger.Warn("Primary "+typeName+" disconnected and no fallback is configured.", log.Err(err))
					return blank, errors.New("all " + typeName + "s failed")
				}
			}
			// If it's a different error, just return it
			return blank, err
		}
		// If there's no error, return the result
		return result, nil
	}

	if m.IsFallbackReady() {
		// Try to run the function on the fallback
		result, err := function(m.GetFallbackClient())
		if err != nil {
			if isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				logger.Warn("Fallback "+typeName+" disconnected", log.Err(err))
				m.SetFallbackReady(false)
				return blank, errors.New("all " + typeName + "s failed")
			}

			// If it's a different error, just return it
			return blank, err
		}
		// If there's no error, return the result
		return result, nil
	}

	return blank, errors.New("no " + typeName + "s were ready")
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

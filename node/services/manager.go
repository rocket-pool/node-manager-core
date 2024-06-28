package services

import "github.com/rocket-pool/node-manager-core/log"

type IClientManager[ClientType any] interface {
	GetPrimaryClient() ClientType
	GetFallbackClient() ClientType
	IsPrimaryReady() bool
	IsFallbackReady() bool
	IsFallbackEnabled() bool
	GetClientTypeName() string
}

type iClientManagerImpl[ClientType any] interface {
	IClientManager[ClientType]

	// Internal functions
	SetPrimaryReady(bool)
	SetFallbackReady(bool)
	RecheckFailTimes(logger *log.Logger)
}

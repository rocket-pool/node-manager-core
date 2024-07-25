package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"time"

	dclient "github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
)

const (
	DockerApiVersion string = "1.40"
)

// ==================
// === Interfaces ===
// ==================

// Provides access to Ethereum client(s) via a fallback-enabled manager, along with utilities for querying the chain and executing transactions
type IEthClientProvider interface {
	// Gets the Execution Client manager
	GetEthClient() *ExecutionClientManager

	// Gets the Execution layer query manager
	GetQueryManager() *eth.QueryManager

	// Gets the Execution layer transaction manager
	GetTransactionManager() *eth.TransactionManager
}

// Provides access to Beacon client(s) via a fallback-enabled manager
type IBeaconClientProvider interface {
	// Gets the Beacon Client manager
	GetBeaconClient() *BeaconClientManager
}

// Provides access to a Docker client
type IDockerProvider interface {
	// Gets the Docker client
	GetDocker() dclient.APIClient
}

// Provides access to the node's loggers
type ILoggerProvider interface {
	// Gets the logger to use for the API server
	GetApiLogger() *log.Logger

	// Gets the logger to use for the automated tasks loop
	GetTasksLogger() *log.Logger
}

// Provides access to the node's wallet
type IWalletProvider interface {
	// Gets the node's wallet
	GetWallet() *wallet.Wallet
}

// Provides access to a context for cancelling long operations upon daemon shutdown
type IContextProvider interface {
	// Gets a base context for the daemon that all operations can derive from
	GetBaseContext() context.Context

	// Cancels the base context when the daemon is shutting down
	CancelContextOnShutdown()
}

// A container for all of the various services used by the node daemon
type IServiceProvider interface {
	IEthClientProvider
	IBeaconClientProvider
	IDockerProvider
	ILoggerProvider
	IWalletProvider
	IContextProvider
	io.Closer
}

// =======================
// === ServiceProvider ===
// =======================

// A container for all of the various services used by the node service
type serviceProvider struct {
	// Services
	nodeWallet *wallet.Wallet
	ecManager  *ExecutionClientManager
	bcManager  *BeaconClientManager
	docker     dclient.APIClient
	txMgr      *eth.TransactionManager
	queryMgr   *eth.QueryManager

	// Context for cancelling long operations
	ctx    context.Context
	cancel context.CancelFunc

	// Logging
	apiLogger   *log.Logger
	tasksLogger *log.Logger
}

// Creates a new ServiceProvider instance based on the given config
func NewServiceProvider(cfg config.IConfig, resources *config.NetworkResources, clientTimeout time.Duration) (IServiceProvider, error) {
	// EC Manager
	var ecManager *ExecutionClientManager
	primaryEcUrl, fallbackEcUrl := cfg.GetExecutionClientUrls()
	primaryEc, err := ethclient.Dial(primaryEcUrl)
	if err != nil {
		return nil, fmt.Errorf("error connecting to primary EC at [%s]: %w", primaryEcUrl, err)
	}
	if fallbackEcUrl != "" {
		// Get the fallback EC url, if applicable
		fallbackEc, err := ethclient.Dial(fallbackEcUrl)
		if err != nil {
			return nil, fmt.Errorf("error connecting to fallback EC at [%s]: %w", fallbackEcUrl, err)
		}
		ecManager = NewExecutionClientManagerWithFallback(primaryEc, fallbackEc, resources.ChainID, clientTimeout)
	} else {
		ecManager = NewExecutionClientManager(primaryEc, resources.ChainID, clientTimeout)
	}

	// Beacon manager
	var bcManager *BeaconClientManager
	primaryBnUrl, fallbackBnUrl := cfg.GetBeaconNodeUrls()
	primaryBc := client.NewStandardHttpClient(primaryBnUrl, clientTimeout)
	if fallbackBnUrl != "" {
		fallbackBc := client.NewStandardHttpClient(fallbackBnUrl, clientTimeout)
		bcManager = NewBeaconClientManagerWithFallback(primaryBc, fallbackBc, resources.ChainID, clientTimeout)
	} else {
		bcManager = NewBeaconClientManager(primaryBc, resources.ChainID, clientTimeout)
	}

	// Docker client
	dockerClient, err := dclient.NewClientWithOpts(dclient.WithVersion(DockerApiVersion))
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}

	return NewServiceProviderWithCustomServices(cfg, resources, ecManager, bcManager, dockerClient)
}

// Creates a new ServiceProvider instance with custom services instead of creating them from the config
func NewServiceProviderWithCustomServices(cfg config.IConfig, resources *config.NetworkResources, ecManager *ExecutionClientManager, bcManager *BeaconClientManager, dockerClient dclient.APIClient) (IServiceProvider, error) {
	// Make the API logger
	loggerOpts := cfg.GetLoggerOptions()
	apiLogger, err := log.NewLogger(cfg.GetApiLogFilePath(), loggerOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating API logger: %w", err)
	}

	// Make the tasks logger
	tasksLogger, err := log.NewLogger(cfg.GetTasksLogFilePath(), loggerOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating tasks logger: %w", err)
	}

	// Wallet
	nodeAddressPath := filepath.Join(cfg.GetNodeAddressFilePath())
	walletDataPath := filepath.Join(cfg.GetWalletFilePath())
	passwordPath := filepath.Join(cfg.GetPasswordFilePath())
	nodeWallet, err := wallet.NewWallet(tasksLogger.Logger, walletDataPath, nodeAddressPath, passwordPath, resources.ChainID)
	if err != nil {
		return nil, fmt.Errorf("error creating node wallet: %w", err)
	}

	// TX Manager
	txMgr, err := eth.NewTransactionManager(ecManager, eth.DefaultSafeGasBuffer, eth.DefaultSafeGasMultiplier)
	if err != nil {
		return nil, fmt.Errorf("error creating transaction manager: %w", err)
	}

	// Query Manager - set the default concurrent run limit to half the CPUs so the EC doesn't get overwhelmed
	concurrentCallLimit := runtime.NumCPU() / 2
	if concurrentCallLimit < 1 {
		concurrentCallLimit = 1
	}
	queryMgr := eth.NewQueryManager(ecManager, resources.MulticallAddress, concurrentCallLimit)

	// Context for handling task cancellation during shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Log startup
	apiLogger.Info("Starting API logger.")
	tasksLogger.Info("Starting Tasks logger.")

	// Create the provider
	provider := &serviceProvider{
		nodeWallet:  nodeWallet,
		ecManager:   ecManager,
		bcManager:   bcManager,
		docker:      dockerClient,
		txMgr:       txMgr,
		queryMgr:    queryMgr,
		ctx:         ctx,
		cancel:      cancel,
		apiLogger:   apiLogger,
		tasksLogger: tasksLogger,
	}
	return provider, nil
}

// Closes the service provider and its underlying services
func (p *serviceProvider) Close() error {
	p.apiLogger.Close()
	p.tasksLogger.Close()
	return nil
}

// ===============
// === Getters ===
// ===============

func (p *serviceProvider) GetWallet() *wallet.Wallet {
	return p.nodeWallet
}

func (p *serviceProvider) GetEthClient() *ExecutionClientManager {
	return p.ecManager
}

func (p *serviceProvider) GetBeaconClient() *BeaconClientManager {
	return p.bcManager
}

func (p *serviceProvider) GetDocker() dclient.APIClient {
	return p.docker
}

func (p *serviceProvider) GetTransactionManager() *eth.TransactionManager {
	return p.txMgr
}

func (p *serviceProvider) GetQueryManager() *eth.QueryManager {
	return p.queryMgr
}

func (p *serviceProvider) GetApiLogger() *log.Logger {
	return p.apiLogger
}

func (p *serviceProvider) GetTasksLogger() *log.Logger {
	return p.tasksLogger
}

func (p *serviceProvider) GetBaseContext() context.Context {
	return p.ctx
}

func (p *serviceProvider) CancelContextOnShutdown() {
	p.cancel()
}

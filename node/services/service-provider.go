package services

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	dclient "github.com/docker/docker/client"
	"github.com/rocket-pool/node-manager-core/beacon/client"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
)

const (
	DockerApiVersion string = "1.40"
)

// A container for all of the various services used by the node service
type ServiceProvider struct {
	// Services
	cfg        config.IConfig
	resources  *config.NetworkResources
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
func NewServiceProvider(cfg config.IConfig) (*ServiceProvider, error) {
	resources := cfg.GetNetworkResources()

	// EC Manager
	var ecManager *ExecutionClientManager
	primaryEcUrl, fallbackEcUrl := cfg.GetExecutionClientUrls()
	timeouts := cfg.GetExecutionClientTimeouts()
	primaryEc, err := eth.NewStandardRpcClient(primaryEcUrl, timeouts.FastTimeout, timeouts.SlowTimeout)
	if err != nil {
		return nil, fmt.Errorf("error connecting to primary EC at [%s]: %w", primaryEcUrl, err)
	}
	if fallbackEcUrl != "" {
		// Get the fallback EC url, if applicable
		fallbackEc, err := eth.NewStandardRpcClient(fallbackEcUrl, timeouts.FastTimeout, timeouts.SlowTimeout)
		if err != nil {
			return nil, fmt.Errorf("error connecting to fallback EC at [%s]: %w", fallbackEcUrl, err)
		}
		ecManager = NewExecutionClientManagerWithFallback(primaryEc, fallbackEc, resources.ChainID, timeouts.RecheckDelay)
	} else {
		ecManager = NewExecutionClientManager(primaryEc, resources.ChainID)
	}

	// Beacon manager
	var bcManager *BeaconClientManager
	primaryBnUrl, fallbackBnUrl := cfg.GetBeaconNodeUrls()
	timeouts = cfg.GetBeaconNodeTimeouts()
	opts := &client.StandardHttpClientOpts{
		FastTimeout: timeouts.FastTimeout,
		SlowTimeout: timeouts.SlowTimeout,
	}
	primaryBc, err := client.NewStandardHttpClient(primaryBnUrl, opts)
	if err != nil {
		return nil, fmt.Errorf("error connecting to primary BC at [%s]: %w", primaryBnUrl, err)
	}
	if fallbackBnUrl != "" {
		fallbackBc, err := client.NewStandardHttpClient(fallbackBnUrl, opts)
		if err != nil {
			return nil, fmt.Errorf("error connecting to fallback BC at [%s]: %w", fallbackBnUrl, err)
		}
		bcManager = NewBeaconClientManagerWithFallback(primaryBc, fallbackBc, resources.ChainID, timeouts.RecheckDelay)
	} else {
		bcManager = NewBeaconClientManager(primaryBc, resources.ChainID)
	}

	// Docker client
	dockerClient, err := dclient.NewClientWithOpts(dclient.WithVersion(DockerApiVersion))
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %w", err)
	}

	return NewServiceProviderWithCustomServices(cfg, resources, ecManager, bcManager, dockerClient)
}

// Creates a new ServiceProvider instance with custom services instead of creating them from the config
func NewServiceProviderWithCustomServices(cfg config.IConfig, resources *config.NetworkResources, ecManager *ExecutionClientManager, bcManager *BeaconClientManager, dockerClient dclient.APIClient) (*ServiceProvider, error) {
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
	provider := &ServiceProvider{
		cfg:         cfg,
		resources:   resources,
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
func (p *ServiceProvider) Close() {
	p.apiLogger.Close()
	p.tasksLogger.Close()
}

// ===============
// === Getters ===
// ===============

func (p *ServiceProvider) GetConfig() config.IConfig {
	return p.cfg
}

func (p *ServiceProvider) GetNetworkResources() *config.NetworkResources {
	return p.resources
}

func (p *ServiceProvider) GetWallet() *wallet.Wallet {
	return p.nodeWallet
}

func (p *ServiceProvider) GetEthClient() *ExecutionClientManager {
	return p.ecManager
}

func (p *ServiceProvider) GetBeaconClient() *BeaconClientManager {
	return p.bcManager
}

func (p *ServiceProvider) GetDocker() dclient.APIClient {
	return p.docker
}

func (p *ServiceProvider) GetTransactionManager() *eth.TransactionManager {
	return p.txMgr
}

func (p *ServiceProvider) GetQueryManager() *eth.QueryManager {
	return p.queryMgr
}

func (p *ServiceProvider) GetApiLogger() *log.Logger {
	return p.apiLogger
}

func (p *ServiceProvider) GetTasksLogger() *log.Logger {
	return p.tasksLogger
}

func (p *ServiceProvider) GetBaseContext() context.Context {
	return p.ctx
}

func (p *ServiceProvider) CancelContextOnShutdown() {
	p.cancel()
}

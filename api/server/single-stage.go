package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/goccy/go-json"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/services"
	"github.com/rocket-pool/node-manager-core/utils"
)

// Wrapper for callbacks used by call runners that follow a common single-stage pattern:
// Create bindings, query the chain, and then do whatever else they want.
// Structs implementing this will handle the caller-specific functionality.
type ISingleStageCallContext[DataType any] interface {
	// Initialize the context with any bootstrapping, requirements checks, or bindings it needs to set up
	Initialize() (types.ResponseStatus, error)

	// Used to get any supplemental state required during initialization - anything in here will be fed into an hd.Query() multicall
	GetState(mc *batch.MultiCaller) error

	// Prepare the response data in whatever way the context needs to do
	PrepareData(data *DataType, opts *bind.TransactOpts) (types.ResponseStatus, error)
}

// Interface for single-stage call context factories - these will be invoked during route handling to create the
// unique context for the route
type ISingleStageGetContextFactory[ContextType ISingleStageCallContext[DataType], DataType any] interface {
	// Create the context for the route
	Create(args url.Values) (ContextType, error)
}

// Interface for queryless call context factories that handle POST requests.
// These will be invoked during route handling to create the unique context for the route
type ISingleStagePostContextFactory[ContextType ISingleStageCallContext[DataType], BodyType any, DataType any] interface {
	// Create the context for the route
	Create(body BodyType) (ContextType, error)
}

// Registers a new route with the router, which will invoke the provided factory to create and execute the context
// for the route when it's called; use this for typical general-purpose calls
func RegisterSingleStageRoute[ContextType ISingleStageCallContext[DataType], DataType any](
	router *mux.Router,
	functionName string,
	factory ISingleStageGetContextFactory[ContextType, DataType],
	logger *slog.Logger,
	serviceProvider *services.ServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		// Log
		args := r.URL.Query()
		logger.Info("Request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
		logger.Debug("Params", slog.String(log.QueryKey, r.URL.RawQuery))

		// Check the method
		if r.Method != http.MethodGet {
			HandleInvalidMethod(logger, w)
			return
		}

		// Create the handler and deal with any input validation errors
		context, err := factory.Create(args)
		if err != nil {
			HandleInputError(logger, w, err)
			return
		}

		// Run the context's processing routine
		status, response, err := runSingleStageRoute[DataType](context, serviceProvider)
		HandleResponse(logger, w, status, response, err)
	})
}

// Registers a new route with the router, which will invoke the provided factory to create and execute the context
// for the route when it's called via POST; use this for typical general-purpose calls
func RegisterSingleStagePost[ContextType ISingleStageCallContext[DataType], BodyType any, DataType any](
	router *mux.Router,
	functionName string,
	factory ISingleStagePostContextFactory[ContextType, BodyType, DataType],
	logger *slog.Logger,
	serviceProvider *services.ServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		// Log
		logger.Info("Request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))

		// Check the method
		if r.Method != http.MethodPost {
			HandleInvalidMethod(logger, w)
			return
		}

		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			HandleInputError(logger, w, fmt.Errorf("error reading request body: %w", err))
			return
		}
		logger.Debug("Body", slog.String(log.BodyKey, string(bodyBytes)))

		// Deserialize the body
		var body BodyType
		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			HandleInputError(logger, w, fmt.Errorf("error deserializing request body: %w", err))
			return
		}

		// Create the handler and deal with any input validation errors
		context, err := factory.Create(body)
		if err != nil {
			HandleInputError(logger, w, err)
			return
		}

		// Run the context's processing routine
		status, response, err := runSingleStageRoute[DataType](context, serviceProvider)
		HandleResponse(logger, w, status, response, err)
	})
}

// Run a route registered with the common single-stage querying pattern
func runSingleStageRoute[DataType any](ctx ISingleStageCallContext[DataType], serviceProvider *services.ServiceProvider) (types.ResponseStatus, *types.ApiResponse[DataType], error) {
	// Get the services
	w := serviceProvider.GetWallet()
	q := serviceProvider.GetQueryManager()

	// Initialize the context with any bootstrapping, requirements checks, or bindings it needs to set up
	status, err := ctx.Initialize()
	if err != nil {
		return status, nil, err
	}

	// Get the context-specific contract state
	err = q.Query(func(mc *batch.MultiCaller) error {
		return ctx.GetState(mc)
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error running chain state query: %w", err)
	}

	// Get the transact opts if this node is ready for transaction
	var opts *bind.TransactOpts
	walletStatus, err := w.GetStatus()
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error getting wallet status: %w", err)
	}
	if utils.IsWalletReady(walletStatus) {
		var err error
		opts, err = w.GetTransactor()
		if err != nil {
			return types.ResponseStatus_Error, nil, fmt.Errorf("error getting node account transactor: %w", err)
		}
	} else {
		opts = &bind.TransactOpts{
			From: walletStatus.Address.NodeAddress,
		}
	}

	// Create the response and data
	data := new(DataType)
	response := &types.ApiResponse[DataType]{
		Data: data,
	}

	// Prep the data with the context-specific behavior
	status, err = ctx.PrepareData(data, opts)
	return status, response, err
}

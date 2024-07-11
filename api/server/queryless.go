package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/services"
	"github.com/rocket-pool/node-manager-core/utils"

	"github.com/gorilla/mux"
)

// Wrapper for callbacks used by call runners that simply run without following a structured pattern of
// querying the chain. This is the most general form of context and can be used by anything as it doesn't
// add any scaffolding.
// Structs implementing this will handle the caller-specific functionality.
type IQuerylessCallContext[DataType any] interface {
	// Prepare the response data in whatever way the context needs to do
	//PrepareData(data *DataType, opts *bind.TransactOpts) error
	PrepareData(data *DataType, opts *bind.TransactOpts) (types.ResponseStatus, error)
}

// Interface for queryless call context factories that handle GET calls.
// These will be invoked during route handling to create the unique context for the route.
type IQuerylessGetContextFactory[ContextType IQuerylessCallContext[DataType], DataType any] interface {
	// Create the context for the route
	Create(args url.Values) (ContextType, error)
}

// Interface for queryless call context factories that handle POST requests.
// These will be invoked during route handling to create the unique context for the route
type IQuerylessPostContextFactory[ContextType IQuerylessCallContext[DataType], BodyType any, DataType any] interface {
	// Create the context for the route
	Create(body BodyType) (ContextType, error)
}

// Registers a new route with the router, which will invoke the provided factory to create and execute the context
// for the route when it's called via GET; use this for typical general-purpose calls
func RegisterQuerylessGet[ContextType IQuerylessCallContext[DataType], DataType any](
	router *mux.Router,
	functionName string,
	factory IQuerylessGetContextFactory[ContextType, DataType],
	logger *slog.Logger,
	serviceProvider services.IServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		// Log
		args := r.URL.Query()
		logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
		logger.Debug("Request params:", slog.String(log.QueryKey, r.URL.RawQuery))

		// Check the method
		if r.Method != http.MethodGet {
			err := HandleInvalidMethod(logger, w)
			if err != nil {
				logger.Error("Error handling response", log.Err(err))
			}
			return
		}

		// Create the handler and deal with any input validation errors
		context, err := factory.Create(args)
		if err != nil {
			err = HandleInputError(logger, w, err)
			if err != nil {
				logger.Error("Error handling response", log.Err(err))
			}
			return
		}

		// Run the context's processing routine
		status, response, err := runQuerylessRoute[DataType](context, serviceProvider)
		err = HandleResponse(logger, w, status, response, err)
		if err != nil {
			logger.Error("Error handling response", log.Err(err))
		}
	})
}

// Registers a new route with the router, which will invoke the provided factory to create and execute the context
// for the route when it's called via POST; use this for typical general-purpose calls
func RegisterQuerylessPost[ContextType IQuerylessCallContext[DataType], BodyType any, DataType any](
	router *mux.Router,
	functionName string,
	factory IQuerylessPostContextFactory[ContextType, BodyType, DataType],
	logger *slog.Logger,
	serviceProvider services.IServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		// Log
		logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))

		// Check the method
		if r.Method != http.MethodPost {
			err := HandleInvalidMethod(logger, w)
			if err != nil {
				logger.Error("Error handling response", log.Err(err))
			}
			return
		}

		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			err = HandleInputError(logger, w, fmt.Errorf("error reading request body: %w", err))
			if err != nil {
				logger.Error("Error handling response", log.Err(err))
			}
			return
		}
		logger.Debug("Request body:", slog.String(log.BodyKey, string(bodyBytes)))

		// Deserialize the body
		var body BodyType
		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			err = HandleInputError(logger, w, fmt.Errorf("error deserializing request body: %w", err))
			if err != nil {
				logger.Error("Error handling response", log.Err(err))
			}
			return
		}

		// Create the handler and deal with any input validation errors
		context, err := factory.Create(body)
		if err != nil {
			err = HandleInputError(logger, w, err)
			if err != nil {
				logger.Error("Error handling response", log.Err(err))
			}
			return
		}

		// Run the context's processing routine
		status, response, err := runQuerylessRoute[DataType](context, serviceProvider)
		err = HandleResponse(logger, w, status, response, err)
		if err != nil {
			logger.Error("Error handling response", log.Err(err))
		}
	})
}

// Run a route registered with no structured chain query pattern
func runQuerylessRoute[DataType any](ctx IQuerylessCallContext[DataType], serviceProvider services.IServiceProvider) (types.ResponseStatus, *types.ApiResponse[DataType], error) {
	// Get the services
	w := serviceProvider.GetWallet()

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
	status, err := ctx.PrepareData(data, opts)
	return status, response, err
}

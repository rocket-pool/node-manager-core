package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/utils/log"
)

const (
	ApiLogColor color.Attribute = color.FgHiBlue
)

type IHandler interface {
	RegisterRoutes(router *mux.Router)
}

type ApiServer struct {
	log        log.ColorLogger
	handlers   []IHandler
	socketPath string
	socket     net.Listener
	server     http.Server
	router     *mux.Router
}

func NewApiServer(socketPath string, handlers []IHandler, baseRoute string, apiVersion string) (*ApiServer, error) {
	// Create the router
	router := mux.NewRouter()

	// Create the manager
	server := &ApiServer{
		log:        log.NewColorLogger(ApiLogColor),
		handlers:   handlers,
		socketPath: socketPath,
		router:     router,
		server: http.Server{
			Handler: router,
		},
	}

	// Register each route
	nmcRouter := router.Host(baseRoute).PathPrefix("/api/v" + apiVersion).Subrouter()
	for _, handler := range server.handlers {
		handler.RegisterRoutes(nmcRouter)
	}

	// Create the socket directory
	socketDir := filepath.Dir(socketPath)
	err := os.MkdirAll(socketDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("error creating socket directory [%s]: %w", socketDir, err)
	}

	return server, nil
}

// Starts listening for incoming HTTP requests
func (s *ApiServer) Start(wg *sync.WaitGroup, socketOwnerUid uint32, socketOwnerGid uint32) error {
	// Remove the socket if it's already there
	_, err := os.Stat(s.socketPath)
	if !errors.Is(err, fs.ErrNotExist) {
		err = os.Remove(s.socketPath)
		if err != nil {
			return fmt.Errorf("error removing socket file: %w", err)
		}
	}

	// Create the socket
	socket, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	s.socket = socket

	// Make it so only the user can write to the socket
	err = os.Chmod(s.socketPath, 0600)
	if err != nil {
		return fmt.Errorf("error setting permissions on socket: %w", err)
	}

	// Set the socket owner to the config file user
	err = os.Chown(s.socketPath, int(socketOwnerUid), int(socketOwnerGid))
	if err != nil {
		return fmt.Errorf("error setting socket owner: %w", err)
	}

	// Start listening
	wg.Add(1)
	go func() {
		err := s.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			s.log.Printlnf("error while listening for HTTP requests: %s", err.Error())
		}
		wg.Done()
	}()

	return nil
}

// Stops the HTTP listener
func (s *ApiServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("error stopping listener: %w", err)
	}
	return nil
}

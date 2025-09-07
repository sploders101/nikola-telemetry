package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/sploders101/nikola-telemetry/dbapi"
	"github.com/sploders101/nikola-telemetry/grpc_services"
	"github.com/sploders101/nikola-telemetry/helpers/config"
	"github.com/sploders101/nikola-telemetry/helpers/tokens"
)

func main() {
	// Read config
	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config.", "error", err.Error())
		os.Exit(1)
	}

	// Set up logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Set up database
	db, err := dbapi.NewDb(config)
	if err != nil {
		slog.Error("An error occurred while opening the db.", "error", err.Error())
		os.Exit(1)
	}

	// Create state managers
	clientManager := tokens.NewClientManager(config)
	tokenManager := tokens.NewTokenManager(config, clientManager)

	// Set up routers
	privateRouter := http.NewServeMux()
	var publicRouter *http.ServeMux
	if config.PublicAddress != nil {
		publicRouter = privateRouter
	} else {
		publicRouter = http.NewServeMux()
	}

	// Add routes
	if err := addPublicRoutes(config, publicRouter); err != nil {
		slog.Error("Failed to load config.", "error", err.Error())
		os.Exit(1)
	}
	if err := addPrivateRoutes(config, privateRouter); err != nil {
		slog.Error("Failed to load config.", "error", err.Error())
		os.Exit(1)
	}

	// Set up gRPC server & services
	grpcServer := grpc.NewServer()
	grpc_services.AddApiRegistrationService(grpcServer, config, db, clientManager, tokenManager)
	reflection.Register(grpcServer)

	// Listen for public routes on public address if one is configured
	// If public address was not configured, public routes were added to the private router.
	if config.PublicAddress != nil {
		go func() {
			slog.Info("Listening on public address.", "address", config.PublicAddress)
			err := http.ListenAndServe(*config.PublicAddress, publicRouter)
			if err != nil {
				slog.Error("An error occurred while listening on public address.", "error", err.Error())
				os.Exit(1)
			}
		}()
	}

	// Listen for private routes & gRPC calls on private router.
	// h2c makes this a little more complicated here. I think this should eventually be e2e encrypted,
	// so this will likely be changed in the future.
	slog.Info("Listening on private address.", "address", config.PrivateAddress)
	h2s := &http2.Server{}
	server := http.Server{
		Addr: config.PrivateAddress,
		Handler: h2c.NewHandler(
			http.HandlerFunc(
				func(response http.ResponseWriter, request *http.Request) {
					if strings.Contains(request.Header.Get("content-type"), "application/grpc") {
						grpcServer.ServeHTTP(response, request)
						return
					}

					privateRouter.ServeHTTP(response, request)
				},
			),
			h2s,
		),
	}
	if err := http2.ConfigureServer(&server, h2s); err != nil {
		slog.Error("An error occurred while setting up h2s.", "error", err.Error())
		os.Exit(1)
	}
	err = server.ListenAndServe()
	if err != nil {
		slog.Error("An error occurred while listening on private address.", "error", err.Error())
	}
}

// Adds public HTTP routes to the given router.
func addPublicRoutes(config config.ConfigFile, router *http.ServeMux) error {
	router.HandleFunc(
		"GET /.well-known/appspecific/com.tesla.3p.public-key.pem",
		func(response http.ResponseWriter, request *http.Request) {
			http.ServeFile(response, request, config.TeslaPubkeyFile)
		},
	)

	return nil
}

// Adds private HTTP routes to the given router.
func addPrivateRoutes(config config.ConfigFile, router *http.ServeMux) error {
	return nil
}

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

	"github.com/sploders101/nikola-telemetry/grpc_services"
	"github.com/sploders101/nikola-telemetry/helpers/config"
	"github.com/sploders101/nikola-telemetry/helpers/tokens"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config.", "error", err.Error())
		os.Exit(1)
	}

	privateRouter := http.NewServeMux()
	var publicRouter *http.ServeMux
	if config.PublicAddress != nil {
		publicRouter = privateRouter
	} else {
		publicRouter = http.NewServeMux()
	}

	if err := addPublicRoutes(config, publicRouter); err != nil {
		slog.Error("Failed to load config.", "error", err.Error())
		os.Exit(1)
	}
	if err := addPrivateRoutes(config, privateRouter); err != nil {
		slog.Error("Failed to load config.", "error", err.Error())
		os.Exit(1)
	}

	tokenManager := tokens.NewTokenManager(config)

	grpcServer := grpc.NewServer()
	grpc_services.AddApiRegistrationService(grpcServer, config, tokenManager)
	reflection.Register(grpcServer)

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

func addPublicRoutes(config config.ConfigFile, router *http.ServeMux) error {
	router.HandleFunc(
		"GET /.well-known/appspecific/com.tesla.3p.public-key.pem",
		func(response http.ResponseWriter, request *http.Request) {
			http.ServeFile(response, request, config.TeslaPubkeyFile)
		},
	)

	return nil
}

func addPrivateRoutes(config config.ConfigFile, router *http.ServeMux) error {
	return nil
}

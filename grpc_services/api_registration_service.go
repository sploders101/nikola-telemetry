package grpc_services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"google.golang.org/grpc"

	pb "github.com/sploders101/nikola-telemetry/gen/shaunkeys/nikola_telemetry/v1"
	"github.com/sploders101/nikola-telemetry/helpers/config"
	"github.com/sploders101/nikola-telemetry/helpers/tokens"
)

type ApiRegistrationService struct {
	pb.UnsafeApiRegistrationServiceServer

	config       config.ConfigFile
	tokenManager *tokens.TokenManager
}

func (self ApiRegistrationService) RegisterApplication(
	_ context.Context,
	request *pb.RegisterApplicationRequest,
) (*pb.RegisterApplicationResponse, error) {
	partnerAccountsUrl, err := url.JoinPath(self.config.TeslaBaseUrl, "/api/1/partner_accounts")
	if err != nil {
		slog.Error("An error occurred while creating the partner accounts URL.", "error", err.Error())
		return nil, err
	}

	bodyBytes, err := json.Marshal(map[string]any{
		"domain": self.config.Domain,
	})
	if err != nil {
		slog.Error("An error occurred while registering the application with Tesla.", "error", err.Error())
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", partnerAccountsUrl, bytes.NewReader(bodyBytes))
	if err != nil {
		slog.Error("An error occurred while creating the request object.", "error", err.Error())
		return nil, err
	}

	partnerToken, err := self.tokenManager.GetPartnerToken()
	if err != nil {
		slog.Error("An error occurred while getting partner token.", "error", err.Error())
		return nil, err
	}

	httpReq.Header.Set("authorization", fmt.Sprintf("Bearer %s", partnerToken))

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		slog.Error("An error occurred while sending partner registration request.", "error", err.Error())
		return nil, err
	}

	respBodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		slog.Error("Error reading response body from tesla.", "error", err.Error())
		return nil, fmt.Errorf("Error reading response body: %w", err)
	}

	if httpResp.StatusCode != 200 {
		slog.Error("Tesla replied with non-200 status.", "status", httpResp.StatusCode, "body", string(respBodyBytes))
		return nil, fmt.Errorf("Tesla replied with status %v.", httpResp.StatusCode)
	}

	return &pb.RegisterApplicationResponse{}, nil
}

func AddApiRegistrationService(server *grpc.Server, config config.ConfigFile, tokenManager *tokens.TokenManager) {
	pb.RegisterApiRegistrationServiceServer(server, ApiRegistrationService{
		config:       config,
		tokenManager: tokenManager,
	})
}

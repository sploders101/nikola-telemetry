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

	"github.com/sploders101/nikola-telemetry/dbapi"
	pb "github.com/sploders101/nikola-telemetry/gen/shaunkeys/nikola_telemetry/v1"
	"github.com/sploders101/nikola-telemetry/helpers/config"
	"github.com/sploders101/nikola-telemetry/helpers/tokens"
)

type ApiRegistrationService struct {
	pb.UnsafeApiRegistrationServiceServer

	db            dbapi.Db
	config        config.ConfigFile
	clientManager tokens.ClientManager
	tokenManager  *tokens.TokenManager
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

	return pb.RegisterApplicationResponse_builder{}.Build(), nil
}

func (self ApiRegistrationService) RegisterUser(
	_ context.Context,
	request *pb.RegisterUserRequest,
) (*pb.RegisterUserResponse, error) {
	user := dbapi.UserDetails{Username: request.GetUsername()}
	if err := self.db.AddUser(&user); err != nil {
		slog.Error("Failed to add user to db.", "error", err.Error())
		return nil, fmt.Errorf("Failed to add user to db: %w", err)
	}

	clientId, err := self.clientManager.GetClientId()
	if err != nil {
		return nil, err
	}

	redirectUri := fmt.Sprintf("https://%s/oauth/redirect/tesla", self.config.Domain)

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", clientId)
	params.Add("redirect_uri", redirectUri)
	params.Add("scope", "openid offline_access user_data vehicle_device_data vehicle_cmds vehicle_charging_cmds")
	params.Add("state", user.RegistrationCode)
	params.Add("show_keypair_step", "true")

	return pb.RegisterUserResponse_builder{
		UserId:           user.Id,
		RegistrationCode: user.RegistrationCode,
		RegistrationUrl:  fmt.Sprintf("https://auth.tesla.com/oauth2/v3/authorize?%s", params.Encode()),
	}.Build(), nil
}

func (self ApiRegistrationService) DeleteUser(
	_ context.Context,
	request *pb.DeleteUserRequest,
) (*pb.DeleteUserResponse, error) {
	if err := self.db.DeleteUser(request.GetUserId()); err != nil {
		return nil, err
	}
	return pb.DeleteUserResponse_builder{}.Build(), nil
}

func AddApiRegistrationService(
	server *grpc.Server,
	config config.ConfigFile,
	db dbapi.Db,
	clientManager tokens.ClientManager,
	tokenManager *tokens.TokenManager,
) {
	pb.RegisterApiRegistrationServiceServer(server, ApiRegistrationService{
		db:            db,
		config:        config,
		clientManager: clientManager,
		tokenManager:  tokenManager,
	})
}

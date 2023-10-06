package naisdevice

import (
	"fmt"
	"path/filepath"

	"github.com/nais/device/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TODO: Denne listen burde hentes fra et sted
// Kan hentes fra naisdevice.GetStatus().GetTenants()
var Tenants = []string{"nav", "dev-nais.io", "fhi-api.com", "naas.mattilsynet.no", "nais.io", "ssb.no"}

func agentConnection() (*grpc.ClientConn, error) {
	userConfigDir, err := config.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("naisdevice config directory: %v", err)
	}
	socket := filepath.Join(userConfigDir, "agent.sock")

	connection, err := grpc.Dial(
		"unix:"+socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, formatGrpcError(err)
	}

	return connection, nil
}

func formatGrpcError(err error) error {
	gerr, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch gerr.Code() {
	case codes.Unavailable:
		return fmt.Errorf("unable to connect to naisdevice; make sure naisdevice is running")
	}
	return fmt.Errorf("%s: %s", gerr.Code(), gerr.Message())
}

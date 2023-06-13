package naisdevice

import (
	"context"
	"fmt"
	"github.com/nais/device/pkg/config"
	"github.com/nais/device/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"path/filepath"
)

var (
	// TODO: Denne listen burde hentes fra et sted
	Tenants = []string{"NAV", "dev-nais.io", "fhi-api.com", "naas.mattilsynet.no", "nais.io", "nav.no", "ssb.no"}
)

func SetTenant(ctx context.Context, tenant string) error {
	connection, err := agentConnection()
	if err != nil {
		return err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	// TODO: naisdevice gir ikke feil hvis man gir den en tenants som ikke er gyldig
	_, err = client.SetActiveTenant(ctx, &pb.SetActiveTenantRequest{Name: tenant})
	if err != nil {
		return formatGrpcError(err)
	}

	return nil
}
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

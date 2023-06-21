package naisdevice

import (
	"context"
	"github.com/nais/device/pkg/pb"
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

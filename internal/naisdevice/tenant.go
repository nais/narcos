package naisdevice

import (
	"context"
	"fmt"

	"github.com/nais/device/pkg/pb"
)

func ListTenants(ctx context.Context) ([]string, error) {
	as, err := agentStatus(ctx)
	if err != nil {
		return nil, err
	}

	var tenants []string
	for _, tenant := range as.GetTenants() {
		tenants = append(tenants, tenant.Name)
	}

	return tenants, nil
}

func GetTenant(ctx context.Context) (string, error) {
	as, err := agentStatus(ctx)
	if err != nil {
		return "", err
	}

	for _, tenant := range as.GetTenants() {
		if tenant.Active {
			return tenant.Name, nil
		}
	}

	return "", fmt.Errorf("no active tenant")
}

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

func agentStatus(ctx context.Context) (*pb.AgentStatus, error) {
	connection, err := agentConnection()
	if err != nil {
		return nil, err
	}

	client := pb.NewDeviceAgentClient(connection)
	defer connection.Close()

	sc, err := client.Status(ctx, &pb.AgentStatusRequest{
		KeepConnectionOnComplete: true,
	})
	if err != nil {
		return nil, formatGrpcError(err)
	}

	s, err := sc.Recv()
	if err != nil {
		return nil, formatGrpcError(err)
	}

	return s, nil
}

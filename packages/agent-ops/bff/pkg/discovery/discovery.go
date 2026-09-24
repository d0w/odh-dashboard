package discovery

import (
	"context"
	"log/slog"
	"time"

	"github.com/opendatahub-io/agent-ops/pkg/fleet"
)

func tempAddGateway(ctx context.Context, registry fleet.Registry, cfg fleet.GatewayConfig) error {
	_, err := registry.Register(ctx, cfg)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}

func PollGateways(ctx context.Context, gatewayRegistry fleet.Registry, interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ctx.Done():
			// done
			return
		case <-ticker.C:
			// find gateways
			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{
				ID:       "derxu-openshell",
				Endpoint: "grpcs://localhost:8080",
				// wed have to assume we get this somehow from a sidecar/operator and it gets mounted to the container
				GatewayCaCert: "/Users/derxu/.config/openshell/gateways/derxu-cluster/mtls/ca.crt",
			})

			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{
				ID:            "openshell-2",
				Endpoint:      "grpcs://localhost:8081",
				GatewayCaCert: "/Users/derxu/.config/openshell/gateways/openshell-2/mtls/ca.crt",
			})
		}
	}
}

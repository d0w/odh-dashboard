// Package gateway handles discovering Openshell gateways in a cluster
package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func tempAddGateway(ctx context.Context, registry *GatewayRegistry, cfg GatewayConfig) error {
	_, err := registry.Register(ctx, cfg)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}

func isGRPCPort(port corev1.ServicePort) bool {
	// Check standard Kubernetes appProtocol field
	if port.AppProtocol != nil && strings.HasPrefix(strings.ToLower(*port.AppProtocol), "grpc") {
		return true
	}

	// Check fallback naming conventions (e.g. "grpc", "grpc-api", "grpc-web")
	if strings.Contains(strings.ToLower(port.Name), "grpc") {
		return true
	}

	return false
}

func getServices(ctx context.Context, k8sClient kubernetes.Interface, applicationName string) ([]corev1.Service, error) {
	svcList, err := k8sClient.CoreV1().Services("").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", applicationName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	return svcList.Items, nil
}

// TODO: Can reuse project's client initialization

// PollGateways will need to find gateway instances
// 1. List services labeled "app.kubernetes.io/name=openshell" and get their grpc port
// 2. Get each gateways ConfigMap and receive oidc issuer, audience, and Ca Cert
// 3. Backend describes self to openshell gateway and receives version + feature flags
func PollGateways(ctx context.Context, k8sClient kubernetes.Interface, gatewayRegistry *GatewayRegistry, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			// done
			return
		case <-ticker.C:
			svcs, err := getServices(ctx, k8sClient, "openshell")
			for _, svc := range svcs {
				for _, port := range svc.Spec.Ports {
					if isGRPCPort(port) {
						fmt.Println(port)
					}
				}
			}
			if err != nil {
				continue
			}

			// find gateways
			tempAddGateway(ctx, gatewayRegistry, GatewayConfig{
				ID:       "derxu-openshell",
				Endpoint: "grpcs://localhost:8080",
				// wed have to assume we get this somehow from a sidecar/operator and it gets mounted to the container
				GatewayCaCert: "/Users/derxu/.config/openshell/gateways/derxu-cluster/mtls/ca.crt",
			})

			tempAddGateway(ctx, gatewayRegistry, GatewayConfig{
				ID:            "openshell-2",
				Endpoint:      "grpcs://localhost:8081",
				GatewayCaCert: "/Users/derxu/.config/openshell/gateways/openshell-2/mtls/ca.crt",
			})
		}
	}
}

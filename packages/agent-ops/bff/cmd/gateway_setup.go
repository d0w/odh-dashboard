package main

import (
	"log/slog"
	"os"
	"strings"

	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"

	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"
)

const (
	defaultPort       = "8080"
	defaultGatewayURL = "localhost:50051"
)

type gatewayClients struct {
	sdk        openshell.ClientInterface
	uploadExec *clients.RawExecClient
}

func (c *gatewayClients) Close() {
	if c.sdk != nil {
		if err := c.sdk.Close(); err != nil {
			slog.Warn("SDK client close failed", "error", err)
		}
	}
	if c.uploadExec != nil {
		if err := c.uploadExec.Close(); err != nil {
			slog.Warn("upload exec client close failed", "error", err)
		}
	}
}

func warnGatewayConfig(gatewayURL, gatewayCACert string, authDisabled bool) {
	if gatewayURL == defaultGatewayURL {
		slog.Warn("gateway URL is the default — verify OPENSHELL_GATEWAY_URL is configured correctly", "url", gatewayURL)
	}
	if gatewayCACert != "" && !strings.HasPrefix(gatewayURL, "grpcs://") && !strings.HasPrefix(gatewayURL, "https://") {
		slog.Warn(
			"gateway CA cert is set but gateway URL has no TLS scheme; use grpcs:// or https:// for TLS gateways",
			"url", gatewayURL,
			"caCert", gatewayCACert,
		)
	}
	if authDisabled {
		slog.Warn("AUTH_DISABLED=true — authentication is OFF; never use this outside local development")
	}
}

func exitOnError(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}

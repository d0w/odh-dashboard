package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
	testclients "github.com/opendatahub-io/agent-ops/internal/clients"
	"github.com/opendatahub-io/agent-ops/internal/middleware"
	"github.com/opendatahub-io/agent-ops/pkg/gateway"

	openshellauth "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/auth"
	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"
	openshellmodels "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/models"
	openshellapi "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/server"
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

func embeddedFeatureFlags() openshellmodels.FeatureFlags {
	flags := openshellmodels.FeatureFlags{
		Terminal:          true,
		FileTransfer:      true,
		Settings:          true,
		GlobalPolicy:      true,
		CredentialRefresh: true,
		Services:          true,
		DraftPolicy:       true,
	}
	// for _, name := range embeddedUnsupportedFeatures {
	// 	if name == "terminal" {
	// 		flags.Terminal = false
	// 	}
	// }
	return flags
}

func gatewayInstanceFactory(ctx context.Context, cfg gateway.GatewayConfig) (gateway.GatewayInstance, error) {
	appClients, err := testclients.NewGatewayClients(
		cfg.Endpoint,
		// TODO: Make based on factory inputs
		cfg.GatewayCaCert,
		// "/Users/derxu/.config/openshell/gateways/derxu-cluster/mtls/tls.crt",
		"",
		// "/Users/derxu/.config/openshell/gateways/derxu-cluster/mtls/tls.key",
		"",
	)
	if err != nil {
		return gateway.GatewayInstance{Handler: nil, CloseFunc: nil}, err
	}
	app := openshellapi.NewApp(
		appClients.SDK,
		appClients.UploadExec,
		middleware.NewAuthReplacerMiddleware.New(openshellauth.Config{}),
		"",
		// TODO: remove statics
		openshellmodels.AuthConfigResponse{
			AdminRole:    "openshell-admin",
			LogoutURL:    "",
			Features:     embeddedFeatureFlags(),
			AuthDisabled: false,
		},
	)
	slog.Info(fmt.Sprintf("Created factory instance: %s", cfg.ID))
	return gateway.GatewayInstance{
		CloseFunc: func() error {
			return nil
		},
		Handler: app.Routes(),
	}, nil
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

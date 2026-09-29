package gateway

import (
	"context"
	"fmt"
	"log/slog"

	testclients "github.com/opendatahub-io/agent-ops/internal/clients"

	openshellauth "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/auth"
	openshellmodels "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/models"
	openshellapi "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/server"
)

// TODO: Move this out
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

func gatewayInstanceFactory(ctx context.Context, cfg GatewayConfig) (*GatewayInstance, error) {
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
		return nil, err
	}
	app := openshellapi.NewApp(
		appClients.SDK,
		appClients.UploadExec,
		openshellauth.New(openshellauth.Config{}),
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
	return &GatewayInstance{
		GatewayConfig: cfg,
		Handler:       app.Routes(),
	}, nil
}

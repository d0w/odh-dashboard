package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/opendatahub-io/agent-ops/pkg/fleet"
)

// GatewayRegistry is a typed version of the generic fleet registry
type GatewayRegistry = fleet.Registry[GatewayConfig, *GatewayInstance]

type GatewayResponse struct {
	ID            string `json:"id"`
	Endpoint      string `json:"endpoint"`
	GatewayCaCert string `json:"ca_cert"`
}

type GatewayConfig struct {
	ID            string `json:"id"`
	Endpoint      string `json:"endpoint"`
	GatewayCaCert string `json:"ca_cert"`
}

// GatewayInstance represents an entry within a fleet registry (which holds a map of instances)
// Every instance in a fleet registry will need to implement http.Handler
type GatewayInstance struct {
	GatewayConfig
	Handler http.Handler
}

func keyFunc(config GatewayConfig) string {
	return config.ID
}

func NewGatewayRegistry() *GatewayRegistry {
	return fleet.NewRegistry[GatewayConfig, *GatewayInstance](gatewayInstanceFactory, keyFunc)
}

// ServeHTTP serves the handler associated with this individual gateway instance
// This will typically be called by the fleet router
func (i *GatewayInstance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	i.Handler.ServeHTTP(w, r)
}

func (i *GatewayInstance) Close() error {
	return nil
}

// MarshalJSON handles custom JSON marshaling
func (i *GatewayInstance) MarshalJSON() ([]byte, error) {
	return json.Marshal(&GatewayResponse{
		ID:            i.ID,
		Endpoint:      i.Endpoint,
		GatewayCaCert: i.GatewayCaCert,
	})
}

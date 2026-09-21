package api

import (
	agentsmock "github.com/opendatahub-io/agent-ops/internal/integrations/agents/mocks"
	"github.com/opendatahub-io/agent-ops/internal/repositories"
)

func testRepositoriesWithAgents() *repositories.Repositories {
	return repositories.NewRepositories(testAgentsFactory())
}

func testAgentsFactory() *agentsmock.Factory {
	return &agentsmock.Factory{Client: agentsmock.NewDemoClient()}
}

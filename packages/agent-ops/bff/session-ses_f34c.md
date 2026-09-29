# TLS setup for OpenShell gateway in K8s cluster

**Session ID:** ses_f34ca2586ffeGR45lQzwjnnSTg
**Created:** 9/22/2026, 6:21:39 PM
**Updated:** 9/23/2026, 1:36:04 PM

---

## User

this backend needs to interface with another openshell gateway in a k8s cluster. right now i'm trying to get tls to work but not sure how. i have the svc/openshell portforwarded. the cluster uses cert-manager and service mesh but cert-manager does not require to be used by openshell. this is the openshell docs: https://docs.nvidia.com/openshell/latest/kubernetes/managing-certificates. don't change code, but investigate how to get this to work

---

## Assistant (Build · GPT-5.6 Terra · 7.0s)

_Thinking:_

**Investigating backend rules**

I need to dig into the special rules for bff Go since it's relevant to understanding this backend. I want to make sure I grasp the development workflow without any code involved. Also, I need to focus on the TLS aspects for security. It might be helpful to webfetch the documentation. I should inspect the code configuration related to TLS integrations and then send out an update once I've gathered the information.

Inspect BFF TLS/client wiring, cluster gateway path, NVIDIA guidance. No code changes.

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/.claude/rules/bff-go.md"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/.claude/rules/bff-go.md</path>
<type>file</type>
<content>
1: ---
2: description: BFF (Backend For Frontend) API patterns and Go conventions for ODH Dashboard packages and distributions
3: globs: "packages/*/bff/**,packages/*/upstream/bff/**,distributions/core-bff/bff/**"
4: alwaysApply: false
5: paths:
6:   - "packages/*/bff/**"
7:   - "packages/*/upstream/bff/**"
8:   - "distributions/core-bff/bff/**"
9: ---
10: 
11: # BFF API Patterns & Go Conventions
12: 
13: ## Which packages have BFFs
14: 
15: gen-ai, model-registry, maas, automl, autorag, mlflow, eval-hub — each with a `bff/` directory containing Go code. The `distributions/core-bff` module also has a Go BFF in `distributions/core-bff/bff/`.
16: 
17: ## Directory structure
18: 
19: ```text
20: bff/
21: ├── cmd/
22: │   ├── main.go            # Entry point, flag parsing, server setup
23: │   └── helpers.go         # getEnvAsInt, getEnvAsString, parseLevel
24: ├── internal/
25: │   ├── api/               # HTTP handlers and routing
26: │   │   ├── app.go         # App struct, NewApp(), Routes()
27: │   │   ├── *_handler.go   # Per-endpoint handlers
28: │   │   ├── middleware.go  # Auth, CORS, logging, client attachment
29: │   │   └── errors.go      # Error response helpers
30: │   ├── config/            # EnvConfig struct
31: │   ├── constants/         # Path constants, header keys
32: │   ├── integrations/      # External service clients
33: │   │   ├── kubernetes/    # K8s client, SAR, identity
34: │   │   └── httpclient/    # HTTP client for upstream APIs
35: │   ├── models/            # DTOs
36: │   ├── repositories/      # Business logic layer
37: │   └── mocks/             # Mock implementations
38: ├── openapi/src/*.yaml     # OpenAPI specs
39: ├── go.mod
40: ├── go.sum
41: └── Makefile
42: ```
43: 
44: ## Router and handlers
45: 
46: All BFFs use `julienschmidt/httprouter`. Handler signature:
47: 
48: ```go
49: func (app *App) MyHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params)
50: ```
51: 
52: ### Route registration
53: 
54: ```go
55: apiRouter := httprouter.New()
56: apiRouter.NotFound = http.HandlerFunc(app.notFoundResponse)
57: apiRouter.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
58: 
59: apiRouter.GET(constants.ModelsListPath,
60:   app.AttachNamespace(
61:     app.RequireAccessToService(
62:       app.AttachLlamaStackClient(app.LlamaStackModelsHandler))))
63: ```
64: 
65: ## Middleware
66: 
67: Middleware wraps handlers — composed inside-out:
68: 
69: | Middleware | Purpose |
70: |---|---|
71: | `InjectRequestIdentity` | Extract user identity from headers |
72: | `AttachNamespace` | Extract and validate namespace param |
73: | `RequireAccessToService` | SAR/SSAR check for namespace access |
74: | `AttachLlamaStackClient` / `AttachClient` | Create service client per request |
75: 
76: Chain: `AttachNamespace → RequireAccess → AttachClient → handler`.
77: 
78: ## Authentication
79: 
80: ### Auth methods
81: 
82: | Method | Header | Usage |
83: |---|---|---|
84: | `internal` | `kubeflow-userid` + `kubeflow-groups` | Kubeflow/ODH deployments |
85: | `user_token` | `Authorization: Bearer <token>` | RHOAI deployments |
86: | `disabled` | N/A | Mock/testing |
87: 
88: Identity stored in `RequestIdentity` (UserID, Groups, Token) via `context.WithValue`.
89: 
90: ### RBAC
91: 
92: SubjectAccessReview / SelfSubjectAccessReview for namespace-level permission checks (e.g., `CanListServicesInNamespace`).
93: 
94: ## Configuration
95: 
96: CLI flags + environment variables via `EnvConfig`:
97: 
98: | Flag | Env Var | Description |
99: |---|---|---|
100: | `-port` | `PORT` | Listen port (default 8080) |
101: | `-auth-method` | `AUTH_METHOD` | `internal`, `user_token`, `disabled` |
102: | `-auth-token-header` | `AUTH_TOKEN_HEADER` | Token header (default `Authorization`) |
103: | `-mock-k8s-client` | `MOCK_K8S_CLIENT` | Enable K8s mocks |
104: | `-allowed-origins` | `ALLOWED_ORIGINS` | CORS origins |
105: | `-log-level` | `LOG_LEVEL` | Log level |
106: | `-insecure-skip-verify` | `INSECURE_SKIP_VERIFY` | Skip TLS verification (default `false`, **local dev only** — forbidden in production/CI) |
107: 
108: > **Security**: `INSECURE_SKIP_VERIFY=true` is permitted only for local debugging against non-production endpoints. Production and CI profiles must fail configuration validation at startup if this flag is enabled.
109: 
110: ## Error handling
111: 
112: Standard error envelope:
113: 
114: ```go
115: type ErrorEnvelope struct {
116:     Error *integrations.HTTPError `json:"error"`
117: }
118: 
119: type HTTPError struct {
120:     StatusCode int `json:"-"`
121:     ErrorResponse
122: }
123: 
124: type ErrorResponse struct {
125:     Code    string `json:"code"`
126:     Message string `json:"message"`
127: }
128: ```
129: 
130: Response helpers: `badRequestResponse`, `unauthorizedResponse`, `forbiddenResponse`, `notFoundResponse`, `serverErrorResponse`, `serviceUnavailableResponse`.
131: 
132: ## Request/Response helpers
133: 
134: ```go
135: func (app *App) WriteJSON(w http.ResponseWriter, status int, data any, headers http.Header) error
136: ```
137: 
138: ## Upstream service calls
139: 
140: BFFs call upstream services via HTTP clients — not generic proxies:
141: 
142: 1. Create `httpclient.NewHTTPClient(logger, baseURL, headers, ...)`
143: 2. Forward user token as `Authorization: Bearer <token>`
144: 3. Service URLs from K8s status (e.g., `LlamaStackDistribution.Status.ServiceURL`) or env vars
145: 
146: ## Testing
147: 
148: ### Unit tests
149: 
150: - `*_test.go` adjacent to implementation
151: - `httptest.NewRecorder()` + `http.NewRequest()` for handler tests
152: - Mock factories for K8s, service clients
153: - Table-driven tests:
154: 
155: ```go
156: tests := []struct {
157:     name    string
158:     input   Request
159:     wantErr bool
160: }{...}
161: for _, tt := range tests {
162:     t.Run(tt.name, func(t *testing.T) { ... })
163: }
164: ```
165: 
166: ### envtest
167: 
168: gen-ai and model-registry use K8s `envtest` for integration tests.
169: 
170: ### Contract tests
171: 
172: - Framework: `@odh-dashboard/contract-tests`
173: - BFF must expose `/healthcheck`, accept mock flags, and have OpenAPI spec
174: - Tests run against mock BFF and validate with `toMatchContract(apiSchema, { ref, status })`
175: - OpenAPI spec in `bff/openapi/src/*.yaml`
176: - Run: `npm run test:contract` (starts mock BFF, runs Jest)
177: 
178: ## Go conventions
179: 
180: - Standard `cmd/` + `internal/` layout
181: - `internal/api/` for all HTTP concerns
182: - `internal/integrations/` for external service clients
183: - `internal/repositories/` for business logic
184: - `internal/models/` for DTOs
185: - `slog` for structured logging
186: - `flag` package for CLI flags with env var fallbacks

(End of file - total 186 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/.claude/rules/security.md"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/.claude/rules/security.md</path>
<type>file</type>
<content>
1: ---
2: description: Security review guidelines for auth, secrets, input validation, and K8s API interactions
3: globs: "backend/**,**/api/**,**/routes/**,**/middleware/**"
4: alwaysApply: false
5: paths:
6:   - "backend/**"
7:   - "**/api/**"
8:   - "**/routes/**"
9:   - "**/middleware/**"
10: ---
11: 
12: # ODH Dashboard Security Review Guidelines
13: 
14: ## Authentication and Authorization
15: 
16: - All backend routes must validate user authentication via OpenShift token
17: - Check RBAC permissions before performing Kubernetes operations
18: - Never expose cluster-admin operations to regular users
19: - Validate that users have access to the namespace they're operating in
20: 
21: ## Secrets and Credentials
22: 
23: - Never hardcode secrets, tokens, API keys, or passwords in source code
24: - Use Kubernetes Secrets for sensitive configuration
25: - Environment variables for runtime configuration (never committed)
26: - Tracked `.env` files must only contain non-sensitive defaults and placeholders — never real credentials
27: - Sensitive local overrides go in `.env.local` (gitignored). See `.env.local.example` for the template
28: 
29: ## Input Validation
30: 
31: - Validate all user input on both frontend and backend
32: - Sanitize inputs to prevent XSS in React components (avoid `dangerouslySetInnerHTML`)
33: - Validate Kubernetes resource names follow RFC 1123: lowercase alphanumeric and hyphens only, max 63 chars, must start and end with alphanumeric
34: - Backend route handlers must validate request body schemas
35: 
36: ## Kubernetes API Security
37: 
38: - Use least-privilege RBAC when creating ServiceAccounts or ClusterRoles
39: - Validate resource ownership before modification or deletion
40: - Use proper namespace scoping — avoid cluster-wide operations unless necessary
41: - Check for proper error handling on 403 (Forbidden) responses
42: 
43: ## Frontend Security
44: 
45: - Use PatternFly components, but treat XSS prevention as an application responsibility
46: - Never render untrusted HTML; sanitize/validate any rich content before rendering
47: - Avoid direct DOM manipulation
48: - Use React's built-in escaping for dynamic content
49: - Validate URLs before navigation: enforce http/https scheme allowlist, reject `javascript:` and `data:` protocols
50: 
51: ## Proxy and Network
52: 
53: - Backend proxies requests to Kubernetes API — validate proxy targets
54: - Do not allow open redirects
55: - Validate WebSocket connections are authenticated
56: - BFF services should validate upstream API responses
57: 
58: ## Dependencies
59: 
60: - Keep dependencies up to date (Renovate is configured for automated updates)
61: - Review new dependency additions for security advisories
62: - Prefer well-maintained packages with active security response

(End of file - total 62 lines)
</content>
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://docs.nvidia.com/openshell/latest/kubernetes/managing-certificates",
  "format": "markdown"
}
```

**Output:**
```
> For clean Markdown of any page, append .md to the page URL.
> For a complete documentation index, see https://docs.nvidia.com/openshell/llms.txt.
> For AI client integration (Claude Code, Cursor, etc.), connect to the MCP server at https://docs.nvidia.com/openshell/_mcp/server.

# Managing Certificates

> Configure the OpenShell Helm chart to use cert-manager for mTLS certificate issuance and automatic renewal.

The OpenShell gateway uses mTLS certificates for transport between the gateway and sandbox supervisors. These certificates are not Kubernetes user authentication; configure OIDC or a trusted access proxy for user access. The Helm chart supports two ways to provision and manage the certificate bundle:

| Mode                            | When to use                                                                                                                                 |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Built-in `pkiInitJob` (default) | The default path. A pre-install Kubernetes Job generates a self-signed CA and certificates during installation. No additional dependencies. |
| cert-manager                    | Production deployments that need automatic certificate rotation managed by a running controller.                                            |

The rest of this page covers switching to cert-manager. The built-in mode requires no configuration.

When `certManager.enabled=true`, cert-manager owns TLS certificate generation.
The chart still runs a JWT-only initialization hook because cert-manager does
not create the sandbox JWT signing Secret required by the gateway. This
cert-manager precedence applies even if `pkiInitJob.enabled` remains true.

## Install cert-manager

Install cert-manager from the OCI registry with CRD support enabled:

```shell
helm upgrade --install cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version v1.20.3 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true \
  --wait
```

Verify the cert-manager pods are running:

```shell
kubectl -n cert-manager get pods
```

## Install OpenShell with cert-manager PKI

Pass the cert-manager values override when installing or upgrading the chart:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set certManager.enabled=true
```

The chart creates a self-signed CA, issues server and client certificates from it, and cert-manager handles renewal before expiry.
The chart also runs a pre-install hook in JWT-only mode to create the gateway's
sandbox JWT signing Secret. That Secret is separate from the cert-manager TLS
certificate Secrets and is mounted at `/etc/openshell-jwt`.

## Using a real Issuer for the server certificate

By default, cert-manager issues both the server and client certificates from
a self-signed CA the chart creates — this rotates automatically, but the
server certificate is still not publicly trusted. `certManager.serverIssuerRef`
overrides the `issuerRef` on the server `Certificate` resource to point at a
real `Issuer` or `ClusterIssuer` instead, for example an ACME issuer:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set certManager.enabled=true \
  --set certManager.serverIssuerRef.name=letsencrypt-prod \
  --set certManager.serverIssuerRef.kind=ClusterIssuer \
  --set certManager.serverDnsNames[0]=openshell.example.com
```

### Dual certificate architecture

When `serverIssuerRef` is set, the chart creates **two** server certificates:

1. **Internal certificate** (`openshell-server-tls`): signed by the chart CA
   with internal SANs (`*.svc.cluster.local`, `localhost`, etc.).
2. **External certificate** (`openshell-server-external-tls`): signed by the
   configured issuer (e.g. ACME) with only the hostnames from
   `certManager.serverDnsNames`.

The gateway uses **SNI** to select which certificate to present:
supervisors connect via internal service names and receive the internal
certificate (verified against the chart CA they already trust), while CLI
users connecting through a Route or ingress use the external hostname and
receive the ACME certificate. This keeps supervisor trust pinned to only
the operator's chart CA — no WebPKI root trust is needed.

Public CAs such as Let's Encrypt reject certificate requests that include
internal-only names per CA/Browser Forum baseline requirements. The chart
validates this at install time and fails with an actionable error if
`certManager.serverDnsNames` contains internal-only entries while
`serverIssuerRef` is set.

You do **not** need to set `server.grpcEndpoint` to the external hostname.
Supervisors connect via the internal service name automatically. Setting
`server.grpcEndpoint` to an external hostname would cause supervisors to
receive the ACME certificate (via SNI) which they cannot verify against the
chart CA.

The default `clientCaFromServerTlsSecret=true` is correct even when
`serverIssuerRef` is set: the internal server certificate is always signed
by the chart CA (the same CA that signs the client certificate), so its
`ca.crt` is the right trust anchor for mTLS verification.

## Next Steps

Return to [Setup](/kubernetes/setup) to complete the installation. For
exposing the gateway externally on OpenShift with a real certificate, see
[OpenShift](/kubernetes/openshift).
```

**Tool: glob**

**Input:**
```json
{
  "path": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "pattern": "**/*.{go,mod,yaml,yml,json,md}"
}
```

**Output:**
```
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/router.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/utils.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/discovery/discovery.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/.golangci.yaml
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/README.md
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/openapi/src/agent-ops.yaml
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/docs/extensions.md
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/helpers/k8s.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/helpers/http.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/helpers/logging.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/errors/errors.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/errors.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/public_helpers.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/helpers.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/middleware.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/testapp.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/namespaces_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/openapi_handler_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/user_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/deploy_handler_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/helpers_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/extensions.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/deploy_validation_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/extensions_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/agent_runtime_detail_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/app.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/agent_runtime_detail_handler_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/agent_runtimes_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/openapi_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/agent_handlers_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/healthcheck_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/agent_lifecycle_handler_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/test_agents_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/delete_handler_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/middleware_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/agent_runtimes_handler_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/cache_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/deploy_validation.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/deploy_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/agent_lifecycle_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/app_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/delete_handler.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/agent_handlers.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/token_k8s_client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/agent_rbac.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/k8mocks/base_testenv.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/k8mocks/mock_factory.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/k8mocks/internal_k8s_client_mock.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/k8mocks/token_k8s_client_mock.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/types.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/factory.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/internal_k8s_client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/agent_enrichment_rbac_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/shared_k8s_client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/agent_rbac_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/agent_enrichment_rbac.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/errors.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/factory_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/client_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/config_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/middleware.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/README.md
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/factory.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/bffmocks/mock_client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/errors_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/config.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/errors.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/url_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/factory_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/manifests_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/deploy_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/client_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/client_detail_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/sandbox.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/k8s_errors.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/tool_connections.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/tool_connections_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/agent_card_enrichment_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/agent_card_enrichment.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/mapper.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/route_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/route.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/factory.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/mapper_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/deploy.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/manifests.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/metadata_test.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/url.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/deploy_types.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/types.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/metadata.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/mocks/demo_client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/mocks/client.go
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/httpclient/http.go

(Results are truncated: showing first 100 results. Consider using a more specific path or pattern.)
```

---

## Assistant (Build · GPT-5.6 Terra · 5.2s)

_Thinking:_



**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "pattern": "OpenShell|openshell|TLS|tls\\.Config|InsecureSkipVerify|certificate|cert-manager|gateway",
  "include": "*.go"
}
```

**Output:**
```
Found 100 matches (more matches available)
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/discovery/discovery.go:
  Line 20: func PollGateways(ctx context.Context, gatewayRegistry fleet.Registry, interval time.Duration) {

  Line 29: 			// find gateways

  Line 30: 			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{

  Line 31: 				ID:       "derxu-openshell",

  Line 35: 			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{

  Line 36: 				ID:       "openshell-2",


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry_test.go:
  Line 24: 	created, err := registry.Register(context.Background(), GatewayConfig{ID: "openshell"})

  Line 26: 		t.Fatalf("register gateway: %v", err)

  Line 29: 		t.Fatal("gateway was not registered")

  Line 32: 	router, err := NewRouter("/agent-ops/api/openshell", registry)

  Line 40: 		"/agent-ops/api/openshell/openshell/v1/models%2Fmy-model?include=details",


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry.go:
  Line 21: // TODO: make package-agnostic. no gateway specific stuff just registry

  Line 57: 		return false, fmt.Errorf("create gateway %q: %w", config.ID, err)

  Line 91: 	gatewayInstance, ok := r.entries[id]

  Line 96: 	fmt.Printf("%#v\n", gatewayInstance)

  Line 104: 	gatewayInstance.Handler.ServeHTTP(w, upstreamReq)

  Line 108: 	// close all gateway instances. be aware of race conditions


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/main.go:
  Line 23: const fleetRouterPrefix = "/agent-ops/api/openshell"

  Line 30: 	flag.StringVar(&certFile, "cert-file", "", "Path to TLS certificate file")

  Line 31: 	flag.StringVar(&keyFile, "key-file", "", "Path to TLS key file")

  Line 43: 	// bundle-paths accepts a comma-separated list of CA bundle file paths to trust for outbound TLS.

  Line 46: 	flag.Func("bundle-paths", "Comma-separated list of PEM CA bundle file paths to trust for outbound TLS (optional)", newOriginParser(&cfg.BundlePaths, defaultBundlePaths))

  Line 51: 	// TLS configuration flags

  Line 52: 	flag.BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-verify", getEnvAsBool("INSECURE_SKIP_VERIFY", false), "Skip TLS certificate verification (useful for development, default: false)")

  Line 97: 	fleetRegistry := fleet.NewRegistry(gatewayInstanceFactory)

  Line 107: 	router.Get("/agent-ops/api/openshell/gateways", func(w http.ResponseWriter, r *http.Request) {

  Line 108: 		// list gateways

  Line 129: 		tlsCfg, err := tlsprofile.ServerTLSConfig(context.Background(), logger)

  Line 131: 			logger.Error("failed to resolve TLS configuration from cluster profile", "error", err)

  Line 134: 		srv.TLSConfig = tlsCfg

  Line 139: 		logger.Info("starting server", "addr", srv.Addr, "TLS enabled", (certFile != "" && keyFile != ""))

  Line 142: 			err = srv.ListenAndServeTLS(certFile, keyFile)

  Line 151: 	// gateway discovery service


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/gateway_setup.go:
  Line 10: 	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"

  Line 14: 	openshellauth "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/auth"

  Line 15: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"

  Line 16: 	openshellmodels "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/models"

  Line 17: 	openshellapi "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/server"

  Line 25: type gatewayClients struct {

  Line 26: 	sdk        openshell.ClientInterface

  Line 30: func (c *gatewayClients) Close() {

  Line 43: func embeddedFeatureFlags() openshellmodels.FeatureFlags {

  Line 44: 	flags := openshellmodels.FeatureFlags{

  Line 61: func gatewayInstanceFactory(ctx context.Context, cfg fleet.GatewayConfig) (fleet.GatewayInstance, error) {

  Line 71: 	app := openshellapi.NewApp(

  Line 74: 		openshellauth.New(openshellauth.Config{}),

  Line 77: 		openshellmodels.AuthConfigResponse{

  Line 78: 			AdminRole:    "openshell-admin",

  Line 93: func warnGatewayConfig(gatewayURL, gatewayCACert string, authDisabled bool) {

  Line 94: 	if gatewayURL == defaultGatewayURL {

  Line 95: 		slog.Warn("gateway URL is the default — verify OPENSHELL_GATEWAY_URL is configured correctly", "url", gatewayURL)

  Line 97: 	if gatewayCACert != "" && !strings.HasPrefix(gatewayURL, "grpcs://") && !strings.HasPrefix(gatewayURL, "https://") {

  Line 99: 			"gateway CA cert is set but gateway URL has no TLS scheme; use grpcs:// or https:// for TLS gateways",

  Line 100: 			"url", gatewayURL,

  Line 101: 			"caCert", gatewayCACert,


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/config/environment.go:
  Line 82: 	// certificates to the HTTP client's Root CAs for outbound TLS connections.

  Line 99: 	// ─── TLS ────────────────────────────────────────────────────

  Line 100: 	// TLS verification settings for HTTP client connections to the Client

  Line 101: 	// InsecureSkipVerify when true, skips TLS certificate verification (useful for development/local setups)

  Line 103: 	InsecureSkipVerify bool


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/mapper/agent_test.go:
  Line 19: // 				agents.LabelOpenShellManagedBy: agents.OpenShellManagedByValue,

  Line 60: // func TestAgentDetailToRuntimeDetail_OpenShellDefaultsResourceType(t *testing.T) {

  Line 63: // 			Name:      "openshell-agent",

  Line 64: // 			Namespace: "openshell-ns",

  Line 66: // 				agents.LabelOpenShellManagedBy: agents.OpenShellManagedByValue,

  Line 67: // 				agents.LabelOpenShellSandboxID: "uuid-123",


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/mapper/agent_card_test.go:
  Line 109: // 	assert.Equal(t, []string{"mTLS"}, discoverAuthenticationMethods(&agents.AgentCardObserved{

  Line 117: // 	assert.Equal(t, []string{"mTLS", "Bearer"}, discoverAuthenticationMethods(&agents.AgentCardObserved{


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/models/rbac_types.go:
  Line 7: // CertificateItem represents a Secret or ConfigMap suitable for certificate usage.


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/clients/gateway.go:
  Line 9: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"

  Line 10: 	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"

  Line 13: // GatewayClients are the per-gateway clients an App needs: the SDK client for

  Line 14: // the OpenShell API, and the raw-exec client for binary-safe uploads.

  Line 16: // One set per gateway. A host fronting several gateways builds several — the

  Line 20: 	SDK        openshell.ClientInterface

  Line 41: // NewGatewayClients dials one OpenShell gateway.

  Line 43: // gatewayURL accepts grpc://, grpcs://, http://, https:// or a bare host:port;

  Line 44: // TLS is inferred from the scheme. Certificate paths are optional, and mTLS

  Line 45: // requires both a client certificate and key or neither.

  Line 46: func NewGatewayClients(gatewayURL, caCert, clientCert, clientKey string) (*GatewayClients, error) {

  Line 47: 	useTLS := strings.HasPrefix(gatewayURL, "grpcs://") || strings.HasPrefix(gatewayURL, "https://")

  Line 48: 	address := NormalizeGatewayAddress(gatewayURL, useTLS)

  Line 51: 		return nil, fmt.Errorf("gateway mTLS requires both a client certificate and key")

  Line 54: 	cfg := openshell.Config{

  Line 56: 		// Auth:    clients.ContextAuthProvider{RequireTLS: useTLS},

  Line 57: 		Auth: openshell.NoAuth(),

  Line 59: 	if useTLS {

  Line 60: 		tlsCfg := &openshell.TLSConfig{CAFile: caCert}

  Line 65: 		cfg.TLS = tlsCfg

  Line 67: 		cfg.TLS = &openshell.TLSConfig{Insecure: true}

  Line 70: 	sdkClient, err := openshell.NewClient(cfg)

  Line 76: 	uploadExec, err := clients.NewRawExecClient(rawHost, caCert, clientCert, clientKey, useTLS)

  Line 87: // NormalizeGatewayAddress maps the accepted gateway URL forms onto the http(s)

  Line 89: func NormalizeGatewayAddress(gatewayURL string, useTLS bool) string {

  Line 91: 	case strings.HasPrefix(gatewayURL, "grpcs://"):

  Line 92: 		return "https://" + strings.TrimPrefix(gatewayURL, "grpcs://")

  Line 93: 	case strings.HasPrefix(gatewayURL, "grpc://"):

  Line 94: 		return "http://" + strings.TrimPrefix(gatewayURL, "grpc://")

  Line 95: 	case strings.HasPrefix(gatewayURL, "https://"), strings.HasPrefix(gatewayURL, "http://"):

  Line 96: 		return gatewayURL


(Results truncated. Consider using a more specific path or pattern.)
```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops",
  "pattern": "openshell|OpenShell|cert-manager|gateway|TLS",
  "include": "*.{yaml,yml,json,md,go,ts,tsx}"
}
```

**Output:**
```
Found 100 matches (more matches available)
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/AGENTS.md:
  Line 236: | `-insecure-skip-verify` | `INSECURE_SKIP_VERIFY` | Skip upstream TLS verify (dev only)      | false         |


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/api/openapi/agent-ops.yaml:
  Line 124:         labeled `openshell.ai/managed-by=openshell`.

  Line 834:           description: Mutual TLS mode for the agent.


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/README.md:
  Line 19: Agent endpoints validate `{ns}` and `{name}` as DNS-1123 identifiers. List discovery loads Sandbox CRs (`agents.x-k8s.io/v1beta1/sandboxes`) labeled `openshell.ai/managed-by=openshell`. Detail loads any authorized Sandbox CR by name (no label filter).

  Line 102: | `-cert-file` | `CERT_FILE` | TLS certificate path (enables TLS when paired with key) |

  Line 103: | `-key-file` | `KEY_FILE` | TLS key path |

  Line 104: | `-insecure-skip-verify` | `INSECURE_SKIP_VERIFY` | Skip upstream TLS verify (dev only) |

  Line 107: TLS: If both `cert-file` and `key-file` are provided the server starts with HTTPS.

  Line 269: ### Disabling TLS verification (development only)

  Line 271: For local Kubeflow installations with self-signed certificates, you may need to disable TLS certificate verification.

  Line 289: > **Warning:** Only use in development. Keep TLS verification enabled in production.


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry.go:
  Line 21: // TODO: make package-agnostic. no gateway specific stuff just registry

  Line 57: 		return false, fmt.Errorf("create gateway %q: %w", config.ID, err)

  Line 91: 	gatewayInstance, ok := r.entries[id]

  Line 96: 	fmt.Printf("%#v\n", gatewayInstance)

  Line 104: 	gatewayInstance.Handler.ServeHTTP(w, upstreamReq)

  Line 108: 	// close all gateway instances. be aware of race conditions


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/docs/agent-runtime-detail-handoff.md:
  Line 97: | Authentication | `agentCard.authenticationMethods` | Derived from `transportSecurity` (mTLS), AuthBridge mode, and `supportsAuthenticatedExtendedCard` |

  Line 258: - **Unknown auth:** return empty array when card has no AuthBridge/mTLS signals; otherwise derive methods.


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry_test.go:
  Line 24: 	created, err := registry.Register(context.Background(), GatewayConfig{ID: "openshell"})

  Line 26: 		t.Fatalf("register gateway: %v", err)

  Line 29: 		t.Fatal("gateway was not registered")

  Line 32: 	router, err := NewRouter("/agent-ops/api/openshell", registry)

  Line 40: 		"/agent-ops/api/openshell/openshell/v1/models%2Fmy-model?include=details",


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/discovery/discovery.go:
  Line 20: func PollGateways(ctx context.Context, gatewayRegistry fleet.Registry, interval time.Duration) {

  Line 29: 			// find gateways

  Line 30: 			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{

  Line 31: 				ID:       "derxu-openshell",

  Line 35: 			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{

  Line 36: 				ID:       "openshell-2",


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/openapi/src/agent-ops.yaml:
  Line 124:         labeled `openshell.ai/managed-by=openshell`.

  Line 834:           description: Mutual TLS mode for the agent.


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/mapper/agent_test.go:
  Line 19: // 				agents.LabelOpenShellManagedBy: agents.OpenShellManagedByValue,

  Line 60: // func TestAgentDetailToRuntimeDetail_OpenShellDefaultsResourceType(t *testing.T) {

  Line 63: // 			Name:      "openshell-agent",

  Line 64: // 			Namespace: "openshell-ns",

  Line 66: // 				agents.LabelOpenShellManagedBy: agents.OpenShellManagedByValue,

  Line 67: // 				agents.LabelOpenShellSandboxID: "uuid-123",


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/config/environment.go:
  Line 82: 	// certificates to the HTTP client's Root CAs for outbound TLS connections.

  Line 99: 	// ─── TLS ────────────────────────────────────────────────────

  Line 100: 	// TLS verification settings for HTTP client connections to the Client

  Line 101: 	// InsecureSkipVerify when true, skips TLS certificate verification (useful for development/local setups)


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/mapper/agent_card.go:
  Line 183: // 		add("mTLS")


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/mapper/agent_card_test.go:
  Line 109: // 	assert.Equal(t, []string{"mTLS"}, discoverAuthenticationMethods(&agents.AgentCardObserved{

  Line 117: // 	assert.Equal(t, []string{"mTLS", "Bearer"}, discoverAuthenticationMethods(&agents.AgentCardObserved{


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/clients/gateway.go:
  Line 9: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"

  Line 10: 	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"

  Line 13: // GatewayClients are the per-gateway clients an App needs: the SDK client for

  Line 14: // the OpenShell API, and the raw-exec client for binary-safe uploads.

  Line 16: // One set per gateway. A host fronting several gateways builds several — the

  Line 20: 	SDK        openshell.ClientInterface

  Line 41: // NewGatewayClients dials one OpenShell gateway.

  Line 43: // gatewayURL accepts grpc://, grpcs://, http://, https:// or a bare host:port;

  Line 44: // TLS is inferred from the scheme. Certificate paths are optional, and mTLS

  Line 46: func NewGatewayClients(gatewayURL, caCert, clientCert, clientKey string) (*GatewayClients, error) {

  Line 47: 	useTLS := strings.HasPrefix(gatewayURL, "grpcs://") || strings.HasPrefix(gatewayURL, "https://")

  Line 48: 	address := NormalizeGatewayAddress(gatewayURL, useTLS)

  Line 51: 		return nil, fmt.Errorf("gateway mTLS requires both a client certificate and key")

  Line 54: 	cfg := openshell.Config{

  Line 56: 		// Auth:    clients.ContextAuthProvider{RequireTLS: useTLS},

  Line 57: 		Auth: openshell.NoAuth(),

  Line 59: 	if useTLS {

  Line 60: 		tlsCfg := &openshell.TLSConfig{CAFile: caCert}

  Line 65: 		cfg.TLS = tlsCfg

  Line 67: 		cfg.TLS = &openshell.TLSConfig{Insecure: true}

  Line 70: 	sdkClient, err := openshell.NewClient(cfg)

  Line 76: 	uploadExec, err := clients.NewRawExecClient(rawHost, caCert, clientCert, clientKey, useTLS)

  Line 87: // NormalizeGatewayAddress maps the accepted gateway URL forms onto the http(s)

  Line 89: func NormalizeGatewayAddress(gatewayURL string, useTLS bool) string {

  Line 91: 	case strings.HasPrefix(gatewayURL, "grpcs://"):

  Line 92: 		return "https://" + strings.TrimPrefix(gatewayURL, "grpcs://")

  Line 93: 	case strings.HasPrefix(gatewayURL, "grpc://"):

  Line 94: 		return "http://" + strings.TrimPrefix(gatewayURL, "grpc://")

  Line 95: 	case strings.HasPrefix(gatewayURL, "https://"), strings.HasPrefix(gatewayURL, "http://"):

  Line 96: 		return gatewayURL

  Line 99: 		if useTLS {

  Line 102: 		return fmt.Sprintf("%s://%s", scheme, gatewayURL)


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/main.go:
  Line 23: const fleetRouterPrefix = "/agent-ops/api/openshell"

  Line 30: 	flag.StringVar(&certFile, "cert-file", "", "Path to TLS certificate file")

  Line 31: 	flag.StringVar(&keyFile, "key-file", "", "Path to TLS key file")

  Line 43: 	// bundle-paths accepts a comma-separated list of CA bundle file paths to trust for outbound TLS.

  Line 46: 	flag.Func("bundle-paths", "Comma-separated list of PEM CA bundle file paths to trust for outbound TLS (optional)", newOriginParser(&cfg.BundlePaths, defaultBundlePaths))

  Line 51: 	// TLS configuration flags

  Line 52: 	flag.BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-verify", getEnvAsBool("INSECURE_SKIP_VERIFY", false), "Skip TLS certificate verification (useful for development, default: false)")

  Line 97: 	fleetRegistry := fleet.NewRegistry(gatewayInstanceFactory)

  Line 107: 	router.Get("/agent-ops/api/openshell/gateways", func(w http.ResponseWriter, r *http.Request) {

  Line 108: 		// list gateways

  Line 129: 		tlsCfg, err := tlsprofile.ServerTLSConfig(context.Background(), logger)

  Line 131: 			logger.Error("failed to resolve TLS configuration from cluster profile", "error", err)

  Line 134: 		srv.TLSConfig = tlsCfg

  Line 139: 		logger.Info("starting server", "addr", srv.Addr, "TLS enabled", (certFile != "" && keyFile != ""))

  Line 142: 			err = srv.ListenAndServeTLS(certFile, keyFile)

  Line 151: 	// gateway discovery service


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/gateway_setup.go:
  Line 10: 	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"

  Line 14: 	openshellauth "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/auth"

  Line 15: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"

  Line 16: 	openshellmodels "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/models"

  Line 17: 	openshellapi "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/server"

  Line 25: type gatewayClients struct {

  Line 26: 	sdk        openshell.ClientInterface


(Results truncated. Consider using a more specific path or pattern.)
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/httpclient/http.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/httpclient/http.go</path>
<type>file</type>
<content>
1: package mrserver
2: 
3: import (
4: 	"crypto/tls"
5: 	"encoding/json"
6: 	"fmt"
7: 	"io"
8: 	"log/slog"
9: 	"net/http"
10: 	"strconv"
11: 
12: 	"github.com/google/uuid"
13: 	helper "github.com/opendatahub-io/agent-ops/internal/helpers"
14: )
15: 
16: type HTTPClientInterface interface {
17: 	GET(url string) ([]byte, error)
18: 	POST(url string, body io.Reader) ([]byte, error)
19: 	PATCH(url string, body io.Reader) ([]byte, error)
20: }
21: 
22: type HTTPClient struct {
23: 	client    *http.Client
24: 	baseURL   string
25: 	RequestID string
26: 	logger    *slog.Logger
27: 	Headers   http.Header
28: }
29: 
30: type ErrorResponse struct {
31: 	Code    string `json:"code"`
32: 	Message string `json:"message"`
33: }
34: 
35: type HTTPError struct {
36: 	StatusCode int `json:"-"`
37: 	ErrorResponse
38: }
39: 
40: func (e *HTTPError) Error() string {
41: 	return fmt.Sprintf("HTTP %d: %s - %s", e.StatusCode, e.Code, e.Message)
42: }
43: 
44: func NewHTTPClient(logger *slog.Logger, RequestID string, baseURL string, headers http.Header, insecureSkipVerify bool) (HTTPClientInterface, error) {
45: 	return &HTTPClient{
46: 		client: &http.Client{Transport: &http.Transport{
47: 			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
48: 		}},
49: 		baseURL:   baseURL,
50: 		RequestID: RequestID,
51: 		logger:    logger,
52: 		Headers:   headers,
53: 	}, nil
54: }
55: 
56: func (c *HTTPClient) GetRequestID() string {
57: 	return c.RequestID
58: }
59: 
60: func (c *HTTPClient) GET(url string) ([]byte, error) {
61: 	requestId := uuid.NewString()
62: 
63: 	fullURL := c.baseURL + url
64: 	req, err := http.NewRequest("GET", fullURL, nil)
65: 	if err != nil {
66: 		return nil, err
67: 	}
68: 
69: 	c.applyHeaders(req)
70: 
71: 	logUpstreamReq(c.logger, requestId, req)
72: 
73: 	response, err := c.client.Do(req)
74: 	if err != nil {
75: 		return nil, err
76: 	}
77: 	defer response.Body.Close()
78: 
79: 	body, err := io.ReadAll(response.Body)
80: 	logUpstreamResp(c.logger, requestId, response, body)
81: 	if err != nil {
82: 		return nil, fmt.Errorf("error reading response body: %w", err)
83: 	}
84: 
85: 	if response.StatusCode != http.StatusOK {
86: 		var errorResponse ErrorResponse
87: 		if err := json.Unmarshal(body, &errorResponse); err != nil {
88: 			// If we can't unmarshal as JSON, create a generic error response with the raw body
89: 			c.logger.Warn("received non-JSON error response",
90: 				"status_code", response.StatusCode,
91: 				"content_type", response.Header.Get("Content-Type"),
92: 				"body_preview", string(body[:min(len(body), 200)]))
93: 
94: 			errorResponse = ErrorResponse{
95: 				Code:    strconv.Itoa(response.StatusCode),
96: 				Message: fmt.Sprintf("HTTP %d: %s", response.StatusCode, string(body)),
97: 			}
98: 		}
99: 		httpError := &HTTPError{
100: 			StatusCode:    response.StatusCode,
101: 			ErrorResponse: errorResponse,
102: 		}
103: 
104: 		if httpError.Code == "" {
105: 			httpError.Code = strconv.Itoa(response.StatusCode)
106: 		}
107: 		return nil, httpError
108: 	}
109: 
110: 	return body, nil
111: }
112: 
113: func (c *HTTPClient) POST(url string, body io.Reader) ([]byte, error) {
114: 	requestId := uuid.NewString()
115: 
116: 	fullURL := c.baseURL + url
117: 	req, err := http.NewRequest("POST", fullURL, body)
118: 	if err != nil {
119: 		return nil, err
120: 	}
121: 
122: 	req.Header.Set("Content-Type", "application/json")
123: 
124: 	c.applyHeaders(req)
125: 
126: 	logUpstreamReq(c.logger, requestId, req)
127: 
128: 	response, err := c.client.Do(req)
129: 	if err != nil {
130: 		return nil, err
131: 	}
132: 	defer response.Body.Close()
133: 
134: 	responseBody, err := io.ReadAll(response.Body)
135: 	logUpstreamResp(c.logger, requestId, response, responseBody)
136: 	if err != nil {
137: 		return nil, fmt.Errorf("error reading response body: %w", err)
138: 	}
139: 
140: 	if response.StatusCode != http.StatusCreated {
141: 		var errorResponse ErrorResponse
142: 		if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
143: 			// If we can't unmarshal as JSON, create a generic error response with the raw body
144: 			c.logger.Warn("received non-JSON error response",
145: 				"status_code", response.StatusCode,
146: 				"content_type", response.Header.Get("Content-Type"),
147: 				"body_preview", string(responseBody[:min(len(responseBody), 200)]))
148: 
149: 			errorResponse = ErrorResponse{
150: 				Code:    strconv.Itoa(response.StatusCode),
151: 				Message: fmt.Sprintf("HTTP %d: %s", response.StatusCode, string(responseBody)),
152: 			}
153: 		}
154: 		httpError := &HTTPError{
155: 			StatusCode:    response.StatusCode,
156: 			ErrorResponse: errorResponse,
157: 		}
158: 
159: 		if httpError.Code == "" {
160: 			httpError.Code = strconv.Itoa(response.StatusCode)
161: 		}
162: 		return nil, httpError
163: 	}
164: 
165: 	return responseBody, nil
166: }
167: 
168: func (c *HTTPClient) PATCH(url string, body io.Reader) ([]byte, error) {
169: 	fullURL := c.baseURL + url
170: 	req, err := http.NewRequest(http.MethodPatch, fullURL, body)
171: 	if err != nil {
172: 		return nil, err
173: 	}
174: 
175: 	requestId := uuid.NewString()
176: 
177: 	req.Header.Set("Content-Type", "application/json")
178: 
179: 	c.applyHeaders(req)
180: 
181: 	logUpstreamReq(c.logger, requestId, req)
182: 
183: 	response, err := c.client.Do(req)
184: 	if err != nil {
185: 		return nil, err
186: 	}
187: 	defer response.Body.Close()
188: 
189: 	responseBody, err := io.ReadAll(response.Body)
190: 	logUpstreamResp(c.logger, requestId, response, responseBody)
191: 	if err != nil {
192: 		return nil, fmt.Errorf("error reading response body: %w", err)
193: 	}
194: 
195: 	if response.StatusCode != http.StatusOK {
196: 		var errorResponse ErrorResponse
197: 		if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
198: 			// If we can't unmarshal as JSON, create a generic error response with the raw body
199: 			c.logger.Warn("received non-JSON error response",
200: 				"status_code", response.StatusCode,
201: 				"content_type", response.Header.Get("Content-Type"),
202: 				"body_preview", string(responseBody[:min(len(responseBody), 200)]))
203: 
204: 			errorResponse = ErrorResponse{
205: 				Code:    strconv.Itoa(response.StatusCode),
206: 				Message: fmt.Sprintf("HTTP %d: %s", response.StatusCode, string(responseBody)),
207: 			}
208: 		}
209: 		httpError := &HTTPError{
210: 			StatusCode:    response.StatusCode,
211: 			ErrorResponse: errorResponse,
212: 		}
213: 
214: 		if httpError.Code == "" {
215: 			httpError.Code = strconv.Itoa(response.StatusCode)
216: 		}
217: 		return nil, httpError
218: 	}
219: 	return responseBody, nil
220: }
221: 
222: func (c *HTTPClient) applyHeaders(req *http.Request) {
223: 	if c.Headers != nil {
224: 		for key, values := range c.Headers {
225: 			for _, value := range values {
226: 				req.Header.Add(key, value)
227: 			}
228: 		}
229: 	}
230: }
231: 
232: func logUpstreamReq(logger *slog.Logger, reqId string, req *http.Request) {
233: 	logger.Debug("Making upstream HTTP request", slog.String("request_id", reqId), slog.Any("request", helper.RequestLogValuer{Request: req}))
234: }
235: 
236: func logUpstreamResp(logger *slog.Logger, reqId string, resp *http.Response, body []byte) {
237: 	logger.Debug("Received upstream HTTP response", slog.String("request_id", reqId), slog.Any("response", helper.ResponseLogValuer{Response: resp}))
238: }

(End of file - total 238 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/url.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/url.go</path>
<type>file</type>
<content>
1: package agents
2: 
3: import (
4: 	"net/url"
5: 	pathpkg "path"
6: 	"strings"
7: )
8: 
9: // BuildSanitizedHTTPURL assembles an http/https URL with a normalized path.
10: func BuildSanitizedHTTPURL(scheme, host, rawPath string) string {
11: 	scheme = strings.TrimSpace(scheme)
12: 	host = strings.TrimSpace(host)
13: 	rawPath = strings.TrimSpace(rawPath)
14: 	if host == "" {
15: 		return ""
16: 	}
17: 	if strings.Contains(rawPath, "..") {
18: 		return ""
19: 	}
20: 
21: 	normalizedPath := rawPath
22: 	if normalizedPath == "" {
23: 		normalizedPath = "/"
24: 	} else if !strings.HasPrefix(normalizedPath, "/") {
25: 		normalizedPath = "/" + normalizedPath
26: 	}
27: 	normalizedPath = pathpkg.Clean(normalizedPath)
28: 	if normalizedPath == "." {
29: 		normalizedPath = "/"
30: 	}
31: 
32: 	parsed := &url.URL{
33: 		Scheme: scheme,
34: 		Host:   host,
35: 		Path:   normalizedPath,
36: 	}
37: 	return SanitizeHTTPURL(parsed.String())
38: }
39: 
40: // SanitizeHTTPURL returns the URL only when it uses an http or https scheme.
41: func SanitizeHTTPURL(raw string) string {
42: 	raw = strings.TrimSpace(raw)
43: 	if raw == "" {
44: 		return ""
45: 	}
46: 
47: 	parsed, err := url.Parse(raw)
48: 	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
49: 		return ""
50: 	}
51: 
52: 	switch strings.ToLower(parsed.Scheme) {
53: 	case "http", "https":
54: 		if parsed.User != nil {
55: 			return ""
56: 		}
57: 		return raw
58: 	default:
59: 		return ""
60: 	}
61: }
62: 
63: // SanitizeResourceURI returns safe http/https agent-card URLs and urn extension identifiers.
64: func SanitizeResourceURI(raw string) string {
65: 	raw = strings.TrimSpace(raw)
66: 	if raw == "" {
67: 		return ""
68: 	}
69: 
70: 	parsed, err := url.Parse(raw)
71: 	if err != nil || parsed.Scheme == "" {
72: 		return ""
73: 	}
74: 	if parsed.User != nil {
75: 		return ""
76: 	}
77: 
78: 	switch strings.ToLower(parsed.Scheme) {
79: 	case "http", "https":
80: 		if parsed.Host == "" {
81: 			return ""
82: 		}
83: 		return raw
84: 	case "urn":
85: 		return raw
86: 	default:
87: 		return ""
88: 	}
89: }

(End of file - total 89 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/client.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/client.go</path>
<type>file</type>
<content>
1: package agents
2: 
3: import "context"
4: 
5: // Client loads agent runtime data from Sandbox CRs or mocks.
6: type Client interface {
7: 	// ListNamespaces returns namespaces where the caller can list agents.
8: 	// enabledOnly is reserved for future filtering and is currently ignored.
9: 	ListNamespaces(ctx context.Context, enabledOnly bool) ([]string, error)
10: 	ListAgents(ctx context.Context, namespace string) (*AgentList, error)
11: 	GetAgent(ctx context.Context, namespace, name string) (*AgentDetail, error)
12: 
13: 	DeployAgent(ctx context.Context, params *DeployAgentParams) (*DeployAgentResult, error)
14: 	DeleteAgent(ctx context.Context, namespace, name string) error
15: 	StopAgent(ctx context.Context, namespace, name string) error
16: 	StartAgent(ctx context.Context, namespace, name string) error
17: 	RestartAgent(ctx context.Context, namespace, name string) error
18: }
19: 
20: // ClientFactory creates a Client for the current request (e.g. with caller identity from context).
21: type ClientFactory interface {
22: 	GetClient(ctx context.Context) (Client, error)
23: }

(End of file - total 23 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/client.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/agents/kubernetes/client.go</path>
<type>file</type>
<content>
1: package kubernetes
2: 
3: import (
4: 	"context"
5: 	"errors"
6: 	"fmt"
7: 	"log/slog"
8: 
9: 	"github.com/opendatahub-io/agent-ops/internal/constants"
10: 	"github.com/opendatahub-io/agent-ops/internal/integrations/agents"
11: 	k8s "github.com/opendatahub-io/agent-ops/internal/integrations/kubernetes"
12: 	corev1 "k8s.io/api/core/v1"
13: 	apierrors "k8s.io/apimachinery/pkg/api/errors"
14: 	"k8s.io/apimachinery/pkg/api/meta"
15: 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
16: )
17: 
18: // Client reads agent Sandbox CRs from the Kubernetes API.
19: type Client struct {
20: 	k8sClient k8s.KubernetesClientInterface
21: 	identity  *k8s.RequestIdentity
22: 	logger    *slog.Logger
23: }
24: 
25: // ListNamespaces returns namespaces visible to the caller.
26: // Sandbox list access is not pre-checked here; forbidden namespaces are
27: // skipped when ListAgents handles K8s 403 responses directly.
28: func (c *Client) ListNamespaces(ctx context.Context, enabledOnly bool) ([]string, error) {
29: 	_ = enabledOnly
30: 
31: 	namespaces, err := c.k8sClient.GetNamespaces(ctx, c.identity)
32: 	if err != nil {
33: 		return nil, fmt.Errorf("failed to list namespaces for agent discovery: %w", err)
34: 	}
35: 
36: 	result := make([]string, 0, len(namespaces))
37: 	for _, namespace := range namespaces {
38: 		result = append(result, namespace.Name)
39: 	}
40: 
41: 	return result, nil
42: }
43: 
44: // ListAgents returns labeled agent sandboxes in a namespace.
45: func (c *Client) ListAgents(ctx context.Context, namespace string) (*agents.AgentList, error) {
46: 	summaries, err := c.listAgentSummaries(ctx, namespace)
47: 	if err != nil {
48: 		return nil, err
49: 	}
50: 	return &agents.AgentList{Items: summaries}, nil
51: }
52: 
53: // GetAgent returns detailed sandbox information for one agent.
54: func (c *Client) GetAgent(ctx context.Context, namespace, name string) (*agents.AgentDetail, error) {
55: 	return c.getAgentDetail(ctx, namespace, name)
56: }
57: 
58: func (c *Client) listAgentSummaries(ctx context.Context, namespace string) ([]agents.AgentSummary, error) {
59: 	dynamicClient, err := c.k8sClient.DynamicClient()
60: 	if err != nil {
61: 		return nil, fmt.Errorf("failed to get dynamic client: %w", err)
62: 	}
63: 
64: 	seen := make(map[string]struct{})
65: 	summaries := make([]agents.AgentSummary, 0)
66: 	var listed bool
67: 	var deferredErr error
68: 
69: 	for _, selector := range agentLabelSelectors() {
70: 		sandboxes, err := listSandboxes(ctx, dynamicClient, namespace, selector)
71: 		if err != nil {
72: 			if meta.IsNoMatchError(err) || apierrors.IsNotFound(err) {
73: 				continue
74: 			}
75: 			if apierrors.IsForbidden(err) {
76: 				if deferredErr == nil {
77: 					deferredErr = agents.ErrForbidden
78: 				}
79: 				c.logger.Debug("skipping sandbox list due to forbidden access",
80: 					slog.String("namespace", namespace),
81: 					slog.String("selector", selector))
82: 				continue
83: 			}
84: 			if isRetryableListError(err) {
85: 				if deferredErr == nil {
86: 					deferredErr = err
87: 				}
88: 				c.logger.Warn("skipping sandbox list due to transient error",
89: 					slog.String("namespace", namespace),
90: 					slog.String("selector", selector),
91: 					slog.Any("error", err))
92: 				continue
93: 			}
94: 			return nil, fmt.Errorf("failed to list sandboxes in namespace %q with selector %q: %w", namespace, selector, err)
95: 		}
96: 		listed = true
97: 
98: 		for _, sandbox := range sandboxes {
99: 			name := sandbox.GetName()
100: 			if _, exists := seen[name]; exists {
101: 				continue
102: 			}
103: 			seen[name] = struct{}{}
104: 			service := mapService(c.getServiceBestEffort(ctx, namespace, name))
105: 			summaries = append(summaries, sandboxToSummary(sandbox, service))
106: 		}
107: 	}
108: 
109: 	if !listed && deferredErr != nil {
110: 		if errors.Is(deferredErr, agents.ErrForbidden) {
111: 			return nil, agents.ErrForbidden
112: 		}
113: 		return nil, fmt.Errorf("failed to list sandboxes in namespace %q: %w", namespace, deferredErr)
114: 	}
115: 
116: 	return summaries, nil
117: }
118: 
119: func (c *Client) getAgentDetail(ctx context.Context, namespace, name string) (*agents.AgentDetail, error) {
120: 	dynamicClient, err := c.k8sClient.DynamicClient()
121: 	if err != nil {
122: 		return nil, fmt.Errorf("failed to get dynamic client: %w", err)
123: 	}
124: 
125: 	obj, err := getSandbox(ctx, dynamicClient, namespace, name)
126: 	if err != nil {
127: 		if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
128: 			return nil, agents.ErrNotFound
129: 		}
130: 		if apierrors.IsForbidden(err) {
131: 			return nil, agents.ErrForbidden
132: 		}
133: 		return nil, fmt.Errorf("failed to get sandbox %q in namespace %q: %w", name, namespace, err)
134: 	}
135: 
136: 	service := mapService(c.getServiceBestEffort(ctx, namespace, name))
137: 	detail := sandboxToDetail(*obj, service)
138: 	return detail, nil
139: }
140: 
141: func (c *Client) getServiceBestEffort(ctx context.Context, namespace, name string) *corev1.Service {
142: 	clientset := c.k8sClient.KubernetesClientset()
143: 	service, err := clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
144: 	if err != nil {
145: 		if !apierrors.IsNotFound(err) {
146: 			c.logger.Debug("failed to get service for agent",
147: 				slog.String("namespace", namespace),
148: 				slog.String("name", name),
149: 				slog.Any("error", err))
150: 		}
151: 		return nil
152: 	}
153: 	return service
154: }
155: 
156: func requestIdentityFromContext(ctx context.Context) (*k8s.RequestIdentity, error) {
157: 	identityVal := ctx.Value(constants.RequestIdentityKey)
158: 	if identityVal == nil {
159: 		return nil, errors.New("missing RequestIdentity in context")
160: 	}
161: 	identity, ok := identityVal.(*k8s.RequestIdentity)
162: 	if !ok || identity == nil {
163: 		return nil, errors.New("invalid RequestIdentity in context")
164: 	}
165: 	return identity, nil
166: }
167: 
168: func isRetryableListError(err error) bool {
169: 	return apierrors.IsServerTimeout(err) ||
170: 		apierrors.IsServiceUnavailable(err) ||
171: 		apierrors.IsTimeout(err) ||
172: 		apierrors.IsTooManyRequests(err)
173: }

(End of file - total 173 lines)
</content>
```

---

## Assistant (Build · GPT-5.6 Terra · 6.3s)

_Thinking:_

**Exploring gateway setup**

I need to focus on the gateway setup and confirm how external client packages come into play. I’ll look up the GOMODCACHE to see if it’s being used properly. I can use the Go environment in the terminal for this. It’s essential that I read the relevant README file for TLS configurations and hardcoded discovery. I might also need to inspect source dependencies through the Go module cache. Debugging with grep for imports could help identify agent ops or BFF packages.

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/gateway_setup.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/gateway_setup.go</path>
<type>file</type>
<content>
1: package main
2: 
3: import (
4: 	"context"
5: 	"fmt"
6: 	"log/slog"
7: 	"os"
8: 	"strings"
9: 
10: 	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
11: 	testclients "github.com/opendatahub-io/agent-ops/internal/clients"
12: 	"github.com/opendatahub-io/agent-ops/pkg/fleet"
13: 
14: 	openshellauth "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/auth"
15: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"
16: 	openshellmodels "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/models"
17: 	openshellapi "github.com/Gkrumbach07/openshell-dashboard/backend/pkg/server"
18: )
19: 
20: const (
21: 	defaultPort       = "8080"
22: 	defaultGatewayURL = "localhost:50051"
23: )
24: 
25: type gatewayClients struct {
26: 	sdk        openshell.ClientInterface
27: 	uploadExec *clients.RawExecClient
28: }
29: 
30: func (c *gatewayClients) Close() {
31: 	if c.sdk != nil {
32: 		if err := c.sdk.Close(); err != nil {
33: 			slog.Warn("SDK client close failed", "error", err)
34: 		}
35: 	}
36: 	if c.uploadExec != nil {
37: 		if err := c.uploadExec.Close(); err != nil {
38: 			slog.Warn("upload exec client close failed", "error", err)
39: 		}
40: 	}
41: }
42: 
43: func embeddedFeatureFlags() openshellmodels.FeatureFlags {
44: 	flags := openshellmodels.FeatureFlags{
45: 		Terminal:          true,
46: 		FileTransfer:      true,
47: 		Settings:          true,
48: 		GlobalPolicy:      true,
49: 		CredentialRefresh: true,
50: 		Services:          true,
51: 		DraftPolicy:       true,
52: 	}
53: 	// for _, name := range embeddedUnsupportedFeatures {
54: 	// 	if name == "terminal" {
55: 	// 		flags.Terminal = false
56: 	// 	}
57: 	// }
58: 	return flags
59: }
60: 
61: func gatewayInstanceFactory(ctx context.Context, cfg fleet.GatewayConfig) (fleet.GatewayInstance, error) {
62: 	appClients, err := testclients.NewGatewayClients(
63: 		cfg.Endpoint,
64: 		"",
65: 		"",
66: 		"",
67: 	)
68: 	if err != nil {
69: 		return fleet.GatewayInstance{Handler: nil, Close: nil}, err
70: 	}
71: 	app := openshellapi.NewApp(
72: 		appClients.SDK,
73: 		appClients.UploadExec,
74: 		openshellauth.New(openshellauth.Config{}),
75: 		"",
76: 		// TODO: remove statics
77: 		openshellmodels.AuthConfigResponse{
78: 			AdminRole:    "openshell-admin",
79: 			LogoutURL:    "",
80: 			Features:     embeddedFeatureFlags(),
81: 			AuthDisabled: true,
82: 		},
83: 	)
84: 	slog.Info(fmt.Sprintf("Created factory instance: %s", cfg.ID))
85: 	return fleet.GatewayInstance{
86: 		Close: func() error {
87: 			return nil
88: 		},
89: 		Handler: app.Routes(),
90: 	}, nil
91: }
92: 
93: func warnGatewayConfig(gatewayURL, gatewayCACert string, authDisabled bool) {
94: 	if gatewayURL == defaultGatewayURL {
95: 		slog.Warn("gateway URL is the default — verify OPENSHELL_GATEWAY_URL is configured correctly", "url", gatewayURL)
96: 	}
97: 	if gatewayCACert != "" && !strings.HasPrefix(gatewayURL, "grpcs://") && !strings.HasPrefix(gatewayURL, "https://") {
98: 		slog.Warn(
99: 			"gateway CA cert is set but gateway URL has no TLS scheme; use grpcs:// or https:// for TLS gateways",
100: 			"url", gatewayURL,
101: 			"caCert", gatewayCACert,
102: 		)
103: 	}
104: 	if authDisabled {
105: 		slog.Warn("AUTH_DISABLED=true — authentication is OFF; never use this outside local development")
106: 	}
107: }
108: 
109: func exitOnError(msg string, err error) {
110: 	slog.Error(msg, "error", err)
111: 	os.Exit(1)
112: }

(End of file - total 112 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/clients/gateway.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/clients/gateway.go</path>
<type>file</type>
<content>
1: // TODO: This whole file should be deleted sicne everything should be upstream and importable
2: package clients
3: 
4: import (
5: 	"fmt"
6: 	"log/slog"
7: 	"strings"
8: 
9: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"
10: 	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
11: )
12: 
13: // GatewayClients are the per-gateway clients an App needs: the SDK client for
14: // the OpenShell API, and the raw-exec client for binary-safe uploads.
15: //
16: // One set per gateway. A host fronting several gateways builds several — the
17: // clients differ only in endpoint, because ContextAuthProvider resolves the
18: // caller's bearer per request rather than binding one at construction.
19: type GatewayClients struct {
20: 	SDK        openshell.ClientInterface
21: 	UploadExec *clients.RawExecClient
22: }
23: 
24: // Close releases both clients, logging rather than failing on error.
25: func (c *GatewayClients) Close() {
26: 	if c == nil {
27: 		return
28: 	}
29: 	if c.SDK != nil {
30: 		if err := c.SDK.Close(); err != nil {
31: 			slog.Warn("SDK client close failed", "error", err)
32: 		}
33: 	}
34: 	if c.UploadExec != nil {
35: 		if err := c.UploadExec.Close(); err != nil {
36: 			slog.Warn("upload exec client close failed", "error", err)
37: 		}
38: 	}
39: }
40: 
41: // NewGatewayClients dials one OpenShell gateway.
42: //
43: // gatewayURL accepts grpc://, grpcs://, http://, https:// or a bare host:port;
44: // TLS is inferred from the scheme. Certificate paths are optional, and mTLS
45: // requires both a client certificate and key or neither.
46: func NewGatewayClients(gatewayURL, caCert, clientCert, clientKey string) (*GatewayClients, error) {
47: 	useTLS := strings.HasPrefix(gatewayURL, "grpcs://") || strings.HasPrefix(gatewayURL, "https://")
48: 	address := NormalizeGatewayAddress(gatewayURL, useTLS)
49: 
50: 	if (clientCert == "") != (clientKey == "") {
51: 		return nil, fmt.Errorf("gateway mTLS requires both a client certificate and key")
52: 	}
53: 
54: 	cfg := openshell.Config{
55: 		Address: address,
56: 		// Auth:    clients.ContextAuthProvider{RequireTLS: useTLS},
57: 		Auth: openshell.NoAuth(),
58: 	}
59: 	if useTLS {
60: 		tlsCfg := &openshell.TLSConfig{CAFile: caCert}
61: 		if clientCert != "" {
62: 			tlsCfg.CertFile = clientCert
63: 			tlsCfg.KeyFile = clientKey
64: 		}
65: 		cfg.TLS = tlsCfg
66: 	} else {
67: 		cfg.TLS = &openshell.TLSConfig{Insecure: true}
68: 	}
69: 
70: 	sdkClient, err := openshell.NewClient(cfg)
71: 	if err != nil {
72: 		return nil, fmt.Errorf("SDK client setup failed: %w", err)
73: 	}
74: 
75: 	rawHost := strings.TrimPrefix(strings.TrimPrefix(address, "https://"), "http://")
76: 	uploadExec, err := clients.NewRawExecClient(rawHost, caCert, clientCert, clientKey, useTLS)
77: 	if err != nil {
78: 		if closeErr := sdkClient.Close(); closeErr != nil {
79: 			slog.Warn("SDK client close failed during setup rollback", "error", closeErr)
80: 		}
81: 		return nil, fmt.Errorf("upload exec client setup failed: %w", err)
82: 	}
83: 
84: 	return &GatewayClients{SDK: sdkClient, UploadExec: uploadExec}, nil
85: }
86: 
87: // NormalizeGatewayAddress maps the accepted gateway URL forms onto the http(s)
88: // address the SDK expects.
89: func NormalizeGatewayAddress(gatewayURL string, useTLS bool) string {
90: 	switch {
91: 	case strings.HasPrefix(gatewayURL, "grpcs://"):
92: 		return "https://" + strings.TrimPrefix(gatewayURL, "grpcs://")
93: 	case strings.HasPrefix(gatewayURL, "grpc://"):
94: 		return "http://" + strings.TrimPrefix(gatewayURL, "grpc://")
95: 	case strings.HasPrefix(gatewayURL, "https://"), strings.HasPrefix(gatewayURL, "http://"):
96: 		return gatewayURL
97: 	default:
98: 		scheme := "http"
99: 		if useTLS {
100: 			scheme = "https"
101: 		}
102: 		return fmt.Sprintf("%s://%s", scheme, gatewayURL)
103: 	}
104: }

(End of file - total 104 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/discovery/discovery.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/discovery/discovery.go</path>
<type>file</type>
<content>
1: package discovery
2: 
3: import (
4: 	"context"
5: 	"log/slog"
6: 	"time"
7: 
8: 	"github.com/opendatahub-io/agent-ops/pkg/fleet"
9: )
10: 
11: func tempAddGateway(ctx context.Context, registry fleet.Registry, cfg fleet.GatewayConfig) error {
12: 	_, err := registry.Register(ctx, cfg)
13: 	if err != nil {
14: 		slog.Error(err.Error())
15: 		return err
16: 	}
17: 	return nil
18: }
19: 
20: func PollGateways(ctx context.Context, gatewayRegistry fleet.Registry, interval time.Duration) {
21: 	ticker := time.NewTicker(interval)
22: 
23: 	for {
24: 		select {
25: 		case <-ctx.Done():
26: 			// done
27: 			return
28: 		case <-ticker.C:
29: 			// find gateways
30: 			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{
31: 				ID:       "derxu-openshell",
32: 				Endpoint: "https://localhost:8080",
33: 			})
34: 
35: 			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{
36: 				ID:       "openshell-2",
37: 				Endpoint: "http://localhost:4000",
38: 			})
39: 		}
40: 	}
41: }

(End of file - total 41 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/main.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/main.go</path>
<type>file</type>
<content>
1: package main
2: 
3: import (
4: 	"context"
5: 	"flag"
6: 	"fmt"
7: 	"log/slog"
8: 	"net/http"
9: 	"os"
10: 	"os/signal"
11: 	"syscall"
12: 	"time"
13: 
14: 	"github.com/go-chi/chi"
15: 	"github.com/go-chi/chi/middleware"
16: 	"github.com/opendatahub-io/agent-ops/internal/config"
17: 	"github.com/opendatahub-io/agent-ops/internal/helpers"
18: 	"github.com/opendatahub-io/agent-ops/pkg/discovery"
19: 	"github.com/opendatahub-io/agent-ops/pkg/fleet"
20: 	tlsprofile "github.com/opendatahub-io/odh-dashboard/pkg/tls"
21: )
22: 
23: const fleetRouterPrefix = "/agent-ops/api/openshell"
24: 
25: func main() {
26: 	var cfg config.EnvConfig
27: 	var certFile, keyFile string
28: 
29: 	flag.IntVar(&cfg.Port, "port", getEnvAsInt("PORT", 4000), "API server port")
30: 	flag.StringVar(&certFile, "cert-file", "", "Path to TLS certificate file")
31: 	flag.StringVar(&keyFile, "key-file", "", "Path to TLS key file")
32: 	flag.BoolVar(&cfg.MockK8Client, "mock-k8s-client", false, "Use mock Kubernetes client")
33: 	flag.BoolVar(&cfg.MockHTTPClient, "mock-http-client", false, "Use mock HTTP client")
34: 	flag.BoolVar(&cfg.DevMode, "dev-mode", false, "Use development mode for access to local K8s cluster")
35: 	flag.IntVar(&cfg.DevModeClientPort, "dev-mode-client-port", getEnvAsInt("DEV_MODE_CLIENT_PORT", 8080), "Use port when in development mode for client")
36: 
37: 	// New deployment mode flag
38: 	flag.Var(&cfg.DeploymentMode, "deployment-mode", "Deployment mode (federated, standalone, or kubeflow alias for standalone)")
39: 
40: 	flag.StringVar(&cfg.StaticAssetsDir, "static-assets-dir", "./static", "Configure frontend static assets root directory")
41: 	flag.TextVar(&cfg.LogLevel, "log-level", parseLevel(getEnvAsString("LOG_LEVEL", "INFO")), "Sets server log level, possible values: error, warn, info, debug")
42: 	flag.Func("allowed-origins", "Sets allowed origins for CORS purposes, accepts a comma separated list of origins or * to allow all, default none", newOriginParser(&cfg.AllowedOrigins, getEnvAsString("ALLOWED_ORIGINS", "")))
43: 	// bundle-paths accepts a comma-separated list of CA bundle file paths to trust for outbound TLS.
44: 	// If not provided via flag, it can be set via BUNDLE_PATHS env var (comma-separated). Defaults to empty.
45: 	defaultBundlePaths := getEnvAsString("BUNDLE_PATHS", "")
46: 	flag.Func("bundle-paths", "Comma-separated list of PEM CA bundle file paths to trust for outbound TLS (optional)", newOriginParser(&cfg.BundlePaths, defaultBundlePaths))
47: 	flag.StringVar(&cfg.AuthMethod, "auth-method", "internal", "Authentication method (internal, user_token, or disabled)")
48: 	flag.StringVar(&cfg.AuthTokenHeader, "auth-token-header", getEnvAsString("AUTH_TOKEN_HEADER", config.DefaultAuthTokenHeader), "Header used to extract the token (e.g., Authorization)")
49: 	flag.StringVar(&cfg.AuthTokenPrefix, "auth-token-prefix", getEnvAsString("AUTH_TOKEN_PREFIX", config.DefaultAuthTokenPrefix), "Prefix used in the token header (e.g., 'Bearer ')")
50: 
51: 	// TLS configuration flags
52: 	flag.BoolVar(&cfg.InsecureSkipVerify, "insecure-skip-verify", getEnvAsBool("INSECURE_SKIP_VERIFY", false), "Skip TLS certificate verification (useful for development, default: false)")
53: 
54: 	// ─── BFF Inter-Communication ─────────────────────────────────
55: 	flag.BoolVar(&cfg.MockBFFClients, "mock-bff-clients",
56: 		getEnvAsBool("MOCK_BFF_CLIENTS", false),
57: 		"Enable mock BFF clients (no real HTTP calls to other BFFs)")
58: 	flag.BoolVar(&cfg.MockAgentClient, "mock-agent-client",
59: 		getEnvAsBool("MOCK_AGENT_CLIENT", false),
60: 		"Enable mock agent data client (demo data instead of Kubernetes; local development only)")
61: 
62: 	// Deprecated flags - kept for backward compatibility
63: 	flag.BoolVar(&cfg.StandaloneMode, "standalone-mode", false, "DEPRECATED: Use -deployment-mode=standalone instead")
64: 	flag.BoolVar(&cfg.FederatedPlatform, "federated-platform", false, "DEPRECATED: Use -deployment-mode=federated instead")
65: 
66: 	flag.Parse()
67: 
68: 	// Handle backward compatibility: if old flags are used, override deployment mode
69: 	if cfg.StandaloneMode {
70: 		cfg.DeploymentMode = config.DeploymentModeStandalone
71: 	} else if cfg.FederatedPlatform {
72: 		cfg.DeploymentMode = config.DeploymentModeFederated
73: 	}
74: 
75: 	// Ensure the deprecated boolean fields are consistent with the new deployment mode
76: 	cfg.StandaloneMode = cfg.DeploymentMode.IsStandaloneMode()
77: 	cfg.FederatedPlatform = cfg.DeploymentMode.IsFederatedMode()
78: 
79: 	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
80: 		Level: cfg.LogLevel,
81: 	}))
82: 
83: 	// validate auth method
84: 	if cfg.AuthMethod != config.AuthMethodInternal &&
85: 		cfg.AuthMethod != config.AuthMethodUser &&
86: 		cfg.AuthMethod != config.AuthMethodDisabled {
87: 		logger.Error("invalid auth method: (must be internal, user_token, or disabled)", "authMethod", cfg.AuthMethod)
88: 		os.Exit(1)
89: 	}
90: 	if cfg.AuthMethod == config.AuthMethodDisabled {
91: 		logger.Warn("AUTH_METHOD=disabled: all authentication and RBAC checks are bypassed — do not use in staging or production")
92: 	}
93: 
94: 	// Only use for logging errors about logging configuration.
95: 	slog.SetDefault(logger)
96: 
97: 	fleetRegistry := fleet.NewRegistry(gatewayInstanceFactory)
98: 	fleetHandler, err := fleet.NewRouter(fleetRouterPrefix, fleetRegistry)
99: 	if err != nil {
100: 		logger.Error("Unable to create fleet router", err.Error())
101: 		os.Exit(1)
102: 	}
103: 
104: 	// TODO: Move out of file and write custom middleware using slog defined above
105: 	router := chi.NewRouter()
106: 	router.Use(middleware.Logger)
107: 	router.Get("/agent-ops/api/openshell/gateways", func(w http.ResponseWriter, r *http.Request) {
108: 		// list gateways
109: 		if err := helpers.WriteJSON(w, http.StatusOK, "hello", nil); err != nil {
110: 			logger.Error("Failed to write JSON response",
111: 				slog.String("method", r.Method),
112: 				slog.String("path", r.URL.Path),
113: 				slog.Int("status", http.StatusOK),
114: 				slog.Any("error", err))
115: 		}
116: 	})
117: 	router.Mount(fleetRouterPrefix, fleetHandler)
118: 
119: 	srv := &http.Server{
120: 		Addr:         fmt.Sprintf(":%d", cfg.Port),
121: 		Handler:      router,
122: 		IdleTimeout:  time.Minute,
123: 		ReadTimeout:  30 * time.Second,
124: 		WriteTimeout: 30 * time.Second,
125: 		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
126: 	}
127: 
128: 	if certFile != "" && keyFile != "" {
129: 		tlsCfg, err := tlsprofile.ServerTLSConfig(context.Background(), logger)
130: 		if err != nil {
131: 			logger.Error("failed to resolve TLS configuration from cluster profile", "error", err)
132: 			os.Exit(1)
133: 		}
134: 		srv.TLSConfig = tlsCfg
135: 	}
136: 
137: 	// Start the server in a goroutine
138: 	go func() {
139: 		logger.Info("starting server", "addr", srv.Addr, "TLS enabled", (certFile != "" && keyFile != ""))
140: 		var err error
141: 		if certFile != "" && keyFile != "" {
142: 			err = srv.ListenAndServeTLS(certFile, keyFile)
143: 		} else {
144: 			err = srv.ListenAndServe()
145: 		}
146: 		if err != nil && err != http.ErrServerClosed {
147: 			logger.Error("HTTP server ListenAndServe", "error", err)
148: 		}
149: 	}()
150: 
151: 	// gateway discovery service
152: 	svcCtx, svcCancel := context.WithCancel(context.Background())
153: 	go discovery.PollGateways(svcCtx, fleetRegistry, 2*time.Second)
154: 
155: 	// Graceful shutdown setup
156: 	shutdownCh := make(chan os.Signal, 1)
157: 	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
158: 
159: 	// Wait for shutdown signal
160: 	<-shutdownCh
161: 	logger.Info("shutting down gracefully...")
162: 
163: 	svcCancel()
164: 
165: 	// Create a context with timeout for the shutdown process
166: 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
167: 	defer cancel()
168: 
169: 	// Shutdown the HTTP server gracefully
170: 	if err := srv.Shutdown(ctx); err != nil {
171: 		logger.Error("server shutdown failed", "error", err)
172: 	}
173: 	// Shutdown the App gracefully
174: 	if err := fleetRegistry.Close(ctx); err != nil {
175: 		logger.Error("failed to shutdown Kubernetes manager", "error", err)
176: 	}
177: 
178: 	logger.Info("server stopped")
179: 	os.Exit(0)
180: }

(End of file - total 180 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/README.md",
  "offset": 230,
  "limit": 75
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/README.md</path>
<type>file</type>
<content>
230: 
231: When serving the UI directly from the BFF there is no need for any CORS headers to be served, by default they are turned off for security reasons.
232: 
233: If you need to enable CORS for any reasons you can add origins to the allow-list in several ways:
234: 
235: ##### Via the make command
236: 
237: Add the following parameter to your command: `ALLOWED_ORIGINS` this takes a comma separated list of origins to permit serving to, alterantively you can specify the value `*` to allow all origins, **Note this is not recommended in production deployments as it poses a security risk**
238: 
239: Examples:
240: 
241: ```shell
242: # Allow only the origin http://example.com:8081
243: make run ALLOWED_ORIGINS="http://example.com:8081"
244: 
245: # Allow the origins http://example.com and http://very-nice.com
246: make run ALLOWED_ORIGINS="http://example.com,http://very-nice.com"
247: 
248: # Allow all origins
249: make run ALLOWED_ORIGINS="*"
250: 
251: # Explicitly disable CORS (default behaviour)
252: make run ALLOWED_ORIGINS=""
253: ```
254: 
255: #### Via environment variable
256: 
257: Setting CORS via environment variable follows the same rules as using the Makefile, simply set the environment variable `ALLOWED_ORIGINS` with the same value as above.
258: 
259: #### Via command line argument
260: 
261: Setting CORS via command line arguments follows the same rules as using the Makefile. Simply add the `--allowed-origins=` flag to your command.
262: 
263: Examples:
264: 
265: ```shell
266: ./bff --allowed-origins="http://my-domain.com,http://my-other-domain.com"
267: ```
268: 
269: ### Disabling TLS verification (development only)
270: 
271: For local Kubeflow installations with self-signed certificates, you may need to disable TLS certificate verification.
272: 
273: **Kubernetes deployment:**
274: 
275: ```yaml
276: env:
277:   - name: INSECURE_SKIP_VERIFY
278:     value: "true"
279: ```
280: 
281: **Local development:**
282: 
283: ```shell
284: ./bin/bff --insecure-skip-verify
285: # or
286: export INSECURE_SKIP_VERIFY=true
287: ```
288: 
289: > **Warning:** Only use in development. Keep TLS verification enabled in production.

(End of file - total 289 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/go.mod"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/go.mod</path>
<type>file</type>
<content>
1: module github.com/opendatahub-io/agent-ops
2: 
3: go 1.25.13
4: 
5: require (
6: 	github.com/Gkrumbach07/openshell-dashboard/backend v0.0.0-20260916180845-978bcb5eb6b5
7: 	github.com/NVIDIA/OpenShell/sdk/go v0.0.0-20260910203850-90dbe5454bb8
8: 	github.com/getkin/kin-openapi v0.149.0
9: 	github.com/go-chi/chi v1.5.5
10: 	github.com/google/uuid v1.6.0
11: 	github.com/julienschmidt/httprouter v1.3.0
12: 	github.com/opendatahub-io/odh-dashboard/pkg/tls v0.0.0-00010101000000-000000000000
13: 	github.com/rs/cors v1.11.1
14: 	github.com/stretchr/testify v1.12.1
15: 	k8s.io/api v0.34.3
16: 	k8s.io/apimachinery v0.34.3
17: 	k8s.io/client-go v0.34.3
18: 	sigs.k8s.io/controller-runtime v0.22.3
19: )
20: 
21: require (
22: 	github.com/beorn7/perks v1.0.1 // indirect
23: 	github.com/cespare/xxhash/v2 v2.3.0 // indirect
24: 	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
25: 	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
26: 	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
27: 	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
28: 	github.com/go-chi/chi/v5 v5.3.1 // indirect
29: 	github.com/go-logr/logr v1.4.3 // indirect
30: 	github.com/go-openapi/jsonpointer v0.22.5 // indirect
31: 	github.com/go-openapi/jsonreference v0.21.0 // indirect
32: 	github.com/go-openapi/swag v0.23.1 // indirect
33: 	github.com/go-openapi/swag/jsonname v0.25.5 // indirect
34: 	github.com/gogo/protobuf v1.3.2 // indirect
35: 	github.com/google/gnostic-models v0.7.0 // indirect
36: 	github.com/google/go-cmp v0.7.0 // indirect
37: 	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
38: 	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
39: 	github.com/josharian/intern v1.0.0 // indirect
40: 	github.com/json-iterator/go v1.1.12 // indirect
41: 	github.com/mailru/easyjson v0.9.0 // indirect
42: 	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
43: 	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
44: 	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
45: 	github.com/oasdiff/yaml v0.1.1 // indirect
46: 	github.com/oasdiff/yaml3 v0.0.14 // indirect
47: 	github.com/pkg/errors v0.9.1 // indirect
48: 	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
49: 	github.com/prometheus/client_golang v1.22.0 // indirect
50: 	github.com/prometheus/client_model v0.6.1 // indirect
51: 	github.com/prometheus/common v0.62.0 // indirect
52: 	github.com/prometheus/procfs v0.15.1 // indirect
53: 	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
54: 	github.com/spf13/pflag v1.0.6 // indirect
55: 	github.com/stretchr/objx v0.5.3 // indirect
56: 	github.com/x448/float16 v0.8.4 // indirect
57: 	go.yaml.in/yaml/v2 v2.4.2 // indirect
58: 	go.yaml.in/yaml/v3 v3.0.5 // indirect
59: 	golang.org/x/net v0.58.0 // indirect
60: 	golang.org/x/oauth2 v0.36.0 // indirect
61: 	golang.org/x/sync v0.22.0 // indirect
62: 	golang.org/x/sys v0.47.0 // indirect
63: 	golang.org/x/term v0.45.0 // indirect
64: 	golang.org/x/text v0.41.0 // indirect
65: 	golang.org/x/time v0.11.0 // indirect
66: 	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
67: 	google.golang.org/grpc v1.83.2 // indirect
68: 	google.golang.org/protobuf v1.36.11 // indirect
69: 	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
70: 	gopkg.in/inf.v0 v0.9.1 // indirect
71: 	gopkg.in/yaml.v3 v3.0.1 // indirect
72: 	k8s.io/apiextensions-apiserver v0.34.1 // indirect
73: 	k8s.io/klog/v2 v2.130.1 // indirect
74: 	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
75: 	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
76: 	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
77: 	sigs.k8s.io/randfill v1.0.0 // indirect
78: 	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
79: 	sigs.k8s.io/yaml v1.6.0 // indirect
80: )
81: 
82: replace github.com/opendatahub-io/odh-dashboard/pkg/tls => ../../../pkg/tls
83: 
84: replace github.com/Gkrumbach07/openshell-dashboard/backend => github.com/d0w/openshell-dashboard/backend v0.0.0-20260921141023-113961fac55c

(End of file - total 84 lines)
</content>
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://docs.nvidia.com/openshell/latest/kubernetes/openshift",
  "format": "markdown"
}
```

**Output:**
```
> For clean Markdown of any page, append .md to the page URL.
> For a complete documentation index, see https://docs.nvidia.com/openshell/llms.txt.
> For AI client integration (Claude Code, Cursor, etc.), connect to the MCP server at https://docs.nvidia.com/openshell/_mcp/server.

# OpenShift

> Install the OpenShell Helm chart on OpenShift, including the SCC binding and chart overrides required by OpenShift's Security Context Constraints.

The OpenShift install path is experimental. It currently requires running sandbox pods under the `privileged` SCC and installing the gateway with TLS disabled. Use only for evaluation on a private network.

OpenShift's [Security Context Constraints](https://docs.openshift.com/container-platform/latest/authentication/managing-security-context-constraints.html) reject the chart's default pod security settings. Installing on OpenShift requires precreating the namespace, granting the `privileged` SCC to the sandbox service account, and overriding a few chart values so the cluster admission controller can assign UIDs and FS groups itself.

OpenShell installs sandbox nftables rules as individual commands. On OpenShift
nodes where optional conntrack or packet log expressions are unavailable, those
optional rules can fail without rolling back the required proxy bypass reject
rules.

## Prerequisites

* OpenShift 4.x cluster with `oc` configured
* Helm 3.x
* [Agent Sandbox](/kubernetes/setup#install-agent-sandbox) controller and CRDs installed

## Install

## Create the namespace

Pre-create the namespace so the SCC binding can be applied before the chart installs:

```shell
oc create ns openshell
```

## Grant the privileged SCC to sandbox pods

Sandbox pods run under the `openshell-sandbox` service account in the `openshell` namespace and require the `privileged` SCC:

```shell
oc adm policy add-scc-to-user privileged -z openshell-sandbox -n openshell
```

## Install the chart with OpenShift overrides

```shell
helm install openshell oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set server.disableTls=true \
  --set podSecurityContext.fsGroup=null \
  --set securityContext.runAsUser=null
```

| Override                                                             | Reason                                                                                    |
| -------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| `server.disableTls=true`                                             | Runs the gateway over plaintext HTTP for simpler evaluation.                              |
| `podSecurityContext.fsGroup=null` / `securityContext.runAsUser=null` | Clear the chart's hardcoded UID and fsGroup so OpenShift's SCC admission can assign them. |

## Wait for the gateway to be ready

```shell
oc -n openshell rollout status statefulset/openshell
```

If you set `workload.kind=deployment`, use
`oc -n openshell rollout status deployment/openshell` instead.

## Connect to the gateway

The gateway is now running over plaintext HTTP. Connect with `oc port-forward`:

```shell
oc -n openshell port-forward svc/openshell 8080:8080
```

Register the gateway with the CLI:

```shell
openshell gateway add http://127.0.0.1:8080 --local --name openshift
openshell status
```

## Production: expose externally with a real certificate

The steps above run the gateway over plaintext HTTP for quick evaluation. For
a real deployment, cert-manager can issue the gateway's server certificate
from a real Issuer or ClusterIssuer (for example, an ACME issuer), and an
OpenShift Route with TLS passthrough exposes it externally while the gateway
keeps terminating its own TLS and mTLS.

Install cert-manager and configure a working `ClusterIssuer` first — see
[Managing Certificates](/kubernetes/managing-certificates) for the
`certManager.serverIssuerRef` details. Configure an OIDC provider as described
in [Access Control](/kubernetes/access-control) — remote gateways authenticate
CLI users via OIDC, not mTLS, so the gateway must know the OIDC issuer URL.
Install the chart with:

```shell
helm install openshell oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set podSecurityContext.fsGroup=null \
  --set securityContext.runAsUser=null \
  --set server.disableTls=false \
  --set certManager.enabled=true \
  --set certManager.serverIssuerRef.name=<cluster-issuer-name> \
  --set certManager.serverIssuerRef.kind=ClusterIssuer \
  --set certManager.serverDnsNames[0]=<external-hostname> \
  --set openshiftRoute.enabled=true \
  --set openshiftRoute.host=<external-hostname> \
  --set server.oidc.issuer=<oidc-issuer-url> \
  --set server.oidc.audience=<oidc-audience>
```

| Override                                         | Reason                                                                                                                                                                                                                                                                                                                                                                                       |
| ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `certManager.serverIssuerRef`                    | Creates a second server certificate from your Issuer or ClusterIssuer for external clients. The gateway uses SNI to present this cert for the external hostname while continuing to present the internal (chart CA) cert to supervisors. The internal certificate's `ca.crt` is the chart CA that also signed the client cert, so the default `clientCaFromServerTlsSecret=true` is correct. |
| `openshiftRoute.enabled` / `openshiftRoute.host` | Creates an OpenShift Route with TLS passthrough — the router forwards the encrypted connection by SNI without decrypting, so the gateway uses the SNI hostname to select the external certificate.                                                                                                                                                                                           |
| `server.oidc.issuer` / `server.oidc.audience`    | Configures server-side OIDC validation. Without these, the gateway expects mTLS client certificates and rejects OIDC-only CLI connections. See [Access Control](/kubernetes/access-control).                                                                                                                                                                                                 |

Register the gateway with the CLI over OIDC. Remote gateways authenticate CLI
users via OIDC, not mTLS — see [Access Control](/kubernetes/access-control):

```shell
openshell gateway add https://<external-hostname> \
  --name openshift \
  --oidc-issuer <issuer-url>
openshell gateway login openshift
```

## Next Steps

* For more on certificate provisioning modes, refer to [Managing Certificates](/kubernetes/managing-certificates).
* To expose the gateway externally through the Kubernetes Gateway API instead of a Route, refer to [Ingress](/kubernetes/ingress).
* To configure OIDC authentication, refer to [Access Control](/kubernetes/access-control).
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://docs.nvidia.com/openshell/latest/kubernetes/setup",
  "format": "markdown"
}
```

**Output:**
```
> For clean Markdown of any page, append .md to the page URL.
> For a complete documentation index, see https://docs.nvidia.com/openshell/llms.txt.
> For AI client integration (Claude Code, Cursor, etc.), connect to the MCP server at https://docs.nvidia.com/openshell/_mcp/server.

# Set Up OpenShell on Kubernetes

> Deploy the OpenShell gateway to a Kubernetes cluster using the official Helm chart from GHCR.

The OpenShell Helm chart is experimental and under active development. Templates, values, and defaults can change between releases. Do not use it in production.

Use the Kubernetes deployment when the gateway should run on a shared cluster, in a cloud environment, or as part of team infrastructure. The Helm chart handles PKI bootstrap, RBAC, sandbox namespace setup, and the gateway workload. It uses a StatefulSet by default for the SQLite database, and can render a Deployment when `server.externalDbSecret` points at an external database.

## Prerequisites

Make sure the following are in place before you install.

| Prerequisite                       | Required | Notes                                                                                                                                  |
| ---------------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| Kubernetes 1.29+ with RBAC enabled | Yes      | No additional notes.                                                                                                                   |
| Helm 3.x                           | Yes      | No additional notes.                                                                                                                   |
| Agent Sandbox controller and CRDs  | Yes      | Install before the OpenShell chart. Refer to [Install Agent Sandbox](#install-agent-sandbox).                                          |
| cert-manager                       | No       | Refer to [Managing Certificates](/kubernetes/managing-certificates). Use cert-manager only if you prefer it over the built-in PKI job. |
| Kubernetes Gateway API             | No       | Refer to [Ingress](/kubernetes/ingress). Use it only for external access without port-forwarding.                                      |

## Install Agent Sandbox

OpenShell uses the [Agent Sandbox](https://agent-sandbox.sigs.k8s.io) Kubernetes SIG project to provision sandbox pods. Install the Agent Sandbox controller and its CRDs on your cluster before installing the OpenShell Helm chart.

Apply the latest release manifest:

```shell
kubectl apply -f https://github.com/kubernetes-sigs/agent-sandbox/releases/latest/download/manifest.yaml
```

This creates the `agent-sandbox-system` namespace, installs the `sandboxes.agents.x-k8s.io` CRD, and starts the controller.

The Helm chart checks for a supported Agent Sandbox API before it creates
gateway resources. This preflight is enabled by default. Disable it only for
offline `helm template` rendering, where Helm cannot discover cluster APIs:

```shell
helm template openshell oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --set agentSandbox.preflight.enabled=false
```

The chart does not install or upgrade the cluster-scoped Agent Sandbox CRDs or
controller.

**Air-gapped clusters:** mirror the manifest above and the `registry.k8s.io/agent-sandbox/agent-sandbox-controller` image referenced inside it to your internal registry, then point the manifest's image reference at your mirror before applying. You will also need to mirror the OpenShell gateway and sandbox images — see the chart's `image.repository` value for the gateway and `server.sandboxImage` / `server.supervisorImage` for the sandbox runtime.

Confirm the controller pod is running before proceeding:

```shell
kubectl -n agent-sandbox-system get pods
```

The controller pod should reach `Running` status within a few seconds. For cluster-specific setup instructions, including KinD and GKE walkthroughs, refer to the [Agent Sandbox getting started guide](https://agent-sandbox.sigs.k8s.io/docs/getting_started/).

### Upgrade Agent Sandbox

OpenShell detects the served Agent Sandbox `Sandbox` API when the Kubernetes gateway first needs it and caches that choice for the gateway process. If you upgrade Agent Sandbox in place, restart the OpenShell gateway after the Agent Sandbox controller and CRD rollout completes so the gateway can detect the served API versions again. Existing sandboxes keep running during the upgrade, and the restarted gateway can continue managing them.

## Install OpenShell

## Create the namespace

```shell
kubectl create namespace openshell
```

## Install the chart

Install from the OCI registry on GHCR. Replace `<version>` with the chart version you want to install.

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell
```

To use the latest development build instead of a stable release:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version 0.0.0-dev \
  --namespace openshell
```

The chart automatically generates PKI secrets on first install using pre-install Helm hooks. No manual secret creation is required.

## Wait for the gateway to be ready

```shell
kubectl -n openshell rollout status statefulset/openshell
```

If you set `workload.kind=deployment`, wait on the Deployment instead:

```shell
kubectl -n openshell rollout status deployment/openshell
```

## Connect to the gateway

For local evaluation, use a port-forward:

```shell
kubectl -n openshell port-forward svc/openshell 8080:8080
```

The port-forward is for local evaluation only. For shared environments, expose the gateway through your ingress controller or access proxy. Refer to [Ingress](/kubernetes/ingress) for an external access option.

## Install the TLS client bundle

The chart generates an mTLS bundle for transport security. Kubernetes deployments do not use that bundle as user authentication; configure OIDC or a trusted access proxy as described in [Access Control](/kubernetes/access-control). For local port-forwarded access, copy the generated bundle so the CLI can verify the gateway certificate:

```shell
mkdir -p ~/.config/openshell/gateways/k8s/mtls
kubectl -n openshell get secret openshell-client-tls \
  -o jsonpath='{.data.ca\.crt}'  | base64 -d > ~/.config/openshell/gateways/k8s/mtls/ca.crt
kubectl -n openshell get secret openshell-client-tls \
  -o jsonpath='{.data.tls\.crt}' | base64 -d > ~/.config/openshell/gateways/k8s/mtls/tls.crt
kubectl -n openshell get secret openshell-client-tls \
  -o jsonpath='{.data.tls\.key}' | base64 -d > ~/.config/openshell/gateways/k8s/mtls/tls.key
```

The server certificate SANs include `localhost` and `127.0.0.1`, so hostname verification passes over the port-forward without extra flags.

## Register with the CLI

In another terminal, register the gateway with the user authentication mode you configured and verify it is reachable. For example, with OIDC:

```shell
openshell gateway add https://127.0.0.1:8080 --local --name k8s \
  --oidc-issuer https://your-idp.example.com/realms/openshell \
  --oidc-client-id openshell-cli
openshell status
```

## Configure Chart Values

The most commonly changed values are:

| Value                                                      | Purpose                                                                                                                                                                                                                                                                                                                                                                                                           |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `image.repository` / `image.tag`                           | Gateway container image. Defaults to `ghcr.io/nvidia/openshell/gateway:latest`.                                                                                                                                                                                                                                                                                                                                   |
| `replicaCount`                                             | Number of gateway replicas. Leave at `1` unless you are explicitly testing multi-replica behavior.                                                                                                                                                                                                                                                                                                                |
| `workload.kind`                                            | Gateway workload controller. Use `statefulset` for SQLite or `deployment` with `server.externalDbSecret`.                                                                                                                                                                                                                                                                                                         |
| `workload.allowMultiReplicaStatefulSet`                    | Allow `replicaCount > 1` with `workload.kind=statefulset`. Prefer Deployment for external database-backed multi-replica gateways.                                                                                                                                                                                                                                                                                 |
| `server.sandboxNamespace`                                  | Namespace where sandbox pods are created. Defaults to the Helm release namespace when left empty.                                                                                                                                                                                                                                                                                                                 |
| `server.externalDbSecret`                                  | Secret containing a PostgreSQL connection URI in the `uri` key. Use when the database is managed outside the chart.                                                                                                                                                                                                                                                                                               |
| `server.telemetryEnabled`                                  | Enable anonymous OpenShell telemetry from the gateway and its sandbox supervisors. Set to `false` to opt out.                                                                                                                                                                                                                                                                                                     |
| `server.sandboxImage`                                      | Default sandbox image used when a sandbox does not specify one.                                                                                                                                                                                                                                                                                                                                                   |
| `server.sandboxImagePullSecrets`                           | Image pull secrets attached to sandbox pods. Referenced Secrets must exist in the sandbox namespace.                                                                                                                                                                                                                                                                                                              |
| `server.grpcEndpoint`                                      | Endpoint that sandbox supervisors use to call back to the gateway. Must be reachable from inside the cluster.                                                                                                                                                                                                                                                                                                     |
| `server.appArmorProfile`                                   | AppArmor profile requested for sandbox agent containers. Defaults to `Unconfined`.                                                                                                                                                                                                                                                                                                                                |
| `server.disableTls`                                        | Run the gateway over plaintext HTTP. Use only behind a trusted transport.                                                                                                                                                                                                                                                                                                                                         |
| `server.auth.allowUnauthenticatedUsers`                    | Accept user-facing calls without OIDC or mTLS credentials. Use only for trusted local development or a fully trusted access proxy.                                                                                                                                                                                                                                                                                |
| `server.enableLoopbackServiceHttp`                         | Enable local plaintext HTTP for loopback sandbox service URLs. Defaults to `true`.                                                                                                                                                                                                                                                                                                                                |
| `pkiInitJob.serverDnsNames` / `certManager.serverDnsNames` | Additional gateway server DNS SANs. Wildcard SANs also enable sandbox service URLs under that domain.                                                                                                                                                                                                                                                                                                             |
| `supervisor.sideloadMethod`                                | How the supervisor binary is delivered into sandbox pods. Leave empty to auto-detect based on cluster version: clusters running Kubernetes 1.35 or later use `image-volume` (ImageVolume GA in 1.36); older clusters use `init-container`. Set explicitly to `image-volume` on Kubernetes 1.33 or 1.34 with the ImageVolume feature gate enabled, or to `init-container` to force the legacy path on any version. |
| `supervisor.topology`                                      | Sandbox pod topology. Refer to [Topology](/kubernetes/topology).                                                                                                                                                                                                                                                                                                                                                  |
| `supervisor.sidecar.proxyUid`                              | Non-root UID used when sidecar process/binary-aware network policy is disabled. The default binary-aware sidecar runs as UID 0 instead. The configured UID must not match the sandbox UID.                                                                                                                                                                                                                        |
| `upstreamProxy`                                            | Operator-owned corporate HTTP forward proxy for policy-approved TLS egress. Refer to [Configure a Corporate Upstream Proxy](#configure-a-corporate-upstream-proxy).                                                                                                                                                                                                                                               |

Use a values file for repeatable deployments:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --values my-values.yaml
```

The chart defaults `server.appArmorProfile` to `Unconfined` because
runtime/default AppArmor profiles can block the supervisor's network namespace
mount setup on AppArmor-enabled nodes. Set `server.appArmorProfile` to an empty
string to omit the field, `RuntimeDefault` to force the runtime default, or
`Localhost/<profile-name>` when you load and manage a localhost profile on each
node.

To use private sandbox images, create a `kubernetes.io/dockerconfigjson` Secret
in the sandbox namespace and reference its name:

```shell
kubectl -n openshell create secret docker-registry regcred \
  --docker-server=registry.example.com \
  --docker-username="$REGISTRY_USER" \
  --docker-password="$REGISTRY_TOKEN"
```

```yaml
server:
  sandboxImage: registry.example.com/team/openshell-sandbox:latest
  sandboxImagePullSecrets:
    - name: regcred
```

## Configure a Corporate Upstream Proxy

Configure a corporate forward proxy when sandbox TLS egress cannot dial the Internet directly. OpenShell evaluates policy and SSRF checks before it opens an HTTP CONNECT tunnel through the proxy. The proxy URL is operator-owned configuration. Sandbox environment variables cannot select, replace, or bypass it.

Create the credential Secret in the sandbox namespace when the proxy requires Basic authentication. The Secret value uses the `user:pass` form.

```shell
kubectl -n openshell create secret generic corporate-proxy-auth \
  --from-literal=credentials="$PROXY_USER:$PROXY_PASSWORD"
```

Add the proxy settings to your Helm values file. Replace the DNS suffixes and CIDRs in `noProxy` with values for your cluster. `noProxy` bypasses only the corporate proxy. OpenShell policy evaluation still applies.

```yaml
upstreamProxy:
  url: http://proxy.corp.example:8080
  noProxy: .svc,.svc.cluster.local,10.96.0.0/12,10.244.0.0/16
  authSecret:
    name: corporate-proxy-auth
    key: credentials
  authAllowInsecure: true

supervisor:
  topology: sidecar
```

Use `authAllowInsecure: true` only when you accept that Basic authentication is cleartext on the connection to an `http://` proxy. The initial release supports `http://` proxy endpoints and TLS CONNECT egress. It does not support HTTPS-to-proxy, custom corporate CA bundles, or forwarding plain HTTP egress through the proxy.

Proxy credentials require `sidecar` topology. It mounts the credential only into the dedicated network supervisor container. OpenShell rejects credential Secrets with `combined` topology because Kubernetes `fsGroup` volume permission handling can make a shared credential mount readable by the sandbox group.

## RBAC

The chart creates the following RBAC resources in the release namespace:

| Resource                         | Scope     | Name                                   |
| -------------------------------- | --------- | -------------------------------------- |
| ServiceAccount                   | Namespace | `openshell`                            |
| ServiceAccount                   | Namespace | `openshell-sandbox` (for sandbox pods) |
| Role + RoleBinding               | Namespace | `openshell-sandbox`                    |
| ClusterRole + ClusterRoleBinding | Cluster   | `openshell-node-reader`                |

The namespaced Role covers sandbox lifecycle and identity:

| API Group         | Resource                        | Verbs                                           |
| ----------------- | ------------------------------- | ----------------------------------------------- |
| `agents.x-k8s.io` | `sandboxes`, `sandboxes/status` | create, delete, get, list, patch, update, watch |
| `""`              | `events`                        | get, list, watch                                |
| `""`              | `pods`                          | get                                             |

The ClusterRole grants node inspection and token validation:

| API Group               | Resource       | Verbs            |
| ----------------------- | -------------- | ---------------- |
| `authentication.k8s.io` | `tokenreviews` | create           |
| `""`                    | `nodes`        | get, list, watch |

To use an existing ServiceAccount instead of creating one, set `serviceAccount.create=false` and supply its name:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set serviceAccount.create=false \
  --set serviceAccount.name=my-existing-sa
```

The ServiceAccount must already have the Role and ClusterRole bindings described above.

## Probes

The gateway exposes `/healthz` for process liveness and `/readyz` for dependency-aware readiness on the health port. The Helm chart wires both into Kubernetes probes:

* `startupProbe` and `livenessProbe` use `/healthz`.
* `readinessProbe` uses `/readyz`, which reflects the latest result of an in-process background database check.

## Next Steps

* To choose between combined and sidecar sandbox pods, refer to [Topology](/kubernetes/topology).
* To enable automatic certificate rotation with cert-manager, refer to [Managing Certificates](/kubernetes/managing-certificates).
* To expose the gateway externally without port-forwarding, refer to [Ingress](/kubernetes/ingress).
* To configure OIDC or reverse-proxy authentication, refer to [Access Control](/kubernetes/access-control).
* To create your first sandbox, refer to [Manage Sandboxes](/sandboxes/manage-sandboxes).
```

---

## Assistant (Build · GPT-5.6 Terra · 7.6s)

_Thinking:_

**Identifying operational issues**

I see there’s a bug with the gateway factory passing an empty CA, and the endpoint is hardcoded to HTTPS localhost:8080. The user wants to test port forwarding, so I need to provide them with exact instructions without any changes. I think I need to identify the expected RawExec client and check whether the CA works. Plus, inspecting the version dependency is crucial to validate the SDK's TLS semantics since it's likely coming from the cache. I’ll execute `go env GOMODCACHE`, which should be safe.

Key finding: gateway fleet path ignores TLS settings. Discovery registers `https://localhost:8080`, then factory passes empty CA/client-cert/key paths. OpenShell default port-forward uses gateway mTLS, so TLS handshake needs all three files from `openshell-client-tls`.

**Tool: bash**

**Input:**
```json
{
  "workdir": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "command": "go env GOMODCACHE",
  "timeout": 120000
}
```

**Output:**
```
/Users/derxu/go/pkg/mod

```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry.go</path>
<type>file</type>
<content>
1: package fleet
2: 
3: import (
4: 	"context"
5: 	"fmt"
6: 	"log/slog"
7: 	"net/http"
8: 	"sync"
9: )
10: 
11: type GatewayConfig struct {
12: 	ID       string
13: 	Endpoint string
14: }
15: 
16: type GatewayInstance struct {
17: 	Handler http.Handler
18: 	Close   func() error
19: }
20: 
21: // TODO: make package-agnostic. no gateway specific stuff just registry
22: // eventually change to interface to allow any object to be created
23: type GatewayFactory func(context.Context, GatewayConfig) (GatewayInstance, error)
24: 
25: type Registry interface {
26: 	ServeHTTP(http.ResponseWriter, *http.Request)
27: 	Register(context.Context, GatewayConfig) (bool, error)
28: 	Remove(context.Context, string) (bool, error)
29: 	Close(context.Context) error
30: }
31: 
32: type GatewayRegistry struct {
33: 	factory GatewayFactory
34: 
35: 	mu      sync.Mutex
36: 	entries map[string]*GatewayInstance
37: }
38: 
39: func NewRegistry(factory GatewayFactory) *GatewayRegistry {
40: 	return &GatewayRegistry{
41: 		factory: factory,
42: 		entries: make(map[string]*GatewayInstance),
43: 	}
44: }
45: 
46: func (r *GatewayRegistry) Register(ctx context.Context, config GatewayConfig) (bool, error) {
47: 	r.mu.Lock()
48: 	defer r.mu.Unlock()
49: 
50: 	_, exists := r.entries[config.ID]
51: 	if exists {
52: 		return false, nil
53: 	}
54: 
55: 	instance, err := r.factory(ctx, config)
56: 	if err != nil {
57: 		return false, fmt.Errorf("create gateway %q: %w", config.ID, err)
58: 	}
59: 
60: 	r.entries[config.ID] = &instance
61: 
62: 	return true, nil
63: }
64: 
65: func (r *GatewayRegistry) Remove(ctx context.Context, id string) (bool, error) {
66: 	r.mu.Lock()
67: 	defer r.mu.Unlock()
68: 
69: 	_, exists := r.entries[id]
70: 	if exists {
71: 		delete(r.entries, id)
72: 		return true, nil
73: 	}
74: 
75: 	return false, nil
76: }
77: 
78: func (r *GatewayRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
79: 	id, upstreamPath, upstreamRawPath, ok := splitGatewayPath(req.URL.Path, req.URL.EscapedPath())
80: 	slog.Info("HIT THE REGISTRY")
81: 	if !ok {
82: 		// return http error
83: 		return
84: 	}
85: 	// if !validGatewayID(id) {
86: 	// 	// return http error
87: 	// 	return
88: 	// }
89: 	fmt.Printf("%#v\n", r.entries)
90: 
91: 	gatewayInstance, ok := r.entries[id]
92: 	if !ok {
93: 		w.Write([]byte("Instance does not exist"))
94: 		return
95: 	}
96: 	fmt.Printf("%#v\n", gatewayInstance)
97: 
98: 	// clones request
99: 	// alter headers, auth, etc. here...
100: 	upstreamReq := req.Clone(req.Context())
101: 	upstreamReq.URL.Path = upstreamPath
102: 	upstreamReq.URL.RawPath = upstreamRawPath
103: 	upstreamReq.RequestURI = upstreamReq.URL.RequestURI()
104: 	gatewayInstance.Handler.ServeHTTP(w, upstreamReq)
105: }
106: 
107: func (r *GatewayRegistry) Close(ctx context.Context) error {
108: 	// close all gateway instances. be aware of race conditions
109: 	return nil
110: 	// errChan := make(chan error, 1)
111: 	// select {
112: 	// case <-ctx.Done():
113: 	// 	return nil
114: 	// case err := <-errChan:
115: 	// 	return err
116: 	// }
117: }

(End of file - total 117 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/router.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/router.go</path>
<type>file</type>
<content>
1: package fleet
2: 
3: import "net/http"
4: 
5: func NewRouter(urlPrefix string, registry Registry) (http.Handler, error) {
6: 	router := http.NewServeMux()
7: 	router.Handle("/", http.StripPrefix(urlPrefix, registry))
8: 	return router, nil
9: }

(End of file - total 9 lines)
</content>
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://docs.nvidia.com/openshell/latest/kubernetes/access-control",
  "format": "markdown"
}
```

**Output:**
```
> For clean Markdown of any page, append .md to the page URL.
> For a complete documentation index, see https://docs.nvidia.com/openshell/llms.txt.
> For AI client integration (Claude Code, Cursor, etc.), connect to the MCP server at https://docs.nvidia.com/openshell/_mcp/server.

# Access Control

> Configure OIDC user authentication or reverse-proxy auth termination for a Kubernetes-deployed OpenShell gateway.

The OpenShell gateway supports two access-control models for human callers on Kubernetes:

| Model                          | When to use                                                                                                                                                                          |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| OIDC (recommended)             | Production deployments. Integrates with an existing identity provider, supports role-based access control, and gives each user their own identity without distributing certificates. |
| Reverse-proxy auth termination | An access proxy (Cloudflare Access, ngrok, corporate SSO) authenticates callers in front of the gateway. The gateway trusts the proxy and skips its own client-cert check.           |

The Helm chart always generates mTLS certificates at install time. The gateway uses them for transport-layer security regardless of which access-control model you choose. The client bundle in the `openshell-client-tls` secret is used internally by sandbox supervisors, not for granting access to individual users.

For how the CLI resolves gateways and stores credentials, refer to [Gateway Authentication](/reference/gateway-auth).

## Sandbox Supervisor Identity

Kubernetes sandbox supervisors authenticate back to the gateway as sandbox workloads. By default, the gateway mints its own sandbox JWTs and Kubernetes sandboxes bootstrap them with a projected ServiceAccount token.

Dynamic provider token grants can use SPIFFE without changing supervisor-to-gateway authentication. Set `server.providerTokenGrants.spiffe.enabled=true` to mount the SPIFFE CSI Workload API socket into gateway and sandbox pods while keeping the projected ServiceAccount token bootstrap and gateway-minted sandbox JWT path.

Provider token grants require a SPIFFE implementation such as SPIRE and identities for the gateway and sandbox pods. The repository's local SPIRE overlay assigns sandbox IDs from the pod's `openshell.io/sandbox-id` annotation, but the gateway validation path only requires the supervisor SVID to be valid and in the same SPIFFE trust domain as the gateway SVID. Provider profiles with `token_grant` metadata cause the sandbox supervisor to request JWT-SVIDs and exchange them for upstream OAuth2 access tokens. Token-exchange profiles also require a gateway SPIFFE identity because the gateway brokers the intermediate token exchange with its own JWT-SVID.

The gateway verifies supervisor JWT-SVIDs with JWT bundles fetched from the SPIFFE Workload API, so intermediate token exchange does not require gateway access to the SPIRE OIDC discovery endpoint or its TLS CA.

## OIDC User Authentication

Set `server.oidc.issuer` to enable OIDC. The gateway validates the `Authorization: Bearer <token>` header on every request against the issuer's JWKS endpoint.

```shell
helm upgrade openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set server.oidc.issuer=https://your-idp.example.com/realms/openshell \
  --set server.oidc.audience=openshell-cli
```

The `audience` value must match the client ID configured in your identity provider for the OpenShell resource server.

### OIDC values reference

| Value                     | Default         | Purpose                                               |
| ------------------------- | --------------- | ----------------------------------------------------- |
| `server.oidc.issuer`      | `""`            | OIDC issuer URL. Empty disables OIDC.                 |
| `server.oidc.audience`    | `openshell-cli` | Expected `aud` claim in the JWT.                      |
| `server.oidc.jwksTtl`     | `3600`          | JWKS key cache TTL in seconds.                        |
| `server.oidc.rolesClaim`  | `""`            | Dot-separated path to the roles array in JWT claims.  |
| `server.oidc.adminRole`   | `""`            | Role name that grants admin access.                   |
| `server.oidc.userRole`    | `""`            | Role name that grants standard user access.           |
| `server.oidc.scopesClaim` | `""`            | Dot-separated path to the scopes array in JWT claims. |

### Auth-only mode vs. RBAC mode

Leave both `adminRole` and `userRole` empty to use auth-only mode: any request with a valid JWT from the configured issuer is accepted, but no role distinction is enforced.

Set both values to enable RBAC mode, where the gateway checks the role claim and enforces access based on the assigned role:

```shell
helm upgrade openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set server.oidc.issuer=https://your-idp.example.com/realms/openshell \
  --set server.oidc.audience=openshell-cli \
  --set server.oidc.rolesClaim=realm_access.roles \
  --set server.oidc.adminRole=openshell-admin \
  --set server.oidc.userRole=openshell-user
```

Both `adminRole` and `userRole` must be set, or both must be empty. Setting only one is not supported.

OIDC RBAC is method-level authorization. It controls which API operations a caller can perform, but provider and sandbox records are not owned by individual OIDC subjects. In shared clusters, treat provider credentials as gateway-wide resources and use separate gateways or external tenancy controls when users must not see or attach each other's providers and sandboxes.

### Provider-specific rolesClaim paths

| Provider           | rolesClaim value     |
| ------------------ | -------------------- |
| Keycloak           | `realm_access.roles` |
| Microsoft Entra ID | `roles`              |
| Okta               | `groups`             |

## Reverse-Proxy Auth Termination

When an access proxy, such as Cloudflare Access, ngrok, or a corporate SSO gateway, handles authentication in front of the OpenShell gateway, you can explicitly allow unauthenticated user calls at the gateway:

```shell
helm upgrade openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set server.auth.allowUnauthenticatedUsers=true
```

The gateway still serves TLS and sandbox supervisors still authenticate with gateway-minted sandbox JWTs. User-facing CLI/API calls without OIDC or mTLS credentials are accepted as an unauthenticated local developer principal. The proxy is responsible for authenticating callers and forwarding only authorized traffic.

To also disable TLS entirely (when the proxy terminates TLS before the request reaches the gateway):

```shell
  --set server.disableTls=true \
  --set server.auth.allowUnauthenticatedUsers=true
```

Only enable unauthenticated users when the gateway is not reachable from outside a trusted local development environment or the proxy path is fully trusted. Never expose a plaintext, auth-disabled gateway to a public network.

Register the gateway with the CLI using the proxy's public URL. The browser-based login flow runs automatically on first use:

```shell
openshell gateway add https://gateway.example.com --name production
```
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://docs.nvidia.com/openshell/latest/kubernetes/ingress",
  "format": "markdown"
}
```

**Output:**
```
> For clean Markdown of any page, append .md to the page URL.
> For a complete documentation index, see https://docs.nvidia.com/openshell/llms.txt.
> For AI client integration (Claude Code, Cursor, etc.), connect to the MCP server at https://docs.nvidia.com/openshell/_mcp/server.

# Ingress

> Expose the OpenShell gateway externally using the Kubernetes Gateway API and a GRPCRoute.

By default, the OpenShell gateway is only reachable inside the cluster. To let CLI clients connect without a `kubectl port-forward`, expose the gateway through an ingress.

OpenShell uses the [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io) for ingress. The chart creates a `GRPCRoute` that routes inbound gRPC traffic to the gateway pod. You need a Gateway API implementation installed on your cluster to fulfill the `GRPCRoute`. This page uses [Envoy Gateway](https://gateway.envoyproxy.io), which the chart is tested with.

## Install Envoy Gateway

Envoy Gateway installs the Gateway API CRDs and controller:

```shell
helm install eg \
  oci://docker.io/envoyproxy/gateway-helm \
  --version v1.8.1 \
  --namespace envoy-gateway-system \
  --create-namespace \
  --wait
```

## Create the GatewayClass

Create the `eg` GatewayClass that the OpenShell chart references:

```shell
kubectl apply -f - <<'EOF'
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
EOF
```

Verify the GatewayClass is accepted:

```shell
kubectl get gatewayclass eg
```

The `ACCEPTED` column should show `True`.

## Install OpenShell with Gateway API enabled

Enable the GRPCRoute and let the chart create a Gateway resource in the `openshell` namespace:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set grpcRoute.enabled=true \
  --set grpcRoute.gateway.create=true \
  --set grpcRoute.gateway.className=eg
```

## Get the external address

After the Gateway is provisioned, Envoy Gateway creates a LoadBalancer service in the `openshell` namespace. Wait for it to get an external address:

```shell
kubectl -n openshell get svc -l gateway.envoyproxy.io/owning-gateway-name=openshell
```

After the `EXTERNAL-IP` is assigned, register the gateway with the CLI:

```shell
openshell gateway add http://<external-ip> --name production
openshell status
```

This setup is plaintext end-to-end and is intended for development. For external access, terminate TLS at the gateway as shown below.

## HTTPS (TLS termination)

Envoy Gateway can terminate TLS at the listener and forward plaintext to the OpenShell gateway pod:

```text
client → HTTPS → Envoy Gateway (terminate TLS) → plaintext → openshell gateway pod
```

Envoy Gateway only terminates TLS here — it does not perform OIDC. Do not enable an Envoy Gateway OIDC `SecurityPolicy` in front of the gateway: that flow relies on browser redirects and cookies and cannot work with the OpenShell CLI or headless agents. Instead, the OpenShell gateway validates an OIDC bearer token that the client sends in the gRPC `authorization` metadata, which Envoy forwards untouched.

Because Envoy terminates TLS, the OpenShell gateway never sees a client certificate, so client mTLS cannot provide identity on this path. Use OIDC bearer tokens for client identity instead.

For an interactive login on a headless machine, set `OPENSHELL_NO_BROWSER=1` before running `openshell gateway add`. The CLI uses the Device Authorization Grant with S256 PKCE and prompts the user to approve the login from another browser. For unattended agents and CI, set `OPENSHELL_OIDC_CLIENT_SECRET` to use the OAuth2 client-credentials grant instead. The client id comes from `--oidc-client-id` (default `openshell-cli`); pass it explicitly if your identity provider uses a different id. Interactive users with a local browser get the Authorization Code flow with PKCE by default.

### Provide a TLS certificate

Create a `kubernetes.io/tls` Secret in the `openshell` namespace with the certificate for your external hostname:

```shell
kubectl -n openshell create secret tls openshell-ingress-tls \
  --cert=tls.crt --key=tls.key
```

The Secret may also be issued by cert-manager, or you can reference the chart's existing `openshell-server-tls` Secret if its SANs include the external hostname.

### Install with HTTPS termination

Enable an HTTPS listener, point it at the Secret, disable gateway-pod TLS so Envoy forwards plaintext, and configure an OIDC issuer for client identity:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set grpcRoute.enabled=true \
  --set grpcRoute.gateway.create=true \
  --set grpcRoute.gateway.className=eg \
  --set grpcRoute.gateway.listener.protocol=HTTPS \
  --set grpcRoute.gateway.listener.port=443 \
  --set 'grpcRoute.gateway.listener.tls.certificateRefs[0].name=openshell-ingress-tls' \
  --set server.disableTls=true \
  --set server.oidc.issuer=https://<issuer> \
  --set 'grpcRoute.hostnames[0]=<external-hostname>'
```

Keep the certificate Secret in the release namespace. Referencing a Secret in another namespace requires a `ReferenceGrant`.

### Register over HTTPS

```shell
openshell gateway add https://<external-hostname> --name production --oidc-issuer https://<issuer>
openshell status
```

See [Authentication](/kubernetes/setup) for OIDC issuer, audience, and roles configuration.

## SSH Relay

Sandbox SSH uses the gateway endpoint registered with the CLI. No separate Helm SSH host or port values are required.

## Next Steps

Return to [Setup](/kubernetes/setup) to complete the installation.
```

---

## Assistant (Build · GPT-5.6 Terra · 11.9s)

_Thinking:_

**Inspecting mod cache sources**

I need to examine the mod cache source. I'm curious if glob paths escape uppercase. It’s important to check if the path exists. I'll need to use globbing that starts with something like `github.com/!gkrumbach07/openshell-dashboard/backend@*/pkg/clients/*.go`. I wonder if the tools support absolute paths for this. Oh, and I should also consider the NVIDIA SDK while I'm at it! Let's see how this goes.

**Tool: glob**

**Input:**
```json
{
  "path": "/Users/derxu/go/pkg/mod/github.com",
  "pattern": "!gkrumbach07/openshell-dashboard/backend@*/pkg/clients/*.go"
}
```

**Output:**
```
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/README.txt
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/autoneg_test.go
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/Makefile
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/autoneg.go
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/LICENSE
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/docs.sh
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/simple_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/openapi3gen.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/Makefile
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/internal/subpkg/sub_type.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/field_info.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/openapi3gen_newschemarefforvalue_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/type_info.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/openapi3gen_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/parameter.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/v2_apis_guru_openapi_directory_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/openapi2.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/helpers.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/issues1010_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/schema.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/path_item.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/marsh.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/doc.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/ref.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/refs.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/response.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/marsh_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/operation.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/security_scheme.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/openapi2_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/refs_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3conv/openapi3_conv_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3conv/doc.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3conv/openapi3_conv.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/value.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/equals_test.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/parse.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/expression.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/README.rst
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/match.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/machine.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/compile.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/match_test.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/uritemplate_test.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/go.sum
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/equals.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/LICENSE
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/escape.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/go.mod
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/uritemplate.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/error.go
/Users/derxu/go/pkg/mod/github.com/yosida95/uritemplate/v3@v3.0.2/prog.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1062_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1008_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue979_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue573_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue847_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/openapi2_conv.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/doc.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/openapi2_conv_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1069_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1091_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue440_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/custom_methods_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue187_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1016_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1049_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue558_test.go
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/bench_test.go
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/utils.go
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/README.md
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/go.sum
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/testdata/swagger.json
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/isatty_others.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/isatty_others_test.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/doc.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/README.md
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/isatty_solaris.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/go.sum
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/isatty_bsd.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/isatty_windows_test.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/isatty_plan9.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/go.test.sh
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/LICENSE
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/go.mod
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/example_test.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/isatty_windows.go
/Users/derxu/go/pkg/mod/github.com/mattn/go-isatty@v0.0.20/isatty_tcgets.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/testdata/apis_guru_openapi_directory/powerdns_local_0_0_13_swagger_yaml__load
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/internal/sortedset_test.go
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/internal/sortedset.go
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/LICENSE
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/go.mod
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/utils_test.go
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/cors_test.go
/Users/derxu/go/pkg/mod/github.com/rs/cors@v1.11.1/cors.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2/header.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/LICENSE
/Users/derxu/go/pkg/mod/github.com/modern-go/reflect2@v1.0.3-0.20250322232337-35a7c28c31ee/unsafe_type.go
/Users/derxu/go/pkg/mod/github.com/modern-go/reflect2@v1.0.3-0.20250322232337-35a7c28c31ee/unsafe_slice.go

(Results are truncated: showing first 100 results. Consider using a more specific path or pattern.)
```

**Tool: glob**

**Input:**
```json
{
  "path": "/Users/derxu/go/pkg/mod/github.com",
  "pattern": "!n!v!i!d!i!a/!open!shell/sdk/go@*/openshell/v1/*.go"
}
```

**Output:**
```
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/README.txt
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/autoneg_test.go
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/Makefile
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/autoneg.go
/Users/derxu/go/pkg/mod/github.com/munnerz/goautoneg@v0.0.0-20191010083416-a7dc8b61c822/LICENSE
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/docs.sh
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/simple_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/openapi3gen.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/internal/subpkg/sub_type.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/field_info.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/openapi3gen_newschemarefforvalue_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/type_info.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3gen/openapi3gen_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3conv/openapi3_conv_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3conv/doc.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3conv/openapi3_conv.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1062_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1008_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue979_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue573_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue847_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/openapi2_conv.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/doc.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/openapi2_conv_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1069_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1091_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue440_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/custom_methods_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue187_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1016_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue1049_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi2conv/issue558_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/errors.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validation_handler.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validation_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue1243_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue743_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validation_discriminator_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue707_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_request.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_response.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/yaml_nan_inf_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue1112_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_request_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/csv_file_upload_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue641_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validation_kit.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_request_example_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue1045_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/req_resp_encoder.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/array_items_nil_panic_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/middleware.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue1134_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue1230_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue789_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue1110_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/ghsa_mmfr_pmjx_hw9w_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue722_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validation_enum_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue201_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue436_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/req_resp_decoder.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_readonly_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/options_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/ghsa_74vm_87hj_r66f_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validation_error_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/encoded_path_route_confusion_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/zip_file_upload_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue733_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue639_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue884_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue689_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_set_default_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/req_resp_decoder_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue625_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/unpack_errors_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue267_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue949_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_request_input.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validation_error_encoder.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_request_custom_methods_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/circular_recursion_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue624_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/middleware_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue1105_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/decode_array_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/authentication_input.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/req_resp_encoder_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue991_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_response_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/options.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/internal.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/upload_arbitrary_file_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/issue1246_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/testdata/issue1100_test.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/testdata/fixtures/petstore.json
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/testdata/petstore.yaml
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/testdata/encoded_path_authorization_bypass.yml
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validate_response_input.go
/Users/derxu/go/pkg/mod/github.com/getkin/kin-openapi@v0.149.0/openapi3filter/validation_error.go

(Results are truncated: showing first 100 results. Consider using a more specific path or pattern.)
```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/go/pkg/mod/github.com",
  "pattern": "type RawExecClient|func NewRawExecClient|type TLSConfig|func NewClient",
  "include": "*.go"
}
```

**Output:**
```
Found 100 matches (more matches available)
/Users/derxu/go/pkg/mod/github.com/rhuss/openshell-sdk-go@v0.3.1/openshell/v1/gateway/gateway.go:
  Line 22: func NewClient(name string, opts ...ClientOption) (*v1.Client, error) {


/Users/derxu/go/pkg/mod/github.com/rhuss/openshell-sdk-go@v0.3.1/openshell/v1/fake/fake.go:
  Line 70: func NewClient(opts ...ClientOption) *Client {


/Users/derxu/go/pkg/mod/github.com/rhuss/openshell-sdk-go@v0.3.1/openshell/v1/types.go:
  Line 43: type TLSConfig = types.TLSConfig


/Users/derxu/go/pkg/mod/github.com/rhuss/openshell-sdk-go@v0.3.1/openshell/v1/types/types.go:
  Line 41: type TLSConfig struct {


/Users/derxu/go/pkg/mod/github.com/rhuss/openshell-sdk-go@v0.3.1/openshell/v1/client.go:
  Line 68: func NewClient(cfg Config) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/azidentity@v1.13.0/client_certificate_credential.go:
  Line 58: func NewClientCertificateCredential(tenantID string, clientID string, certs []*x509.Certificate, key crypto.PrivateKey, options *ClientCertificateCredentialOptions) (*ClientCertificateCredential, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/azidentity@v1.13.0/client_assertion_credential.go:
  Line 53: func NewClientAssertionCredential(tenantID, clientID string, getAssertion func(context.Context) (string, error), options *ClientAssertionCredentialOptions) (*ClientAssertionCredential, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/azidentity@v1.13.0/client_secret_credential.go:
  Line 47: func NewClientSecretCredential(tenantID string, clientID string, clientSecret string, options *ClientSecretCredentialOptions) (*ClientSecretCredential, error) {


/Users/derxu/go/pkg/mod/github.com/opendatahub-io/ogx-k8s-operator@v0.0.0-20260512184943-d419c039b7b9/api/v1alpha1/llamastackdistribution_types.go:
  Line 152: type TLSConfig struct {


/Users/derxu/go/pkg/mod/github.com/opendatahub-io/ogx-k8s-operator@v0.0.0-20260512184943-d419c039b7b9/api/v1beta1/provider_types_inference.go:
  Line 22: type TLSConfig struct {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/storage/azblob@v1.6.2/blob/client.go:
  Line 39: func NewClient(blobURL string, cred azcore.TokenCredential, options *ClientOptions) (*Client, error) {

  Line 56: func NewClientWithNoCredential(blobURL string, options *ClientOptions) (*Client, error) {

  Line 70: func NewClientWithSharedKeyCredential(blobURL string, cred *SharedKeyCredential, options *ClientOptions) (*Client, error) {

  Line 87: func NewClientFromConnectionString(connectionString, containerName, blobName string, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/aws/smithy-go@v1.28.1/transport/http/client.go:
  Line 39: func NewClientHandler(client ClientDo) ClientHandler {

  Line 45: func NewClientHandlerWithOptions(client ClientDo, opts ...func(*ClientHandler)) ClientHandler {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/storage/azblob@v1.6.2/service/client.go:
  Line 42: func NewClient(serviceURL string, cred azcore.TokenCredential, options *ClientOptions) (*Client, error) {

  Line 59: func NewClientWithNoCredential(serviceURL string, options *ClientOptions) (*Client, error) {

  Line 73: func NewClientWithSharedKeyCredential(serviceURL string, cred *SharedKeyCredential, options *ClientOptions) (*Client, error) {

  Line 89: func NewClientFromConnectionString(connectionString string, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/storage/azblob@v1.6.2/container/client.go:
  Line 44: func NewClient(containerURL string, cred azcore.TokenCredential, options *ClientOptions) (*Client, error) {

  Line 61: func NewClientWithNoCredential(containerURL string, options *ClientOptions) (*Client, error) {

  Line 75: func NewClientWithSharedKeyCredential(containerURL string, cred *SharedKeyCredential, options *ClientOptions) (*Client, error) {

  Line 91: func NewClientFromConnectionString(connectionString string, containerName string, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/storage/azblob@v1.6.2/pageblob/client.go:
  Line 38: func NewClient(blobURL string, cred azcore.TokenCredential, options *ClientOptions) (*Client, error) {

  Line 55: func NewClientWithNoCredential(blobURL string, options *ClientOptions) (*Client, error) {

  Line 69: func NewClientWithSharedKeyCredential(blobURL string, cred *blob.SharedKeyCredential, options *ClientOptions) (*Client, error) {

  Line 86: func NewClientFromConnectionString(connectionString, containerName, blobName string, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260915184529-b670fa5e7760/pkg/clients/rawexec.go:
  Line 25: type RawExecClient struct {

  Line 34: func NewRawExecClient(address, caFile, clientCert, clientKey string, useTLS bool) (*RawExecClient, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/storage/azblob@v1.6.2/client.go:
  Line 33: func NewClient(serviceURL string, cred azcore.TokenCredential, options *ClientOptions) (*Client, error) {

  Line 48: func NewClientWithNoCredential(serviceURL string, options *ClientOptions) (*Client, error) {

  Line 63: func NewClientWithSharedKeyCredential(serviceURL string, cred *SharedKeyCredential, options *ClientOptions) (*Client, error) {

  Line 77: func NewClientFromConnectionString(connectionString string, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/storage/azblob@v1.6.2/internal/base/clients.go:
  Line 62: func NewClient[T any](inner *T) *Client[T] {


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/rawexec.go:
  Line 25: type RawExecClient struct {

  Line 34: func NewRawExecClient(address, caFile, clientCert, clientKey string, useTLS bool) (*RawExecClient, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/storage/azblob@v1.6.2/appendblob/client.go:
  Line 37: func NewClient(blobURL string, cred azcore.TokenCredential, options *ClientOptions) (*Client, error) {

  Line 55: func NewClientWithNoCredential(blobURL string, options *ClientOptions) (*Client, error) {

  Line 70: func NewClientWithSharedKeyCredential(blobURL string, cred *blob.SharedKeyCredential, options *ClientOptions) (*Client, error) {

  Line 88: func NewClientFromConnectionString(connectionString, containerName, blobName string, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/storage/azblob@v1.6.2/blockblob/client.go:
  Line 47: func NewClient(blobURL string, cred azcore.TokenCredential, options *ClientOptions) (*Client, error) {

  Line 64: func NewClientWithNoCredential(blobURL string, options *ClientOptions) (*Client, error) {

  Line 79: func NewClientWithSharedKeyCredential(blobURL string, cred *blob.SharedKeyCredential, options *ClientOptions) (*Client, error) {

  Line 97: func NewClientFromConnectionString(connectionString, containerName, blobName string, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/azcore@v1.19.1/arm/client.go:
  Line 35: func NewClient(moduleName, moduleVersion string, cred azcore.TokenCredential, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/openai/openai-go/v2@v2.7.1/realtime/clientsecret.go:
  Line 33: func NewClientSecretService(opts ...option.RequestOption) (r ClientSecretService) {


/Users/derxu/go/pkg/mod/github.com/openai/openai-go/v2@v2.7.1/client.go:
  Line 73: func NewClient(opts ...option.RequestOption) (r Client) {


/Users/derxu/go/pkg/mod/github.com/aws/aws-sdk-go-v2@v1.44.0/internal/protocoltest/mockHttpClient.go:
  Line 24: func NewClient() *HTTPClient {


/Users/derxu/go/pkg/mod/github.com/!azure/azure-sdk-for-go/sdk/azcore@v1.19.1/core.go:
  Line 127: func NewClient(moduleName, moduleVersion string, plOpts runtime.PipelineOptions, options *ClientOptions) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/gateway/gateway.go:
  Line 23: func NewClient(name string, opts ...ClientOption) (*v1.Client, error) {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/abev3/abev3.go:
  Line 2686: func NewClient(server string, opts ...ClientOption) (*Client, error) {

  Line 14025: func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/helloworldv3/helloworldv3.go:
  Line 270: func NewClient(server string, opts ...ClientOption) (*Client, error) {

  Line 2324: func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {


/Users/derxu/go/pkg/mod/github.com/k8s-manifest-kit/renderer-helm@v0.4.0/pkg/container/client.go:
  Line 26: func NewClient(ref string, opts ...ClientOption) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/responsebody/client/response_body_service/response_body_service_client.go:
  Line 23: func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {

  Line 35: func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/fake/fake.go:
  Line 74: func NewClient(opts ...ClientOption) *Client {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/echo/client/echo_service/echo_service_client.go:
  Line 23: func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {

  Line 35: func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/client.go:
  Line 70: func NewClient(cfg Config) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/oidc/credentials_auth.go:
  Line 30: func NewClientCredentialsAuth(opts ...LoginOption) (types.AuthProvider, error) {


/Users/derxu/go/pkg/mod/github.com/openai/openai-go@v1.12.0/client.go:
  Line 68: func NewClient(opts ...option.RequestOption) (r Client) {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/unannotatedecho/client/unannotated_echo_service/unannotated_echo_service_client.go:
  Line 23: func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {

  Line 35: func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {


/Users/derxu/go/pkg/mod/github.com/gorilla/websocket@v1.5.3/client.go:
  Line 40: func NewClient(netConn net.Conn, u *url.URL, requestHeader http.Header, readBufSize, writeBufSize int) (c *Conn, response *http.Response, err error) {


/Users/derxu/go/pkg/mod/github.com/gorilla/websocket@v1.5.4-0.20250319132907-e064f32e3674/client.go:
  Line 39: func NewClient(netConn net.Conn, u *url.URL, requestHeader http.Header, readBufSize, writeBufSize int) (c *Conn, response *http.Response, err error) {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/types.go:
  Line 47: type TLSConfig = types.TLSConfig


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/types/types.go:
  Line 45: type TLSConfig struct {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/abe/client/a_bit_of_everything/a_bit_of_everything_client.go:
  Line 23: func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {

  Line 35: func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/abe/client/camel_case_service_name/camel_case_service_name_client.go:
  Line 23: func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {

  Line 35: func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/abe/client/echo_rpc/echo_rpc_client.go:
  Line 23: func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {

  Line 35: func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/abe/client/snake_enum_service/snake_enum_service_client.go:
  Line 23: func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {

  Line 35: func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260902135142-387aea069a8b/openshell/v1/gateway/gateway.go:
  Line 22: func NewClient(name string, opts ...ClientOption) (*v1.Client, error) {


/Users/derxu/go/pkg/mod/github.com/!gkrumbach07/openshell-dashboard/backend@v0.0.0-20260916180845-978bcb5eb6b5/internal/sdkclient/rawexec.go:
  Line 25: type RawExecClient struct {

  Line 34: func NewRawExecClient(address, caFile, clientCert, clientKey string, useTLS bool) (*RawExecClient, error) {


/Users/derxu/go/pkg/mod/github.com/grpc-ecosystem/grpc-gateway/v2@v2.30.0/examples/internal/clients/generateunboundmethods/client/generate_unbound_methods_echo_service/generate_unbound_methods_echo_service_client.go:
  Line 23: func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {

  Line 35: func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {


/Users/derxu/go/pkg/mod/github.com/aws/aws-sdk-go-v2/service/internal/s3shared@v1.19.40/endpoint_error.go:
  Line 98: func NewClientPartitionMismatchError(resource arn.Resource, clientPartitionID, clientRegion string, err error) ConfigurationError {

  Line 109: func NewClientRegionMismatchError(resource arn.Resource, clientPartitionID, clientRegion string, err error) ConfigurationError {

  Line 131: func NewClientConfiguredForFIPSError(resource arn.Resource, clientPartitionID, clientRegion string, err error) ConfigurationError {

  Line 153: func NewClientConfiguredForAccelerateError(resource arn.Resource, clientPartitionID, clientRegion string, err error) ConfigurationError {

  Line 164: func NewClientConfiguredForCrossRegionFIPSError(resource arn.Resource, clientPartitionID, clientRegion string, err error) ConfigurationError {

  Line 175: func NewClientConfiguredForDualStackError(resource arn.Resource, clientPartitionID, clientRegion string, err error) ConfigurationError {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260902135142-387aea069a8b/openshell/v1/fake/fake.go:
  Line 74: func NewClient(opts ...ClientOption) *Client {


/Users/derxu/go/pkg/mod/github.com/!gkrumbach07/openshell-dashboard/backend@v0.0.0-20260916180940-45d3f1f22227/pkg/sdkclient/rawexec.go:
  Line 25: type RawExecClient struct {

  Line 34: func NewRawExecClient(address, caFile, clientCert, clientKey string, useTLS bool) (*RawExecClient, error) {


/Users/derxu/go/pkg/mod/github.com/redis/go-redis/v9@v9.20.1/maintnotifications/e2e/config_parser_test.go:
  Line 274: func NewClientFactory(config *RedisConnectionConfig) *ClientFactory {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260902135142-387aea069a8b/openshell/v1/client.go:
  Line 72: func NewClient(cfg Config) (*Client, error) {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260902135142-387aea069a8b/openshell/v1/oidc/credentials_auth.go:
  Line 30: func NewClientCredentialsAuth(opts ...LoginOption) (types.AuthProvider, error) {


/Users/derxu/go/pkg/mod/github.com/redis/go-redis/v9@v9.20.1/redis.go:
  Line 1340: func NewClient(opt *Options) *Client {


/Users/derxu/go/pkg/mod/github.com/!gkrumbach07/openshell-dashboard/backend@v0.0.0-20260909185808-701454a4a47e/internal/sdkclient/rawexec.go:
  Line 25: type RawExecClient struct {

  Line 34: func NewRawExecClient(address, caFile, clientCert, clientKey string, useTLS bool) (*RawExecClient, error) {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260902135142-387aea069a8b/openshell/v1/types.go:
  Line 47: type TLSConfig = types.TLSConfig


/Users/derxu/go/pkg/mod/github.com/aws/aws-sdk-go@v1.55.6/example/aws/request/httptrace/client.go:
  Line 10: func NewClient(cfg ClientConfig) *http.Client {


(Results truncated. Consider using a more specific path or pattern.)
```

**Tool: bash**

**Input:**
```json
{
  "workdir": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "command": "git status --short",
  "timeout": 120000
}
```

**Output:**
```
 M cmd/gateway_setup.go
 M cmd/main.go
 M go.mod
 M go.sum
 M pkg/fleet/registry.go
 M pkg/fleet/router.go
 D temp-bff-1/README.md
 D temp-bff-1/go.mod
 D temp-bff-1/registry.go
 D temp-bff-1/registry_test.go
?? internal/clients/
?? pkg/discovery/
?? pkg/fleet/registry_test.go

```

---

## Assistant (Build · GPT-5.6 Terra · 5.4s)

_Thinking:_



**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/types/types.go"
}
```

**Output:**
```
<path>/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/types/types.go</path>
<type>file</type>
<content>
1: // SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
2: // SPDX-License-Identifier: Apache-2.0
3: 
4: package types
5: 
6: import "time"
7: 
8: // SandboxPhase represents the lifecycle phase of a sandbox.
9: type SandboxPhase string
10: 
11: // SandboxPhase values for sandbox lifecycle.
12: const (
13: 	SandboxProvisioning SandboxPhase = "Provisioning"
14: 	SandboxReady        SandboxPhase = "Ready"
15: 	SandboxError        SandboxPhase = "Error"
16: 	SandboxDeleting     SandboxPhase = "Deleting"
17: 	SandboxUnknown      SandboxPhase = "Unknown"
18: 	SandboxStopping     SandboxPhase = "Stopping"
19: 	SandboxStopped      SandboxPhase = "Stopped"
20: 	SandboxStarting     SandboxPhase = "Starting"
21: 	SandboxCompleted    SandboxPhase = "Completed"
22: )
23: 
24: // EventType classifies watch events.
25: type EventType string
26: 
27: // EventType values for watch events.
28: const (
29: 	EventAdded    EventType = "ADDED"
30: 	EventModified EventType = "MODIFIED"
31: 	EventDeleted  EventType = "DELETED"
32: 	EventError    EventType = "ERROR"
33: )
34: 
35: // StreamType identifies which output stream a chunk belongs to.
36: type StreamType string
37: 
38: // StreamType values for exec output.
39: const (
40: 	StreamStdout StreamType = "stdout"
41: 	StreamStderr StreamType = "stderr"
42: )
43: 
44: // TLSConfig holds TLS connection settings.
45: type TLSConfig struct {
46: 	CertFile string
47: 	KeyFile  string
48: 	CAFile   string
49: 	Insecure bool
50: }
51: 
52: // RetryPolicy configures automatic retry behavior for failed RPCs.
53: type RetryPolicy struct {
54: 	MaxRetries  int
55: 	InitialWait time.Duration
56: 	MaxWait     time.Duration
57: }

(End of file - total 57 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/client.go"
}
```

**Output:**
```
<path>/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/client.go</path>
<type>file</type>
<content>
1: // SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
2: // SPDX-License-Identifier: Apache-2.0
3: 
4: package v1
5: 
6: import (
7: 	"context"
8: 	"sync"
9: 
10: 	"github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/types"
11: 
12: 	internalgrpc "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1/internal/grpc"
13: 	"google.golang.org/grpc"
14: )
15: 
16: // Config holds all settings needed to create a Client.
17: type Config = types.Config
18: 
19: // ClientInterface defines the top-level SDK surface.
20: type ClientInterface interface {
21: 	Sandboxes() SandboxInterface
22: 	SandboxTemplates() SandboxTemplateInterface
23: 	CreateSandboxFromTemplate(ctx context.Context, workspace, name, templateName string, spec *SandboxSpec, labels map[string]string, opts ...CreateOptions) (*Sandbox, error)
24: 	Providers() ProviderInterface
25: 	Services() ServiceInterface
26: 	Exec() ExecInterface
27: 	Files() FileInterface
28: 	Health() HealthInterface
29: 	SSH() SSHInterface
30: 	TCP() TCPInterface
31: 	Config() ConfigInterface
32: 	Policy() PolicyInterface
33: 	Workspaces() WorkspaceInterface
34: 	Close() error
35: }
36: 
37: // SandboxInterface is defined in sandbox.go
38: 
39: // ProviderInterface is defined in provider.go
40: 
41: // ExecInterface is defined in exec.go
42: 
43: // FileInterface is defined in file.go
44: 
45: // Client implements ClientInterface. It holds a gRPC connection and provides
46: // sub-client accessors following the Kubernetes client-go pattern.
47: type Client struct {
48: 	conn   *grpc.ClientConn
49: 	config Config
50: 
51: 	closeOnce sync.Once
52: 	closeErr  error
53: 
54: 	sandboxes      SandboxInterface
55: 	templateCreate SandboxTemplateCreateInterface
56: 	templates      SandboxTemplateInterface
57: 	providers      ProviderInterface
58: 	services       ServiceInterface
59: 	exec           ExecInterface
60: 	files          FileInterface
61: 	health         HealthInterface
62: 	ssh            SSHInterface
63: 	tcp            TCPInterface
64: 	cfg            ConfigInterface
65: 	policy         PolicyInterface
66: 	workspaces     WorkspaceInterface
67: }
68: 
69: // NewClient creates a new SDK client connected to the given gateway.
70: func NewClient(cfg Config) (*Client, error) {
71: 	if cfg.Address == "" {
72: 		return nil, &StatusError{Code: ErrorInvalidArgument, Message: "address must not be empty"}
73: 	}
74: 
75: 	if cfg.Auth == nil {
76: 		cfg.Auth = NoAuth()
77: 	}
78: 
79: 	var tlsParams *internalgrpc.TLSParams
80: 	if cfg.TLS != nil {
81: 		tlsParams = &internalgrpc.TLSParams{
82: 			CertFile: cfg.TLS.CertFile,
83: 			KeyFile:  cfg.TLS.KeyFile,
84: 			CAFile:   cfg.TLS.CAFile,
85: 			Insecure: cfg.TLS.Insecure,
86: 		}
87: 	}
88: 
89: 	conn, err := internalgrpc.NewConnection(cfg.Address, tlsParams, cfg.Auth)
90: 	if err != nil {
91: 		return nil, err
92: 	}
93: 
94: 	c := &Client{
95: 		conn:   conn,
96: 		config: cfg,
97: 	}
98: 
99: 	sandboxes := newSandboxClient(conn)
100: 	c.sandboxes = sandboxes
101: 	c.templateCreate = sandboxes
102: 	c.templates = newSandboxTemplateClient(conn)
103: 	c.providers = newProviderClient(conn)
104: 	c.services = newServiceClient(conn)
105: 	c.exec = newExecClient(conn, c.sandboxes)
106: 	c.files = newFileClient(conn, c.sandboxes)
107: 	c.health = newHealthClient(conn)
108: 	c.ssh = newSSHClient(conn, c.sandboxes)
109: 	c.tcp = newTCPClient(conn, c.sandboxes, c.ssh)
110: 	c.cfg = newConfigClient(conn, c.sandboxes)
111: 	c.policy = newPolicyClient(conn)
112: 	c.workspaces = newWorkspaceClient(conn)
113: 
114: 	return c, nil
115: }
116: 
117: // Sandboxes returns the sandbox sub-client.
118: func (c *Client) Sandboxes() SandboxInterface { return c.sandboxes }
119: 
120: // SandboxTemplates returns the reusable sandbox template sub-client.
121: func (c *Client) SandboxTemplates() SandboxTemplateInterface { return c.templates }
122: 
123: // CreateSandboxFromTemplate creates a sandbox from a named workload template
124: // without changing the legacy Sandboxes() interface.
125: func (c *Client) CreateSandboxFromTemplate(ctx context.Context, workspace, name, templateName string, spec *SandboxSpec, labels map[string]string, opts ...CreateOptions) (*Sandbox, error) {
126: 	return c.templateCreate.CreateFromTemplate(ctx, workspace, name, templateName, spec, labels, opts...)
127: }
128: 
129: // Providers returns the provider sub-client.
130: func (c *Client) Providers() ProviderInterface { return c.providers }
131: 
132: // Services returns the service sub-client.
133: func (c *Client) Services() ServiceInterface { return c.services }
134: 
135: // Exec returns the exec sub-client.
136: func (c *Client) Exec() ExecInterface { return c.exec }
137: 
138: // Files returns the file sub-client.
139: func (c *Client) Files() FileInterface { return c.files }
140: 
141: // Health returns the health sub-client.
142: func (c *Client) Health() HealthInterface { return c.health }
143: 
144: // SSH returns the SSH session sub-client.
145: func (c *Client) SSH() SSHInterface { return c.ssh }
146: 
147: // TCP returns the TCP port forwarding sub-client.
148: func (c *Client) TCP() TCPInterface { return c.tcp }
149: 
150: // Config returns the configuration sub-client.
151: func (c *Client) Config() ConfigInterface { return c.cfg }
152: 
153: // Policy returns the policy management sub-client.
154: func (c *Client) Policy() PolicyInterface { return c.policy }
155: 
156: // Workspaces returns the workspace management sub-client.
157: func (c *Client) Workspaces() WorkspaceInterface { return c.workspaces }
158: 
159: // Close closes the underlying gRPC connection. Safe to call multiple times.
160: func (c *Client) Close() error {
161: 	c.closeOnce.Do(func() {
162: 		c.closeErr = c.conn.Close()
163: 	})
164: 	return c.closeErr
165: }

(End of file - total 165 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/internal/transport/transport.go"
}
```

**Error:**
```
File not found: /Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/internal/transport/transport.go
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/rawexec.go"
}
```

**Output:**
```
<path>/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/rawexec.go</path>
<type>file</type>
<content>
1: package clients
2: 
3: import (
4: 	"context"
5: 	"crypto/tls"
6: 	"crypto/x509"
7: 	"fmt"
8: 	"io"
9: 	"os"
10: 	"strings"
11: 
12: 	pb "github.com/NVIDIA/OpenShell/sdk/go/proto/openshellv1"
13: 	"google.golang.org/grpc"
14: 	"google.golang.org/grpc/credentials"
15: 	"google.golang.org/grpc/credentials/insecure"
16: )
17: 
18: // RawExecClient performs the gateway's non-TTY, stdin-carrying ExecSandbox RPC,
19: // which the OpenShell Go SDK does not expose (its Run has no stdin and its
20: // Interactive forces a PTY). It exists solely so binary file uploads run
21: // through a clean pipe — `dd` with Stdin bytes and Tty=false — instead of a
22: // PTY, whose line discipline (EOF/flow-control bytes, CR/LF translation, echo)
23: // silently corrupts binary content. It shares the same address, TLS, and
24: // per-request bearer forwarding as the main SDK client.
25: type RawExecClient struct {
26: 	conn   *grpc.ClientConn
27: 	client pb.OpenShellClient
28: }
29: 
30: // NewRawExecClient dials the gateway. address is host:port (no URL scheme).
31: // When useTLS is set, caFile (optional) verifies the server and clientCert +
32: // clientKey (optional, both required together) enable mTLS client auth — the
33: // same knobs the SDK client uses, so upload honors gateway mTLS too.
34: func NewRawExecClient(address, caFile, clientCert, clientKey string, useTLS bool) (*RawExecClient, error) {
35: 	var creds credentials.TransportCredentials
36: 	if useTLS {
37: 		tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
38: 		if caFile != "" {
39: 			pem, err := os.ReadFile(caFile)
40: 			if err != nil {
41: 				return nil, fmt.Errorf("read gateway CA cert: %w", err)
42: 			}
43: 			pool := x509.NewCertPool()
44: 			if !pool.AppendCertsFromPEM(pem) {
45: 				return nil, fmt.Errorf("no valid certificates in %s", caFile)
46: 			}
47: 			tlsCfg.RootCAs = pool
48: 		}
49: 		if clientCert != "" && clientKey != "" {
50: 			cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
51: 			if err != nil {
52: 				return nil, fmt.Errorf("load gateway client cert/key: %w", err)
53: 			}
54: 			tlsCfg.Certificates = []tls.Certificate{cert}
55: 		}
56: 		creds = credentials.NewTLS(tlsCfg)
57: 	} else {
58: 		creds = insecure.NewCredentials()
59: 	}
60: 	conn, err := grpc.NewClient(address,
61: 		grpc.WithTransportCredentials(creds),
62: 		grpc.WithPerRPCCredentials(ContextAuthProvider{RequireTLS: useTLS}),
63: 	)
64: 	if err != nil {
65: 		return nil, err
66: 	}
67: 	return &RawExecClient{conn: conn, client: pb.NewOpenShellClient(conn)}, nil
68: }
69: 
70: // Close closes the underlying gRPC connection.
71: func (r *RawExecClient) Close() error { return r.conn.Close() }
72: 
73: // ExecWithStdin runs command in the sandbox (identified by UUID) with stdin
74: // piped in and no TTY, returning merged stdout+stderr and the process exit
75: // code. exitCode is -1 if the gateway sent no exit event.
76: func (r *RawExecClient) ExecWithStdin(ctx context.Context, sandboxID string, command []string, stdin []byte) (string, int, error) {
77: 	stream, err := r.client.ExecSandbox(ctx, &pb.ExecSandboxRequest{
78: 		SandboxId: sandboxID,
79: 		Command:   command,
80: 		Stdin:     stdin,
81: 		Tty:       false,
82: 	})
83: 	if err != nil {
84: 		return "", 0, err
85: 	}
86: 	var out strings.Builder
87: 	exitCode := -1
88: 	for {
89: 		ev, recvErr := stream.Recv()
90: 		if recvErr == io.EOF {
91: 			break
92: 		}
93: 		if recvErr != nil {
94: 			return out.String(), exitCode, recvErr
95: 		}
96: 		switch p := ev.Payload.(type) {
97: 		case *pb.ExecSandboxEvent_Stdout:
98: 			out.Write(p.Stdout.GetData())
99: 		case *pb.ExecSandboxEvent_Stderr:
100: 			out.Write(p.Stderr.GetData())
101: 		case *pb.ExecSandboxEvent_Exit:
102: 			exitCode = int(p.Exit.GetExitCode())
103: 		}
104: 	}
105: 	return out.String(), exitCode, nil
106: }

(End of file - total 106 lines)
</content>
```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6",
  "pattern": "CAFile|CertFile|KeyFile|Insecure",
  "include": "*.go"
}
```

**Output:**
```
Found 44 matches
/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/gateway/gateway_test.go:
  Line 53: 	client, err := NewClient("test-gw", WithTLS(&types.TLSConfig{Insecure: true}))

  Line 80: 		WithTLS(&types.TLSConfig{Insecure: true}),

  Line 97: 		WithTLS(&types.TLSConfig{Insecure: true}),

  Line 157: 		WithTLS(&types.TLSConfig{Insecure: true}),

  Line 168: 		WithTLS(&types.TLSConfig{Insecure: true}),

  Line 381: 	client, err := NewClient("", WithTLS(&types.TLSConfig{Insecure: true}))


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/gateway/gateway.go:
  Line 52: 		sdkCfg.TLS = &types.TLSConfig{Insecure: true}


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/client_test.go:
  Line 23: 		TLS:     &TLSConfig{Insecure: true},

  Line 48: 		TLS:     &TLSConfig{Insecure: true},

  Line 62: 		TLS:     &TLSConfig{Insecure: true},


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/types/types.go:
  Line 46: 	CertFile string

  Line 47: 	KeyFile  string

  Line 48: 	CAFile   string

  Line 49: 	Insecure bool


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/internal/grpc/conn.go:
  Line 21: 	CertFile string

  Line 22: 	KeyFile  string

  Line 23: 	CAFile   string

  Line 24: 	Insecure bool

  Line 42: 		if tlsCfg != nil && (tlsCfg.CAFile != "" || tlsCfg.CertFile != "" || tlsCfg.KeyFile != "") {

  Line 43: 			return nil, fmt.Errorf("grpc connect: TLS parameters (CAFile/CertFile/KeyFile) are ignored with plaintext (http://) address")

  Line 73: 		InsecureSkipVerify: cfg.Insecure, //nolint:gosec // user-requested skip for dev gateways

  Line 76: 	if cfg.CAFile != "" {

  Line 77: 		caCert, err := os.ReadFile(cfg.CAFile)

  Line 88: 	if cfg.CertFile != "" && cfg.KeyFile != "" {

  Line 89: 		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)

  Line 94: 	} else if cfg.CertFile != "" || cfg.KeyFile != "" {

  Line 95: 		return nil, fmt.Errorf("both CertFile and KeyFile must be provided for client certificate authentication")


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/internal/grpc/conn_test.go:
  Line 42: func TestNewConnectionInsecureTLSConfig(t *testing.T) {

  Line 43: 	conn, err := NewConnection("127.0.0.1:1", &TLSParams{Insecure: true}, nil)

  Line 54: func TestNewConnectionHTTPWithInsecureAuth(t *testing.T) {

  Line 70: 	_, err := NewConnection("http://127.0.0.1:1", &TLSParams{CAFile: "/some/ca.pem"}, nil)


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/oidc/discovery_test.go:
  Line 252: func TestDiscover_RejectsInsecureRemoteEndpoint(t *testing.T) {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/oidc/token_test.go:
  Line 38: func TestWriteToken_ReplacesInsecureExistingFileWithOwnerOnlyFile(t *testing.T) {


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/edge/doc.go:
  Line 66: //	    TLS:     &v1.TLSConfig{Insecure: true}, // local tunnel


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/edge/tunnel_test.go:
  Line 29: 		InsecureSkipVerify: true,

  Line 235: 			InsecureSkipVerify: true,


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/proto/sandboxv1/sandbox.pb.go:
  Line 2190: 	AllowInsecureTransport bool `protobuf:"varint,7,opt,name=allow_insecure_transport,json=allowInsecureTransport,proto3" json:"allow_insecure_transport,omitempty"`

  Line 2267: func (x *SupervisorMiddlewareService) GetAllowInsecureTransport() bool {

  Line 2269: 		return x.AllowInsecureTransport

  Line 2457: 	"\x18allow_insecure_transport\x18\a \x01(\bR\x16allowInsecureTransportJ\x04\b\x04\x10\x05R\atimeout*\x97\x01\n" +


/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/client.go:
  Line 82: 			CertFile: cfg.TLS.CertFile,

  Line 83: 			KeyFile:  cfg.TLS.KeyFile,

  Line 84: 			CAFile:   cfg.TLS.CAFile,

  Line 85: 			Insecure: cfg.TLS.Insecure,

```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c",
  "pattern": "CAFile|caFile|clientCert|clientKey|InsecureSkipVerify",
  "include": "*.go"
}
```

**Output:**
```
Found 10 matches
/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/gateway_setup.go:
  Line 45: 		tlsCfg := &openshell.TLSConfig{CAFile: gatewayCACert}


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/inbound_tls_test.go:
  Line 156: 			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/rawexec.go:
  Line 31: // When useTLS is set, caFile (optional) verifies the server and clientCert +

  Line 32: // clientKey (optional, both required together) enable mTLS client auth — the

  Line 34: func NewRawExecClient(address, caFile, clientCert, clientKey string, useTLS bool) (*RawExecClient, error) {

  Line 38: 		if caFile != "" {

  Line 39: 			pem, err := os.ReadFile(caFile)

  Line 45: 				return nil, fmt.Errorf("no valid certificates in %s", caFile)

  Line 49: 		if clientCert != "" && clientKey != "" {

  Line 50: 			cert, err := tls.LoadX509KeyPair(clientCert, clientKey)

```

---

## Assistant (Build · GPT-5.6 Terra · 5.3s)

_Thinking:_

**Determining connection requirements**

I need to check the connection details, possibly focusing on whether "Go" can confirm the TLS host for HTTPS maps. It seems important to ensure that the connection is secure and functioning properly, so I’ll look into the specifics of how to read this information. Understanding the configurations might help clarify what’s needed for the connection to be established successfully. I’ll keep it in mind as I analyze any relevant documentation or resources.

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/internal/grpc/conn.go"
}
```

**Output:**
```
<path>/Users/derxu/go/pkg/mod/github.com/!n!v!i!d!i!a/!open!shell/sdk/go@v0.0.0-20260921185137-251f77e2b8f6/openshell/v1/internal/grpc/conn.go</path>
<type>file</type>
<content>
1: // SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
2: // SPDX-License-Identifier: Apache-2.0
3: 
4: // Package grpc provides gRPC connection setup utilities.
5: package grpc
6: 
7: import (
8: 	"crypto/tls"
9: 	"crypto/x509"
10: 	"fmt"
11: 	"os"
12: 	"strings"
13: 
14: 	"google.golang.org/grpc"
15: 	"google.golang.org/grpc/credentials"
16: 	"google.golang.org/grpc/credentials/insecure"
17: )
18: 
19: // TLSParams holds TLS settings without importing the v1 package.
20: type TLSParams struct {
21: 	CertFile string
22: 	KeyFile  string
23: 	CAFile   string
24: 	Insecure bool
25: }
26: 
27: // NewConnection creates a gRPC client connection.
28: // The address may include an http:// or https:// scheme (as written by the
29: // upstream gateway). The scheme drives transport selection: http:// uses
30: // plaintext, https:// or no scheme uses TLS.
31: func NewConnection(address string, tlsCfg *TLSParams, auth credentials.PerRPCCredentials) (*grpc.ClientConn, error) {
32: 	usePlaintext := false
33: 	if strings.HasPrefix(address, "http://") {
34: 		usePlaintext = true
35: 		address = strings.TrimPrefix(address, "http://")
36: 	} else {
37: 		address = strings.TrimPrefix(address, "https://")
38: 	}
39: 	opts := []grpc.DialOption{}
40: 
41: 	if usePlaintext {
42: 		if tlsCfg != nil && (tlsCfg.CAFile != "" || tlsCfg.CertFile != "" || tlsCfg.KeyFile != "") {
43: 			return nil, fmt.Errorf("grpc connect: TLS parameters (CAFile/CertFile/KeyFile) are ignored with plaintext (http://) address")
44: 		}
45: 		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
46: 	} else if tlsCfg != nil {
47: 		creds, err := buildTLSCredentials(tlsCfg)
48: 		if err != nil {
49: 			return nil, fmt.Errorf("tls config: %w", err)
50: 		}
51: 		opts = append(opts, grpc.WithTransportCredentials(creds))
52: 	} else {
53: 		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})))
54: 	}
55: 
56: 	if auth != nil {
57: 		if usePlaintext && auth.RequireTransportSecurity() {
58: 			return nil, fmt.Errorf("grpc connect: auth provider requires transport security but address uses plaintext (http://)")
59: 		}
60: 		opts = append(opts, grpc.WithPerRPCCredentials(auth))
61: 	}
62: 
63: 	conn, err := grpc.NewClient(address, opts...)
64: 	if err != nil {
65: 		return nil, fmt.Errorf("grpc connect: %w", err)
66: 	}
67: 	return conn, nil
68: }
69: 
70: func buildTLSCredentials(cfg *TLSParams) (credentials.TransportCredentials, error) {
71: 	tlsConfig := &tls.Config{
72: 		MinVersion:         tls.VersionTLS12,
73: 		InsecureSkipVerify: cfg.Insecure, //nolint:gosec // user-requested skip for dev gateways
74: 	}
75: 
76: 	if cfg.CAFile != "" {
77: 		caCert, err := os.ReadFile(cfg.CAFile)
78: 		if err != nil {
79: 			return nil, fmt.Errorf("read CA file: %w", err)
80: 		}
81: 		pool := x509.NewCertPool()
82: 		if !pool.AppendCertsFromPEM(caCert) {
83: 			return nil, fmt.Errorf("invalid CA certificate")
84: 		}
85: 		tlsConfig.RootCAs = pool
86: 	}
87: 
88: 	if cfg.CertFile != "" && cfg.KeyFile != "" {
89: 		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
90: 		if err != nil {
91: 			return nil, fmt.Errorf("load client cert: %w", err)
92: 		}
93: 		tlsConfig.Certificates = []tls.Certificate{cert}
94: 	} else if cfg.CertFile != "" || cfg.KeyFile != "" {
95: 		return nil, fmt.Errorf("both CertFile and KeyFile must be provided for client certificate authentication")
96: 	}
97: 
98: 	return credentials.NewTLS(tlsConfig), nil
99: }

(End of file - total 99 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/gateway_setup.go"
}
```

**Output:**
```
<path>/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/gateway_setup.go</path>
<type>file</type>
<content>
1: package main
2: 
3: import (
4: 	"fmt"
5: 	"log/slog"
6: 	"os"
7: 	"strings"
8: 
9: 	openshell "github.com/NVIDIA/OpenShell/sdk/go/openshell/v1"
10: 
11: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/clients"
12: )
13: 
14: type gatewayClients struct {
15: 	sdk        openshell.ClientInterface
16: 	uploadExec *clients.RawExecClient
17: }
18: 
19: func (c *gatewayClients) Close() {
20: 	if c.sdk != nil {
21: 		if err := c.sdk.Close(); err != nil {
22: 			slog.Warn("SDK client close failed", "error", err)
23: 		}
24: 	}
25: 	if c.uploadExec != nil {
26: 		if err := c.uploadExec.Close(); err != nil {
27: 			slog.Warn("upload exec client close failed", "error", err)
28: 		}
29: 	}
30: }
31: 
32: func newGatewayClients(gatewayURL, gatewayCACert, gatewayClientCert, gatewayClientKey string) (*gatewayClients, error) {
33: 	useTLS := strings.HasPrefix(gatewayURL, "grpcs://") || strings.HasPrefix(gatewayURL, "https://")
34: 	sdkAddress := normalizeGatewayAddress(gatewayURL, useTLS)
35: 
36: 	if (gatewayClientCert == "") != (gatewayClientKey == "") {
37: 		return nil, fmt.Errorf("gateway mTLS requires both --gateway-client-cert and --gateway-client-key")
38: 	}
39: 
40: 	sdkCfg := openshell.Config{
41: 		Address: sdkAddress,
42: 		Auth:    clients.ContextAuthProvider{RequireTLS: useTLS},
43: 	}
44: 	if useTLS {
45: 		tlsCfg := &openshell.TLSConfig{CAFile: gatewayCACert}
46: 		if gatewayClientCert != "" {
47: 			tlsCfg.CertFile = gatewayClientCert
48: 			tlsCfg.KeyFile = gatewayClientKey
49: 		}
50: 		sdkCfg.TLS = tlsCfg
51: 	} else {
52: 		sdkCfg.TLS = &openshell.TLSConfig{Insecure: true}
53: 	}
54: 
55: 	sdkClient, err := openshell.NewClient(sdkCfg)
56: 	if err != nil {
57: 		return nil, fmt.Errorf("SDK client setup failed: %w", err)
58: 	}
59: 
60: 	rawHost := strings.TrimPrefix(strings.TrimPrefix(sdkAddress, "https://"), "http://")
61: 	uploadExec, err := clients.NewRawExecClient(rawHost, gatewayCACert, gatewayClientCert, gatewayClientKey, useTLS)
62: 	if err != nil {
63: 		if closeErr := sdkClient.Close(); closeErr != nil {
64: 			slog.Warn("SDK client close failed during setup rollback", "error", closeErr)
65: 		}
66: 		return nil, fmt.Errorf("upload exec client setup failed: %w", err)
67: 	}
68: 
69: 	return &gatewayClients{sdk: sdkClient, uploadExec: uploadExec}, nil
70: }
71: 
72: func normalizeGatewayAddress(gatewayURL string, useTLS bool) string {
73: 	switch {
74: 	case strings.HasPrefix(gatewayURL, "grpcs://"):
75: 		return "https://" + strings.TrimPrefix(gatewayURL, "grpcs://")
76: 	case strings.HasPrefix(gatewayURL, "grpc://"):
77: 		return "http://" + strings.TrimPrefix(gatewayURL, "grpc://")
78: 	case strings.HasPrefix(gatewayURL, "https://"), strings.HasPrefix(gatewayURL, "http://"):
79: 		return gatewayURL
80: 	default:
81: 		scheme := "http"
82: 		if useTLS {
83: 			scheme = "https"
84: 		}
85: 		return fmt.Sprintf("%s://%s", scheme, gatewayURL)
86: 	}
87: }
88: 
89: func warnGatewayConfig(gatewayURL, gatewayCACert string, authDisabled bool) {
90: 	if gatewayURL == defaultGatewayURL {
91: 		slog.Warn("gateway URL is the default — verify OPENSHELL_GATEWAY_URL is configured correctly", "url", gatewayURL)
92: 	}
93: 	if gatewayCACert != "" && !strings.HasPrefix(gatewayURL, "grpcs://") && !strings.HasPrefix(gatewayURL, "https://") {
94: 		slog.Warn(
95: 			"gateway CA cert is set but gateway URL has no TLS scheme; use grpcs:// or https:// for TLS gateways",
96: 			"url", gatewayURL,
97: 			"caCert", gatewayCACert,
98: 		)
99: 	}
100: 	if authDisabled {
101: 		slog.Warn("AUTH_DISABLED=true — authentication is OFF; never use this outside local development")
102: 	}
103: }
104: 
105: func exitOnError(msg string, err error) {
106: 	slog.Error(msg, "error", err)
107: 	os.Exit(1)
108: }

(End of file - total 108 lines)
</content>
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://docs.nvidia.com/openshell/latest/kubernetes/managing-certificates.md",
  "format": "markdown"
}
```

**Output:**
```
> For clean Markdown of any page, append .md to the page URL.
> For a complete documentation index, see https://docs.nvidia.com/openshell/llms.txt.
> For AI client integration (Claude Code, Cursor, etc.), connect to the MCP server at https://docs.nvidia.com/openshell/_mcp/server.

# Managing Certificates

> Configure the OpenShell Helm chart to use cert-manager for mTLS certificate issuance and automatic renewal.

The OpenShell gateway uses mTLS certificates for transport between the gateway and sandbox supervisors. These certificates are not Kubernetes user authentication; configure OIDC or a trusted access proxy for user access. The Helm chart supports two ways to provision and manage the certificate bundle:

| Mode                            | When to use                                                                                                                                 |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Built-in `pkiInitJob` (default) | The default path. A pre-install Kubernetes Job generates a self-signed CA and certificates during installation. No additional dependencies. |
| cert-manager                    | Production deployments that need automatic certificate rotation managed by a running controller.                                            |

The rest of this page covers switching to cert-manager. The built-in mode requires no configuration.

When `certManager.enabled=true`, cert-manager owns TLS certificate generation.
The chart still runs a JWT-only initialization hook because cert-manager does
not create the sandbox JWT signing Secret required by the gateway. This
cert-manager precedence applies even if `pkiInitJob.enabled` remains true.

## Install cert-manager

Install cert-manager from the OCI registry with CRD support enabled:

```shell
helm upgrade --install cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version v1.20.3 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true \
  --wait
```

Verify the cert-manager pods are running:

```shell
kubectl -n cert-manager get pods
```

## Install OpenShell with cert-manager PKI

Pass the cert-manager values override when installing or upgrading the chart:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set certManager.enabled=true
```

The chart creates a self-signed CA, issues server and client certificates from it, and cert-manager handles renewal before expiry.
The chart also runs a pre-install hook in JWT-only mode to create the gateway's
sandbox JWT signing Secret. That Secret is separate from the cert-manager TLS
certificate Secrets and is mounted at `/etc/openshell-jwt`.

## Using a real Issuer for the server certificate

By default, cert-manager issues both the server and client certificates from
a self-signed CA the chart creates — this rotates automatically, but the
server certificate is still not publicly trusted. `certManager.serverIssuerRef`
overrides the `issuerRef` on the server `Certificate` resource to point at a
real `Issuer` or `ClusterIssuer` instead, for example an ACME issuer:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set certManager.enabled=true \
  --set certManager.serverIssuerRef.name=letsencrypt-prod \
  --set certManager.serverIssuerRef.kind=ClusterIssuer \
  --set certManager.serverDnsNames[0]=openshell.example.com
```

### Dual certificate architecture

When `serverIssuerRef` is set, the chart creates **two** server certificates:

1. **Internal certificate** (`openshell-server-tls`): signed by the chart CA
   with internal SANs (`*.svc.cluster.local`, `localhost`, etc.).
2. **External certificate** (`openshell-server-external-tls`): signed by the
   configured issuer (e.g. ACME) with only the hostnames from
   `certManager.serverDnsNames`.

The gateway uses **SNI** to select which certificate to present:
supervisors connect via internal service names and receive the internal
certificate (verified against the chart CA they already trust), while CLI
users connecting through a Route or ingress use the external hostname and
receive the ACME certificate. This keeps supervisor trust pinned to only
the operator's chart CA — no WebPKI root trust is needed.

Public CAs such as Let's Encrypt reject certificate requests that include
internal-only names per CA/Browser Forum baseline requirements. The chart
validates this at install time and fails with an actionable error if
`certManager.serverDnsNames` contains internal-only entries while
`serverIssuerRef` is set.

You do **not** need to set `server.grpcEndpoint` to the external hostname.
Supervisors connect via the internal service name automatically. Setting
`server.grpcEndpoint` to an external hostname would cause supervisors to
receive the ACME certificate (via SNI) which they cannot verify against the
chart CA.

The default `clientCaFromServerTlsSecret=true` is correct even when
`serverIssuerRef` is set: the internal server certificate is always signed
by the chart CA (the same CA that signs the client certificate), so its
`ca.crt` is the right trust anchor for mTLS verification.

## Next Steps

Return to [Setup](/kubernetes/setup) to complete the installation. For
exposing the gateway externally on OpenShift with a real certificate, see
[OpenShift](/kubernetes/openshift).
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://docs.nvidia.com/openshell/latest/kubernetes/mesh",
  "format": "markdown"
}
```

**Output:**
```
# Page Not Found

This page does not exist.

## Similar pages

- [Setup](https://docs.nvidia.com/openshell/openshell/kubernetes/setup.md)
- [Ingress](https://docs.nvidia.com/openshell/openshell/kubernetes/ingress.md)
- [OpenShift](https://docs.nvidia.com/openshell/openshell/kubernetes/openshift.md)

```

---

## Assistant (Build · GPT-5.6 Terra · 7.7s)

_Thinking:_



SDK supports CA plus client cert/key; both its gRPC client and raw-exec client require same bundle. Next: verify forwarded listener presents OpenShell TLS and accepts client cert.

**Tool: bash**

**Input:**
```json
{
  "workdir": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "command": "openssl s_client -connect 127.0.0.1:8080 -servername localhost -brief </dev/null",
  "timeout": 10000
}
```

**Output:**
```
C02163FB01000000:error:8000003D:system library:BIO_connect:Connection refused:crypto/bio/bio_sock2.c:183:calling connect()
C02163FB01000000:error:10000067:BIO routines:BIO_connect:connect error:crypto/bio/bio_sock2.c:185:
connect:errno=61

```

**Tool: bash**

**Input:**
```json
{
  "workdir": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "command": "lsof -nP -iTCP:8080 -sTCP:LISTEN",
  "timeout": 10000
}
```

**Output:**
```
(no output)
```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "pattern": "NewGatewayClients\\(",
  "include": "*.go"
}
```

**Output:**
```
Found 2 matches
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/clients/gateway.go:
  Line 46: func NewGatewayClients(gatewayURL, caCert, clientCert, clientKey string) (*GatewayClients, error) {


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/gateway_setup.go:
  Line 62: 	appClients, err := testclients.NewGatewayClients(

```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "pattern": "GatewayConfig\\{",
  "include": "*.go"
}
```

**Output:**
```
Found 3 matches
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/fleet/registry_test.go:
  Line 24: 	created, err := registry.Register(context.Background(), GatewayConfig{ID: "openshell"})


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/pkg/discovery/discovery.go:
  Line 30: 			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{

  Line 35: 			tempAddGateway(ctx, gatewayRegistry, fleet.GatewayConfig{

```

---

## Assistant (Build · GPT-5.6 Terra · 37.8s)

_Thinking:_



_Thinking:_



_Thinking:_

**Evaluating certificate options**

I'm looking into the `kubectl` command for service management, specifically for an openshell service. I need to consider the cert manager case too. I'm weighing the default built-in CA against the self-signed internal issuer supported by cert manager. I wonder if both options might involve different client secret names. It seems like I have to clarify that aspect to ensure proper configuration and understanding moving forward.

_Thinking:_



**Finding**

Current shell has no listener on `127.0.0.1:8080`; TLS probe got `connection refused`.

OpenShell default port-forward needs OpenShell mTLS bundle, not cluster/service-mesh CA:

```sh
kubectl -n <openshell-namespace> port-forward svc/openshell 8080:8080
```

Extract bundle:

```sh
umask 077
mkdir -p ~/.config/openshell/gateways/agent-ops/mtls

kubectl -n <openshell-namespace> get secret openshell-client-tls \
  -o jsonpath='{.data.ca\.crt}' | base64 -d \
  > ~/.config/openshell/gateways/agent-ops/mtls/ca.crt

kubectl -n <openshell-namespace> get secret openshell-client-tls \
  -o jsonpath='{.data.tls\.crt}' | base64 -d \
  > ~/.config/openshell/gateways/agent-ops/mtls/tls.crt

kubectl -n <openshell-namespace> get secret openshell-client-tls \
  -o jsonpath='{.data.tls\.key}' | base64 -d \
  > ~/.config/openshell/gateways/agent-ops/mtls/tls.key
```

Verify TLS plus client authentication:

```sh
openssl s_client -connect 127.0.0.1:8080 -servername localhost \
  -CAfile ~/.config/openshell/gateways/agent-ops/mtls/ca.crt \
  -cert ~/.config/openshell/gateways/agent-ops/mtls/tls.crt \
  -key ~/.config/openshell/gateways/agent-ops/mtls/tls.key \
  -verify_return_error -alpn h2 </dev/null
```

**Cert-manager / mesh**

- cert-manager optional. Built-in OpenShell PKI works.
- With `certManager.enabled=true`, OpenShell still creates `openshell-client-tls`; use it same way.
- Service mesh mTLS separate from OpenShell gateway TLS. Do not use mesh CA to validate gateway certificate.
- Port-forward uses OpenShell cert SANs for `localhost` and `127.0.0.1`, per docs.
- For BFF deployed in cluster, use `https://openshell.<namespace>.svc.cluster.local:8080`, with same CA/client bundle mounted into BFF pod. Do not target `localhost`.
- cert-manager rotation requires BFF restart under current implementation: gateway clients load PEM files once during fleet registration.

**Current backend blocker**

TLS cannot work through config only:

- `pkg/discovery/discovery.go:32` hardcodes `https://localhost:8080`.
- `cmd/gateway_setup.go:62-67` creates gateway clients with empty CA, client cert, client key.
- `BUNDLE_PATHS` and `INSECURE_SKIP_VERIFY` do not reach this gateway client path.
- `internal/clients/gateway.go:60-64` supports all needed files, but caller supplies none.

Expected failures: untrusted private CA, then client-certificate-required if server mTLS enforced. `InsecureSkipVerify` would not solve missing client certificate and should not be used.

Separate next issue: `cmd/gateway_setup.go:57` uses `openshell.NoAuth()`. If gateway uses OIDC, TLS will pass but API calls still need forwarded bearer-token auth.

No files changed.

---

## User

ignore the fact that the portforward is not on. does it make sense to have the bff which is supposed to talk to the gateway literally take all the private secrets/certs of the gateway? 

---

## Assistant (Build · GPT-5.6 Terra · 21.5s)

_Thinking:_

**Analyzing authentication options**

I need to consider certificate resources, but for a general answer, I don't need detailed specifics. The user mentioned a BFF that should interact with a gateway. Authentication should use an OIDC bearer token instead of a shared OpenShell mTLS certificate. I see that OpenShell documentation states mTLS is for transport, not for Kubernetes user authentication. If the gateway is set up for OIDC, then it shouldn't need a client private cert. I’ll need to clarify the details and confirm my understanding.

_Thinking:_

**Evaluating BFF and TLS setup**

It seems that the BFF shouldn't require a certificate unless it's functioning as an internal supervisor. I'm wondering if the gateway TLS universally requires a client certificate, or if it might permit OIDC without it. The setup instructions mention copying the generated bundle to the CLI for verification, which likely includes the TLS cert and key for mTLS. However, the documentation clarifies that this bundle is internal and doesn't authenticate users. I need to check the server configuration and handshake for clarity.

_Thinking:_



No. BFF should never receive gateway server key, CA/issuer private key, JWT-signing Secret, or arbitrary gateway Secrets.

Recommended path:

- BFF trusts gateway with public CA bundle only.
- Browser user token reaches BFF.
- BFF validates token, forwards it to OpenShell.
- OpenShell uses same OIDC issuer/audience and authorizes user request.
- Service mesh mTLS handles workload-to-workload transport separately.

`openshell-client-tls` contains client private key. It can make local port-forward testing work, but sharing it with BFF makes BFF a holder of shared OpenShell mTLS identity. Bad production design: no caller attribution, hard rotation, credential reuse across supervisors/BFF.

If gateway TLS policy truly requires client mTLS from BFF, provision dedicated BFF client cert/key:

- Same trusted client CA used by gateway.
- Secret mounted read-only only into BFF service account.
- Namespace-scoped RBAC permits BFF to read only this Secret.
- cert-manager rotates it.
- Never expose it to frontend or pass through API responses.

OpenShell docs distinguish mTLS transport from user identity. For normal BFF-to-gateway requests, prefer OIDC bearer forwarding. Current BFF uses `openshell.NoAuth()`, so this needs fixing after TLS; otherwise OIDC-enabled gateway rejects calls even when TLS succeeds.

---

## User

so how do i make this code work with best production practice?

---

## Assistant (Build · GPT-5.6 Terra · 7.2s)

_Thinking:_

**Considering code changes**

I’m thinking about whether code changes need to be implemented based on the user’s question about architecture. The developer mentions following the default unless a specific plan is provided. I need to propose exact code changes carefully, especially due to the ambiguity surrounding the OIDC infra and BFF gateway configuration. I should inspect the auth middleware to see if the user token is present. The user likely expects a design rather than actual code changes, so I’ll need to investigate accordingly.

Map existing request-auth path, then give production wiring with smallest required changes.

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/config/environment.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/config/environment.go</path>
<type>file</type>
<content>
1: package config
2: 
3: import (
4: 	"fmt"
5: 	"log/slog"
6: 	"strings"
7: )
8: 
9: const (
10: 	// AuthMethodDisabled authentication is disabled, useful for testing and mock-only local dev.
11: 	AuthMethodDisabled = "disabled"
12: 
13: 	// AuthMethodInternal uses the credentials of the running backend.
14: 	// If running inside the cluster, it uses the pod's service account.
15: 	// If running locally (e.g. for development), it uses the current user's kubeconfig context.
16: 	// This is the default authentication method.
17: 	AuthMethodInternal = "internal"
18: 
19: 	// AuthMethodUser uses a user-provided Bearer token for authentication.
20: 	AuthMethodUser = "user_token"
21: 
22: 	// DefaultAuthTokenHeader is the standard header for Bearer token auth.
23: 	DefaultAuthTokenHeader = "Authorization"
24: 
25: 	// DefaultAuthTokenPrefix is the prefix used in the Authorization header.
26: 	// note: the space here is intentional, as the prefix is "Bearer " (with a space).
27: 	DefaultAuthTokenPrefix = "Bearer "
28: )
29: 
30: // DeploymentMode represents the deployment mode enum
31: type DeploymentMode string
32: 
33: const (
34: 	// DeploymentModeFederated represents the federated platform mode
35: 	DeploymentModeFederated DeploymentMode = "federated"
36: 	// DeploymentModeStandalone represents the standalone mode
37: 	DeploymentModeStandalone DeploymentMode = "standalone"
38: )
39: 
40: // String implements the fmt.Stringer interface
41: func (d DeploymentMode) String() string {
42: 	return string(d)
43: }
44: 
45: // Set implements the flag.Value interface
46: func (d *DeploymentMode) Set(value string) error {
47: 	switch strings.ToLower(value) {
48: 	case "federated":
49: 		*d = DeploymentModeFederated
50: 	case "standalone", "kubeflow":
51: 		// kubeflow is accepted as an alias for standalone (legacy Makefile / frontend naming).
52: 		*d = DeploymentModeStandalone
53: 	default:
54: 		return fmt.Errorf("invalid deployment mode: %s (must be federated, standalone, or kubeflow)", value)
55: 	}
56: 	return nil
57: }
58: 
59: // IsStandaloneMode returns true if the deployment mode is standalone
60: func (d DeploymentMode) IsStandaloneMode() bool {
61: 	return d == DeploymentModeStandalone
62: }
63: 
64: // IsFederatedMode returns true if the deployment mode is federated
65: func (d DeploymentMode) IsFederatedMode() bool {
66: 	return d == DeploymentModeFederated
67: }
68: 
69: type EnvConfig struct {
70: 	Port               int
71: 	MockK8Client       bool
72: 	MockHTTPClient     bool
73: 	DevMode            bool
74: 	DeploymentMode     DeploymentMode
75: 	DevModeClientPort  int
76: 	DevModeCatalogPort int
77: 	StaticAssetsDir    string
78: 	LogLevel           slog.Level
79: 	AllowedOrigins     []string
80: 	// BundlePaths is a list of filesystem paths to PEM-encoded CA bundle files.
81: 	// If provided, the application will attempt to load these files and add the
82: 	// certificates to the HTTP client's Root CAs for outbound TLS connections.
83: 	// Missing or unreadable files are ignored.
84: 	BundlePaths []string
85: 
86: 	// ─── AUTH ───────────────────────────────────────────────────
87: 	// Specifies the authentication method used by the server.
88: 	// Valid values: "internal", "user_token", or "disabled"
89: 	AuthMethod string
90: 
91: 	// Header used to extract the authentication token.
92: 	// Default is "Authorization" and can be overridden via CLI/env for proxy integration scenarios.
93: 	AuthTokenHeader string
94: 
95: 	// Optional prefix to strip from the token header value.
96: 	// Default is "Bearer ", can be set to empty if the token is sent without a prefix.
97: 	AuthTokenPrefix string
98: 
99: 	// ─── TLS ────────────────────────────────────────────────────
100: 	// TLS verification settings for HTTP client connections to the Client
101: 	// InsecureSkipVerify when true, skips TLS certificate verification (useful for development/local setups)
102: 	// Default is false (secure) for production environments
103: 	InsecureSkipVerify bool
104: 
105: 	// ─── BFF INTER-COMMUNICATION ─────────────────────────────────
106: 	// MockBFFClients enables mock mode for BFF inter-communication clients.
107: 	// When true, BFF clients return mock responses instead of making real HTTP calls.
108: 	MockBFFClients bool
109: 
110: 	// MockAgentClient enables mock mode for the agent data source client.
111: 	// When true, agent APIs return built-in demo data instead of querying the cluster.
112: 	// Local development only — routes are live with no RBAC gate; do not enable in staging or production.
113: 	MockAgentClient bool
114: 
115: 	// ─── DEPRECATED ─────────────────────────────────────────────
116: 	// The following fields are deprecated and maintained for backward compatibility
117: 	// Use DeploymentMode instead
118: 	StandaloneMode    bool
119: 	FederatedPlatform bool
120: }

(End of file - total 120 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/middleware.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/middleware.go</path>
<type>file</type>
<content>
1: package api
2: 
3: import (
4: 	"context"
5: 	"fmt"
6: 	"log/slog"
7: 	"net/http"
8: 	"runtime/debug"
9: 	"strings"
10: 
11: 	"github.com/google/uuid"
12: 	"github.com/julienschmidt/httprouter"
13: 	"github.com/opendatahub-io/agent-ops/internal/config"
14: 	"github.com/opendatahub-io/agent-ops/internal/constants"
15: 	helper "github.com/opendatahub-io/agent-ops/internal/helpers"
16: 	k8s "github.com/opendatahub-io/agent-ops/internal/integrations/kubernetes"
17: 	"github.com/rs/cors"
18: )
19: 
20: func (app *App) RecoverPanic(next http.Handler) http.Handler {
21: 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
22: 		defer func() {
23: 			if err := recover(); err != nil {
24: 				w.Header().Set("Connection", "close")
25: 				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
26: 				app.logger.Error("Recovered from panic", slog.String("stack_trace", string(debug.Stack())))
27: 			}
28: 		}()
29: 
30: 		next.ServeHTTP(w, r)
31: 	})
32: }
33: 
34: func (app *App) InjectRequestIdentity(next http.Handler) http.Handler {
35: 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
36: 		//skip use headers check if we are not on /api/v1 (i.e. we are on /healthcheck and / (static fe files) )
37: 		if !strings.HasPrefix(r.URL.Path, ApiPathPrefix) && !strings.HasPrefix(r.URL.Path, PathPrefix+ApiPathPrefix) {
38: 			next.ServeHTTP(w, r)
39: 			return
40: 		}
41: 
42: 		// If authentication is disabled, skip identity extraction.
43: 		if app.config.AuthMethod == config.AuthMethodDisabled {
44: 			next.ServeHTTP(w, r)
45: 			return
46: 		}
47: 
48: 		identity, error := app.kubernetesClientFactory.ExtractRequestIdentity(r.Header)
49: 		if error != nil {
50: 			app.unauthorizedResponse(w, r, error)
51: 			return
52: 		}
53: 
54: 		ctx := context.WithValue(r.Context(), constants.RequestIdentityKey, identity)
55: 		next.ServeHTTP(w, r.WithContext(ctx))
56: 	})
57: }
58: 
59: func (app *App) EnableCORS(next http.Handler) http.Handler {
60: 	if len(app.config.AllowedOrigins) == 0 {
61: 		// CORS is disabled, this middleware becomes a noop.
62: 		return next
63: 	}
64: 
65: 	c := cors.New(cors.Options{
66: 		AllowedOrigins:     app.config.AllowedOrigins,
67: 		AllowCredentials:   true,
68: 		AllowedMethods:     []string{"GET", "PUT", "POST", "PATCH", "DELETE"},
69: 		AllowedHeaders:     []string{constants.KubeflowUserIDHeader, constants.KubeflowUserGroupsIdHeader},
70: 		Debug:              app.config.LogLevel == slog.LevelDebug,
71: 		OptionsPassthrough: false,
72: 	})
73: 
74: 	return c.Handler(next)
75: }
76: 
77: func (app *App) EnableTelemetry(next http.Handler) http.Handler {
78: 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
79: 		// Adds a unique id to the context to allow tracing of requests
80: 		traceId := uuid.NewString()
81: 		ctx := context.WithValue(r.Context(), constants.TraceIdKey, traceId)
82: 
83: 		// logger will only be nil in tests.
84: 		if app.logger != nil {
85: 			traceLogger := app.logger.With(slog.String("trace_id", traceId))
86: 			ctx = context.WithValue(ctx, constants.TraceLoggerKey, traceLogger)
87: 
88: 			traceLogger.Debug("Incoming HTTP request", slog.Any("request", helper.RequestLogValuer{Request: r}))
89: 		}
90: 		next.ServeHTTP(w, r.WithContext(ctx))
91: 	})
92: }
93: 
94: func (app *App) AttachNamespace(next func(http.ResponseWriter, *http.Request, httprouter.Params)) httprouter.Handle {
95: 	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
96: 		namespace := r.URL.Query().Get(string(constants.NamespaceHeaderParameterKey))
97: 		if namespace == "" {
98: 			app.badRequestResponse(w, r, fmt.Errorf("missing required query parameter: %s", constants.NamespaceHeaderParameterKey))
99: 			return
100: 		}
101: 		if !isValidDNS1123Label(namespace) {
102: 			app.badRequestResponse(w, r, fmt.Errorf("invalid namespace %q", namespace))
103: 			return
104: 		}
105: 
106: 		ctx := context.WithValue(r.Context(), constants.NamespaceHeaderParameterKey, namespace)
107: 		r = r.WithContext(ctx)
108: 
109: 		next(w, r, ps)
110: 	}
111: }
112: 
113: // AttachNamespaceFromParam reads a namespace from an httprouter path parameter and injects it
114: // into the request context under NamespaceHeaderParameterKey.
115: // This middleware must be placed before RequireAccessToService for path-scoped routes.
116: func (app *App) AttachNamespaceFromParam(paramName string, next func(http.ResponseWriter, *http.Request, httprouter.Params)) httprouter.Handle {
117: 	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
118: 		namespace := ps.ByName(paramName)
119: 		if namespace == "" {
120: 			app.badRequestResponse(w, r, fmt.Errorf("missing required path parameter: %s", paramName))
121: 			return
122: 		}
123: 		if !isValidDNS1123Label(namespace) {
124: 			app.badRequestResponse(w, r, fmt.Errorf("invalid namespace %q", namespace))
125: 			return
126: 		}
127: 
128: 		ctx := context.WithValue(r.Context(), constants.NamespaceHeaderParameterKey, namespace)
129: 		r = r.WithContext(ctx)
130: 
131: 		next(w, r, ps)
132: 	}
133: }
134: 
135: // authenticateAgentRequest validates caller identity for agent routes when auth is enabled.
136: func (app *App) authenticateAgentRequest(w http.ResponseWriter, r *http.Request) (*k8s.RequestIdentity, bool) {
137: 	if app.config.AuthMethod == config.AuthMethodDisabled {
138: 		return nil, true
139: 	}
140: 
141: 	identity, ok := r.Context().Value(constants.RequestIdentityKey).(*k8s.RequestIdentity)
142: 	if !ok || identity == nil {
143: 		app.unauthorizedResponse(w, r, fmt.Errorf("missing RequestIdentity in context"))
144: 		return nil, false
145: 	}
146: 
147: 	if app.kubernetesClientFactory == nil {
148: 		app.serverErrorResponse(w, r, fmt.Errorf("kubernetes client factory is not configured"))
149: 		return nil, false
150: 	}
151: 
152: 	if err := app.kubernetesClientFactory.ValidateRequestIdentity(identity); err != nil {
153: 		app.unauthorizedResponse(w, r, err)
154: 		return nil, false
155: 	}
156: 
157: 	return identity, true
158: }
159: 
160: // RequireAuthenticatedForAgents ensures agent API routes receive a valid identity when auth is enabled.
161: // Namespace-scoped SSAR filtering for list results is performed in the Kubernetes agent client.
162: func (app *App) RequireAuthenticatedForAgents(next func(http.ResponseWriter, *http.Request, httprouter.Params)) httprouter.Handle {
163: 	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
164: 		if _, ok := app.authenticateAgentRequest(w, r); !ok {
165: 			return
166: 		}
167: 		next(w, r, ps)
168: 	}
169: }
170: 
171: // RequireAccessToService validates identity and checks whether the user can list services in the
172: // namespace injected by AttachNamespace or AttachNamespaceFromParam.
173: // This middleware must be placed after InjectRequestIdentity and a namespace-attachment middleware.
174: func (app *App) RequireAccessToService(next func(http.ResponseWriter, *http.Request, httprouter.Params)) httprouter.Handle {
175: 	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
176: 		if app.config.AuthMethod == config.AuthMethodDisabled {
177: 			next(w, r, ps)
178: 			return
179: 		}
180: 
181: 		ctx := r.Context()
182: 
183: 		identity, ok := app.authenticateAgentRequest(w, r)
184: 		if !ok {
185: 			return
186: 		}
187: 
188: 		namespace, ok := ctx.Value(constants.NamespaceHeaderParameterKey).(string)
189: 		if !ok || namespace == "" {
190: 			app.badRequestResponse(w, r, fmt.Errorf("missing namespace in context"))
191: 			return
192: 		}
193: 
194: 		k8sClient, err := app.kubernetesClientFactory.GetClient(ctx)
195: 		if err != nil {
196: 			app.serverErrorResponse(w, r, fmt.Errorf("failed to get Kubernetes client: %w", err))
197: 			return
198: 		}
199: 
200: 		allowed, err := k8sClient.CanListServicesInNamespace(ctx, identity, namespace)
201: 		if err != nil {
202: 			app.serverErrorResponse(w, r, fmt.Errorf("failed to check namespace access: %w", err))
203: 			return
204: 		}
205: 
206: 		if !allowed {
207: 			app.forbiddenResponse(w, r, "user does not have permission to access services in this namespace")
208: 			return
209: 		}
210: 
211: 		logger := helper.GetContextLoggerFromReq(r)
212: 		logger.Debug("User authorized to access services in namespace", "namespace", namespace)
213: 
214: 		next(w, r, ps)
215: 	}
216: }

(End of file - total 216 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/types.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/types.go</path>
<type>file</type>
<content>
1: package kubernetes
2: 
3: type ServiceDetails struct {
4: 	Name                string
5: 	DisplayName         string
6: 	Description         string
7: 	ClusterIP           string
8: 	HTTPPort            int32
9: 	IsHTTPS             bool
10: 	ExternalAddressRest string
11: }
12: 
13: type RequestIdentity struct {
14: 	UserID string
15: 	Groups []string
16: 	Token  string
17: }
18: 
19: type BearerToken struct {
20: 	raw string
21: }
22: 
23: func NewBearerToken(t string) BearerToken {
24: 	return BearerToken{raw: t}
25: }
26: 
27: func (t BearerToken) String() string {
28: 	return "[REDACTED]"
29: }
30: 
31: func (t BearerToken) Raw() string {
32: 	return t.raw
33: }

(End of file - total 33 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/auth.go"
}
```

**Output:**
```
<path>/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/auth.go</path>
<type>file</type>
<content>
1: package clients
2: 
3: import (
4: 	"context"
5: 
6: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/auth"
7: )
8: 
9: // ContextAuthProvider implements grpc.PerRPCCredentials by reading the JWT
10: // from the request context on every gRPC call. This allows a single shared
11: // SDK client to forward per-request user tokens to the gateway.
12: type ContextAuthProvider struct {
13: 	RequireTLS bool
14: }
15: 
16: func (c ContextAuthProvider) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
17: 	token := auth.TokenFromContext(ctx)
18: 	if token == "" {
19: 		return nil, nil
20: 	}
21: 	return map[string]string{"authorization": "Bearer " + token}, nil
22: }
23: 
24: func (c ContextAuthProvider) RequireTransportSecurity() bool {
25: 	return c.RequireTLS
26: }

(End of file - total 26 lines)
</content>
```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "pattern": "AuthToken|Authorization|RequestIdentity|ContextAuthProvider|AuthMethod",
  "include": "*.go"
}
```

**Output:**
```
Found 100 matches (more matches available)
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/helpers/logging.go:
  Line 31: 	"Authorization",

  Line 34: 	"Proxy-Authorization",


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/clients/gateway.go:
  Line 17: // clients differ only in endpoint, because ContextAuthProvider resolves the

  Line 56: 		// Auth:    clients.ContextAuthProvider{RequireTLS: useTLS},


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/config/environment.go:
  Line 10: 	// AuthMethodDisabled authentication is disabled, useful for testing and mock-only local dev.

  Line 11: 	AuthMethodDisabled = "disabled"

  Line 13: 	// AuthMethodInternal uses the credentials of the running backend.

  Line 17: 	AuthMethodInternal = "internal"

  Line 19: 	// AuthMethodUser uses a user-provided Bearer token for authentication.

  Line 20: 	AuthMethodUser = "user_token"

  Line 22: 	// DefaultAuthTokenHeader is the standard header for Bearer token auth.

  Line 23: 	DefaultAuthTokenHeader = "Authorization"

  Line 25: 	// DefaultAuthTokenPrefix is the prefix used in the Authorization header.

  Line 27: 	DefaultAuthTokenPrefix = "Bearer "

  Line 89: 	AuthMethod string

  Line 92: 	// Default is "Authorization" and can be overridden via CLI/env for proxy integration scenarios.

  Line 93: 	AuthTokenHeader string

  Line 97: 	AuthTokenPrefix string


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/constants/keys.go:
  Line 11: 	RequestIdentityKey contextKey = "requestIdentityKey"

  Line 13: 	// For config.AuthMethodInternal


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/internal_k8s_client.go:
  Line 73: func (kc *InternalKubernetesClient) GetNamespaces(ctx context.Context, identity *RequestIdentity) ([]corev1.Namespace, error) {

  Line 142: 					response, err := kc.authorizationClient().AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})

  Line 210: func (kc *InternalKubernetesClient) IsClusterAdmin(identity *RequestIdentity) (bool, error) {

  Line 235: func (kc *InternalKubernetesClient) GetUser(identity *RequestIdentity) (string, error) {

  Line 240: func (kc *InternalKubernetesClient) CanListServicesInNamespace(ctx context.Context, identity *RequestIdentity, namespace string) (bool, error) {

  Line 257: 		response, err := kc.authorizationClient().AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/token_k8s_client.go:
  Line 25: func (kc *TokenKubernetesClient) IsClusterAdmin(_ *RequestIdentity) (bool, error) {

  Line 43: 	resp, err := kc.Client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})

  Line 113: // RequestIdentity is unused because the token already represents the user identity.

  Line 114: func (kc *TokenKubernetesClient) CanListServicesInNamespace(ctx context.Context, _ *RequestIdentity, namespace string) (bool, error) {

  Line 129: 		resp, err := kc.Client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})

  Line 144: // RequestIdentity is unused because the token already represents the user identity.

  Line 145: func (kc *TokenKubernetesClient) CanAccessServiceInNamespace(ctx context.Context, _ *RequestIdentity, namespace, serviceName string) (bool, error) {

  Line 160: 	resp, err := kc.Client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})

  Line 173: // RequestIdentity is unused because the token already represents the user identity.

  Line 175: func (kc *TokenKubernetesClient) GetNamespaces(ctx context.Context, _ *RequestIdentity) ([]corev1.Namespace, error) {

  Line 188: func (kc *TokenKubernetesClient) GetUser(_ *RequestIdentity) (string, error) {


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/agent_rbac.go:
  Line 14: 	identity *RequestIdentity,

  Line 40: 	response, err := kc.authorizationClient().AuthorizationV1().SubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})

  Line 72: 	resp, err := kc.Client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/factory.go:
  Line 18: 	switch cfg.AuthMethod {

  Line 20: 	case config.AuthMethodInternal:

  Line 27: 	case config.AuthMethodUser:

  Line 31: 	case config.AuthMethodDisabled:

  Line 35: 		return nil, fmt.Errorf("invalid auth method: %q", cfg.AuthMethod)

  Line 45: 	ExtractRequestIdentity(httpHeader http.Header) (*RequestIdentity, error)

  Line 46: 	ValidateRequestIdentity(identity *RequestIdentity) error

  Line 72: 	val := ctx.Value(constants.RequestIdentityKey)

  Line 74: 		return nil, fmt.Errorf("missing RequestIdentity in context")

  Line 77: 	identity, ok := val.(*RequestIdentity)

  Line 104: func (f *StaticClientFactory) ExtractRequestIdentity(httpHeader http.Header) (*RequestIdentity, error) {

  Line 122: 	identity := &RequestIdentity{

  Line 129: func (f *StaticClientFactory) ValidateRequestIdentity(identity *RequestIdentity) error {

  Line 168: 		Header:                     cfg.AuthTokenHeader,

  Line 169: 		Prefix:                     cfg.AuthTokenPrefix,

  Line 174: func (f *TokenClientFactory) ExtractRequestIdentity(httpHeader http.Header) (*RequestIdentity, error) {

  Line 188: 	return &RequestIdentity{

  Line 193: func (f *TokenClientFactory) ValidateRequestIdentity(identity *RequestIdentity) error {

  Line 207: 	identityVal := ctx.Value(constants.RequestIdentityKey)

  Line 209: 		return nil, fmt.Errorf("missing RequestIdentity in context")

  Line 212: 	identity, ok := identityVal.(*RequestIdentity)


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/agent_enrichment_rbac_test.go:
  Line 24: 		&RequestIdentity{UserID: "user@example.com"},

  Line 45: 		&RequestIdentity{UserID: "user@example.com"},


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/agent_enrichment_rbac.go:
  Line 23: 	identity *RequestIdentity,

  Line 39: 	_ *RequestIdentity,


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/client.go:
  Line 15: 	GetNamespaces(ctx context.Context, identity *RequestIdentity) ([]corev1.Namespace, error)

  Line 16: 	IsClusterAdmin(identity *RequestIdentity) (bool, error)

  Line 17: 	GetUser(identity *RequestIdentity) (string, error)

  Line 20: 	CanListServicesInNamespace(ctx context.Context, identity *RequestIdentity, namespace string) (bool, error)

  Line 22: 	CanAccessAgentCardEnrichment(ctx context.Context, identity *RequestIdentity, namespace string) (AgentCardEnrichmentAccess, error)


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/k8mocks/mock_factory.go:
  Line 23: 	switch cfg.AuthMethod {

  Line 24: 	case config.AuthMethodInternal:

  Line 31: 	case config.AuthMethodUser:

  Line 38: 	case config.AuthMethodDisabled:

  Line 42: 		return nil, fmt.Errorf("invalid auth method: %q", cfg.AuthMethod)

  Line 86: func (f *MockedStaticClientFactory) ExtractRequestIdentity(httpHeader http.Header) (*k8s.RequestIdentity, error) {

  Line 87: 	return f.realFactoryWithoutClient.ExtractRequestIdentity(httpHeader)

  Line 89: func (f *MockedStaticClientFactory) ValidateRequestIdentity(identity *k8s.RequestIdentity) error {

  Line 90: 	return f.realFactoryWithoutClient.ValidateRequestIdentity(identity)

  Line 113: 		AuthMethod:      config.AuthMethodUser,

  Line 114: 		AuthTokenHeader: config.DefaultAuthTokenHeader,

  Line 115: 		AuthTokenPrefix: config.DefaultAuthTokenPrefix,

  Line 128: func (f *MockedTokenClientFactory) ExtractRequestIdentity(httpHeader http.Header) (*k8s.RequestIdentity, error) {

  Line 129: 	return f.realK8sFactory.ExtractRequestIdentity(httpHeader)

  Line 132: func (f *MockedTokenClientFactory) ValidateRequestIdentity(identity *k8s.RequestIdentity) error {

  Line 133: 	return f.realK8sFactory.ValidateRequestIdentity(identity)

  Line 139: 	val := ctx.Value(constants.RequestIdentityKey)

  Line 141: 		return nil, fmt.Errorf("missing RequestIdentity in context")

  Line 144: 	identity, ok := val.(*k8s.RequestIdentity)


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/deploy_handler_test.go:
  Line 21: 		config:       config.EnvConfig{AuthMethod: config.AuthMethodDisabled},


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/types.go:
  Line 13: type RequestIdentity struct {


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/main.go:
  Line 47: 	flag.StringVar(&cfg.AuthMethod, "auth-method", "internal", "Authentication method (internal, user_token, or disabled)")

  Line 48: 	flag.StringVar(&cfg.AuthTokenHeader, "auth-token-header", getEnvAsString("AUTH_TOKEN_HEADER", config.DefaultAuthTokenHeader), "Header used to extract the token (e.g., Authorization)")

  Line 49: 	flag.StringVar(&cfg.AuthTokenPrefix, "auth-token-prefix", getEnvAsString("AUTH_TOKEN_PREFIX", config.DefaultAuthTokenPrefix), "Prefix used in the token header (e.g., 'Bearer ')")

  Line 84: 	if cfg.AuthMethod != config.AuthMethodInternal &&

  Line 85: 		cfg.AuthMethod != config.AuthMethodUser &&

  Line 86: 		cfg.AuthMethod != config.AuthMethodDisabled {

  Line 87: 		logger.Error("invalid auth method: (must be internal, user_token, or disabled)", "authMethod", cfg.AuthMethod)

  Line 90: 	if cfg.AuthMethod == config.AuthMethodDisabled {


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/namespaces_handler.go:
  Line 18: 	if app.config.AuthMethod == config.AuthMethodDisabled {


(Results truncated. Consider using a more specific path or pattern.)
```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c",
  "pattern": "type ContextAuthProvider|ContextAuthProvider|Authorization",
  "include": "*.go"
}
```

**Output:**
```
Found 19 matches
/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/rawexec.go:
  Line 62: 		grpc.WithPerRPCCredentials(ContextAuthProvider{RequireTLS: useTLS}),


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/auth.go:
  Line 9: // ContextAuthProvider implements grpc.PerRPCCredentials by reading the JWT

  Line 12: type ContextAuthProvider struct {

  Line 16: func (c ContextAuthProvider) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {

  Line 24: func (c ContextAuthProvider) RequireTransportSecurity() bool {


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/sdkclient/auth.go:
  Line 9: // ContextAuthProvider implements grpc.PerRPCCredentials by reading the JWT

  Line 12: type ContextAuthProvider struct {

  Line 16: func (c ContextAuthProvider) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {

  Line 24: func (c ContextAuthProvider) RequireTransportSecurity() bool {


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/auth/proxy_test.go:
  Line 65: 	req.Header.Set("Authorization", "Bearer bearer-jwt")

  Line 85: 	req.Header.Set("Authorization", "Bearer bearer-jwt")

  Line 95: func TestHandler_MalformedAuthorizationRejected(t *testing.T) {

  Line 103: 	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

  Line 109: 		t.Fatalf("status = %d, want 401 for non-Bearer Authorization", w.Code)

  Line 124: 	req.Header.Set("Authorization", "Bearer stray-bearer")


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/auth/proxy.go:
  Line 8: //  2. An explicit `Authorization: Bearer` header (API clients, tests).

  Line 78: 			if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "Bearer ") {


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/gateway_setup.go:
  Line 42: 		Auth:    clients.ContextAuthProvider{RequireTLS: useTLS},


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/main.go:
  Line 8: // Authorization: Bearer from API clients — and forwards the token to the

```

---

## Assistant (Build · GPT-5.6 Terra · 5.3s)

_Thinking:_



**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/factory.go",
  "offset": 150,
  "limit": 80
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/factory.go</path>
<type>file</type>
<content>
150: //
151: // ─── TOKEN FACTORY (USER TOKEN) ────────────────────────────────────────────────
152: // uses a user-provided Bearer token for client creation.
153: // each user has a separate client instance.
154: //
155: 
156: type TokenClientFactory struct {
157: 	Logger *slog.Logger
158: 	Header string
159: 	Prefix string
160: 	// NewTokenKubernetesClientFn is the function used to create token-based Kubernetes clients.
161: 	// This can be overridden by downstream code to provide custom client creation logic.
162: 	NewTokenKubernetesClientFn func(token string, logger *slog.Logger) (KubernetesClientInterface, error)
163: }
164: 
165: func NewTokenClientFactory(logger *slog.Logger, cfg config.EnvConfig) KubernetesClientFactory {
166: 	return &TokenClientFactory{
167: 		Logger:                     logger,
168: 		Header:                     cfg.AuthTokenHeader,
169: 		Prefix:                     cfg.AuthTokenPrefix,
170: 		NewTokenKubernetesClientFn: NewTokenKubernetesClient,
171: 	}
172: }
173: 
174: func (f *TokenClientFactory) ExtractRequestIdentity(httpHeader http.Header) (*RequestIdentity, error) {
175: 	raw := httpHeader.Get(f.Header)
176: 	if raw == "" {
177: 		return nil, fmt.Errorf("missing required Header: %s", f.Header)
178: 	}
179: 
180: 	token := raw
181: 	if f.Prefix != "" {
182: 		if !strings.HasPrefix(raw, f.Prefix) {
183: 			return nil, fmt.Errorf("expected token Header %s to start with Prefix %q", f.Header, f.Prefix)
184: 		}
185: 		token = strings.TrimPrefix(raw, f.Prefix)
186: 	}
187: 
188: 	return &RequestIdentity{
189: 		Token: strings.TrimSpace(token),
190: 	}, nil
191: }
192: 
193: func (f *TokenClientFactory) ValidateRequestIdentity(identity *RequestIdentity) error {
194: 
195: 	if identity == nil {
196: 		return errors.New("missing identity")
197: 	}
198: 
199: 	if identity.Token == "" {
200: 		return errors.New("token is required for token-based authentication")
201: 	}
202: 
203: 	return nil
204: }
205: 
206: func (f *TokenClientFactory) GetClient(ctx context.Context) (KubernetesClientInterface, error) {
207: 	identityVal := ctx.Value(constants.RequestIdentityKey)
208: 	if identityVal == nil {
209: 		return nil, fmt.Errorf("missing RequestIdentity in context")
210: 	}
211: 
212: 	identity, ok := identityVal.(*RequestIdentity)
213: 	if !ok || identity.Token == "" {
214: 		return nil, fmt.Errorf("invalid or missing identity token")
215: 	}
216: 
217: 	return f.NewTokenKubernetesClientFn(identity.Token, f.Logger)
218: }

(End of file - total 218 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/auth/proxy.go"
}
```

**Output:**
```
<path>/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/auth/proxy.go</path>
<type>file</type>
<content>
1: // Package auth provides the BFF's authentication middleware. The BFF never
2: // validates tokens itself — the gateway does that against its own OIDC JWKS.
3: // The middleware only decides where the bearer for a request comes from, in
4: // precedence order:
5: //
6: //  1. The auth-proxy header (oauth2-proxy / kube-auth-proxy injects
7: //     `x-forwarded-access-token`).
8: //  2. An explicit `Authorization: Bearer` header (API clients, tests).
9: //
10: // There is deliberately no third source: the BFF holds no sessions and runs
11: // no OIDC flows (ADR 0002). Deployments that need browser login put an auth
12: // proxy in front; the proxy authenticates WebSocket upgrades too, since it
13: // injects the token header on the upgrade request like any other.
14: package auth
15: 
16: import (
17: 	"context"
18: 	"fmt"
19: 	"net/http"
20: 	"strings"
21: )
22: 
23: type contextKey int
24: 
25: const (
26: 	tokenContextKey contextKey = iota
27: 	userContextKey
28: )
29: 
30: // Config holds auth middleware settings.
31: type Config struct {
32: 	TokenHeader string
33: 	UserHeader  string
34: 	Disabled    bool
35: }
36: 
37: // Middleware extracts the request's bearer token and stores it on the
38: // request context for the gateway client to forward.
39: type Middleware struct {
40: 	cfg Config
41: }
42: 
43: // New builds the middleware.
44: func New(cfg Config) *Middleware {
45: 	if cfg.TokenHeader == "" {
46: 		cfg.TokenHeader = "x-forwarded-access-token"
47: 	}
48: 	if cfg.UserHeader == "" {
49: 		cfg.UserHeader = "x-auth-request-user"
50: 	}
51: 	return &Middleware{cfg: cfg}
52: }
53: 
54: // Disabled reports whether auth validation is turned off.
55: func (m *Middleware) Disabled() bool {
56: 	return m.cfg.Disabled
57: }
58: 
59: // TokenHeader returns the configured header name for the bearer token.
60: func (m *Middleware) TokenHeader() string {
61: 	return m.cfg.TokenHeader
62: }
63: 
64: // Handler resolves the request's bearer token and stores it on the request
65: // context. When auth is disabled, a synthetic dev-user identity is injected
66: // and any tokens on the request are ignored, so a misconfigured proxy cannot
67: // leak credentials to the gateway in dev mode.
68: func (m *Middleware) Handler(next http.Handler) http.Handler {
69: 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
70: 		if m.cfg.Disabled {
71: 			ctx := context.WithValue(r.Context(), userContextKey, "dev-user")
72: 			next.ServeHTTP(w, r.WithContext(ctx))
73: 			return
74: 		}
75: 
76: 		token := r.Header.Get(m.cfg.TokenHeader)
77: 		if token == "" {
78: 			if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "Bearer ") {
79: 				token = strings.TrimPrefix(bearer, "Bearer ")
80: 			}
81: 		}
82: 		if token == "" {
83: 			writeUnauthorized(w, "not authenticated")
84: 			return
85: 		}
86: 
87: 		ctx := context.WithValue(r.Context(), tokenContextKey, token)
88: 		if user := r.Header.Get(m.cfg.UserHeader); user != "" {
89: 			ctx = context.WithValue(ctx, userContextKey, user)
90: 		}
91: 		next.ServeHTTP(w, r.WithContext(ctx))
92: 	})
93: }
94: 
95: func writeUnauthorized(w http.ResponseWriter, message string) {
96: 	w.Header().Set("Content-Type", "application/json")
97: 	w.WriteHeader(http.StatusUnauthorized)
98: 	fmt.Fprintf(w, `{"code":"unauthorized","message":%q}`, message)
99: }
100: 
101: // TokenFromContext returns the bearer token stored by the middleware, or ""
102: // when absent.
103: func TokenFromContext(ctx context.Context) string {
104: 	token, _ := ctx.Value(tokenContextKey).(string)
105: 	return token
106: }
107: 
108: // WithToken returns a context carrying the given bearer token. Useful in tests.
109: func WithToken(ctx context.Context, token string) context.Context {
110: 	return context.WithValue(ctx, tokenContextKey, token)
111: }
112: 
113: // UserFromContext returns the authenticated username, or "" when absent.
114: func UserFromContext(ctx context.Context) string {
115: 	user, _ := ctx.Value(userContextKey).(string)
116: 	return user
117: }

(End of file - total 117 lines)
</content>
```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/main.go",
  "offset": 1,
  "limit": 150
}
```

**Output:**
```
<path>/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/main.go</path>
<type>file</type>
<content>
1: // The OpenShell Dashboard BFF: serves the REST API for the React frontend and
2: // (optionally) the built static assets, talking to the OpenShell gateway over
3: // gRPC with per-request bearer token forwarding.
4: //
5: // The BFF is a token relay (ADR 0002): authentication is owned by an external
6: // auth proxy (oauth2-proxy, kube-auth-proxy, ...) which injects the user's
7: // bearer token as an HTTP header. The BFF reads that header — or an explicit
8: // Authorization: Bearer from API clients — and forwards the token to the
9: // gateway, which validates it against its own OIDC JWKS. The BFF never runs
10: // OIDC flows, never holds sessions, and never validates tokens.
11: package main
12: 
13: import (
14: 	"context"
15: 	"flag"
16: 	"log/slog"
17: 	"net"
18: 	"net/http"
19: 	"os"
20: 	"os/signal"
21: 	"syscall"
22: 
23: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/auth"
24: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/models"
25: 	"github.com/Gkrumbach07/openshell-dashboard/backend/pkg/server"
26: )
27: 
28: const (
29: 	defaultPort       = "8080"
30: 	defaultGatewayURL = "localhost:50051"
31: )
32: 
33: func envOr(key, fallback string) string {
34: 	if value := os.Getenv(key); value != "" {
35: 		return value
36: 	}
37: 	return fallback
38: }
39: 
40: func main() {
41: 	var (
42: 		port              = flag.String("port", envOr("PORT", defaultPort), "listen port (env PORT)")
43: 		listenAddress     = flag.String("listen-address", envOr("LISTEN_ADDRESS", ""), "listen address (env LISTEN_ADDRESS)")
44: 		gatewayURL        = flag.String("gateway-url", envOr("OPENSHELL_GATEWAY_URL", defaultGatewayURL), "OpenShell gateway gRPC endpoint (env OPENSHELL_GATEWAY_URL)")
45: 		gatewayCACert     = flag.String("gateway-ca-cert", envOr("GATEWAY_CA_CERT", ""), "path to CA cert for gateway TLS (env GATEWAY_CA_CERT)")
46: 		gatewayClientCert = flag.String("gateway-client-cert", envOr("GATEWAY_CLIENT_CERT", ""), "path to client certificate for gateway mTLS (env GATEWAY_CLIENT_CERT)")
47: 		gatewayClientKey  = flag.String("gateway-client-key", envOr("GATEWAY_CLIENT_KEY", ""), "path to client key for gateway mTLS (env GATEWAY_CLIENT_KEY)")
48: 		tlsCert           = flag.String("tls-cert", envOr("TLS_CERT_FILE", ""), "path to server cert for inbound HTTPS (env TLS_CERT_FILE)")
49: 		tlsKey            = flag.String("tls-key", envOr("TLS_KEY_FILE", ""), "path to server key for inbound HTTPS (env TLS_KEY_FILE)")
50: 		staticDir         = flag.String("static-dir", envOr("STATIC_DIR", ""), "frontend static assets directory (env STATIC_DIR)")
51: 		authDisabled      = flag.Bool("auth-disabled", envOr("AUTH_DISABLED", "false") == "true", "skip auth — dev only (env AUTH_DISABLED)")
52: 		tokenHeader       = flag.String("auth-token-header", envOr("AUTH_TOKEN_HEADER", "x-forwarded-access-token"), "header injected by auth proxy containing the bearer token (env AUTH_TOKEN_HEADER)")
53: 		userHeader        = flag.String("auth-user-header", envOr("AUTH_USER_HEADER", "x-auth-request-user"), "header injected by auth proxy containing the username (env AUTH_USER_HEADER)")
54: 		adminRole         = flag.String("admin-role", envOr("ADMIN_ROLE", "admin"), "role name that grants platform admin access (env ADMIN_ROLE)")
55: 		logoutURL         = flag.String("logout-url", envOr("LOGOUT_URL", "/oauth2/sign_out"), "auth proxy sign-out URL to redirect to on logout (env LOGOUT_URL)")
56: 	)
57: 	flag.Parse()
58: 
59: 	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
60: 	slog.SetDefault(logger)
61: 
62: 	warnGatewayConfig(*gatewayURL, *gatewayCACert, *authDisabled)
63: 	if err := validateInboundTLS(*tlsCert, *tlsKey); err != nil {
64: 		slog.Error(err.Error())
65: 		os.Exit(1)
66: 	}
67: 
68: 	authMiddleware := auth.New(auth.Config{
69: 		Disabled:    *authDisabled,
70: 		TokenHeader: *tokenHeader,
71: 		UserHeader:  *userHeader,
72: 	})
73: 
74: 	authCfg := models.AuthConfigResponse{
75: 		AuthDisabled: *authDisabled,
76: 		AdminRole:    *adminRole,
77: 		LogoutURL:    *logoutURL,
78: 		Features: models.FeatureFlags{
79: 			Terminal:          envOr("FEATURE_TERMINAL", "true") == "true",
80: 			FileTransfer:      envOr("FEATURE_FILE_TRANSFER", "true") == "true",
81: 			Settings:          envOr("FEATURE_SETTINGS", "true") == "true",
82: 			GlobalPolicy:      envOr("FEATURE_GLOBAL_POLICY", "true") == "true",
83: 			CredentialRefresh: envOr("FEATURE_CREDENTIAL_REFRESH", "true") == "true",
84: 			Services:          envOr("FEATURE_SERVICES", "true") == "true",
85: 			DraftPolicy:       envOr("FEATURE_DRAFT_POLICY", "true") == "true",
86: 		},
87: 	}
88: 
89: 	clients, err := newGatewayClients(*gatewayURL, *gatewayCACert, *gatewayClientCert, *gatewayClientKey)
90: 	if err != nil {
91: 		exitOnError("gateway client setup failed", err)
92: 	}
93: 	defer clients.Close()
94: 
95: 	app := server.NewApp(clients.sdk, clients.uploadExec, authMiddleware, *staticDir, authCfg)
96: 
97: 	addr := net.JoinHostPort(*listenAddress, *port)
98: 	slog.Info("openshell-dashboard BFF listening",
99: 		"addr", addr,
100: 		"scheme", inboundScheme(*tlsCert, *tlsKey),
101: 		"gateway", *gatewayURL,
102: 		"static", *staticDir,
103: 		"authDisabled", *authDisabled,
104: 	)
105: 
106: 	server := newInboundServer(addr, app.Routes())
107: 	errCh := make(chan error, 1)
108: 	go func() {
109: 		errCh <- serveInbound(server, *tlsCert, *tlsKey)
110: 	}()
111: 
112: 	sigCh := make(chan os.Signal, 1)
113: 	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
114: 
115: 	select {
116: 	case err := <-errCh:
117: 		if err != nil {
118: 			exitOnError("server exited", err)
119: 		}
120: 		return
121: 	case sig := <-sigCh:
122: 		signal.Stop(sigCh)
123: 		slog.Info("shutting down BFF", "signal", sig.String())
124: 	}
125: 
126: 	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
127: 	defer cancel()
128: 	if err := server.Shutdown(shutdownCtx); err != nil {
129: 		exitOnError("server shutdown failed", err)
130: 	}
131: 
132: 	if err := <-errCh; err != nil && err != http.ErrServerClosed {
133: 		exitOnError("server exited", err)
134: 	}
135: }

(End of file - total 135 lines)
</content>
```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c",
  "pattern": "TokenFromContext|WithToken|Auth\\.Handler|auth\\.New",
  "include": "*.go"
}
```

**Output:**
```
Found 16 matches
/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/handlers/gateway_handler_test.go:
  Line 28: 	handler := NewGatewayHandler(services.NewGatewayService(sdk), auth.New(auth.Config{}), models.AuthConfigResponse{})

  Line 53: 	handler := NewGatewayHandler(services.NewGatewayService(sdk), auth.New(auth.Config{}), models.AuthConfigResponse{})

  Line 63: 	handler := NewGatewayHandler(services.NewGatewayService(&mockSDK{}), auth.New(auth.Config{}), models.AuthConfigResponse{})

  Line 77: 	handler := NewGatewayHandler(services.NewGatewayService(sdk), auth.New(auth.Config{}), models.AuthConfigResponse{})

  Line 89: 		auth.New(auth.Config{Disabled: true}),


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/auth.go:
  Line 17: 	token := auth.TokenFromContext(ctx)


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/auth/proxy_test.go:
  Line 17: 		gotToken = TokenFromContext(r.Context())

  Line 61: 		gotToken = TokenFromContext(r.Context())

  Line 80: 		gotToken = TokenFromContext(r.Context())

  Line 118: 		gotToken = TokenFromContext(r.Context())


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/auth/proxy.go:
  Line 101: // TokenFromContext returns the bearer token stored by the middleware, or ""

  Line 103: func TokenFromContext(ctx context.Context) string {

  Line 108: // WithToken returns a context carrying the given bearer token. Useful in tests.

  Line 109: func WithToken(ctx context.Context, token string) context.Context {


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/clients/sdkclient/auth.go:
  Line 17: 	token := auth.TokenFromContext(ctx)


/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/cmd/server/main.go:
  Line 68: 	authMiddleware := auth.New(auth.Config{

```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/constants/keys.go"
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/constants/keys.go</path>
<type>file</type>
<content>
1: package constants
2: 
3: type contextKey string
4: 
5: // NOTE: If you are adding any HTTP headers, they need to also be added to the EnableCORS middleware
6: // to ensure requests are not blocked when using CORS.
7: const (
8: 	NamespaceHeaderParameterKey contextKey = "namespace"
9: 
10: 	// The following keys are used to store the user access token in the context
11: 	RequestIdentityKey contextKey = "requestIdentityKey"
12: 
13: 	// For config.AuthMethodInternal
14: 	// Kubeflow authorization operates using custom authentication headers:
15: 	// Note: The functionality for `kubeflow-groups` is not fully operational at Kubeflow platform at this time
16: 	// but it's supported on Mod Arch BFF
17: 	KubeflowUserIDHeader       = "kubeflow-userid" // kubeflow-userid :contains the user's email address
18: 	KubeflowUserGroupsIdHeader = "kubeflow-groups" // kubeflow-groups : Holds a comma-separated list of user groups
19: 
20: 	TraceIdKey     contextKey = "TraceIdKey"
21: 	TraceLoggerKey contextKey = "TraceLoggerKey"
22: )
23: 
24: // BFFTarget represents a target BFF service (re-exported from bffclient package)
25: type BFFTarget string
26: 
27: // BFFClientKey returns a context key for storing BFF clients by target.
28: // This allows multiple BFF clients to be stored in the same context.
29: // Usage: ctx.Value(constants.BFFClientKey("maas"))
30: func BFFClientKey(target BFFTarget) contextKey {
31: 	return contextKey("BFFClientKey_" + string(target))
32: }

(End of file - total 32 lines)
</content>
```

---

## Assistant (Build · GPT-5.6 Terra · 8.8s)

_Thinking:_



_Thinking:_



**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/server",
  "pattern": "func \\(.*\\) Routes|auth\\.Handler|type App|func NewApp",
  "include": "*.go"
}
```

**Output:**
```
Found 4 matches
/Users/derxu/go/pkg/mod/github.com/d0w/openshell-dashboard/backend@v0.0.0-20260921141023-113961fac55c/pkg/server/app.go:
  Line 23: type App struct { //nolint:govet // fieldalignment: readability over padding

  Line 47: func NewApp(sdkClient openshell.ClientInterface, execUpload services.StdinExecer, authMiddleware *auth.Middleware, staticDir string, authCfg models.AuthConfigResponse) *App {

  Line 78: func (app *App) Routes() http.Handler {

  Line 92: 			r.Use(app.auth.Handler)

```

**Tool: read**

**Input:**
```json
{
  "filePath": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/factory.go",
  "offset": 1,
  "limit": 145
}
```

**Output:**
```
<path>/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/kubernetes/factory.go</path>
<type>file</type>
<content>
1: package kubernetes
2: 
3: import (
4: 	"context"
5: 	"errors"
6: 	"fmt"
7: 	"log/slog"
8: 	"net/http"
9: 	"strings"
10: 
11: 	"github.com/opendatahub-io/agent-ops/internal/config"
12: 	"github.com/opendatahub-io/agent-ops/internal/constants"
13: 	"k8s.io/client-go/kubernetes"
14: 	"k8s.io/client-go/rest"
15: )
16: 
17: func NewKubernetesClientFactory(cfg config.EnvConfig, logger *slog.Logger) (KubernetesClientFactory, error) {
18: 	switch cfg.AuthMethod {
19: 
20: 	case config.AuthMethodInternal:
21: 		k8sFactory, err := NewStaticClientFactory(logger)
22: 		if err != nil {
23: 			return nil, fmt.Errorf("failed to create static client factory: %w", err)
24: 		}
25: 		return k8sFactory, nil
26: 
27: 	case config.AuthMethodUser:
28: 		k8sFactory := NewTokenClientFactory(logger, cfg)
29: 		return k8sFactory, nil
30: 
31: 	case config.AuthMethodDisabled:
32: 		return nil, nil
33: 
34: 	default:
35: 		return nil, fmt.Errorf("invalid auth method: %q", cfg.AuthMethod)
36: 	}
37: }
38: 
39: // ─── STATIC FACTORY (INTERNAL) ──────────────────────────────────────────
40: // uses the credentials of the running backend to create a single instance of the client
41: // If running inside the cluster, it uses the pod's service account.
42: // If running locally (e.g. for development), it uses the current user's kubeconfig context.
43: type KubernetesClientFactory interface {
44: 	GetClient(ctx context.Context) (KubernetesClientInterface, error)
45: 	ExtractRequestIdentity(httpHeader http.Header) (*RequestIdentity, error)
46: 	ValidateRequestIdentity(identity *RequestIdentity) error
47: }
48: 
49: type StaticClientFactory struct {
50: 	Logger     *slog.Logger
51: 	sarClient  KubernetesClientInterface
52: 	restConfig *rest.Config
53: }
54: 
55: func NewStaticClientFactory(logger *slog.Logger) (KubernetesClientFactory, error) {
56: 	client, err := newInternalKubernetesClient(logger)
57: 	if err != nil {
58: 		return nil, fmt.Errorf("failed to create service account client: %w", err)
59: 	}
60: 	internal, ok := client.(*InternalKubernetesClient)
61: 	if !ok || internal.restConfig == nil {
62: 		return nil, fmt.Errorf("internal kubernetes client is missing rest config")
63: 	}
64: 	return &StaticClientFactory{
65: 		sarClient:  client,
66: 		restConfig: rest.CopyConfig(internal.restConfig),
67: 		Logger:     logger,
68: 	}, nil
69: }
70: 
71: func (f *StaticClientFactory) GetClient(ctx context.Context) (KubernetesClientInterface, error) {
72: 	val := ctx.Value(constants.RequestIdentityKey)
73: 	if val == nil {
74: 		return nil, fmt.Errorf("missing RequestIdentity in context")
75: 	}
76: 
77: 	identity, ok := val.(*RequestIdentity)
78: 	if !ok || identity == nil || identity.UserID == "" {
79: 		return nil, fmt.Errorf("invalid or missing request identity")
80: 	}
81: 
82: 	impersonatedCfg := rest.CopyConfig(f.restConfig)
83: 	impersonatedCfg.Impersonate = rest.ImpersonationConfig{
84: 		UserName: identity.UserID,
85: 		Groups:   identity.Groups,
86: 	}
87: 	clearClientCertificateFields(impersonatedCfg)
88: 
89: 	clientset, err := kubernetes.NewForConfig(impersonatedCfg)
90: 	if err != nil {
91: 		return nil, fmt.Errorf("failed to create impersonated kubernetes client: %w", err)
92: 	}
93: 
94: 	return &InternalKubernetesClient{
95: 		SharedClientLogic: SharedClientLogic{
96: 			Client: clientset,
97: 			Logger: f.Logger,
98: 		},
99: 		restConfig:   impersonatedCfg,
100: 		sarClientset: f.sarClient.KubernetesClientset(),
101: 	}, nil
102: }
103: 
104: func (f *StaticClientFactory) ExtractRequestIdentity(httpHeader http.Header) (*RequestIdentity, error) {
105: 
106: 	userID := httpHeader.Get(constants.KubeflowUserIDHeader)
107: 	//`kubeflow-userid`: Contains the user's email address.
108: 	if userID == "" {
109: 		return nil, errors.New("missing required kubeflow-userid header")
110: 	}
111: 
112: 	userGroupsHeader := httpHeader.Get(constants.KubeflowUserGroupsIdHeader)
113: 	// Note: The functionality for `kubeflow-groups` is not fully operational at Kubeflow platform at this time
114: 	// but it's supported on Mod Arch BFF
115: 	//`kubeflow-groups`: Holds a comma-separated list of user groups.
116: 	groups := []string{}
117: 	if userGroupsHeader != "" {
118: 		for _, g := range strings.Split(userGroupsHeader, ",") {
119: 			groups = append(groups, strings.TrimSpace(g))
120: 		}
121: 	}
122: 	identity := &RequestIdentity{
123: 		UserID: userID,
124: 		Groups: groups,
125: 	}
126: 	return identity, nil
127: }
128: 
129: func (f *StaticClientFactory) ValidateRequestIdentity(identity *RequestIdentity) error {
130: 	if identity == nil {
131: 		return errors.New("missing identity")
132: 	}
133: 	if identity.UserID == "" {
134: 		return errors.New("user ID (kubeflow-userid) required for internal authentication")
135: 	}
136: 	return nil
137: }
138: 
139: // clearClientCertificateFields prevents client-cert credentials from overriding impersonation.
140: func clearClientCertificateFields(cfg *rest.Config) {
141: 	if cfg == nil {
142: 		return
143: 	}
144: 	cfg.CertData = nil
145: 	cfg.CertFile = ""

(Showing lines 1-145 of 218. Use offset=146 to continue.)
</content>
```

**Tool: grep**

**Input:**
```json
{
  "path": "/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff",
  "pattern": "InjectRequestIdentity|Routes\\(\\)|NewApp\\(",
  "include": "*.go"
}
```

**Output:**
```
Found 23 matches
/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/middleware.go:
  Line 34: func (app *App) InjectRequestIdentity(next http.Handler) http.Handler {

  Line 173: // This middleware must be placed after InjectRequestIdentity and a namespace-attachment middleware.


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/testapp.go:
  Line 15: // WithOpenAPIHandler sets the OpenAPI handler. Required when calling Routes() in tests.

  Line 31: // Tests that call Routes() must pass WithOpenAPIHandler(NewOpenAPIHandler(...)).


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/middleware_test.go:
  Line 263: func TestInjectRequestIdentity_SkipsAuthWhenDisabled(t *testing.T) {

  Line 272: 	handler := app.InjectRequestIdentity(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

  Line 283: func TestInjectRequestIdentity_NilFactoryPanicsWithoutDisabledAuth(t *testing.T) {

  Line 292: 		app.InjectRequestIdentity(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

  Line 298: func TestInjectRequestIdentity_SkipsNonAPIPaths(t *testing.T) {

  Line 307: 	handler := app.InjectRequestIdentity(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/openapi_handler_test.go:
  Line 52: 	handler := app.Routes()


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/app_test.go:
  Line 26: 	app.Routes().ServeHTTP(rr, req)

  Line 35: 	app.Routes().ServeHTTP(rr, req)

  Line 40: 	_, err := NewApp(config.EnvConfig{

  Line 58: 	app, err := NewApp(config.EnvConfig{

  Line 68: func TestInjectRequestIdentity_UnauthorizedWithoutToken(t *testing.T) {

  Line 80: 	handler := app.InjectRequestIdentity(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/api/app.go:
  Line 83: func NewApp(cfg config.EnvConfig, logger *slog.Logger) (*App, error) {

  Line 216: func (app *App) Routes() http.Handler {

  Line 305: 	combinedMux.Handle("/", app.RecoverPanic(app.EnableTelemetry(app.EnableCORS(app.InjectRequestIdentity(appMux)))))


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/internal/integrations/bffclient/middleware.go:
  Line 29: //	    app.InjectRequestIdentity(


/Users/derxu/Documents/code/odh-dashboard/packages/agent-ops/bff/cmd/gateway_setup.go:
  Line 71: 	app := openshellapi.NewApp(

  Line 89: 		Handler: app.Routes(),

```

---

## Assistant (Build · GPT-5.6 Terra · 35.2s)

_Thinking:_



_Thinking:_



_Thinking:_



**Production Shape**

Use OIDC bearer-token relay plus server-authenticated TLS.

```text
dashboard auth proxy
  -> BFF: verified user JWT
  -> OpenShell: same user JWT over TLS
  -> gateway validates JWT against OIDC JWKS
```

BFF mounts gateway CA public cert only. No OpenShell server key, CA key, JWT-signing Secret, or shared `openshell-client-tls` private key.

If gateway deployment enforces TLS client certificates, provision dedicated BFF cert/key. Never reuse supervisor/client bundle. This should be fallback, not default identity mechanism.

**OpenShell Configuration**

Configure gateway:

```yaml
certManager:
  enabled: true

server:
  oidc:
    issuer: https://<issuer>
    audience: <audience>
```

BFF uses in-cluster endpoint:

```text
https://openshell.<openshell-namespace>.svc.cluster.local:8080
```

Expose gateway only inside cluster. Mesh `STRICT` mTLS may wrap this connection; OpenShell TLS remains application TLS inside mesh transport.

**Required Code Changes**

1. Replace hardcoded gateway discovery endpoint in `pkg/discovery/discovery.go`.
   - Read `OPENSHELL_GATEWAY_URL`.
   - Production default must not be `localhost`.

2. Add explicit gateway TLS config.
   - `OPENSHELL_GATEWAY_CA_FILE`
   - Optional `OPENSHELL_GATEWAY_CLIENT_CERT_FILE`
   - Optional `OPENSHELL_GATEWAY_CLIENT_KEY_FILE`
   - Fail startup if HTTPS URL lacks CA file, or only one client-cert/key path exists.
   - Do not reuse unrelated `BUNDLE_PATHS` or `INSECURE_SKIP_VERIFY`.

3. Pass CA path into `NewGatewayClients`.
   - Current `cmd/gateway_setup.go:62-67` passes four empty strings.
   - CA-only TLS should be normal production configuration.

4. Replace `openshell.NoAuth()` in `internal/clients/gateway.go:57`.
   - Use upstream `clients.ContextAuthProvider{RequireTLS: useTLS}`.
   - This forwards bearer token per gRPC call.

5. Keep upstream OpenShell dashboard auth middleware enabled.
   - `openshellauth.New(openshellauth.Config{})` already extracts `x-forwarded-access-token`, falling back to `Authorization`.
   - Ensure only trusted ingress/auth proxy can set these headers.

6. Require real user JWT inbound.
   - `AUTH_METHOD=internal` provides `kubeflow-userid` headers, not an OIDC bearer JWT.
   - Use `AUTH_METHOD=user_token` when dashboard passes a user token.
   - If platform cannot provide matching OIDC JWT, implement trusted OAuth token exchange. Never synthesize gateway identity from `kubeflow-userid`.

**CA Delivery**

Mount only `ca.crt` into BFF. Prefer separate ConfigMap/public trust bundle. Mounting a single Secret key is acceptable when necessary:

```yaml
volumes:
  - name: openshell-ca
    secret:
      secretName: openshell-server-tls
      items:
        - key: ca.crt
          path: ca.crt
```

BFF ServiceAccount does not need Secret-read RBAC for mounted volumes.

**mTLS Fallback**

Only if OpenShell gateway requires client certificate at TLS layer:

- cert-manager issues unique `agent-ops-bff` client certificate from gateway-trusted issuer.
- Mount only its `tls.crt` and `tls.key` into BFF.
- Keep OIDC bearer forwarding. Client cert authenticates workload transport, JWT identifies user.
- Rotate BFF client on Secret renewal. Current SDK client loads PEM once, so restart/recreate gateway clients on rotation.

**Tests**

- Unit: reject missing CA, invalid endpoint scheme, orphan cert/key.
- Unit: request JWT reaches gRPC `authorization` metadata.
- Integration: valid JWT succeeds; missing/invalid JWT gets denied; BFF cannot call gateway with only `kubeflow-userid`.
- Integration: CA trust failure fails closed; no `INSECURE_SKIP_VERIFY` production path.

---

## User

is the ca.crt enough for openshell? I was getting errors with ERROR openshell_server: TLS handshake failed error=received fatal alert: BadCertificate client=127.0.0.1:45998
2026-09-22T22:16:50.215725Z ERROR openshell_server: TLS handshake failed error=received fatal alert: BadCertificate client=127.0.0.1:58332
2026-09-22T22:16:51.285550Z ERROR openshell_server: TLS handshake failed error=received fatal alert: UnknownCA client=127.0.0.1:58350 

---

## Assistant (Build · GPT-5.6 Terra · 18.4s)

_Thinking:_



_Thinking:_

**Considering webfetch options**

I'm thinking about whether I should use webfetch to access GitHub directly. The user is seeking a straightforward answer, so maybe I should explore browsing options. I wonder if inspecting the values from a public GitHub repository would be helpful here. There could also be relevant documentation from Nvidia about LLMs that might include useful links. I’ll figure out if using the GitHub API is the best way to proceed.

`UnknownCA` and `BadCertificate` came from client toward gateway: client rejects gateway certificate. Verify trust/SAN first; then test whether gateway also mandates client mTLS.

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/helm-chart/values.yaml",
  "format": "text"
}
```

**Error:**
```
StatusCode: non 2xx status code (404 GET https://raw.githubusercontent.com/NVIDIA/OpenShell/main/helm-chart/values.yaml)
```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents",
  "format": "text"
}
```

**Output:**
```
[
  {
    "name": ".agents",
    "path": ".agents",
    "sha": "d1cff481982f5afceaad7a0f85046e632ee48867",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.agents?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/.agents",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/d1cff481982f5afceaad7a0f85046e632ee48867",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.agents?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/d1cff481982f5afceaad7a0f85046e632ee48867",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/.agents"
    }
  },
  {
    "name": ".cargo",
    "path": ".cargo",
    "sha": "ae0517287fb2b28e109a37ce219628c82fa72cd1",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.cargo?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/.cargo",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/ae0517287fb2b28e109a37ce219628c82fa72cd1",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.cargo?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/ae0517287fb2b28e109a37ce219628c82fa72cd1",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/.cargo"
    }
  },
  {
    "name": ".claude",
    "path": ".claude",
    "sha": "94804dd1318d6bf0039fc67432d09c6dae4fe886",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.claude?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/.claude",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/94804dd1318d6bf0039fc67432d09c6dae4fe886",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.claude?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/94804dd1318d6bf0039fc67432d09c6dae4fe886",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/.claude"
    }
  },
  {
    "name": ".config",
    "path": ".config",
    "sha": "5587f2068758ef8cff4ed6c88612f467dd762466",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.config?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/.config",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/5587f2068758ef8cff4ed6c88612f467dd762466",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.config?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/5587f2068758ef8cff4ed6c88612f467dd762466",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/.config"
    }
  },
  {
    "name": ".dockerignore",
    "path": ".dockerignore",
    "sha": "192b57b6cefc47e15a776be6d6c0618ce6fcb054",
    "size": 341,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.dockerignore?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/.dockerignore",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/192b57b6cefc47e15a776be6d6c0618ce6fcb054",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/.dockerignore",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.dockerignore?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/192b57b6cefc47e15a776be6d6c0618ce6fcb054",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/.dockerignore"
    }
  },
  {
    "name": ".env.example",
    "path": ".env.example",
    "sha": "d043379e75158f866e768025b06e938f159b781f",
    "size": 684,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.env.example?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/.env.example",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d043379e75158f866e768025b06e938f159b781f",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/.env.example",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.env.example?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d043379e75158f866e768025b06e938f159b781f",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/.env.example"
    }
  },
  {
    "name": ".gitattributes",
    "path": ".gitattributes",
    "sha": "af03ae26af336eee26a9f25f042dc5a4f9102c3b",
    "size": 543,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.gitattributes?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/.gitattributes",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/af03ae26af336eee26a9f25f042dc5a4f9102c3b",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/.gitattributes",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.gitattributes?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/af03ae26af336eee26a9f25f042dc5a4f9102c3b",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/.gitattributes"
    }
  },
  {
    "name": ".github",
    "path": ".github",
    "sha": "cd838ac39d310345d7e9d370d23ba385ab9d248a",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.github?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/.github",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/cd838ac39d310345d7e9d370d23ba385ab9d248a",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.github?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/cd838ac39d310345d7e9d370d23ba385ab9d248a",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/.github"
    }
  },
  {
    "name": ".gitignore",
    "path": ".gitignore",
    "sha": "b4e8cd16317f1e6e09130ada29310f82cb6e41a2",
    "size": 3885,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.gitignore?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/.gitignore",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b4e8cd16317f1e6e09130ada29310f82cb6e41a2",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/.gitignore",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.gitignore?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b4e8cd16317f1e6e09130ada29310f82cb6e41a2",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/.gitignore"
    }
  },
  {
    "name": ".markdownlint-cli2.jsonc",
    "path": ".markdownlint-cli2.jsonc",
    "sha": "125df0f812ebadb206338e562d2e763626aadc3c",
    "size": 1521,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.markdownlint-cli2.jsonc?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/.markdownlint-cli2.jsonc",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/125df0f812ebadb206338e562d2e763626aadc3c",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/.markdownlint-cli2.jsonc",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.markdownlint-cli2.jsonc?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/125df0f812ebadb206338e562d2e763626aadc3c",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/.markdownlint-cli2.jsonc"
    }
  },
  {
    "name": ".opencode",
    "path": ".opencode",
    "sha": "76a26f4179e0f43c0cc50a64bc6e5d5111d30cdc",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.opencode?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/.opencode",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/76a26f4179e0f43c0cc50a64bc6e5d5111d30cdc",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.opencode?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/76a26f4179e0f43c0cc50a64bc6e5d5111d30cdc",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/.opencode"
    }
  },
  {
    "name": ".packit.yaml",
    "path": ".packit.yaml",
    "sha": "d3b92eafae007f5b7a1bb70ab2c3ade0c1dc19a8",
    "size": 3910,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.packit.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/.packit.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d3b92eafae007f5b7a1bb70ab2c3ade0c1dc19a8",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/.packit.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.packit.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d3b92eafae007f5b7a1bb70ab2c3ade0c1dc19a8",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/.packit.yaml"
    }
  },
  {
    "name": ".python-version",
    "path": ".python-version",
    "sha": "a6d9ada03ee63903081d538cc0a139a1b9658908",
    "size": 7,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.python-version?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/.python-version",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a6d9ada03ee63903081d538cc0a139a1b9658908",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/.python-version",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.python-version?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a6d9ada03ee63903081d538cc0a139a1b9658908",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/.python-version"
    }
  },
  {
    "name": ".trivyignore.yaml",
    "path": ".trivyignore.yaml",
    "sha": "4cc4e32ffe9f4647c0d32b9c8b8578edc020c1af",
    "size": 1713,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.trivyignore.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/.trivyignore.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4cc4e32ffe9f4647c0d32b9c8b8578edc020c1af",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/.trivyignore.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/.trivyignore.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4cc4e32ffe9f4647c0d32b9c8b8578edc020c1af",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/.trivyignore.yaml"
    }
  },
  {
    "name": "AGENTS.md",
    "path": "AGENTS.md",
    "sha": "b1a0ffeb60e4465206fe508a2890593e92f96149",
    "size": 22662,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/AGENTS.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/AGENTS.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b1a0ffeb60e4465206fe508a2890593e92f96149",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/AGENTS.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/AGENTS.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b1a0ffeb60e4465206fe508a2890593e92f96149",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/AGENTS.md"
    }
  },
  {
    "name": "CI.md",
    "path": "CI.md",
    "sha": "35802a42dca09ff8857e4cdd03b81e0fac033156",
    "size": 27568,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/CI.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/CI.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/35802a42dca09ff8857e4cdd03b81e0fac033156",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/CI.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/CI.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/35802a42dca09ff8857e4cdd03b81e0fac033156",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/CI.md"
    }
  },
  {
    "name": "CLAUDE.md",
    "path": "CLAUDE.md",
    "sha": "43c994c2d3617f947bcb5adf1933e21dabe46bb5",
    "size": 11,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/CLAUDE.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/CLAUDE.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/43c994c2d3617f947bcb5adf1933e21dabe46bb5",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/CLAUDE.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/CLAUDE.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/43c994c2d3617f947bcb5adf1933e21dabe46bb5",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/CLAUDE.md"
    }
  },
  {
    "name": "CODE_OF_CONDUCT.md",
    "path": "CODE_OF_CONDUCT.md",
    "sha": "4575d16002f1fa93fd385608c44462679987cb7c",
    "size": 335,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/CODE_OF_CONDUCT.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/CODE_OF_CONDUCT.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4575d16002f1fa93fd385608c44462679987cb7c",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/CODE_OF_CONDUCT.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/CODE_OF_CONDUCT.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4575d16002f1fa93fd385608c44462679987cb7c",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/CODE_OF_CONDUCT.md"
    }
  },
  {
    "name": "CONTRIBUTING.md",
    "path": "CONTRIBUTING.md",
    "sha": "4213239b35a729e39c56d26f4aa0b6b66011f7ad",
    "size": 35156,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/CONTRIBUTING.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/CONTRIBUTING.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4213239b35a729e39c56d26f4aa0b6b66011f7ad",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/CONTRIBUTING.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/CONTRIBUTING.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4213239b35a729e39c56d26f4aa0b6b66011f7ad",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/CONTRIBUTING.md"
    }
  },
  {
    "name": "Cargo.lock",
    "path": "Cargo.lock",
    "sha": "53a3654ebfb186e913c38960e53208097004bbfb",
    "size": 214762,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/Cargo.lock?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/Cargo.lock",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/53a3654ebfb186e913c38960e53208097004bbfb",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/Cargo.lock",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/Cargo.lock?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/53a3654ebfb186e913c38960e53208097004bbfb",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/Cargo.lock"
    }
  },
  {
    "name": "Cargo.toml",
    "path": "Cargo.toml",
    "sha": "2e7237190f3f0c6fc11b6b6a4f9c5d14eaac6a5a",
    "size": 6121,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/Cargo.toml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/Cargo.toml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2e7237190f3f0c6fc11b6b6a4f9c5d14eaac6a5a",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/Cargo.toml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/Cargo.toml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2e7237190f3f0c6fc11b6b6a4f9c5d14eaac6a5a",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/Cargo.toml"
    }
  },
  {
    "name": "DCO",
    "path": "DCO",
    "sha": "49b8cb0549267a8176467738b172a63d86eff436",
    "size": 1366,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/DCO?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/DCO",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/49b8cb0549267a8176467738b172a63d86eff436",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/DCO",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/DCO?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/49b8cb0549267a8176467738b172a63d86eff436",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/DCO"
    }
  },
  {
    "name": "GOVERNANCE.md",
    "path": "GOVERNANCE.md",
    "sha": "e2c7cfbbef31331f76e2a596b6dbd052736b4834",
    "size": 13590,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/GOVERNANCE.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/GOVERNANCE.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/e2c7cfbbef31331f76e2a596b6dbd052736b4834",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/GOVERNANCE.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/GOVERNANCE.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/e2c7cfbbef31331f76e2a596b6dbd052736b4834",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/GOVERNANCE.md"
    }
  },
  {
    "name": "LICENSE",
    "path": "LICENSE",
    "sha": "114889973c62b6ba2b03ced9baa8e47bec23e6c8",
    "size": 11437,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/LICENSE?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/LICENSE",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/114889973c62b6ba2b03ced9baa8e47bec23e6c8",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/LICENSE",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/LICENSE?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/114889973c62b6ba2b03ced9baa8e47bec23e6c8",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/LICENSE"
    }
  },
  {
    "name": "MAINTAINERS.md",
    "path": "MAINTAINERS.md",
    "sha": "3ecde4ded45862567aa59bb3b9b726e40dc3b80d",
    "size": 1047,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/MAINTAINERS.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/MAINTAINERS.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/3ecde4ded45862567aa59bb3b9b726e40dc3b80d",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/MAINTAINERS.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/MAINTAINERS.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/3ecde4ded45862567aa59bb3b9b726e40dc3b80d",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/MAINTAINERS.md"
    }
  },
  {
    "name": "README.md",
    "path": "README.md",
    "sha": "d441d5bcf99dfa1967dfd7d67e86ae86ebdf365e",
    "size": 24801,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/README.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/README.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d441d5bcf99dfa1967dfd7d67e86ae86ebdf365e",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/README.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/README.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d441d5bcf99dfa1967dfd7d67e86ae86ebdf365e",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/README.md"
    }
  },
  {
    "name": "SECURITY.md",
    "path": "SECURITY.md",
    "sha": "76ba2ce72d79f59ccc3cad027c7d1f30397b41ed",
    "size": 1908,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/SECURITY.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/SECURITY.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/76ba2ce72d79f59ccc3cad027c7d1f30397b41ed",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/SECURITY.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/SECURITY.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/76ba2ce72d79f59ccc3cad027c7d1f30397b41ed",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/SECURITY.md"
    }
  },
  {
    "name": "STYLEGUIDE.md",
    "path": "STYLEGUIDE.md",
    "sha": "e7178f455557f515ebbd260feaf7e67f4f14d064",
    "size": 905,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/STYLEGUIDE.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/STYLEGUIDE.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/e7178f455557f515ebbd260feaf7e67f4f14d064",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/STYLEGUIDE.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/STYLEGUIDE.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/e7178f455557f515ebbd260feaf7e67f4f14d064",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/STYLEGUIDE.md"
    }
  },
  {
    "name": "TESTING.md",
    "path": "TESTING.md",
    "sha": "42495fc23a8f5e1b67cea12b7240e17a3c0f6414",
    "size": 17018,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/TESTING.md?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/TESTING.md",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/42495fc23a8f5e1b67cea12b7240e17a3c0f6414",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/TESTING.md",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/TESTING.md?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/42495fc23a8f5e1b67cea12b7240e17a3c0f6414",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/TESTING.md"
    }
  },
  {
    "name": "THIRD-PARTY-NOTICES",
    "path": "THIRD-PARTY-NOTICES",
    "sha": "8125e5692f0356036202264119839a12ec2830eb",
    "size": 856598,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/THIRD-PARTY-NOTICES?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/THIRD-PARTY-NOTICES",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8125e5692f0356036202264119839a12ec2830eb",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/THIRD-PARTY-NOTICES",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/THIRD-PARTY-NOTICES?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8125e5692f0356036202264119839a12ec2830eb",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/THIRD-PARTY-NOTICES"
    }
  },
  {
    "name": "about.toml",
    "path": "about.toml",
    "sha": "7d93b7361e8d3ea599a58ba188d220831065e169",
    "size": 713,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/about.toml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/about.toml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/7d93b7361e8d3ea599a58ba188d220831065e169",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/about.toml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/about.toml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/7d93b7361e8d3ea599a58ba188d220831065e169",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/about.toml"
    }
  },
  {
    "name": "architecture",
    "path": "architecture",
    "sha": "2f5cc026baec3d85be253c76eb450751b9f090a7",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/architecture?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/architecture",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/2f5cc026baec3d85be253c76eb450751b9f090a7",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/architecture?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/2f5cc026baec3d85be253c76eb450751b9f090a7",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/architecture"
    }
  },
  {
    "name": "buf.yaml",
    "path": "buf.yaml",
    "sha": "77fde365649d49c302d73add8d48e7182b52e237",
    "size": 1288,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/buf.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/buf.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/77fde365649d49c302d73add8d48e7182b52e237",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/buf.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/buf.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/77fde365649d49c302d73add8d48e7182b52e237",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/buf.yaml"
    }
  },
  {
    "name": "crates",
    "path": "crates",
    "sha": "f77076e88f93759b39d1cad21659e5b16381317b",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/crates?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/crates",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/f77076e88f93759b39d1cad21659e5b16381317b",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/crates?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/f77076e88f93759b39d1cad21659e5b16381317b",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/crates"
    }
  },
  {
    "name": "deny.toml",
    "path": "deny.toml",
    "sha": "62ce9a909f9f73938bb6b2433072371ffff88fea",
    "size": 2421,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deny.toml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deny.toml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/62ce9a909f9f73938bb6b2433072371ffff88fea",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deny.toml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deny.toml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/62ce9a909f9f73938bb6b2433072371ffff88fea",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deny.toml"
    }
  },
  {
    "name": "deploy",
    "path": "deploy",
    "sha": "116c01ea1e0f4bd3802571197969ba40e33d7fa3",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/116c01ea1e0f4bd3802571197969ba40e33d7fa3",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/116c01ea1e0f4bd3802571197969ba40e33d7fa3",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy"
    }
  },
  {
    "name": "docs",
    "path": "docs",
    "sha": "918e35f95f9abc01a39dce8cbcec212d79e649d2",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/docs?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/docs",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/918e35f95f9abc01a39dce8cbcec212d79e649d2",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/docs?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/918e35f95f9abc01a39dce8cbcec212d79e649d2",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/docs"
    }
  },
  {
    "name": "e2e",
    "path": "e2e",
    "sha": "57dcd0cc44f901fa9e11fc715e0be55b978ca9ff",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/e2e?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/e2e",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/57dcd0cc44f901fa9e11fc715e0be55b978ca9ff",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/e2e?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/57dcd0cc44f901fa9e11fc715e0be55b978ca9ff",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/e2e"
    }
  },
  {
    "name": "examples",
    "path": "examples",
    "sha": "91bbc3ea888b3f7cdc4804838131e4d92b707c69",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/examples?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/examples",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/91bbc3ea888b3f7cdc4804838131e4d92b707c69",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/examples?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/91bbc3ea888b3f7cdc4804838131e4d92b707c69",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/examples"
    }
  },
  {
    "name": "fern",
    "path": "fern",
    "sha": "99ec0e8d5108371dadbc37e3cdda70e92ea7a23a",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/fern?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/fern",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/99ec0e8d5108371dadbc37e3cdda70e92ea7a23a",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/fern?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/99ec0e8d5108371dadbc37e3cdda70e92ea7a23a",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/fern"
    }
  },
  {
    "name": "flake.lock",
    "path": "flake.lock",
    "sha": "b5861792acaefc1ab5b93ff400b302f90ea074e1",
    "size": 3100,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/flake.lock?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/flake.lock",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b5861792acaefc1ab5b93ff400b302f90ea074e1",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/flake.lock",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/flake.lock?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b5861792acaefc1ab5b93ff400b302f90ea074e1",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/flake.lock"
    }
  },
  {
    "name": "flake.nix",
    "path": "flake.nix",
    "sha": "0aaf9273c7cf4a557a54acf0d1c703154e4b609d",
    "size": 6010,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/flake.nix?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/flake.nix",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0aaf9273c7cf4a557a54acf0d1c703154e4b609d",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/flake.nix",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/flake.nix?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0aaf9273c7cf4a557a54acf0d1c703154e4b609d",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/flake.nix"
    }
  },
  {
    "name": "install.sh",
    "path": "install.sh",
    "sha": "00f8141135cdcccb9fc2610f22fade10d7f6cedb",
    "size": 36385,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/install.sh?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/install.sh",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/00f8141135cdcccb9fc2610f22fade10d7f6cedb",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/install.sh",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/install.sh?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/00f8141135cdcccb9fc2610f22fade10d7f6cedb",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/install.sh"
    }
  },
  {
    "name": "mise.lock",
    "path": "mise.lock",
    "sha": "ec8cccf089523dd69568c67a7b171808a8ebed76",
    "size": 32446,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/mise.lock?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/mise.lock",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ec8cccf089523dd69568c67a7b171808a8ebed76",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/mise.lock",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/mise.lock?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ec8cccf089523dd69568c67a7b171808a8ebed76",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/mise.lock"
    }
  },
  {
    "name": "mise.toml",
    "path": "mise.toml",
    "sha": "9504f82781d27ec77fa183cf922fe6ea72d8a2d1",
    "size": 3027,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/mise.toml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/mise.toml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/9504f82781d27ec77fa183cf922fe6ea72d8a2d1",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/mise.toml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/mise.toml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/9504f82781d27ec77fa183cf922fe6ea72d8a2d1",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/mise.toml"
    }
  },
  {
    "name": "nix",
    "path": "nix",
    "sha": "3f6c06536daebb7b984612b59488784f501effd4",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/nix?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/nix",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/3f6c06536daebb7b984612b59488784f501effd4",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/nix?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/3f6c06536daebb7b984612b59488784f501effd4",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/nix"
    }
  },
  {
    "name": "openshell.spec",
    "path": "openshell.spec",
    "sha": "ee9d7d5f5a827e00b774486d2f39c2ec9aa54baa",
    "size": 11765,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/openshell.spec?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/openshell.spec",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ee9d7d5f5a827e00b774486d2f39c2ec9aa54baa",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/openshell.spec",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/openshell.spec?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ee9d7d5f5a827e00b774486d2f39c2ec9aa54baa",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/openshell.spec"
    }
  },
  {
    "name": "proto",
    "path": "proto",
    "sha": "abd2b3924f15aaf7e7dd0dcaf9f52c0d02854b84",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/proto?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/proto",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/abd2b3924f15aaf7e7dd0dcaf9f52c0d02854b84",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/proto?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/abd2b3924f15aaf7e7dd0dcaf9f52c0d02854b84",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/proto"
    }
  },
  {
    "name": "providers",
    "path": "providers",
    "sha": "87456308999b01cc63899a87a2c788e55ed2f08a",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/providers?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/providers",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/87456308999b01cc63899a87a2c788e55ed2f08a",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/providers?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/87456308999b01cc63899a87a2c788e55ed2f08a",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/providers"
    }
  },
  {
    "name": "pyproject.toml",
    "path": "pyproject.toml",
    "sha": "33ca25ddf003f1de101c4ebc90820041825c892b",
    "size": 3145,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/pyproject.toml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/pyproject.toml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/33ca25ddf003f1de101c4ebc90820041825c892b",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/pyproject.toml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/pyproject.toml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/33ca25ddf003f1de101c4ebc90820041825c892b",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/pyproject.toml"
    }
  },
  {
    "name": "python",
    "path": "python",
    "sha": "6c6454e1bde75b6cbecc4a38443b9c6ec6380140",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/python?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/python",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/6c6454e1bde75b6cbecc4a38443b9c6ec6380140",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/python?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/6c6454e1bde75b6cbecc4a38443b9c6ec6380140",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/python"
    }
  },
  {
    "name": "rfc",
    "path": "rfc",
    "sha": "c15b6345e049e426043f20c2dc445ff489d73efc",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/rfc?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/rfc",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/c15b6345e049e426043f20c2dc445ff489d73efc",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/rfc?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/c15b6345e049e426043f20c2dc445ff489d73efc",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/rfc"
    }
  },
  {
    "name": "rust-toolchain.toml",
    "path": "rust-toolchain.toml",
    "sha": "13beef5d1ae968d441fcb9a22c0d8972a6361f07",
    "size": 230,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/rust-toolchain.toml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/rust-toolchain.toml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/13beef5d1ae968d441fcb9a22c0d8972a6361f07",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/rust-toolchain.toml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/rust-toolchain.toml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/13beef5d1ae968d441fcb9a22c0d8972a6361f07",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/rust-toolchain.toml"
    }
  },
  {
    "name": "scripts",
    "path": "scripts",
    "sha": "e362b82ac77bed1dc675d639a1faefcdc46449e1",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/scripts?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/scripts",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/e362b82ac77bed1dc675d639a1faefcdc46449e1",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/scripts?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/e362b82ac77bed1dc675d639a1faefcdc46449e1",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/scripts"
    }
  },
  {
    "name": "sdk",
    "path": "sdk",
    "sha": "1e72ff1687852b375a571dea7d0449d6537f3d89",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/sdk?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/sdk",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/1e72ff1687852b375a571dea7d0449d6537f3d89",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/sdk?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/1e72ff1687852b375a571dea7d0449d6537f3d89",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/sdk"
    }
  },
  {
    "name": "skills",
    "path": "skills",
    "sha": "85ed669442d9a839394946c2afb0a1ed9958cc62",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/skills?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/skills",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/85ed669442d9a839394946c2afb0a1ed9958cc62",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/skills?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/85ed669442d9a839394946c2afb0a1ed9958cc62",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/skills"
    }
  },
  {
    "name": "snap",
    "path": "snap",
    "sha": "9a5f60b4fd59be1080d1d130451fd4b38d7d2413",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/snap?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/snap",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/9a5f60b4fd59be1080d1d130451fd4b38d7d2413",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/snap?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/9a5f60b4fd59be1080d1d130451fd4b38d7d2413",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/snap"
    }
  },
  {
    "name": "snapcraft.yaml",
    "path": "snapcraft.yaml",
    "sha": "0e2fc821cb4d94ccd7ac65503c064760eb202757",
    "size": 5556,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/snapcraft.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/snapcraft.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0e2fc821cb4d94ccd7ac65503c064760eb202757",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/snapcraft.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/snapcraft.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0e2fc821cb4d94ccd7ac65503c064760eb202757",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/snapcraft.yaml"
    }
  },
  {
    "name": "tasks",
    "path": "tasks",
    "sha": "084b2b2f7679be60651a799aba84efb9dc27bdd6",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/tasks?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/tasks",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/084b2b2f7679be60651a799aba84efb9dc27bdd6",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/tasks?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/084b2b2f7679be60651a799aba84efb9dc27bdd6",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/tasks"
    }
  },
  {
    "name": "telemetry",
    "path": "telemetry",
    "sha": "7e7432a65fe485a8ec848e2cb56bea546615c247",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/telemetry?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/telemetry",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/7e7432a65fe485a8ec848e2cb56bea546615c247",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/telemetry?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/7e7432a65fe485a8ec848e2cb56bea546615c247",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/telemetry"
    }
  },
  {
    "name": "tests",
    "path": "tests",
    "sha": "461246c4ec8c27aac8be3ba2676397ecc340d59b",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/tests?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/tests",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/461246c4ec8c27aac8be3ba2676397ecc340d59b",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/tests?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/461246c4ec8c27aac8be3ba2676397ecc340d59b",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/tests"
    }
  },
  {
    "name": "uv.lock",
    "path": "uv.lock",
    "sha": "063ef739e64174bd5adbdd262ba17f4f1374a760",
    "size": 108461,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/uv.lock?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/uv.lock",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/063ef739e64174bd5adbdd262ba17f4f1374a760",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/uv.lock",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/uv.lock?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/063ef739e64174bd5adbdd262ba17f4f1374a760",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/uv.lock"
    }
  }
]

```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://docs.nvidia.com/openshell/latest/kubernetes/managing-certificates",
  "format": "markdown"
}
```

**Output:**
```
> For clean Markdown of any page, append .md to the page URL.
> For a complete documentation index, see https://docs.nvidia.com/openshell/llms.txt.
> For AI client integration (Claude Code, Cursor, etc.), connect to the MCP server at https://docs.nvidia.com/openshell/_mcp/server.

# Managing Certificates

> Configure the OpenShell Helm chart to use cert-manager for mTLS certificate issuance and automatic renewal.

The OpenShell gateway uses mTLS certificates for transport between the gateway and sandbox supervisors. These certificates are not Kubernetes user authentication; configure OIDC or a trusted access proxy for user access. The Helm chart supports two ways to provision and manage the certificate bundle:

| Mode                            | When to use                                                                                                                                 |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Built-in `pkiInitJob` (default) | The default path. A pre-install Kubernetes Job generates a self-signed CA and certificates during installation. No additional dependencies. |
| cert-manager                    | Production deployments that need automatic certificate rotation managed by a running controller.                                            |

The rest of this page covers switching to cert-manager. The built-in mode requires no configuration.

When `certManager.enabled=true`, cert-manager owns TLS certificate generation.
The chart still runs a JWT-only initialization hook because cert-manager does
not create the sandbox JWT signing Secret required by the gateway. This
cert-manager precedence applies even if `pkiInitJob.enabled` remains true.

## Install cert-manager

Install cert-manager from the OCI registry with CRD support enabled:

```shell
helm upgrade --install cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version v1.20.3 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true \
  --wait
```

Verify the cert-manager pods are running:

```shell
kubectl -n cert-manager get pods
```

## Install OpenShell with cert-manager PKI

Pass the cert-manager values override when installing or upgrading the chart:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set certManager.enabled=true
```

The chart creates a self-signed CA, issues server and client certificates from it, and cert-manager handles renewal before expiry.
The chart also runs a pre-install hook in JWT-only mode to create the gateway's
sandbox JWT signing Secret. That Secret is separate from the cert-manager TLS
certificate Secrets and is mounted at `/etc/openshell-jwt`.

## Using a real Issuer for the server certificate

By default, cert-manager issues both the server and client certificates from
a self-signed CA the chart creates — this rotates automatically, but the
server certificate is still not publicly trusted. `certManager.serverIssuerRef`
overrides the `issuerRef` on the server `Certificate` resource to point at a
real `Issuer` or `ClusterIssuer` instead, for example an ACME issuer:

```shell
helm upgrade --install openshell \
  oci://ghcr.io/nvidia/openshell/helm-chart \
  --version <version> \
  --namespace openshell \
  --set certManager.enabled=true \
  --set certManager.serverIssuerRef.name=letsencrypt-prod \
  --set certManager.serverIssuerRef.kind=ClusterIssuer \
  --set certManager.serverDnsNames[0]=openshell.example.com
```

### Dual certificate architecture

When `serverIssuerRef` is set, the chart creates **two** server certificates:

1. **Internal certificate** (`openshell-server-tls`): signed by the chart CA
   with internal SANs (`*.svc.cluster.local`, `localhost`, etc.).
2. **External certificate** (`openshell-server-external-tls`): signed by the
   configured issuer (e.g. ACME) with only the hostnames from
   `certManager.serverDnsNames`.

The gateway uses **SNI** to select which certificate to present:
supervisors connect via internal service names and receive the internal
certificate (verified against the chart CA they already trust), while CLI
users connecting through a Route or ingress use the external hostname and
receive the ACME certificate. This keeps supervisor trust pinned to only
the operator's chart CA — no WebPKI root trust is needed.

Public CAs such as Let's Encrypt reject certificate requests that include
internal-only names per CA/Browser Forum baseline requirements. The chart
validates this at install time and fails with an actionable error if
`certManager.serverDnsNames` contains internal-only entries while
`serverIssuerRef` is set.

You do **not** need to set `server.grpcEndpoint` to the external hostname.
Supervisors connect via the internal service name automatically. Setting
`server.grpcEndpoint` to an external hostname would cause supervisors to
receive the ACME certificate (via SNI) which they cannot verify against the
chart CA.

The default `clientCaFromServerTlsSecret=true` is correct even when
`serverIssuerRef` is set: the internal server certificate is always signed
by the chart CA (the same CA that signs the client certificate), so its
`ca.crt` is the right trust anchor for mTLS verification.

## Next Steps

Return to [Setup](/kubernetes/setup) to complete the installation. For
exposing the gateway externally on OpenShift with a real certificate, see
[OpenShift](/kubernetes/openshift).
```

---

## Assistant (Build · GPT-5.6 Terra · 21.7s)

_Thinking:_

**Investigating client issues**

I'm reviewing a situation where the client secret is used in tests but the data can only be mounted. It looks like the usual SNI settings are in place for gateways. However, the logs indicate "UnknownCA," which suggests that the client might be loading the wrong Certificate Authority or that the server is unexpectedly serving an external certificate. There could also be an issue with connecting by an incorrect hostname or SNI. I need to explore these points further.

_Thinking:_



No. For default OpenShell internal gateway, `ca.crt` alone likely not enough.

OpenShell chart creates mTLS server and client certificates from chart CA. Direct BFF-to-Service connection needs:

```text
ca.crt       verify gateway server certificate
tls.crt      BFF client identity
tls.key      BFF client private key
```

Not needed:

```text
openshell-server-tls/tls.key  gateway private key
CA/Issuer private key         certificate authority key
OpenShell JWT-signing Secret  sandbox authentication key
```

Your errors identify trust failure first:

- `received fatal alert: UnknownCA`: client rejected gateway server certificate. Wrong/missing CA, wrong SNI, or gateway presented unexpected external certificate.
- `received fatal alert: BadCertificate`: client rejected presented server certificate. Usually hostname/SAN, certificate chain, expiry, or key usage issue.

These alerts were received by gateway, so gateway logged client rejection. They do **not** prove gateway rejected absent client certificate.

Use endpoint `https://localhost:8080` for port-forward, not bare host or external hostname. Verify in two stages:

```sh
# Stage 1: server certificate trust and SAN.
openssl s_client -connect 127.0.0.1:8080 -servername localhost \
  -CAfile ca.crt -verify_return_error </dev/null

# Stage 2: expected OpenShell internal mTLS.
openssl s_client -connect 127.0.0.1:8080 -servername localhost \
  -CAfile ca.crt -cert tls.crt -key tls.key \
  -verify_return_error -alpn h2 </dev/null
```

Stage 1 must validate server chain. It may then fail because gateway requires client cert. Stage 2 should complete TLS.

For production, do not mount generic `openshell-client-tls` into BFF. Create BFF-specific client cert:

```sh
kubectl -n <openshell-namespace> get certificate openshell-client-tls -o yaml
```

Copy its `spec.issuerRef` into a new cert-manager `Certificate`:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: agent-ops-openshell-client
  namespace: <bff-namespace>
spec:
  secretName: agent-ops-openshell-client-tls
  duration: 2160h
  renewBefore: 360h
  commonName: agent-ops-bff
  usages:
    - client auth
  issuerRef:
    name: <same issuer as openshell-client-tls>
    kind: Issuer
```

Issuer must be accessible from BFF namespace. If chart issuer is namespaced, create equivalent issuer in BFF namespace backed by same protected CA, or keep certificate Secret in gateway namespace and mount only into authorized BFF workload.

BFF mounts:

- Internal CA, preferably only `ca.crt`.
- Its own `tls.crt` and `tls.key`.
- No gateway server keys or shared supervisor credentials.

Then code configures all three paths. OIDC bearer forwarding still needed for user authorization; mTLS identifies BFF workload, not dashboard user.

---

## User

apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  annotations:
    meta.helm.sh/release-name: openshell
    meta.helm.sh/release-namespace: openshell
  resourceVersion: '84241479'
  name: openshell-ca
  uid: b8039f20-52b4-4baa-bc26-bd2d13ce9ac4
  creationTimestamp: '2026-09-17T19:50:21Z'
  generation: 1
  managedFields:
    - apiVersion: cert-manager.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        'f:metadata':
          'f:annotations':
            'f:meta.helm.sh/release-name': {}
            'f:meta.helm.sh/release-namespace': {}
          'f:labels':
            'f:app.kubernetes.io/instance': {}
            'f:app.kubernetes.io/managed-by': {}
            'f:app.kubernetes.io/name': {}
            'f:app.kubernetes.io/version': {}
            'f:helm.sh/chart': {}
        'f:spec':
          'f:commonName': {}
          'f:isCA': {}
          'f:issuerRef':
            'f:group': {}
            'f:kind': {}
            'f:name': {}
          'f:privateKey':
            'f:algorithm': {}
            'f:size': {}
          'f:secretName': {}
      manager: helm
      operation: Apply
      time: '2026-09-17T19:50:21Z'
    - apiVersion: cert-manager.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        'f:status':
          .: {}
          'f:conditions':
            .: {}
            'k:{"type":"Ready"}':
              .: {}
              'f:lastTransitionTime': {}
              'f:message': {}
              'f:observedGeneration': {}
              'f:reason': {}
              'f:status': {}
              'f:type': {}
          'f:notAfter': {}
          'f:notBefore': {}
          'f:renewalTime': {}
      manager: cert-manager-certificates-readiness
      operation: Update
      subresource: status
      time: '2026-09-17T19:50:21Z'
  namespace: openshell
  labels:
    app.kubernetes.io/instance: openshell
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: openshell
    app.kubernetes.io/version: 0.0.116
    helm.sh/chart: helm-chart-0.0.116
spec:
  commonName: openshell-ca
  isCA: true
  issuerRef:
    group: cert-manager.io
    kind: Issuer
    name: openshell-selfsigned
  privateKey:
    algorithm: ECDSA
    size: 256
  secretName: openshell-ca-tls
status:
  conditions:
    - lastTransitionTime: '2026-09-17T19:50:21Z'
      message: Certificate is up to date and has not expired
      observedGeneration: 1
      reason: Ready
      status: 'True'
      type: Ready
  notAfter: '2026-12-16T18:14:39Z'
  notBefore: '2026-09-17T18:14:39Z'
  renewalTime: '2026-11-16T18:14:39Z' 

---

## Assistant (Build · GPT-5.6 Terra · 22.5s)

_Thinking:_

**Exploring certificate namespace issues**

I need to consider that a certificate's target secret must be in the same namespace as the certificate itself. If the BFF is in a different namespace, it can't issue with the issuer. One option is to deploy the BFF in the openshell namespace with the certificate secret, but that could present issues since the pod can only mount secrets from its own namespace. Alternatively, I could create an `Issuer` in the BFF namespace, but that would compromise control by copying the CA key there. It may be better to use cert-manager's `CertificateRequestPolicy` or approver to manage this, since standard issuers can't cross namespaces.

_Thinking:_

**Considering ClusterIssuer options**

I'm thinking about creating a `ClusterIssuer` that references the root secret, `openshell-ca-tls`. However, the secret's location might be determined by the cluster resource's secret namespace, which could complicate things. I also need to check if the `ca.secretName` in the `ClusterIssuer` is configurable within this namespace. Ultimately, I need to copy the CA secret to the cert-manager namespace, which raises concerns about storing the root key there, as it won't be in the BFF.

_Thinking:_



This confirms chart CA root. Critical: `openshell-selfsigned` creates root only; BFF must never reference or receive `openshell-ca-tls`. Find leaf-signing Issuer before creating BFF certificate.

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy?ref=main",
  "format": "text"
}
```

**Output:**
```
[
  {
    "name": "deb",
    "path": "deploy/deb",
    "sha": "090d0c72de1c44eba1e3f622a8865b46762d60db",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/deb?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/deb",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/090d0c72de1c44eba1e3f622a8865b46762d60db",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/deb?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/090d0c72de1c44eba1e3f622a8865b46762d60db",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/deb"
    }
  },
  {
    "name": "docker",
    "path": "deploy/docker",
    "sha": "2e31de25374098dba7767154c823b1b42cb982f4",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/docker?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/docker",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/2e31de25374098dba7767154c823b1b42cb982f4",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/docker?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/2e31de25374098dba7767154c823b1b42cb982f4",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/docker"
    }
  },
  {
    "name": "helm",
    "path": "deploy/helm",
    "sha": "53a33010d6629c5c3b21de958081f18ad51d1a0c",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/helm",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/53a33010d6629c5c3b21de958081f18ad51d1a0c",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/53a33010d6629c5c3b21de958081f18ad51d1a0c",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/helm"
    }
  },
  {
    "name": "kube",
    "path": "deploy/kube",
    "sha": "c164b24b50d4ed0e659e7158438fa38a21d3365a",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/kube?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/kube",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/c164b24b50d4ed0e659e7158438fa38a21d3365a",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/kube?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/c164b24b50d4ed0e659e7158438fa38a21d3365a",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/kube"
    }
  },
  {
    "name": "man",
    "path": "deploy/man",
    "sha": "99fc2b5ed12cc308ecc2a55b2c745b7c3c577a87",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/man?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/man",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/99fc2b5ed12cc308ecc2a55b2c745b7c3c577a87",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/man?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/99fc2b5ed12cc308ecc2a55b2c745b7c3c577a87",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/man"
    }
  },
  {
    "name": "rpm",
    "path": "deploy/rpm",
    "sha": "e0f50461f0e6cb5111130d3ae618bb0bbc265f49",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/rpm?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/rpm",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/e0f50461f0e6cb5111130d3ae618bb0bbc265f49",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/rpm?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/e0f50461f0e6cb5111130d3ae618bb0bbc265f49",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/rpm"
    }
  },
  {
    "name": "sbom",
    "path": "deploy/sbom",
    "sha": "fbf269d2d22aff85b63a8679503f8a28c91ccfa7",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/sbom?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/sbom",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/fbf269d2d22aff85b63a8679503f8a28c91ccfa7",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/sbom?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/fbf269d2d22aff85b63a8679503f8a28c91ccfa7",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/sbom"
    }
  }
]

```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/main?recursive=1",
  "format": "text"
}
```

**Output:**
```
{
  "sha": "11f1fe5806797d8c211c29b411017197a3d810e8",
  "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/11f1fe5806797d8c211c29b411017197a3d810e8",
  "tree": [
    {
      "path": ".agents",
      "mode": "040000",
      "type": "tree",
      "sha": "c051e11f371f390d574b74b4e89cdaad11db2ef1",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/c051e11f371f390d574b74b4e89cdaad11db2ef1"
    },
    {
      "path": ".agents/skills",
      "mode": "040000",
      "type": "tree",
      "sha": "6a9eb68cd10a1b74350248b4aeabc89019a868e7",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/6a9eb68cd10a1b74350248b4aeabc89019a868e7"
    },
    {
      "path": ".agents/skills/build-from-issue",
      "mode": "040000",
      "type": "tree",
      "sha": "27c969f268ac2b701f344e4a90588e2b91b6c945",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/27c969f268ac2b701f344e4a90588e2b91b6c945"
    },
    {
      "path": ".agents/skills/build-from-issue/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "89d9e04215680ef6d28d1f55a991c71790cd0a66",
      "size": 29705,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/89d9e04215680ef6d28d1f55a991c71790cd0a66"
    },
    {
      "path": ".agents/skills/build-openshell-mxc-windows",
      "mode": "040000",
      "type": "tree",
      "sha": "864f66e4679d64c5039e2b335055209e06f290de",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/864f66e4679d64c5039e2b335055209e06f290de"
    },
    {
      "path": ".agents/skills/build-openshell-mxc-windows/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "21b49f1a5d9409f2d233ed584c00b64aeef99019",
      "size": 19629,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/21b49f1a5d9409f2d233ed584c00b64aeef99019"
    },
    {
      "path": ".agents/skills/build-openshell-mxc-windows/reference.md",
      "mode": "100644",
      "type": "blob",
      "sha": "5ee638e1a9bd7f28b3063152610633eb2d4f6948",
      "size": 7915,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/5ee638e1a9bd7f28b3063152610633eb2d4f6948"
    },
    {
      "path": ".agents/skills/create-github-issue",
      "mode": "040000",
      "type": "tree",
      "sha": "b05968827e12fe8e42985f6774d8fbe5f253f7b2",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/b05968827e12fe8e42985f6774d8fbe5f253f7b2"
    },
    {
      "path": ".agents/skills/create-github-issue/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "2510b7927201e5f0c72654b581ca37fc20bd3aea",
      "size": 5106,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2510b7927201e5f0c72654b581ca37fc20bd3aea"
    },
    {
      "path": ".agents/skills/create-github-pr",
      "mode": "040000",
      "type": "tree",
      "sha": "84f4cb1f4e57e0a2e389a729b8fbbbb787935597",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/84f4cb1f4e57e0a2e389a729b8fbbbb787935597"
    },
    {
      "path": ".agents/skills/create-github-pr/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "5eafc91f9bf31c1fc78e9c9449a8075d210b7b63",
      "size": 7234,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/5eafc91f9bf31c1fc78e9c9449a8075d210b7b63"
    },
    {
      "path": ".agents/skills/create-rfc",
      "mode": "040000",
      "type": "tree",
      "sha": "e33d2baf6f01933b7b3a13b227f54ca40254cd23",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/e33d2baf6f01933b7b3a13b227f54ca40254cd23"
    },
    {
      "path": ".agents/skills/create-rfc/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "6df87ef315772c8f0104b42d9f945ddec5891226",
      "size": 2647,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/6df87ef315772c8f0104b42d9f945ddec5891226"
    },
    {
      "path": ".agents/skills/create-rfc/agents",
      "mode": "040000",
      "type": "tree",
      "sha": "98a2c4d0fae79d901816b4fb62ab8a47d7d3d09d",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/98a2c4d0fae79d901816b4fb62ab8a47d7d3d09d"
    },
    {
      "path": ".agents/skills/create-rfc/agents/openai.yaml",
      "mode": "100644",
      "type": "blob",
      "sha": "8717be4ff26d709f6b8a061c0477a7e4a0f85698",
      "size": 174,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8717be4ff26d709f6b8a061c0477a7e4a0f85698"
    },
    {
      "path": ".agents/skills/create-spike",
      "mode": "040000",
      "type": "tree",
      "sha": "ea866293a699144ecba287225658b03fde3d8650",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/ea866293a699144ecba287225658b03fde3d8650"
    },
    {
      "path": ".agents/skills/create-spike/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "baafd173dac466ad66fd71c0ca3efa8c56f2b63d",
      "size": 17338,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/baafd173dac466ad66fd71c0ca3efa8c56f2b63d"
    },
    {
      "path": ".agents/skills/fix-security-issue",
      "mode": "040000",
      "type": "tree",
      "sha": "76dc8a6c3a87e3be02179a4287ed2084a522cc83",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/76dc8a6c3a87e3be02179a4287ed2084a522cc83"
    },
    {
      "path": ".agents/skills/fix-security-issue/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "1452cbe66c4c206d322348bd0b84605ff3fccae2",
      "size": 12810,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1452cbe66c4c206d322348bd0b84605ff3fccae2"
    },
    {
      "path": ".agents/skills/helm-dev-environment",
      "mode": "040000",
      "type": "tree",
      "sha": "dfc58703e2cae7f26635fb46a29dd8e459155bfa",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/dfc58703e2cae7f26635fb46a29dd8e459155bfa"
    },
    {
      "path": ".agents/skills/helm-dev-environment/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "cb3023f476fda686c4f2374070fdbd2a3f6da481",
      "size": 19307,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/cb3023f476fda686c4f2374070fdbd2a3f6da481"
    },
    {
      "path": ".agents/skills/launch-openshell-gator",
      "mode": "040000",
      "type": "tree",
      "sha": "ab50f1d55b1651c9c2855fbcaf575442eebd38ec",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/ab50f1d55b1651c9c2855fbcaf575442eebd38ec"
    },
    {
      "path": ".agents/skills/launch-openshell-gator/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "80a84ca74424bbefe755f6f4ea998792cbd1ded4",
      "size": 20298,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/80a84ca74424bbefe755f6f4ea998792cbd1ded4"
    },
    {
      "path": ".agents/skills/review-github-pr",
      "mode": "040000",
      "type": "tree",
      "sha": "dac4e3041758330dededbcf1827e5f0cda05ff47",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/dac4e3041758330dededbcf1827e5f0cda05ff47"
    },
    {
      "path": ".agents/skills/review-github-pr/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "de56cb477044d36f163876f791721123eb0de790",
      "size": 6944,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/de56cb477044d36f163876f791721123eb0de790"
    },
    {
      "path": ".agents/skills/review-security-issue",
      "mode": "040000",
      "type": "tree",
      "sha": "3e76c5cfdc5d196be03af24b990bd9799619ad79",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/3e76c5cfdc5d196be03af24b990bd9799619ad79"
    },
    {
      "path": ".agents/skills/review-security-issue/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "efb054df8045f5f2a4aa45d428718ce3fbe37956",
      "size": 8180,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/efb054df8045f5f2a4aa45d428718ce3fbe37956"
    },
    {
      "path": ".agents/skills/sbom",
      "mode": "040000",
      "type": "tree",
      "sha": "e344f7b368745ea7679247f159c7ca846c723c45",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/e344f7b368745ea7679247f159c7ca846c723c45"
    },
    {
      "path": ".agents/skills/sbom/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "0848363c82a48cf271a623ef991fa4d6c2b480a3",
      "size": 5262,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0848363c82a48cf271a623ef991fa4d6c2b480a3"
    },
    {
      "path": ".agents/skills/sync-agent-infra",
      "mode": "040000",
      "type": "tree",
      "sha": "611ea5e86839f06a7f08ded72b861f4ccf52b811",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/611ea5e86839f06a7f08ded72b861f4ccf52b811"
    },
    {
      "path": ".agents/skills/sync-agent-infra/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "f2800afca7f6435530b7d9410f8f1a4c9a163962",
      "size": 14564,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f2800afca7f6435530b7d9410f8f1a4c9a163962"
    },
    {
      "path": ".agents/skills/test-release-canary",
      "mode": "040000",
      "type": "tree",
      "sha": "9ebabbd7744c181adbcba3ab8ac533a3f785c91d",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/9ebabbd7744c181adbcba3ab8ac533a3f785c91d"
    },
    {
      "path": ".agents/skills/test-release-canary/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "7f279778c4736d6987d2015a7f3dfe80d13f15bb",
      "size": 9599,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/7f279778c4736d6987d2015a7f3dfe80d13f15bb"
    },
    {
      "path": ".agents/skills/triage-issue",
      "mode": "040000",
      "type": "tree",
      "sha": "df85e18a917b0dd2a5dad46f596cd4272cb97c7e",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/df85e18a917b0dd2a5dad46f596cd4272cb97c7e"
    },
    {
      "path": ".agents/skills/triage-issue/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "76ab01bdbe1b67699851a0b96a1bf73aade9f1ed",
      "size": 14264,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/76ab01bdbe1b67699851a0b96a1bf73aade9f1ed"
    },
    {
      "path": ".agents/skills/tui-development",
      "mode": "040000",
      "type": "tree",
      "sha": "8215d6a126834c6121201fdc3dadbd64d5cc59e5",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/8215d6a126834c6121201fdc3dadbd64d5cc59e5"
    },
    {
      "path": ".agents/skills/tui-development/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "055884e1ed91770dc60fd1605883c6342dcb8bb8",
      "size": 34058,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/055884e1ed91770dc60fd1605883c6342dcb8bb8"
    },
    {
      "path": ".agents/skills/update-docs",
      "mode": "040000",
      "type": "tree",
      "sha": "dc372d2ddc7f74bbe973cd88c8e8fc3920806c48",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/dc372d2ddc7f74bbe973cd88c8e8fc3920806c48"
    },
    {
      "path": ".agents/skills/update-docs/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "86edcd2738763ed29cfa6e633a914085d6b2e6cc",
      "size": 8652,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/86edcd2738763ed29cfa6e633a914085d6b2e6cc"
    },
    {
      "path": ".agents/skills/watch-github-actions",
      "mode": "040000",
      "type": "tree",
      "sha": "85d410cbe1384cf640dbd201db4e48225a767987",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/85d410cbe1384cf640dbd201db4e48225a767987"
    },
    {
      "path": ".agents/skills/watch-github-actions/SKILL.md",
      "mode": "100644",
      "type": "blob",
      "sha": "5d926cc647afe7f892596f44de2b91d8ae1a9f8d",
      "size": 5719,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/5d926cc647afe7f892596f44de2b91d8ae1a9f8d"
    },
    {
      "path": ".cargo",
      "mode": "040000",
      "type": "tree",
      "sha": "ae0517287fb2b28e109a37ce219628c82fa72cd1",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/ae0517287fb2b28e109a37ce219628c82fa72cd1"
    },
    {
      "path": ".cargo/config.toml",
      "mode": "100644",
      "type": "blob",
      "sha": "0005fc2bd01ef970c58e627e268fca0635efcdee",
      "size": 409,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0005fc2bd01ef970c58e627e268fca0635efcdee"
    },
    {
      "path": ".claude",
      "mode": "040000",
      "type": "tree",
      "sha": "94804dd1318d6bf0039fc67432d09c6dae4fe886",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/94804dd1318d6bf0039fc67432d09c6dae4fe886"
    },
    {
      "path": ".claude/README.md",
      "mode": "100644",
      "type": "blob",
      "sha": "b6bbd9eda3e92881636f4eac580b25f5e0f7b483",
      "size": 593,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b6bbd9eda3e92881636f4eac580b25f5e0f7b483"
    },
    {
      "path": ".claude/agent-memory",
      "mode": "040000",
      "type": "tree",
      "sha": "8a3e05a35fd1d65b9a4a4dc99c97dc81a56bb8a4",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/8a3e05a35fd1d65b9a4a4dc99c97dc81a56bb8a4"
    },
    {
      "path": ".claude/agent-memory/arch-doc-writer",
      "mode": "040000",
      "type": "tree",
      "sha": "f4143d75950526fdda031e4ecf0181be860ae6ea",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/f4143d75950526fdda031e4ecf0181be860ae6ea"
    },
    {
      "path": ".claude/agent-memory/arch-doc-writer/MEMORY.md",
      "mode": "100644",
      "type": "blob",
      "sha": "67844988a06a01916372be45b93c2ccc84095aa1",
      "size": 16830,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/67844988a06a01916372be45b93c2ccc84095aa1"
    },
    {
      "path": ".claude/agent-memory/principal-engineer-reviewer",
      "mode": "040000",
      "type": "tree",
      "sha": "fbc95665cf1ff88e08cf16abf475dd7053797ce1",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/fbc95665cf1ff88e08cf16abf475dd7053797ce1"
    },
    {
      "path": ".claude/agent-memory/principal-engineer-reviewer/MEMORY.md",
      "mode": "100644",
      "type": "blob",
      "sha": "505250c1d85a4da747899eab9148fa424bf4027c",
      "size": 1478,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/505250c1d85a4da747899eab9148fa424bf4027c"
    },
    {
      "path": ".claude/agents",
      "mode": "040000",
      "type": "tree",
      "sha": "8129615bac175c081ebb16a27995bab3c5302aef",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/8129615bac175c081ebb16a27995bab3c5302aef"
    },
    {
      "path": ".claude/agents/arch-doc-writer.md",
      "mode": "100644",
      "type": "blob",
      "sha": "90fbd707ea682cbcd843ff2612064a534dd3612c",
      "size": 14532,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/90fbd707ea682cbcd843ff2612064a534dd3612c"
    },
    {
      "path": ".claude/agents/principal-engineer-reviewer.md",
      "mode": "100644",
      "type": "blob",
      "sha": "ee316121447483111de1c202a418f6e149e8d75d",
      "size": 13611,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ee316121447483111de1c202a418f6e149e8d75d"
    },
    {
      "path": ".claude/skills",
      "mode": "120000",
      "type": "blob",
      "sha": "ac173151dd1e96ecb9092b70767e7ba4a7de9061",
      "size": 18,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ac173151dd1e96ecb9092b70767e7ba4a7de9061"
    },
    {
      "path": ".config",
      "mode": "040000",
      "type": "tree",
      "sha": "5587f2068758ef8cff4ed6c88612f467dd762466",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/5587f2068758ef8cff4ed6c88612f467dd762466"
    },
    {
      "path": ".config/nextest.toml",
      "mode": "100644",
      "type": "blob",
      "sha": "b8267fb5f5e42177fd8b0f26967d375cea789894",
      "size": 808,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b8267fb5f5e42177fd8b0f26967d375cea789894"
    },
    {
      "path": ".dockerignore",
      "mode": "100644",
      "type": "blob",
      "sha": "192b57b6cefc47e15a776be6d6c0618ce6fcb054",
      "size": 341,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/192b57b6cefc47e15a776be6d6c0618ce6fcb054"
    },
    {
      "path": ".env.example",
      "mode": "100644",
      "type": "blob",
      "sha": "d043379e75158f866e768025b06e938f159b781f",
      "size": 684,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d043379e75158f866e768025b06e938f159b781f"
    },
    {
      "path": ".gitattributes",
      "mode": "100644",
      "type": "blob",
      "sha": "af03ae26af336eee26a9f25f042dc5a4f9102c3b",
      "size": 543,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/af03ae26af336eee26a9f25f042dc5a4f9102c3b"
    },
    {
      "path": ".github",
      "mode": "040000",
      "type": "tree",
      "sha": "a41ddfd1be032dbd5e8e61e36dcb876113f12981",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/a41ddfd1be032dbd5e8e61e36dcb876113f12981"
    },
    {
      "path": ".github/CODEOWNERS",
      "mode": "100644",
      "type": "blob",
      "sha": "4d32b414c162d33082e3a7047dd6b0a1c3b4e108",
      "size": 251,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4d32b414c162d33082e3a7047dd6b0a1c3b4e108"
    },
    {
      "path": ".github/DISCUSSION_TEMPLATE",
      "mode": "040000",
      "type": "tree",
      "sha": "99348c340aae2718090741e5f4ba335fb5cf2e4e",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/99348c340aae2718090741e5f4ba335fb5cf2e4e"
    },
    {
      "path": ".github/DISCUSSION_TEMPLATE/vouch-request.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "b5ee84de5391178251ff8ff847e978237b2481b7",
      "size": 1438,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b5ee84de5391178251ff8ff847e978237b2481b7"
    },
    {
      "path": ".github/ISSUE_TEMPLATE",
      "mode": "040000",
      "type": "tree",
      "sha": "fe1d25aff969f954648b63f47148a5df69dc1581",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/fe1d25aff969f954648b63f47148a5df69dc1581"
    },
    {
      "path": ".github/ISSUE_TEMPLATE/bug_report.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "dd774e25c8bdf8a8529a0d1f1a43ffc32a9e6da3",
      "size": 3302,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/dd774e25c8bdf8a8529a0d1f1a43ffc32a9e6da3"
    },
    {
      "path": ".github/ISSUE_TEMPLATE/config.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "0f6486fde8e421e18377d16f5f13d03ac8942c0d",
      "size": 986,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0f6486fde8e421e18377d16f5f13d03ac8942c0d"
    },
    {
      "path": ".github/ISSUE_TEMPLATE/feature_request.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "10ae9cbe50df1a553e5fbc86b1b50467e2b366a0",
      "size": 3937,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/10ae9cbe50df1a553e5fbc86b1b50467e2b366a0"
    },
    {
      "path": ".github/PULL_REQUEST_TEMPLATE.md",
      "mode": "100644",
      "type": "blob",
      "sha": "2efbaac5d1489e3e958b4fcac6e5bc739da41b8b",
      "size": 811,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2efbaac5d1489e3e958b4fcac6e5bc739da41b8b"
    },
    {
      "path": ".github/VOUCHED.td",
      "mode": "100644",
      "type": "blob",
      "sha": "f569e39f4ed27382ed8b2b838dc79907760f2a70",
      "size": 476,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f569e39f4ed27382ed8b2b838dc79907760f2a70"
    },
    {
      "path": ".github/actionlint-matcher.json",
      "mode": "100644",
      "type": "blob",
      "sha": "4613e1617bfe25b72dbf381d0da6e309112fbcbf",
      "size": 435,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4613e1617bfe25b72dbf381d0da6e309112fbcbf"
    },
    {
      "path": ".github/actionlint-sarif-template.txt",
      "mode": "100644",
      "type": "blob",
      "sha": "01c589e21867d51cdc6e129d7aa4e17a38e6fc99",
      "size": 1757,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/01c589e21867d51cdc6e129d7aa4e17a38e6fc99"
    },
    {
      "path": ".github/actionlint.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "35aa72ea7845311c5735961d3b171d47e4232834",
      "size": 492,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/35aa72ea7845311c5735961d3b171d47e4232834"
    },
    {
      "path": ".github/actions",
      "mode": "040000",
      "type": "tree",
      "sha": "759de2574214ecf64e7db72f7d4285eb769cbb56",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/759de2574214ecf64e7db72f7d4285eb769cbb56"
    },
    {
      "path": ".github/actions/build-docker-image",
      "mode": "040000",
      "type": "tree",
      "sha": "44f3a10e2c8d33a2a718b4cbc9d4d7b90628c192",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/44f3a10e2c8d33a2a718b4cbc9d4d7b90628c192"
    },
    {
      "path": ".github/actions/build-docker-image/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "b3b19f244768684f0661b9c7e888e7d692afb74b",
      "size": 3203,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b3b19f244768684f0661b9c7e888e7d692afb74b"
    },
    {
      "path": ".github/actions/build-rust-binary",
      "mode": "040000",
      "type": "tree",
      "sha": "abb6f7443a316042a0c95d6d1618f3c0f2f1a8a1",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/abb6f7443a316042a0c95d6d1618f3c0f2f1a8a1"
    },
    {
      "path": ".github/actions/build-rust-binary/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "2bc4e3ce8c568a44b394e63ecfed696d055ab720",
      "size": 2837,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2bc4e3ce8c568a44b394e63ecfed696d055ab720"
    },
    {
      "path": ".github/actions/check-job-results",
      "mode": "040000",
      "type": "tree",
      "sha": "05abb039414e082698ef004562c564cbe02fdfa0",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/05abb039414e082698ef004562c564cbe02fdfa0"
    },
    {
      "path": ".github/actions/check-job-results/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "5d2aa445bbe571322314b14158c0b66140beda19",
      "size": 886,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/5d2aa445bbe571322314b14158c0b66140beda19"
    },
    {
      "path": ".github/actions/pr-gate",
      "mode": "040000",
      "type": "tree",
      "sha": "5fe6ae4eb5b173c1ad3a95ece283872e6cd0627a",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/5fe6ae4eb5b173c1ad3a95ece283872e6cd0627a"
    },
    {
      "path": ".github/actions/pr-gate/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "e21fb488eb3814395a412353ea46f8d5de718032",
      "size": 2476,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/e21fb488eb3814395a412353ea46f8d5de718032"
    },
    {
      "path": ".github/actions/pr-merge-base",
      "mode": "040000",
      "type": "tree",
      "sha": "8a8d6cf8b8c6f203a2b1bee7b930dc9df7b89d82",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/8a8d6cf8b8c6f203a2b1bee7b930dc9df7b89d82"
    },
    {
      "path": ".github/actions/pr-merge-base/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "a702e9c0b4d8813bb7243f4498f7b2e22464b28a",
      "size": 1361,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a702e9c0b4d8813bb7243f4498f7b2e22464b28a"
    },
    {
      "path": ".github/actions/release-helm-oci",
      "mode": "040000",
      "type": "tree",
      "sha": "6306c959309f2986948a270ea1eafdffeb7aaaf4",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/6306c959309f2986948a270ea1eafdffeb7aaaf4"
    },
    {
      "path": ".github/actions/release-helm-oci/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "2d1a1b86b25ddfdf585e5dc08ff6779147b7430f",
      "size": 4815,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2d1a1b86b25ddfdf585e5dc08ff6779147b7430f"
    },
    {
      "path": ".github/actions/setup-buildx",
      "mode": "040000",
      "type": "tree",
      "sha": "376ea21a7d998cec92575fb4f643cc69a94c6069",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/376ea21a7d998cec92575fb4f643cc69a94c6069"
    },
    {
      "path": ".github/actions/setup-buildx/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "ace7c60f5eb120d571fbfbb4a68e391904ccee18",
      "size": 1216,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ace7c60f5eb120d571fbfbb4a68e391904ccee18"
    },
    {
      "path": ".github/actions/setup-e2e-cli",
      "mode": "040000",
      "type": "tree",
      "sha": "93d94c6064b91ee002a9d907b31bddbbce0f0353",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/93d94c6064b91ee002a9d907b31bddbbce0f0353"
    },
    {
      "path": ".github/actions/setup-e2e-cli/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "d7d1c8a29655e728becdaffca2c518b1666916ae",
      "size": 2108,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d7d1c8a29655e728becdaffca2c518b1666916ae"
    },
    {
      "path": ".github/actions/setup-e2e-driver",
      "mode": "040000",
      "type": "tree",
      "sha": "93fa907815ec69b9adae646d992df380978d6d80",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/93fa907815ec69b9adae646d992df380978d6d80"
    },
    {
      "path": ".github/actions/setup-e2e-driver/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "c2172b2a63405a6ed5809e687d3f7f00d2007ebb",
      "size": 1256,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/c2172b2a63405a6ed5809e687d3f7f00d2007ebb"
    },
    {
      "path": ".github/actions/setup-e2e-gateway",
      "mode": "040000",
      "type": "tree",
      "sha": "17dd2029d71f2c6f13499eee08c2057d5baa00af",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/17dd2029d71f2c6f13499eee08c2057d5baa00af"
    },
    {
      "path": ".github/actions/setup-e2e-gateway/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "255a7f4e4c10953c47d4c8c6936e1bee0dcb8b22",
      "size": 852,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/255a7f4e4c10953c47d4c8c6936e1bee0dcb8b22"
    },
    {
      "path": ".github/actions/setup-e2e-kind",
      "mode": "040000",
      "type": "tree",
      "sha": "b3729badfb92c7cdee93770f5320e379a1ddd466",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/b3729badfb92c7cdee93770f5320e379a1ddd466"
    },
    {
      "path": ".github/actions/setup-e2e-kind/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "e617a27717fc78e94201ad75c3f0ff4d1ca29300",
      "size": 3361,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/e617a27717fc78e94201ad75c3f0ff4d1ca29300"
    },
    {
      "path": ".github/actions/setup-e2e-podman",
      "mode": "040000",
      "type": "tree",
      "sha": "3a5f221d887c2ba224e42d77b8c034f19726c58b",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/3a5f221d887c2ba224e42d77b8c034f19726c58b"
    },
    {
      "path": ".github/actions/setup-e2e-podman/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "3e6f481f4f52db68908aa0bafb9e8f5585c7fd76",
      "size": 4970,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/3e6f481f4f52db68908aa0bafb9e8f5585c7fd76"
    },
    {
      "path": ".github/actions/setup-e2e-vm-driver",
      "mode": "040000",
      "type": "tree",
      "sha": "b25a665c87613abb5fbc3301140ddaba76fa1c92",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/b25a665c87613abb5fbc3301140ddaba76fa1c92"
    },
    {
      "path": ".github/actions/setup-e2e-vm-driver/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "a074ed97a8c8ff963c1da802d535ce3ef5e9daad",
      "size": 742,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a074ed97a8c8ff963c1da802d535ce3ef5e9daad"
    },
    {
      "path": ".github/actions/setup-e2e-vm",
      "mode": "040000",
      "type": "tree",
      "sha": "9b162e5611d9d0f65341c50363df59836f3f41a6",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/9b162e5611d9d0f65341c50363df59836f3f41a6"
    },
    {
      "path": ".github/actions/setup-e2e-vm/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "6776f9927f6b94b4c00d8e9504ed9625e1dcd3d4",
      "size": 1884,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/6776f9927f6b94b4c00d8e9504ed9625e1dcd3d4"
    },
    {
      "path": ".github/actions/setup-mise",
      "mode": "040000",
      "type": "tree",
      "sha": "e49befaab9d4a240bf12004944aa8531e80aeef9",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/e49befaab9d4a240bf12004944aa8531e80aeef9"
    },
    {
      "path": ".github/actions/setup-mise/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "ecc82f47bea99f61aa23ee3df78d599e256d5c74",
      "size": 1525,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ecc82f47bea99f61aa23ee3df78d599e256d5c74"
    },
    {
      "path": ".github/actions/setup-nix",
      "mode": "040000",
      "type": "tree",
      "sha": "ab7b70fe4c947a8e55e736cff4624e942596c218",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/ab7b70fe4c947a8e55e736cff4624e942596c218"
    },
    {
      "path": ".github/actions/setup-nix/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "3cd6940c057e9bce8082dd7764b2b2a00520246f",
      "size": 762,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/3cd6940c057e9bce8082dd7764b2b2a00520246f"
    },
    {
      "path": ".github/actions/setup-rust",
      "mode": "040000",
      "type": "tree",
      "sha": "86c63ca52340091fffc682ae6a90e627dce58694",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/86c63ca52340091fffc682ae6a90e627dce58694"
    },
    {
      "path": ".github/actions/setup-rust/action.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "f6b579d26a7ecfcc3fd44a0dd835e541bd68236f",
      "size": 1240,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f6b579d26a7ecfcc3fd44a0dd835e541bd68236f"
    },
    {
      "path": ".github/codeql",
      "mode": "040000",
      "type": "tree",
      "sha": "b0b76fd348271bce3f1a4f38a017908d727f1b58",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/b0b76fd348271bce3f1a4f38a017908d727f1b58"
    },
    {
      "path": ".github/codeql/codeql-config.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "5fc088727286eb565fd4050105b1710ebedf2c23",
      "size": 420,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/5fc088727286eb565fd4050105b1710ebedf2c23"
    },
    {
      "path": ".github/copy-pr-bot.yaml",
      "mode": "100644",
      "type": "blob",
      "sha": "4cfbdc7f05f17ca41bf53589524bb0f9564d1891",
      "size": 59,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4cfbdc7f05f17ca41bf53589524bb0f9564d1891"
    },
    {
      "path": ".github/dependabot.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "29aeab9aa13ef48b29c6b5cb433adc0cdda0eb96",
      "size": 153,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/29aeab9aa13ef48b29c6b5cb433adc0cdda0eb96"
    },
    {
      "path": ".github/workflows",
      "mode": "040000",
      "type": "tree",
      "sha": "2e9f1d3e41108519872f4310c25f881d018e4a45",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/2e9f1d3e41108519872f4310c25f881d018e4a45"
    },
    {
      "path": ".github/workflows/branch-checks.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "04f1dc87eab068fcc55bd4f3a42f64438882fddb",
      "size": 15968,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/04f1dc87eab068fcc55bd4f3a42f64438882fddb"
    },
    {
      "path": ".github/workflows/branch-docs.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "9333a6ee1e7ab10f396e43c4c6cea76e21f1439a",
      "size": 2993,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/9333a6ee1e7ab10f396e43c4c6cea76e21f1439a"
    },
    {
      "path": ".github/workflows/branch-e2e.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "99c68b979374b7aeac10620a5e5011c8d61cb343",
      "size": 19183,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/99c68b979374b7aeac10620a5e5011c8d61cb343"
    },
    {
      "path": ".github/workflows/build-binaries.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "ad34ee792680078fe7ae176c9264a7b4f650ff13",
      "size": 4746,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ad34ee792680078fe7ae176c9264a7b4f650ff13"
    },
    {
      "path": ".github/workflows/build-images.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "1aa513d5780afbfb15b119904eada8c7e707c558",
      "size": 1506,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1aa513d5780afbfb15b119904eada8c7e707c558"
    },
    {
      "path": ".github/workflows/build-rpm.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "0b07af316087f3f946c226ff548f06a28595b70f",
      "size": 4018,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0b07af316087f3f946c226ff548f06a28595b70f"
    },
    {
      "path": ".github/workflows/build-vm-driver.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "c20706cdb253f3d707ebd081d9f24d2623b0ab96",
      "size": 3726,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/c20706cdb253f3d707ebd081d9f24d2623b0ab96"
    },
    {
      "path": ".github/workflows/cargo-deny.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "c85e9ac758ca57f5e59709dd40b16a6de29e89b9",
      "size": 1533,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/c85e9ac758ca57f5e59709dd40b16a6de29e89b9"
    },
    {
      "path": ".github/workflows/ci-image.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "581bc2d080543d1dffebf2437d29a5baef33b744",
      "size": 3406,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/581bc2d080543d1dffebf2437d29a5baef33b744"
    },
    {
      "path": ".github/workflows/codeql.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "288b39b962533f5a7264855461ef52394d9de6f9",
      "size": 5909,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/288b39b962533f5a7264855461ef52394d9de6f9"
    },
    {
      "path": ".github/workflows/codex-security.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "348817ba8957a58d7803d080a81ae77b8dfcef33",
      "size": 12487,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/348817ba8957a58d7803d080a81ae77b8dfcef33"
    },
    {
      "path": ".github/workflows/dco.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "1bb69a15d95e73a44da88105a1792d85b1753c6c",
      "size": 1816,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1bb69a15d95e73a44da88105a1792d85b1753c6c"
    },
    {
      "path": ".github/workflows/deb-package.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "1ec52dc3a1c300e4471114c763ad0df0590405dd",
      "size": 3194,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1ec52dc3a1c300e4471114c763ad0df0590405dd"
    },
    {
      "path": ".github/workflows/dependency-review.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "fc162fd2cbda0342339eaa97028e5e6d6ecec4ef",
      "size": 3494,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/fc162fd2cbda0342339eaa97028e5e6d6ecec4ef"
    },
    {
      "path": ".github/workflows/e2e-docker-test.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "24eb4a54f72ebfc51ad0d886fe35af1f3495e8c2",
      "size": 4297,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/24eb4a54f72ebfc51ad0d886fe35af1f3495e8c2"
    },
    {
      "path": ".github/workflows/e2e-gpu-test.yaml",
      "mode": "100644",
      "type": "blob",
      "sha": "b363a0509b493229eb01c0de7029118a1b01e92c",
      "size": 3089,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b363a0509b493229eb01c0de7029118a1b01e92c"
    },
    {
      "path": ".github/workflows/e2e-kubernetes-test.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "4aaa93264433aeda77301e1829651ae67bd1ee71",
      "size": 5543,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4aaa93264433aeda77301e1829651ae67bd1ee71"
    },
    {
      "path": ".github/workflows/e2e-label-help.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "b412cdc6f220d557869653d218a3445f30ec852f",
      "size": 4492,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b412cdc6f220d557869653d218a3445f30ec852f"
    },
    {
      "path": ".github/workflows/e2e-podman-test.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "4ca40da39cb08e278afeed50a53142a19984e690",
      "size": 3069,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4ca40da39cb08e278afeed50a53142a19984e690"
    },
    {
      "path": ".github/workflows/e2e-vm-test.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "954de9669ebcabf2d3ea541bc93476c77e8ff6b7",
      "size": 1623,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/954de9669ebcabf2d3ea541bc93476c77e8ff6b7"
    },
    {
      "path": ".github/workflows/helm-lint.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "0ca4473d901d8c2f320763a08510cd7497ea0548",
      "size": 2811,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/0ca4473d901d8c2f320763a08510cd7497ea0548"
    },
    {
      "path": ".github/workflows/integration-runner.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "1048e5abd37d1adda8d7a19759021a6c249973d1",
      "size": 2903,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1048e5abd37d1adda8d7a19759021a6c249973d1"
    },
    {
      "path": ".github/workflows/integration-test.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "bf23de2ae512987c852f4b54ed397b0b020e67d9",
      "size": 1587,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/bf23de2ae512987c852f4b54ed397b0b020e67d9"
    },
    {
      "path": ".github/workflows/issue-triage.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "4aa3d6697ee42b74e008a5231a39048c7f388a1a",
      "size": 1829,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4aa3d6697ee42b74e008a5231a39048c7f388a1a"
    },
    {
      "path": ".github/workflows/package-release-binaries.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "fb8df4ada260e3f4516b498bd6bb6e6c367eb601",
      "size": 7934,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/fb8df4ada260e3f4516b498bd6bb6e6c367eb601"
    },
    {
      "path": ".github/workflows/prepare-integration-inputs.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "1fb14e583500d033ee7dbacffec77f8cd1ac62eb",
      "size": 4523,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1fb14e583500d033ee7dbacffec77f8cd1ac62eb"
    },
    {
      "path": ".github/workflows/publish-docs-website.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "5c99492b6ef15f985a6f3351238e15da78957a48",
      "size": 1954,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/5c99492b6ef15f985a6f3351238e15da78957a48"
    },
    {
      "path": ".github/workflows/release-auto-tag.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "b0ae1a268631f9ed7633b205fe072b67f557fcb1",
      "size": 3087,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b0ae1a268631f9ed7633b205fe072b67f557fcb1"
    },
    {
      "path": ".github/workflows/release-canary.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "7a458e7ef7d96c78f438dca89fd1cef9ea081ad6",
      "size": 14406,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/7a458e7ef7d96c78f438dca89fd1cef9ea081ad6"
    },
    {
      "path": ".github/workflows/release-dev.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "c188ba3dc6a39677b6fccf418b9018d954efe34f",
      "size": 22325,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/c188ba3dc6a39677b6fccf418b9018d954efe34f"
    },
    {
      "path": ".github/workflows/release-tag.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "10465d251dc2686ad1b7277fc2e81115c7b6ba50",
      "size": 33340,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/10465d251dc2686ad1b7277fc2e81115c7b6ba50"
    },
    {
      "path": ".github/workflows/release-vm-kernel.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "8a31c3adbed937d174e3654efcd5141a3aab9b64",
      "size": 10185,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8a31c3adbed937d174e3654efcd5141a3aab9b64"
    },
    {
      "path": ".github/workflows/required-ci-gates.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "b5b521a8f3ff6011979551ab4756d60bba304ea5",
      "size": 10979,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b5b521a8f3ff6011979551ab4756d60bba304ea5"
    },
    {
      "path": ".github/workflows/rpm-package.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "5e2e50ac59071bb5941ee2a0aacc0d86cc5edbb4",
      "size": 1616,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/5e2e50ac59071bb5941ee2a0aacc0d86cc5edbb4"
    },
    {
      "path": ".github/workflows/security-scan.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "36a7fec8c37ae489f3cacc93e16de04b63345ae3",
      "size": 4307,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/36a7fec8c37ae489f3cacc93e16de04b63345ae3"
    },
    {
      "path": ".github/workflows/snap-package.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "7eb78d3a2799b043bb154b2c76df7d9d106b3b61",
      "size": 6287,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/7eb78d3a2799b043bb154b2c76df7d9d106b3b61"
    },
    {
      "path": ".github/workflows/stale.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "e45ccbf5ebb81c24c358969c10d10c06c38a582c",
      "size": 2152,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/e45ccbf5ebb81c24c358969c10d10c06c38a582c"
    },
    {
      "path": ".github/workflows/sync-docs.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "63c4a6b8b8a3ac1657872610a388dea38bd76b96",
      "size": 8455,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/63c4a6b8b8a3ac1657872610a388dea38bd76b96"
    },
    {
      "path": ".github/workflows/trivy-changes.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "15e9b7af33f56cca3309588e93d4b8517e5fbc13",
      "size": 6468,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/15e9b7af33f56cca3309588e93d4b8517e5fbc13"
    },
    {
      "path": ".github/workflows/trivy-scan.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "cad728132172e2cb37a60ab75dd986271eb191d2",
      "size": 7455,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/cad728132172e2cb37a60ab75dd986271eb191d2"
    },
    {
      "path": ".github/workflows/vouch-check.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "2eeeb949f9a69cf1bb550180fb4c23a008a4484b",
      "size": 4515,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2eeeb949f9a69cf1bb550180fb4c23a008a4484b"
    },
    {
      "path": ".github/workflows/vouch-command.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "2c258e8e137ab9474129d1e8b3384fbc9134f723",
      "size": 7383,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2c258e8e137ab9474129d1e8b3384fbc9134f723"
    },
    {
      "path": ".github/workflows/windows-msvc.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "a268629a22154770204722aa1bf2a1c30825a3c2",
      "size": 7872,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a268629a22154770204722aa1bf2a1c30825a3c2"
    },
    {
      "path": ".github/workflows/workflow-security.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "a098b2c4ab110f730a73a92c64f5f5b0f05187c6",
      "size": 7720,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a098b2c4ab110f730a73a92c64f5f5b0f05187c6"
    },
    {
      "path": ".github/zizmor.yml",
      "mode": "100644",
      "type": "blob",
      "sha": "7b1e035b77c43c0d776bf376152238076c541880",
      "size": 474,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/7b1e035b77c43c0d776bf376152238076c541880"
    },
    {
      "path": ".gitignore",
      "mode": "100644",
      "type": "blob",
      "sha": "b4e8cd16317f1e6e09130ada29310f82cb6e41a2",
      "size": 3885,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b4e8cd16317f1e6e09130ada29310f82cb6e41a2"
    },
    {
      "path": ".markdownlint-cli2.jsonc",
      "mode": "100644",
      "type": "blob",
      "sha": "125df0f812ebadb206338e562d2e763626aadc3c",
      "size": 1521,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/125df0f812ebadb206338e562d2e763626aadc3c"
    },
    {
      "path": ".opencode",
      "mode": "040000",
      "type": "tree",
      "sha": "76a26f4179e0f43c0cc50a64bc6e5d5111d30cdc",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/76a26f4179e0f43c0cc50a64bc6e5d5111d30cdc"
    },
    {
      "path": ".opencode/agents",
      "mode": "040000",
      "type": "tree",
      "sha": "b47aaa5b6fc000b62074a124d36e18752a7cc121",
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/b47aaa5b6fc000b62074a124d36e18752a7cc121"
    },
    {
      "path": ".opencode/agents/arch-doc-writer.md",
      "mode": "100644",
      "type": "blob",
      "sha": "ac0e2540a203650c52918cbda22dcba07373ff43",
      "size": 9499,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ac0e2540a203650c52918cbda22dcba07373ff43"
    },
    {
      "path": ".opencode/agents/principal-engineer-reviewer.md",
      "mode": "100644",
      "type": "blob",
      "sha": "68c3a86d0ef1f1480f96e9e25050511d621a9ef8",
      "size": 6693,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/68c3a86d0ef1f1480f96e9e25050511d621a9ef8"
    },
    {
      "path": ".packit.yaml",
      "mode": "100644",
      "type": "blob",
      "sha": "d3b92eafae007f5b7a1bb70ab2c3ade0c1dc19a8",
      "size": 3910,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d3b92eafae007f5b7a1bb70ab2c3ade0c1dc19a8"
    },
    {
      "path": ".python-version",
      "mode": "100644",
      "type": "blob",
      "sha": "a6d9ada03ee63903081d538cc0a139a1b9658908",
      "size": 7,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a6d9ada03ee63903081d538cc0a139a1b9658908"
    },
    {
      "path": ".trivyignore.yaml",
      "mode": "100644",
      "type": "blob",
      "sha": "4cc4e32ffe9f4647c0d32b9c8b8578edc020c1af",
      "size": 1713,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4cc4e32ffe9f4647c0d32b9c8b8578edc020c1af"
    },
    {
      "path": "AGENTS.md",
      "mode": "100644",
      "type": "blob",
      "sha": "b1a0ffeb60e4465206fe508a2890593e92f96149",
      "size": 22662,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b1a0ffeb60e4465206fe508a2890593e92f96149"
    },
    {
      "path": "CI.md",
      "mode": "100644",
      "type": "blob",
      "sha": "1324ca9d7f46c185d633e9fb4d11c1aa8ced9b60",
      "size": 28230,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1324ca9d7f46c185d633e9fb4d11c1aa8ced9b60"
    },
    {
      "path": "CLAUDE.md",
      "mode": "100644",
      "type": "blob",
      "sha": "43c994c2d3617f947bcb5adf1933e21dabe46bb5",
      "size": 11,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/43c994c2d3617f947bcb5adf1933e21dabe46bb5"
    },
    {
      "path": "CODE_OF_CONDUCT.md",
      "mode": "100644",
      "type": "blob",
      "sha": "4575d16002f1fa93fd385608c44462679987cb7c",
      "size": 335,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4575d16002f1fa93fd385608c44462679987cb7c"
    },
    {
      "path": "CONTRIBUTING.md",
      "mode": "100644",
      "type": "blob",
      "sha": "4213239b35a729e39c56d26f4aa0b6b66011f7ad",
      "size": 35156,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/4213239b35a729e39c56d26f4aa0b6b66011f7ad"
    },
    {
      "path": "Cargo.lock",
      "mode": "100644",
      "type": "blob",
      "sha": "53a3654ebfb186e913c38960e53208097004bbfb",
      "size": 214762,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/53a3654ebfb186e913c38960e53208097004bbfb"
    },
    {
      "path": "Cargo.toml",
      "mode": "100644",
      "type": "blob",
      "sha": "2e7237190f3f0c6fc11b6b6a4f9c5d14eaac6a5a",
      "size": 6121,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/2e7237190f3f0c6fc11b6b6a4f9c5d14eaac6a5a"
    },
    {
      "path": "DCO",
      "mode": "100644",
      "type": "blob",
      "sha": "49b8cb0549267a8176467738b172a63d86eff436",
      "size": 1366,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/49b8cb0549267a8176467738b172a63d86eff436"
    },
    {
      "path": "GOVERNANCE.md",
      "mode": "100644",
      "type": "blob",
      "sha": "e2c7cfbbef31331f76e2a596b6dbd052736b4834",
      "size": 13590,
      "url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/e2c7cfbbef31331f76e2a596b6dbd052736b4834"
    },
    {
      "path": "LICENSE",
      "mode": "100644",
      "type": "blob",

...636404 bytes truncated...

The tool call succeeded but the output was truncated. Full output saved to: /Users/derxu/.local/share/opencode/tool-output/tool_0ced499ad001jpYlIGryPCMUy3
Use the Task tool to have explore agent process this file with Grep and Read (with offset/limit). Do NOT read the full file yourself - delegate to save context.
```

---

## Assistant (Build · GPT-5.6 Terra · 4.2s)

_Thinking:_

**Inspecting the chart**

I see that there's a need to inspect something, but I'm not sure we really need to go that deep. I think I could check the `deploy/helm/helm-chart/templates/cert-manager.yaml` file directly on GitHub, to see what's there. I might also need to list out the possible paths or directories for deployment. There’s quite a bit to think through, but I want to make sure I’m approaching this correctly!

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm?ref=main",
  "format": "text"
}
```

**Output:**
```
[
  {
    "name": "openshell-workspace",
    "path": "deploy/helm/openshell-workspace",
    "sha": "9ca7cf03327c888cfca559703d0ee7c0a52ad77a",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell-workspace?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/helm/openshell-workspace",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/9ca7cf03327c888cfca559703d0ee7c0a52ad77a",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell-workspace?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/9ca7cf03327c888cfca559703d0ee7c0a52ad77a",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/helm/openshell-workspace"
    }
  },
  {
    "name": "openshell",
    "path": "deploy/helm/openshell",
    "sha": "4cc0cc293488a3f46fbcd35ac2fec6a9724b790c",
    "size": 0,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/helm/openshell",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/4cc0cc293488a3f46fbcd35ac2fec6a9724b790c",
    "download_url": null,
    "type": "dir",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/trees/4cc0cc293488a3f46fbcd35ac2fec6a9724b790c",
      "html": "https://github.com/NVIDIA/OpenShell/tree/main/deploy/helm/openshell"
    }
  },
  {
    "name": "test-split-ownership.sh",
    "path": "deploy/helm/test-split-ownership.sh",
    "sha": "a7a865363bfbf4bace9be572bb6317914ba6fef3",
    "size": 2601,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/test-split-ownership.sh?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/test-split-ownership.sh",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a7a865363bfbf4bace9be572bb6317914ba6fef3",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/test-split-ownership.sh",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/test-split-ownership.sh?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/a7a865363bfbf4bace9be572bb6317914ba6fef3",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/test-split-ownership.sh"
    }
  }
]

```

---

## Assistant (Build · GPT-5.6 Terra · 4.1s)

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates?ref=main",
  "format": "text"
}
```

**Output:**
```
[
  {
    "name": "_gateway-workload.tpl",
    "path": "deploy/helm/openshell/templates/_gateway-workload.tpl",
    "sha": "67ed46d140fb83634520f4c877ed8d0451c006d2",
    "size": 12588,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/_gateway-workload.tpl?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/_gateway-workload.tpl",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/67ed46d140fb83634520f4c877ed8d0451c006d2",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/_gateway-workload.tpl",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/_gateway-workload.tpl?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/67ed46d140fb83634520f4c877ed8d0451c006d2",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/_gateway-workload.tpl"
    }
  },
  {
    "name": "_helpers.tpl",
    "path": "deploy/helm/openshell/templates/_helpers.tpl",
    "sha": "03713b9c269956e2044dd0196cc4df4691d0d402",
    "size": 13251,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/_helpers.tpl?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/_helpers.tpl",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/03713b9c269956e2044dd0196cc4df4691d0d402",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/_helpers.tpl",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/_helpers.tpl?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/03713b9c269956e2044dd0196cc4df4691d0d402",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/_helpers.tpl"
    }
  },
  {
    "name": "agent-sandbox-preflight.yaml",
    "path": "deploy/helm/openshell/templates/agent-sandbox-preflight.yaml",
    "sha": "83a742f67093fbb6087c59c885a29011ecd1effd",
    "size": 864,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/agent-sandbox-preflight.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/agent-sandbox-preflight.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/83a742f67093fbb6087c59c885a29011ecd1effd",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/agent-sandbox-preflight.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/agent-sandbox-preflight.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/83a742f67093fbb6087c59c885a29011ecd1effd",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/agent-sandbox-preflight.yaml"
    }
  },
  {
    "name": "backend-tls-policy.yaml",
    "path": "deploy/helm/openshell/templates/backend-tls-policy.yaml",
    "sha": "8d7a4fd557861c48f8633aee5b9427d0f85365ad",
    "size": 1219,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/backend-tls-policy.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/backend-tls-policy.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8d7a4fd557861c48f8633aee5b9427d0f85365ad",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/backend-tls-policy.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/backend-tls-policy.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8d7a4fd557861c48f8633aee5b9427d0f85365ad",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/backend-tls-policy.yaml"
    }
  },
  {
    "name": "cert-manager-pki.yaml",
    "path": "deploy/helm/openshell/templates/cert-manager-pki.yaml",
    "sha": "838534e62ed80e9952a37f57e5dc6e39db810662",
    "size": 6161,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/cert-manager-pki.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/cert-manager-pki.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/838534e62ed80e9952a37f57e5dc6e39db810662",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/cert-manager-pki.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/cert-manager-pki.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/838534e62ed80e9952a37f57e5dc6e39db810662",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/cert-manager-pki.yaml"
    }
  },
  {
    "name": "certgen.yaml",
    "path": "deploy/helm/openshell/templates/certgen.yaml",
    "sha": "b2ce8d2a82f7126c42deb1ca1418b5edb1956473",
    "size": 6237,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/certgen.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/certgen.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b2ce8d2a82f7126c42deb1ca1418b5edb1956473",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/certgen.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/certgen.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/b2ce8d2a82f7126c42deb1ca1418b5edb1956473",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/certgen.yaml"
    }
  },
  {
    "name": "clusterrole.yaml",
    "path": "deploy/helm/openshell/templates/clusterrole.yaml",
    "sha": "8bd68eeb5c136439edd914a5233f89ef4b1e85d4",
    "size": 4999,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/clusterrole.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/clusterrole.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8bd68eeb5c136439edd914a5233f89ef4b1e85d4",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/clusterrole.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/clusterrole.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8bd68eeb5c136439edd914a5233f89ef4b1e85d4",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/clusterrole.yaml"
    }
  },
  {
    "name": "clusterrolebinding.yaml",
    "path": "deploy/helm/openshell/templates/clusterrolebinding.yaml",
    "sha": "9b3b254586d48a80ca8b85cc36664c01d57efbe2",
    "size": 644,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/clusterrolebinding.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/clusterrolebinding.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/9b3b254586d48a80ca8b85cc36664c01d57efbe2",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/clusterrolebinding.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/clusterrolebinding.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/9b3b254586d48a80ca8b85cc36664c01d57efbe2",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/clusterrolebinding.yaml"
    }
  },
  {
    "name": "credential-secrets-role.yaml",
    "path": "deploy/helm/openshell/templates/credential-secrets-role.yaml",
    "sha": "f6187c9acb6997fe61bf9d4d9d21796a9a99a4b7",
    "size": 1086,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/credential-secrets-role.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/credential-secrets-role.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f6187c9acb6997fe61bf9d4d9d21796a9a99a4b7",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/credential-secrets-role.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/credential-secrets-role.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f6187c9acb6997fe61bf9d4d9d21796a9a99a4b7",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/credential-secrets-role.yaml"
    }
  },
  {
    "name": "credential-secrets-rolebinding.yaml",
    "path": "deploy/helm/openshell/templates/credential-secrets-rolebinding.yaml",
    "sha": "3a9ee0bddc6a46bcbb2787b782c5a412dc435aa9",
    "size": 818,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/credential-secrets-rolebinding.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/credential-secrets-rolebinding.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/3a9ee0bddc6a46bcbb2787b782c5a412dc435aa9",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/credential-secrets-rolebinding.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/credential-secrets-rolebinding.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/3a9ee0bddc6a46bcbb2787b782c5a412dc435aa9",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/credential-secrets-rolebinding.yaml"
    }
  },
  {
    "name": "credential-storage-key-encryption-key-secret.yaml",
    "path": "deploy/helm/openshell/templates/credential-storage-key-encryption-key-secret.yaml",
    "sha": "1e53d84bfb4e50e117e131f29ba2ba72540df821",
    "size": 1313,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/credential-storage-key-encryption-key-secret.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/credential-storage-key-encryption-key-secret.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1e53d84bfb4e50e117e131f29ba2ba72540df821",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/credential-storage-key-encryption-key-secret.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/credential-storage-key-encryption-key-secret.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1e53d84bfb4e50e117e131f29ba2ba72540df821",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/credential-storage-key-encryption-key-secret.yaml"
    }
  },
  {
    "name": "deployment.yaml",
    "path": "deploy/helm/openshell/templates/deployment.yaml",
    "sha": "f94900b1363d246471a3c0e3974945536709740b",
    "size": 626,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/deployment.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/deployment.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f94900b1363d246471a3c0e3974945536709740b",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/deployment.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/deployment.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f94900b1363d246471a3c0e3974945536709740b",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/deployment.yaml"
    }
  },
  {
    "name": "gateway-config.yaml",
    "path": "deploy/helm/openshell/templates/gateway-config.yaml",
    "sha": "8c50b4053cd3c60acad8e942572c8ed689228d13",
    "size": 14603,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/gateway-config.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/gateway-config.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8c50b4053cd3c60acad8e942572c8ed689228d13",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/gateway-config.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/gateway-config.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/8c50b4053cd3c60acad8e942572c8ed689228d13",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/gateway-config.yaml"
    }
  },
  {
    "name": "gateway.yaml",
    "path": "deploy/helm/openshell/templates/gateway.yaml",
    "sha": "73c3d7603748b2c59811482e7729083d38bbd347",
    "size": 1727,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/gateway.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/gateway.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/73c3d7603748b2c59811482e7729083d38bbd347",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/gateway.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/gateway.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/73c3d7603748b2c59811482e7729083d38bbd347",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/gateway.yaml"
    }
  },
  {
    "name": "grpcroute.yaml",
    "path": "deploy/helm/openshell/templates/grpcroute.yaml",
    "sha": "362067fda3df9c804dc871ec4bee1a3e4aa34194",
    "size": 845,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/grpcroute.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/grpcroute.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/362067fda3df9c804dc871ec4bee1a3e4aa34194",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/grpcroute.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/grpcroute.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/362067fda3df9c804dc871ec4bee1a3e4aa34194",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/grpcroute.yaml"
    }
  },
  {
    "name": "network-policy-ack.yaml",
    "path": "deploy/helm/openshell/templates/network-policy-ack.yaml",
    "sha": "780d4314c9a47beb520bf1d72e20c6bcfb627b74",
    "size": 405,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/network-policy-ack.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/network-policy-ack.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/780d4314c9a47beb520bf1d72e20c6bcfb627b74",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/network-policy-ack.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/network-policy-ack.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/780d4314c9a47beb520bf1d72e20c6bcfb627b74",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/network-policy-ack.yaml"
    }
  },
  {
    "name": "networkpolicy.yaml",
    "path": "deploy/helm/openshell/templates/networkpolicy.yaml",
    "sha": "c9e4a760e30867a80c7418e385f2414e88854cdf",
    "size": 1326,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/networkpolicy.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/networkpolicy.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/c9e4a760e30867a80c7418e385f2414e88854cdf",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/networkpolicy.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/networkpolicy.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/c9e4a760e30867a80c7418e385f2414e88854cdf",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/networkpolicy.yaml"
    }
  },
  {
    "name": "peer-role.yaml",
    "path": "deploy/helm/openshell/templates/peer-role.yaml",
    "sha": "59ae7479addf7d37f36bf6fa7a007301778c1fb0",
    "size": 1092,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/peer-role.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/peer-role.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/59ae7479addf7d37f36bf6fa7a007301778c1fb0",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/peer-role.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/peer-role.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/59ae7479addf7d37f36bf6fa7a007301778c1fb0",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/peer-role.yaml"
    }
  },
  {
    "name": "peer-service.yaml",
    "path": "deploy/helm/openshell/templates/peer-service.yaml",
    "sha": "f5d93af245807860a4f1da9222f08931ff57acb4",
    "size": 556,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/peer-service.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/peer-service.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f5d93af245807860a4f1da9222f08931ff57acb4",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/peer-service.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/peer-service.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/f5d93af245807860a4f1da9222f08931ff57acb4",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/peer-service.yaml"
    }
  },
  {
    "name": "role.yaml",
    "path": "deploy/helm/openshell/templates/role.yaml",
    "sha": "d43f5de0ab02c6facc66be3937e2a17b75b46594",
    "size": 2111,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/role.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/role.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d43f5de0ab02c6facc66be3937e2a17b75b46594",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/role.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/role.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/d43f5de0ab02c6facc66be3937e2a17b75b46594",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/role.yaml"
    }
  },
  {
    "name": "rolebinding.yaml",
    "path": "deploy/helm/openshell/templates/rolebinding.yaml",
    "sha": "32f11644bf02856344680594801153cf83a0dda0",
    "size": 826,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/rolebinding.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/rolebinding.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/32f11644bf02856344680594801153cf83a0dda0",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/rolebinding.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/rolebinding.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/32f11644bf02856344680594801153cf83a0dda0",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/rolebinding.yaml"
    }
  },
  {
    "name": "route.yaml",
    "path": "deploy/helm/openshell/templates/route.yaml",
    "sha": "459a0ac46286b3a88c1eb9d4d5bf17d3b6f8ab47",
    "size": 2407,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/route.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/route.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/459a0ac46286b3a88c1eb9d4d5bf17d3b6f8ab47",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/route.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/route.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/459a0ac46286b3a88c1eb9d4d5bf17d3b6f8ab47",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/route.yaml"
    }
  },
  {
    "name": "service.yaml",
    "path": "deploy/helm/openshell/templates/service.yaml",
    "sha": "ebad42eab15a76ec20fcb5003ac9747642ef04e1",
    "size": 849,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/service.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/service.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ebad42eab15a76ec20fcb5003ac9747642ef04e1",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/service.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/service.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/ebad42eab15a76ec20fcb5003ac9747642ef04e1",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/service.yaml"
    }
  },
  {
    "name": "serviceaccount.yaml",
    "path": "deploy/helm/openshell/templates/serviceaccount.yaml",
    "sha": "1a9245d4dcadfb1a725f1fac0c6ae5b09e1dc1c9",
    "size": 1068,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/serviceaccount.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/serviceaccount.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1a9245d4dcadfb1a725f1fac0c6ae5b09e1dc1c9",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/serviceaccount.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/serviceaccount.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/1a9245d4dcadfb1a725f1fac0c6ae5b09e1dc1c9",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/serviceaccount.yaml"
    }
  },
  {
    "name": "statefulset.yaml",
    "path": "deploy/helm/openshell/templates/statefulset.yaml",
    "sha": "10d0839f6030ed378b81632805870ac352afd5da",
    "size": 871,
    "url": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/statefulset.yaml?ref=main",
    "html_url": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/statefulset.yaml",
    "git_url": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/10d0839f6030ed378b81632805870ac352afd5da",
    "download_url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/statefulset.yaml",
    "type": "file",
    "_links": {
      "self": "https://api.github.com/repos/NVIDIA/OpenShell/contents/deploy/helm/openshell/templates/statefulset.yaml?ref=main",
      "git": "https://api.github.com/repos/NVIDIA/OpenShell/git/blobs/10d0839f6030ed378b81632805870ac352afd5da",
      "html": "https://github.com/NVIDIA/OpenShell/blob/main/deploy/helm/openshell/templates/statefulset.yaml"
    }
  }
]

```

---

## Assistant (Build · GPT-5.6 Terra · 3.7s)

_Thinking:_



**Tool: webfetch**

**Input:**
```json
{
  "url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/cert-manager-pki.yaml",
  "format": "text"
}
```

**Output:**
```
{{- if .Values.certManager.enabled }}
# SPDX-FileCopyrightText: Copyright (c) 2025-2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: {{ include "openshell.issuerSelfSigned" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "openshell.labels" . | nindent 4 }}
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ include "openshell.fullname" . }}-ca
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "openshell.labels" . | nindent 4 }}
spec:
  isCA: true
  commonName: openshell-ca
  secretName: {{ .Values.certManager.caSecretName | quote }}
  privateKey:
    algorithm: ECDSA
    size: 256
  issuerRef:
    name: {{ include "openshell.issuerSelfSigned" . }}
    kind: Issuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: {{ include "openshell.fullname" . }}-ca-issuer
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "openshell.labels" . | nindent 4 }}
spec:
  ca:
    secretName: {{ .Values.certManager.caSecretName | quote }}
---
{{- $externalServerIssuer := .Values.certManager.serverIssuerRef.name }}
{{- if and (not .Values.certManager.clientCaFromServerTlsSecret) (eq .Values.server.tls.clientCaSecretName "openshell-server-client-ca") }}
{{- fail "certManager.clientCaFromServerTlsSecret is false but server.tls.clientCaSecretName is still the default (openshell-server-client-ca), which nothing creates when cert-manager owns TLS. Set server.tls.clientCaSecretName to the secret containing the client CA (e.g. the value of certManager.caSecretName, which defaults to openshell-ca-tls), or set it to empty to disable mTLS client verification." }}
{{- end }}
{{- if $externalServerIssuer }}
{{- if not .Values.certManager.serverDnsNames }}
{{- fail "certManager.serverIssuerRef.name is set but certManager.serverDnsNames is empty — the external certificate requires at least one externally-resolvable DNS name." }}
{{- end }}
{{- range .Values.certManager.serverDnsNames }}
{{- /* Single-label names (e.g. "openshell") are also rejected by ACME CAs but are intentionally not checked here — the guard targets recognisable internal-network patterns. */ -}}
{{- if or (eq . "localhost") (hasSuffix ".localhost" .) (hasSuffix ".svc.cluster.local" .) (hasSuffix ".svc" .) (eq . "host.docker.internal") (eq . "host.containers.internal") }}
{{- fail (printf "certManager.serverIssuerRef.name is set (external issuer) but certManager.serverDnsNames contains %q — external CAs (e.g. ACME / Let's Encrypt) reject internal-only names per CA/Browser Forum baseline requirements. Override certManager.serverDnsNames with your externally-resolvable hostname(s)." .) }}
{{- end }}
{{- end }}
{{- end }}
# Internal server certificate — always issued by the chart’s own CA with
# internal SANs.  Supervisors connect via internal hostnames and verify
# this cert against the chart CA they already trust.
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ include "openshell.fullname" . }}-server
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "openshell.labels" . | nindent 4 }}
spec:
  secretName: {{ .Values.server.tls.certSecretName | quote }}
  duration: {{ .Values.certManager.certificateDuration | quote }}
  renewBefore: {{ .Values.certManager.certificateRenewBefore | quote }}
  commonName: {{ include "openshell.fullname" . }}
  dnsNames:
    {{- range (include "openshell.defaultServerDnsNames" . | fromYamlArray) }}
    - {{ . | quote }}
    {{- end }}
    {{- if not $externalServerIssuer }}
    {{- range .Values.certManager.serverDnsNames }}
    - {{ . | quote }}
    {{- end }}
    {{- end }}
  {{- if .Values.certManager.serverIpAddresses }}
  ipAddresses:
    {{- toYaml .Values.certManager.serverIpAddresses | nindent 4 }}
  {{- end }}
  privateKey:
    algorithm: ECDSA
    size: 256
  usages:
    - server auth
    - digital signature
    - key encipherment
  issuerRef:
    name: {{ include "openshell.fullname" . }}-ca-issuer
    kind: Issuer
    group: cert-manager.io
{{- if $externalServerIssuer }}
---
# External server certificate — issued by the operator-configured issuer
# (e.g. ACME/Let’s Encrypt) with only externally-resolvable SANs.
# The gateway uses SNI to present this cert for external hostnames.
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ include "openshell.fullname" . }}-server-external
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "openshell.labels" . | nindent 4 }}
spec:
  secretName: {{ include "openshell.fullname" . }}-server-external-tls
  duration: {{ .Values.certManager.certificateDuration | quote }}
  renewBefore: {{ .Values.certManager.certificateRenewBefore | quote }}
  {{- if .Values.certManager.serverDnsNames }}
  commonName: {{ first .Values.certManager.serverDnsNames | quote }}
  {{- end }}
  dnsNames:
    {{- range .Values.certManager.serverDnsNames }}
    - {{ . | quote }}
    {{- end }}
  privateKey:
    algorithm: ECDSA
    size: 256
  usages:
    - server auth
    - digital signature
    - key encipherment
  issuerRef:
    name: {{ .Values.certManager.serverIssuerRef.name }}
    kind: {{ .Values.certManager.serverIssuerRef.kind | default "Issuer" }}
    group: {{ .Values.certManager.serverIssuerRef.group | default "cert-manager.io" }}
{{- end }}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ include "openshell.fullname" . }}-client
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "openshell.labels" . | nindent 4 }}
spec:
  secretName: {{ .Values.server.tls.clientTlsSecretName | quote }}
  duration: {{ .Values.certManager.certificateDuration | quote }}
  renewBefore: {{ .Values.certManager.certificateRenewBefore | quote }}
  commonName: openshell-client
  privateKey:
    algorithm: ECDSA
    size: 256
  usages:
    - client auth
    - digital signature
    - key encipherment
  issuerRef:
    name: {{ include "openshell.fullname" . }}-ca-issuer
    kind: Issuer
    group: cert-manager.io
{{- end }}

```

---

## Assistant (Build · GPT-5.6 Terra · 31.2s)

_Thinking:_

**Considering client certificate issuance**

I’m thinking we can create a client-auth CA externally, which has the public CA in the secret in OpenShell. It would involve configuring the gateway to trust it. Issuing the client cert via the `ClusterIssuer` seems like a feasible route. But, wait, the Backend for Frontend (BFF) needs a client cert from external issuance, which makes me wonder if custom CA creation is possible, given that their deployment chart root is temporary, lasting about 90 days.

_Thinking:_



_Thinking:_



This confirms exact chain:

```text
openshell-selfsigned
  -> openshell-ca (root certificate)
  -> openshell-ca-tls (root public cert + private key)
  -> openshell-ca-issuer (CA Issuer)
  -> openshell-server-tls / openshell-client-tls
```

Use `openshell-ca-issuer` for BFF client certificate. **Do not** use `openshell-selfsigned`; that creates self-signed leaf cert, rejected by gateway.

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: agent-ops-bff-client
  namespace: openshell
spec:
  secretName: agent-ops-bff-client-tls
  commonName: agent-ops-bff
  duration: 2160h
  renewBefore: 360h
  privateKey:
    algorithm: ECDSA
    size: 256
  usages:
    - client auth
    - digital signature
    - key encipherment
  issuerRef:
    name: openshell-ca-issuer
    kind: Issuer
    group: cert-manager.io
```

This certificate is trusted: chart gateway defaults to trusting client certificates signed by this CA.

BFF needs only:

```text
CA:         openshell-ca-tls/tls.crt
Client cert: agent-ops-bff-client-tls/tls.crt
Client key:  agent-ops-bff-client-tls/tls.key
```

Never mount `openshell-ca-tls/tls.key`. That is root CA private key.

`Issuer` and `Certificate` secrets are namespace-scoped. If BFF runs outside `openshell` namespace:

1. Create BFF leaf certificate in `openshell`, signed by `openshell-ca-issuer`.
2. Replicate only `agent-ops-bff-client-tls` to BFF namespace through controlled GitOps/External Secrets/secret reflector.
3. Replicate CA public certificate separately, or mount only `tls.crt` from a controlled trust Secret.
4. BFF ServiceAccount gets no API permission to read either secret; kubelet mounts them read-only.

Your current `UnknownCA` is expected until BFF trusts `openshell-ca-tls/tls.crt`. It means BFF rejected gateway server certificate. It is not a missing client-cert error.

One important lifecycle risk: root CA expires/renews in December, with renewal scheduled November 16. Ensure CA trust bundle and BFF client leaf update before then; current Go SDK clients load TLS PEM at client creation, so BFF must restart/recreate gateway clients after certificate rotation.

---

## User

the bff will not in the same namespace. It will be in redhat-ods-applications. Furthermore, multiple gateways will be in different namespaces. is it not enough to just have the bff have its own certs? They are all under cert manager anyways too

---

## Assistant (Build · GPT-5.6 Terra · 27.5s)

_Thinking:_



_Thinking:_



_Thinking:_



Cert-manager alone not shared trust. Check chart client-CA override, then map scalable multi-gateway design.

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/gateway-config.yaml",
  "format": "text"
}
```

**Output:**
```
# SPDX-FileCopyrightText: Copyright (c) 2025-2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
{{/*
ConfigMap holding the gateway TOML config file (RFC 0003).

The gateway reads `/etc/openshell/gateway.toml` (mounted from this ConfigMap)
at startup. CLI flags and OPENSHELL_* env vars on the gateway workload container
still override anything in this file.

One value is intentionally NOT rendered here:
  - server.dbUrl              → passed via OPENSHELL_DB_URL env var (from Secret)
                                 when server.externalDbSecret is set, otherwise
                                 --db-url arg for SQLite
*/}}
{{- $credentialDrivers := list -}}
{{- $otlp := .Values.server.otlp | default dict -}}
{{- if .Values.server.credentialDrivers.kubernetesSecrets.enabled -}}
{{- $credentialDrivers = append $credentialDrivers "kubernetes-secrets" -}}
{{- end -}}
{{- if .Values.server.credentialDrivers.vault.enabled -}}
{{- $credentialDrivers = append $credentialDrivers "vault" -}}
{{- end -}}
{{- if and .Values.certManager.serverIssuerRef.name (not .Values.certManager.enabled) }}
{{- fail "certManager.serverIssuerRef.name is set but certManager.enabled is false \u2014 the external server certificate, its Secret mount, and the gateway TLS configuration all require cert-manager to be enabled. Set certManager.enabled=true or remove certManager.serverIssuerRef.name." }}
{{- end }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "openshell.fullname" . }}-config
  labels:
    {{- include "openshell.labels" . | nindent 4 }}
data:
  gateway.toml: |
    [openshell]
    version = 2

    [openshell.gateway]
    name                  = {{ .Values.server.name | default (include "openshell.fullname" .) | quote }}
    bind_address          = "0.0.0.0:{{ .Values.service.port }}"
    {{- if .Values.service.healthPort }}
    health_bind_address   = "0.0.0.0:{{ .Values.service.healthPort }}"
    {{- end }}
    {{- if .Values.service.metricsPort }}
    metrics_bind_address  = "0.0.0.0:{{ .Values.service.metricsPort }}"
    {{- end }}
    log_level             = {{ .Values.server.logLevel | quote }}
    compute_driver        = "kubernetes"
    {{- if $credentialDrivers }}
    credential_drivers    = [{{- range $i, $driver := $credentialDrivers }}{{ if $i }}, {{ end }}{{ $driver | quote }}{{- end }}]
    {{- end }}
    {{- $policyValidationFailureMode := .Values.server.policyValidationFailureMode }}
    {{- if not (has $policyValidationFailureMode (list "fail_closed" "retain_last_valid")) }}
    {{- fail "server.policyValidationFailureMode must be fail_closed or retain_last_valid" }}
    {{- end }}
    policy_validation_failure_mode = {{ $policyValidationFailureMode | quote }}
    {{- if .Values.server.disableTls }}
    disable_tls            = true
    {{- end }}
    enable_loopback_service_http = {{ .Values.server.enableLoopbackServiceHttp }}
    {{- $sans := list -}}
    {{- if and .Values.certManager.enabled .Values.certManager.serverDnsNames }}
    {{- $sans = .Values.certManager.serverDnsNames }}
    {{- else if and .Values.pkiInitJob.enabled .Values.pkiInitJob.serverDnsNames }}
    {{- $sans = .Values.pkiInitJob.serverDnsNames }}
    {{- end }}
    {{- if $sans }}
    server_sans = [{{- range $i, $san := $sans }}{{ if $i }}, {{ end }}{{ $san | quote }}{{- end }}]
    {{- end }}
    {{- $rlRequests := int .Values.server.grpcRateLimit.requests }}
    {{- $rlWindowSeconds := int .Values.server.grpcRateLimit.windowSeconds }}
    {{- if or (lt $rlRequests 0) (lt $rlWindowSeconds 0) }}
    {{- fail "server.grpcRateLimit.requests and server.grpcRateLimit.windowSeconds must not be negative; they map to unsigned gateway settings" }}
    {{- end }}
    {{- if and (gt $rlRequests 0) (gt $rlWindowSeconds 0) }}
    grpc_rate_limit_requests       = {{ $rlRequests }}
    grpc_rate_limit_window_seconds = {{ $rlWindowSeconds }}
    {{- else if or (gt $rlRequests 0) (gt $rlWindowSeconds 0) }}
    {{- fail "server.grpcRateLimit requires both requests and windowSeconds to be positive to enable rate limiting, or both 0/unset to disable it" }}
    {{- end }}

    {{- if $otlp.endpoint }}

    [openshell.gateway.otlp]
    endpoint = {{ $otlp.endpoint | quote }}
    {{- if $otlp.serviceName }}
    service_name = {{ $otlp.serviceName | quote }}
    {{- end }}
    {{- end }}

    {{- if not .Values.server.disableTls }}

    [openshell.gateway.tls]
    cert_path             = "/etc/openshell-tls/server/tls.crt"
    key_path              = "/etc/openshell-tls/server/tls.key"
    {{- if eq (include "openshell.gatewayClientCaEnabled" .) "true" }}
    client_ca_path        = "/etc/openshell-tls/client-ca/ca.crt"
    {{- end }}
    {{- if .Values.certManager.serverIssuerRef.name }}
    external_cert_path    = "/etc/openshell-tls/server-external/tls.crt"
    external_key_path     = "/etc/openshell-tls/server-external/tls.key"
    external_server_names = [{{- range $i, $name := .Values.certManager.serverDnsNames }}{{ if $i }}, {{ end }}{{ $name | quote }}{{- end }}]
    {{- end }}
    {{- end }}

    {{- if .Values.server.auth.allowUnauthenticatedUsers }}

    [openshell.gateway.auth]
    allow_unauthenticated_users = true
    {{- end }}

    [openshell.gateway.gateway_jwt]
    signing_key_path = "/etc/openshell-jwt/signing.pem"
    public_key_path  = "/etc/openshell-jwt/public.pem"
    kid_path         = "/etc/openshell-jwt/kid"
    gateway_id       = {{ .Values.server.sandboxJwt.gatewayId | default (include "openshell.fullname" .) | quote }}
    ttl_secs         = {{ .Values.server.sandboxJwt.ttlSecs | default 3600 }}

    {{- if .Values.server.oidc.issuer }}

    [openshell.gateway.oidc]
    issuer                           = {{ .Values.server.oidc.issuer | quote }}
    dangerously_allow_insecure_http = {{ .Values.server.oidc.dangerouslyAllowInsecureHttp }}
    jwks_allowed_origins             = {{ .Values.server.oidc.jwksAllowedOrigins | toJson }}
    audience                         = {{ .Values.server.oidc.audience | quote }}
    jwks_ttl_secs                    = {{ .Values.server.oidc.jwksTtl }}
    {{- if .Values.server.oidc.rolesClaim }}
    roles_claim   = {{ .Values.server.oidc.rolesClaim | quote }}
    {{- end }}
    {{- if .Values.server.oidc.adminRole }}
    admin_role    = {{ .Values.server.oidc.adminRole | quote }}
    {{- end }}
    {{- if .Values.server.oidc.userRole }}
    user_role     = {{ .Values.server.oidc.userRole | quote }}
    {{- end }}
    {{- if .Values.server.oidc.scopesClaim }}
    scopes_claim  = {{ .Values.server.oidc.scopesClaim | quote }}
    {{- end }}
    {{- end }}

    [openshell.drivers.kubernetes]
    allow_driver_config          = {{ .Values.server.drivers.kubernetes.allowDriverConfig }}
    namespace                    = {{ include "openshell.sandboxNamespace" . | quote }}
    default_image                = {{ .Values.server.sandboxImage | quote }}
    {{- if include "openshell.sandboxRuntimeImageOverrideEnabled" . }}
    sandbox_runtime_image        = {{ include "openshell.sandboxRuntimeImage" . | quote }}
    {{- end }}
    {{- if include "openshell.supervisorImageOverrideEnabled" . }}
    supervisor_image             = {{ include "openshell.supervisorImage" . | quote }}
    {{- end }}
    {{- if .Values.server.hostGatewayIP }}
    host_gateway_ip              = {{ .Values.server.hostGatewayIP | quote }}
    {{- end }}
    {{- if not .Values.server.disableTls }}
    client_tls_secret_name       = {{ .Values.server.tls.clientTlsSecretName | quote }}
    {{- end }}
    workspace_mode               = {{ .Values.server.drivers.kubernetes.workspaceMode | default "shared" | quote }}
    gateway_id                   = {{ .Values.server.sandboxJwt.gatewayId | default (include "openshell.fullname" .) | quote }}
    grpc_endpoint                = {{ include "openshell.grpcEndpoint" . | quote }}
    service_account_name         = {{ include "openshell.sandboxServiceAccountName" . | quote }}
    {{- if .Values.server.enableUserNamespaces }}
    enable_user_namespaces       = true
    {{- end }}
    {{- if .Values.server.drivers.kubernetes.operatorNamespaceLabel }}
    operator_namespace_label     = {{ .Values.server.drivers.kubernetes.operatorNamespaceLabel | quote }}
    {{- end }}
    {{- if .Values.server.drivers.kubernetes.operatorNamespaceFile }}
    operator_namespace_file      = {{ .Values.server.drivers.kubernetes.operatorNamespaceFile | quote }}
    {{- end }}
    sa_token_ttl_secs            = {{ .Values.server.sandboxJwt.k8sSaTokenTtlSecs | default 3600 }}
    {{- if .Values.upstreamProxy.url }}
    https_proxy                  = {{ .Values.upstreamProxy.url | quote }}
    {{- end }}
    {{- if .Values.upstreamProxy.noProxy }}
    no_proxy                     = {{ .Values.upstreamProxy.noProxy | quote }}
    {{- end }}
    {{- if .Values.upstreamProxy.authSecret.name }}
    proxy_auth_secret_name       = {{ .Values.upstreamProxy.authSecret.name | quote }}
    {{- end }}
    {{- if .Values.upstreamProxy.authSecret.key }}
    proxy_auth_secret_key        = {{ .Values.upstreamProxy.authSecret.key | quote }}
    {{- end }}
    {{- if and .Values.upstreamProxy.authSecret.name .Values.upstreamProxy.authSecret.key }}
    proxy_auth_allow_insecure    = {{ .Values.upstreamProxy.authAllowInsecure }}
    {{- end }}
    {{- if .Values.upstreamProxy.connectByHostname }}
    proxy_connect_by_hostname    = true
    {{- end }}
    {{- if .Values.upstreamProxy.caBundle.configMapName }}
    proxy_ca_bundle              = "/etc/openshell-tls/proxy-ca/ca.crt"
    {{- end }}
    {{- if .Values.server.providerTokenGrants.spiffe.enabled }}
    provider_spiffe_workload_api_socket_path = {{ .Values.server.providerTokenGrants.spiffe.workloadApiSocketPath | quote }}
    {{- end }}
    {{- if .Values.server.sandboxImagePullPolicy }}
    image_pull_policy            = {{ include "openshell.canonicalImagePullPolicy" .Values.server.sandboxImagePullPolicy | quote }}
    {{- end }}
    {{- $sandboxImagePullSecretNames := list -}}
    {{- range .Values.server.sandboxImagePullSecrets }}
    {{- if .name }}
    {{- $sandboxImagePullSecretNames = append $sandboxImagePullSecretNames .name }}
    {{- end }}
    {{- end }}
    {{- if $sandboxImagePullSecretNames }}
    image_pull_secrets           = [{{- range $i, $name := $sandboxImagePullSecretNames }}{{ if $i }}, {{ end }}{{ $name | quote }}{{- end }}]
    {{- end }}
    {{- if .Values.server.workspaceDefaultStorageSize }}
    workspace_default_storage_size = {{ .Values.server.workspaceDefaultStorageSize | quote }}
    {{- end }}
    {{- if .Values.server.workspaceStorageClass }}
    workspace_storage_class      = {{ .Values.server.workspaceStorageClass | quote }}
    {{- end }}
    {{- if .Values.server.defaultRuntimeClassName }}
    default_runtime_class_name   = {{ .Values.server.defaultRuntimeClassName | quote }}
    {{- end }}
    {{- if .Values.supervisor.image.pullPolicy }}
    supervisor_image_pull_policy = {{ include "openshell.canonicalImagePullPolicy" .Values.supervisor.image.pullPolicy | quote }}
    {{- end }}
    {{- if .Values.sandboxRuntime.image.pullPolicy }}
    sandbox_runtime_image_pull_policy = {{ include "openshell.canonicalImagePullPolicy" .Values.sandboxRuntime.image.pullPolicy | quote }}
    {{- end }}

    [openshell.drivers.kubernetes.resource_admission]
    enabled = {{ .Values.server.drivers.kubernetes.resourceAdmission.enabled }}
    {{- if ne .Values.server.drivers.kubernetes.resourceAdmission.requiredLabels nil }}
    [openshell.drivers.kubernetes.resource_admission.required_labels]
    {{- range $key, $value := .Values.server.drivers.kubernetes.resourceAdmission.requiredLabels }}
    {{ $key | quote }} = {{ $value | quote }}
    {{- end }}
    {{- end }}

    [openshell.drivers.kubernetes.managed_ssh_ingress]
    enabled                       = {{ .Values.networkPolicy.enabled }}
    gateway_namespace             = {{ .Release.Namespace | quote }}
    gateway_pod_selector          = { "app.kubernetes.io/name" = {{ include "openshell.name" . | quote }}, "app.kubernetes.io/instance" = {{ .Release.Name | quote }} }

    [openshell.drivers.kubernetes.sandbox_runtime]
    network_policy_enforced      = {{ .Values.supervisor.sandboxRuntime.networkPolicyEnforced }}
    boundary_port                = {{ .Values.supervisor.sandboxRuntime.boundaryPort | default 5500 }}

    {{- if not $credentialDrivers }}

    [openshell.gateway.credential_storage]
    key_encryption_key_env = {{ include "openshell.credentialStorageKeyEncryptionKeyEnvName" . | quote }}
    {{- end }}

    {{- if .Values.server.credentialDrivers.kubernetesSecrets.enabled }}

    [openshell.credential_drivers.kubernetes-secrets]
    namespace = {{ include "openshell.credentialKubernetesSecretsNamespace" . | quote }}
    allow_reference_namespace = {{ .Values.server.credentialDrivers.kubernetesSecrets.allowReferenceNamespace }}
    workspace_mode               = {{ .Values.server.drivers.kubernetes.workspaceMode | default "shared" | quote }}
    gateway_id                   = {{ .Values.server.sandboxJwt.gatewayId | default (include "openshell.fullname" .) | quote }}
    {{- end }}

    {{- if .Values.server.credentialDrivers.vault.enabled }}

    [openshell.credential_drivers.vault]
    address = {{ .Values.server.credentialDrivers.vault.address | quote }}
    {{- if .Values.server.credentialDrivers.vault.caConfigMapName }}
    ca_bundle = "/etc/openshell-tls/vault-ca/ca.crt"
    {{- end }}
    mount = {{ .Values.server.credentialDrivers.vault.mount | quote }}
    kv_version = {{ .Values.server.credentialDrivers.vault.kvVersion | quote }}
    auth_method = {{ .Values.server.credentialDrivers.vault.authMethod | quote }}
    {{- if .Values.server.credentialDrivers.vault.role }}
    role = {{ .Values.server.credentialDrivers.vault.role | quote }}
    {{- end }}
    {{- if .Values.server.credentialDrivers.vault.kubernetesAuthMount }}
    kubernetes_auth_mount = {{ .Values.server.credentialDrivers.vault.kubernetesAuthMount | quote }}
    {{- end }}
    {{- if .Values.server.credentialDrivers.vault.serviceAccountTokenPath }}
    service_account_token_path = {{ .Values.server.credentialDrivers.vault.serviceAccountTokenPath | quote }}
    {{- end }}
    {{- if .Values.server.credentialDrivers.vault.tokenPath }}
    token_path = {{ .Values.server.credentialDrivers.vault.tokenPath | quote }}
    {{- end }}
    {{- if .Values.server.credentialDrivers.vault.timeoutSecs }}
    timeout_secs = {{ .Values.server.credentialDrivers.vault.timeoutSecs }}
    {{- end }}
    {{- end }}

```

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/values.yaml",
  "format": "text"
}
```

**Output:**
```
# SPDX-FileCopyrightText: Copyright (c) 2025-2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# Default values for OpenShell

# -- Number of OpenShell gateway replicas. Values greater than 1 require
# server.externalDbSecret because the default SQLite backend is per pod.
replicaCount: 1

workload:
  # -- Gateway workload controller kind. Use `statefulset` for the default
  # SQLite database, or `deployment` when server.externalDbSecret points at an
  # external database.
  kind: statefulset
  # -- Allow replicaCount > 1 while rendering a StatefulSet. Prefer
  # workload.kind=deployment for external database-backed multi-replica
  # gateways; this override exists for operators who explicitly require
  # StatefulSet identity or storage semantics.
  allowMultiReplicaStatefulSet: false

image:
  # -- Gateway image repository.
  repository: ghcr.io/nvidia/openshell/gateway
  # -- Gateway image pull policy.
  pullPolicy: IfNotPresent
  # -- Gateway image tag. Defaults to the chart appVersion when empty.
  tag: ""

# Trusted workload-side runtime image.
sandboxRuntime:
  image:
    # -- Sandbox runtime image repository. Changing it uses the effective gateway image tag unless tag is also set.
    repository: ghcr.io/nvidia/openshell/sandbox
    # -- Sandbox runtime image pull policy. Defaults to the gateway image pull policy when empty.
    pullPolicy: ""
    # -- Sandbox runtime image tag override. Empty uses the version pinned into the gateway unless repository is changed.
    tag: ""

# Trusted control-side runtime image.
supervisor:
  image:
    # -- Supervisor image repository. Changing it uses the effective gateway image tag unless tag is also set.
    repository: ghcr.io/nvidia/openshell/supervisor
    # -- Sandbox supervisor pull policy. Leave unset to use the Kubernetes
    # image default. Prefer always, if_not_present, or never; the chart also
    # accepts legacy Kubernetes spellings Always, IfNotPresent, and Never.
    pullPolicy: null
    # -- Supervisor image tag override. Empty uses the version pinned into the gateway unless repository is changed.
    tag: ""
  sandboxRuntime:
    # -- Required operator acknowledgement that the cluster CNI enforces NetworkPolicy.
    networkPolicyEnforced: false
    # -- Workload boundary TLS listener port.
    boundaryPort: 5500

# -- Operator-owned corporate forward proxy for policy-approved TLS egress
# from Kubernetes sandboxes. The workload cannot select or override it.
upstreamProxy:
  # -- Proxy URL in http://host:port or https://host:port form. An https:// proxy
  # whose certificate is not publicly trusted also needs caBundle below.
  url: ""
  # -- Comma-separated destinations that bypass only the corporate proxy.
  noProxy: ""
  authSecret:
    # -- Existing Secret in the sandbox namespace containing a user:pass value.
    name: ""
    # -- Secret key containing the proxy credential.
    key: ""
  # -- Required when authSecret is configured because Basic auth to an HTTP proxy is cleartext.
  authAllowInsecure: false
  # -- Last-resort option for hostname-filtering proxy ACLs. It lets the proxy resolve CONNECT targets.
  connectByHostname: false
  caBundle:
    # -- ConfigMap in the release namespace holding the corporate proxy CA bundle.
    # Required for an https:// proxy with a private CA, and for a TLS-intercepting
    # proxy that re-signs upstream certificates. The gateway reads it and stages it
    # into each sandbox's immutable supervisor bootstrap Secret, so the anchor stays
    # in the gateway's trust domain rather than the workload namespace.
    # Supply only the CA that signs your proxy's certificate, or that the proxy
    # re-signs intercepted upstream certificates with. Public roots already come
    # from the supervisor image and its TLS stack, so a full merged trust bundle
    # (for example an OpenShift config.openshift.io/inject-trusted-cabundle
    # ConfigMap) adds hundreds of kilobytes of duplicated roots and can exceed
    # the sandbox boundary's control-frame budget.
    configMapName: ""
    # -- Key inside that ConfigMap. Change this only to reuse an existing
    # ConfigMap whose key is not ca.crt.
    key: ca.crt

# -- Image pull secrets attached to gateway and helper pods.
imagePullSecrets: []
# -- Override the chart name used in generated resource names.
nameOverride: "openshell"
# -- Override the full generated resource name.
fullnameOverride: ""

serviceAccount:
  # -- Create a service account for the gateway.
  create: true
  # -- Annotations to add to the generated service account.
  annotations: {}
  # -- Existing service account name to use when serviceAccount.create is false.
  name: ""

sandboxServiceAccount:
  # -- Create a service account for sandbox pods.
  create: true
  # -- Annotations to add to the generated sandbox service account.
  annotations: {}
  # -- Existing service account name for sandbox pods when sandboxServiceAccount.create is false.
  name: ""

# Namespace-scoped resources needed to run sandboxes. Disable this when the
# gateway and workspace prerequisites are managed as separate Helm releases
# using the openshell-workspace chart.
workspaceResources:
  # -- Create the sandbox ServiceAccount, Role, RoleBinding, and NetworkPolicy
  # from this chart. Disable for a gateway-only release.
  enabled: true

# -- Extra annotations to add to the gateway pod.
podAnnotations: {}
# -- Extra labels to add to the gateway pod.
podLabels: {}

podSecurityContext:
  # -- fsGroup assigned to the gateway pod.
  fsGroup: 1000

securityContext:
  # -- Require the gateway container to run as a non-root user.
  runAsNonRoot: true
  # -- UID assigned to the gateway container.
  runAsUser: 1000
  # -- Whether the gateway container can gain additional privileges.
  allowPrivilegeEscalation: false
  capabilities:
    # -- Linux capabilities dropped from the gateway container.
    drop:
      - ALL

service:
  # -- Kubernetes Service type for the gateway.
  type: ClusterIP
  # -- Gateway gRPC/HTTP service port.
  port: 8080
  # -- Gateway health service port.
  healthPort: 8081
  # -- Gateway metrics service port.
  metricsPort: 9090

# Agent Sandbox is a cluster-scoped prerequisite for the Kubernetes compute
# driver. OpenShell deliberately does not install its CRDs or controller.
# Enable this check for live Helm installs to fail before creating gateway
# resources when neither supported Sandbox API is served. Disable it for
# offline `helm template` and lint workflows, which cannot discover APIs.
agentSandbox:
  preflight:
    # -- Check the live cluster for a supported Agent Sandbox API before rendering gateway resources. Disable only for offline rendering and linting.
    enabled: true

# Pod restart behavior and health probe tuning.
podLifecycle:
  # -- Grace period, in seconds, before Kubernetes terminates the gateway pod.
  terminationGracePeriodSeconds: 5

probes:
  startup:
    # -- Startup probe period, in seconds.
    periodSeconds: 2
    # -- Startup probe timeout, in seconds.
    timeoutSeconds: 1
    # -- Startup probe failure threshold before the container is killed.
    failureThreshold: 30
  liveness:
    # -- Liveness probe initial delay, in seconds.
    initialDelaySeconds: 2
    # -- Liveness probe period, in seconds.
    periodSeconds: 5
    # -- Liveness probe timeout, in seconds.
    timeoutSeconds: 1
    # -- Liveness probe failure threshold before the container is restarted.
    failureThreshold: 3
  readiness:
    # -- Readiness probe initial delay, in seconds.
    initialDelaySeconds: 1
    # -- Readiness probe period, in seconds.
    periodSeconds: 2
    # -- Readiness probe timeout, in seconds.
    timeoutSeconds: 1
    # -- Readiness probe failure threshold before the pod is marked not ready.
    failureThreshold: 3

# -- Gateway pod resource requests and limits.
resources: {}

# -- Node selector for the gateway pod.
nodeSelector: {}

# -- Tolerations for the gateway pod.
tolerations: []

# -- Affinity rules for the gateway pod.
affinity: {}

# Server configuration
server:
  # -- Operator-facing gateway name. Defaults to the chart fullname so all
  # replicas in one installation share an identity. Set explicitly when one
  # telemetry collector receives spans from multiple namespaces or clusters.
  name: ""
  # -- Gateway log level.
  logLevel: info
  # OpenTelemetry trace export over OTLP/gRPC. Leave endpoint empty to disable.
  otlp:
    # -- OTLP/gRPC collector endpoint, conventionally using port 4317.
    endpoint: ""
    # -- Gateway OpenTelemetry service name. Empty uses openshell-gateway.
    serviceName: ""
  # -- Enable anonymous OpenShell telemetry from the gateway and the sandbox
  # supervisors it launches.
  telemetryEnabled: true
  # -- Namespace where sandbox pods are created. Defaults to the Helm release
  # namespace (.Release.Namespace) when left empty.
  sandboxNamespace: ""
  # -- Gateway database URL (used for the default SQLite backend).
  dbUrl: "sqlite:/var/openshell/openshell.db"
  # -- Name of a pre-existing Opaque Secret containing a PostgreSQL
  # connection URI (key: uri). When set, the gateway reads OPENSHELL_DB_URL
  # from this Secret instead of using dbUrl. The Secret must contain a
  # `uri` key, e.g. postgresql://user:pass@host:5432/dbname.
  externalDbSecret: ""
  # -- Default sandbox image used when requests do not specify one.
  sandboxImage: "nvcr.io/nvidia/base/ubuntu:24.04"
  # -- Pull policy for sandbox pods. Leave unset to use the Kubernetes image
  # default (Always for :latest, IfNotPresent otherwise). Prefer always,
  # if_not_present, or never; the chart also accepts legacy Kubernetes spellings
  # Always, IfNotPresent, and Never.
  sandboxImagePullPolicy: null
  # -- Image pull secrets attached to sandbox pods. Referenced Secrets must exist
  # in the sandbox namespace.
  sandboxImagePullSecrets: []
  # -- Default storage size for the workspace PVC in sandbox pods.
  # Uses Kubernetes quantity syntax (e.g. "2Gi", "10Gi", "500Mi").
  # Empty = built-in default (2Gi).
  workspaceDefaultStorageSize: ""
  # -- Kubernetes StorageClass for the workspace PVC in sandbox pods.
  # Empty (default) = omit storageClassName, using the cluster's default
  # StorageClass. Set this on clusters with no default StorageClass, otherwise
  # the workspace PVC stays Pending and the sandbox never starts.
  workspaceStorageClass: ""
  # -- Default Kubernetes runtimeClassName for sandbox pods.
  # Applied when a CreateSandbox request does not specify one.
  # Empty (default) = omit the field, using the cluster's default RuntimeClass.
  # Set to a RuntimeClass name (e.g. "kata-containers", "nvidia") to apply it
  # to all sandboxes that don't explicitly override it.
  defaultRuntimeClassName: ""
  # -- gRPC endpoint sandboxes call back into the gateway. Leave empty to derive
  # it from the chart fullname, release namespace, service port, and
  # disableTls flag, for example https://openshell.openshell.svc.cluster.local:8080.
  # Override only when sandboxes must reach the gateway via a different
  # hostname (e.g. an external ingress or a host alias).
  grpcEndpoint: ""
  # The gateway terminates TLS directly. Its client CA authenticates sandbox
  # callbacks; OIDC-enabled listeners permit bearer-only user clients while
  # still validating any client certificate they present.
  # -- Host gateway IP for sandbox pod hostAliases. When set, sandbox pods get
  # hostAliases entries mapping host.docker.internal and host.openshell.internal
  # to this IP, allowing them to reach services running on the Docker host.
  # Auto-detected by the cluster entrypoint script.
  hostGatewayIP: ""
  # -- Enable Kubernetes user namespace isolation (hostUsers: false) for sandbox
  # pods. Requires Kubernetes 1.33+ with user namespace support available
  # (beta through 1.35, GA in 1.36+), plus a supporting container runtime and
  # Linux 5.12+. When enabled, container UID 0 maps to an unprivileged host
  # UID and capabilities become namespaced.
  enableUserNamespaces: false
  # Kubernetes compute driver settings.
  drivers:
    kubernetes:
      # -- Allow caller driver JSON; external resources still require approval.
      allowDriverConfig: false
      resourceAdmission:
        # -- Require operator approval labels on external sandbox attachments (GPU attachments exempt).
        enabled: true
        # -- Replacement label map; null uses the built-in admission labels. Empty map is invalid when enabled.
        requiredLabels: null
      # -- How workspaces map to Kubernetes namespaces.
      # "shared" (default): all sandboxes in a single namespace.
      # "managed": auto-creates per-workspace namespaces.
      # "operator": uses pre-provisioned namespaces.
      workspaceMode: "shared"
      # -- K8s label selector for namespace discovery in operator mode.
      # The driver watches namespaces matching this label.
      operatorNamespaceLabel: ""
      # -- Path to a JSON file containing an array of namespace names
      # allowed in operator mode. Hot-reloaded on change.
      operatorNamespaceFile: ""
  # -- Disable TLS entirely - the server listens on plaintext HTTP.
  # Set to true when a reverse proxy / tunnel terminates TLS at the edge.
  disableTls: false
  # -- Enable plaintext HTTP routing for loopback sandbox service URLs on
  # TLS-enabled gateways.
  enableLoopbackServiceHttp: true
  # -- Posture when a candidate sandbox policy fails validation. `fail_closed`
  # deactivates the previous policy; `retain_last_valid` keeps it active.
  policyValidationFailureMode: fail_closed
  # Optional gateway-wide gRPC request rate limit. Applies only to gRPC API
  # traffic after protocol multiplexing; health, metrics, and loopback service
  # HTTP routes are not rate limited. Both values must be positive to enable the
  # limit, otherwise it is omitted from the rendered config and stays disabled.
  grpcRateLimit:
    # -- Maximum gRPC requests allowed per window. Must be positive (alongside
    # windowSeconds) to enable rate limiting; 0 (default) disables it.
    requests: 0
    # -- gRPC rate-limit window length in seconds. Must be positive (alongside
    # requests) to enable rate limiting; 0 (default) disables it.
    windowSeconds: 0
  # Default credential storage settings (used when no credential driver is
  # enabled). The gateway encrypts provider credentials in the database using
  # AES-256-GCM with a key-encryption key (KEK). By default, the Helm chart
  # generates and retains a KEK Secret. For GitOps / helm-template workflows
  # where `lookup` is unavailable, reference a pre-created Secret instead.
  credentialStorage:
    # -- Name of a pre-existing Secret containing the key-encryption key.
    # When set, the chart does NOT generate a new Secret; it references this
    # one instead. The Secret must contain a key named "key-encryption-key"
    # with a base64-encoded 32-byte value. Required for GitOps workflows that
    # render manifests with `helm template` (where `lookup` is unavailable).
    existingSecret: ""
  # Provider credential drivers store provider credential secret material in an
  # external or native backend. When no driver is enabled, the gateway uses its
  # default encrypted database credential storage with a retained Kubernetes
  # Secret for the shared key-encryption key.
  credentialDrivers:
    kubernetesSecrets:
      # -- Enable the in-tree Kubernetes Secret credential driver.
      # WARNING: The RBAC Role grants read/write access to ALL Secrets in the
      # configured namespace. Use a dedicated namespace to limit blast radius.
      enabled: false
      # -- Namespace where OpenShell-managed provider Secret objects are stored.
      # Empty = Helm release namespace. A dedicated namespace is RECOMMENDED
      # to isolate OpenShell-managed Secrets from other workloads.
      namespace: ""
      # -- Deprecated compatibility field. Credential storage no longer supports user-authored namespace references.
      allowReferenceNamespace: false
      rbac:
        # -- Create a Role/RoleBinding granting the gateway ServiceAccount read/write access to managed provider Secrets.
        create: true
    vault:
      # -- Enable the in-tree Vault credential driver.
      enabled: false
      # -- Vault service base URL. Non-loopback endpoints must use HTTPS, for example https://vault.vault.svc.cluster.local:8200.
      address: ""
      # -- ConfigMap containing the private Vault CA certificate bundle in the ca.crt key. Leave empty to use platform trust roots.
      caConfigMapName: ""
      # -- Default KV mount name.
      mount: secret
      # -- Default KV engine version. Use "1" or "2".
      kvVersion: "2"
      # -- Authentication method. Use "kubernetes" in-cluster or "token_file" for local/dev validation.
      authMethod: kubernetes
      # -- Vault Kubernetes auth role when authMethod is kubernetes.
      role: ""
      # -- Vault Kubernetes auth mount.
      kubernetesAuthMount: kubernetes
      # -- ServiceAccount token path used for Kubernetes auth.
      serviceAccountTokenPath: /var/run/secrets/kubernetes.io/serviceaccount/token
      # -- Mounted token file path when authMethod is token_file.
      tokenPath: ""
      # -- HTTP request timeout in seconds. Empty = driver default.
      timeoutSecs: ""
  auth:
    # -- UNSAFE: accept unauthenticated CLI/user requests as a local developer
    # principal. Intended only for trusted local Skaffold/k3d development or a
    # fully trusted fronting proxy. Leave false for shared or production clusters.
    allowUnauthenticatedUsers: false
  tls:
    # -- K8s secret (type kubernetes.io/tls) with tls.crt and tls.key for the server.
    certSecretName: openshell-server-tls
    # -- Enable mTLS client certificate authentication. When false, the gateway
    # runs HTTPS-only without requiring client certificates (use OIDC for auth
    # instead). Must be false when using BackendTLSPolicy because ingress
    # proxies cannot present client certificates to the backend.
    enableMtls: true
    # -- K8s secret with ca.crt for client certificate verification (mTLS).
    # Only used when enableMtls is true. Set to "" to disable client certificate
    # verification for HTTPS-only mode.
    clientCaSecretName: openshell-server-client-ca
    # -- K8s secret mounted into sandbox pods for mTLS to the server.
    clientTlsSecretName: openshell-client-tls
  # Gateway-minted sandbox JWT signing keys. The certgen hook generates an
  # Ed25519 keypair and writes it to a secret containing signing.pem (PKCS#8),
  # public.pem (SPKI), and kid (plain text). The hook runs in full PKI mode when
  # pkiInitJob.enabled=true unless certManager.enabled=true, which takes
  # precedence and runs the hook in JWT-only mode.
  sandboxJwt:
    # -- Name of the Opaque Secret holding the signing key material. Empty
    # falls back to the chart fullname with "-jwt-keys" appended.
    signingSecretName: ""
    # -- Stable gateway identity embedded in iss/aud of every minted token.
    # Defaults to the release name so HA replicas share identity.
    gatewayId: ""
    # -- Token TTL in seconds. Defaults to 3600 (1h).
    ttlSecs: 3600
    # -- Lifetime (seconds) of the projected ServiceAccount token kubelet
    # writes into each sandbox pod for the IssueSandboxToken bootstrap
    # exchange. Kubelet enforces a minimum of 600s; the driver clamps
    # values outside [600, 86400]. Default 3600 — generous, since the
    # supervisor consumes the token within seconds of pod start.
    k8sSaTokenTtlSecs: 3600
    # -- File mode for the mounted JWT signing key Secret. Default 0400
    # (owner-read only). Override to 0440 or 0444 if the container UID
    # does not match the volume file owner.
    secretDefaultMode: ""
  # Dynamic provider token grants. When SPIFFE is enabled here, both the
  # gateway and sandbox supervisors mount the SPIFFE Workload API socket so
  # token-exchange profiles can use gateway- and sandbox-scoped JWT-SVIDs.
  # Supervisor-to-gateway authentication still uses gateway-minted sandbox JWTs.
  providerTokenGrants:
    spiffe:
      # -- Mount the SPIFFE Workload API socket into gateway and sandbox pods for dynamic provider token grants.
      enabled: false
      # -- Path to the SPIFFE Workload API socket mounted into gateway and sandbox pods.
      workloadApiSocketPath: /spiffe-workload-api/spire-agent.sock
  # OIDC (OpenID Connect) configuration for JWT-based authentication.
  # When issuer is set, the server validates Bearer tokens on gRPC requests.
  oidc:
    # -- OIDC issuer URL (e.g. https://keycloak.example.com/realms/openshell).
    issuer: ""
    # -- Development only: permit cleartext OIDC requests to numeric loopback
    # addresses. This never permits HTTP to hostnames or non-loopback addresses.
    dangerouslyAllowInsecureHttp: false
    # -- Additional trusted HTTPS origins allowed to serve JWKS. The issuer
    # origin is always allowed. Entries must not include a path or query.
    jwksAllowedOrigins: []
    # -- Expected audience claim for the API resource server.
    # This should match the server's --oidc-audience, NOT the CLI client ID.
    audience: "openshell-cli"
    # -- JWKS key cache TTL in seconds. Must be greater than zero.
    jwksTtl: 3600
    # -- Dot-separated path to the roles array in the JWT claims.
    # Keycloak: "realm_access.roles", Entra ID: "roles", Okta: "groups".
    rolesClaim: ""
    # -- Role name for admin access. Leave empty (with userRole also empty) for
    # authentication-only mode. Both must be set or both empty.
    adminRole: ""
    # -- Role name for standard user access.
    userRole: ""
    # -- Dot-separated path to the scopes array in the JWT claims.
    scopesClaim: ""
    # -- Name of a ConfigMap containing a CA certificate bundle (key: ca.crt)
    # for verifying the OIDC issuer's TLS certificate. Required when the
    # issuer uses a non-public CA (e.g. OpenShift ingress, private PKI).
    caConfigMapName: ""

# NetworkPolicy restricting SSH ingress on sandbox pods to the gateway only.
networkPolicy:
  # -- Restrict SSH ingress on sandbox pods to the gateway. In managed mode,
  # the driver applies the equivalent policy to each workspace namespace.
  enabled: true

# Built-in TLS PKI bootstrap via a pre-install/pre-upgrade hook Job.
# Runs `openshell-gateway generate-certs` to create the server and client TLS
# Secrets in-cluster. Key material is written directly to K8s Secrets and
# never appears in Helm release history. Idempotent: existing secrets are
# left untouched on upgrade. Reuses the gateway image - no extra image to
# mirror in air-gapped environments.
#
# The server certificate already includes the built-in cluster SANs
# (`openshell`, `openshell.openshell.svc`, the cluster.local FQDN, `localhost`,
# `openshell.localhost`, `*.openshell.localhost`, `host.docker.internal`, and
# `127.0.0.1`) baked into the gateway binary. The lists below are additional
# SANs appended on top. Wildcard DNS SANs also enable sandbox service URLs under
# that domain, for example `*.apps.example.com` enables
# `<sandbox>--<service>.apps.example.com`.
pkiInitJob:
  # -- Run a pre-install/pre-upgrade Job that creates gateway and client mTLS
  # Secrets. When certManager.enabled=true, cert-manager owns TLS and this same
  # hook runs in JWT-only mode even if pkiInitJob.enabled remains true.
  enabled: true
  # -- Extra DNS SANs to append to the server certificate.
  serverDnsNames: []
  # -- Extra IP SANs to append to the server certificate.
  serverIpAddresses: []
  # -- Maximum time in seconds for the certgen hook to poll for cert-manager
  # certificates. When using cert-manager with BackendTLSPolicy, the hook
  # polls for this many seconds waiting for the certificate to be issued,
  # then creates the backend CA ConfigMap. The Job deadline is set to
  # (timeoutSeconds + 30) to allow time for ConfigMap creation and cleanup.
  # Increase this if cert-manager takes longer than 120 seconds to issue
  # certificates.
  timeoutSeconds: 120
  # -- Fail the helm install/upgrade if cert-manager does not issue the
  # certificate within the polling timeout. When true (default), the install
  # fails immediately if the timeout is reached, providing clear feedback that
  # BackendTLSPolicy is non-functional. When false, the hook succeeds with a
  # warning and you can run `helm upgrade` after cert-manager issues the
  # certificate to create the backend CA ConfigMap. If you set this to false and
  # see "TLS error: Secret is not supplied by SDS" when connecting to the gateway,
  # check if the TLS secret exists and run `helm upgrade` to create the ConfigMap.
  failOnTimeout: true

# cert-manager Certificate/Issuer resources (requires cert-manager CRDs in-cluster).
# Does not install cert-manager itself.
certManager:
  # -- Create cert-manager Issuer and Certificate resources. When enabled,
  # cert-manager owns TLS and the chart runs a JWT-only certgen hook to create
  # the sandbox JWT signing Secret that cert-manager does not manage.
  enabled: false
  # -- Secret created for the intermediate CA (Certificate with isCA: true).
  caSecretName: openshell-ca-tls
  # -- Override the issuerRef for the external server Certificate (e.g. a real
  # LetsEncrypt/ACME ClusterIssuer for a publicly-trusted cert on an external hostname).
  # When set, the chart creates a second server certificate from this issuer
  # with only the hostnames in serverDnsNames; the internal server certificate
  # is always signed by the chart's own CA. Leave name empty to use the chart
  # CA for all server certificates (default). Requires certManager.enabled=true.
  serverIssuerRef:
    name: ""
    kind: ""
    group: ""
  # -- Mount gateway client CA from the internal server TLS secret's ca.crt.
  # The internal server certificate is always signed by the chart CA — the same
  # CA that signs the client (mTLS) certificate — so the default (true) is
  # correct for all configurations, including when serverIssuerRef is set.
  # Only set to false if you mount the client CA from a separate secret via
  # server.tls.clientCaSecretName.
  clientCaFromServerTlsSecret: true
  # -- Duration for cert-manager-issued certificates.
  certificateDuration: 8760h
  # -- Renewal window for cert-manager-issued certificates.
  certificateRenewBefore: 720h
  # -- DNS SANs on the cert-manager-issued server certificate.
  serverDnsNames:
    - openshell
    - openshell.openshell.svc
    - openshell.openshell.svc.cluster.local
    - localhost
    - openshell.localhost
    - "*.openshell.localhost"
    - host.docker.internal
  # -- IP SANs on the cert-manager-issued server certificate.
  serverIpAddresses:
    - 127.0.0.1

# Kubernetes Gateway API - HTTPRoute and Gateway resources.
# Requires a Gateway API controller in the cluster. Install Envoy Gateway via
# the skaffold.yaml releases or independently:
#   helm install eg oci://docker.io/envoyproxy/gateway-helm \
#     --version v1.4.1 -n envoy-gateway-system --create-namespace
grpcRoute:
  # -- Create a Gateway API GRPCRoute for the gateway service.
  enabled: false
  # -- Hostnames the GRPCRoute matches on. Leave empty to match all hosts.
  hostnames: []
  gateway:
    # -- When true, a Gateway resource is created in the release namespace.
    # Set to false and provide name/namespace to attach to a pre-existing Gateway.
    create: false
    # -- GatewayClass to reference. Envoy Gateway installs one named "eg".
    className: "eg"
    # -- Name of the Gateway resource. Defaults to the chart fullname.
    name: ""
    # -- Namespace of the Gateway referenced by the GRPCRoute parentRef.
    # Defaults to the release namespace.
    namespace: ""
    # Listener settings (only used when gateway.create is true).
    listener:
      # -- Listener port for the generated Gateway resource. Use 443 with protocol HTTPS.
      port: 80
      # -- Listener protocol for the generated Gateway resource: HTTP or HTTPS.
      # HTTPS terminates TLS at the Envoy Gateway listener; pair it with
      # server.disableTls=true so Envoy forwards plaintext to the gateway pod,
      # and use OIDC for client identity (the gateway never sees the client cert).
      protocol: HTTP
      # -- "Same" restricts attached routes to the release namespace; "All" allows any namespace.
      allowedRoutes: Same
      # TLS settings for the listener. Used only when protocol is HTTPS
      # (mode is always Terminate).
      tls:
        # -- certificateRefs for the HTTPS listener. Required when protocol is
        # HTTPS. Each entry needs a `name` pointing at a kubernetes.io/tls Secret
        # in the Gateway's namespace. May reference a cert-manager-issued Secret
        # or the existing openshell-server-tls Secret (its SANs must include the
        # external hostname).
        certificateRefs: []
  # BackendTLSPolicy for end-to-end TLS between the Gateway proxy and the
  # OpenShell gateway pod. When enabled, the Gateway proxy terminates
  # client-facing TLS at the listener and re-encrypts when connecting to the
  # backend service, validating the backend's certificate against the
  # specified CA. Requires the gateway pod to serve TLS (server.disableTls
  # must be false). Supported on OpenShift 4.22+ and other platforms with
  # BackendTLSPolicy support in the Gateway API implementation.
  backendTLSPolicy:
    # -- Create a BackendTLSPolicy resource for end-to-end TLS between the
    # Gateway proxy and the OpenShell gateway pod. The traffic flow is:
    # client → HTTPS → Gateway (terminate) → TLS (re-encrypt) → gateway pod.
    # Requires server.disableTls=false and server.tls.enableMtls=false. The
    # certgen hook auto-creates the backend CA ConfigMap.
    enabled: false
    # -- Name of the ConfigMap containing the CA certificate (key: ca.crt) used
    # to validate the gateway pod's TLS certificate. Defaults to
    # `<fullname>-backend-ca` when empty. The certgen hook auto-creates this:
    # with pkiInitJob (default), immediately on install/upgrade; with
    # cert-manager, the hook polls for pkiInitJob.timeoutSeconds seconds
    # waiting for cert-manager to issue the server certificate, then creates the
    # ConfigMap. A single install usually succeeds; if cert-manager takes longer,
    # increase pkiInitJob.timeoutSeconds. By default (pkiInitJob.failOnTimeout=true),
    # the install fails if the timeout is reached; set failOnTimeout=false to allow
    # the install to succeed and run `helm upgrade` after the certificate is issued.
    caCertificateConfigMapName: ""
    # -- Hostname the Gateway proxy validates against the backend's TLS
    # certificate SAN. Defaults to the service FQDN
    # (`<fullname>.<namespace>.svc.cluster.local`) when empty, which matches
    # the SAN included by both cert-manager and the pkiInitJob.
    hostname: ""

# OpenShift Route with TLS passthrough. The gateway terminates its own
# TLS/mTLS; the router only forwards based on SNI, so it never sees plaintext
# or the client certificate. Requires server.disableTls=false and a server
# cert whose SANs include the Route host (see certManager.serverIssuerRef).
openshiftRoute:
  # -- Create an OpenShift Route with TLS passthrough.
  enabled: false
  # -- Hostname for the Route. Must match a SAN on the gateway's server cert.
  host: ""
  # -- Extra annotations on the Route (e.g. haproxy.router.openshift.io/*).
  annotations: {}

```

---

## Assistant (Build · GPT-5.6 Terra · 30.9s)

_Thinking:_

**Evaluating mTLS and security policies**

I’m wondering whether disabling mTLS for supervisors might create issues. I’m comparing OIDC with the use of minted JWT within the gateway. The K8s docs mention that mTLS is essential for security but disabling it could be acceptable if the service mesh provides adequate protection. I need to ask if it’s enough to just have the BFF manage certificates. Also, there are implications for how cert managers and issuers work that need clarification.

_Thinking:_

**Exploring mTLS options in sandbox**

I’m thinking that the best setup might involve the service mesh covering all sandbox supervisors. If we go with option B, keeping application mTLS ensures that a BFF leaf certificate can work across all gateways, provided they trust the issuing CA. This would need a central issuer and proper configurations for each gateway trust bundle. The user seems to want code solutions rather than new infrastructure. I can recommend option B since it preserves mTLS while addressing their existing mesh.

_Thinking:_

**Considering server trust**

I’m thinking about how gateways mount the same bundle. That makes sense. I realize there's still a need for all the root certificates from the trusted Certificate Authorities (CA) for the BFF server. It’s important to ensure the server's trustworthiness. I want to double-check that I’m doing this correctly so that everything is secure and works smoothly. It's a bit technical, but getting it right is so important!

_Thinking:_

**Thinking about concise answers**

I'm considering that the user may not want an overly exhaustive response. I realize I need to provide a correct caveat to clarify any potential misunderstandings. It’s important that I answer their question concisely; perhaps including a diagram will help. However, I shouldn't introduce any code changes in my response. So, keeping it straightforward with clarity is my goal!

Yes. One BFF client certificate can serve every gateway **only when every gateway trusts its issuing CA**.

cert-manager does not establish shared trust. Each default OpenShell release creates its own `openshell-ca`; a BFF cert issued by another cert-manager Issuer is rejected unless gateway trusts that issuer's CA.

**Best Choices**

1. **Mesh + OIDC, no OpenShell client mTLS for BFF**
   
   Configure every gateway:

   ```yaml
   server:
     tls:
       enableMtls: false
     oidc:
       issuer: https://<issuer>
       audience: <openshell-audience>
   ```

   BFF needs only server CA trust, plus per-request OIDC bearer forwarding.

   This matches chart comment: HTTPS-only with OIDC. Mesh `STRICT` mTLS provides workload identity and encrypted pod transport. Do this only if mesh covers gateway, BFF, and sandbox-supervisor paths; otherwise OpenShell supervisor transport loses application-layer mTLS.

2. **Keep OpenShell mTLS**
   
   Use one platform client-CA and one BFF leaf cert:

   ```text
   platform client CA
     -> agent-ops BFF client cert/key

   every OpenShell gateway trusts platform client CA
     + its existing chart CA for sandbox supervisors
   ```

   BFF mounts one cert/key in `redhat-ods-applications`. No gateway private keys. No per-gateway BFF certs.

**For Option 2**

Each gateway must replace its client CA Secret with a bundle:

```text
ca.crt =
  chart's openshell-ca public certificate
  + platform BFF-client-CA public certificate
```

Helm values per gateway:

```yaml
certManager:
  clientCaFromServerTlsSecret: false

server:
  tls:
    enableMtls: true
    clientCaSecretName: openshell-client-ca-bundle
```

This preserves chart-issued supervisor client certificates while accepting BFF cert from common platform CA.

BFF gets leaf from a `ClusterIssuer` or external issuer in `redhat-ods-applications`. Issuer private key stays in cert-manager control plane, never BFF namespace.

**BFF Server Trust**

Every default OpenShell installation has separate internal server CA. BFF must trust all gateway server CA public certs:

```text
gateway-a openshell-ca public cert
gateway-b openshell-ca public cert
gateway-c openshell-ca public cert
```

Deliver merged public CA bundle to BFF with trust-manager/GitOps. BFF uses it as `RootCAs`. This is public trust data, not gateway private material.

**Recommendation**

Use Option 1 if mesh reliably covers all OpenShell traffic. Simpler: no BFF private TLS key; OIDC remains user identity.

Use Option 2 if OpenShell application mTLS must remain enabled. One BFF cert is enough, but only after every gateway trusts its CA.

---

## User

lets go with option 2. expand on it

---

## Assistant (Build · GPT-5.6 Terra · 15.5s)

_Thinking:_

**Clarifying chart template**

I need to determine the correct mount for my chart template, which involves fetching the workload mount accurately. The documentation indicates that the values are secured with ca.crt, so I believe using an Opaque type should suffice. It’s essential to ensure that I’ve followed the right approach to avoid any misconfiguration. I’ll double-check the specifics to confirm everything aligns correctly!

_Thinking:_



_Thinking:_



Detailing shared client-CA model: one BFF leaf, per-gateway trust bundles preserving sandbox certs, rotation and auth boundaries.

**Tool: webfetch**

**Input:**
```json
{
  "url": "https://raw.githubusercontent.com/NVIDIA/OpenShell/main/deploy/helm/openshell/templates/_gateway-workload.tpl",
  "format": "text"
}
```

**Output:**
```
# SPDX-FileCopyrightText: Copyright (c) 2025-2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

{{/*
Gateway pod template shared by the StatefulSet and Deployment workload shapes.
*/}}
{{- define "openshell.gatewayPodTemplate" -}}
metadata:
  annotations:
    # Roll the gateway workload when the rendered gateway TOML changes - the
    # gateway only reads /etc/openshell/gateway.toml at startup, so without
    # this annotation a `helm upgrade` that only mutates the ConfigMap would
    # leave pods running with stale config.
    checksum/gateway-config: {{ include (print $.Template.BasePath "/gateway-config.yaml") . | sha256sum }}
    {{- with .Values.podAnnotations }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
  labels:
    {{- include "openshell.labels" . | nindent 4 }}
    {{- with .Values.podLabels }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
spec:
  terminationGracePeriodSeconds: {{ .Values.podLifecycle.terminationGracePeriodSeconds }}
  {{- with .Values.imagePullSecrets }}
  imagePullSecrets:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  serviceAccountName: {{ include "openshell.serviceAccountName" . }}
  {{- if .Values.server.hostGatewayIP }}
  hostAliases:
    - ip: {{ .Values.server.hostGatewayIP | quote }}
      hostnames:
        - host.docker.internal
        - host.openshell.internal
  {{- end }}
  {{- with .Values.podSecurityContext }}
  securityContext:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  containers:
    - name: openshell-gateway
      securityContext:
        {{- toYaml .Values.securityContext | nindent 8 }}
      image: {{ include "openshell.image" . | quote }}
      imagePullPolicy: {{ .Values.image.pullPolicy }}
      args:
        - --config
        - /etc/openshell/gateway.toml
        {{- if not .Values.server.externalDbSecret }}
        - --db-url
        - {{ .Values.server.dbUrl | quote }}
        {{- end }}
      env:
        - name: OPENSHELL_REPLICA_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: OPENSHELL_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: OPENSHELL_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        {{- if eq (include "openshell.workloadKind" .) "deployment" }}
        - name: OPENSHELL_POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: OPENSHELL_PEER_ENDPOINT
          value: {{ printf "%s://$(OPENSHELL_POD_IP):%d" (ternary "http" "https" (default false .Values.server.disableTls)) (int .Values.service.port) | quote }}
        {{- end }}
        - name: OPENSHELL_SERVICE_ACCOUNT_NAME
          value: {{ include "openshell.serviceAccountName" . | quote }}
        - name: OPENSHELL_PEER_SERVICE_NAME
          value: {{ include "openshell.peerServiceName" . | quote }}
        - name: OPENSHELL_PEER_TOKEN_AUDIENCE
          value: "openshell-gateway-peer"
        - name: OPENSHELL_PEER_SERVICE_ACCOUNT_TOKEN_FILE
          value: /var/run/secrets/openshell-peer/token
        - name: OPENSHELL_PEER_POD_LABELS
          value: {{ printf "app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s" (include "openshell.name" .) .Release.Name | quote }}
        {{- if not .Values.server.disableTls }}
        - name: OPENSHELL_PEER_TLS_SERVER_NAME
          value: {{ printf "%s.%s.svc.cluster.local" (include "openshell.fullname" .) .Release.Namespace | quote }}
        {{- if or .Values.pkiInitJob.enabled .Values.certManager.enabled }}
        - name: OPENSHELL_PEER_TLS_CA_FILE
          value: /etc/openshell-tls/server/ca.crt
        {{- end }}
        {{- if eq (include "openshell.gatewayClientCaEnabled" .) "true" }}
        - name: OPENSHELL_PEER_TLS_CERT_FILE
          value: /etc/openshell-tls/peer-client/tls.crt
        - name: OPENSHELL_PEER_TLS_KEY_FILE
          value: /etc/openshell-tls/peer-client/tls.key
        {{- end }}
        {{- end }}
        {{- if not (or .Values.server.credentialDrivers.kubernetesSecrets.enabled .Values.server.credentialDrivers.vault.enabled) }}
        - name: {{ include "openshell.credentialStorageKeyEncryptionKeyEnvName" . }}
          valueFrom:
            secretKeyRef:
              name: {{ include "openshell.credentialStorageKeyEncryptionKeySecretName" . }}
              key: {{ include "openshell.credentialStorageKeyEncryptionKeySecretKey" . }}
        {{- end }}
        {{- if .Values.server.externalDbSecret }}
        - name: OPENSHELL_DB_URL
          valueFrom:
            secretKeyRef:
              name: {{ .Values.server.externalDbSecret }}
              key: uri
        {{- end }}
        # Most gateway settings live in the ConfigMap-backed TOML file
        # mounted at /etc/openshell/gateway.toml. Secret-bearing settings use
        # env vars that the TOML references by name. Some process-level
        # settings consumed by libraries outside gateway code also remain here.
        {{- if and .Values.server.oidc.issuer .Values.server.oidc.caConfigMapName }}
        # OIDC issuer custom-CA: rustls/reqwest read SSL_CERT_FILE for
        # outbound TLS verification. This is a process-level env var
        # consumed by the TLS stack itself, not by gateway code, so it
        # cannot be represented in the gateway TOML schema.
        - name: SSL_CERT_FILE
          value: /etc/openshell-tls/oidc-ca/ca.crt
        {{- end }}
        - name: OPENSHELL_TELEMETRY_ENABLED
          value: {{ .Values.server.telemetryEnabled | quote }}
        {{- if .Values.server.providerTokenGrants.spiffe.enabled }}
        - name: OPENSHELL_GATEWAY_SPIFFE_WORKLOAD_API_SOCKET
          value: {{ .Values.server.providerTokenGrants.spiffe.workloadApiSocketPath | quote }}
        {{- end }}
      volumeMounts:
        {{- if eq (include "openshell.workloadKind" .) "statefulset" }}
        - name: openshell-data
          mountPath: /var/openshell
        {{- end }}
        # ConfigMap directory mounts expose keys through atomic-writer symlinks,
        # while the gateway intentionally rejects symlinked configuration.
        # The checksum annotation above rolls pods when this subPath changes.
        - name: gateway-config
          mountPath: /etc/openshell/gateway.toml
          subPath: gateway.toml
          readOnly: true
        - name: sandbox-jwt
          mountPath: /etc/openshell-jwt
          readOnly: true
        - name: gateway-peer-token
          mountPath: /var/run/secrets/openshell-peer
          readOnly: true
        {{- if not .Values.server.disableTls }}
        - name: tls-cert
          mountPath: /etc/openshell-tls/server
          readOnly: true
        {{- if .Values.certManager.serverIssuerRef.name }}
        - name: tls-external-cert
          mountPath: /etc/openshell-tls/server-external
          readOnly: true
        {{- end }}
        {{- if eq (include "openshell.gatewayClientCaEnabled" .) "true" }}
        - name: peer-client-tls
          mountPath: /etc/openshell-tls/peer-client
          readOnly: true
        - name: tls-client-ca
          mountPath: /etc/openshell-tls/client-ca
          readOnly: true
        {{- end }}
        {{- end }}
        {{- if and .Values.server.oidc.issuer .Values.server.oidc.caConfigMapName }}
        - name: oidc-ca
          mountPath: /etc/openshell-tls/oidc-ca
          readOnly: true
        {{- end }}
        {{- if and .Values.server.credentialDrivers.vault.enabled .Values.server.credentialDrivers.vault.caConfigMapName }}
        - name: vault-ca
          mountPath: /etc/openshell-tls/vault-ca
          readOnly: true
        {{- end }}
        {{- if .Values.upstreamProxy.caBundle.configMapName }}
        - name: upstream-proxy-ca
          mountPath: /etc/openshell-tls/proxy-ca
          readOnly: true
        {{- end }}
        {{- if .Values.server.providerTokenGrants.spiffe.enabled }}
        - name: spiffe-workload-api
          mountPath: {{ dir .Values.server.providerTokenGrants.spiffe.workloadApiSocketPath | quote }}
          readOnly: true
        {{- end }}
      ports:
        - name: grpc
          containerPort: {{ .Values.service.port }}
          protocol: TCP
        - name: health
          containerPort: {{ .Values.service.healthPort }}
          protocol: TCP
        {{- if .Values.service.metricsPort }}
        - name: metrics
          containerPort: {{ .Values.service.metricsPort }}
          protocol: TCP
        {{- end }}
      startupProbe:
        httpGet:
          path: /healthz
          port: health
        periodSeconds: {{ .Values.probes.startup.periodSeconds }}
        timeoutSeconds: {{ .Values.probes.startup.timeoutSeconds }}
        failureThreshold: {{ .Values.probes.startup.failureThreshold }}
      livenessProbe:
        httpGet:
          path: /healthz
          port: health
        initialDelaySeconds: {{ .Values.probes.liveness.initialDelaySeconds }}
        periodSeconds: {{ .Values.probes.liveness.periodSeconds }}
        timeoutSeconds: {{ .Values.probes.liveness.timeoutSeconds }}
        failureThreshold: {{ .Values.probes.liveness.failureThreshold }}
      readinessProbe:
        httpGet:
          path: /readyz
          port: health
        initialDelaySeconds: {{ .Values.probes.readiness.initialDelaySeconds }}
        periodSeconds: {{ .Values.probes.readiness.periodSeconds }}
        timeoutSeconds: {{ .Values.probes.readiness.timeoutSeconds }}
        failureThreshold: {{ .Values.probes.readiness.failureThreshold }}
      resources:
        {{- toYaml .Values.resources | nindent 8 }}
  volumes:
    - name: gateway-config
      configMap:
        name: {{ include "openshell.fullname" . }}-config
    - name: sandbox-jwt
      secret:
        secretName: {{ include "openshell.sandboxJwtSecretName" . }}
        defaultMode: {{ .Values.server.sandboxJwt.secretDefaultMode | default 0400 }}
    - name: gateway-peer-token
      projected:
        defaultMode: 0400
        sources:
          - serviceAccountToken:
              path: token
              audience: openshell-gateway-peer
              expirationSeconds: 3600
    {{- if not .Values.server.disableTls }}
    - name: tls-cert
      secret:
        secretName: {{ .Values.server.tls.certSecretName }}
    {{- if .Values.certManager.serverIssuerRef.name }}
    - name: tls-external-cert
      secret:
        secretName: {{ include "openshell.fullname" . }}-server-external-tls
    {{- end }}
    {{- if eq (include "openshell.gatewayClientCaEnabled" .) "true" }}
    - name: peer-client-tls
      secret:
        secretName: {{ .Values.server.tls.clientTlsSecretName }}
    - name: tls-client-ca
      secret:
        {{- if or (and .Values.pkiInitJob.enabled (not .Values.certManager.enabled)) (and .Values.certManager.enabled .Values.certManager.clientCaFromServerTlsSecret) }}
        secretName: {{ .Values.server.tls.certSecretName }}
        items:
          - key: ca.crt
            path: ca.crt
        {{- else }}
        secretName: {{ .Values.server.tls.clientCaSecretName }}
        {{- end }}
    {{- end }}
    {{- end }}
    {{- if and .Values.server.oidc.issuer .Values.server.oidc.caConfigMapName }}
    - name: oidc-ca
      configMap:
        name: {{ .Values.server.oidc.caConfigMapName }}
    {{- end }}
    {{- if and .Values.server.credentialDrivers.vault.enabled .Values.server.credentialDrivers.vault.caConfigMapName }}
    - name: vault-ca
      configMap:
        name: {{ .Values.server.credentialDrivers.vault.caConfigMapName }}
        items:
          - key: ca.crt
            path: ca.crt
    {{- end }}
    {{- if .Values.upstreamProxy.caBundle.configMapName }}
    - name: upstream-proxy-ca
      configMap:
        name: {{ .Values.upstreamProxy.caBundle.configMapName | quote }}
        items:
          # The mounted filename stays fixed so the rendered proxy_ca_bundle
          # path does not depend on the operator's ConfigMap key.
          - key: {{ .Values.upstreamProxy.caBundle.key | default "ca.crt" | quote }}
            path: ca.crt
    {{- end }}
    {{- if .Values.server.providerTokenGrants.spiffe.enabled }}
    - name: spiffe-workload-api
      csi:
        driver: csi.spiffe.io
        readOnly: true
    {{- end }}
  {{- with .Values.nodeSelector }}
  nodeSelector:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.affinity }}
  affinity:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.tolerations }}
  tolerations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}

```

---

## Assistant (Build · GPT-5.6 Terra · 44.0s)

_Thinking:_



_Thinking:_

**Evaluating production parameters**

I’m considering that there might be some issues related to the scope of the project. It’s interesting to note that “best production” indicates that the fleet proxy relies on the openshell dashboard middleware, rather than using the existing agent-ops authentication. This distinction could be crucial in understanding how production is managed and integrated, so I'll need to look more closely into the specifics to see how this might impact our approach.

_Thinking:_



**Target Design**

One BFF leaf certificate. One platform client-CA. Every OpenShell gateway trusts that client-CA.

```text
platform client CA
  └─ agent-ops BFF client certificate
       └─ mounted only in redhat-ods-applications

gateway-a client CA bundle
  ├─ gateway-a chart CA       # preserves its sandbox supervisors
  └─ platform client CA       # accepts BFF

gateway-b client CA bundle
  ├─ gateway-b chart CA
  └─ platform client CA
```

cert-manager only automates issuance. It does not make separately issued certificates mutually trusted.

**Do Not Use**

Do not issue BFF cert from `gateway-a`'s `openshell-ca-issuer`.

That cert works only against gateway-a. Every default OpenShell release owns a distinct `openshell-ca`, so it fails against gateway-b.

**1. Create Platform Client CA**

Use existing internal PKI-backed `ClusterIssuer`, or create dedicated platform `ClusterIssuer` for OpenShell workload client certificates.

Requirements:

- Internal CA. Never Let's Encrypt/public web PKI.
- Can issue certificates with `client auth` EKU.
- Available from `redhat-ods-applications`.
- Issuer private key remains in cert-manager control-plane namespace.
- Only platform GitOps/cert-management identities can create `Certificate` resources against issuer.

Call it, for example, `openshell-bff-client`.

**2. Issue Single BFF Certificate**

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: agent-ops-openshell-client
  namespace: redhat-ods-applications
spec:
  secretName: agent-ops-openshell-client-tls
  commonName: agent-ops-bff
  duration: 2160h
  renewBefore: 360h
  privateKey:
    algorithm: ECDSA
    size: 256
  usages:
    - client auth
    - digital signature
    - key encipherment
  issuerRef:
    name: openshell-bff-client
    kind: ClusterIssuer
    group: cert-manager.io
```

BFF pod mounts only this Secret:

```text
tls.crt
tls.key
```

No BFF ServiceAccount Secret-read RBAC required. Kubelet mounts named Secret into BFF pod read-only.

**3. Configure Every Gateway**

For each OpenShell Helm release:

```yaml
certManager:
  clientCaFromServerTlsSecret: false

server:
  tls:
    enableMtls: true
    clientCaSecretName: openshell-client-ca-bundle
```

Create `openshell-client-ca-bundle` in that specific gateway namespace:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: openshell-client-ca-bundle
  namespace: <gateway-namespace>
type: Opaque
stringData:
  ca.crt: |
    # Public root from this gateway's openshell-ca-tls/tls.crt.
    -----BEGIN CERTIFICATE-----
    <gateway-specific-chart-ca>
    -----END CERTIFICATE-----

    # Public root/intermediate used by ClusterIssuer openshell-bff-client.
    -----BEGIN CERTIFICATE-----
    <platform-bff-client-ca>
    -----END CERTIFICATE-----
```

First CA keeps chart-issued `openshell-client-tls` working for sandbox supervisors. Second CA lets gateway accept BFF leaf cert. Bundle contains public certificates only.

Do not commit generated Secret data. Materialize bundle through GitOps secret integration, External Secrets, or controlled cert-management automation.

**4. Give BFF Server Trust**

BFF must verify each gateway server certificate. Default OpenShell installs create distinct server CAs, so BFF needs merged public bundle:

```text
gateway-a openshell-ca-tls/tls.crt
gateway-b openshell-ca-tls/tls.crt
gateway-c openshell-ca-tls/tls.crt
```

Mount merged bundle as BFF `gateway-ca-bundle.pem`.

BFF connection configuration:

```text
Gateway endpoint: https://<service>.<gateway-namespace>.svc.cluster.local:8080
Server CA file:  /etc/openshell/tls/gateway-ca-bundle.pem
Client cert:     /etc/openshell/tls/client/tls.crt
Client key:      /etc/openshell/tls/client/tls.key
```

Use service FQDN in each endpoint. It drives hostname validation and SNI. Do not use `localhost` outside local port-forward tests.

**5. Required BFF Changes**

- Remove hardcoded gateway URLs in `pkg/discovery/discovery.go`.
- Load gateway endpoint inventory from config or Kubernetes discovery.
- Pass CA bundle, BFF client cert, BFF client key into `NewGatewayClients`.
- Replace `openshell.NoAuth()` in `internal/clients/gateway.go` with `clients.ContextAuthProvider{RequireTLS: true}`.
- Retain OpenShell dashboard auth middleware so inbound user bearer token enters request context.
- Configure every gateway with same OIDC issuer/audience. BFF mTLS cert authenticates workload; forwarded OIDC token authenticates and authorizes user.
- Fail startup when HTTPS endpoint lacks CA bundle or only one client cert/key path exists.
- Reject `INSECURE_SKIP_VERIFY` in production.

**Rotation**

- BFF leaf renewal: restart/recreate BFF gateway clients. Current SDK loads PEM at construction.
- Gateway chart CA renewal: update BFF server trust bundle before new gateway server certificate becomes active.
- Platform BFF-client CA renewal: maintain old and new CA in every gateway client-CA bundle during overlap.
- Update gateway bundle: roll gateway workload. Chart restarts on config checksum changes, not arbitrary Secret-content changes; use a Secret reloader or controlled rollout.

**Validation Matrix**

```text
BFF cert signed by platform client CA -> every gateway succeeds
BFF cert signed by gateway-a chart CA -> only gateway-a succeeds
No BFF cert                         -> gateway rejects mTLS
Unknown BFF issuer                  -> gateway rejects mTLS
Valid BFF cert + no user JWT        -> gateway rejects authorization
Valid BFF cert + valid user JWT     -> succeeds if role permits
```

This preserves OpenShell mTLS, supports many gateway namespaces, keeps root/server private keys out of BFF, and avoids one BFF certificate per gateway.

---

## User

give me how to create clusterissuer and how to set helm chart values via cli

---

## Assistant (Build · GPT-5.6 Terra · 50.6s)

_Thinking:_

**Creating a ClusterIssuer**

The user wants guidance on creating a ClusterIssuer and updating Helm values, so I need to provide clear, procedural instructions. It's important to be security-conscious here, as creating a ClusterIssuer and configuring charts involves irreversible steps. 

To create a simple self-signed ClusterIssuer, I would start with an API version and metadata setup. Next, I'd create a CA Certificate in the cert-manager namespace, defining its characteristics and the issuer reference. Finally, I'd establish the ClusterIssuer using the CA, ensuring proper secret name usage within the cert-manager namespace.

_Thinking:_

**Verifying ClusterIssuer Secrets**

According to the cert-manager documentation, ClusterIssuer secrets remain in the `--cluster-resource-namespace`, which defaults to the `cert-manager` namespace. To set up a CA certificate, the `issuerRef` referencing the ClusterIssuer named `openshell-bff-selfsigned` should have its secret in the cert-manager namespace.

I note that a certificate leaf in the Red Hat namespace can utilize this cluster issuer. However, for distributing the "BFF client CA public root" across gateways, I'll provide a CLI command to create the secret using a process substitution method by combining outputs from different secrets.

_Thinking:_



_Thinking:_

**Configuring TLS Settings**

I'm deciding to use the root `tls.crt`. For the leaf `usages`, modern certificate managers recognize options like `client auth`, `digital signature`, and maybe `key encipherment`. I want to ensure I use exactly what the chart specifies. The user is asking about helm chart values, particularly including the client CA bundle secret in the config. It might be necessary to do a `helm upgrade` in a loop and check the rollout status with `kubectl`.

_Thinking:_

**Evaluating Helm Chart Settings**

I might need to adjust `certManager.clientCaFromServerTlsSecret` to be `false`. If the chart template condition for `server tls clientCaSecretName` is still set to the default `openshell-server-client-ca`, it could fail. We should work together to update our secret. I wonder if the user is specifically referring to the BFF helm chart? They only mentioned helm chart values, possibly for OpenShell charts. It's important to note that the BFF deployment must mount the leaf and CA trust bundle separately.

_Thinking:_

**Evaluating chart values**

I’m considering how to open the chart values via CLI. It seems I want to keep any updates related to the BFF separate, without making changes to the code. I need to ensure my response is longer and includes some caveats. I'll steer clear of nested bullet points and keep the headers short and linear. It’s all about being concise and direct, almost in a "caveman" style. Let’s put that into practice!

**Prerequisite**

Use dedicated private CA for BFF client certificates. Do not use public ACME issuer.

`ClusterIssuer` CA Secret lives in cert-manager cluster-resource namespace. Usually `cert-manager`; verify controller flag before applying:

```sh
kubectl -n cert-manager get deployment cert-manager \
  -o jsonpath='{.spec.template.spec.containers[0].args}'
```

Look for `--cluster-resource-namespace`. Commands below assume `cert-manager`.

**Create Client CA**

```sh
kubectl apply -f - <<'EOF'
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: openshell-bff-selfsigned
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: openshell-bff-client-ca
  namespace: cert-manager
spec:
  isCA: true
  commonName: openshell-bff-client-ca
  secretName: openshell-bff-client-ca
  duration: 8760h
  renewBefore: 720h
  privateKey:
    algorithm: ECDSA
    size: 256
  usages:
    - cert sign
    - crl sign
    - digital signature
  issuerRef:
    name: openshell-bff-selfsigned
    kind: ClusterIssuer
    group: cert-manager.io
EOF

kubectl -n cert-manager wait \
  --for=condition=Ready certificate/openshell-bff-client-ca \
  --timeout=2m
```

Create issuer backed by that CA:

```sh
kubectl apply -f - <<'EOF'
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: openshell-bff-client
spec:
  ca:
    secretName: openshell-bff-client-ca
EOF

kubectl wait --for=condition=Ready clusterissuer/openshell-bff-client --timeout=2m
```

`openshell-bff-client-ca/tls.key` is CA private key. It must remain only in cert-manager cluster-resource namespace.

**Issue BFF Leaf**

```sh
kubectl apply -f - <<'EOF'
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: agent-ops-openshell-client
  namespace: redhat-ods-applications
spec:
  secretName: agent-ops-openshell-client-tls
  commonName: agent-ops-bff
  duration: 2160h
  renewBefore: 360h
  privateKey:
    algorithm: ECDSA
    size: 256
  usages:
    - client auth
    - digital signature
    - key encipherment
  issuerRef:
    name: openshell-bff-client
    kind: ClusterIssuer
    group: cert-manager.io
EOF

kubectl -n redhat-ods-applications wait \
  --for=condition=Ready certificate/agent-ops-openshell-client \
  --timeout=2m
```

BFF later mounts `agent-ops-openshell-client-tls/tls.crt` and `tls.key`.

**Create Gateway Trust Bundle**

Run once per gateway namespace. This produces `ca.crt` containing that gateway's existing chart CA plus BFF client CA.

```sh
export GATEWAY_NAMESPACE=<gateway-namespace>
export TMP_DIR="$(mktemp -d)"

kubectl -n "$GATEWAY_NAMESPACE" get secret openshell-ca-tls \
  -o jsonpath='{.data.tls\.crt}' | openssl base64 -d -A \
  > "$TMP_DIR/gateway-ca.crt"

kubectl -n cert-manager get secret openshell-bff-client-ca \
  -o jsonpath='{.data.tls\.crt}' | openssl base64 -d -A \
  > "$TMP_DIR/bff-client-ca.crt"

cat "$TMP_DIR/gateway-ca.crt" "$TMP_DIR/bff-client-ca.crt" \
  > "$TMP_DIR/ca.crt"

kubectl -n "$GATEWAY_NAMESPACE" create secret generic openshell-client-ca-bundle \
  --from-file=ca.crt="$TMP_DIR/ca.crt" \
  --dry-run=client -o yaml | kubectl apply -f -
```

This preserves gateway's default supervisor client certificate trust. Do not replace bundle with only BFF CA.

**Set Helm Values**

Resource shows chart version `0.0.116`. Upgrade each release after bundle exists:

```sh
export GATEWAY_NAMESPACE=<gateway-namespace>
export OPENSHELL_VERSION=0.0.116

helm upgrade openshell oci://ghcr.io/nvidia/openshell/helm-chart \
  --version "$OPENSHELL_VERSION" \
  --namespace "$GATEWAY_NAMESPACE" \
  --reuse-values \
  --set certManager.enabled=true \
  --set certManager.clientCaFromServerTlsSecret=false \
  --set server.tls.enableMtls=true \
  --set server.tls.clientCaSecretName=openshell-client-ca-bundle \
  --wait
```

Use actual release name if not `openshell`.

The chart mounts `openshell-client-ca-bundle/ca.crt` at gateway TLS client-CA path. Gateway then accepts both:

```text
chart sandbox client certificates
BFF certificates signed by openshell-bff-client
```

Check rollout against actual workload kind:

```sh
kubectl -n "$GATEWAY_NAMESPACE" rollout status statefulset/openshell
```

Or:

```sh
kubectl -n "$GATEWAY_NAMESPACE" rollout status deployment/openshell
```

**BFF Trust Bundle**

BFF needs public roots for every gateway server CA. Build a PEM bundle from each gateway namespace's `openshell-ca-tls/tls.crt`, mount it as one CA file. BFF client uses:

```text
CA file:     merged gateway-server-ca.pem
Client cert: agent-ops-openshell-client-tls/tls.crt
Client key:  agent-ops-openshell-client-tls/tls.key
```

**Security Gate**

A `ClusterIssuer` can sign certificates from every namespace. Without policy, any subject allowed to create `Certificate` resources may mint a trusted gateway-client certificate.

Restrict `Certificate` creation with RBAC and cert-manager approver-policy before production rollout. Allow only controlled GitOps identity to request `openshell-bff-client` certificates.

---

## User

why are we creating two clusterissuers here. is that necessary?

---

## Assistant (Build · GPT-5.6 Terra · 17.3s)

_Thinking:_



No. Previous sequence used two signing stages, but only one needs cluster scope.

```text
self-signed bootstrap Issuer
  -> creates private CA certificate

CA ClusterIssuer
  -> creates BFF leaf certificates across namespaces
```

`openshell-bff-selfsigned` does not need cross-namespace use. Make it a namespaced `Issuer` in `cert-manager`, not `ClusterIssuer`:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: openshell-bff-selfsigned
  namespace: cert-manager
spec:
  selfSigned: {}
```

Then CA certificate references:

```yaml
issuerRef:
  name: openshell-bff-selfsigned
  kind: Issuer
```

Keep only this cluster-scoped issuer:

```text
ClusterIssuer/openshell-bff-client
```

It signs leaf certificates in `redhat-ods-applications`.

If cluster already has internal PKI `ClusterIssuer` capable of `client auth`, create neither. Issue BFF leaf directly from existing issuer and add its public CA chain to every gateway's `openshell-client-ca-bundle`.

Do not use one self-signed `ClusterIssuer` directly for BFF leaf. That creates self-signed leaf certificate, not shared CA hierarchy; gateway trust becomes brittle and non-scalable.

---

## User

can we not create two whole issuers. is there a deafult one we can use for the clusterissuer for bff-client

---

## Assistant (Build · GPT-5.6 Terra · 27.0s)

_Thinking:_



_Thinking:_

**Analyzing certificate requirements**

I’m considering whether the certificate needs to be signed by a client CA, especially if we’re using a service mesh. They've opted to keep app mTLS, so maybe we need to centralize the client CA. It seems I should mention that we don't need the ClusterIssuer called bff-client if an existing one, like `internal-pki`, is available. The certificate can set issuerRef directly, and a separate namespace issuer isn’t necessary. I’ll work on providing an answer.

cert-manager has no built-in default `ClusterIssuer`. Installing cert-manager creates controllers and CRDs, not a cluster CA.

Check cluster first:

```sh
kubectl get clusterissuer
kubectl get issuer -A
```

Use existing `ClusterIssuer` only if it is private PKI capable of issuing `client auth` certificates:

```sh
kubectl get clusterissuer <name> -o yaml
```

Valid issuer types:

```text
spec.ca
spec.vault
corporate/external cert-manager issuer
```

Usually invalid for this purpose:

```text
spec.acme          public web-PKI, not workload client auth
spec.selfSigned    produces self-signed leaf, not shared trust hierarchy
```

If cluster has suitable issuer, create no issuers. Point BFF certificate at it:

```yaml
issuerRef:
  name: <existing-private-clusterissuer>
  kind: ClusterIssuer
  group: cert-manager.io
```

Then add that issuer's public CA chain to every gateway's `openshell-client-ca-bundle`.

If cluster has no suitable issuer, one shared CA plus one `ClusterIssuer` is minimum clean design. Earlier bootstrap self-signed issuer exists only to make that CA automatically. It is not required if platform team can provide a CA Secret or existing private PKI issuer.

Do not repurpose `openshell-ca-issuer`. It is namespace-scoped and each OpenShell release owns a different CA, so it cannot support one BFF certificate across multiple gateways.

---

## User

where did you get openshell-bff-client-ca from>

---

## Assistant (Build · GPT-5.6 Terra · 9.9s)

_Thinking:_



`openshell-bff-client-ca` came from my proposed design. It is not existing cluster resource.

Resources confirmed from your YAML:

```text
Certificate/openshell-ca
Secret/openshell-ca-tls
Issuer/openshell-selfsigned
```

OpenShell chart also creates:

```text
Issuer/openshell-ca-issuer
```

`openshell-bff-client-ca` would exist only if platform creates dedicated shared client CA. Do not reference it unless you create it.

Next fact needed:

```sh
kubectl get clusterissuer
```

Use one real existing private PKI `ClusterIssuer` from that output, if available. Otherwise shared BFF certificate across independently installed OpenShell gateways needs new shared CA/issuer infrastructure.

---

## User

how do i get the secretname if its in "redhat-ods-applications"

---

## Assistant (Build · GPT-5.6 Terra · 13.3s)

_Thinking:_

**Clarifying deployment requirements**

I need to copy or move a CA source to the cluster resource namespace. For the deployment, the `secretName` references the implicit `agent...` namespace. If they ask about `clientCaSecretName` in chart values, I should clarify that the chart release in another namespace can’t directly refer to Secrets in Red Hat. The Secret must be in the gateway namespace, with no cross-namespace references. The gateway should only access the CA public bundle for its local use.

Kubernetes Secrets are namespace-scoped. Secret name never includes namespace.

BFF certificate Secret:

```yaml
metadata:
  namespace: redhat-ods-applications
spec:
  secretName: agent-ops-openshell-client-tls
```

cert-manager creates:

```text
Secret/agent-ops-openshell-client-tls
Namespace/redhat-ods-applications
```

Get it:

```sh
kubectl -n redhat-ods-applications get secret agent-ops-openshell-client-tls
```

Get name from Certificate:

```sh
kubectl -n redhat-ods-applications get certificate agent-ops-openshell-client \
  -o jsonpath='{.spec.secretName}{"\n"}'
```

For OpenShell Helm:

```yaml
server:
  tls:
    clientCaSecretName: openshell-client-ca-bundle
```

That Secret must exist in **each gateway's namespace**. Helm cannot mount a Secret from `redhat-ods-applications`.

Gateway does not need BFF leaf Secret. It needs only BFF issuer's public CA in local `openshell-client-ca-bundle/ca.crt`.

If asking about `ClusterIssuer.spec.ca.secretName`: it cannot reference a Secret in `redhat-ods-applications` either. `ClusterIssuer` reads its CA key from cert-manager’s configured cluster-resource namespace, usually `cert-manager`.

---

## User

well secretname is referenced in clusterissuer though

---

## Assistant (Build · GPT-5.6 Terra · 14.2s)

_Thinking:_



Correct. `ClusterIssuer.spec.ca.secretName` references CA signing-key Secret.

It cannot reference a Secret in `redhat-ods-applications` unless cert-manager was configured with:

```text
--cluster-resource-namespace=redhat-ods-applications
```

Usually it is:

```text
--cluster-resource-namespace=cert-manager
```

So this:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: openshell-bff-client
spec:
  ca:
    secretName: openshell-bff-client-ca
```

means:

```text
Secret/openshell-bff-client-ca
Namespace/cert-manager
Keys: tls.crt, tls.key
```

It does **not** reference BFF leaf Secret.

BFF leaf is separate, created in its namespace:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: agent-ops-openshell-client
  namespace: redhat-ods-applications
spec:
  secretName: agent-ops-openshell-client-tls
  issuerRef:
    name: openshell-bff-client
    kind: ClusterIssuer
```

Flow:

```text
cert-manager/openshell-bff-client-ca
  -> ClusterIssuer/openshell-bff-client
  -> redhat-ods-applications/agent-ops-openshell-client-tls
```

Do not set `ClusterIssuer.spec.ca.secretName` to `agent-ops-openshell-client-tls`. That is leaf cert/key, not CA signing material.

---


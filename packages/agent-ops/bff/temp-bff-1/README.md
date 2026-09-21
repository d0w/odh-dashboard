# Multi-Gateway BFF POC

Isolated proof of concept for mounting one upstream OpenShell `server.App` per gateway below:

```text
/api/openshell/{gateway-id}/api/v1/...
```

`Registry` owns dynamically registered gateway instances. Each instance contains an HTTP handler plus the close function for resources held outside upstream `server.App`, such as its SDK and raw-exec gRPC clients.

Consumer contracts:

- `GatewayRegistrar`: discovery/reconciliation adds and removes gateways.
- `GatewayRouter`: root BFF mounts request dispatch.
- `GatewayRegistry`: root BFF owns full lifecycle, including shutdown.

Discovery code must depend on `GatewayRegistrar`, not concrete `*Registry`.

```go
instance := tempbff.GatewayInstance{
    Handler: upstreamApp.Routes(),
    Close:   gatewayClients.Close,
}
```

Gateway creation and deletion come from trusted in-process code only. Browser requests only select an existing RFC 1123 gateway ID; they cannot set endpoint or TLS data.

`Remove` immediately unpublishes a handler, waits for active requests to finish, then closes its clients. This avoids interrupting uploads, terminal streams, or normal RPCs with `ClientConn.Close`.

Run POC tests:

```shell
cd temp-bff-1
go test -race ./...
```

This POC does not import upstream packages. Production integration should supply a factory that creates OpenShell clients, calls `server.NewApp(...)`, and returns `GatewayInstance{Handler: app.Routes(), Close: clients.Close}`. Keep `server.App` static assets disabled because Agent Ops owns static asset serving.

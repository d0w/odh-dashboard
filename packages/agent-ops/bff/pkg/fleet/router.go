// Package fleet handles upper level HTTP routing for a map of registered instances.
//
// Each instance should have their own HTTP Handlers which fleet will proxy to.
package fleet

import "net/http"

func NewRouter[Config any, Entry Instance](
	urlPrefix string,
	registry *Registry[Config, Entry],
) (http.Handler, error) {
	router := http.NewServeMux()
	router.Handle("/", http.StripPrefix(urlPrefix, registry))
	return router, nil
}

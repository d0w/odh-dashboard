package fleet

import "net/http"

func NewRouter(urlPrefix string, registry Registry) (http.Handler, error) {
	router := http.NewServeMux()
	router.Handle("/", http.StripPrefix(urlPrefix, registry))
	return router, nil
}

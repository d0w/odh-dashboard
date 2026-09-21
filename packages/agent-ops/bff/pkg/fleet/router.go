package fleet

import "net/http"

func NewRouter(urlPrefix string, registry Registry) (http.Handler, error) {
	router := http.NewServeMux()
	// router.Handle(urlPrefix+"/", http.StripPrefix(urlPrefix, registry))
	router.Handle("/", registry)
	return router, nil
}

package http

import (
	"net/http"
)

func EnforeMethod(method string, handler http.Handler) http.Handler {
	return &httpMethodCheckWrapper{method, handler}
}

type httpMethodCheckWrapper struct {
	AllowedMethod string
	Next          http.Handler
}

func (h *httpMethodCheckWrapper) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != h.AllowedMethod {
		resp.WriteHeader(http.StatusMethodNotAllowed)

		resp.Write([]byte("Method not allowed.\n"))
	} else {
		h.Next.ServeHTTP(resp, req)
	}
}

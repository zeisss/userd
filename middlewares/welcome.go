package middlewares

import (
	"net/http"

	httputil "../http"
)

// --------------------------------------------------------------------------------------------

type WelcomeHandler struct{}

func (h WelcomeHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	result := struct {
		Message string `json:"message"`
	}{
		"Welcome! This is userd.",
	}
	httputil.WriteJSONResponse(resp, http.StatusOK, result)
}

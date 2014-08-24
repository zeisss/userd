package http

import (
	"encoding/json"
	"net/http"
)

func WriteBadRequest(resp http.ResponseWriter, req *http.Request, msg ...string) {
	resp.WriteHeader(http.StatusBadRequest)

	for _, s := range msg {
		resp.Write([]byte(s))
	}
}

func WriteNotFound(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusNotFound)
}

func WriteNoContent(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusNoContent)
}

func WriteJSONResponse(resp http.ResponseWriter, code int, data interface{}) {
	resp.WriteHeader(code)
	resp.Header().Add("content-Type", "application/json; charset=UTF8")

	if err := json.NewEncoder(resp).Encode(data); err != nil {
		panic(err)
	}
}

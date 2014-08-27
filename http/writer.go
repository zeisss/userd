package http

import (
	"encoding/json"
	"net/http"
)

func WriteBadRequest(resp http.ResponseWriter, req *http.Request, msg ...string) {
	if len(msg) == 0 {
		WriteJSONErrorPage(resp, http.StatusBadRequest, "Bad request.")
	} else if len(msg) == 1 {
		WriteJSONErrorPage(resp, http.StatusBadRequest, msg[0])
	} else {
		panic("To many arguments to WriteBadRequest() only 0,1 supported.")
	}
}

func WriteNotFound(resp http.ResponseWriter) {
	WriteJSONErrorPage(resp, http.StatusNotFound, "Resource not found.")

}

func WriteNoContent(resp http.ResponseWriter) {
	resp.WriteHeader(http.StatusNoContent)
}

func WriteJSONErrorPage(resp http.ResponseWriter, code int, message string) {
	body := map[string]string{
		"msg": message,
	}

	WriteJSONResponse(resp, code, body)
}

func WriteJSONResponse(resp http.ResponseWriter, code int, data interface{}) {
	resp.WriteHeader(code)
	resp.Header().Add("content-Type", "application/json; charset=UTF8")

	if err := json.NewEncoder(resp).Encode(data); err != nil {
		panic(err)
	}
}

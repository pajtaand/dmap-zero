package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string
}

func WriteResponse(w http.ResponseWriter, code int, resp interface{}) {
	w.WriteHeader(code)
	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
	}
}

func WriteErrorResponse(w http.ResponseWriter, code int, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	http.Error(w, errMsg, code)
}

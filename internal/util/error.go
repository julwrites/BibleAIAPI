package util

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func JSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response := ErrorResponse{}
	response.Error.Code = code
	response.Error.Message = message
	json.NewEncoder(w).Encode(response)
}

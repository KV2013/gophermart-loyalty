package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KV2013/gophermart-loyalty/internal/model"
)

func writeJSONError(res http.ResponseWriter, err error, status int) error {
	body, err := json.Marshal(model.APIErrorResponse{Error: err.Error(), StatusCode: status})
	if err != nil {
		return err
	}
	res.WriteHeader(status)
	if _, err := res.Write(body); err != nil {
		return err
	}

	return nil
}

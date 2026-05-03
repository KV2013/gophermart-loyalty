package model

type APIErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code"`
}

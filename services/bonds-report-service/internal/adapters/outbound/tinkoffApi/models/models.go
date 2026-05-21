package models

type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

func NewHTTPResponse(statusCode int, body []byte) *HTTPResponse {
	return &HTTPResponse{
		StatusCode: statusCode,
		Body:       body,
	}
}

package handlers

import "errors"

var (
	errInvalidRequestBody error = errors.New("invalid request body")
	errGetData            error = errors.New("could not get data")
	errSaveMessage        error = errors.New("internal server error")
	errEmptyOffset        error = errors.New("offset is empty")
	errEmptyLimit         error = errors.New("limit is empty")
	errEmptyChatID        error = errors.New("chatID is empty")
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type MediaItem struct {
	Type    string `json:"type"`
	Media   string `json:"media"`
	Caption string `json:"caption,omitempty"`
}

type UploadedImage struct {
	FieldName string
	FileName  string
	Data      []byte
}


type SendPhotoRequest struct {
	ChatID   int
	Caption  string
	FileName string
	Photo    []byte
}
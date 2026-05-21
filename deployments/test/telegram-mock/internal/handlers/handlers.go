package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"telegram-mock/internal/service"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	offsetValue = "offset"
	limitValue  = "limit"
	chatIDValue = "chat_id"
	textValue   = "text"
	mediaVlaue  = "media"
)

type Handler struct {
	logger  *slog.Logger
	service Service
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=Service
type Service interface {
	Updates(ctx context.Context, offset int, limit int) []service.Update
	SaveMessage(ctx context.Context, message service.IncomingMessage)
}

func NewHandler(logger *slog.Logger, srvc Service) *Handler {
	return &Handler{
		logger:  logger,
		service: srvc,
	}
}

func (h *Handler) GetUpdates(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	offset, err := validateOffset(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, service.UpdatesResponse{Ok: false, Result: []service.Update{}})
	}

	limit, err := validateLimit(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, service.UpdatesResponse{Ok: false, Result: []service.Update{}})
	}

	updates := h.service.Updates(ctx, offset, limit)

	UpdatesResponse := service.UpdatesResponse{
		Ok:     true,
		Result: updates,
	}

	return c.JSON(http.StatusOK, UpdatesResponse)
}

func (h *Handler) PostMessage(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var message service.IncomingMessage

	err := c.Bind(&message)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	h.service.SaveMessage(ctx, message)

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) SendMessage(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	text := c.QueryParams().Get(textValue)

	chatID, err := validateChatID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	h.logger.DebugContext(ctx, "send message request", slog.Int("chatID", chatID), slog.String("text", text))

	return c.JSON(http.StatusOK, "success")
}

func (h *Handler) SendPhoto(c echo.Context) error {
	req, err := h.parseSendPhotoRequest(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	h.logger.Info("photo received",
		"chat_id", req.ChatID,
		"caption", req.Caption,
		"file_name", req.FileName,
		"photo_size", len(req.Photo),
	)

	return c.JSON(http.StatusOK, map[string]any{
		"ok": true,
		"result": map[string]any{
			"chat_id":    req.ChatID,
			"caption":    req.Caption,
			"file_name":  req.FileName,
			"photo_size": len(req.Photo),
		},
	})
}

func (h *Handler) parseSendPhotoRequest(c echo.Context) (SendPhotoRequest, error) {
	var req SendPhotoRequest

	chatIDStr := c.FormValue("chat_id")
	if chatIDStr == "" {
		return req, echo.NewHTTPError(http.StatusBadRequest, "chat_id is required")
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, "invalid chat_id")
	}
	req.ChatID = chatID

	req.Caption = c.FormValue("caption")

	fileHeader, err := c.FormFile("photo")
	if err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, "photo is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, "can't open photo")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, "can't read photo")
	}

	req.FileName = fileHeader.Filename
	req.Photo = data

	return req, nil
}

func (h *Handler) SendMediaGroup(c echo.Context) error {
	const op = "handlers.SendMediaGroup"

	chatIDStr := c.FormValue("chat_id")
	if chatIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "chat_id is required")
	}

	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid chat_id")
	}

	mediaStr := c.FormValue(mediaVlaue)
	if mediaStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "media is required")
	}

	var media []MediaItem
	if err := json.Unmarshal([]byte(mediaStr), &media); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid media json")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid multipart form")
	}

	var images []UploadedImage

	for fieldName, files := range form.File {
		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "can't open uploaded file")
			}

			data, readErr := io.ReadAll(f)
			_ = f.Close()
			if readErr != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "can't read uploaded file")
			}

			images = append(images, UploadedImage{
				FieldName: fieldName,
				FileName:  fh.Filename,
				Data:      data,
			})
		}
	}

	h.logger.Info("received media group",
		slog.Int("chat_id", chatID),
		slog.Int("media_count", len(media)),
		slog.Int("files_count", len(images)),
	)

	return c.JSON(http.StatusOK, map[string]any{
		"ok":      true,
		"chat_id": chatID,
		"media":   media,
		"files":   len(images),
	})
}

func validateOffset(c echo.Context) (int, error) {
	offset := c.QueryParams().Get(offsetValue)
	if offset == "" {
		return 0, errEmptyOffset
	}
	res, err := strconv.Atoi(offset)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func validateLimit(c echo.Context) (int, error) {
	limit := c.QueryParams().Get(limitValue)
	if limit == "" {
		return 0, errEmptyLimit
	}
	res, err := strconv.Atoi(limit)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func validateChatID(c echo.Context) (int, error) {
	chatID := c.QueryParams().Get(chatIDValue)

	if chatID == "" {
		return 0, errEmptyChatID
	}
	res, err := strconv.Atoi(chatID)
	if err != nil {
		return 0, err
	}
	return res, nil
}

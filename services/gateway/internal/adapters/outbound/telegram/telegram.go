package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gladinov/e"
	bondreportservice "main.go/internal/adapters/outbound/bondReportService"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	logger   *slog.Logger
	host     string
	basePath string
	client   http.Client
}

const (
	getUpdatesMethod     = "getUpdates"
	sendUpdateMethod     = "sendMessage"
	sendPhotoMethod      = "sendPhoto"
	sendMediaGroupMethod = "sendMediaGroup"
)

func New(logger *slog.Logger, host string, token string) *Client {
	return &Client{
		logger:   logger,
		host:     host,
		basePath: newBasePath(token),
		client:   http.Client{Timeout: defaultTimeout},
	}
}

func newBasePath(token string) string {
	return "bot" + token
}

func (c *Client) Updates(ctx context.Context, offset int, limit int) (updates []Update, err error) {
	defer func() { err = e.WrapIfErr("can`t get updates", err) }()

	const op = "telegram.Updates"

	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(ctx, getUpdatesMethod, q)
	if err != nil {
		return nil, err
	}

	var res UpdatesResponse

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int, text string) error {
	const op = "telegram.SendMessage"
	logg := c.logger.With(
		slog.String("op", op),
	)
	defer func() { logg.DebugContext(ctx, "success") }()
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)

	_, err := c.doRequest(ctx, sendUpdateMethod, q)
	if err != nil {
		return e.Wrap("can`t send message", err)
	}
	return nil
}

func (c *Client) doRequest(ctx context.Context, method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("can`t do request", err) }()

	const op = "telegram.doRequest"

	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, parseErr(err, c.basePath)
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func parseErr(err error, token string) error {
	if err == nil {
		return nil
	}
	const hiddenToken = "bot_hidden_token"
	msg := strings.ReplaceAll(err.Error(), token, hiddenToken)
	return errors.New(msg)
}

func (c *Client) SendImageFromBuffer(ctx context.Context, chatID int, imageData []byte, caption string) error {
	const op = "telegram.SendImageFromBuffer"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("chat_id", strconv.Itoa(chatID))
	if caption != "" {
		writer.WriteField("caption", caption)
	}

	part, err := writer.CreateFormFile("photo", "image.png")
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, bytes.NewReader(imageData)); err != nil {
		return err
	}

	writer.Close()

	_, err = c.doMultipartRequest(ctx, "sendPhoto", body, writer.FormDataContentType())
	return err
}

func (c *Client) SendMediaGroupFromBuffer(ctx context.Context, chatID int, images []*bondreportservice.ImageData) error {
	const op = "telegram.SendMediaGroupFromBuffer"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	if len(images) == 0 {
		return errors.New("no images to send")
	}

	// Ограничение Telegram
	if len(images) > 10 {
		images = images[:10]
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем chat_id
	writer.WriteField("chat_id", strconv.Itoa(chatID))

	// Подготавливаем медиа-группу
	media := make([]map[string]string, len(images))
	for i, img := range images {
		media[i] = map[string]string{
			"type":  "photo",
			"media": "attach://image_" + strconv.Itoa(i),
		}

		// Подпись только для первого изображения
		if i == 0 && img.Caption != "" {
			media[i]["caption"] = img.Caption
		}
	}

	mediaJSON, _ := json.Marshal(media)
	writer.WriteField("media", string(mediaJSON))

	// Добавляем изображения из буферов
	for i, img := range images {
		part, err := writer.CreateFormFile("image_"+strconv.Itoa(i), img.Name)
		if err != nil {
			return fmt.Errorf("can't create form file: %v", err)
		}

		// Копируем байты из буфера
		if _, err := io.Copy(part, bytes.NewReader(img.Data)); err != nil {
			return fmt.Errorf("can't copy image data: %v", err)
		}
	}

	writer.Close()

	_, err := c.doMultipartRequest(ctx, "sendMediaGroup", body, writer.FormDataContentType())
	if err != nil {
		return fmt.Errorf("can't send media group: %v", err)
	}

	return nil
}

func (c *Client) doMultipartRequest(ctx context.Context, method string, body *bytes.Buffer, contentType string) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("can't do multipart request", err) }()

	const op = "telegram.doMultipartRequest"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, parseErr(err, c.basePath)
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram API error: %s - %s", resp.Status, string(bodyResponse))
	}

	return bodyResponse, nil
}

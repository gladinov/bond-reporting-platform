package transport

import (
	"bonds-report-service/internal/adapters/outbound/moex/models"
	"bonds-report-service/internal/utils/logging"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	httpheaders "github.com/gladinov/contracts/http"
	"github.com/gladinov/contracts/trace"
	"github.com/gladinov/e"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TransportClient
type TransportClient interface {
	DoRequest(ctx context.Context, path string, query url.Values, requestBody io.Reader) (*models.HTTPResponse, error)
}

type Transport struct {
	logger *slog.Logger
	host   string
	client *http.Client
}

func NewTransport(logger *slog.Logger, host string, timeout time.Duration) *Transport {
	return &Transport{
		logger: logger,
		host:   host,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (t *Transport) DoRequest(ctx context.Context,
	path string,
	query url.Values,
	requestBody io.Reader,
) (_ *models.HTTPResponse, err error) {
	const op = "transport.doRequest"
	logg := t.logger.With()
	defer logging.LogOperation_Debug(ctx, logg, op, &err)()

	u := url.URL{
		Scheme: "http",
		Host:   t.host,
		Path:   path,
	}

	var method string
	if requestBody != nil {
		method = http.MethodPost
	} else {
		method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), requestBody)
	if err != nil {
		errMsg := "could not create http.NewRequest"
		logging.LoggHTTPError(ctx, logg, req, errMsg, op, err)
		return nil, e.WrapIfErr(errMsg, err)
	}

	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", "application/json")
	reqWithTraceID := t.setHeaders(ctx, req)

	response, err := t.client.Do(reqWithTraceID)
	if err != nil {
		errMsg := "could not do request"
		logging.LoggHTTPError(ctx, logg, req, errMsg, op, err)
		return nil, e.WrapIfErr(errMsg, err)
	}

	if response != nil {
		defer response.Body.Close()
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		errMsg := "could not read body"
		logging.LoggHTTPError(ctx, logg, req, errMsg, op, err)
		return nil, e.WrapIfErr(errMsg, err)
	}

	HTTPResponse := models.NewHTTPResponse(response.StatusCode, body)

	return HTTPResponse, nil
}

func (t *Transport) setHeaders(ctx context.Context, req *http.Request) *http.Request {
	const op = "transport.SetHeaders"
	logg := t.logger.With(slog.String("op", op))

	traceID, ok := trace.TraceIDFromContext(ctx)
	if !ok {
		logg.WarnContext(ctx, "hasn't traceID in ctx")
	}
	req.Header.Set(httpheaders.HeaderTraceID, traceID)

	return req
}

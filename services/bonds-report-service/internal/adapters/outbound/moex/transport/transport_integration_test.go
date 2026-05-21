//go:build integration

package transport

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransport_DoRequest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	t.Run("Success GET returns 200", func(t *testing.T) {
		// Поднимаем тестовый HTTP-сервер
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"ok"}`))
		}))
		defer ts.Close()

		tr := &Transport{
			logger: logger,
			host:   ts.Listener.Addr().String(),
			client: ts.Client(),
		}

		ctx := context.Background()
		resp, err := tr.DoRequest(ctx, "/test", url.Values{}, bytes.NewBuffer(nil))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, string(resp.Body), "ok")
	})

	t.Run("Error: invalid URL", func(t *testing.T) {
		tr := NewTransport(logger, "http://[::1]:0", 10*time.Second)
		_, err := tr.DoRequest(ctx, "/test", url.Values{}, nil)
		assert.Error(t, err)
	})

	t.Run("Error: server returns 500", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`internal error`))
		}))
		defer ts.Close()

		tr := &Transport{
			logger: logger,
			host:   ts.Listener.Addr().String(),
			client: ts.Client(),
		}

		resp, err := tr.DoRequest(ctx, "/test", url.Values{}, nil)
		assert.NoError(t, err) // DoRequest возвращает HTTPResponse даже при 500
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "internal error")
	})

	t.Run("POST request body", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, `{"foo":"bar"}`, string(body))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message":"received"}`))
		}))
		defer ts.Close()

		tr := &Transport{
			logger: logger,
			host:   ts.Listener.Addr().String(),
			client: ts.Client(),
		}

		body := bytes.NewBufferString(`{"foo":"bar"}`)

		resp, err := tr.DoRequest(ctx, "/test", url.Values{}, body)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "received")
	})
}

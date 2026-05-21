//go:build unit

package moex

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"moex/internal/clients/moex/mocks"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoRequest_Realistic(t *testing.T) {
	moexHost := "iss.moex.com"
	ctx := context.Background()
	logg := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Success", func(t *testing.T) {
		mockBody := "<ValCurs>data</ValCurs>"
		client := mocks.NewMockClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockBody)),
			}, nil
		})

		transport := &Transport{
			logger: logg,
			host:   moexHost,
			client: &client,
		}
		body, err := transport.DoRequest(ctx, "/mock", url.Values{})
		require.NoError(t, err)
		require.Equal(t, mockBody, string(body))
	})

	t.Run("Network error (client.Do)", func(t *testing.T) {
		client := mocks.NewMockClient(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network unreachable")
		})

		transport := &Transport{logger: logg, host: moexHost, client: &client}
		_, err := transport.DoRequest(ctx, "/mock", url.Values{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "network unreachable")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		transport := &Transport{
			logger: logg,
			host:   "://bad_host",
			client: &http.Client{},
		}
		_, err := transport.DoRequest(ctx, "/mock", url.Values{})
		require.Error(t, err)
		require.ErrorContains(t, err, "could not create http.NewRequest")
	})

	t.Run("Error reading body", func(t *testing.T) {
		client := mocks.NewMockClient(func(req *http.Request) (*http.Response, error) {
			badReader := &faultyReader{}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(badReader),
			}, nil
		})

		transport := &Transport{logger: logg, host: moexHost, client: &client}
		_, err := transport.DoRequest(ctx, "/mock", url.Values{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "could not read body")
	})
}

type faultyReader struct{}

func (f *faultyReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

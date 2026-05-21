package telegram

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseErr(t *testing.T) {
	const hiddenToken = "bot_hidden_token"

	t.Run("success", func(t *testing.T) {
		in := `Get "https://api.telegram.org/bot_for_test/getUpdates?limit=100&offset=0": dial tcp 149.154.166.110:443: i/o timeout`
		want := `Get "https://api.telegram.org/bot_hidden_token/getUpdates?limit=100&offset=0": dial tcp 149.154.166.110:443: i/o timeout`
		token := "bot_for_test"
		err := errors.New(in)
		newErr := parseErr(err, token)

		got := newErr.Error()
		require.Equal(t, want, got)
		t.Log(got)
	})

	t.Run("nil", func(t *testing.T) {
		token := "bot_for_test"
		var err error
		got := parseErr(err, token)

		require.Equal(t, err, got)
	})

	t.Run("no token in err", func(t *testing.T) {
		in := `Get "https://api.telegram.org//getUpdates?limit=100&offset=0": dial tcp 149.154.166.110:443: i/o timeout`
		want := `Get "https://api.telegram.org//getUpdates?limit=100&offset=0": dial tcp 149.154.166.110:443: i/o timeout`
		token := "bot_for_test"
		err := errors.New(in)
		newErr := parseErr(err, token)

		got := newErr.Error()
		require.Equal(t, want, got)
		t.Log(got)
	})
}

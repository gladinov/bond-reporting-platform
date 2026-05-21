package tokenauth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gladinov/cryptotoken"
	"github.com/gladinov/e"
	"github.com/gladinov/valuefromcontext"
	"github.com/redis/go-redis/v9"
	"main.go/internal/adapters/outbound/tinkoffApi"
	storage "main.go/internal/repository"
)

type TokenStatus int

const (
	TokenError TokenStatus = iota
	TokenFound
	TokenInserted
)

var ErrIncorrectToken = errors.New("incorrect token")

type TokenAuthService struct {
	logger       *slog.Logger
	redis        *redis.Client
	storage      storage.Storage
	tinkoffApi   *tinkoffApi.Client
	tokenCrypter *cryptotoken.TokenCrypter
}

func NewTokenAuthService(logger *slog.Logger,
	redis *redis.Client,
	storage storage.Storage,
	tinkoffApi *tinkoffApi.Client,
	tokenCrypter *cryptotoken.TokenCrypter,
) *TokenAuthService {
	return &TokenAuthService{
		logger:       logger,
		redis:        redis,
		storage:      storage,
		tinkoffApi:   tinkoffApi,
		tokenCrypter: tokenCrypter,
	}
}

func (t *TokenAuthService) Auth(ctx context.Context, text string, username string) (TokenStatus, error) {
	const op = "telegram.auth"

	logg := t.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished")
	}()

	chatIDStr, err := valuefromcontext.GetChatIDFromCtxStr(ctx)
	if err != nil {
		return TokenError, fmt.Errorf("%s:%w", op, err)
	}

	haveTokenInRedisErr := t.redis.Get(ctx, chatIDStr).Err()
	if haveTokenInRedisErr != nil && haveTokenInRedisErr != redis.Nil {
		return TokenError, fmt.Errorf("%s:%w", op, err)
	}
	if haveTokenInRedisErr == nil {
		logg.DebugContext(ctx, "token found in redis")
		return TokenFound, nil
	}

	haveToken, err := t.checkUserToken(ctx)
	if err != nil {
		return TokenError, fmt.Errorf("%s:%w", op, err)
	}

	switch haveToken {
	case true:
		logg.InfoContext(ctx, "existing user token reused")
		tokenInBase64, err := t.storage.PickToken(ctx)
		if err != nil {
			return TokenError, fmt.Errorf("%s:%w", op, err)
		}

		err = t.cacheToken(ctx, chatIDStr, tokenInBase64)
		if err != nil {
			return TokenError, fmt.Errorf("%s:%w", op, err)
		}
		logg.InfoContext(ctx, "token stored and cached")
		return TokenInserted, nil

	case false:
		logg.InfoContext(ctx, "user provided new token")
		err := t.isToken(ctx, text)
		if err != nil {
			return TokenError, ErrIncorrectToken
		}
		tokenInBase64, err := t.tokenToBase64(text)
		if err != nil {
			return TokenError, fmt.Errorf("%s:%w", op, err)
		}
		err = t.storage.Save(ctx, username, tokenInBase64)
		if err != nil {
			return TokenError, fmt.Errorf("%s:%w", op, err)
		}
		err = t.cacheToken(ctx, chatIDStr, tokenInBase64)
		if err != nil {
			return TokenError, fmt.Errorf("%s:%w", op, err)
		}
		logg.InfoContext(ctx, "token stored and cached")
		return TokenInserted, nil
	}
	return TokenFound, nil
}

func (t *TokenAuthService) isToken(ctx context.Context, text string) error {
	const op = "telegram.isToken"

	chatID, err := valuefromcontext.GetChatIDFromCtxStr(ctx)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	log := t.logger.With(
		slog.String("op", op),
		slog.String("chat_id", chatID),
	)

	if len(text) == 88 { // TODO:модифицировать проверку
		tokenInBase64, err := t.tokenToBase64(text)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		err = t.tinkoffApi.CheckToken(ctx, tokenInBase64)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}

		log.Info("token validated")
		return nil
	}

	return errors.New("is not token")
}

func (t *TokenAuthService) checkUserToken(ctx context.Context) (res bool, err error) {
	const op = "processor:checkUserToken"

	isExists, err := t.storage.IsExistsToken(ctx)
	if err != nil {
		return false, err
	}

	return isExists, nil
}

func (t *TokenAuthService) tokenToBase64(token string) (string, error) {
	const op = "telegram.tokenToBase64"
	logg := t.logger.With(slog.String("op", op))
	encryptedToken, err := t.tokenCrypter.EncryptToken(token)
	if err != nil {
		logg.Debug("err encrypt token", slog.Any("error", err))
		return "", e.WrapIfErr("could not encrypt token", err)
	}
	tokenInBase64, err := encryptedToken.ToBase64()
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}
	return tokenInBase64, nil
}

func (t *TokenAuthService) cacheToken(ctx context.Context, chatID, token string) error {
	expiry := time.Until(time.Now().AddDate(5, 0, 0))
	return t.redis.Set(ctx, chatID, token, expiry).Err()
}

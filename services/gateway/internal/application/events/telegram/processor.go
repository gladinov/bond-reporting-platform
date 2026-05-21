package telegram

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gladinov/e"
	bondreportservice "main.go/internal/adapters/outbound/bondReportService"
	"main.go/internal/adapters/outbound/telegram"
	"main.go/internal/adapters/outbound/tinkoffApi"
	"main.go/internal/application/events"
	tokenauth "main.go/internal/application/tokenAuth"
)

type Processor struct {
	logger            *slog.Logger
	tg                MessageSender
	tinkoffApi        *tinkoffApi.Client
	bondReportService *bondreportservice.Client
	tokenAuthService  *tokenauth.TokenAuthService
	kafka             Producer
}

type MessageSender interface {
	SendImageFromBuffer(ctx context.Context, chatID int, imageData []byte, caption string) error
	SendMediaGroupFromBuffer(ctx context.Context, chatID int, images []*bondreportservice.ImageData) error
	SendMessage(ctx context.Context, chatID int, text string) error
}

type Producer interface {
	PublishRequest(
		ctx context.Context,
		reportKind string,
		chatID string,
		traceID string,
	) error
}

type Meta struct {
	ChatID   int
	Username string
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

func NewProccesor(
	logger *slog.Logger,
	client *telegram.Client,
	tinkoffApiClient *tinkoffApi.Client,
	bondReportServiceClient *bondreportservice.Client,
	tokenAuthService *tokenauth.TokenAuthService,
	kafka Producer,
) *Processor {
	return &Processor{
		logger:            logger,
		tg:                client,
		tinkoffApi:        tinkoffApiClient,
		bondReportService: bondReportServiceClient,
		tokenAuthService:  tokenAuthService,
		kafka:             kafka,
	}
}

func (p *Processor) Process(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(ctx, event)
	default:
		return e.Wrap("can't process message", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(ctx context.Context, event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	if err := p.doCmd(ctx, event.Text, meta.ChatID, meta.Username); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}
	return res, nil
}

package bondreportservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"time"

	httpheaders "github.com/gladinov/contracts/http"
	trace "github.com/gladinov/contracts/trace"

	"github.com/gladinov/valuefromcontext"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	logger *slog.Logger
	host   string
	client http.Client
}

func New(logger *slog.Logger, host string) *Client {
	return &Client{
		logger: logger,
		host:   host,
		client: http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) GetAccountsList(ctx context.Context) (AccountListResponce, error) {
	const op = "bondreportservice.GetAccountsList"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	pth := path.Join("bondReportService", "accounts")
	u := url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   pth,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return AccountListResponce{}, fmt.Errorf("%s: %w", op, err)
	}

	reqWithHeaders, err := c.setHeaders(ctx, req)
	if err != nil {
		return AccountListResponce{}, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(reqWithHeaders)
	if err != nil {
		return AccountListResponce{}, fmt.Errorf("%s: %w", op, err)
	}

	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountListResponce{}, fmt.Errorf("%s: %w", op, err)
	}
	if resp.StatusCode != http.StatusOK {
		var statusErr map[string]string
		err = json.Unmarshal(body, &statusErr)
		if err != nil {
			return AccountListResponce{}, fmt.Errorf("%s: %w", op, err)
		}
		return AccountListResponce{}, fmt.Errorf("%s:"+statusErr["error"], op)
	}

	var accountResponce AccountListResponce
	err = json.Unmarshal(body, &accountResponce)
	if err != nil {
		return AccountListResponce{}, fmt.Errorf("%s: %w", op, err)
	}
	return accountResponce, nil
}

func (c *Client) GetUsd(ctx context.Context) (UsdResponce, error) {
	const op = "bondreportservice.GetUsd"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	pth := path.Join("bondReportService", "getUSD")

	u := url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   pth,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return UsdResponce{}, fmt.Errorf("%s:%w", op, err)
	}

	reqWithHeaders, err := c.setHeaders(ctx, req)
	if err != nil {
		return UsdResponce{}, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(reqWithHeaders)
	if err != nil {
		return UsdResponce{}, fmt.Errorf("%s:%w", op, err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UsdResponce{}, fmt.Errorf("%s:%w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		var statusErr map[string]string
		err := json.Unmarshal(body, &statusErr)
		if err != nil {
			return UsdResponce{}, fmt.Errorf("%s:%w", op, err)
		}
		return UsdResponce{}, fmt.Errorf("%s:"+statusErr["error"], op)
	}

	var usdResponce UsdResponce
	err = json.Unmarshal(body, &usdResponce)
	if err != nil {
		return UsdResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	return usdResponce, nil
}

func (c *Client) GetBondReportsByFifo(ctx context.Context) error {
	const op = "bondreportservice.GetBondReportsByFifo"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	pth := path.Join("bondReportService", "getBondReportsByFifo")
	u := url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   pth,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	reqWithHeaders, err := c.setHeaders(ctx, req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(reqWithHeaders)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	if resp.StatusCode != http.StatusNoContent {
		var statusErr map[string]string
		err := json.Unmarshal(body, &statusErr)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
		return fmt.Errorf("%s:"+statusErr["error"], op)
	}
	return nil
}

func (c *Client) GetBondReports(ctx context.Context) (BondReportsResponce, error) {
	const op = "bondreportservice.GetBondReports"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	pth := path.Join("bondReportService", "getBondReports")
	u := url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   pth,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return BondReportsResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	reqWithHeaders, err := c.setHeaders(ctx, req)
	if err != nil {
		return BondReportsResponce{}, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(reqWithHeaders)
	if err != nil {
		return BondReportsResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return BondReportsResponce{}, fmt.Errorf("%s:%w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		var statusErr map[string]string
		err := json.Unmarshal(body, &statusErr)
		if err != nil {
			return BondReportsResponce{}, fmt.Errorf("%s:%w", op, err)
		}
		return BondReportsResponce{}, fmt.Errorf("%s:"+statusErr["error"], op)
	}
	var bondReportResponce BondReportsResponce
	err = json.Unmarshal(body, &bondReportResponce)
	if err != nil {
		return BondReportsResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	return bondReportResponce, nil
}

func (c *Client) GetPortfolioStructure(ctx context.Context) (PortfolioStructureForEachAccountResponce, error) {
	const op = "bondreportservice.GetPortfolioStructure"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	pth := path.Join("bondReportService", "getPortfolioStructure")
	u := url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   pth,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return PortfolioStructureForEachAccountResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	reqWithHeaders, err := c.setHeaders(ctx, req)
	if err != nil {
		return PortfolioStructureForEachAccountResponce{}, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(reqWithHeaders)
	if err != nil {
		return PortfolioStructureForEachAccountResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PortfolioStructureForEachAccountResponce{}, fmt.Errorf("%s:%w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		var statusErr map[string]string
		err := json.Unmarshal(body, &statusErr)
		if err != nil {
			return PortfolioStructureForEachAccountResponce{}, fmt.Errorf("%s:%w", op, err)
		}
		return PortfolioStructureForEachAccountResponce{}, fmt.Errorf("%s:"+statusErr["error"], op)
	}
	var bondReportResponce PortfolioStructureForEachAccountResponce
	err = json.Unmarshal(body, &bondReportResponce)
	if err != nil {
		return PortfolioStructureForEachAccountResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	return bondReportResponce, nil
}

func (c *Client) GetUnionPortfolioStructure(ctx context.Context) (UnionPortfolioStructureResponce, error) {
	const op = "bondreportservice.GetUnionPortfolioStructure"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	pth := path.Join("bondReportService", "getUnionPortfolioStructure")
	u := url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   pth,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return UnionPortfolioStructureResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	reqWithHeaders, err := c.setHeaders(ctx, req)
	if err != nil {
		return UnionPortfolioStructureResponce{}, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(reqWithHeaders)
	if err != nil {
		return UnionPortfolioStructureResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UnionPortfolioStructureResponce{}, fmt.Errorf("%s:%w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		var statusErr map[string]string
		err := json.Unmarshal(body, &statusErr)
		if err != nil {
			return UnionPortfolioStructureResponce{}, fmt.Errorf("%s:%w", op, err)
		}
		return UnionPortfolioStructureResponce{}, fmt.Errorf("%s:"+statusErr["error"], op)
	}
	var bondReportResponce UnionPortfolioStructureResponce
	err = json.Unmarshal(body, &bondReportResponce)
	if err != nil {
		return UnionPortfolioStructureResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	return bondReportResponce, nil
}

func (c *Client) GetUnionPortfolioStructureWithSber(ctx context.Context) (UnionPortfolioStructureWithSberResponce, error) {
	const op = "bondreportservice.GetUnionPortfolioStructureWithSber"

	start := time.Now()
	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	pth := path.Join("bondReportService", "getUnionPortfolioStructureWithSber")
	u := url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   pth,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return UnionPortfolioStructureWithSberResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	reqWithHeaders, err := c.setHeaders(ctx, req)
	if err != nil {
		return UnionPortfolioStructureWithSberResponce{}, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.client.Do(reqWithHeaders)
	if err != nil {
		return UnionPortfolioStructureWithSberResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UnionPortfolioStructureWithSberResponce{}, fmt.Errorf("%s:%w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		var statusErr map[string]string
		err := json.Unmarshal(body, &statusErr)
		if err != nil {
			return UnionPortfolioStructureWithSberResponce{}, fmt.Errorf("%s:%w", op, err)
		}
		return UnionPortfolioStructureWithSberResponce{}, fmt.Errorf("%s:"+statusErr["error"], op)
	}
	var bondReportResponce UnionPortfolioStructureWithSberResponce
	err = json.Unmarshal(body, &bondReportResponce)
	if err != nil {
		return UnionPortfolioStructureWithSberResponce{}, fmt.Errorf("%s:%w", op, err)
	}
	return bondReportResponce, nil
}

func (c *Client) setHeaders(ctx context.Context, req *http.Request) (*http.Request, error) {
	const op = "bondreportservice.SetHeaders"

	logg := c.logger.With(slog.String("op", op))
	logg.DebugContext(ctx, "start")
	defer func() {
		logg.InfoContext(ctx, "finished")
	}()

	chatID, err := valuefromcontext.GetChatIDFromCtxStr(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	req.Header.Set(httpheaders.HeaderChatID, chatID)
	traceID, ok := trace.TraceIDFromContext(ctx)
	if !ok {
		logg.WarnContext(ctx, "hasn't traceID in ctx")
	}
	req.Header.Set(httpheaders.HeaderTraceID, traceID)

	return req, nil
}

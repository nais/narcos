package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	genqlientgraphql "github.com/Khan/genqlient/graphql"
)

const (
	defaultEndpoint = "https://fasit.nais.io/query"
)

const (
	maxQueryRecords = 200
	cacheTTL        = time.Minute
)

var desktopOAuth struct {
	clientID         string
	clientSecret     string
	tokenEndpoint    string
	loadRefreshToken func() (string, error)
}

type tokenResponse struct {
	IDToken          string `json:"id_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func ConfigureDesktopOAuth(clientID, clientSecret, tokenEndpoint string, loadRefreshToken func() (string, error)) {
	desktopOAuth.clientID = clientID
	desktopOAuth.clientSecret = clientSecret
	desktopOAuth.tokenEndpoint = tokenEndpoint
	desktopOAuth.loadRefreshToken = loadRefreshToken
}

func DefaultEndpoint() string {
	return defaultEndpoint
}

func MaxQueryRecords() int {
	return maxQueryRecords
}

func CacheTTL() time.Duration {
	return cacheTTL
}

type iapTransport struct {
	wrapped http.RoundTripper

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func newIAPTransport() http.RoundTripper {
	return &iapTransport{
		wrapped: http.DefaultTransport,
	}
}

func (t *iapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.tokenFor(req.Context())
	if err != nil {
		return nil, err
	}

	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+token)
	return t.wrapped.RoundTrip(req)
}

func (t *iapTransport) tokenFor(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.token != "" && time.Until(t.expiresAt) > time.Minute {
		return t.token, nil
	}

	token, err := t.fetchToken(ctx)
	if err != nil {
		return "", err
	}

	t.token = token
	t.expiresAt = time.Now().Add(45 * time.Minute)
	return t.token, nil
}

func (t *iapTransport) fetchToken(ctx context.Context) (string, error) {
	if desktopOAuth.loadRefreshToken == nil || desktopOAuth.tokenEndpoint == "" {
		return "", fmt.Errorf("desktop OAuth is not configured")
	}

	refreshToken, err := desktopOAuth.loadRefreshToken()
	if err != nil {
		return "", fmt.Errorf("load stored Fasit refresh token: %w", err)
	}
	if refreshToken == "" {
		return "", fmt.Errorf("not logged in to Fasit — run: narc fasit login")
	}

	values := url.Values{
		"client_id":     {desktopOAuth.clientID},
		"client_secret": {desktopOAuth.clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, desktopOAuth.tokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("create OAuth refresh request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := (&http.Client{Transport: t.wrapped}).Do(request)
	if err != nil {
		return "", fmt.Errorf("exchange refresh token for ID token: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read OAuth refresh response: %w", err)
	}

	var payload tokenResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode OAuth refresh response: %w", err)
	}

	if payload.Error != "" {
		return "", fmt.Errorf("OAuth refresh failed: %s (%s)", payload.Error, payload.ErrorDescription)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("OAuth refresh failed with status %s", response.Status)
	}

	if payload.IDToken == "" {
		return "", fmt.Errorf("OAuth refresh response did not include an id_token")
	}

	return payload.IDToken, nil
}

type QueryRecord struct {
	OpName    string
	Duration  time.Duration
	Error     string
	Timestamp time.Time
	CacheHit  bool
}

type QueryLogger interface {
	QueryLog() []QueryRecord
}

type CacheInfo struct {
	Hits   uint64
	Misses uint64
}

type CacheEntryDetail struct {
	Key       string
	OpName    string
	Age       time.Duration
	TTL       time.Duration
	SizeBytes int
}

type CacheInspector interface {
	CacheEntries() []CacheEntryDetail
	CacheInfo() CacheInfo
}

type cacheEntry struct {
	data    json.RawMessage
	created time.Time
}

type Client struct {
	inner   genqlientgraphql.Client
	mu      sync.Mutex
	records []QueryRecord
	cache   map[string]cacheEntry
	hits    uint64
	misses  uint64
}

func NewClient(ctx context.Context) (*Client, error) {
	httpClient := &http.Client{Transport: newIAPTransport()}

	return &Client{
		inner: genqlientgraphql.NewClient(defaultEndpoint, httpClient),
		cache: make(map[string]cacheEntry),
	}, nil
}

func cacheKey(req *genqlientgraphql.Request) string {
	variables, _ := json.Marshal(req.Variables)
	return req.OpName + ":" + string(variables)
}

func (c *Client) MakeRequest(ctx context.Context, req *genqlientgraphql.Request, resp *genqlientgraphql.Response) error {
	isMutation := strings.HasPrefix(strings.TrimSpace(req.Query), "mutation")

	if !isMutation {
		key := cacheKey(req)
		c.mu.Lock()
		cached, ok := c.cache[key]
		c.mu.Unlock()

		if ok && time.Since(cached.created) < cacheTTL {
			c.mu.Lock()
			c.hits++
			c.records = append(c.records, QueryRecord{
				OpName:    req.OpName,
				Timestamp: time.Now(),
				CacheHit:  true,
			})
			if len(c.records) > maxQueryRecords {
				c.records = c.records[len(c.records)-maxQueryRecords:]
			}
			c.mu.Unlock()

			return json.Unmarshal(cached.data, resp.Data)
		}

		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
	}

	slog.Info("graphql request start", "op", req.OpName)
	start := time.Now()
	err := c.inner.MakeRequest(ctx, req, resp)
	duration := time.Since(start)

	record := QueryRecord{
		OpName:    req.OpName,
		Duration:  duration,
		Timestamp: start,
	}

	if err != nil {
		slog.Warn("graphql request failed", "op", req.OpName, "duration", duration, "error", err)
		record.Error = err.Error()
	} else {
		slog.Info("graphql request done", "op", req.OpName, "duration", duration)
	}

	c.mu.Lock()
	c.records = append(c.records, record)
	if len(c.records) > maxQueryRecords {
		c.records = c.records[len(c.records)-maxQueryRecords:]
	}

	if err == nil && !isMutation {
		key := cacheKey(req)
		if data, marshalErr := json.Marshal(resp.Data); marshalErr == nil {
			c.cache[key] = cacheEntry{data: data, created: time.Now()}
		}
	}

	if isMutation && err == nil {
		c.cache = make(map[string]cacheEntry)
	}

	c.mu.Unlock()

	return err
}

func (c *Client) InvalidateAll() {
	c.mu.Lock()
	c.cache = make(map[string]cacheEntry)
	c.mu.Unlock()
}

func (c *Client) QueryLog() []QueryRecord {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]QueryRecord, len(c.records))
	copy(out, c.records)
	return out
}

func (c *Client) CacheEntries() []CacheEntryDetail {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	entries := make([]CacheEntryDetail, 0, len(c.cache))
	for key, entry := range c.cache {
		age := now.Sub(entry.created)
		ttl := max(cacheTTL-age, time.Duration(0))

		opName, _, found := strings.Cut(key, ":")
		if !found {
			opName = key
		}

		entries = append(entries, CacheEntryDetail{
			Key:       key,
			OpName:    opName,
			Age:       age,
			TTL:       ttl,
			SizeBytes: len(entry.data),
		})
	}

	return entries
}

func (c *Client) CacheInfo() CacheInfo {
	c.mu.Lock()
	defer c.mu.Unlock()

	return CacheInfo{Hits: c.hits, Misses: c.misses}
}

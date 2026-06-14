// Package openstax is the library behind the openstax command.
package openstax

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://openstax.org",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client talks to the OpenStax CMS API.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, url)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get: %w", lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// Books fetches all OpenStax free textbooks.
func (c *Client) Books(ctx context.Context, limit int) ([]Book, error) {
	url := c.cfg.BaseURL + "/apps/cms/api/books/"
	raw, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var resp wireResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	books := make([]Book, 0, len(resp.Books))
	for i, w := range resp.Books {
		if limit > 0 && i >= limit {
			break
		}
		books = append(books, wireToBook(w, i+1))
	}
	return books, nil
}

// Search returns books whose title or subjects contain query (case-insensitive).
func (c *Client) Search(ctx context.Context, query string, limit int) ([]Book, error) {
	all, err := c.Books(ctx, 0)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var out []Book
	for _, b := range all {
		if strings.Contains(strings.ToLower(b.Title), q) || strings.Contains(strings.ToLower(b.Subjects), q) {
			out = append(out, b)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func wireToBook(w wireBook, rank int) Book {
	return Book{
		Rank:     rank,
		Title:    w.Title,
		Subjects: strings.Join(w.Subjects, "; "),
		State:    w.BookState,
		URL:      w.WebviewRexLink,
		PDF:      w.HighResolutionPDFURL,
	}
}

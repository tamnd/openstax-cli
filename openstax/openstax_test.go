package openstax_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/openstax-cli/openstax"
)

func booksPayload(books []map[string]any) []byte {
	b, _ := json.Marshal(map[string]any{"books": books})
	return b
}

func TestBooks(t *testing.T) {
	payload := booksPayload([]map[string]any{
		{
			"id":                      1,
			"slug":                    "books/algebra-1",
			"title":                   "Algebra 1",
			"subjects":                []string{"Mathematics"},
			"book_state":              "live",
			"webview_rex_link":        "https://openstax.org/books/algebra-1/pages/1",
			"high_resolution_pdf_url": "https://example.com/algebra-1.pdf",
		},
		{
			"id":                      2,
			"slug":                    "books/biology-2e",
			"title":                   "Biology 2e",
			"subjects":                []string{"Science"},
			"book_state":              "live",
			"webview_rex_link":        "https://openstax.org/books/biology-2e/pages/1",
			"high_resolution_pdf_url": "https://example.com/biology-2e.pdf",
		},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	cfg := openstax.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := openstax.NewClient(cfg)
	books, err := c.Books(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 2 {
		t.Fatalf("got %d books, want 2", len(books))
	}
	if books[0].Title != "Algebra 1" {
		t.Errorf("title = %q, want Algebra 1", books[0].Title)
	}
	if books[0].Rank != 1 {
		t.Errorf("rank = %d, want 1", books[0].Rank)
	}
}

func TestBooksLimit(t *testing.T) {
	items := make([]map[string]any, 5)
	for i := range items {
		items[i] = map[string]any{"id": i, "title": "Book", "subjects": []string{}, "book_state": "live", "webview_rex_link": "https://example.com", "high_resolution_pdf_url": ""}
	}
	payload := booksPayload(items)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	cfg := openstax.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := openstax.NewClient(cfg)
	books, err := c.Books(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 3 {
		t.Fatalf("got %d books, want 3 (limit applied)", len(books))
	}
}

func TestSearch(t *testing.T) {
	payload := booksPayload([]map[string]any{
		{"id": 1, "title": "Algebra 1", "subjects": []string{"Mathematics"}, "book_state": "live", "webview_rex_link": "", "high_resolution_pdf_url": ""},
		{"id": 2, "title": "Biology 2e", "subjects": []string{"Science"}, "book_state": "live", "webview_rex_link": "", "high_resolution_pdf_url": ""},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	cfg := openstax.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := openstax.NewClient(cfg)
	books, err := c.Search(context.Background(), "algebra", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 {
		t.Fatalf("got %d books, want 1", len(books))
	}
	if books[0].Title != "Algebra 1" {
		t.Errorf("title = %q", books[0].Title)
	}
}

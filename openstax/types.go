package openstax

// Book is a free OpenStax textbook.
type Book struct {
	Rank     int    `json:"rank"`
	Title    string `json:"title"`
	Subjects string `json:"subjects"`
	State    string `json:"state"`
	URL      string `json:"url"`
	PDF      string `json:"pdf"`
}

// wire types mirror the OpenStax CMS API response.
type wireResponse struct {
	Books []wireBook `json:"books"`
}

type wireBook struct {
	ID        int      `json:"id"`
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Subjects  []string `json:"subjects"`
	BookState string   `json:"book_state"`

	WebviewRexLink       string `json:"webview_rex_link"`
	HighResolutionPDFURL string `json:"high_resolution_pdf_url"`
}

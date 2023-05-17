package responses

type ResponseServiceLink struct {
	Result string `json:"result,omitempty"`
}

type ResponseListShortLinks struct {
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
}

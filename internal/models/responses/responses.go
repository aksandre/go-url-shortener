package responses

type ResponseServiceLink struct {
	Result string `json:"result,omitempty"`
}

type ResponseListShortLinks struct {
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
}

type RowBatchServiceLink struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}
type ResponseBatchServiceLinks []RowBatchServiceLink

package requests

type RequestServiceLink struct {
	URL string `json:"url,omitempty"`
}

type RowBatchServiceLink struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
}
type RequestBatchServiceLinks []RowBatchServiceLink

package model_requests_responses

type RequestServiceLink struct {
	Url string `json:"url,omitempty"`
}

type ResponseServiceLink struct {
	Result string `json:"result,omitempty"`
}

type ResponseListShortLinks struct {
	ShortUrl    string `json:"short_url,omitempty"`
	OriginalUrl string `json:"original_url,omitempty"`
}

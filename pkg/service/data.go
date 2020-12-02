package service

type UrlRequests struct {
	Requests []string `json:"requests"`
}

type UrlResponse struct {
	Url     string `json:"url"`
	Content string `json:"content"`
	Error   error  `json:"-"`
}

type UrlResponses struct {
	Responses []UrlResponse `json:"responses"`
}

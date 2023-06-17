package service

import (
	"context"
	modelsResponses "go-url-shortener/internal/models/responses"
)

type RowShortLink modelsResponses.ResponseListShortLinks
type ListShortLinks []RowShortLink

// ключ - полная ссылка, значение - короткая ссылка c хостом
type BatchShortLinks map[string]string

type ServiceShortInterface interface {
	GetBatchShortLink(ctx context.Context, listFullURL []string) (dataBatch BatchShortLinks, err error)
	AddNewFullURL(ctx context.Context, fullURL string) (serviceLink string, err error)
	GetServiceLinkByURL(ctx context.Context, fullURL string) (serviceLink string, err error)
	GetFullLinkByShort(ctx context.Context, shortLink string) (fullURL string, err error)
	GetDataShortLinks(ctx context.Context, listFullURL any) (shortLinks ListShortLinks, err error)
	SetLength(length int)
}

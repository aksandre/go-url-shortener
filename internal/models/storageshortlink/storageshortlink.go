package storageshortlink

import "context"

// тип для одной запись с данными о короткой ссылке
type RowStorageShortLink struct {
	ShortLink string
	FullURL   string
	UUID      string
}

// вид хранения ссылок в памяти
// ключом является короткая ссылка
type DataStorageShortLink map[string]RowStorageShortLink

// фильтр для получения коротких ссылок
type FilterOptionsQuery struct {
	ListFullURL []string
}

type OptionsQuery struct {
	Filter FilterOptionsQuery
}

// тип для хранилища данных ссылок
type StorageShortInterface interface {
	GetFullLinkByShort(ctx context.Context, shortLink string) (fullURL string, err error)
	GetShortLinkByURL(ctx context.Context, fullURL string) (shortLink string, err error)
	AddShortLinkForURL(ctx context.Context, fullURL, shortLink string) (err error)
	// добавление коротких ссылок группой
	AddBatchShortLinks(ctx context.Context, dataBatch DataStorageShortLink) (err error)
	// установка всех данных хранилища
	SetData(ctx context.Context, data DataStorageShortLink) (err error)
	GetCountLink(ctx context.Context) (count int, err error)
	GetShortLinks(ctx context.Context, options *OptionsQuery) (shortLinks DataStorageShortLink, err error)
	Init(ctx context.Context) (err error)
	ClearStorage(ctx context.Context) (err error)
}

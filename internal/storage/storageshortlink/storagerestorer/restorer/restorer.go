package restorer

type RowDataRestorer struct {
	ShortLink string
	FullURL   string
	UUID      string
}

type Restorer interface {
	WriteRow(dataRow RowDataRestorer) (err error)
	ReadRow() (dataRow RowDataRestorer, err error)
	ReadAll() (allRows []RowDataRestorer, err error)
	ClearRows() (err error)
}

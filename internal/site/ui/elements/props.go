package elements

type Action struct {
	Label   string
	Href    string
	Method  string
	Style   string
	Enabled bool
}

type FieldError struct {
	Field   string
	Message string
}

type PaginationData struct {
	Page       int
	TotalPages int
	BaseHref   string
}

type MediaThumbnailData struct {
	URL   string
	Alt   string
	Label string
}

type AccountActionsData struct {
	Email string
	Token string
}

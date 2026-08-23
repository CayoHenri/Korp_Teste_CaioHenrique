package query

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Pagination struct {
	Page     int
	PageSize int
}

func NewPagination(page, pageSize int) Pagination {
	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Pagination{Page: page, PageSize: pageSize}
}

func (pagination Pagination) Offset() int {
	return (pagination.Page - 1) * pagination.PageSize
}

type Criteria[F any] struct {
	Filters    F
	Pagination Pagination
}

type Page[T any] struct {
	Items      []T   `json:"itens"`
	Total      int64 `json:"total"`
	Page       int   `json:"pagina"`
	PageSize   int   `json:"tamanhoPagina"`
	TotalPages int   `json:"totalPaginas"`
}

func NewPage[T any](items []T, total int64, pagination Pagination) Page[T] {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))
	}
	return Page[T]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}
}

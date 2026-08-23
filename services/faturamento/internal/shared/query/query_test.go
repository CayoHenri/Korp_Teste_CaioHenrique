package query

import "testing"

func TestNewPaginationAplicaPadroesELimite(t *testing.T) {
	defaultPagination := NewPagination(0, 0)
	if defaultPagination.Page != 1 || defaultPagination.PageSize != 20 {
		t.Fatalf("paginacao padrao inesperada: %+v", defaultPagination)
	}
	limitedPagination := NewPagination(2, 500)
	if limitedPagination.PageSize != 100 || limitedPagination.Offset() != 100 {
		t.Fatalf("limite ou offset inesperado: %+v", limitedPagination)
	}
}

func TestNewPageCalculaTotalDePaginas(t *testing.T) {
	page := NewPage([]int{1, 2}, 21, NewPagination(2, 20))
	if page.TotalPages != 2 || page.Total != 21 || page.Page != 2 {
		t.Fatalf("metadados inesperados: %+v", page)
	}
}

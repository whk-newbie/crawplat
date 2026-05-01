package httpx

type PaginatedResponse struct {
	Items  any   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type PaginationParams struct {
	Limit  int
	Offset int
}

func DefaultPagination(limit, offset int) PaginationParams {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return PaginationParams{Limit: limit, Offset: offset}
}

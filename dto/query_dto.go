package dto

type QueryDto struct {
	All       bool   `query:"all"`
	Page      int    `query:"page"`
	PerPage   int    `query:"per_page"`
	Search    string `query:"search"`
	SortField string `query:"sort_field"`
	SortOrder string `query:"sort_order"`
}

type QueryResponse struct {
	Total      int    `json:"total"`
	PerPage    int    `json:"per_page"`
	CurPage    int    `json:"current_page"`
	NextCursor string `json:"next_cursor,omitempty"`
	LastPage   int    `json:"last_page"`
	Data       any    `json:"data"`
}

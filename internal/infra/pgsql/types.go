package pgsql

type PageResult[T any] struct {
	List      []T   `json:"list"`
	Total     int64 `json:"total"`
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	PageCount int   `json:"page_count"`
}

// 聚合查询结果
type AggregateResult struct {
	Field string
	Count int64
	Sum   float64
	Max   float64
	Min   float64
	Avg   float64
}

package utils

// PageData 分页列表数据，JSON 字段与常见前端约定一致（list / pageNum / pageSize / total）。
type PageData struct {
	List     any   `json:"list"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

// NewPageData 构造分页数据，通常作为 Success / SuccessPage 的 data 载荷。
func NewPageData(list any, pageNum, pageSize int, total int64) *PageData {
	return &PageData{
		List:     list,
		PageNum:  pageNum,
		PageSize: pageSize,
		Total:    total,
	}
}

package ringapi

// OffsetLimit 分页
type OffsetLimit struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

package redis

// Element zset返回结构定义
type Element struct {
	Member string `json:"member,omitempty"`
	Score  int64  `json:"score,omitempty"`
}

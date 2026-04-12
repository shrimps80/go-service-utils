package utils

// MapPtr 将 []*A 按索引一一映射为 []*B；items[i] 为 nil 时对应结果为 nil，否则为 fn(items[i])。
func MapPtr[A, B any](items []*A, fn func(*A) *B) []*B {
	if len(items) == 0 {
		return nil
	}
	out := make([]*B, len(items))
	for i, it := range items {
		if it == nil {
			continue
		}
		out[i] = fn(it)
	}
	return out
}

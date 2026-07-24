package converter

// StripUnavailable 深度剔除 nil / "" 字段，对应 JS 版 stripUnavailable。
// 对象若所有字段被剔除后为空，则该对象本身也被剔除（返回 nil, false）；
// 数组则始终保留（即使为空数组），与原实现行为一致。
func StripUnavailable(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case []interface{}:
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			if stripped, ok := StripUnavailable(item); ok {
				result = append(result, stripped)
			}
		}
		return result, true
	case anyMap:
		result := anyMap{}
		for key, item := range v {
			if stripped, ok := StripUnavailable(item); ok {
				result[key] = stripped
			}
		}
		if len(result) == 0 {
			return nil, false
		}
		return result, true
	case nil:
		return nil, false
	case string:
		if v == "" {
			return nil, false
		}
		return v, true
	default:
		return v, true
	}
}

// stripMap 是 StripUnavailable 的便捷封装：始终返回 anyMap（可能为空 map）。
func stripMap(m anyMap) anyMap {
	v, ok := StripUnavailable(m)
	if !ok {
		return anyMap{}
	}
	result, _ := asMap(v)
	return result
}

// firstDefined 对应 JS 的 `??`（nullish coalescing）：返回第一个非 nil 的值。
func firstDefined(values ...interface{}) interface{} {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

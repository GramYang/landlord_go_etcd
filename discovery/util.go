package discovery

import "strconv"

func AnyToString(data interface{}) string {
	switch v := data.(type) {
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', 2, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case string:
		return v
	default:
		return ""
	}
}
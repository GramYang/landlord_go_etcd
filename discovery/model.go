package discovery

import "strings"

const (
	ServiceKeyPrefix = "_svcdesc_"
	KVKeyPrefix = "_kv_"
)

func IsServiceKey(key string) bool {
	return strings.HasPrefix(key, ServiceKeyPrefix)
}

func IsKVKey(key string) bool {
	return strings.HasPrefix(key, KVKeyPrefix)
}

func GetSvcIDByServiceKey(rawkey string) string {
	if IsServiceKey(rawkey) {
		return rawkey[len(ServiceKeyPrefix):]
	}
	return ""
}

func GetKVKey(rawkey string) string {
	if IsKVKey(rawkey) {
		return rawkey[len(KVKeyPrefix):]
	}
	return ""
}

func GetSvcNameByID(svcid string) string {
	return svcid[:strings.Index(svcid, "#")]
}
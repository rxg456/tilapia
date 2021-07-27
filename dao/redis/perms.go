package redis

const (
	// all perms key
	AllPermsKey = "all_perms_key"
)

func DelRedisAllPermKey() {
	key := AllPermsKey
	Rdb.Del(key).Err()
}

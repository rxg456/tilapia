package redis

import "time"

func SetValByKey(key string, val interface{}, expiration time.Duration) error {
	_, err := Rdb.Set(key, val, expiration).Result()

	return err
}

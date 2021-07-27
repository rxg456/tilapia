package redis

import "time"

func SetValByKey(key string, val interface{}, expiration time.Duration) error {
	_, err := Rdb.Set(key, val, expiration).Result()

	return err
}

func SetValBySetKey(key string, val interface{}) error {
	_, err := Rdb.SAdd(key, val).Result()

	return err
}

func CheckMemberByKey(key string, val interface{}) bool {
	isMember, _ := Rdb.SIsMember(key, val).Result()
	return isMember
}

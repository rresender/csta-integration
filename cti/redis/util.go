package redis

import (
	"fmt"
	"log"

	"github.com/garyburd/redigo/redis"
)

//TTL Time to live
const TTL = "60"

// Ping redis
func Ping() {

	conn := Pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		log.Panicf("cannot 'PING' db: %v", err)
	}
}

// Get a value
func Get(key string) ([]byte, error) {

	conn := Pool.Get()
	defer conn.Close()

	var data []byte
	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return data, fmt.Errorf("error getting key %s: %v", key, err)
	}
	return data, err
}

// Set a key
func Set(key string, value []byte) error {

	conn := Pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error setting key %s to %s: %v", key, v, err)
	}
	return err
}

// Exists a key
func Exists(key string) (bool, error) {

	conn := Pool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
	}
	return ok, err
}

// Expire a key
func Expire(key string) (bool, error) {

	conn := Pool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXPIRE", key, TTL))
	if err != nil {
		return ok, fmt.Errorf("error checking if key %s expire: %v", key, err)
	}
	return ok, err
}

// Delete a key
func Delete(key string) error {

	conn := Pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return err
}

// GetKeys all
func GetKeys(pattern string) ([]string, error) {

	conn := Pool.Get()
	defer conn.Close()

	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}

// Incr a counter
func Incr(counterKey string) (int, error) {

	conn := Pool.Get()
	defer conn.Close()

	return redis.Int(conn.Do("INCR", counterKey))
}

// PushValue into a set
func PushValue(list string, value string) error {
	conn := Pool.Get()
	defer conn.Close()
	_, err := conn.Do("SADD", list, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error Pushing value %s to list %s: %v", value, list, err)
	}
	return err
}

// RemoveValue from a set
func RemoveValue(list string, value string) error {
	conn := Pool.Get()
	defer conn.Close()
	_, err := conn.Do("SREM", list, 0, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error removing value %s from list %s: %v", value, list, err)
	}
	return err
}

// GetValues from a set
func GetValues(key string) []string {
	conn := Pool.Get()
	defer conn.Close()

	value, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
		log.Fatal(err)
	}

	return value
}

// Close Pool
func Close() {
	Pool.Close()
}

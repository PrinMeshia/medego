package cache

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func (c *RedisCache) Has(str string) (bool, error) {
	key := c.buildKey(str)
	conn := c.Conn.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (c *RedisCache) buildKey(str string) string {
	return fmt.Sprintf("%s:%s", c.Prefix, str)
}

func (c *RedisCache) Get(str string) (interface{}, error) {
	key := c.buildKey(str)
	conn := c.Conn.Get()
	defer conn.Close()

	cacheEntry, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(cacheEntry))
	if err != nil {
		return nil, err
	}

	return decoded[key], nil
}

func (c *RedisCache) Set(str string, value interface{}, expires ...int) error {
	key := c.buildKey(str)
	conn := c.Conn.Get()
	defer conn.Close()

	entry := Entry{key: value}
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	var cmd string
	if len(expires) > 0 {
		cmd = "SETEX"
	} else {
		cmd = "SET"
	}

	args := redis.Args{key, encoded}
	if len(expires) > 0 {
		args = append(args, expires[0])
	}

	_, err = conn.Do(cmd, args...)
	return err
}

func (c *RedisCache) Forget(str string) error {
	key := c.buildKey(str)
	conn := c.Conn.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return err
}

func (c *RedisCache) EmptyByMatch(str string) error {
	pattern := fmt.Sprintf("%s:%s*", c.Prefix, str)
	keys, err := c.getKeys(pattern)
	if err != nil {
		return err
	}

	conn := c.Conn.Get()
	defer conn.Close()

	for _, x := range keys {
		_, err := conn.Do("DEL", x)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RedisCache) Empty() error {
	pattern := fmt.Sprintf("%s:*", c.Prefix)
	keys, err := c.getKeys(pattern)
	if err != nil {
		return err
	}

	conn := c.Conn.Get()
	defer conn.Close()

	for _, x := range keys {
		_, err := conn.Do("DEL", x)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RedisCache) getKeys(pattern string) ([]string, error) {
	conn := c.Conn.Get()
	defer conn.Close()

	iter := 0
	keys := []string{}

	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, err
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

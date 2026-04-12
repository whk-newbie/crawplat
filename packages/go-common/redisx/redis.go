package redisx

import (
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
)

func NewClient(addr string) (*redis.Client, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.New("redis addr is required")
	}
	return redis.NewClient(&redis.Options{Addr: addr}), nil
}

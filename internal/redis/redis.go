package redis

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

// Pool manages a pool of connections of redis.
type Pool interface {
	GetContext(context.Context) (redis.Conn, error)
	Close() error
}

// New returns Pool implementation.
func New(url string) Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(url)
			if err != nil {
				return nil, errors.Wrapf(err, "faild to dial url: %v", url)
			}

			if _, err := c.Do("PING"); err != nil {
				c.Close()
				return nil, errors.Wrap(err, "faild to write command: PING")
			}
			return c, nil
		},
	}
}

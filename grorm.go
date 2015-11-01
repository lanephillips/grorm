package grorm

import (
	"github.com/garyburd/redigo/redis"
	// "reflect"
)

type Conn struct {
	conn redis.Conn
}

func NewConn() (*Conn, error) {
	var c Conn
	// TODO: user should be able to pass in own args
	rc, err := redis.Dial("tcp", ":6379")
    if err != nil {
        return nil, err
    }

    c.conn = rc
    return &c, nil
}

func (c *Conn) Close() {
	c.conn.Close()
}

func (c *Conn) Save(object interface{}) error {
	return nil
}

func (c *Conn) Load(id int64, object interface{}) error {
	return nil
}
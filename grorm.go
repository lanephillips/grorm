package grorm

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"reflect"
)

type Conn struct {
	conn redis.Conn
	appPrefix string
}

func NewConn(appPrefix string) (*Conn, error) {
	var c Conn
	// TODO: user should be able to pass in own args
	rc, err := redis.Dial("tcp", ":6379")
    if err != nil {
        return nil, err
    }

    c.conn = rc
    c.appPrefix = appPrefix

    return &c, nil
}

func (c *Conn) Close() {
	c.conn.Close()
}

// object must be a pointer to a settable object
func (c *Conn) Save(object interface{}) error {
	t := reflect.TypeOf(object)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t.Kind() != reflect.Struct {
			return fmt.Errorf("Object is a pointer to %v, not a pointer to struct.", t.Kind())
		}
	} else {
		return fmt.Errorf("Object is a %v, not a pointer to struct.", t.Kind())
	}
	v := reflect.ValueOf(object).Elem()

	id := v.FieldByName("Id")
	if !id.IsValid() || id.Kind() != reflect.Uint64 {
		return fmt.Errorf("Object does not have an Id field.")
	}

	if id.Uint() == 0 {
		keyName := fmt.Sprintf("%v:%v", c.appPrefix, t.Name())
		newId, err := redis.Uint64(c.conn.Do("INCR", keyName))
		if err != nil {
			return err
		}
		newIdV := reflect.ValueOf(newId)
		// fmt.Printf("Attempting to set %v to %v.\n", id, newIdV)
		id.Set(newIdV)
		// fmt.Printf("Incremented %v to %v.\n", keyName, newId)
	}

	key := fmt.Sprintf("%v:%v:%v", c.appPrefix, t.Name(), id.Uint())
	args := []interface{}{ key }

	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if name == "Id" {
			continue
		}
		value := v.Field(i).Interface()

		args = append(args, name, value)
	}

	_, err := c.conn.Do("HMSET", args...)
	if err != nil {
		return err
	}
	fmt.Printf("Saved %v to %v.\n", args, key)

	return nil
}

func (c *Conn) Load(id int64, object interface{}) error {
	return nil
}
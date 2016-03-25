package grorm

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"encoding/json"
	"reflect"
	"strconv"
)

type store interface {
	save(object interface{}) error
	load(id uint64, object interface{}) error
	delete(typeName string, id uint64) error
	close()
}

// po is pointer to settable struct object
func copyJsonToObject(r io.Reader, po reflect.Value) error {
	if po.Kind() != reflect.Ptr {
		return newInternalError(nil, "Argument is a %v, not a pointer to struct.", po.Kind())
	}
	o := po.Elem()
	if o.Kind() != reflect.Struct {
		return newInternalError(nil, "Argument is a pointer to %v, not a pointer to struct.", o.Kind())
	}

	// fill in fields from JSON
	// TODO: accept zero value for unspecified field unless annotated otherwise
	d := json.NewDecoder(r)
	m := map[string]interface{}{}
	err := d.Decode(&m)
	if err != nil {
		return newBadRequestError(err, "Malformed JSON")
	}

	for k, v := range m {
		// error if id is specified
		if k == "Id" {
			return newBadRequestError(nil, "You may not set the 'Id' field.")
		}

		f := o.FieldByName(k)
		// error on extra fields
		if !f.IsValid() || !f.CanSet() {
			return newBadRequestError(nil, "Request body specifies field '%v' which cannot be set.", k)
		}

		v2 := reflect.ValueOf(v)
		if !v2.Type().AssignableTo(f.Type()) {
			return newBadRequestError(nil, "Field '%v' cannot take the value %v.", k, v)
		}

		// actually set value
		// TODO: probably should try to catch panic if possible, might have forgotten something above
		// TODO: can we defer and recover and set the return value of this function?
		f.Set(v2)
	}
	return nil
}

type redisStore struct {
	conn redis.Conn
	appPrefix string
}

func newRedisStore(appPrefix string) (*redisStore, error) {
	// TODO: user should be able to pass in own args
	rc, err := redis.Dial("tcp", ":6379")
    if err != nil {
        return nil, err
    }

	var c redisStore
    c.conn = rc
    c.appPrefix = appPrefix

    return &c, nil
}

func (c *redisStore) close() {
	c.conn.Close()
}

// object must be a pointer to a settable object
func (c *redisStore) save(object interface{}) error {
	t := reflect.TypeOf(object)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t.Kind() != reflect.Struct {
			return newInternalError(nil, "Object is a pointer to %v, not a pointer to struct.", t.Kind())
		}
	} else {
		return newInternalError(nil, "Object is a %v, not a pointer to struct.", t.Kind())
	}
	v := reflect.ValueOf(object).Elem()

	id := v.FieldByName("Id")
	if !id.IsValid() || id.Kind() != reflect.Uint64 {
		return newInternalError(nil, "Object does not have an Id field.")
	}

	if id.Uint() == 0 {
		keyName := fmt.Sprintf("%v:%v", c.appPrefix, t.Name())
		newId, err := redis.Uint64(c.conn.Do("INCR", keyName))
		if err != nil {
			return err
		}
		newIdV := reflect.ValueOf(PrimaryKey(newId))
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

func (c *redisStore) load(id uint64, object interface{}) error {
	t := reflect.TypeOf(object)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t.Kind() != reflect.Struct {
			return newInternalError(nil, "Object is a pointer to %v, not a pointer to struct.", t.Kind())
		}
	} else {
		return newInternalError(nil, "Object is a %v, not a pointer to struct.", t.Kind())
	}
	v := reflect.ValueOf(object).Elem()

	idf := v.FieldByName("Id")
	if !idf.IsValid() || idf.Kind() != reflect.Uint64 {
		return newInternalError(nil, "Object does not have an Id field.")
	}
	idf.Set(reflect.ValueOf(PrimaryKey(id)))

	key := fmt.Sprintf("%v:%v:%v", c.appPrefix, t.Name(), id)
	values, err := redis.StringMap(c.conn.Do("HGETALL", key))
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return newNotFoundError(nil, "%v with id %v not found.", t.Name(), id)
	}

	for name, value := range values {
		f, ok := t.FieldByName(name)
		if !ok {
			continue
		}

		var value2 reflect.Value

		switch f.Type.Kind() {
		case reflect.String:
			value2 = reflect.ValueOf(value)

		case reflect.Int64:
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				continue
			}
			value2 = reflect.ValueOf(i)

		case reflect.Uint64:
			i, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				continue
			}
			value2 = reflect.ValueOf(i)
			
		default:
			continue
		}

		v.FieldByIndex(f.Index).Set(value2)
		// fmt.Printf("Set %v to %v.\n", f, value2)
	}

	return nil
}

func (c *redisStore) delete(typeName string, id uint64) error {
	key := fmt.Sprintf("%v:%v:%v", c.appPrefix, typeName, id)
	count, err := redis.Int64(c.conn.Do("DEL", key))
	if err != nil {
		return err
	}
	if count == 0 {
		return newNotFoundError(nil, "%v with id %v not found.", typeName, id)
	}
	return nil
}

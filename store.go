package grorm

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

type store interface {
	save(object interface{}) error
	load(id uint64, object interface{}) error
	delete(typeName string, id uint64) error
	close()
}

func copyJsonToObject(r io.Reader, mv *metaValue) (err error) {
	// fill in fields from JSON
	// TODO: accept zero value for unspecified field unless annotated otherwise
	d := json.NewDecoder(r)
	m := map[string]interface{}{}
	err = d.Decode(&m)
	if err != nil {
		return newBadRequestError(err, "Malformed JSON")
	}
	
	// want these accessible in case of panic
	var k string
	var v interface{}
	
	defer func() {
		if r := recover(); r != nil {
			// internal error because we should have already considered all possible panics
			err = newInternalError(nil, "Panic setting field '%v' to the value %v: %v.", k, v, r)
		}
	}()

	o := mv.p.Elem()
	for k, v = range m {
		// error if id is specified
		tf, ok := mv.mt.t.FieldByName(k)
		// error on extra fields
		if !ok {
			return newBadRequestError(nil, "Request body specifies field '%v' which cannot be set.", k)
		}
		if tf.Index[0] == mv.mt.id {
			return newBadRequestError(nil, "You may not set the primary key field.")
		}

		f := o.Field(tf.Index[0])
		v2 := reflect.ValueOf(v)
		
		// if JSON value is "null", consider it as the "zero" value of the target field's type
		if v == nil {
			v2 = reflect.Zero(f.Type())
		}
		
		// golang json decodes all numbers into float64, so convert if needed
		// TODO: float64 can't represent all possible int64 values
		if f.Kind() == reflect.Int64 && v2.Kind() == reflect.Float64 {
			v2 = reflect.ValueOf(int64(v.(float64)))
		} else if f.Kind() == reflect.Uint64 && v2.Kind() == reflect.Float64 {
			v2 = reflect.ValueOf(uint64(v.(float64)))
		} else if f.Type() == timeType && v2.Kind() == reflect.String {
			t := &time.Time{}
			err = t.UnmarshalText([]byte(v.(string)))
			if err != nil {
				return newBadRequestError(err, "Malformed date: %v.", v)
			}
			v2 = reflect.ValueOf(*t)
		} else if f.Type() == byteSliceType && v2.Kind() == reflect.String {
			bs := []byte{}
			err = json.Unmarshal([]byte("\"" + v.(string) + "\""), &bs)
			if err != nil {
				return newBadRequestError(err, "Malformed base64: %v.", v)
			}
			v2 = reflect.ValueOf(bs)
		}

		if !v2.Type().AssignableTo(f.Type()) {
			return newBadRequestError(nil, "Field '%v' with type %v cannot take the %v value %v.", k, f.Type(), v2.Type(), v)
		}

		// actually set value
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
	rc, err := redis.Dial("tcp", "localhost:6379")
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
func (c *redisStore) save(mv *metaValue) error {
	if mv.getPrimaryKey() == 0 {
		keyName := fmt.Sprintf("%v:%v", c.appPrefix, mv.mt.t.Name())
		newId, err := redis.Uint64(c.conn.Do("INCR", keyName))
		if err != nil {
			return err
		}
		mv.setPrimaryKey(newId)
	}

	key := mv.getKeyString(c.appPrefix)
	args := []interface{}{ key }

	v := mv.p.Elem()
	for i := 0; i < mv.mt.t.NumField(); i++ {
		if mv.mt.id == i {
			continue
		}

		name := mv.mt.t.Field(i).Name
		value := v.Field(i).Interface()
		
		if mv.mt.t.Field(i).Type == timeType {
			timev, ok := value.(time.Time)
			if !ok {
				return newInternalError(nil, "Shouldn't happen.")
			}
			bs, err := timev.MarshalText()
			if err != nil {
				return newInternalError(err, "Shouldn't happen.")
			}
			value = bs
		}

		args = append(args, name, value)
	}

	_, err := c.conn.Do("HMSET", args...)
	if err != nil {
		return err
	}
	fmt.Printf("Saved %v to %v.\n", args, key)

	return nil
}

func (c *redisStore) load(id uint64, mv *metaValue) error {
	mv.setPrimaryKey(id)
	v := mv.p.Elem()

	key := mv.getKeyString(c.appPrefix)
	values, err := redis.StringMap(c.conn.Do("HGETALL", key))
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return newNotFoundError(nil, "%v with id %v not found.", mv.mt.t.Name(), id)
	}

	for name, value := range values {
		f, ok := mv.mt.t.FieldByName(name)
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

		case reflect.Float64:
			n, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			value2 = reflect.ValueOf(n)

		case reflect.Bool:
			n, err := strconv.ParseBool(value)
			if err != nil {
				continue
			}
			value2 = reflect.ValueOf(n)
			
		default:
			if f.Type == byteSliceType {
				// fmt.Printf("I want to parse '%v' as a byte array.\n", value)
				value2 = reflect.ValueOf([]byte(value))
			} else if f.Type == timeType {
				// fmt.Println("I want to parse '"+value+"' as a time.")
				t := &time.Time{}
				err = t.UnmarshalText([]byte(value))
				if err != nil {
					return newInternalError(err, "Malformed date: %v.", value)
				}
				value2 = reflect.ValueOf(*t)
			} else {
				continue
			}
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

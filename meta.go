package grorm

import (
	"fmt"
	"reflect"
	"time"
)

var primaryKeyType = reflect.TypeOf(PrimaryKey(0))
var timeType = reflect.TypeOf(time.Time{})

type metaType struct {
	// type of the object this metaType represents
	t reflect.Type
	// index of the identifier field
	id int
	// and its name
	idName string
}

type metaValue struct {
	// value is a pointer to an instance of a registered type
	p reflect.Value
	// the metaType for the value's Type
	mt *metaType
}

func getMetaType(exampleValue interface{}) (*metaType, error) {
	t := reflect.TypeOf(exampleValue)

	if t.Kind() != reflect.Struct {
		return nil, newConfigurationError(nil, "A struct type is required.")
	}

	md := &metaType{ t, -1, "" }

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type == primaryKeyType {
			if md.id < 0 {
				md.id = i
				md.idName = f.Name
			} else {
				return nil, newConfigurationError(nil, "Type '%v' has more than one primary key field.", t.Name())
			}
		}
	}

	if md.id < 0 {
		return nil, newConfigurationError(nil, "Type '%v' is missing a primary key field.", t.Name())
	}

	// TODO: build ACL objects from field annotations

	return md, nil
}

func (mt *metaType) newValue() *metaValue {
	p := reflect.New(mt.t)
	return &metaValue{ p, mt }
}

func (mv *metaValue) setPrimaryKey(id uint64) {
	value := mv.p.Elem().Field(mv.mt.id)
	value.Set(reflect.ValueOf(PrimaryKey(id)))
}

func (mv *metaValue) getPrimaryKey() uint64 {
	value := mv.p.Elem().Field(mv.mt.id)
	return value.Uint()
}

func (mv *metaValue) getKeyString(prefix string) string {
	value := mv.p.Elem().Field(mv.mt.id)
	return fmt.Sprintf("%v:%v:%v", prefix, mv.mt.t.Name(), value.Uint())	
}

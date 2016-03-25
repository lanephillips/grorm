package grorm

import (
	"reflect"
)

var intPrimaryKeyType = reflect.TypeOf(IntPrimaryKey(0))

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
	t *metaType
}

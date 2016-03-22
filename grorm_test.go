package grorm

import (
    "testing"
)

type Thing struct {
	id int
	name string
}

func TestServer(t *testing.T) {
	r := NewRouter()
	r.RegisterType(Thing{}, nil)
	r.ListenAndServe(":8080")
}

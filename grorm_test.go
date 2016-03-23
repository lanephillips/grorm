package grorm

import (
    "testing"
)

type Thing struct {
	id uint64
	name string
}

func TestServer(t *testing.T) {
	r := NewRouter()
	err := r.RegisterType(Thing{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.ListenAndServe(":8080")
}

package grorm

import (
    "testing"
)

type Thing struct {
	Id uint64
	Name string
}

func TestServer(t *testing.T) {
	r, err := NewServer("grormtest")
	if err != nil {
		t.Fatal(err)
	}

	err = r.RegisterType(Thing{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	
	r.ListenAndServe(":8080")
}

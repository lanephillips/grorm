package grorm

import (
    "testing"
	"time"
)

type Thing struct {
	Id PrimaryKey
	Name string
	Num int64
	Unum uint64
	Float float64
	Flag bool
	Date time.Time
	Data []byte
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

	r.ListenAndServe("localhost:8080")
}

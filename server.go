package grorm

import (
    "fmt"
    "log"
    "net/http"
    "reflect"
    "strings"
)

type Server struct {
	types map[string]reflect.Type
	router *router
	store *redisStore
}

func NewServer(appName string) (*Server, error) {
	s := &Server{}

	s.types = map[string]reflect.Type{}

	c, err := newRedisStore(appName)
	if err != nil {
		return nil, err
	}
	s.store = c

	s.router = newRouter(s)

	return s, nil
}

func (r *Server) RegisterType(object interface{}, nameOrNil *string) error {
	t := reflect.TypeOf(object)

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("A struct type is required.")
	}

	// TODO: check that struct conforms
	f, ok := t.FieldByName("Id")
	if !ok {
		return fmt.Errorf("An 'Id' field is required.")
	}
	if f.Type.Kind() != reflect.Uint64 {
		return fmt.Errorf("Type of 'Id' must be uint64.")
	}

	// TODO: build ACL objects from field annotations

	// TODO: field names are still capitalized, do we really want Go style to leak through? maybe have pluggable mappers
	if nameOrNil == nil {
		s := strings.ToLower(t.Name())
		nameOrNil = &s
	}

	r.types[*nameOrNil] = t
	log.Printf("Registered type %v with path \"%v\".\n", t, *nameOrNil)
	return nil
}

func (r *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r.router)
}

// TODO: disable above, require TLS
func (r *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
    return http.ListenAndServeTLS(addr, certFile, keyFile, r.router)
}

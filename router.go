package grorm

import (
    "fmt"
    "log"
    "net/http"
    "reflect"
    "strings"
)

type Router struct {
	types map[string]reflect.Type
}

func NewRouter() *Router {
	var r Router
	r.types = map[string]reflect.Type{}
	return &r
}

func (r *Router) RegisterType(object interface{}, nameOrNil *string) error {
	t := reflect.TypeOf(object)

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("A struct type is required.")
	}

	if nameOrNil == nil {
		s := strings.ToLower(t.Name())
		nameOrNil = &s
	}

	// TODO: build ACL objects from field annotations
	// TODO: check that struct conforms

	r.types[*nameOrNil] = t
	log.Printf("Registered type %v with path \"%v\".\n", t, *nameOrNil)
	return nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	// TODO: rate limiting
	// TODO: authenticate user
	// TODO: strip api prefix

	// tokenize path
	path := strings.Split(rq.URL.Path, "/")
	if len(path) < 2 {
		s := fmt.Sprintf("%v not found.", path)
		http.Error(w, s, http.StatusNotFound)
		return
	}

	// look up type name
	t, ok := r.types[path[1]]
	if !ok {
		http.Error(w, path[1] + " not found.", http.StatusNotFound)
		return
	}

	// TODO: retrieve object or list from redis
	// TODO: demux method
	// TODO: test ACL for method and user

	if rq.Method == "GET" {
		// TODO: search all objects filter with query parms
		// TODO: or get object by id
		// TODO: or get scalar field
		// TODO: or get list field
	    fmt.Fprintf(w, "Hi there, I love %v!", t)
	    return
	}

	if rq.Method == "POST" {
		// TODO: create new object from JSON in body
		// TODO: or add object id to list field

		// TODO: accept zero value for unspecified field unless annotated otherwise
		// TODO: error on extra fields
		// TODO: error if id is specified
	}

	if rq.Method == "PUT" {
		// TODO: update object with id with values from JSON
		// TODO: or update scalar field
		// TODO: or set whole list field
	}

	if rq.Method == "DELETE" {
		// TODO: delete object with id
		// TODO: or remove object id from a list field
	}

    http.Error(w, "Bad request.", http.StatusBadRequest)
}

func (r *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r)
}

// TODO: disable above, require TLS
func (r *Router) ListenAndServeTLS(addr, certFile, keyFile string) error {
    return http.ListenAndServeTLS(addr, certFile, keyFile, r)
}
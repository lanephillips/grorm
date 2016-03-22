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

	r.types[*nameOrNil] = t
	log.Printf("Registered type %v with path \"%v\".\n", t, *nameOrNil)
	return nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", rq.URL.Path[1:])
}

func (r *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r)
}

// TODO: disable above, require TLS
func (r *Router) ListenAndServeTLS(addr, certFile, keyFile string) error {
    return http.ListenAndServeTLS(addr, certFile, keyFile, r)
}
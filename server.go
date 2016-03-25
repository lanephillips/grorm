package grorm

import (
    "net/http"
    "time"
)

// marks the struct field as the identifier of its object, taking int values
type PrimaryKey uint64

// TODO: support these, too
type DateCreated time.Time
type DateModified time.Time
type IntIndex int64
type StringIndex string
type DateIndex time.Time

type Server struct {
	router *router
	resolver *resolver
	store *redisStore
}

func NewServer(appName string) (*Server, error) {
	s := &Server{}

	c, err := newRedisStore(appName)
	if err != nil {
		return nil, err
	}
	s.store = c

	s.router = newRouter(s)
	s.resolver = newResolver(s)

	return s, nil
}

func (r *Server) RegisterType(object interface{}, nameOrNil *string) error {
	return r.resolver.registerType(object, nameOrNil)
}

func (r *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r.router)
}

// TODO: disable above, require TLS
func (r *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
    return http.ListenAndServeTLS(addr, certFile, keyFile, r.router)
}

package grorm

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

type router struct {
	server *Server
}

func newRouter(server *Server) *router {
	return &router{ server }
}

func (r *router) handleRequest(w http.ResponseWriter, rq *http.Request) error {
	// TODO: rate limiting
	// TODO: authenticate user
	// TODO: strip api prefix

	// tokenize path
	path := strings.Split(rq.URL.Path, "/")

	// TODO: retrieve object or list from redis
	// TODO: demux method
	// TODO: test ACL for method and user

	if rq.Method == "GET" {
		mt, mv, err := r.server.resolver.resolvePathObject(path)
		if err != nil {
			return err
		}

		if mv == nil {
			// TODO: search all objects filter with query parms
			return newHttpError(http.StatusNotImplemented, nil, mt.t.Name())
		} else {
			// or get object by id
			w.Header().Set("Content-Type", "application/json")
			e := json.NewEncoder(w)
			e.Encode(mv.p.Interface())
			return nil
		}

		// TODO: or get scalar field
		// TODO: or get list field
	}

	if rq.Method == "POST" {
		// path should just be a type
		// TODO: or add object id to list field
		mt, id, err := r.server.resolver.resolvePath(path)
		if err != nil {
			return err
		}
		if id != nil {
			return newBadRequestError(nil, "You can't POST to an Id.")
		}

		// create new object from JSON in body
		mv := mt.newValue()
		err = copyJsonToObject(rq.Body, mv)
		if err != nil {
			return err
		}

		// save the object to DB
		err = r.server.store.save(mv)
		if err != nil {
			return err
		}

		// return JSON including the new id
		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.Encode(mv.p.Interface())
	    return nil
	}

	if rq.Method == "PUT" {
		// TODO: update object with id with values from JSON
		_, mv, err := r.server.resolver.resolvePathObject(path)
		if err != nil {
			return err
		}
		if mv == nil {
			return newBadRequestError(nil, "No object Id was given.")
		}

		err = copyJsonToObject(rq.Body, mv)
		if err != nil {
			return err
		}
		// save the object to DB
		err = r.server.store.save(mv)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.Encode(mv.p.Interface())
		return nil

		// TODO: or update scalar field
		// TODO: or set whole list field
	}

	if rq.Method == "DELETE" {
		// delete object with id
		mt, id, err := r.server.resolver.resolvePath(path)
		if err != nil {
			return err
		}
		if id == nil {
			return newBadRequestError(nil, "Missing Id.")
		}

		err = r.server.store.delete(mt.t.Name(), *id)
		if err != nil {
			return err
		}
		return nil
		// TODO: or remove object id from a list field
	}

    return newHttpError(http.StatusMethodNotAllowed, nil, "Method not allowed.")
}

func (r *router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	err := r.handleRequest(w, rq)

	switch err2 := err.(type) {
	case nil:

	case *httpError:
		http.Error(w, err.Error(), err2.status)
	case *notFoundError:
		http.Error(w, err.Error(), http.StatusNotFound)
	case *badRequestError:
		http.Error(w, err.Error(), http.StatusBadRequest)
	case *internalError:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	default:
		fmt.Println("Unhandled error: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

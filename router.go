package grorm

import (
    "encoding/json"
    "fmt"
    "net/http"
    "reflect"
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
		t, po, err := r.server.resolver.resolvePathObject(path)
		if err != nil {
			return err
		}

		if po == nil {
			// TODO: search all objects filter with query parms
			return newHttpError(http.StatusNotImplemented, nil, (*t).Name())
		} else {
			// or get object by id
			w.Header().Set("Content-Type", "application/json")
			e := json.NewEncoder(w)
			e.Encode(po.Interface())
			return nil
		}

		// TODO: or get scalar field
		// TODO: or get list field
	}

	if rq.Method == "POST" {
		// path should just be a type
		// TODO: or add object id to list field
		t, id, err := r.server.resolver.resolvePath(path)
		if err != nil {
			return err
		}
		if id != nil {
			return newBadRequestError(nil, "You can't POST to an Id.")
		}

		// create new object from JSON in body
		po := reflect.New(*t)
		err = copyJsonToObject(rq.Body, po)
		if err != nil {
			return err
		}

		// save the object to DB
		err = r.server.store.save(po.Interface())
		if err != nil {
			return err
		}

		// return JSON including the new id
		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.Encode(po.Interface())
	    return nil
	}

	if rq.Method == "PUT" {
		// TODO: update object with id with values from JSON
		_, po, err := r.server.resolver.resolvePathObject(path)
		if err != nil {
			return err
		}
		if po == nil {
			return newBadRequestError(nil, "No object Id was given.")
		}

		err = copyJsonToObject(rq.Body, *po)
		if err != nil {
			return err
		}
		// save the object to DB
		err = r.server.store.save(po.Interface())
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.Encode((*po).Interface())
		return nil

		// TODO: or update scalar field
		// TODO: or set whole list field
	}

	if rq.Method == "DELETE" {
		// delete object with id
		t, id, err := r.server.resolver.resolvePath(path)
		if err != nil {
			return err
		}
		if id == nil {
			return newBadRequestError(nil, "Missing Id.")
		}

		err = r.server.store.delete((*t).Name(), *id)
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

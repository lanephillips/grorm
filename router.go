package grorm

import (
    "encoding/json"
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

func (r *router) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
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
		if err == errBadId || err == errPathExtra {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == errNotFound {
			http.Error(w, rq.URL.Path + " not found.", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if po == nil {
			// TODO: search all objects filter with query parms
			http.Error(w, (*t).Name(), http.StatusNotImplemented)
			return
		} else {
			// or get object by id
			w.Header().Set("Content-Type", "application/json")
			e := json.NewEncoder(w)
			e.Encode(po.Interface())
			return
		}

		// TODO: or get scalar field
		// TODO: or get list field
	}

	if rq.Method == "POST" {
		// path should just be a type
		// TODO: or add object id to list field
		t, id, err := r.server.resolver.resolvePath(path)
		if err == errBadId || err == errPathExtra {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// TODO: too much error checking, maybe make more specialized implementations of resolvePath
		if id != nil {
			http.Error(w, "You can't POST to an Id.", http.StatusBadRequest)
			return
		}
		if err == errNotFound {
			http.Error(w, rq.URL.Path + " not found.", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// create new object from JSON in body
		po := reflect.New(*t)
		err = copyJsonToObject(rq.Body, po)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// save the object to DB
		err = r.server.store.save(po.Interface())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// return JSON including the new id
		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.Encode(po.Interface())
	    return
	}

	if rq.Method == "PUT" {
		// TODO: update object with id with values from JSON
		_, po, err := r.server.resolver.resolvePathObject(path)
		if err == errBadId || err == errPathExtra {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == errNotFound {
			http.Error(w, rq.URL.Path + " not found.", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if po == nil {
			http.Error(w, "No object Id was given.", http.StatusBadRequest)
			return
		}

		err = copyJsonToObject(rq.Body, *po)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		e.Encode((*po).Interface())
		return

		// TODO: or update scalar field
		// TODO: or set whole list field
	}

	if rq.Method == "DELETE" {
		// delete object with id
		t, id, err := r.server.resolver.resolvePath(path)
		if err == errBadId || err == errPathExtra {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == errNotFound {
			http.Error(w, rq.URL.Path + " not found.", http.StatusNotFound)
			return
		}
		if id == nil {
			http.Error(w, "Missing Id.", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = r.server.store.delete((*t).Name(), *id)
		if err == errNotFound {
			http.Error(w, rq.URL.Path + " not found.", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
		// TODO: or remove object id from a list field
	}

    http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
}
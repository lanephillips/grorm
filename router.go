package grorm

import (
    "fmt"
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
		o := po.Elem()

		// fill in fields from JSON
		// TODO: this all will probably need to be factored out for PUT
		// TODO: accept zero value for unspecified field unless annotated otherwise
		d := json.NewDecoder(rq.Body)
		m := map[string]interface{}{}
		err = d.Decode(&m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for k, v := range m {
			// error if id is specified
			if k == "Id" {
				http.Error(w, "You may not set the 'Id' field.", http.StatusBadRequest)
				return
			}

			f := o.FieldByName(k)
			// error on extra fields
			if !f.IsValid() || !f.CanSet() {
				http.Error(w, fmt.Sprintf("Request body specifies field '%v' which cannot be set.", k), http.StatusBadRequest)
				return
			}

			v2 := reflect.ValueOf(v)
			if !v2.Type().AssignableTo(f.Type()) {
				http.Error(w, fmt.Sprintf("Field '%v' cannot take the value %v.", k, v), http.StatusBadRequest)
				return
			}

			// actually set value
			// TODO: probably should try to catch panic if possible, might have forgotten something above
			f.Set(v2)
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
		// TODO: or update scalar field
		// TODO: or set whole list field
		http.Error(w, "Not implemented.", http.StatusNotImplemented)
		return
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
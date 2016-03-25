package grorm

import (
    "reflect"
    "strconv"
    "strings"
)

type resolver struct {
	server *Server
	types map[string]reflect.Type
}

func newResolver(server *Server) *resolver {
	var r resolver
	r.server = server
	r.types = map[string]reflect.Type{}
	return &r
}

func (r *resolver) registerType(object interface{}, nameOrNil *string) error {
	t := reflect.TypeOf(object)

	if t.Kind() != reflect.Struct {
		return newConfigurationError(nil, "A struct type is required.")
	}

	// TODO: check that struct conforms
	f, ok := t.FieldByName("Id")
	if !ok {
		return newConfigurationError(nil, "An 'Id' field is required.")
	}
	if f.Type.Kind() != reflect.Uint64 {
		return newConfigurationError(nil, "Type of 'Id' must be uint64.")
	}

	// TODO: build ACL objects from field annotations

	// TODO: field names are still capitalized, do we really want Go style to leak through? maybe have pluggable mappers
	if nameOrNil == nil {
		s := strings.ToLower(t.Name())
		nameOrNil = &s
	}

	r.types[*nameOrNil] = t
	// log.Printf("Registered type %v with path \"%v\".\n", t, *nameOrNil)
	return nil
}

func (r *resolver) resolvePath(path []string) (*reflect.Type, *uint64, error) {
	// this happens when we split on / and path starts with /
	if len(path) > 1 && path[0] == "" {
		path = path[1:]
	}

	if len(path) == 0 {
		return nil, nil, newNotFoundError(nil, "/ not found.")
	}

	// look up type name
	name, path := path[0], path[1:]
	t, ok := r.types[name]
	if !ok {
		return nil, nil, newNotFoundError(nil, "Type %v not found.", name)
	}

	if len(path) == 0 {
		// path only went as far as type name
		return &t, nil, nil
	}

	// parse id
	sid, path := path[0], path[1:]
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		return nil, nil, newBadRequestError(err, "Malformed Id.")
	}

	if len(path) > 0 {
		// don't allow extra junk after id
		return nil, nil, newBadRequestError(nil, "Extra chars in path.")
	}

	return &t, &id, nil
}

// value has kind pointer to struct
// will return notFoundError if id parses but retrieves no object
func (r *resolver) resolvePathObject(path []string) (*reflect.Type, *reflect.Value, error) {
	t, id, err := r.resolvePath(path)
	if err != nil {
		return nil, nil, err
	}

	if id == nil {
		return t, nil, nil
	}

	po := reflect.New(*t)
	err = r.server.store.load(*id, po.Interface())
	if err != nil {
		return nil, nil, err
	}

	return t, &po, nil
}

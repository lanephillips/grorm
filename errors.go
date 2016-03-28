package grorm

import "fmt"

// used for errors that occur during web app startup, these are the fault of the package user
type configurationError struct {
	message string
	err error
}

func (e *configurationError) Error() string {
	if e.err != nil {
		return e.message + ": " + e.err.Error()
	}
	return e.message
}

func newConfigurationError(err error, format string, args ...interface{}) error {
	s := fmt.Sprintf(format, args...)
	return &configurationError{ s, err }
}

// represents an error that was probably caused by the client sending garbage
// usually maps to 400
type badRequestError struct {
	message string
	err error
}

func (e *badRequestError) Error() string {
	if e.err != nil {
		return e.message + ": " + e.err.Error()
	}
	return e.message
}

func newBadRequestError(err error, format string, args ...interface{}) error {
	s := fmt.Sprintf(format, args...)
	return &badRequestError{ s, err }
}

// represents an error that is not the client's fault, but requested object simply isn't there
// usually maps to 404
type notFoundError struct {
	message string
	err error
}

func (e *notFoundError) Error() string {
	if e.err != nil {
		return e.message + ": " + e.err.Error()
	}
	return e.message
}

func newNotFoundError(err error, format string, args ...interface{}) error {
	s := fmt.Sprintf(format, args...)
	return &notFoundError{ s, err }
}

// represents an error that really shouldn't happen, these are the package maintainer's fault
// usually maps to 500
type internalError struct {
	message string
	err error
}

func (e *internalError) Error() string {
	if e.err != nil {
		return e.message + ": " + e.err.Error()
	}
	return e.message
}

func newInternalError(err error, format string, args ...interface{}) error {
	s := fmt.Sprintf(format, args...)
	return &internalError{ s, err }
}

// allows us to specify something other than the default status code
type httpError struct {
	status int
	message string
	err error
}

func (e *httpError) Error() string {
	if e.err != nil {
		return e.message + ": " + e.err.Error()
	}
	return e.message
}

func newHttpError(status int, err error, format string, args ...interface{}) error {
	s := fmt.Sprintf(format, args...)
	return &httpError{ status, s, err }
}

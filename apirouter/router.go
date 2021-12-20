package apirouter

import (
	"errors"
)

var ErrInvalidHandler = errors.New("invalid handler")

type Router interface {
	AddRoute(method string, path string, handler interface{}) (Route, error)
}

type Route interface{}

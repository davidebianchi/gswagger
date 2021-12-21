# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### BREAKING CHANGES

- modified apirouter.Router interface
  - sorting addRoute arguments in a different manner: first method and then path
  - AddRoute now returns a Route object and an error
  - added `SwaggerHandler` required method to router interface
  - AddRoute method now take `interface{}` as handler argument instead of `HandlerFunc`. This is to allow the router to be used with other handler types. For the basic gorilla apirouter implementation, only the `func(w http.ResponseWriter, req *http.Request)` type is supported. If you need other handler types, you can create it using the `apirouter.Router` interface.

To migrate, all the router implementation must be updated with the Router interface change.

Before:

```go
type Router interface {
  AddRoute(path, method string, handler HandlerFunc) Route
}
```

After:

```go
type Router interface {
  AddRoute(method, path string, handler interface{}) (Route, error)
  SwaggerHandler(contentType string, json []byte) interface{}
}
```

### Added

- add support for [echo](https://echo.labstack.com/) router


## v0.3.0 - 10-11-2021

### Added

- handle router with path prefix
- add SubRouter method to use a new sub router

## v0.2.0 - 16-10-2021

### BREAKING CHANGES

Introduced the `apirouter.Router` interface, which abstract the used router.
Changed function are:

- Router struct now support `apirouter.Router` interface
- NewRouter function accepted router is an `apirouter.Router`
- AddRawRoute now accept `apirouter.Handler` iterface
- AddRawRoute returns an interface instead of *mux.Router. This interface is the Route returned by specific Router
- AddRoute now accept `apirouter.Handler` iterface
- AddRoute returns an interface instead of *mux.Router. This interface is the Route returned by specific Router

## v0.1.0

Initial release

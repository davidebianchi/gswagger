# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### BREAKING CHANGES

- `apirouter.NewGorillaMuxRouter` is now `gorilla.NewRouter` (exposed by package `github.com/davidebianchi/gswagger/support/gorilla`).

## 0.6.1 - 17-11-2022

### Changed

- change jsonschema lib to `mia-platform/jsonschema v0.1.0`. This update removes the `patternProperties` with `additionalProperties` from all schemas
- remove use of deprecated io/ioutil lib

## 0.6.0 - 04-11-2022

### Added

- Tags support to `router.AddRoute` accepted definition

## 0.5.1 - 03-10-2022

### Fixed

- upgrade deps

## v0.5.0 - 05-08-2022

### Added

- path params are auto generated if not set

## v0.4.0 - 02-08-2022

### Changed

- change jsonschema lib to `invopop/jsonschema v0.5.0`. This updates remove the `additionalProperties: true` from all the schemas, as it is the default value

### BREAKING CHANGES

- modified Router interface by sorting addRoute arguments in a different manner: first method and then path
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
  AddRoute(method, path string, handler HandlerFunc) Route
}
```

### Updates

- kin-openapi@v0.98.0
- go-openapi/swag@v0.21.1
- labstack/echo/v4@v4.7.2

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

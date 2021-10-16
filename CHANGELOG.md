# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

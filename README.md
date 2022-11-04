<div align="center">

[![Build Status][github-actions-svg]][github-actions]
[![Coverage Status](https://coveralls.io/repos/github/davidebianchi/gswagger/badge.svg?branch=main)](https://coveralls.io/github/davidebianchi/gswagger?branch=main)
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]

# gswagger

</div>

Generate an openapi spec dynamically based on the types used to handle request and response.

It works with any router which support handler net/http HandlerFunc compatible.

The routers supported out of the box are:

- [gorilla-mux](https://github.com/gorilla/mux)

This lib uses [kin-openapi] to automatically generate and serve a swagger file.

To convert struct to schemas, we use [jsonschema] library.  
The struct must contains the appropriate struct tags to be inserted in json schema to generate the schema dynamically.  
It is always possible to add a totally custom swagger schema using [kin-openapi].

## Usage

To add a router not handled out of the box, it must implements the [Router interface](./apirouter/router.go).

An example usage of this lib with gorilla mux:

```go
context := context.Background()
muxRouter := mux.NewRouter()

router, err := swagger.NewRouter(apirouter.NewGorillaMuxRouter(muxRouter), swagger.Options{
  Context: context,
  Openapi: &openapi3.T{
    Info: &openapi3.Info{
      Title:   "my title",
      Version: "1.0.0",
    },
  },
})

okHandler := func(w http.ResponseWriter, req *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("OK"))
}

type User struct {
  Name        string   `json:"name" jsonschema:"title=The user name,required" jsonschema_extras:"example=Jane"`
  PhoneNumber int      `json:"phone" jsonschema:"title=mobile number of user"`
  Groups      []string `json:"groups,omitempty" jsonschema:"title=groups of the user,default=users"`
  Address     string   `json:"address" jsonschema:"title=user address"`
}
type Users []User
type errorResponse struct {
  Message string `json:"message"`
}

router.AddRoute(http.MethodPost, "/users", okHandler, swagger.Definitions{
  RequestBody: &swagger.ContentValue{
    Content: swagger.Content{
      "application/json": {Value: User{}},
    },
  },
  Responses: map[int]swagger.ContentValue{
    201: {
      Content: swagger.Content{
        "text/html": {Value: ""},
      },
    },
    401: {
      Content: swagger.Content{
        "application/json": {Value: &errorResponse{}},
      },
      Description: "invalid request",
    },
  },
})

router.AddRoute(http.MethodGet, "/users", okHandler, swagger.Definitions{
  Responses: map[int]swagger.ContentValue{
    200: {
      Content: swagger.Content{
        "application/json": {Value: &[]User{}},
      },
    },
  },
})

carSchema := openapi3.NewObjectSchema().WithProperties(map[string]*openapi3.Schema{
  "foo": openapi3.NewStringSchema(),
  "bar": openapi3.NewIntegerSchema().WithMax(15).WithMin(5),
})
requestBody := openapi3.NewRequestBody().WithJSONSchema(carSchema)
operation := swagger.NewOperation()
operation.AddRequestBody(requestBody)

router.AddRawRoute(http.MethodPost, "/cars", okHandler, operation)
```

This configuration will output the schema shown [here](testdata/users_employees.json).

## Auto generated path params schema

The path params, if not set in schema, are auto generated from the path. Currently, it is supported only the path params like `{myPath}`.

For example, with this use case:

```golang
okHandler := func(w http.ResponseWriter, req *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("OK"))
}

_, err := router.AddRoute(http.MethodGet, "/users/{userId}", okHandler, Definitions{
  Querystring: ParameterValue{
    "query": {
      Schema: &Schema{Value: ""},
    },
  },
})
require.NoError(t, err)

_, err = router.AddRoute(http.MethodGet, "/cars/{carId}/drivers/{driverId}", okHandler, Definitions{})
require.NoError(t, err)
```

The generated oas schema will contains `userId`, `carId` and `driverId` as path params set to string.
If only one params is set, you must specify manually all the path params.

The generated file for this test case is [here](./testdata/params-autofill.json).

## SubRouter

It is possible to create a new sub router from the swagger.Router.
It is possible to add a prefix to all the routes created under the specific router (instead of use the router specific methods, if given, or repeat the prefix for every route).

It could also be useful if you need a sub router to create a group of APIs which use the same middleware (for example,this could be achieved by the SubRouter features of gorilla mux, for example).

To see the SubRouter example, please see the [SubRouter test](./integration_test.go).

### FAQ

1. How to add format `binary`?
Formats `date-time`, `email`, `hostname`, `ipv4`, `ipv6`, `uri` could be added with tag `jsonschema`. Others format could be added with tag `jsonschema_extra`. Not all the formats are supported (see discovered unsupported formats [here](#discovered-unsupported-schema-features)).

1. How to add a swagger with `allOf`?
You can create manually a swagger with `allOf` using the `AddRawRoute` method.

1. How to add a swagger with `anyOf`?
You can create manually a swagger with `anyOf` using the `AddRawRoute` method.

1. How to add a swagger with `oneOf`?
You can create manually a swagger with `oneOf` using the `AddRawRoute` method, or use the [jsonschema] struct tag.

#### Discovered unsupported schema features

*Formats*:

- `uuid` is unsupported by [kin-openapi]

## Versioning

We use [SemVer][semver] for versioning. For the versions available,
see the [tags on this repository](https://github.com/davidebianchi/gswagger/tags).

<!-- Reference -->
[kin-openapi]: https://github.com/getkin/kin-openapi
[jsonschema]: https://github.com/invopop/jsonschemaa
[github-actions]: https://github.com/davidebianchi/gswagger/actions
[github-actions-svg]: https://github.com/davidebianchi/gswagger/workflows/Test%20and%20build/badge.svg
[godoc-svg]: https://godoc.org/github.com/davidebianchi/gswagger?status.svg
[godoc-link]: https://godoc.org/github.com/davidebianchi/gswagger
[go-report-card]: https://goreportcard.com/badge/github.com/davidebianchi/gswagger
[go-report-card-link]: https://goreportcard.com/report/github.com/davidebianchi/gswagger
[semver]: https://semver.org/

## gorilla swagger

Initial test&try to generate a swagger dynamically.

It uses [gorilla-mux](https://github.com/gorilla/mux) and [kin-openapi](https://github.com/getkin/kin-openapi)
to automatically generate and serve a swagger file.

To convert struct to schemas, we use [this library](https://github.com/alecthomas/jsonschema).
The struct should contains the appropriate struct tags to be inserted in json schema.

## Usage

An example usage of this lib:

```go
context := context.Background()
r := mux.NewRouter()
router, _ := gswagger.New(r, gswagger.Options{
  Context: context,
  Openapi: &openapi3.Swagger{
    Info: &openapi3.Info{
      Title:   "my swagger title",
      Version: "1.0.0",
    },
  },
})

okHandler := func(w http.ResponseWriter, req *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("OK"))
}

router.AddRoute(http.MethodPost, "/users", okHandler, Definitions{
  RequestBody: &gswagger.ContentValue{
    Content: gswagger.Content{
      "application/json": {Value: User{}},
    },
  },
  Responses: map[int]gswagger.ContentValue{
    201: {
      Content: gswagger.Content{
        "text/html": {Value: ""},
      },
    },
    401: {
      Content: gswagger.Content{
        "application/json": {Value: &errorResponse{}},
      },
      Description: "invalid request",
    },
  },
})

router.AddRoute(http.MethodGet, "/users", okHandler, Definitions{
  Responses: map[int]ContentValue{
    200: {
      Content: Content{
        "application/json": {Value: &Users{}},
      },
    },
  },
})

router.AddRoute(http.MethodGet, "/employees", okHandler, Definitions{
  Responses: map[int]ContentValue{
    200: {
      Content: Content{
        "application/json": {Value: &Employees{}},
      },
    },
  },
})
```

This configuration will output the schema shown [here](testdata/users_employees.json)

### FAQ

1. How to add format `binary`?
Formats `date-time`, `email`, `hostname`, `ipv4`, `ipv6`, `uri` could be added with tag `jsonschema`. Others format could be added with tag `jsonschema_extra`. Not all the formats are supported (see discovered unsupported formats [here](#discovered-unsupported-schema-features)).

#### Discovered unsupported schema features

*Formats*:

* `uuid` is unsupported by kin-openapi

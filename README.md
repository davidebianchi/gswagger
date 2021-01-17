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
router, _ := swagger.NewRouter(r, gswagger.Options{
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

carSchema := openapi3.NewObjectSchema().WithProperties(map[string]*openapi3.Schema{
  "foo": openapi3.NewStringSchema(),
  "bar": openapi3.NewIntegerSchema().WithMax(15).WithMin(5),
})
operation := swagger.NewOperation()
openapi3.NewRequestBody().WithJSONSchema(carSchema)
operation.AddRequestBody(requestBody)

router.AddRawRoute(http.MethodPost, "/cars", okHandler, operation)
```

This configuration will output the schema shown [here](testdata/users_employees.json)

### FAQ

1. How to add format `binary`?
Formats `date-time`, `email`, `hostname`, `ipv4`, `ipv6`, `uri` could be added with tag `jsonschema`. Others format could be added with tag `jsonschema_extra`. Not all the formats are supported (see discovered unsupported formats [here](#discovered-unsupported-schema-features)).

1. How to add a swagger with `allOf`?
You can create manually a swagger with `allOf` using the `AddRawRoute` method.

#### Discovered unsupported schema features

*Formats*:

* `uuid` is unsupported by kin-openapi

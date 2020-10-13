## gorilla swagger

Initial test&try to generate a swagger dynamically.

It uses [gorilla-mux](https://github.com/gorilla/mux) and [kin-openapi](github.com/getkin/kin-openapi)
to automatically generate and serve a swagger file.

To convert struct to schemas, we use [this library](https://github.com/alecthomas/jsonschema).
The struct should contains the appropriate struct tags to be inserted in json schema.

## gorilla swagger

Initial test&try to generate a swagger dynamically.

It uses [gorilla-mux](https://github.com/gorilla/mux) and [kin-openapi](github.com/getkin/kin-openapi)
to automatically generate and serve a swagger file.

To convert struct to schemas, we use [this library](https://github.com/alecthomas/jsonschema).
The struct should contains the appropriate struct tags to be inserted in json schema.

### FAQ

1. How to add format `binary`?
Formats `date-time`, `email`, `hostname`, `ipv4`, `ipv6`, `uri` could be added with tag `jsonschema`. Others format could be added with tag `jsonschema_extra`. Not all the formats are supported (see discovered unsupported formats [here](#discovered-unsupported-schema-features)).

#### Discovered unsupported schema features

*Formats*:

* `uuid` is unsupported by kin-openapi

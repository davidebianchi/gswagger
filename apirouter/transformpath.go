package apirouter

import (
	"strings"
)

func TransformPathParamsWithColon(path string) string {
	pathParams := strings.Split(path, "/")
	for i, param := range pathParams {
		if strings.HasPrefix(param, ":") {
			pathParams[i] = strings.Replace(param, ":", "{", 1) + "}"
		}
	}
	return strings.Join(pathParams, "/")
}

package swagger

// func validateRequest(r Router, req *http.Request) error {
// 	// Find route
// 	route, pathParams, err := r.swaggerRouter.FindRoute(req.Method, req.URL)
// 	if err != nil {
// 		return err
// 	}

// 	// Validate request
// 	requestValidationInput := &openapi3filter.RequestValidationInput{
// 		Request:    req,
// 		PathParams: pathParams,
// 		Route:      route,
// 		// TODO: add query params
// 	}

// 	return openapi3filter.ValidateRequest(req.Context(), requestValidationInput)
// }

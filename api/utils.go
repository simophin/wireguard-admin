package api

import (
	"fmt"
	"net/http"
)

func mustScanQueryParameter(out interface{}, r *http.Request, name string, format string, optional bool) bool {
	values := r.URL.Query()[name]
	if len(values) == 0 {
		if !optional {
			panic(displayableError{
				Name:        badRequest,
				Description: fmt.Sprintf("Parameter %v is missing", name),
				StatusCode:  400,
			})
		}

		return false
	}

	n, err := fmt.Scanf(format, out)
	if n == 0 || err != nil {
		panic(displayableError{
			Name:        badRequest,
			Description: fmt.Sprintf("Parameter %v is malformed", name),
			StatusCode:  400,
		})
	}

	return true
}

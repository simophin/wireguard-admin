package api

import (
	"fmt"
	"net/http"
	"regexp"
)

func mustScanQueryParameter(out interface{}, r *http.Request, name string, re regexp.Regexp, optional bool) bool {
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

}

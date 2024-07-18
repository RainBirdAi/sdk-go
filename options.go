package sdk

import "net/http"

// requestOptions is a struct with options that can be used within the sdk endpoints
type requestOptions struct {
	headers http.Header
}

// AddOptionsHeaders helps adding the headers to the request
func (ro *requestOptions) AddOptionsHeaders(req *http.Request) {
	if ro != nil {
		for k, vals := range ro.headers {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
	}
}

// CreateOption add optional parameter to `createFile` mutation.
type CreateOption func(o requestOptions) requestOptions

// AddHeaders assigns the headers to the requestOptions struct
func AddHeaders(headers http.Header) CreateOption {
	return func(o requestOptions) requestOptions {
		o.headers = headers.Clone()
		return o
	}
}

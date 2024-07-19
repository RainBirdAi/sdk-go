package sdk

import "net/http"

// requestOptions is a struct with options that can be used within the sdk endpoints
type queryOption struct {
	headers http.Header
}

// addOptionsHeaders helps adding the headers to the request
func (ro *queryOption) addOptionsHeaders(req *http.Request) {
	if ro != nil {
		for k, vals := range ro.headers {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
	}
}

// RequestOption add optional parameter to sdk requests
type QueryOption func(o queryOption) queryOption

// AddHeaders assigns the headers to the requestOptions struct
func AddHeaders(headers http.Header) QueryOption {
	return func(o queryOption) queryOption {
		if headers != nil {
			o.headers = headers.Clone()
			for k, vals := range headers {
				for _, v := range vals {
					o.headers.Add(k, v)
				}
			}
		}
		return o
	}
}

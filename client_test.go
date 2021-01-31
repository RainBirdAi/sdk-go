package sdk

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientNewSessionValidation(t *testing.T) {
	testCases := []struct {
		description string
		client      *Client
		kmid        string
		expectErr   error
	}{
		{
			description: "Missing API Key",
			client: &Client{
				EnvironmentURL: EnvCommunity,
			},
			kmid:      "12345678-1234-1234-1234567890ab",
			expectErr: ErrClientMissingAPIKey,
		},
		{
			description: "Missing EnvironmentURL",
			client: &Client{
				APIKey: "abcdefgh-abcd-abcd-abcdefghijkl",
			},
			kmid:      "12345678-1234-1234-1234567890ab",
			expectErr: ErrClientMissingEnvironmentURL,
		},
		{
			description: "Missing KMID",
			client: &Client{
				EnvironmentURL: EnvCommunity,
				APIKey:         "abcdefgh-abcd-abcd-abcdefghijkl",
			},
			expectErr: ErrNewSessionInvalidKMID,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			result, err := tc.client.NewSession(tc.kmid, "")
			assert.Nil(t, result)
			assert.Equal(t, tc.expectErr, err)
		})
	}
}

func TestClientNewSessionStartCalls(t *testing.T) {
	testCases := []struct {
		description string
		apiKey      string
		keyEncoded  string
		kmid        string
		contextID   string

		expectCallURI string
		returnBody    string
		returnCode    int

		expectCalls int
		expectID    string
		expectErr   error
	}{
		{
			description: "Missing API Key",
			kmid:        "12345678-1234-1234-1234567890ab",
			expectErr:   ErrClientMissingAPIKey,
		},
		{
			description: "Missing KMID",
			apiKey:      "abcdefgh-abcd-abcd-abcdefghijkl",
			expectErr:   ErrNewSessionInvalidKMID,
		},
		{
			description: "Simple success",
			apiKey:      "abcdefgh-abcd-abcd-abcdefghijkl",
			keyEncoded:  "Basic YWJjZGVmZ2gtYWJjZC1hYmNkLWFiY2RlZmdoaWprbDo=",
			kmid:        "12345678-1234-1234-1234567890ab",

			expectCallURI: "/start/12345678-1234-1234-1234567890ab",
			returnBody:    `{"id":"success-id"}`,
			returnCode:    http.StatusOK,

			expectCalls: 1,
			expectID:    "success-id",
			expectErr:   nil,
		},
		{
			description: "Correctly request context",
			apiKey:      "abcdefgh-abcd-abcd-abcdefghijkl",
			keyEncoded:  "Basic YWJjZGVmZ2gtYWJjZC1hYmNkLWFiY2RlZmdoaWprbDo=",
			kmid:        "12345678-1234-1234-1234567890ab",
			contextID:   "foo",

			expectCallURI: "/start/12345678-1234-1234-1234567890ab?contextid=foo",
			returnBody:    `{"id":"success-id"}`,
			returnCode:    http.StatusOK,

			expectCalls: 1,
			expectID:    "success-id",
			expectErr:   nil,
		},
		{
			description: "Handle 400",
			apiKey:      "abcdefgh-abcd-abcd-abcdefghijkl",
			keyEncoded:  "Basic YWJjZGVmZ2gtYWJjZC1hYmNkLWFiY2RlZmdoaWprbDo=",
			kmid:        "12345678-1234-1234-1234567890ab",

			expectCallURI: "/start/12345678-1234-1234-1234567890ab",
			returnBody:    `Bad request!`,
			returnCode:    http.StatusBadRequest,

			expectCalls: 1,
			expectID:    "",
			expectErr:   errors.New("API returned an error 400: Bad request!"),
		},
		// TODO: Engine header
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			calls := 0

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++

					assert.Equal(t, tc.expectCallURI, r.RequestURI)
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, r.Header.Get("Authorization"), tc.keyEncoded)
					assert.Equal(t, r.Header.Get("Accept"), "application/json")
					// TODO: Engine

					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(tc.returnCode)
					w.Write([]byte(tc.returnBody))
				}),
			)
			defer srv.Close()

			client := &Client{
				APIKey:         tc.apiKey,
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}

			result, err := client.NewSession(tc.kmid, tc.contextID)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectCalls, calls)
			if tc.expectID != "" {
				require.NotNil(t, result)
				assert.Equal(t, tc.expectID, result.ID)
			}
		})
	}
}

func TestClientHTTP(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Not used
		}),
	)
	defer srv.Close()
	custom := srv.Client()

	testCases := []struct {
		description string
		client      *Client
		expectRet   *http.Client
	}{
		{
			description: "Default",
			client:      &Client{},
			expectRet:   http.DefaultClient,
		},
		{
			description: "Custom",
			client: &Client{
				HTTPClient: custom,
			},
			expectRet: custom,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expectRet, tc.client.HTTP())
		})
	}
}

func TestClientVersion(t *testing.T) {
	testCases := []struct {
		description string

		expectURI    string
		responseCode int
		responseBody string

		expectCalls int64
		expectRet   string
		expectErr   error
	}{
		{
			description: "Simple success",

			expectURI:    "/version",
			responseCode: http.StatusOK,
			responseBody: "2.3.4",

			expectCalls: 1,
			expectRet:   "2.3.4",
			expectErr:   nil,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			var calls int64

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					atomic.AddInt64(&calls, 1)

					assert.Equal(t, tc.expectURI, r.RequestURI)
					assert.Equal(t, http.MethodGet, r.Method)

					w.WriteHeader(tc.responseCode)
					w.Write([]byte(tc.responseBody))
				}),
			)
			defer srv.Close()

			client := Client{
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}

			result, err := client.Version()
			assert.Equal(t, tc.expectRet, result)
			assert.Equal(t, tc.expectErr, err)
		})
	}
}

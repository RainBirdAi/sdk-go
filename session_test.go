package sdk

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionInject(t *testing.T) {
	testCases := []struct {
		description  string
		facts        []InjectFact
		responseCode int
		expectCalls  int64
		expectBody   string
		expectErr    error
	}{
		{
			description:  "Empty",
			facts:        []InjectFact{},
			responseCode: http.StatusOK,
			expectCalls:  2,
			expectBody:   "[]",
			expectErr:    nil,
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

					switch calls {
					case 1:
						// Start (as part of making the session)
						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"id":"success-id"}`))
					case 2:
						// Inject endpoint
						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
						assert.Equal(t, "/success-id/inject", r.RequestURI)
						assert.Equal(t, http.MethodPost, r.Method)

						body, err := ioutil.ReadAll(r.Body)
						require.Nil(t, err)
						assert.Equal(t, tc.expectBody, string(body))

						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(tc.responseCode)
						w.Write([]byte(""))
					default:
						t.Fatal("Unexpected call")
					}
				}),
			)
			defer srv.Close()

			client := &Client{
				APIKey:         "1234567890-1234-1234-1234-1234567890ab",
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}

			session, err := client.NewSession("kmid", "")
			require.Nil(t, err)

			err = session.Inject(tc.facts)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectCalls, calls)
		})
	}
}

func TestSessionQuery(t *testing.T) {
	testCases := []struct {
		description string
		sub         string
		rel         string
		obj         string

		responseCode int
		responseBody string
		expectCalls  int64
		expectBody   string

		expectQuestion *Question
		expectAnswers  *[]Answer
		expectErr      error
	}{
		{
			description: "Empty",
			sub:         "",
			rel:         "",
			obj:         "",

			expectCalls: 1,
			expectErr:   ErrQueryBlankRelationship,
		},
		{
			description: "John - lives in - ? leading to answer",
			sub:         "John",
			rel:         "lives in",
			obj:         "",

			responseCode: http.StatusOK,
			responseBody: `{
				"result": [{
					"certainty": 100,
					"factID": "thisisthefactid",
					"object": "England",
					"relationship": "lives in",
					"subject": "John"
				}]
			}`,
			expectCalls: 2,
			expectBody:  `{"subject":"John","relationship":"lives in"}`,

			expectQuestion: nil,
			expectAnswers: &[]Answer{
				{
					Subject:      "John",
					Relationship: "lives in",
					Object:       "England",
					Certainty:    100,
					FactID:       "thisisthefactid",
				},
			},
			expectErr: nil,
		},
		{
			description: "Handle bad request",
			sub:         "John",
			rel:         "lives in",
			obj:         "",

			responseCode: http.StatusBadRequest,
			responseBody: "Foo bar baz",
			expectCalls:  2,
			expectBody:   `{"subject":"John","relationship":"lives in"}`,

			expectQuestion: nil,
			expectAnswers:  nil,
			expectErr:      fmt.Errorf("API returned error 400: Foo bar baz"),
		},
		{
			description: "Handle internal server error",
			sub:         "John",
			rel:         "lives in",
			obj:         "",

			responseCode: http.StatusInternalServerError,
			responseBody: `{ "error": "some JSON returned" }`,
			expectCalls:  2,
			expectBody:   `{"subject":"John","relationship":"lives in"}`,

			expectQuestion: nil,
			expectAnswers:  nil,
			expectErr:      fmt.Errorf(`API returned error 500: { "error": "some JSON returned" }`),
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

					switch calls {
					case 1:
						// Start (as part of making the session)
						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"id":"success-id"}`))
					case 2:
						// Query endpoint
						assert.Equal(t, "application/json", r.Header.Get("Accept"))
						assert.Equal(t, "/success-id/query", r.RequestURI)
						assert.Equal(t, http.MethodPost, r.Method)

						body, err := ioutil.ReadAll(r.Body)
						require.Nil(t, err)
						assert.Equal(t, tc.expectBody, string(body))

						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(tc.responseCode)
						w.Write([]byte(tc.responseBody))
					default:
						t.Fatal("Unexpected call")
					}
				}),
			)
			defer srv.Close()

			client := &Client{
				APIKey:         "1234567890-1234-1234-1234-1234567890ab",
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}

			session, err := client.NewSession("kmid", "")
			require.Nil(t, err)

			question, answers, err := session.Query(tc.sub, tc.rel, tc.obj)
			assert.Equal(t, tc.expectQuestion, question)
			assert.Equal(t, tc.expectAnswers, answers)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectCalls, calls)
		})
	}
}

func TestSessionResponse(t *testing.T) {
	testCases := []struct {
		description string
		answers     []QAnswer

		expectCalls  int64
		expectBody   string
		responseCode int
		responseBody string

		expectQuestion *Question
		expectAnswers  *[]Answer
		expectErr      error
	}{
		{
			description: "Simple response that returns a question",
			answers: []QAnswer{
				{
					Subject:      "John",
					Relationship: "Speaks",
					Object:       "English",
					CF:           "100",
				},
			},

			expectCalls:  2,
			expectBody:   `{"answers":[{"subject":"John","relationship":"Speaks","object":"English","cf":"100"}]}`,
			responseCode: http.StatusOK,
			responseBody: `{
				"question": {
					"subject":"John",
					"dataType":"string",
					"relationship":"lives in",
					"type":"Second Form Object",
					"plural":false,
					"allowCF":true,
					"allowUnknown":false,
					"canAdd":true,
					"prompt":"Where does John live?",
					"knownAnswers":[]
				}
			}`,

			expectQuestion: &Question{
				AllowCF:      true,
				AllowUnknown: false,
				CanAdd:       true,
				Concepts:     nil,
				DataType:     "string",
				KnownAnswers: []KnownAnswer{},
				Plural:       false,
				Prompt:       "Where does John live?",
				Relationship: "lives in",
				Subject:      "John",
				Object:       "",
				Type:         "Second Form Object",
			},
			expectAnswers: nil,
			expectErr:     nil,
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

					switch calls {
					case 1:
						// Start (as part of making the session)
						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"id":"success-id"}`))
					case 2:
						// Query endpoint
						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
						assert.Equal(t, "/success-id/response", r.RequestURI)
						assert.Equal(t, http.MethodPost, r.Method)

						body, err := ioutil.ReadAll(r.Body)
						require.Nil(t, err)
						assert.Equal(t, tc.expectBody, string(body))

						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(tc.responseCode)
						w.Write([]byte(tc.responseBody))
					default:
						t.Fatal("Unexpected call")
					}
				}),
			)
			defer srv.Close()

			client := &Client{
				APIKey:         "1234567890-1234-1234-1234-1234567890ab",
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}

			session, err := client.NewSession("kmid", "")
			require.Nil(t, err)

			question, answers, err := session.Response(tc.answers)
			assert.Equal(t, tc.expectQuestion, question)
			assert.Equal(t, tc.expectAnswers, answers)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectCalls, calls)
		})
	}
}

func TestSessionUndo(t *testing.T) {
	testCases := []struct {
		description string

		responseCode int
		responseBody string

		expectQuestion *Question
		expectAnswers  *[]Answer
		expectErr      error
	}{
		{
			description: "Success back to question",

			responseCode: http.StatusOK,
			responseBody: `{
				"question": {
					"subject":"John",
					"dataType":"string",
					"relationship":"lives in",
					"type":"Second Form Object",
					"plural":false,
					"allowCF":true,
					"allowUnknown":false,
					"canAdd":true,
					"prompt":"Where does John live?",
					"knownAnswers":[]
				}
			}`,
			expectQuestion: &Question{
				AllowCF:      true,
				AllowUnknown: false,
				CanAdd:       true,
				Concepts:     nil,
				DataType:     "string",
				KnownAnswers: []KnownAnswer{},
				Plural:       false,
				Prompt:       "Where does John live?",
				Relationship: "lives in",
				Subject:      "John",
				Object:       "",
				Type:         "Second Form Object",
			},
			expectAnswers: nil,
			expectErr:     nil,
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

					switch calls {
					case 1:
						// Start (as part of making the session)
						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"id":"success-id"}`))
					case 2:
						// Query endpoint
						assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
						assert.Equal(t, "/success-id/undo", r.RequestURI)
						assert.Equal(t, http.MethodPost, r.Method)

						body, err := ioutil.ReadAll(r.Body)
						require.Nil(t, err)
						assert.Equal(t, "{}", string(body))

						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(tc.responseCode)
						w.Write([]byte(tc.responseBody))
					default:
						t.Fatal("Unexpected call")
					}
				}),
			)
			defer srv.Close()

			client := &Client{
				APIKey:         "1234567890-1234-1234-1234-1234567890ab",
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}

			session, err := client.NewSession("kmid", "")
			require.Nil(t, err)

			question, answers, err := session.Undo()
			assert.Equal(t, tc.expectQuestion, question)
			assert.Equal(t, tc.expectAnswers, answers)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, int64(2), calls)
		})
	}
}

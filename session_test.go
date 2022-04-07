package sdk

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionInject(t *testing.T) {
	stringPtr := func(s string) *string { return &s }

	testCases := []struct {
		description string
		facts       []InjectFact
		engine      *string

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
		{
			description:  "Specific engine",
			facts:        []InjectFact{},
			engine:       stringPtr("alternate"),
			responseCode: http.StatusOK,
			expectCalls:  2,
			expectBody:   "[]",
			expectErr:    nil,
		},
		{
			description: "Some facts",
			facts: []InjectFact{
				{
					Subject:      "foo1",
					Relationship: "bar1",
					Object:       "baz1",
					Certainty:    "121",
				},
				{
					Subject:      "foo2",
					Relationship: "bar2",
					Object:       "baz2",
					Certainty:    "122",
				},
			},
			responseCode: http.StatusOK,
			expectCalls:  2,
			expectBody:   `[{"subject":"foo1","relationship":"bar1","object":"baz1","cf":"121"},{"subject":"foo2","relationship":"bar2","object":"baz2","cf":"122"}]`,
			expectErr:    nil,
		},
		{
			description:  "Bad request",
			facts:        []InjectFact{},
			responseCode: http.StatusBadRequest,
			expectCalls:  2,
			expectBody:   "[]",
			expectErr:    errors.New("API returned error 400"),
		},
		{
			description:  "Bad request",
			facts:        []InjectFact{},
			responseCode: http.StatusInternalServerError,
			expectCalls:  2,
			expectBody:   "[]",
			expectErr:    errors.New("API returned error 500"),
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
						assert.Equal(
							t,
							"application/json",
							r.Header.Get("Content-Type"),
						)
						assert.Equal(t, "/success-id/inject", r.RequestURI)
						assert.Equal(t, http.MethodPost, r.Method)
						if tc.engine != nil {
							assert.Equal(
								t,
								"alternate",
								r.Header.Get("x-rainbird-engine"),
							)
						}

						body, err := ioutil.ReadAll(r.Body)
						require.Nil(t, err)
						assert.Equal(t, tc.expectBody, string(body))

						w.Header().Add("Content-Type", "application/json")
						w.WriteHeader(tc.responseCode)
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
			if tc.engine != nil {
				client.Engine = *tc.engine
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
	stringPtr := func(s string) *string { return &s }

	testCases := []struct {
		description string
		sub         string
		rel         string
		obj         string
		engine      *string

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
			description: "John - lives in - ? leading to answer (alt engine)",
			sub:         "John",
			rel:         "lives in",
			obj:         "",
			engine:      stringPtr("alternate"),

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
						if tc.engine != nil {
							assert.Equal(
								t,
								"alternate",
								r.Header.Get("x-rainbird-engine"),
							)
						}

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
			if tc.engine != nil {
				client.Engine = *tc.engine
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
	stringPtr := func(s string) *string { return &s }

	testCases := []struct {
		description string
		answers     []QAnswer
		engine      *string

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
		{
			description: "Simple response that returns a question (alt engine)",
			answers: []QAnswer{
				{
					Subject:      "John",
					Relationship: "Speaks",
					Object:       "English",
					CF:           "100",
				},
			},
			engine: stringPtr("alternate"),

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
		{
			description: "Bad request",
			answers: []QAnswer{
				{
					Subject:      "John",
					Relationship: "Speaks",
					Object:       "English",
					CF:           "100",
				},
			},

			expectCalls:    2,
			expectBody:     `{"answers":[{"subject":"John","relationship":"Speaks","object":"English","cf":"100"}]}`,
			responseCode:   http.StatusBadRequest,
			responseBody:   "Foo bar baz",
			expectQuestion: nil,
			expectAnswers:  nil,
			expectErr:      errors.New("API returned error 400: Foo bar baz"),
		},
		{
			description: "Internal server error",
			answers: []QAnswer{
				{
					Subject:      "John",
					Relationship: "Speaks",
					Object:       "English",
					CF:           "100",
				},
			},

			expectCalls:    2,
			expectBody:     `{"answers":[{"subject":"John","relationship":"Speaks","object":"English","cf":"100"}]}`,
			responseCode:   http.StatusInternalServerError,
			responseBody:   "Foo bar baz",
			expectQuestion: nil,
			expectAnswers:  nil,
			expectErr:      errors.New("API returned error 500: Foo bar baz"),
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
						if tc.engine != nil {
							assert.Equal(
								t,
								"alternate",
								r.Header.Get("x-rainbird-engine"),
							)
						}

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
			if tc.engine != nil {
				client.Engine = *tc.engine
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
	stringPtr := func(s string) *string { return &s }

	testCases := []struct {
		description string
		engine      *string

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
		{
			description: "Success back to question (alt engine)",
			engine:      stringPtr("alternate"),

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
		{
			description: "Bad request",

			responseCode:   http.StatusBadRequest,
			responseBody:   "Foo bar baz",
			expectQuestion: nil,
			expectAnswers:  nil,
			expectErr:      errors.New("API returned error 400: Foo bar baz"),
		},
		{
			description: "Internal server error",

			responseCode:   http.StatusInternalServerError,
			responseBody:   "Foo bar baz",
			expectQuestion: nil,
			expectAnswers:  nil,
			expectErr:      errors.New("API returned error 500: Foo bar baz"),
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
						if tc.engine != nil {
							assert.Equal(
								t,
								"alternate",
								r.Header.Get("x-rainbird-engine"),
							)
						}

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
			if tc.engine != nil {
				client.Engine = *tc.engine
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

func TestInteractionLog(t *testing.T) {
	stringPtr := func(s string) *string { return &s }
	sessionID := "1234-5678"
	time, _ := time.Parse("2006-01-02T03:04:05", "2022-02-15T00:00:00")

	testCases := []struct {
		description          string
		kmid                 string
		engine               *string
		responseBody         *string
		responseCode         int
		expectInteractionLog []InteractionEvent
		expectErr            error
	}{
		{
			description:  "Correct interaction log for Start returned",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`[
				{
					"event": "start", 
					"values": {
						"start": {
							"sessionID": "1",
							"useDraft": true,
							"kmVersionID": "1234-5678"
						}
					},
					"created": "2022-02-15T00:00:00Z"
				}
			]`),
			expectInteractionLog: []InteractionEvent{
				{
					Event:   StartEvent,
					Created: time,
					Data: Start{
						SessionID:   "1",
						KmVersionID: "1234-5678",
						UseDraft:    true,
					},
				},
			},
			expectErr: nil,
		},
		{
			description:  "Correct interaction log for Query, Questions returned",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`[
				{
					"event": "query", 
					"values": {
						"query": {
							"subject": "Dan",
							"relationship": "speaks"
						}
					},
					"created": "2022-02-15T00:00:00Z"
				},
				{
					"event": "question", 
					"values": {
						"questions": [
							{
								"subject":"Dan",
								"dataType":"string",
								"relationship":"speaks",
								"type":"Second Form Object",
								"plural":false,
								"allowCF":true,
								"allowUnknown":false,
								"canAdd":true,
								"prompt":"Where does Dan live?",
								"knownAnswers":[]
							},
							{
								"subject":"Tom",
								"dataType":"string",
								"relationship":"speaks",
								"type":"Second Form Object",
								"plural":false,
								"allowCF":true,
								"allowUnknown":false,
								"canAdd":true,
								"prompt":"Where does Tom live?",
								"knownAnswers":[]
							}
						]
					},
					"created": "2022-02-15T00:00:00Z"
				}
			]`),
			expectInteractionLog: []InteractionEvent{
				{
					Event:   QueryEvent,
					Created: time,
					Data: Query{
						Subject:      stringPtr("Dan"),
						Object:       nil,
						Relationship: "speaks",
					},
				},
				{
					Event:   QuestionEvent,
					Created: time,
					Data: []Question{
						{
							AllowCF:      true,
							AllowUnknown: false,
							CanAdd:       true,
							Concepts:     nil,
							DataType:     "string",
							KnownAnswers: []KnownAnswer{},
							Plural:       false,
							Prompt:       "Where does Dan live?",
							Relationship: "speaks",
							Subject:      "Dan",
							Object:       "",
							Type:         "Second Form Object",
						},
						{
							AllowCF:      true,
							AllowUnknown: false,
							CanAdd:       true,
							Concepts:     nil,
							DataType:     "string",
							KnownAnswers: []KnownAnswer{},
							Plural:       false,
							Prompt:       "Where does Tom live?",
							Relationship: "speaks",
							Subject:      "Tom",
							Object:       "",
							Type:         "Second Form Object",
						},
					},
				},
			},
			expectErr: nil,
		},
		{
			description:  "Correct interaction log for Injects returned",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`[
				{
					"event": "inject", 
					"values": {
						"facts": [
							{
								"subject": "Dan",
								"object": "English",
								"relationship": "speaks",
								"cf": "100"
							},
							{
								"subject": "Dan",
								"object": "German",
								"relationship": "speaks",
								"cf": "90"
							},
							{
								"subject": "Dan",
								"object": "French",
								"relationship": "speaks",
								"cf": "80"
							}
						]
					},
					"created": "2022-02-15T00:00:00Z"
				}
			]`),
			expectInteractionLog: []InteractionEvent{
				{
					Event:   InjectEvent,
					Created: time,
					Data: []InjectFact{
						{
							Subject:      "Dan",
							Object:       "English",
							Relationship: "speaks",
							Certainty:    "100",
						},
						{
							Subject:      "Dan",
							Object:       "German",
							Relationship: "speaks",
							Certainty:    "90",
						},
						{
							Subject:      "Dan",
							Object:       "French",
							Relationship: "speaks",
							Certainty:    "80",
						},
					},
				},
			},
			expectErr: nil,
		},
		{
			description:  "Correct interaction log for Answers returned",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`[
				{
					"event": "answer", 
					"values": {
						"answers": [
							{
								"subject": "Dan",
								"object": "English",
								"relationship": "speaks",
								"cf": "100",
								"answer": "England"
							},
							{
								"subject": "Dan",
								"object": "German",
								"relationship": "speaks",
								"cf": "90"
							},
							{
								"subject": "Dan",
								"object": "French",
								"relationship": "speaks",
								"cf": "80"
							}
						]
					},
					"created": "2022-02-15T00:00:00Z"
				}
			]`),
			expectInteractionLog: []InteractionEvent{
				{
					Event:   AnswerEvent,
					Created: time,
					Data: []QAnswer{
						{
							Subject:      "Dan",
							Object:       "English",
							Relationship: "speaks",
							CF:           "100",
							Answer:       "England",
						},
						{
							Subject:      "Dan",
							Object:       "German",
							Relationship: "speaks",
							CF:           "90",
							Answer:       "",
						},
						{
							Subject:      "Dan",
							Object:       "French",
							Relationship: "speaks",
							CF:           "80",
							Answer:       "",
						},
					},
				},
			},
			expectErr: nil,
		},
		{
			description:  "Correct interaction log for Datasources, Result returned",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`[
				{
					"event": "datasource",
					"values": {
						"datasources": [
							{
								"relationship": "speaks",
								"certainty": 100
							},
							{
								"relationship": "speaks",
								"certainty": 90
							},
							{
								"relationship": "speaks",
								"certainty": 80
							}
						]
					},
					"created": "2022-02-15T00:00:00Z"
				},
				{
					"event": "result",
					"values": {
						"results": [
							{
								"certainty": 100,
								"factID": "thisisthefactid",
								"object": "England",
								"relationship": "lives in",
								"subject": "Dan"
							},
							{
								"certainty": 90,
								"factID": "thisisthefactid",
								"object": "England",
								"relationship": "lives in",
								"subject": "Tom"
							}
						]
					},
					"created": "2022-02-15T00:00:00Z"
				}
			]`),
			expectInteractionLog: []InteractionEvent{
				{
					Event:   DatasourceEvent,
					Created: time,
					Data: []Datasource{
						{
							Relationship: "speaks",
							Certainty:    100,
						},
						{
							Relationship: "speaks",
							Certainty:    90,
						},
						{
							Relationship: "speaks",
							Certainty:    80,
						},
					},
				},
				{
					Event:   ResultEvent,
					Created: time,
					Data: []Answer{
						{
							Subject:      "Dan",
							Relationship: "lives in",
							Object:       "England",
							Certainty:    100,
							FactID:       "thisisthefactid",
						},
						{
							Subject:      "Tom",
							Relationship: "lives in",
							Object:       "England",
							Certainty:    90,
							FactID:       "thisisthefactid",
						},
					},
				},
			},
			expectErr: nil,
		},
		{
			description:          "Bad request",
			responseCode:         http.StatusBadRequest,
			responseBody:         stringPtr("Foo bar baz"),
			expectInteractionLog: nil,
			expectErr:            errors.New("API returned error 400: Foo bar baz"),
		},
		{
			description:          "Internal server error",
			responseCode:         http.StatusInternalServerError,
			responseBody:         stringPtr("Foo bar baz"),
			expectInteractionLog: nil,
			expectErr:            errors.New("API returned error 500: Foo bar baz"),
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "/analysis/interactions/"+sessionID, r.RequestURI)

					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(tc.responseCode)
					w.Write([]byte(*tc.responseBody))
				}),
			)
			defer srv.Close()

			client := &Client{
				APIKey:         "1234567890-1234-1234-1234-1234567890ab",
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}

			if tc.engine != nil {
				client.Engine = *tc.engine
			}

			session, err := client.ResumeSession(sessionID)
			require.Nil(t, err)

			interactionLog, err := session.Interactions(nil)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectInteractionLog, interactionLog)
		})
	}
}

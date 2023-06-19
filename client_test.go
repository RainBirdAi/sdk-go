package sdk

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stringPtr(s string) *string { return &s }
func intPtr(s int) *int          { return &s }
func boolPtr(s bool) *bool       { return &s }

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
			result, err := tc.client.NewSession(tc.kmid, "", nil, nil)
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
		engine      *string
		useDraft    *bool
		version     *int

		expectCallURI string
		returnBody    *string
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
			returnBody:    stringPtr(`{"id":"success-id"}`),
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
			returnBody:    stringPtr(`{"id":"success-id"}`),
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
			returnBody:    stringPtr(`Bad request`),
			returnCode:    http.StatusBadRequest,

			expectCalls: 1,
			expectID:    "",
			expectErr:   errors.New("API returned an error 400: Bad request"),
		},
		{
			description: "Handle good code but no ID",
			apiKey:      "abcdefgh-abcd-abcd-abcdefghijkl",
			keyEncoded:  "Basic YWJjZGVmZ2gtYWJjZC1hYmNkLWFiY2RlZmdoaWprbDo=",
			kmid:        "12345678-1234-1234-1234567890ab",

			expectCallURI: "/start/12345678-1234-1234-1234567890ab",
			returnBody:    stringPtr(`{}`),
			returnCode:    http.StatusOK,

			expectCalls: 1,
			expectID:    "",
			expectErr:   errors.New("API returned no error but no ID either"),
		},
		{
			description: "Specific engine",
			apiKey:      "abcdefgh-abcd-abcd-abcdefghijkl",
			keyEncoded:  "Basic YWJjZGVmZ2gtYWJjZC1hYmNkLWFiY2RlZmdoaWprbDo=",
			kmid:        "12345678-1234-1234-1234567890ab",
			engine:      stringPtr("alternateengine"),

			expectCallURI: "/start/12345678-1234-1234-1234567890ab",
			returnBody:    stringPtr(`{}`),
			returnCode:    http.StatusOK,

			expectCalls: 1,
			expectID:    "",
			expectErr:   errors.New("API returned no error but no ID either"),
		},
		{
			description: "Success with useDraft",
			apiKey:      "abcdefgh-abcd-abcd-abcdefghijkl",
			keyEncoded:  "Basic YWJjZGVmZ2gtYWJjZC1hYmNkLWFiY2RlZmdoaWprbDo=",
			kmid:        "12345678-1234-1234-1234567890ab",
			useDraft:    boolPtr(true),

			expectCallURI: "/start/12345678-1234-1234-1234567890ab?useDraft=true",
			returnBody:    stringPtr(`{"id":"success-id"}`),
			returnCode:    http.StatusOK,

			expectCalls: 1,
			expectID:    "success-id",
			expectErr:   nil,
		},
		{
			description: "Success with version",
			apiKey:      "abcdefgh-abcd-abcd-abcdefghijkl",
			keyEncoded:  "Basic YWJjZGVmZ2gtYWJjZC1hYmNkLWFiY2RlZmdoaWprbDo=",
			kmid:        "12345678-1234-1234-1234567890ab",
			version:     intPtr(1),

			expectCallURI: "/start/12345678-1234-1234-1234567890ab?version=1",
			returnBody:    stringPtr(`{"id":"success-id"}`),
			returnCode:    http.StatusOK,

			expectCalls: 1,
			expectID:    "success-id",
			expectErr:   nil,
		},
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
					if tc.engine != nil {
						assert.Equal(
							t,
							*tc.engine,
							r.Header.Get("x-rainbird-engine"),
						)
					}

					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(tc.returnCode)

					if tc.returnBody != nil {
						w.Write([]byte(*tc.returnBody))
					}
				}),
			)
			defer srv.Close()

			client := &Client{
				APIKey:         tc.apiKey,
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}
			if tc.engine != nil {
				client.Engine = *tc.engine
			}

			result, err := client.NewSession(tc.kmid, tc.contextID, tc.useDraft, tc.version)
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
		{
			description: "Error code",

			expectURI:    "/version",
			responseCode: http.StatusInternalServerError,
			responseBody: "Foobarbaz",

			expectCalls: 1,
			expectRet:   "",
			expectErr:   errors.New("API returned error code 500: Foobarbaz"),
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

func TestEvidence(t *testing.T) {
	sessionID := "1234"
	factID := "WA:RF:1234"

	testCases := []struct {
		description    string
		kmid           string
		engine         *string
		responseBody   *string
		responseCode   int
		expectEvidence *Evidence
		expectErr      error
	}{
		{
			description:    "Bad request",
			responseCode:   http.StatusBadRequest,
			responseBody:   stringPtr("Foo bar baz"),
			expectEvidence: nil,
			expectErr:      errors.New("API returned error 400: Foo bar baz"),
		},
		{
			description:    "Internal server error",
			responseCode:   http.StatusInternalServerError,
			responseBody:   stringPtr("Foo bar baz"),
			expectEvidence: nil,
			expectErr:      errors.New("API returned error 500: Foo bar baz"),
		},
		{
			description:  "Returns correct evidence for rule",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`{
				"factID": "WA:RF:1234",
				"source": "rule",
				"fact": {
					"subject": {
						"type": "Person",
						"value": "Dan",
						"dataType": "string"
					},
					"relationship": {
						"type": "might speak"
					},
					"object": {
						"type": "Language",
						"value": "English",
						"dataType": "string"
					},
					"certainty": 100
				},
				"time": 123456789,
				"rule": {
					"bindings": {
						"S": "Dan",
						"O": 100,
						"COUNTRY": "England"
					},
					"conditions": [
						{
							"subject": "Dan",
							"relationship": "lives in",
							"object": "England",
							"salience": 100,
							"certainty": 100,
							"factID": "WA:RF:1234",
							"objectType": "string",
							"factKey": "30d74cc5-7eb6-4c88-8e12-4897b47e3ee7"
						},
						{
							"subject": "England",
							"relationship": "has national language",
							"object": "English",
							"salience": 100,
							"certainty": 100,
							"factID": "WA:RF:1234",
							"objectType": "string"
						},
						{
							"wasMet": true,
							"salience": 100,
							"expression": {
								"text": "1 gte 0"
							}
						}
					]
				}
			}`),
			expectEvidence: &Evidence{
				Fact: Fact{
					ID:     factID,
					Source: "rule",
					Subject: ConceptInstance{
						Type:     stringPtr("Person"),
						Value:    "Dan",
						DataType: "string",
					},
					Relationship: "might speak",
					Object: ConceptInstance{
						Type:     stringPtr("Language"),
						Value:    "English",
						DataType: "string",
					},
					Certainty: 100,
				},
				Rule: &Rule{
					Bindings: map[string]interface{}{
						"S":       "Dan",
						"O":       100.0,
						"COUNTRY": "England",
					},
					Conditions: []Conditioner{
						ConditionRelationship{
							Subject:      "Dan",
							Relationship: "lives in",
							Certainty:    100,
							Object:       "England",
							salience:     100,
							FactID:       factID,
							ObjectType:   "string",
							FactKey:      stringPtr("30d74cc5-7eb6-4c88-8e12-4897b47e3ee7"),
						},
						ConditionRelationship{
							Subject:      "England",
							Relationship: "has national language",
							Certainty:    100,
							Object:       "English",
							salience:     100,
							FactID:       factID,
							ObjectType:   "string",
							FactKey:      nil,
						},
						ConditionExpression{
							WasMet:   true,
							salience: 100,
							Expression: Expression{
								Text: "1 gte 0",
							},
						},
					},
				},
				Time: 123456789,
			},
			expectErr: nil,
		},
		{
			description:  "Returns correct evidence for datasource",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`{
				"factID": "WA:DF:1234",
				"source": "datasource",
				"fact": {
					"subject": {
						"type": "location",
						"value": "London",
						"dataType": "string"
					},
					"relationship": {
						"type": "has temperature"
					},
					"object": {
						"type": "temperature",
						"value": "290.00",
						"dataType": "string"
					},
					"certainty": 100
				},
				"time": 123456789
			}`),
			expectEvidence: &Evidence{
				Fact: Fact{
					ID:     "WA:DF:1234",
					Source: "datasource",
					Subject: ConceptInstance{
						Type:     stringPtr("location"),
						Value:    "London",
						DataType: "string",
					},
					Relationship: "has temperature",
					Object: ConceptInstance{
						Type:     stringPtr("temperature"),
						Value:    "290.00",
						DataType: "string",
					},
					Certainty: 100,
				},
				Rule: nil,
				Time: 123456789,
			},
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "/analysis/evidence/"+factID+"/"+sessionID, r.RequestURI)

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

			evidence, err := client.Evidence(sessionID, factID, nil)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectEvidence, evidence)
		})
	}

}

func TestInteractionLog(t *testing.T) {
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
							Object:       interface{}(nil),
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
							Object:       interface{}(nil),
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

			interactionLog, err := client.Interactions(sessionID, nil)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectInteractionLog, interactionLog)
		})
	}
}

func TestSessionKmVersion(t *testing.T) {
	sessionID := "1234"
	created, _ := time.Parse(time.RFC3339, "2022-04-20T13:13:31.000Z")

	testCases := []struct {
		description  string
		kmid         string
		engine       *string
		responseBody *string
		responseCode int
		expectKm     *KnowledgeMap
		expectErr    error
	}{
		{
			description:  "Bad request",
			responseCode: http.StatusBadRequest,
			responseBody: stringPtr("Foo bar baz"),
			expectKm:     nil,
			expectErr:    errors.New("API returned error 400: Foo bar baz"),
		},
		{
			description:  "Internal server error",
			responseCode: http.StatusInternalServerError,
			responseBody: stringPtr("Foo bar baz"),
			expectKm:     nil,
			expectErr:    errors.New("API returned error 500: Foo bar baz"),
		},
		{
			description:  "Returns version info",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`{"km": {
				"id": "5042ae3a-723a-45fa-bd6c-2d09e96e75f6",
				"name": "concept types",
				"versionID": "5042ae3a-723a-45fa-bd6c-2d09e96e75f6",
				"versionCreated": "2022-04-20T13:13:31.000Z",
				"versionStatus": "Draft"
			}}`),
			expectKm: &KnowledgeMap{
				ID:             "5042ae3a-723a-45fa-bd6c-2d09e96e75f6",
				Name:           "concept types",
				VersionID:      "5042ae3a-723a-45fa-bd6c-2d09e96e75f6",
				VersionCreated: &created,
				VersionStatus:  "Draft",
			},
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)

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

			km, err := client.KnowledgeMapVersion(sessionID)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectKm, km)
		})
	}

}

func TestSessionFacts(t *testing.T) {
	sessionID := "1234"

	testCases := []struct {
		description  string
		kmid         string
		engine       *string
		responseBody *string
		responseCode int
		expectFacts  *Facts
		expectErr    error
	}{
		{
			description:  "Bad request",
			responseCode: http.StatusBadRequest,
			responseBody: stringPtr("Foo bar baz"),
			expectFacts:  nil,
			expectErr:    errors.New("API returned error 400: Foo bar baz"),
		},
		{
			description:  "Internal server error",
			responseCode: http.StatusInternalServerError,
			responseBody: stringPtr("Foo bar baz"),
			expectFacts:  nil,
			expectErr:    errors.New("API returned error 500: Foo bar baz"),
		},
		{
			description:  "Returns facts info",
			responseCode: http.StatusOK,
			responseBody: stringPtr(`{
				"facts": {
					"global": [],
					"context": [],
					"local": [
						{
							"id": "WA:AF:029f78551644e583e9214c6fa1cb85ee83949f97f3068ec0bca26b750e3106e9",
							"subject": {
								"concept": "Subject",
								"value": "Dan",
								"dataType": "string"
							},
							"relationship": "has date plural",
							"object": {
								"concept": "Date plural",
								"value": 1655856000000,
								"dataType": "date"
							},
							"certainty": 100,
							"source": "answer"
						},
						{
							"id": "WA:AF:49a8af17bddfb9a5cafe73ee3332cb056d4dc9e5e9895e4c2e2180c586523fe1",
							"subject": {
								"concept": "Subject",
								"value": "Dan",
								"dataType": "string"
							},
							"relationship": "has number singular",
							"object": {
								"concept": "Number singular",
								"value": 8,
								"dataType": "number"
							},
							"certainty": 100,
							"source": "answer"
						},
						{
							"id": "WA:AF:2cef310b47a6dd3072a59c8752e507753deedbda073eba2fa5189b9058ef795a",
							"subject": {
								"concept": "Subject",
								"value": "Dan",
								"dataType": "string"
							},
							"relationship": "has string plural",
							"object": {
								"concept": "String plural",
								"value": "string plural 2",
								"dataType": "string"
							},
							"certainty": 100,
							"source": "answer"
						},
						{
							"id": "WA:AF:fcd65f85fc5f01e76575ae5b245e71f8fe258cfa1c6e2da3795620b08010a8ba",
							"subject": {
								"concept": "Subject",
								"value": "Dan",
								"dataType": "string"
							},
							"relationship": "has truth",
							"object": {
								"concept": "Truth",
								"value": true,
								"dataType": "boolean"
							},
							"certainty": 100,
							"source": "answer"
						},
						{
							"id": "WA:RF:e70b10add03a2745860a81b9625fe0d9fe62373ebcb599281a9fc3f6813318d3",
							"subject": {
								"concept": "Subject",
								"value": "Dan",
								"dataType": "string"
							},
							"relationship": "relationship",
							"object": {
								"concept": "Result object",
								"value": "string singular 1 string plural 2 8 5.15164 1656028800000 1655856000000 true",
								"dataType": "string"
							},
							"certainty": 100,
							"source": "rule"
						}
					]
				}
			}`),
			expectFacts: &Facts{
				Global:  []Fact{},
				Context: []Fact{},
				Local: []Fact{
					{
						ID: "WA:AF:029f78551644e583e9214c6fa1cb85ee83949f97f3068ec0bca26b750e3106e9",
						Subject: ConceptInstance{
							Concept:  stringPtr("Subject"),
							Value:    "Dan",
							DataType: "string",
						},
						Relationship: "has date plural",
						Object: ConceptInstance{
							Concept:  stringPtr("Date plural"),
							Value:    float64(1655856000000),
							DataType: "date",
						},
						Certainty: 100,
						Source:    "answer",
					},
					{
						ID: "WA:AF:49a8af17bddfb9a5cafe73ee3332cb056d4dc9e5e9895e4c2e2180c586523fe1",
						Subject: ConceptInstance{
							Concept:  stringPtr("Subject"),
							Value:    "Dan",
							DataType: "string",
						},
						Relationship: "has number singular",
						Object: ConceptInstance{
							Concept:  stringPtr("Number singular"),
							Value:    float64(8),
							DataType: "number",
						},
						Certainty: 100,
						Source:    "answer",
					},
					{
						ID: "WA:AF:2cef310b47a6dd3072a59c8752e507753deedbda073eba2fa5189b9058ef795a",
						Subject: ConceptInstance{
							Concept:  stringPtr("Subject"),
							Value:    "Dan",
							DataType: "string",
						},
						Relationship: "has string plural",
						Object: ConceptInstance{
							Concept:  stringPtr("String plural"),
							Value:    "string plural 2",
							DataType: "string",
						},
						Certainty: 100,
						Source:    "answer",
					},
					{
						ID: "WA:AF:fcd65f85fc5f01e76575ae5b245e71f8fe258cfa1c6e2da3795620b08010a8ba",
						Subject: ConceptInstance{
							Concept:  stringPtr("Subject"),
							Value:    "Dan",
							DataType: "string",
						},
						Relationship: "has truth",
						Object: ConceptInstance{
							Concept:  stringPtr("Truth"),
							Value:    true,
							DataType: "boolean",
						},
						Certainty: 100,
						Source:    "answer",
					},
					{
						ID: "WA:RF:e70b10add03a2745860a81b9625fe0d9fe62373ebcb599281a9fc3f6813318d3",
						Subject: ConceptInstance{
							Concept:  stringPtr("Subject"),
							Value:    "Dan",
							DataType: "string",
						},
						Relationship: "relationship",
						Object: ConceptInstance{
							Concept:  stringPtr("Result object"),
							Value:    "string singular 1 string plural 2 8 5.15164 1656028800000 1655856000000 true",
							DataType: "string",
						},
						Certainty: 100,
						Source:    "rule",
					},
				},
			},
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)

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

			facts, err := client.SessionFacts(sessionID)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectFacts, facts)
		})
	}

}

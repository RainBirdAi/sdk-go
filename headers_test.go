package sdk

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkedHeaders are the headers the SDK may set. Each call must send exactly
// the ones it expects, and none of the others.
var checkedHeaders = []string{
	"Authorization",
	"Accept",
	"Content-Type",
	"x-evidence-key",
	"x-interaction-key",
	"x-rainbird-engine",
}

func TestCallHeaders(t *testing.T) {
	const apiKey = "api-key"
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":"))

	testCases := []struct {
		description string
		call        func(c *Client) error

		expectHeaders map[string]string
	}{
		{
			description: "NewSession",
			call: func(c *Client) error {
				_, err := c.NewSession("km-id", NoContext, nil, nil)
				return err
			},
			expectHeaders: map[string]string{
				"Authorization": basicAuth,
				"Accept":        "application/json",
			},
		},
		{
			description: "Version",
			call: func(c *Client) error {
				_, err := c.Version()
				return err
			},
			expectHeaders: map[string]string{},
		},
		{
			description: "Evidence",
			call: func(c *Client) error {
				_, err := c.Evidence("session-id", "fact-id", nil)
				return err
			},
			expectHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			description: "Evidence with evidence key",
			call: func(c *Client) error {
				_, err := c.Evidence("session-id", "fact-id", stringPtr("evidence-key"))
				return err
			},
			expectHeaders: map[string]string{
				"Content-Type":   "application/json",
				"x-evidence-key": "evidence-key",
			},
		},
		{
			description: "Interactions",
			call: func(c *Client) error {
				_, err := c.Interactions("session-id", nil)
				return err
			},
			expectHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			description: "Interactions with interaction key",
			call: func(c *Client) error {
				_, err := c.Interactions("session-id", stringPtr("interaction-key"))
				return err
			},
			expectHeaders: map[string]string{
				"Content-Type":      "application/json",
				"x-interaction-key": "interaction-key",
			},
		},
		{
			description: "KnowledgeMapVersion",
			call: func(c *Client) error {
				_, err := c.KnowledgeMapVersion("session-id")
				return err
			},
			expectHeaders: map[string]string{
				"Authorization": basicAuth,
			},
		},
		{
			description: "SessionFacts",
			call: func(c *Client) error {
				_, err := c.SessionFacts("session-id")
				return err
			},
			expectHeaders: map[string]string{
				"Authorization": basicAuth,
			},
		},
		{
			description: "Session Inject",
			call: func(c *Client) error {
				s, _ := c.ResumeSession("session-id")
				return s.Inject([]InjectFact{{Subject: "John", Relationship: "lives in", Object: "England"}})
			},
			expectHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			description: "Session Query",
			call: func(c *Client) error {
				s, _ := c.ResumeSession("session-id")
				_, _, err := s.Query("John", "lives in", "")
				return err
			},
			expectHeaders: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
		},
		{
			description: "Session Query with engine header option",
			call: func(c *Client) error {
				s, _ := c.ResumeSession("session-id")
				_, _, err := s.Query("John", "lives in", "", AddHeaders(http.Header{
					"x-rainbird-engine": []string{"Experimental"},
				}))
				return err
			},
			expectHeaders: map[string]string{
				"Accept":            "application/json",
				"Content-Type":      "application/json",
				"x-rainbird-engine": "Experimental",
			},
		},
		{
			description: "Session Response",
			call: func(c *Client) error {
				s, _ := c.ResumeSession("session-id")
				_, _, err := s.Response([]QAnswer{{Subject: "John", Relationship: "lives in", Object: "England"}})
				return err
			},
			expectHeaders: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
		},
		{
			description: "Session Undo",
			call: func(c *Client) error {
				s, _ := c.ResumeSession("session-id")
				_, _, err := s.Undo()
				return err
			},
			expectHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			var calls int64

			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					atomic.AddInt64(&calls, 1)

					for _, h := range checkedHeaders {
						expect, shouldBePresent := tc.expectHeaders[h]
						_, present := r.Header[http.CanonicalHeaderKey(h)]
						if assert.Equal(t, shouldBePresent, present, "header %s presence", h) {
							assert.Equal(t, expect, r.Header.Get(h), "header %s value", h)
						}
					}

					switch {
					case strings.HasPrefix(r.URL.Path, "/start/"):
						w.Write([]byte(`{"id":"session-id"}`))
					case strings.HasPrefix(r.URL.Path, "/analysis/interactions/"):
						w.Write([]byte(`[]`))
					default:
						w.Write([]byte(`{}`))
					}
				},
			))
			defer srv.Close()

			c := &Client{
				APIKey:         apiKey,
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
			}

			require.NoError(t, tc.call(c))
			assert.Equal(t, int64(1), atomic.LoadInt64(&calls))
		})
	}
}

func TestQueryEngineHeaderOption(t *testing.T) {
	testCases := []struct {
		description string
		options     []QueryOption

		expectEngine []string
		expectOther  string
	}{
		{
			description: "Engine header set",
			options: []QueryOption{
				AddHeaders(http.Header{"x-rainbird-engine": []string{"Experimental"}}),
			},
			expectEngine: []string{"Experimental"},
		},
		{
			description: "Engine header set alongside another option",
			options: []QueryOption{
				AddHeaders(http.Header{"x-rainbird-engine": []string{"Experimental"}}),
				AddHeaders(http.Header{"x-other": []string{"other"}}),
			},
			expectEngine: []string{"Experimental"},
			expectOther:  "other",
		},
		{
			description:  "Engine header not set without option",
			options:      nil,
			expectEngine: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			var calls int64

			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					atomic.AddInt64(&calls, 1)
					assert.Equal(t, "/session-id/query", r.URL.Path)
					assert.Equal(t, tc.expectEngine, r.Header.Values("x-rainbird-engine"))
					assert.Equal(t, tc.expectOther, r.Header.Get("x-other"))
					w.Write([]byte(`{}`))
				},
			))
			defer srv.Close()

			c := &Client{EnvironmentURL: srv.URL, HTTPClient: srv.Client()}
			s, _ := c.ResumeSession("session-id")

			_, _, err := s.Query("John", "lives in", "", tc.options...)
			require.NoError(t, err)
			assert.Equal(t, int64(1), atomic.LoadInt64(&calls))
		})
	}
}

func TestQueryEngineHeaderNotCarriedOver(t *testing.T) {
	t.Parallel()

	var engines []string

	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			engines = append(engines, r.URL.Path+"="+r.Header.Get("x-rainbird-engine"))
			w.Write([]byte(`{}`))
		},
	))
	defer srv.Close()

	c := &Client{EnvironmentURL: srv.URL, HTTPClient: srv.Client()}
	s, _ := c.ResumeSession("session-id")

	_, _, err := s.Query("John", "lives in", "", AddHeaders(http.Header{
		"x-rainbird-engine": []string{"Experimental"},
	}))
	require.NoError(t, err)
	_, _, err = s.Query("John", "lives in", "")
	require.NoError(t, err)
	_, _, err = s.Response([]QAnswer{{Subject: "John", Relationship: "lives in", Object: "England"}})
	require.NoError(t, err)
	_, _, err = s.Undo()
	require.NoError(t, err)

	assert.Equal(t, []string{
		"/session-id/query=Experimental",
		"/session-id/query=",
		"/session-id/response=",
		"/session-id/undo=",
	}, engines)
}

func TestClientEngineHeader(t *testing.T) {
	calls := []struct {
		description string
		call        func(c *Client) error
	}{
		{"NewSession", func(c *Client) error {
			_, err := c.NewSession("km-id", NoContext, nil, nil)
			return err
		}},
		{"Version", func(c *Client) error {
			_, err := c.Version()
			return err
		}},
		{"Evidence", func(c *Client) error {
			_, err := c.Evidence("session-id", "fact-id", nil)
			return err
		}},
		{"Interactions", func(c *Client) error {
			_, err := c.Interactions("session-id", nil)
			return err
		}},
		{"KnowledgeMapVersion", func(c *Client) error {
			_, err := c.KnowledgeMapVersion("session-id")
			return err
		}},
		{"SessionFacts", func(c *Client) error {
			_, err := c.SessionFacts("session-id")
			return err
		}},
		{"Session Inject", func(c *Client) error {
			s, _ := c.ResumeSession("session-id")
			return s.Inject([]InjectFact{{Subject: "John", Relationship: "lives in", Object: "England"}})
		}},
		{"Session Query", func(c *Client) error {
			s, _ := c.ResumeSession("session-id")
			_, _, err := s.Query("John", "lives in", "")
			return err
		}},
		{"Session Response", func(c *Client) error {
			s, _ := c.ResumeSession("session-id")
			_, _, err := s.Response([]QAnswer{{Subject: "John", Relationship: "lives in", Object: "England"}})
			return err
		}},
		{"Session Undo", func(c *Client) error {
			s, _ := c.ResumeSession("session-id")
			_, _, err := s.Undo()
			return err
		}},
	}

	engines := []struct {
		description  string
		engine       string
		expectEngine []string
	}{
		{"Engine set to v2.0", "v2.0", []string{"v2.0"}},
		{"Engine set to rbx", "rbx", []string{"rbx"}},
		{"Engine not set", "", nil},
	}

	for _, e := range engines {
		e := e // Capture
		for _, tc := range calls {
			tc := tc // Capture
			t.Run(e.description+"/"+tc.description, func(t *testing.T) {
				t.Parallel()
				var requests int64

				srv := httptest.NewServer(http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						atomic.AddInt64(&requests, 1)
						assert.Equal(t, e.expectEngine, r.Header.Values("x-rainbird-engine"))

						switch {
						case strings.HasPrefix(r.URL.Path, "/start/"):
							w.Write([]byte(`{"id":"session-id"}`))
						case strings.HasPrefix(r.URL.Path, "/analysis/interactions/"):
							w.Write([]byte(`[]`))
						default:
							w.Write([]byte(`{}`))
						}
					},
				))
				defer srv.Close()

				c := &Client{
					APIKey:         "api-key",
					EnvironmentURL: srv.URL,
					HTTPClient:     srv.Client(),
					Engine:         e.engine,
				}

				require.NoError(t, tc.call(c))
				assert.Equal(t, int64(1), atomic.LoadInt64(&requests))
			})
		}
	}
}

// The session is pinned to one engine, and engine-router picks the engine per
// request, so a query sent to a different engine wouldn't find the session.
// The session's engine therefore wins over a per-request header option.
func TestSessionEngineOverridesQueryOption(t *testing.T) {
	testCases := []struct {
		description   string
		sessionEngine string
		optionEngine  string
	}{
		{"Session v2.0, option rbx", "v2.0", "rbx"},
		{"Session rbx, option v2.0", "rbx", "v2.0"},
		{"Session and option both rbx", "rbx", "rbx"},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			var requests int64

			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					atomic.AddInt64(&requests, 1)
					assert.Equal(t, []string{tc.sessionEngine}, r.Header.Values("x-rainbird-engine"))
					w.Write([]byte(`{}`))
				},
			))
			defer srv.Close()

			c := &Client{EnvironmentURL: srv.URL, HTTPClient: srv.Client(), Engine: tc.sessionEngine}
			s, _ := c.ResumeSession("session-id")

			_, _, err := s.Query("John", "lives in", "", AddHeaders(http.Header{
				"x-rainbird-engine": []string{tc.optionEngine},
			}))
			require.NoError(t, err)
			assert.Equal(t, int64(1), atomic.LoadInt64(&requests))
		})
	}
}

func TestClientEngineSession(t *testing.T) {
	testCases := []struct {
		description string
		engine      string
		expect      []string
	}{
		{"Engine set to rbx", "rbx", []string{"rbx"}},
		{"Engine set to v2.0", "v2.0", []string{"v2.0"}},
		{"Engine not set", "", nil},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			sent := map[string][]string{}

			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					sent[r.URL.Path] = r.Header.Values("x-rainbird-engine")

					if strings.HasPrefix(r.URL.Path, "/start/") {
						w.Write([]byte(`{"id":"session-id"}`))
						return
					}
					w.Write([]byte(`{}`))
				},
			))
			defer srv.Close()

			c := &Client{
				APIKey:         "api-key",
				EnvironmentURL: srv.URL,
				HTTPClient:     srv.Client(),
				Engine:         tc.engine,
			}

			s, err := c.NewSession("km-id", NoContext, nil, nil)
			require.NoError(t, err)
			_, _, err = s.Query("John", "lives in", "")
			require.NoError(t, err)
			_, _, err = s.Response([]QAnswer{{Subject: "John", Relationship: "lives in", Object: "England"}})
			require.NoError(t, err)
			_, _, err = s.Undo()
			require.NoError(t, err)

			assert.Equal(t, map[string][]string{
				"/start/km-id":         tc.expect,
				"/session-id/query":    tc.expect,
				"/session-id/response": tc.expect,
				"/session-id/undo":     tc.expect,
			}, sent)
		})
	}
}

func TestSessionEngineNotChangedByClient(t *testing.T) {
	t.Parallel()

	sent := map[string][]string{}

	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			sent[r.URL.Path] = r.Header.Values("x-rainbird-engine")

			if strings.HasPrefix(r.URL.Path, "/start/") {
				w.Write([]byte(`{"id":"session-id"}`))
				return
			}
			w.Write([]byte(`{}`))
		},
	))
	defer srv.Close()

	c := &Client{
		APIKey:         "api-key",
		EnvironmentURL: srv.URL,
		HTTPClient:     srv.Client(),
		Engine:         "rbx",
	}

	s, err := c.NewSession("km-id", NoContext, nil, nil)
	require.NoError(t, err)

	c.Engine = "v2.0"

	_, _, err = s.Query("John", "lives in", "")
	require.NoError(t, err)

	assert.Equal(t, map[string][]string{
		"/start/km-id":      {"rbx"},
		"/session-id/query": {"rbx"},
	}, sent)
}

package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Client is a grouped set of config for interacting with a Rainbird engine
type Client struct {
	// APIKey is the user's authentication for the session. Required.
	APIKey string
	// Engine allows interaction to be with explicit engines, especially
	// experimental ones. Advanced users only. Optional.
	Engine string
	// EnvironmentURL is the URL of the API to use. This is most commonly
	// EnvCommunity. Required.
	EnvironmentURL string

	// HTTPClient allows the user to provide a client other than
	// http.DefaultClient with which to make network calls. Good for testing,
	// intercepting, or other advanced uses. Optional.
	HTTPClient *http.Client
}

var (
	// ErrClientMissingAPIKey is returned when an operation is attempted with a
	// client for which it requires an API key but has none
	ErrClientMissingAPIKey = errors.New("Client Missing API Key")
	// ErrClientMissingEnvironmentURL is returned when an operation is attempted
	// with a client for which it must contact an API, but no API URL has been
	// configured
	ErrClientMissingEnvironmentURL = errors.New("Client Missing Environment URL")
	// ErrNewSessionInvalidKMID is returned when a new session is being started,
	// but the KMID is invalid (for example, due to being empty like "")
	ErrNewSessionInvalidKMID = errors.New("NewSessionInvalidKMID")
)

// NoContext is the default way to interact with the engine; without defining a
// context in which to work
const NoContext string = ""

// HTTP retrieves an appropriate client for making HTTP calls
func (c *Client) HTTP() *http.Client {
	if c.HTTPClient == nil {
		return http.DefaultClient
	}
	return c.HTTPClient
}

// NewSession creates a new engine interaction session
func (c *Client) NewSession(kmID string, contextID string) (*Session, error) {
	if c.APIKey == "" {
		return nil, ErrClientMissingAPIKey
	}
	if c.EnvironmentURL == "" {
		return nil, ErrClientMissingEnvironmentURL
	}
	if kmID == "" {
		return nil, ErrNewSessionInvalidKMID
	}

	url := c.EnvironmentURL + "/start/" + kmID
	if contextID != NoContext {
		url += "?contextid=" + contextID
	}

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.APIKey, "")
	req.Header.Set("Accept", "application/json")
	if c.Engine != "" {
		req.Header.Set("x-rainbird-engine", c.Engine)
	}

	resp, err := c.HTTP().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, errors.New("Empty response body")
	}

	// TODO: Stream this rather than ReadAll
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TODO: The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API returned an error: %s", string(rawBody))
	}

	var body struct {
		Error string
		ID    string `json:"id"`
	}
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, err
	}
	if body.ID == "" {
		return nil, errors.New("API returned no error but no ID either")
	}

	return &Session{
		ID:     body.ID,
		client: c,
	}, nil
}

// ResumeSession resumes a session from the given ID
func (c *Client) ResumeSession(sessionID string) (*Session, error) {
	return &Session{ID: sessionID, client: c}, nil
}

// Version contacts the engine API and retrieves its version information
func (c *Client) Version() (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.EnvironmentURL+"/version",
		nil,
	)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTP().Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		// TODO: This should mean we have an error message but not the case
		return "", fmt.Errorf("API returned an error code")
	}

	apiVersion, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(apiVersion), err
}

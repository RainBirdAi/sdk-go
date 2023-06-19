package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
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

// KnowledgeMap represents the knowledge map version information returned from the session endpoint
type KnowledgeMap struct {
	ID             string     `json:"id,omitempty"`
	Name           string     `json:"name,omitempty"`
	VersionID      string     `json:"versionID,omitempty"`
	VersionNumber  *int       `json:"versionNumber,omitempty"`
	VersionCreated *time.Time `json:"versionCreated,omitempty"`
	VersionStatus  string     `json:"versionStatus,omitempty"`
}

// String makes *Evidence satisfy fmt.Stringer
func (s *KnowledgeMap) String() string {
	str := fmt.Sprintf(
		"%s (%s %s), %s",
		s.ID,
		s.Name,
		s.VersionID,
		s.VersionStatus,
	)

	if s.VersionCreated == nil {
		return str
	}

	return str + " " + s.VersionCreated.Format(time.RFC3339)
}

var _ fmt.Stringer = (*KnowledgeMap)(nil)

// Facts contains the three types of Fact a session can have
type Facts struct {
	Global  []Fact `json:"global,omitempty"`
	Context []Fact `json:"context,omitempty"`
	Local   []Fact `json:"local,omitempty"`
}

// Flatten helper function that returns all facts in one slice
func (f *Facts) Flatten() []Fact {
	var allFacts []Fact
	allFacts = append(allFacts, f.Global...)
	allFacts = append(allFacts, f.Context...)
	allFacts = append(allFacts, f.Local...)
	return allFacts
}

// String makes *Facts satisfy fmt.Stringer
func (f *Facts) String() string {
	var facts []string

	for _, f := range f.Flatten() {
		facts = append(facts, f.String())
	}

	return strings.Join(facts, "\n")
}

var _ fmt.Stringer = (*Facts)(nil)

// HTTP retrieves an appropriate client for making HTTP calls
func (c *Client) HTTP() *http.Client {
	if c.HTTPClient == nil {
		return http.DefaultClient
	}
	return c.HTTPClient
}

// NewSession creates a new engine interaction session
func (c *Client) NewSession(kmID string, contextID string, useDraft *bool, version *int) (*Session, error) {
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
	if useDraft != nil && *useDraft {
		url += "?useDraft=true"
	}
	if version != nil {
		versionStr := strconv.Itoa(*version)
		url += "?version=" + versionStr
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
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		rawBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(
			"API returned an error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var body struct {
		Error string
		ID    string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
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
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf(
			"API returned error code %d: %s",
			resp.StatusCode,
			body,
		)
	}

	return string(body), err
}

// Evidence returns Evidence for a given factID
func (c *Client) Evidence(sessionID string, factID string, evidenceKey *string) (*Evidence, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.EnvironmentURL+"/analysis/evidence/"+factID+"/"+sessionID,
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if evidenceKey != nil {
		req.Header.Set("x-evidence-key", *evidenceKey)
	}

	resp, err := c.HTTP().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		rawBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var response evidenceResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	evidence, err := asEvidence(response)
	if err != nil {
		return nil, err
	}

	return &evidence, nil
}

// Interactions returns an array of time-stamped session events
func (c *Client) Interactions(sessionID string, interactionKey *string) ([]InteractionEvent, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.EnvironmentURL+"/analysis/interactions/"+sessionID,
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if interactionKey != nil {
		req.Header.Set("x-interaction-key", *interactionKey)
	}

	resp, err := c.HTTP().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		rawBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var response []InteractionResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return InteractionEvents(response)
}

// KnowledgeMapVersion returns information about the knowledge map for a session
func (c *Client) KnowledgeMapVersion(sessionID string) (*KnowledgeMap, error) {
	resp, err := c.session(sessionID, "version")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessionKm struct {
		KnowledgeMap `json:"km,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&sessionKm)
	if err != nil {
		return nil, err
	}

	return &sessionKm.KnowledgeMap, nil
}

// SessionFacts returns facts from the given session
func (c *Client) SessionFacts(sessionID string) (*Facts, error) {
	resp, err := c.session(sessionID, "facts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessionFacts struct {
		Facts `json:"facts,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&sessionFacts)
	if err != nil {
		return nil, err
	}

	return &sessionFacts.Facts, nil
}

func (c *Client) session(sessionID, filterStr string) (*http.Response, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.EnvironmentURL+"/analysis/session/"+sessionID+"?filter="+filterStr,
		nil,
	)
	req.SetBasicAuth(c.APIKey, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTP().Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		rawBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	return resp, nil
}

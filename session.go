package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Session is a started engine interaction with a knowledge map
type Session struct {
	ID string

	client *Client
}

// InjectFact is the structure of facts to add to a session via the Inject call
type InjectFact struct {
	Subject      string `json:"subject"`
	Relationship string `json:"relationship"`
	Object       string `json:"object"`
	Certainty    string `json:"cf"`
}

// QAnswer is a user's answer to an Question from the engine
type QAnswer struct {
	Subject      string `json:"subject"`
	Relationship string `json:"relationship"`
	Object       string `json:"object"`
	CF           string `json:"cf"`
}

var (
	// ErrQueryBlankRelationship is given when relationship is "" in a query
	// A relationship must always be provided.
	ErrQueryBlankRelationship = errors.New("Blank Relationship")
)

// Inject adds facts to a running session
func (s *Session) Inject(facts []InjectFact) error {
	// TODO: Inject takes invalid JSON and doesn't match the API docs!
	payload, err := json.Marshal(&facts)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		s.client.EnvironmentURL+"/"+s.ID+"/inject",
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.client.Engine != "" {
		req.Header.Set("x-rainbird-engine", s.client.Engine)
	}

	resp, err := s.client.HTTP().Do(req)
	if err != nil {
		return err
	}
	// The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned error %d", resp.StatusCode)
	}
	return nil
}

// Query is the first interaction with a session, setting a goal. Leaving sub,
// obj and/or both blank ("") will instruct the engine in what you wish to find
// out. For example, s.Query("John", "speaks", "") will instruct the engine
// that you wish to find out which languages John speaks.
func (s *Session) Query(sub, rel, obj string) (*Question, *[]Answer, error) {
	if rel == "" {
		return nil, nil, ErrQueryBlankRelationship
	}

	payloadS := struct {
		Subject      string `json:"subject,omitempty"`
		Relationship string `json:"relationship"`
		Object       string `json:"object,omitempty"`
	}{
		Subject:      sub,
		Relationship: rel,
		Object:       obj,
	}
	payload, err := json.Marshal(&payloadS)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		s.client.EnvironmentURL+"/"+s.ID+"/query",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if s.client.Engine != "" {
		req.Header.Set("x-rainbird-engine", s.client.Engine)
	}

	resp, err := s.client.HTTP().Do(req)
	if err != nil {
		return nil, nil, err
	}
	if resp.Body == nil {
		return nil, nil, errors.New("Empty response body")
	}

	// TODO: Stream this rather than ReadAll
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var body struct {
		Error    string
		Question *Question
		Result   *[]Answer
	}
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, nil, err
	}

	return body.Question, body.Result, nil
}

// Response submits a user response to the engine, and must be a response to
// a Question the engine has asked.
func (s *Session) Response(answers []QAnswer) (*Question, *[]Answer, error) {
	payloadS := struct {
		Answers []QAnswer `json:"answers"`
	}{
		Answers: answers,
	}
	payload, err := json.Marshal(&payloadS)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		s.client.EnvironmentURL+"/"+s.ID+"/response",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if s.client.Engine != "" {
		req.Header.Set("x-rainbird-engine", s.client.Engine)
	}

	resp, err := s.client.HTTP().Do(req)
	if err != nil {
		return nil, nil, err
	}
	if resp.Body == nil {
		return nil, nil, errors.New("Empty response body")
	}

	// TODO: Stream this rather than ReadAll
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var body struct {
		Error    string
		Question *Question
		Result   *[]Answer
	}
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, nil, err
	}

	return body.Question, body.Result, nil
}

// Undo steps the engine back in the case of a mistake, for example if a
// Response has been given in error.
func (s *Session) Undo() (*Question, *[]Answer, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		s.client.EnvironmentURL+"/"+s.ID+"/undo",
		strings.NewReader("{}"),
	)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.client.Engine != "" {
		req.Header.Set("x-rainbird-engine", s.client.Engine)
	}

	resp, err := s.client.HTTP().Do(req)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Stream this rather than ReadAll
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var body struct {
		Error    string
		Question *Question
		Result   *[]Answer
	}
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, nil, err
	}

	return body.Question, body.Result, nil
}

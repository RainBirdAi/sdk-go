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

type Session struct {
	ID string

	client *Client
}

type InjectFact struct {
	Subject      string `json:"subject"`
	Relationship string `json:"relationship"`
	Object       string `json:"object"`
	Certainty    string `json:"cf"`
}

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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		// TODO: This should mean we have an error message but not the case
		return fmt.Errorf("API returned an error code")
	}
	return nil
}

func (s *Session) Query(sub, rel, obj string) (*Question, *[]Answer, error) {
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

	resp, err := http.DefaultClient.Do(req)
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

	// TODO: The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("API returned an error: %s", string(rawBody))
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

func (s *Session) Response(
	sub, rel, obj, certainty string,
) (*Question, *[]Answer, error) {
	type rAnswer struct {
		Subject      string `json:"subject"`
		Relationship string `json:"relationship"`
		Object       string `json:"object"`
		CF           string `json:"cf"`
	}

	payloadS := struct {
		Answers []rAnswer `json:"answers"`
	}{
		Answers: []rAnswer{
			{
				Subject:      sub,
				Relationship: rel,
				Object:       obj,
				CF:           certainty,
			},
		},
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

	resp, err := http.DefaultClient.Do(req)
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

	// TODO: The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("API returned an error: %s", string(rawBody))
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Stream this rather than ReadAll
	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// TODO: The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		return nil, nil, fmt.Errorf("API returned an error: %s", string(rawBody))
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

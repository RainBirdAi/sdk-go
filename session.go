package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Session struct {
	id string

	Client *Client
}

func (s *Session) Query(sub, rel, obj string) (*Question, *Answer, error) {
	payloadS := struct {
		Session_id   string `json:"session_id"`
		Subject      string `json:"subject,omitempty"`
		Relationship string `json:"relationship"`
		Object       string `json:"object,omitempty"`
	}{
		Session_id:   s.id,
		Subject:      sub,
		Relationship: rel,
		Object:       obj,
	}
	payload, err := json.Marshal(&payloadS) // TODO: stream
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(string(payload))

	req, err := http.NewRequest(
		http.MethodPost,
		s.Client.EnvironmentURL+"/"+s.id+"/query",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

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

	var body struct {
		Error    string
		Question Question
		// TODO: Answer
	}
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, nil, err
	}

	// TODO: This is written to match API documentation, but the actual returned
	// data doesn't match that
	if resp.StatusCode >= 400 {
		fmt.Println(string(rawBody))
		return nil, nil, fmt.Errorf("API returned an error: %s", body.Error)
	}

	// TODO: Answers
	return &body.Question, nil, nil
}

func (s *Session) Inject(facts []Fact) (*Response, error) {
	return nil, errors.New("Not implemented")
}

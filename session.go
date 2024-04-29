package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Session is a started engine interaction with a knowledge map
type Session struct {
	ID string

	client *Client
}

// InjectFact is the structure of facts to add to a session via the Inject call
type InjectFact struct {
	Subject      string      `json:"subject"`
	Relationship string      `json:"relationship"`
	Object       interface{} `json:"object"`
	Certainty    string      `json:"cf"`
	CertFactor   *int        `json:"certainty,omitempty"`
}

// CertaintyFactor returns the certainty factor from QAnswer or InjectFact
func CertaintyFactor(cf string, cfPointer *int) int {
	if cfPointer != nil {
		return *cfPointer
	}
	cfValue, _ := strconv.Atoi(cf)
	return cfValue
}

func (i *InjectFact) CertaintyFactor() int {
	return CertaintyFactor(i.Certainty, i.CertFactor)
}

// String makes *InjectFact satisfy fmt.Stringer
func (i *InjectFact) String() string {
	return fmt.Sprintf(
		"%s - %s - %s",
		i.Subject,
		i.Relationship,
		i.Object,
	)
}

// Satisfy interfaces
var _ fmt.Stringer = (*InjectFact)(nil)

// QAnswer is a user's answer to an Question from the engine
type QAnswer struct {
	Subject      string      `json:"subject"`
	Relationship string      `json:"relationship"`
	Object       interface{} `json:"object"`
	CF           string      `json:"cf"`
	Certainty    *int        `json:"certainty,omitempty"`
	Answer       string      `json:"answer,omitempty"`
}

func (q *QAnswer) CertaintyFactor() int {
	return CertaintyFactor(q.CF, q.Certainty)
}

// String makes *QAnswer satisfy fmt.Stringer
func (q *QAnswer) String() string {
	return fmt.Sprintf(
		"%s - %s - %v [%3d%%]",
		q.Subject,
		q.Relationship,
		q.Object,
		q.CertaintyFactor(),
	)
}

// Satisfy interfaces
var _ fmt.Stringer = (*QAnswer)(nil)

var (
	// ErrQueryBlankRelationship is given when relationship is "" in a query
	// A relationship must always be provided.
	ErrQueryBlankRelationship = errors.New("blank relationship")
)

// Inject adds facts to a running session
func (s *Session) Inject(facts []InjectFact) error {
	// NOTE: Inject takes invalid JSON and doesn't match the API docs!
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

	resp, err := s.client.HTTP().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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
func (s *Session) Query(sub, rel string, obj interface{}) (*Question, []Answer, error) {
	if rel == "" {
		return nil, nil, ErrQueryBlankRelationship
	}

	payloadS := struct {
		Subject      string      `json:"subject,omitempty"`
		Relationship string      `json:"relationship"`
		Object       interface{} `json:"object,omitempty"`
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

	resp, err := s.client.HTTP().Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		rawBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}

		return nil, nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var body struct {
		Error    string
		Question *Question
		Result   []Answer
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, nil, err
	}

	return body.Question, body.Result, nil
}

// Response submits a user response to the engine, and must be a response to
// a Question the engine has asked.
func (s *Session) Response(answers []QAnswer) ([]Question, []Answer, error) {
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

	resp, err := s.client.HTTP().Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		rawBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}

		return nil, nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var body struct {
		Error          string     `json:"error"`
		Question       *Question  `json:"question"`
		ExtraQuestions []Question `json:"extraQuestions"`
		Result         []Answer   `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, nil, err
	}

	// Ensure the first question is placed at the beginning of the questions slice
	questions := []Question{}
	if body.Question != nil {
		questions = append(questions, *body.Question)
	}
	questions = append(questions, body.ExtraQuestions...)
	return questions, body.Result, nil
}

// Undo steps the engine back in the case of a mistake, for example if a
// Response has been given in error.
func (s *Session) Undo() ([]Question, []Answer, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		s.client.EnvironmentURL+"/"+s.ID+"/undo",
		strings.NewReader("{}"),
	)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.HTTP().Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// The API doesn't match documentation. Workaround.
	if resp.StatusCode >= 400 {
		rawBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}

		return nil, nil, fmt.Errorf(
			"API returned error %d: %s",
			resp.StatusCode,
			string(rawBody),
		)
	}

	var body struct {
		Error          string     `json:"error"`
		Question       *Question  `json:"question"`
		ExtraQuestions []Question `json:"extraQuestions"`
		Result         []Answer   `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, nil, err
	}

	// Ensure the first question is placed at the beginning of the questions slice
	questions := []Question{}
	if body.Question != nil {
		questions = append(questions, *body.Question)
	}
	questions = append(questions, body.ExtraQuestions...)
	return questions, body.Result, nil
}

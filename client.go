package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	APIKey         string
	Engine         string
	EnvironmentURL string
}

func (c *Client) NewSession(kmID string) (*Session, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.EnvironmentURL+"/start/"+kmID,
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

	resp, err := http.DefaultClient.Do(req)
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
		Id    string
	}
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, err
	}

	if body.Id == "" {
		return nil, errors.New("API returned no error but no ID either!")
	}

	return &Session{
		ID:     body.Id,
		client: c,
	}, nil
}

func (c *Client) ResumeSession(sessionID string) (*Session, error) {
	return &Session{ID: sessionID, client: c}, nil
}

func (c *Client) Version() (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.EnvironmentURL+"/version",
		nil,
	)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
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

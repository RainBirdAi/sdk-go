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

	var body struct {
		Error string
		Id    string
	}
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, err
	}

	// TODO: This is written to match API documentation, but the actual returned
	// data doesn't match that
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API returned an error: %s", body.Error)
	}
	if body.Id == "" {
		return nil, errors.New("API returned no error but no ID either!")
	}

	return &Session{
		id: body.Id,

		Client: c,
	}, nil
}

package sdk

import (
	"errors"
)

type Session struct {
	Client *Client
}

func (s *Session) Query(fact Fact) (*Response, error) {
	return nil, errors.New("Not implemented")
}

func (s *Session) Inject(facts []Fact) (*Response, error) {
	return nil, errors.New("Not implemented")
}

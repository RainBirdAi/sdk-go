package sdk

type Session struct {
	Client *Client
}

func (s *Session) Query(fact Fact) (*Response, error) {
	return nil, nil
}

func (s *Session) Inject(facts []Fact) (*Response, error) {
	return nil, nil
}

package sdk

type Session struct {
	Client *Client
}

func (s *Session) Query(sub, rel, obj string) (*Response, error) {
	return nil, nil
}

package sdk

type Client struct {
	APIKey         string
	EnvironmentURL string
}

func (c *Client) NewSession(kmID string) (*Session, error) {
	return &Session{
		Client: c,
	}, nil
}

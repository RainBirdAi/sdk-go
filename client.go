package sdk

type Client struct {
	APIKey         string
	EnvironmentURL string
}

func (c *Client) NewSession(kmID string) *Session {
	return &Session{Client: c}
}

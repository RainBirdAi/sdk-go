# Rainbird SDK (Go)

For more information on the company, our unique reasoning engine, and the benefits of understandable, explainable AI: https://rainbird.ai

## Usage

Create a client using your API key and pointing to the desired environment (commonly Community or Enterprise):

```go
client := sdk.Client{
	APIKey:         os.Getenv("RB_API_KEY"),
	EnvironmentURL: sdk.EnvCommunity,
}
```

Use that client to create a session with a knowledge map by its ID:

```go
session, err := client.NewSession(kmID)
```

Query that session, and Rainbird will either ask you a question or give you answers:

```go
question, answers, err := session.Query("John", "speaks", "")
```

## Known issues

Rainbird's API documentation currently describes a response format for errors that the API doesn't adhere to. This SDK works around those limitations where possible.

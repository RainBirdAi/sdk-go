# Rainbird SDK (Go)

Beautiful, automatically generated code documentation is available at https://pkg.go.dev/gitlab.com/rainbird-ai/sdk-go

For more information on the company, our unique reasoning engine, and the benefits of understandable, explainable AI please visit our corporate site at https://rainbird.ai

## Usage

Create a client using your API key and pointing to the desired environment (commonly Community or Enterprise):

```go
client := sdk.Client{
	APIKey:         os.Getenv("RB_API_KEY"),
	EnvironmentURL: sdk.EnvCommunity,
}
```

Use that client to create a session with a knowledge map by its ID. By passing a non empty contextID, the session will be connected to the context scope with that ID:

```go
session, err := client.NewSession(kmID, contextID)
```

Query that session, and Rainbird will either ask you a question or give you answers:

```go
question, answers, err := session.Query("John", "speaks", "")
```

## Known issues

Rainbird's API documentation currently describes a response format for errors that the API doesn't adhere to. This SDK works around those limitations where possible.

The engine and supporting documentation currently used very mixed terminology. For reference, this SDK will be endeavouring to move terminology to be more in line with current accepted terminology.
* The user makes a Query
* The engine asks a Question
* The user provides an Answer
* The engine gives a Result

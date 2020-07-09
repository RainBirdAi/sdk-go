Golang SDK for Rainbird's reasoning engine

For more information: https://rainbird.ai

Sample usage:

```go
client := sdk.Client{
	APIKey:         os.Getenv("RB_API_KEY"),
	EnvironmentURL: sdk.EnvCommunity,
}

session, err := client.NewSession(kmID)
if err != nil {
	...
}

question, answers, err := session.Query("John", "speaks", "")
...
```

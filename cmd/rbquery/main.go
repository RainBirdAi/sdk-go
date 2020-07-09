package main

import (
	"fmt"
	"os"

	sdk "gitlab.com/JohnAnthony/rainbird-sdk"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Printf("%s <kmID> <subject> <relationship> <object>\n", os.Args[0])
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("RB_API_KEY - required to authenticate")
}

func main() {
	// TODO: Remove panics, handle more gracefully
	if len(os.Args) != 5 {
		usage()
		os.Exit(1)
	}
	kmID := os.Args[1]
	sub := os.Args[2]
	rel := os.Args[3]
	obj := os.Args[4]
	if sub == "?" {
		sub = ""
	}
	if obj == "?" {
		obj = ""
	}

	apiKey := os.Getenv("RB_API_KEY")
	if apiKey == "" {
		panic("Missing environment variable RB_API_KEY")
	}

	client := sdk.Client{
		APIKey:         apiKey,
		EnvironmentURL: sdk.EnvCommunity,
	}

	session, err := client.NewSession(kmID)
	if err != nil {
		panic(err)
	}
	question, answer, err := session.Query(sub, rel, obj)
	if err != nil {
		panic(err)
	}
	fmt.Println(question)
	fmt.Println(answer)
}

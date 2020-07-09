package main

import (
	"errors"
	"fmt"
	"os"

	sdk "gitlab.com/JohnAnthony/rainbird-sdk"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Printf("  %s help\n", os.Args[0])
	fmt.Println("    - Give this help text")
	fmt.Printf("  %s start <kmid>\n", os.Args[0])
	fmt.Println("    - Start a new session for knowledge map <kmid>, receive a session ID for querying")
	fmt.Printf("  %s query <session> <subject> <relationship> <object>\n", os.Args[0])
	fmt.Println("    - Make a query against <session>, receive question or answers")
}

func main() {
	if len(os.Args) == 1 {
		usage()
		os.Exit(0)
	}

	var err error

	switch os.Args[1] {
	case "help":
		usage()
		os.Exit(0)
	case "query":
		if len(os.Args) != 6 {
			usage()
			os.Exit(0)
		}
		err = cmdQuery(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	case "start":
		if len(os.Args) != 3 {
			usage()
			os.Exit(0)
		}
		err = cmdStart(os.Args[2])
	default:
		fmt.Printf("ERR: Unknown operation '%s'\n", os.Args[1])
		fmt.Printf("::\n\n")
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("ERR: %s\n", err)
	}
}

func cmdQuery(sessionId, sub, rel, obj string) error {
	client := sdk.Client{
		EnvironmentURL: sdk.EnvCommunity,
	}

	session, err := client.ResumeSession(sessionId)
	if err != nil {
		return err
	}

	if sub == "?" {
		sub = ""
	}
	if obj == "?" {
		obj = ""
	}

	question, answers, err := session.Query(sub, rel, obj)
	if err != nil {
		panic(err)
	} else if question != nil {
		fmt.Printf("QUESTION: %s\n", question)
	} else if answers != nil {
		fmt.Println("ANSWERS:")
		for _, a := range *answers {
			fmt.Printf("  %s", a)
		}
		fmt.Println("")
	}
	return nil
}

func cmdStart(kmID string) error {
	apiKey := os.Getenv("RB_API_KEY")
	if apiKey == "" {
		return errors.New(
			"Missing environment variable RB_API_KEY - you must set this",
		)
	}

	client := sdk.Client{
		APIKey:         apiKey,
		EnvironmentURL: sdk.EnvCommunity,
	}

	session, err := client.NewSession(kmID)
	if err != nil {
		return err
	}

	fmt.Println(session.ID)
	return nil
}

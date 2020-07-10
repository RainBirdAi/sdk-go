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
	fmt.Printf("  %s inject <session> <subject> <relationship> <object> <certainty>\n", os.Args[0])
	fmt.Println("    - Inject a fect into <session>")
	fmt.Printf("  %s query <session> <subject> <relationship> <object>\n", os.Args[0])
	fmt.Println("    - Make a query against <session>, receive question or answers")
	fmt.Printf("  %s response <session> <subject> <relationship> <object> <certainty>\n", os.Args[0])
	fmt.Println("    - Response to pending question in <session>")
	fmt.Printf("  %s start <kmid>\n", os.Args[0])
	fmt.Println("    - Start a new session for knowledge map <kmid>, receive a session ID for querying")
	fmt.Printf("  %s undo <session>\n", os.Args[0])
	fmt.Println("    - Roll back <session> by one interaction")
	fmt.Printf("  %s version\n", os.Args[0])
	fmt.Println("    - Report CLI and API versions")
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
	case "inject":
		if len(os.Args) != 7 {
			usage()
			os.Exit(0)
		}
		err = cmdInject(
			os.Args[2],
			os.Args[3],
			os.Args[4],
			os.Args[5],
			os.Args[6],
		)
	case "query":
		if len(os.Args) != 6 {
			usage()
			os.Exit(0)
		}
		err = cmdQuery(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	case "response":
		if len(os.Args) != 7 {
			usage()
			os.Exit(0)
		}
		err = cmdResponse(
			os.Args[2],
			os.Args[3],
			os.Args[4],
			os.Args[5],
			os.Args[6],
		)
	case "start":
		if len(os.Args) != 3 {
			usage()
			os.Exit(0)
		}
		err = cmdStart(os.Args[2])
	case "undo":
		if len(os.Args) != 3 {
			usage()
			os.Exit(0)
		}
		err = cmdUndo(os.Args[2])
	case "version":
		if len(os.Args) != 2 {
			usage()
			os.Exit(0)
		}
		cmdVersion()
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

func cmdInject(sessionID, sub, rel, obj, cf string) error {
	client := sdk.Client{
		EnvironmentURL: sdk.EnvCommunity,
	}

	session, err := client.ResumeSession(sessionID)
	if err != nil {
		return err
	}

	return session.Inject([]sdk.InjectFact{{
		Subject:      sub,
		Relationship: rel,
		Object:       obj,
		Certainty:    cf,
	}})
}

func cmdQuery(sessionID, sub, rel, obj string) error {
	client := sdk.Client{
		EnvironmentURL: sdk.EnvCommunity,
	}

	session, err := client.ResumeSession(sessionID)
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
		return err
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

func cmdResponse(sessionID, sub, rel, obj, cf string) error {
	client := sdk.Client{
		EnvironmentURL: sdk.EnvCommunity,
	}

	session, err := client.ResumeSession(sessionID)
	if err != nil {
		return err
	}

	question, answers, err := session.Response(sub, rel, obj, cf)
	if err != nil {
		return err
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
		return errors.New("Missing required environment variable RB_API_KEY")
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

func cmdUndo(sessionID string) error {
	client := sdk.Client{
		EnvironmentURL: sdk.EnvCommunity,
	}

	session, err := client.ResumeSession(sessionID)
	if err != nil {
		return err
	}

	question, answers, err := session.Undo()
	if err != nil {
		return err
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

func cmdVersion() {
	client := sdk.Client{
		EnvironmentURL: sdk.EnvCommunity,
	}
	api, err := client.Version()

	fmt.Printf("SDK: %s\n", sdk.Version)
	if err != nil {
		fmt.Println("API: ERR!")
		fmt.Println(err)
		return
	}
	fmt.Printf("API: %s\n", api)
}

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/tsuru/gnuflag"

	sdk "gitlab.com/rainbird-ai/sdk-go"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Printf("  %s help\n", os.Args[0])
	fmt.Println("    - Give this help text")
	fmt.Printf("  %s config\n", os.Args[0])
	fmt.Println("    - Show running configuration and exit")
	fmt.Printf("  %s inject <session> <subject> <relationship> <object> <certainty>\n", os.Args[0])
	fmt.Println("    - Inject a fact into <session>")
	fmt.Printf("  %s query <session> <subject> <relationship> <object>\n", os.Args[0])
	fmt.Println("    - Make a query against <session>, receive question or answers")
	fmt.Printf("  %s response <session> <subject> <relationship> <object> <certainty>\n", os.Args[0])
	fmt.Println("    - Response to pending question in <session>")
	fmt.Printf("  %s start <kmid> <contextid> --useDraft --version\n", os.Args[0])
	fmt.Println("    - Start a new session for knowledge map <kmid>, context ID <contextid>, useDraft flag --useDraft and specific version number flag --version, receive a session ID for querying")
	fmt.Printf("  %s undo <session>\n", os.Args[0])
	fmt.Println("    - Roll back <session> by one interaction")
	fmt.Printf("  %s version\n", os.Args[0])
	fmt.Println("    - Report CLI and API versions")
	fmt.Printf("  %s interactions <session> <interactionKey>\n", os.Args[0])
	fmt.Println("    - Get the interaction log for a session")
	fmt.Printf(" %s evidence <session> <factID> <evidenceKey>\n", os.Args[0])
	fmt.Println("    - Get the evidence for a given fact")
	fmt.Printf(" %s km <session>\n", os.Args[0])
	fmt.Println("    - Get knowledge map information from a session")
	fmt.Printf(" %s facts <session>\n", os.Args[0])
	fmt.Println("    - Get the facts from a session")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  RB_API_KEY - credentials to interact with Rainbird (Required)")
	fmt.Println("  RB_ENGINE - Select engine (Default is API's default, usually Laserbeak)")
	fmt.Println("  RB_API_URL - API URL (Default is https://api.rainbird.ai)")
}

var client = sdk.Client{
	APIKey:         os.Getenv("RB_API_KEY"),
	EnvironmentURL: sdk.EnvCommunity,
}

func main() {
	if len(os.Args) == 1 {
		usage()
		os.Exit(0)
	}

	if os.Getenv("RB_API_URL") != "" {
		client.EnvironmentURL = os.Getenv("RB_API_URL")
	}

	var err error

	switch os.Args[1] {
	case "help":
		usage()
		os.Exit(0)
	case "config":
		if len(os.Args) != 2 {
			usage()
			os.Exit(0)
		}
		cmdConfig()
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
		var useDraft = gnuflag.Bool("useDraft", false, "Use draft version")
		var version = gnuflag.Int("version", -1, "Use a specific version")
		gnuflag.Parse(true)

		if version != nil && *version == -1 {
			version = nil
		}

		if len(os.Args) < 4 {
			usage()
			os.Exit(0)
		}
		err = cmdStart(os.Args[2], os.Args[3], useDraft, version)
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
	case "interactions":
		if len(os.Args) != 4 {
			usage()
			os.Exit(0)
		}
		err = cmdInteractionsLog(os.Args[2], &os.Args[3])
	case "evidence":
		if len(os.Args) != 5 {
			usage()
			os.Exit(0)
		}
		err = cmdEvidence(os.Args[2], os.Args[3], &os.Args[4])
	case "km":
		if len(os.Args) != 3 {
			usage()
			os.Exit(0)
		}
		err = cmdSessionKmVersion(os.Args[2])
	case "facts":
		if len(os.Args) != 3 {
			usage()
			os.Exit(0)
		}
		err = cmdSessionFacts(os.Args[2])
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

func cmdConfig() {
	envVars := []struct {
		name     string
		required bool
	}{
		{"RB_API_KEY", true},
		{"RB_API_URL", false},
		{"RB_ENGINE", false},
	}

	for _, v := range envVars {
		fmt.Printf("%s\t", v.name)
		if os.Getenv(v.name) != "" {
			fmt.Printf("%s\n", os.Getenv(v.name))
			continue
		}
		if !v.required {
			fmt.Println("-- default --")
			continue
		}
		fmt.Println("!! missing !!")
	}
}

func cmdInject(sessionID, sub, rel, obj, cf string) error {
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

	questions, answers, err := session.Query(sub, rel, obj)
	if err != nil {
		return err
	}

	fmt.Println("QUESTIONS:")
	for _, q := range questions {
		fmt.Printf("  %s\n", &q)
	}

	fmt.Println("ANSWERS:")
	for _, a := range answers {
		fmt.Printf("  %s\n", &a)
	}
	return nil
}

func cmdResponse(sessionID, sub, rel, obj, cf string) error {
	session, err := client.ResumeSession(sessionID)
	if err != nil {
		return err
	}

	questions, answers, err := session.Response([]sdk.QAnswer{{
		Subject:      sub,
		Relationship: rel,
		Object:       obj,
		CF:           cf,
	}})

	if err != nil {
		return err
	}

	fmt.Println("QUESTIONS:")
	for _, q := range questions {
		fmt.Printf("  %s\n", &q)
	}

	fmt.Println("ANSWERS:")
	for _, a := range answers {
		fmt.Printf("  %s\n", &a)
	}

	return nil
}

func cmdStart(kmID string, contextID string, useDraft *bool, version *int) error {
	if client.APIKey == "" {
		return errors.New("missing required environment variable RB_API_KEY")
	}

	session, err := client.NewSession(kmID, contextID, useDraft, version)
	if err != nil {
		return err
	}

	fmt.Println(session.ID)
	return nil
}

func cmdUndo(sessionID string) error {
	session, err := client.ResumeSession(sessionID)
	if err != nil {
		return err
	}

	questions, answers, err := session.Undo()

	if err != nil {
		return err
	}

	fmt.Println("QUESTIONS:")
	for _, q := range questions {
		fmt.Printf("*  %s\n", &q)
	}

	fmt.Println("ANSWERS:")
	for _, a := range answers {
		fmt.Printf("*  %s\n", &a)
	}
	return nil
}

func cmdVersion() {
	api, err := client.Version()
	fmt.Printf("SDK: %s\n", sdk.Version)
	if err != nil {
		fmt.Println("API: ERR!")
		fmt.Println(err)
		return
	}
	fmt.Printf("API: %s\n", api)
}

func cmdInteractionsLog(sessionID string, interactionKey *string) error {
	interactions, err := client.Interactions(sessionID, interactionKey)
	if err != nil {
		return err
	}

	fmt.Println("INTERACTIONS:")
	for _, i := range interactions {
		event := i.Event
		fmt.Printf("Event: %s\n", event)
		fmt.Printf("Created: %s\n", i.Created)
		switch event {
		case sdk.StartEvent:
			start := i.Data.(sdk.Start)
			fmt.Printf("Start: %s\n\n", &start)
		case sdk.QuestionEvent:
			questions := i.Data.([]sdk.Question)
			for _, q := range questions {
				fmt.Printf("Questions: %s\n\n", &q)
			}
		case sdk.AnswerEvent:
			answers := i.Data.([]sdk.QAnswer)
			for _, a := range answers {
				fmt.Printf("Answer: %s\n\n", &a)
			}
		case sdk.QueryEvent:
			query := i.Data.(sdk.Query)
			fmt.Printf("Query: %s\n\n", &query)
		case sdk.InjectEvent:
			facts := i.Data.([]sdk.InjectFact)
			for _, f := range facts {
				fmt.Printf("Inject: %s\n\n", &f)
			}
		case sdk.DatasourceEvent:
			datasources := i.Data.([]sdk.Datasource)
			for _, d := range datasources {
				fmt.Printf("Datasource: %s\n\n", &d)
			}
		case sdk.ResultEvent:
			results := i.Data.([]sdk.Answer)
			for _, r := range results {
				fmt.Printf("Result: %s\n\n", &r)
			}
		}
	}
	fmt.Println("")
	return nil
}

func cmdEvidence(sessionID string, factID string, evidenceKey *string) error {
	evidence, err := client.Evidence(sessionID, factID, evidenceKey)
	if err != nil {
		return err
	}

	fmt.Printf("EVIDENCE: %s\n", evidence)
	return nil
}

func cmdSessionKmVersion(sessionID string) error {
	km, err := client.KnowledgeMapVersion(sessionID)
	if err != nil {
		return err
	}

	fmt.Printf("KM: %s\n", km)
	return nil
}

func cmdSessionFacts(sessionID string) error {
	facts, err := client.SessionFacts(sessionID)
	if err != nil {
		return err
	}

	fmt.Printf("FACTS: %s\n", facts)
	return nil
}

package sdk

import (
	"encoding/json"
	"fmt"
	"time"
)

// Event describes the various events that can be returned via the interactions endpoint
type Event int

// All possible event types
const (
	StartEvent Event = iota
	QuestionEvent
	AnswerEvent
	QueryEvent
	InjectEvent
	DatasourceEvent
	ResultEvent
)

func (e Event) String() string {
	return [7]string{"start", "question", "answer", "query", "inject", "datasource", "result"}[e]
}

var toEvent = map[string]Event{
	StartEvent.String():      StartEvent,
	QuestionEvent.String():   QuestionEvent,
	AnswerEvent.String():     AnswerEvent,
	QueryEvent.String():      QueryEvent,
	InjectEvent.String():     InjectEvent,
	DatasourceEvent.String(): DatasourceEvent,
	ResultEvent.String():     ResultEvent,
}

// InteractionEvent describes an event entry in the interactions log
type InteractionEvent struct {
	Event   Event       `json:"event"`
	Created time.Time   `json:"created"`
	Data    interface{} `json:"data"`
}

// InteractionResponse is the raw response from the api
type InteractionResponse struct {
	Event   string          `json:"event"`
	Values  json.RawMessage `json:"values"`
	Created string          `json:"created"`
}

// Start represents the start of a Rainbird session
type Start struct {
	SessionID   string `json:"sessionID"`
	UseDraft    bool   `json:"useDraft"`
	KmVersionID string `json:"kmVersionID"`
}

// String makes *Start satisfy fmt.Stringer
func (s *Start) String() string {
	return fmt.Sprintf(
		"%s (draft? %t) version %s",
		s.SessionID,
		s.UseDraft,
		s.KmVersionID,
	)
}

// Satisfy interfaces
var _ fmt.Stringer = (*Datasource)(nil)

// InteractionEvents converts the api response to interactionEvent domain objects
func InteractionEvents(response []InteractionResponse) ([]InteractionEvent, error) {
	var events []InteractionEvent
	for _, res := range response {
		event, ok := toEvent[res.Event]
		if !ok {
			return nil, fmt.Errorf("unsupported event: %s returned in the interactions log response", res.Event)
		}
		time, err := parseEventTime(res.Created)
		if err != nil {
			return nil, err
		}
		interaction := &InteractionEvent{
			Event:   event,
			Created: *time,
		}
		switch event {
		case StartEvent:
			start := struct {
				Start `json:"start"`
			}{}
			err := json.Unmarshal(res.Values, &start)
			if err != nil {
				return nil, err
			}
			interaction.Data = start.Start
		case QuestionEvent:
			question := struct {
				Questions []Question `json:"questions"`
			}{}
			err := json.Unmarshal(res.Values, &question)
			if err != nil {
				return nil, err
			}
			interaction.Data = question.Questions
		case AnswerEvent:
			answer := struct {
				Answers []QAnswer `json:"answers"`
			}{}
			err := json.Unmarshal(res.Values, &answer)
			if err != nil {
				return nil, err
			}
			interaction.Data = answer.Answers
		case QueryEvent:
			query := struct {
				Query `json:"query"`
			}{}
			err := json.Unmarshal(res.Values, &query)
			if err != nil {
				return nil, err
			}
			interaction.Data = query.Query
		case InjectEvent:
			inject := struct {
				Facts []InjectFact `json:"facts"`
			}{}
			err := json.Unmarshal(res.Values, &inject)
			if err != nil {
				return nil, err
			}
			interaction.Data = inject.Facts
		case DatasourceEvent:
			datasource := struct {
				DataSources []Datasource `json:"datasources"`
			}{}
			err := json.Unmarshal(res.Values, &datasource)
			if err != nil {
				return nil, err
			}
			interaction.Data = datasource.DataSources
		case ResultEvent:
			result := struct {
				Results []Answer `json:"results"`
			}{}
			err := json.Unmarshal(res.Values, &result)
			if err != nil {
				return nil, err
			}
			interaction.Data = result.Results
		}
		events = append(events, *interaction)
	}
	return events, nil
}

// Datasource describes a datasource inject event
type Datasource struct {
	Relationship string `json:"relationship,omitempty"`
	Certainty    int    `json:"certainty,omitempty"`
}

// String makes *interactionInject satisfy fmt.Stringer
func (i *Datasource) String() string {
	return fmt.Sprintf(
		"%s [%3d%%]",
		i.Relationship,
		i.Certainty,
	)
}

// Satisfy interfaces
var _ fmt.Stringer = (*Datasource)(nil)

// Query describes a user made query to the engine
type Query struct {
	Subject      *string `json:"subject,omitempty"`
	Relationship string  `json:"relationship,omitempty"`
	Object       *string `json:"object,omitempty"`
}

// String makes *query satisfy fmt.Stringer
func (q *Query) String() string {
	sub := "?"
	obj := "?"
	if q.Subject != nil {
		sub = *q.Subject
	}
	if q.Object != nil {
		obj = *q.Object
	}
	return fmt.Sprintf(
		"%s - %s - %s",
		sub,
		q.Relationship,
		obj,
	)
}

// Satisfy interfaces
var _ fmt.Stringer = (*Query)(nil)

func parseEventTime(timeStr string) (*time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

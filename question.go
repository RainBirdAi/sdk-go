package sdk

import "fmt"

// QuestionConcept is how the engine expressed concepts when asking a Question
type QuestionConcept struct {
	ConceptType     string      `json:"conceptType,omitempty"`
	Name            interface{} `json:"name,omitempty"`
	Type            string      `json:"type,omitempty"`
	Value           interface{} `json:"value,omitempty"`
	InvalidResponse bool        `json:"invalidResponse,omitempty"`
	FSID            uint64      `json:"fsid,omitempty"`
}

// KnownAnswer is a hint provided by the engine for information it already knows
type KnownAnswer struct {
	CF           float64     `json:"cf,omitempty"`
	Object       interface{} `json:"object,omitempty"`
	Relationship struct {
		Name string `json:"name,omitempty"`
	} `json:"relationship,omitempty"`
	Subject string `json:"subject,omitempty"`
}

// Question is the structure of a request from the engine when asking for more
// information from the user
type Question struct {
	AllowCF        bool                  `json:"allowCF,omitempty"`
	AllowUnknown   bool                  `json:"allowUnknown,omitempty"`
	CanAdd         bool                  `json:"canAdd,omitempty"`
	Concepts       []QuestionConcept     `json:"concepts,omitempty"`
	DataType       string                `json:"dataType,omitempty"`
	KnownAnswers   []KnownAnswer         `json:"knownAnswers,omitempty"`
	Object         interface{}           `json:"object,omitempty"`
	ObjectMetadata map[string][]MetaData `json:"objectMetadata,omitempty"`
	ObjectType     string                `json:"objectType,omitempty"`
	Plural         bool                  `json:"plural,omitempty"`
	Prompt         string                `json:"prompt,omitempty"`
	Relationship   string                `json:"relationship,omitempty"`
	Subject        string                `json:"subject,omitempty"`
	Type           string                `json:"type,omitempty"`
}

// String makes *Question satisfy fmt.Stringer
func (q *Question) String() string {
	sub := q.Subject
	if sub == "" {
		sub = "?"
	}
	obj := objectStr(q)

	return fmt.Sprintf("%s (%s - %s - %s)", q.Prompt, sub, q.Relationship, obj)
}

// objectStr handles the casting of an object and can return early
func objectStr(q *Question) string {
	if q.Object == nil {
		return "?"
	}

	strObj, ok := q.Object.(string)
	if ok && strObj == "" {
		return "?"
	}

	return fmt.Sprintf("%v", strObj)
}

// Satisfy interfaces
var _ fmt.Stringer = (*Question)(nil)

package sdk

import "fmt"

// QuestionConcept is how the engine expressed concepts when asking a Question
type QuestionConcept struct {
	ConceptType string
	FSID        uint64
	Name        interface{}
	Type        string
	Value       interface{}
}

// Question is the structure of a request from the engine when asking for more
// information from the user
type Question struct {
	AllowCF      bool
	AllowUnknown bool
	CanAdd       bool
	Concepts     []QuestionConcept
	DataType     string
	KnownAnswers []KnownAnswer
	Plural       bool
	Prompt       string
	Relationship string
	Subject      interface{}
	Object       interface{}
	Type         string
}

// String makes *Question satisfy fmt.Stringer
func (q *Question) String() string {
	sub := subjectStr(q)
	obj := objectStr(q)

	return fmt.Sprintf("%s (%s - %s - %s)", q.Prompt, sub, q.Relationship, obj)
}

// toString handles the casting of an object and can return early
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

func subjectStr(q *Question) string {
	if q.Subject == nil {
		return "?"
	}

	strSubj, ok := q.Subject.(string)
	if ok && strSubj == "" {
		return "?"
	}
	return fmt.Sprintf("%v", strSubj)
}

// Satisfy interfaces
var _ fmt.Stringer = (*Question)(nil)

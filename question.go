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
	Subject      string
	Object       interface{}
	Type         string
}

// String makes *Question satisfy fmt.Stringer
func (q *Question) String() string {
	sub := q.Subject
	if sub == "" {
		sub = "?"
	}

	obj := q.Object
	strObj, ok := obj.(string)
	if obj == nil || (ok && strObj == "") {
		obj = "?"
	}

	return fmt.Sprintf("%s (%s - %s - %s)", q.Prompt, sub, q.Relationship, obj)
}

// Satisfy interfaces
var _ fmt.Stringer = (*Question)(nil)

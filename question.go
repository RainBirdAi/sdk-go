package sdk

import "fmt"

type Question struct {
	AllowCF      bool
	AllowUnknown bool
	CanAdd       bool
	Concepts     []struct {
		ConceptType string
		FSID        uint64
		Name        string
		Type        string
		Value       string
	}
	DataType     string
	KnownAnswers []Answer
	Plural       bool
	Prompt       string
	Relationship string
	Subject      string
	Object       string
	Type         string
}

func (q Question) String() string {
	sub := q.Subject
	obj := q.Object

	if sub == "" {
		sub = "?"
	}
	if obj == "" {
		obj = "?"
	}

	return fmt.Sprintf("%s (%s - %s - %s)", q.Prompt, sub, q.Relationship, obj)
}

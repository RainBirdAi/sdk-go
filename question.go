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
	var sub, obj string

	if q.Subject.String() == "" {
		sub = "?"
	} else {
		sub = q.Subject.String()
	}
	if q.Object.String() == "" {
		obj = "?"
	} else {
		obj = q.Object.String()
	}

	return fmt.Sprintf("%s (%s - %s - %s)", q.Prompt, sub, q.Relationship, obj)
}

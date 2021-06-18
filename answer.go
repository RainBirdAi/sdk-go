package sdk

import "fmt"

// MetaData is a collection of additional information about Concepts
type MetaData struct {
	Data     string
	DataType string
}

// Answer is data provided by the engine as a result/decision
type Answer struct {
	Certainty       uint64
	FactID          string
	Object          interface{}
	ObjectMetadata  map[string][]MetaData
	ObjectValue     interface{}
	Relationship    string
	Subject         string
	SubjectMetadata map[string][]MetaData
	SubjectValue    interface{}
}

// KnownAnswer is a hint provided by the engine for information it already knows
type KnownAnswer struct {
	CF           float64
	Object       interface{}
	Relationship struct {
		Name string
	}
	Subject string
}

// String makes *Answer satisfy fmt.Stringer
func (a *Answer) String() string {
	obj := a.Object
	if obj == nil {
		obj = ""
	}

	return fmt.Sprintf(
		"%s - %s - %v [%3d%%]",
		a.Subject,
		a.Relationship,
		obj,
		a.Certainty,
	)
}

// Satisfy interfaces
var _ fmt.Stringer = (*Answer)(nil)

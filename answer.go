package sdk

import "fmt"

// MetaData is a collection of additional information about Concepts
type MetaData struct {
	Data     string
	DataType string
}

// Answer is data provided by the engine as a result/decision
type Answer struct {
	Subject         string                `json:"subject,omitempty"`
	Object          interface{}           `json:"object,omitempty"`
	Certainty       uint64                `json:"certainty,omitempty"`
	FactID          string                `json:"factID,omitempty"`
	Relationship    string                `json:"relationship,omitempty"`
	SubjectMetadata map[string][]MetaData `json:"subjectMetadata,omitempty"`
	SubjectValue    interface{}           `json:"subjectValue,omitempty"`
	ObjectMetadata  map[string][]MetaData `json:"objectMetadata,omitempty"`
	ObjectValue     interface{}           `json:"objectValue,omitempty"`
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

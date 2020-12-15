package sdk

import "fmt"

type MetaData struct {
	Data     string
	DataType string
}

type Answer struct {
	Certainty       uint64
	FactID          string
	Object          interface{}
	ObjectMetadata  map[string][]MetaData
	Relationship    string
	Subject         string
	SubjectMetadata map[string][]MetaData
}

type KnownAnswer struct {
	CF           float64
	Object       interface{}
	Relationship Relationship
	Subject      string
}

type Relationship struct {
	Name string
}

func (a Answer) String() string {
	return fmt.Sprintf(
		"%s - %s - %s [%3d%%]",
		a.Subject,
		a.Relationship,
		a.Object,
		a.Certainty,
	)
}

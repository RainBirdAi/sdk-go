package sdk

import "fmt"

type MetaData struct {
	Data     string
	DataType string
}

type Answer struct {
	Certainty        uint64
	FactID           string
	Object           string
	ObjectMetadata   map[string][]MetaData
	Relationship     string
	RelationshipType string
	Subject          string
	SubjectMetadata  map[string][]MetaData
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

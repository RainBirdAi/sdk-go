package sdk

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

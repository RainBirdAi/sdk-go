package sdk

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
	DataType string
	// KnownAnswers TODO
	Plural       bool
	Prompt       string
	Relationship string
	Subject      string
	Type         string
}

func (q Question) String() string {
	return q.Prompt
}

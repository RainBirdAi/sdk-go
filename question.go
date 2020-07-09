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
	Promps       string
	Relationship string
	Subject      string
	Type         string
}

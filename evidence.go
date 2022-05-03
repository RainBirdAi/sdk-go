package sdk

type Evidence struct {
	FactID string
	Source string
	Fact   InjectFact
	Rule   Rule
	Time   int
}

type Rule struct {
	Bindings interface{}
}

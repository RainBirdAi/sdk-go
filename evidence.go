package sdk

import "fmt"

type Evidence struct {
	FactID string
	Source string
	Fact   Fact
	Rule   Rule
	Time   int
}

// String makes *Answer satisfy fmt.Stringer
func (e *Evidence) String() string {
	return fmt.Sprintf("%s", e.Fact.Subject)
}

type Fact struct {
	Subject      ConceptInstance
	Relationship Relationship
	Object       ConceptInstance
	Certainty    int
}
type Relationship struct {
	Type string
}

type ConceptInstance struct {
	Type     string
	Value    interface{}
	dataType string
}

type Rule struct {
	Bindings struct {
		conditions []Condition
	}
}

type Condition interface {
	Salience() int
}

type ConditionRelationship struct {
	Certainty    int
	FactID       string
	FactKey      *string
	Object       string
	ObjectType   string
	Relationship string
	Subject      string
	salience     int
}

func (cr *ConditionRelationship) Salience() int {
	return cr.salience
}

type ConditionExpression struct {
	WasMet     bool
	Expression struct {
		Text string
	}
	salience int
}

func (ce *ConditionExpression) Salience() int {
	return ce.salience
}

// `{
//     "factID": "WA:RF:66812f77defed9cb592da054e89c498d717019dafe3daadd9e0afdd4b2c999b0",
//     "source": "rule",
//     "fact": {
//         "subject": {
//             "type": "Person",
//             "value": "Lucy",
//             "dataType": "string"
//         },
//         "relationship": {
//             "type": "might speak"
//         },
//         "object": {
//             "type": "Language",
//             "value": "English",
//             "dataType": "string"
//         },
//         "certainty": 100
//     },
//     "time": 1651662791288,
//     "rule": {
//         "bindings": {
//             "S": "Lucy",
//             "O": "English",
//             "COUNTRY": "England"
//         },
//         "conditions": [
//             {
//                 "subject": "Lucy",
//                 "relationship": "lives in",
//                 "object": "England",
//                 "salience": 100,
//                 "certainty": 100,
//                 "factID": "WA:AF:e0d46da0883aa26a782a60adb2d97714c2294b996ebc73d14100a005def815bb",
//                 "objectType": "string",
//                 "factKey": "30d74cc5-7eb6-4c88-8e12-4897b47e3ee7"
//             },
//             {
//                 "subject": "England",
//                 "relationship": "has national language",
//                 "object": "English",
//                 "salience": 100,
//                 "certainty": 100,
//                 "factID": "WA:KF:bda2ac8b66ff2b24e91b0d9e162037cbd05b62094a9461fe504d2ce50e1d691e",
//                 "objectType": "string"
//             }
//         ]
//     }
// }`

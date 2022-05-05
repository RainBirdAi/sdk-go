package sdk

import (
	"errors"
	"fmt"
)

type ConditionType = int

const (
	RelationshipType ConditionType = iota
	ExpressionType
)

// EvidenceResponse is the raw response from the api
type EvidenceResponse struct {
	FactID       string       `json:"factId,omitempty"`
	Source       string       `json:"source,omitempty"`
	Fact         Fact         `json:"fact,omitempty"`
	RuleResponse ruleResponse `json:"rule,omitempty"`
	Time         int          `json:"time,omitempty"`
}

// Evidence is an abstraction upon EvidenceResponse
type Evidence struct {
	FactID string `json:"factId,omitempty"`
	Source string `json:"source,omitempty"`
	Fact   Fact   `json:"fact,omitempty"`
	Rule   Rule   `json:"rule,omitempty"`
	Time   int    `json:"time,omitempty"`
}

// String makes *Answer satisfy fmt.Stringer
func (e *Evidence) String() string {
	return fmt.Sprintf("%s", e.Fact.Subject)
}

func AsEvidence(response EvidenceResponse) (Evidence, error) {
	rule, err := asRule(response.RuleResponse)
	if err != nil {
		return Evidence{}, err
	}

	return Evidence{
		FactID: response.FactID,
		Source: response.Source,
		Fact:   response.Fact,
		Rule:   rule,
		Time:   response.Time,
	}, nil
}

type Fact struct {
	Subject      ConceptInstance `json:"subject,omitempty"`
	Relationship Relationship    `json:"relationship,omitempty"`
	Object       ConceptInstance `json:"object,omitempty"`
	Certainty    int             `json:"certainty,omitempty"`
}

type Relationship struct {
	Type string `json:"type,omitempty"`
}

type ConceptInstance struct {
	Type     string      `json:"type,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	DataType string      `json:"dataType,omitempty"`
}

type ruleResponse struct {
	Bindings   map[string]string `json:"bindings,omitempty"`
	Conditions []rawCondition    `json:"conditions,omitempty"`
}

// Rule is an abstraction of RuleReponse
type Rule struct {
	Bindings   map[string]string
	Conditions []Condition
}

func asRule(response ruleResponse) (Rule, error) {
	conditions, err := asConditions(response.Conditions)
	if err != nil {
		return Rule{}, err
	}

	return Rule{
		Bindings:   response.Bindings,
		Conditions: conditions,
	}, nil
}

type Condition interface {
	Type() ConditionType
	Salience() int
}

func asConditions(raw []rawCondition) ([]Condition, error) {
	var conditions []Condition
	for _, c := range raw {
		if c.Relationship != "" {
			rel := ConditionRelationship{
				Certainty:    c.Certainty,
				FactID:       c.FactID,
				FactKey:      c.FactKey,
				Object:       c.Object,
				ObjectType:   c.ObjectType,
				Relationship: c.Relationship,
				Subject:      c.Subject,
				salience:     c.Salience,
			}
			conditions = append(conditions, rel)
		} else if c.Expression.Text != "" {
			exp := ConditionExpression{
				WasMet:     c.WasMet,
				Expression: c.Expression,
				salience:   c.Salience,
			}
			conditions = append(conditions, exp)
		} else {
			return conditions, errors.New("unsupported condition type")
		}
	}
	return conditions, nil
}

// rawCondition encapsulates any kind of condition that can be present.
type rawCondition struct {
	Certainty    int         `json:"certainty,omitempty"`
	FactID       string      `json:"factID,omitempty"`
	FactKey      *string     `json:"factKey,omitempty"`
	Object       interface{} `json:"object,omitempty"`
	ObjectType   string      `json:"objectType,omitempty"`
	Relationship string      `json:"relationship,omitempty"`
	Subject      interface{} `json:"subject,omitempty"`
	Salience     int         `json:"salience,omitempty"`
	WasMet       bool        `json:"wasMet,omitempty"`
	Expression   struct {
		Text string `json:"text,omitempty"`
	} `json:"expression,omitempty"`
}

type ConditionRelationship struct {
	Certainty    int         `json:"certainty,omitempty"`
	FactID       string      `json:"factID,omitempty"`
	FactKey      *string     `json:"factKey,omitempty"`
	Object       interface{} `json:"object,omitempty"`
	ObjectType   string      `json:"objectType,omitempty"`
	Relationship string      `json:"relationship,omitempty"`
	Subject      interface{} `json:"subject,omitempty"`
	salience     int         `json:"salience,omitempty"`
}

func (cr ConditionRelationship) Type() ConditionType {
	return RelationshipType
}

func (cr ConditionRelationship) Salience() int {
	return cr.salience
}

type ConditionExpression struct {
	WasMet     bool       `json:"wasMet,omitempty"`
	Expression Expression `json:"expression,omitempty"`
	salience   int        `json:"salience,omitempty"`
}

type Expression struct {
	Text string `json:"text,omitempty"`
}

func (cr ConditionExpression) Type() ConditionType {
	return ExpressionType
}

func (ce ConditionExpression) Salience() int {
	return ce.salience
}

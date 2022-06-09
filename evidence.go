package sdk

import (
	"errors"
	"fmt"
)

// ConditionType describes the various types of evidence conditions that can be returned
type ConditionType = int

// All ConditionTypes
const (
	RelationshipType ConditionType = iota
	ExpressionType
)

// evidenceResponse is the raw response from the api
type evidenceResponse struct {
	FactID       string        `json:"factId,omitempty"`
	Source       string        `json:"source,omitempty"`
	Fact         factResponse  `json:"fact,omitempty"`
	RuleResponse *ruleResponse `json:"rule,omitempty"`
	Time         int           `json:"time,omitempty"`
}

// Evidence is produced from the raw EvidenceResponse
type Evidence struct {
	Fact Fact  `json:"fact,omitempty"`
	Rule *Rule `json:"rule,omitempty"`
	Time int   `json:"time,omitempty"`
}

// String makes *Evidence satisfy fmt.Stringer
func (e *Evidence) String() string {
	return e.Fact.String()
}

var _ fmt.Stringer = (*Evidence)(nil)

func asEvidence(response evidenceResponse) (Evidence, error) {
	var rule *Rule
	if response.RuleResponse != nil {
		r, err := asRule(*response.RuleResponse)
		if err != nil {
			return Evidence{}, err
		}
		rule = r
	}
	fr := response.Fact

	return Evidence{
		Fact: Fact{
			ID:           response.FactID,
			Source:       response.Source,
			Subject:      fr.Subject,
			Relationship: fr.Relationship.Type,
			Object:       fr.Object,
			Certainty:    fr.Certainty,
		},
		Rule: rule,
		Time: response.Time,
	}, nil
}

type factResponse struct {
	Subject      ConceptInstance `json:"subject,omitempty"`
	Relationship Relationship    `json:"relationship,omitempty"`
	Object       ConceptInstance `json:"object,omitempty"`
	Certainty    int             `json:"certainty,omitempty"`
}

type Fact struct {
	ID           string          `json:"id,omitempty"`
	Source       string          `json:"source,omitempty"`
	Subject      ConceptInstance `json:"subject,omitempty"`
	Relationship string          `json:"relationship,omitempty"`
	Object       ConceptInstance `json:"object,omitempty"`
	Certainty    int             `json:"certainty,omitempty"`
}

// String makes *Fact satifsy fmt.Stringer
func (f *Fact) String() string {
	return fmt.Sprintf(
		"%s (%s): %v, %s, %v (%d)",
		f.ID,
		f.Source,
		f.Subject.String(),
		f.Relationship,
		f.Object.String(),
		f.Certainty,
	)
}

var _ fmt.Stringer = (*Fact)(nil)

// Relationship connects concepts in a knowledge map
type Relationship struct {
	Type string `json:"type,omitempty"`
}

// ConceptInstance is an instance of a knowledge map concept
type ConceptInstance struct {
	Type     *string     `json:"type,omitempty"`
	Concept  *string     `json:"concept,omitempty"`
	Value    interface{} `json:"value,omitempty"`
	DataType string      `json:"dataType,omitempty"`
}

// GetConcept returns the name of the concept - sometimes refered to as the type
func (ci *ConceptInstance) GetConcept() string {
	if ci.Concept == nil {
		return *ci.Type
	}
	return *ci.Concept
}

func (ci *ConceptInstance) String() string {
	return fmt.Sprintf("%s", ci.Value)
}

var _ fmt.Stringer = (*ConceptInstance)(nil)

type ruleResponse struct {
	Bindings   map[string]interface{} `json:"bindings,omitempty"`
	Conditions []rawCondition         `json:"conditions,omitempty"`
}

// Rule describes the conditions upon which a fact was produced
type Rule struct {
	Bindings   map[string]interface{}
	Conditions []Conditioner
}

func asRule(response ruleResponse) (*Rule, error) {
	conditions, err := asConditions(response.Conditions)
	if err != nil {
		return nil, err
	}

	return &Rule{
		Bindings:   response.Bindings,
		Conditions: conditions,
	}, nil
}

// Conditioner is a common interface for the various ConditionTypes
type Conditioner interface {
	Type() ConditionType
	Salience() int
}

// asConditions converts the raw response conditions to conditions
func asConditions(raw []rawCondition) ([]Conditioner, error) {
	var conditions []Conditioner
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

// ConditionRelationship describes a condition created from a relationship
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

// Type of ConditionType
func (cr ConditionRelationship) Type() ConditionType {
	return RelationshipType
}

// Salience value of the given condition
func (cr ConditionRelationship) Salience() int {
	return cr.salience
}

// ConditionExpression describes a condition created from an expression
type ConditionExpression struct {
	WasMet     bool       `json:"wasMet,omitempty"`
	Expression Expression `json:"expression,omitempty"`
	salience   int        `json:"salience,omitempty"`
}

// Expression is the text representation of a knowledge map expression
type Expression struct {
	Text string `json:"text,omitempty"`
}

// Type of ConditionType
func (ce ConditionExpression) Type() ConditionType {
	return ExpressionType
}

// Salience value of the given condition
func (ce ConditionExpression) Salience() int {
	return ce.salience
}

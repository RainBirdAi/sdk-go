package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuestionToString(t *testing.T) {
	testCases := []struct {
		description string
		question    Question
		expect      string
	}{
		{
			description: "Empty",
			question:    Question{},
			expect:      " (? -  - ?)",
		},
		{
			description: "Unknown object",
			question: Question{
				Prompt:       "Which language does John speak?",
				Subject:      "John",
				Relationship: "speaks",
			},
			expect: "Which language does John speak? (John - speaks - ?)",
		},
		{
			description: "Unknown subject",
			question: Question{
				Prompt:       "Who lives in England?",
				Relationship: "lives in",
				Object:       "England",
			},
			expect: "Who lives in England? (? - lives in - England)",
		},
		{
			description: "Unknown object as empty string",
			question: Question{
				Prompt:       "Who lives in England?",
				Relationship: "lives in",
				Object:       "",
			},
			expect: "Who lives in England? (? - lives in - ?)",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expect, tc.question.String())
		})
	}
}

package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuestionToString(t *testing.T) {
	testCases := []struct {
		description string
		answer      Question
		expect      string
	}{
		{
			description: "Empty",
			answer:      Question{},
			expect:      " (? -  - ?)",
		},
		{
			description: "Unknown object",
			answer: Question{
				Prompt:       "Which language does John speak?",
				Subject:      "John",
				Relationship: "speaks",
			},
			expect: "Which language does John speak? (John - speaks - ?)",
		},
		{
			description: "Unknown subject",
			answer: Question{
				Prompt:       "Who lives in England?",
				Relationship: "lives in",
				Object:       "England",
			},
			expect: "Who lives in England? (? - lives in - England)",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expect, tc.answer.String())
		})
	}
}

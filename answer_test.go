package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnswerToString(t *testing.T) {
	testCases := []struct {
		description string
		answer      Answer
		expect      string
	}{
		{
			description: "Empty",
			answer:      Answer{},
			expect:      " -  -  [  0%]",
		},
		{
			description: "John speaks English 50%",
			answer: Answer{
				Subject:      "John",
				Relationship: "speaks",
				Object:       "English",
				Certainty:    50,
			},
			expect: "John - speaks - English [ 50%]",
		},
		{
			description: "Dan speaks German 100%",
			answer: Answer{
				Subject:      "Dan",
				Relationship: "speaks",
				Object:       "German",
				Certainty:    100,
			},
			expect: "Dan - speaks - German [100%]",
		},
		{
			description: "John favourite number 137 35%",
			answer: Answer{
				Subject:      "John",
				Relationship: "favourite number",
				Object:       137,
				Certainty:    35,
			},
			expect: "John - favourite number - 137 [ 35%]",
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

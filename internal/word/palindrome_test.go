package word_test

import (
	"fmt"
	"testing"

	"github.com/cdimonaco/e2e-framework-usage-demo-talk/internal/word"
	"github.com/stretchr/testify/assert"
)

func TestIsPalindrome(t *testing.T) {
	type testCase struct {
		inputString    string
		expectedResult bool
	}

	for _, tc := range []testCase{
		{
			inputString:    "aba",
			expectedResult: true,
		},
		{
			inputString:    "rancher",
			expectedResult: false,
		},
		{
			inputString:    "aBA",
			expectedResult: true,
		},
		{
			inputString:    "aba-aba",
			expectedResult: true,
		},
		{
			inputString:    "level",
			expectedResult: true,
		},
	} {
		expectationText := "be a palindrome"
		if !tc.expectedResult {
			expectationText = "not be a palindrome"
		}
		t.Run(fmt.Sprintf("%s should %s", tc.inputString, expectationText), func(t *testing.T) {
			assert.Equal(t, tc.expectedResult, word.IsPalindrome(tc.inputString))
		})
	}
}

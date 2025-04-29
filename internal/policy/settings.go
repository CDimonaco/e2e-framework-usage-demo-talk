package policy

import (
	"encoding/json"
	"fmt"

	"github.com/cdimonaco/e2e-framework-usage-demo-talk/internal/word"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
)

type AllowedPalindromeError struct {
	Field string
}

func (e AllowedPalindromeError) Error() string {
	return fmt.Sprintf("%s is not a palindrome, it could not be used as allowed palindrome", e.Field)
}

type Settings struct {
	AllowedPalindromes []string `json:"allowed_palindromes"`
}

func NewSettingsFromValidationRequest(
	validationReq *kubewarden_protocol.ValidationRequest,
) (*Settings, error) {
	var settings Settings
	err := json.Unmarshal(validationReq.Settings, &settings)
	if err != nil {
		return nil, fmt.Errorf(
			"could not create a settings from a validation request: %w",
			err,
		)
	}
	return &settings, nil
}

// Check if the allowed palindromes are really AllowedPalindromes.
func (s *Settings) Validate() error {
	// Cannot use slices package functions, not supported by tinygo
	for _, ap := range s.AllowedPalindromes {
		if !word.IsPalindrome(ap) {
			return AllowedPalindromeError{Field: ap}
		}
	}
	return nil
}

func (s *Settings) IsAnAllowedPalindrome(palindrome string) bool {
	for _, ap := range s.AllowedPalindromes {
		if ap == palindrome {
			return true
		}
	}
	return false
}

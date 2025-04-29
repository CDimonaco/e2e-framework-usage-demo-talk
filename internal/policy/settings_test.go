package policy_test

import (
	"testing"

	"github.com/cdimonaco/e2e-framework-usage-demo-talk/internal/policy"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsCreationSuccess(t *testing.T) {
	t.Run("should create settings without errors and with provided allowed palindromes", func(t *testing.T) {
		vr := kubewarden_protocol.ValidationRequest{
			Settings: []byte(`
			{
				"allowed_palindromes": ["bob", "aba"]
			}`),
		}

		expectedSettings := policy.Settings{
			AllowedPalindromes: []string{"bob", "aba"},
		}
		settings, err := policy.NewSettingsFromValidationRequest(&vr)
		require.NoError(t, err)
		assert.Equal(t, expectedSettings, *settings)
	})
	t.Run("should create settings whith empty palindromes when they are not provided", func(t *testing.T) {
		vr := kubewarden_protocol.ValidationRequest{
			Settings: []byte(`
			{
			}`),
		}

		expectedSettings := policy.Settings{
			AllowedPalindromes: nil,
		}
		settings, err := policy.NewSettingsFromValidationRequest(&vr)
		require.NoError(t, err)
		assert.Equal(t, expectedSettings, *settings)
	})
}

func TestSettingsCreationError(t *testing.T) {
	vr := kubewarden_protocol.ValidationRequest{
		Settings: []byte(`{`),
	}
	settings, err := policy.NewSettingsFromValidationRequest(&vr)
	assert.Nil(t, settings)
	assert.ErrorContains(t, err, "could not create a settings from a validation request")
}

func TestSettingsValidationErrors(t *testing.T) {
	type testCase struct {
		name          string
		settings      policy.Settings
		expectedError error
	}

	for _, tc := range []testCase{
		{
			name: "rancher not pass the allowed palindrome validation",
			settings: policy.Settings{
				AllowedPalindromes: []string{"rancher", "aba"},
			},
			expectedError: policy.AllowedPalindromeError{Field: "rancher"},
		},
		{
			name: "carmine not pass the allowed palindrome validation",
			settings: policy.Settings{
				AllowedPalindromes: []string{"aba", "obo", "carmine"},
			},
			expectedError: policy.AllowedPalindromeError{Field: "carmine"},
		},
		{
			name: "palindrome validation without errors",
			settings: policy.Settings{
				AllowedPalindromes: []string{"aba", "level", "ebe"},
			},
			expectedError: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.settings.Validate()
			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestIsAnAllowedPalindrome(t *testing.T) {
	t.Run("should return true if a palindrome is alloed", func(t *testing.T) {
		settings := policy.Settings{
			AllowedPalindromes: []string{"level", "aba"},
		}

		found := settings.IsAnAllowedPalindrome("level")
		assert.True(t, found)
	})

	t.Run("should return false if a palindrome is not allowed", func(t *testing.T) {
		settings := policy.Settings{
			AllowedPalindromes: []string{"level", "aba"},
		}

		found := settings.IsAnAllowedPalindrome("ebe")
		assert.False(t, found)
	})
}

package policy_test

import (
	"encoding/json"
	"testing"

	"github.com/cdimonaco/e2e-framework-usage-demo-talk/internal/policy"
	"github.com/francoispqt/onelog"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	metav1 "github.com/kubewarden/k8s-objects/apimachinery/pkg/apis/meta/v1"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
	kubewarden_testing "github.com/kubewarden/policy-sdk-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSettingsSuccess(t *testing.T) {
	validateSettings := policy.NewValidateSettings(&onelog.Logger{})
	validSettings := `
	{
		"allowed_palindromes": ["bob", "aba"]
	}`

	var protocolValidationResult kubewarden_protocol.SettingsValidationResponse
	result, err := validateSettings([]byte(validSettings))
	require.NoError(t, err)
	err = json.Unmarshal(result, &protocolValidationResult)
	require.NoError(t, err)
	assert.True(t, protocolValidationResult.Valid)
}

func TestValidateSettingsErrors(t *testing.T) {
	validateSettings := policy.NewValidateSettings(&onelog.Logger{})
	type testCase struct {
		expectedErrorContent string
		settings             []byte
		name                 string
	}

	for _, tc := range []testCase{
		{
			name: "should reject settings validation when the settings could not be unmarshaled",
			settings: []byte(
				`{`,
			),
			expectedErrorContent: "policy settings not valid, error during the unmarshal",
		},
		{
			name: "should reject settings validation when settings domain validation fails",
			settings: []byte(
				`
				{
					"allowed_palindromes": ["rancher"]
				}
				`,
			),
			expectedErrorContent: "provided settings are not valid",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var protocolValidationResult kubewarden_protocol.SettingsValidationResponse
			result, err := validateSettings(tc.settings)
			require.NoError(t, err)
			err = json.Unmarshal(result, &protocolValidationResult)
			require.NoError(t, err)
			assert.False(t, protocolValidationResult.Valid)
			assert.Contains(t, *protocolValidationResult.Message, tc.expectedErrorContent)
		})
	}
}

func TestValidateSuccess(t *testing.T) {
	validate := policy.NewValidate(&onelog.Logger{})

	type testCase struct {
		name     string
		settings policy.Settings
		pod      corev1.Pod
	}

	for _, tc := range []testCase{
		{
			name:     "should pass validation when settings are empty and the labels key are not palindrome",
			settings: policy.Settings{},
			pod: corev1.Pod{
				Metadata: &metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						"env": "development",
					},
				},
			},
		},
		{
			name:     "should pass validation when settings are empty and the pod has no labels",
			settings: policy.Settings{},
			pod: corev1.Pod{
				Metadata: &metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
			},
		},
		{
			name: "should pass validation when settings provide a list of allowed palindromes and the pod contains a palindrome allowed label key", //nolint:lll
			settings: policy.Settings{
				AllowedPalindromes: []string{"level"},
			},
			pod: corev1.Pod{
				Metadata: &metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						"level": "error",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var response kubewarden_protocol.ValidationResponse
			payload, err := kubewarden_testing.BuildValidationRequest(&tc.pod, &tc.settings)
			require.NoError(t, err)
			result, err := validate(payload)
			require.NoError(t, err)
			err = json.Unmarshal(result, &response)
			require.NoError(t, err)
			assert.True(t, response.Accepted)
		})
	}
}

func TestValidatePodLabelsError(t *testing.T) {
	validate := policy.NewValidate(&onelog.Logger{})
	type testCase struct {
		name                 string
		settings             policy.Settings
		pod                  corev1.Pod
		expectedErrorContent string
	}

	for _, tc := range []testCase{
		{
			name:                 "should return error when settings are empty and there is a label key palindrome",
			settings:             policy.Settings{},
			expectedErrorContent: "pod label with key level not allowed, the word is a palindrome",
			pod: corev1.Pod{
				Metadata: &metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						"level": "error",
					},
				},
			},
		},
		{
			name: "should return error when allowed palindromes are provided in the settings and there is a label key palindrome not explicitely allowed", //nolint:lll
			settings: policy.Settings{
				AllowedPalindromes: []string{"aba"},
			},
			expectedErrorContent: "pod label with key level not allowed, the word is a palindrome",
			pod: corev1.Pod{
				Metadata: &metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						"level": "error",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var response kubewarden_protocol.ValidationResponse
			payload, err := kubewarden_testing.BuildValidationRequest(&tc.pod, &tc.settings)
			require.NoError(t, err)
			result, err := validate(payload)
			require.NoError(t, err)
			err = json.Unmarshal(result, &response)
			require.NoError(t, err)
			assert.False(t, response.Accepted)
			assert.Nil(t, response.Code)
			assert.Contains(t, *response.Message, tc.expectedErrorContent)
		})
	}
}

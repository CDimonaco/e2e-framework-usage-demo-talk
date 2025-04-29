package policy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cdimonaco/e2e-framework-usage-demo-talk/internal/word"
	"github.com/francoispqt/onelog"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
	"github.com/tidwall/gjson"
	"github.com/wapc/wapc-guest-tinygo"
)

const httpBadRequestStatusCode = 400

func NewValidate(logger *onelog.Logger) wapc.Function {
	ctxLogger := logger.With(func(e onelog.Entry) {
		e.String("context", "validate")
	})
	return func(payload []byte) ([]byte, error) {
		validationRequest := kubewarden_protocol.ValidationRequest{}
		err := json.Unmarshal(payload, &validationRequest)
		if err != nil {
			ctxLogger.ErrorWithFields("could not unmarshal validation request", func(e onelog.Entry) {
				e.Err("error", err)
			})
			return kubewarden.RejectRequest(
				kubewarden.Message(err.Error()),
				kubewarden.Code(httpBadRequestStatusCode))
		}

		settings, err := NewSettingsFromValidationRequest(&validationRequest)
		if err != nil {
			ctxLogger.ErrorWithFields("could not create settings from validation request", func(e onelog.Entry) {
				e.Err("error", err)
			})
			return kubewarden.RejectRequest(
				kubewarden.Message(err.Error()),
				kubewarden.Code(httpBadRequestStatusCode))
		}

		podMetadata := gjson.GetBytes(
			validationRequest.Request.Object,
			"metadata",
		)

		rawPodLabels := podMetadata.Get("labels")
		podName := podMetadata.Get("name").String()

		var invalidLabelErr error
		rawPodLabels.ForEach(func(key, _ gjson.Result) bool {
			labelKey := key.String()

			if word.IsPalindrome(labelKey) && !settings.IsAnAllowedPalindrome(labelKey) {
				invalidLabelErr = fmt.Errorf("pod label with key %s not allowed, the word is a palindrome", labelKey)
			}

			return invalidLabelErr == nil
		})

		if invalidLabelErr != nil {
			ctxLogger.InfoWithFields("could not validate pod, palindrome label keys found", func(e onelog.Entry) {
				e.String("pod_name", podName)
				e.String("allowed_palindromes", strings.Join(settings.AllowedPalindromes, ","))
			})
			return kubewarden.RejectRequest(
				kubewarden.Message(invalidLabelErr.Error()),
				kubewarden.NoCode,
			)
		}

		return kubewarden.AcceptRequest()
	}
}

func NewValidateSettings(logger *onelog.Logger) wapc.Function {
	ctxLogger := logger.With(func(e onelog.Entry) {
		e.String("context", "validate_settings")
	})
	return func(payload []byte) ([]byte, error) {
		var policySettings Settings
		err := json.Unmarshal(payload, &policySettings)
		if err != nil {
			ctxLogger.ErrorWithFields("could not unmarshal policy settings", func(e onelog.Entry) {
				e.Err("error", err)
			})
			return kubewarden.RejectSettings(
				kubewarden.Message(fmt.Sprintf("policy settings not valid, error during the unmarshal: %v", err)),
			)
		}

		err = policySettings.Validate()
		if err != nil {
			ctxLogger.ErrorWithFields("policy settings not valid", func(e onelog.Entry) {
				e.Err("error", err)
			})
			return kubewarden.RejectSettings(
				kubewarden.Message(fmt.Sprintf("provided settings are not valid: %v", err)),
			)
		}

		return kubewarden.AcceptSettings()
	}
}

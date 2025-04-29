package main

import (
	"github.com/cdimonaco/e2e-framework-usage-demo-talk/internal/policy"
	onelog "github.com/francoispqt/onelog"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	logWriter := kubewarden.KubewardenLogWriter{}
	logger := onelog.New(
		&logWriter,
		onelog.ALL,
	)

	validatePolicy := policy.NewValidate(logger)
	validateSettings := policy.NewValidateSettings(logger)

	wapc.RegisterFunctions(wapc.Functions{
		"validate":          validatePolicy,
		"validate_settings": validateSettings,
	})
}

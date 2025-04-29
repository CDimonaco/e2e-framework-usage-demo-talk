//go:build integration

package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	kubewardenv1 "github.com/kubewarden/kubewarden-controller/api/policies/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/pkg/utils"
)

func TestPolicyApply(t *testing.T) {
	policyWithoutSettingsFeature := features.New("Policy without settings enforcing").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			p := utils.RunCommand(fmt.Sprintf("kubectl apply -f %s", "fixtures/policy-descriptor.yml"))
			assert.NoError(t, p.Err())
			dep := kubewardenv1.ClusterAdmissionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "palindrome-label-pod",
					Namespace: "kubewarden",
				},
			}
			// wait for the deployment of the policy
			err := wait.For(
				conditions.New(c.Client().Resources()).ResourceMatch(&dep, func(object k8s.Object) bool {
					clusterAdmissionPolicy := object.(*kubewardenv1.ClusterAdmissionPolicy)
					return clusterAdmissionPolicy.Status.PolicyStatus == kubewardenv1.PolicyStatusActive
				}),
				wait.WithTimeout(time.Minute*1),
			)
			assert.NoError(t, err)

			return ctx
		}).
		Assess("should not create a pod with a palindrome label", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			p := utils.RunCommand(fmt.Sprintf("kubectl apply -f %s", "fixtures/palindrome-pod-descriptor.yml"))
			assert.Error(t, p.Err())
			assert.Contains(t, p.Result(), "pod label with key level not allowed, the word is a palindrome")
			return ctx
		}).
		Assess("should create a pod with a non palindrome label", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			p := utils.RunCommand(fmt.Sprintf("kubectl apply -f %s", "fixtures/non-palindrome-pod-descriptor.yml"))
			assert.NoError(t, p.Err())
			return ctx
		}).Feature()

	testEnv.Test(t, policyWithoutSettingsFeature)
}

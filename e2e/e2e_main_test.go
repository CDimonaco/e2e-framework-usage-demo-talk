//go:build integration

package e2e_test

import (
	"context"
	"os"
	"testing"
	"time"

	kubewardenv1 "github.com/kubewarden/kubewarden-controller/api/policies/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/types"
	"sigs.k8s.io/e2e-framework/support/k3d"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

var (
	testEnv     env.Environment
	clusterName string
)

func TestMain(m *testing.M) {
	testEnv = env.New()
	clusterName = envconf.RandomName("policyteste2e", 16)
	namespace := envconf.RandomName("k3d-ns", 16)

	testEnv.Setup(
		envfuncs.CreateClusterWithOpts(k3d.NewProvider(), clusterName, k3d.WithImage("rancher/k3s:v1.31.5-k3s1")),
		envfuncs.CreateNamespace(namespace),
		installKubewarden(),
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyCluster(clusterName),
	)

	os.Exit(testEnv.Run(m))
}

func installKubewarden() types.EnvFunc {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		// Install kubewarden with helm
		manager := helm.New(config.KubeconfigFile())
		err := manager.RunRepo(helm.WithArgs("add", "kubewarden", "https://charts.kubewarden.io"))
		if err != nil {
			return ctx, err
		}
		err = manager.RunRepo(helm.WithArgs("update", "kubewarden"))
		if err != nil {
			return ctx, err
		}
		// install crds
		err = manager.RunInstall(
			helm.WithNamespace("kubewarden"),
			helm.WithArgs("--create-namespace"),
			helm.WithName("kubewarden-crds"),
			helm.WithReleaseName("kubewarden/kubewarden-crds"),
			helm.WithWait(),
			helm.WithVersion("1.13.0"),
		)
		if err != nil {
			return ctx, err
		}
		// install controller
		err = manager.RunInstall(
			helm.WithNamespace("kubewarden"),
			helm.WithReleaseName("kubewarden/kubewarden-controller"),
			helm.WithName("kubewarden-controller"),
			helm.WithWait(),
			helm.WithVersion("4.1.0"),
		)
		if err != nil {
			return ctx, err
		}
		// install defaults
		err = manager.RunInstall(
			helm.WithNamespace("kubewarden"),
			helm.WithReleaseName("kubewarden/kubewarden-defaults"),
			helm.WithName("kubewarden-defaults"),
			helm.WithWait(),
			helm.WithVersion("2.8.1"),
		)
		if err != nil {
			return ctx, err
		}

		// Wait for the policy server to be alive
		client, err := config.NewClient()
		if err != nil {
			return ctx, err
		}
		dep := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "policy-server-default", Namespace: "kubewarden"},
		}
		// wait for the deployment to finish becoming available
		err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(&dep, appsv1.DeploymentAvailable, v1.ConditionTrue), wait.WithTimeout(time.Minute*1))
		if err != nil {
			return ctx, err
		}

		// Load the kubewarden types into the test client
		scheme := config.Client().Resources().GetControllerRuntimeClient().Scheme()
		kubewardenv1.AddToScheme(scheme)

		return ctx, nil
	}
}

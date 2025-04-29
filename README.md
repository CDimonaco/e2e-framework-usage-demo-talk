# Palindrome Label Policy

This `kubewarden` policy ensures that no pod with a palindrome label key can be deployed on a Kubernetes cluster unless the label key is explicitly whitelisted in the policy settings.

## Introduction

This `kubewarden` policy can be configured with the following settings:

```json
{
  "allowed_palindromes": ["level"]
}
```

Settings are validated to ensure that only valid palindromes can be added to the `allowed_palindromes` list. If a non-palindrome is included, validation will fail.

The settings are optional. When not provided, the policy will reject all palindrome label keys by default.

## Code Organization

The code structure follows a standard Go module layout:

```
.
├── CODEOWNERS
├── LICENSE
├── Makefile
├── README.md
├── artifacthub-repo.yml
├── e2e
│   ├── e2e.bats
│   └── fixtures
│       ├── non-palindrome-label-pod.json
│       └── palindrome-label-pod.json
├── example
│   ├── palindrome-pod-descriptor.yml
│   ├── policy-descriptor.yml
│   └── policy-server-with-registry-secret.yml
├── go.mod
├── go.sum
├── internal
│   ├── policy
│   │   ├── settings.go
│   │   ├── settings_test.go
│   │   ├── validate.go
│   │   └── validate_test.go
│   └── word
│       ├── palindrome.go
│       └── palindrome_test.go
├── k3d.yml
├── main.go
├── metadata.yml
├── renovate.json
└── settings.sample.json
```

## Examples

### Policy Descriptor

```yaml
apiVersion: policies.kubewarden.io/v1
kind: ClusterAdmissionPolicy
metadata:
  name: palindrome-label-pod
spec:
  module: registry://ghcr.io/cdimonaco/policies/e2e-framework-usage-demo-talk:latest
  rules:
  - apiGroups: [""]
    apiVersions: ["v1"]
    resources: ["pods"]
    operations:
    - CREATE
    - UPDATE
  mutating: false
  settings:
    allowed_palindromes: ["level"]
```

This policy uses an OCI artifact built in the CI pipeline with the specified settings, allowing only the palindrome `level` as a valid label key.

### Rejected Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: palindrome-pod
  labels:
    aba: "yaba"
spec:
  containers:
    - name: hello
      image: hello-world:latest
```

### Accepted Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: palindrome-pod
  labels:
    level: "debug"
spec:
  containers:
    - name: hello
      image: hello-world:latest
```

## Using This Repository

### Running Unit Tests

```sh
make test
```

### Running E2E Tests

```sh
make e2e-tests
```

### Running the Linter

```sh
make lint
```

### Running E2E Cluster Tests

E2E Cluster Tests are implemented using [e2e-framework](https://github.com/kubernetes-sigs/e2e-framework). Each test run spawns a fresh `k3d` cluster, deploys `kubewarden` using the official Helm chart, applies the policy, and performs assertions to ensure that only pods with non-palindrome labels can be deployed. The latest tag of the OCI artifact built on GitHub is used for testing.

After each test run, the cluster is destroyed, ensuring reproducible and isolated testing environments.

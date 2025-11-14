# Claude notes

This repository should provide tooling and help for Konveyor community focusing on developers. Konveyor provides a toolkit for application analysis to help modernize them.

Information in markdown files are processed to CLAUDE.md and this file should be installed to `~/go/src/github.com/konveyor`, then `claude` should be executed in this directory, so claude code have access to konveyor components and can effectively help working with them.

More information could be found in each component directory that should be checked out locally in subdirectories. If they are missing, clone them.

To understand relations between components, check Dockerfiles, go.mod files or Makefiles in respective repos. CI and release process uses withub actions and workflows.

Component | Repository | Description
-- | -- | --
operator | https://github.com/konveyor/operator | kubernetes operator installing Konveyor
hub | https://github.com/konveyor/hub | API service for Konveyor, also provides start-minikube and install-tackle make targets, that can setup Konveyor locally with minikube
kantra | https://github.com/konveyor/kantra | CLI tool
analyzer-lsp | https://github.com/konveyor/analyzer-lsp | core component running the application analysis, used by Hub as well as kantra
rulesets | https://github.com/konveyor/rulesets | set of yaml rules describing patterns in analyzed application, that should be reported as issues
ci | https://github.com/konveyor/ci | CI workflows set and overall status, nightly jobs are executed there
go-konveyor-tests | https://github.com/konveyor/go-konveyor-tests | API test suite
kantra-cli-tests | https://github.com/konveyor-ecosystem/kantra-cli-tests | CLI test suite
enhancements | https://github.com/konveyor/enhancements | Konveyor documentation of new features, check open PRs to get currently open features

## AI Agents (experimental, PoC)

### CI Debugger Agent

Experiment created as part of day of learning Q3 2025.

An intelligent AI agent that automatically debugs GitHub Actions workflow failures and proposes fixes using OpenAI GPT-4.

**Location**: [agent-ci-debugger/](agent-ci-debugger/)

**Features**:
- Fetches workflow run data from GitHub Actions using GitHub CLI
- Parses and analyzes failed job logs
- Uses OpenAI GPT-4o-mini (configurable) to identify root causes and propose fixes
- Generates comprehensive reports with actionable recommendations

**Quick Start**:
```bash
cd agent-ci-debugger
export OPENAI_API_KEY="your-api-key"
./example-usage.sh https://github.com/konveyor/ci/actions/runs/RUNID
```

See [agent-ci-debugger/README.md](agent-ci-debugger/README.md) for detailed documentation.

## Locally running Konveyor

To run Konveyor locally, you can use the hub repository which provides convenient make targets for setting up minikube and installing Tackle.

### Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [minikube](https://minikube.sigs.k8s.io/docs/)
- [Docker](https://docs.docker.com/get-docker/) or [Podman](https://podman.io/getting-started/installation)

### Setup Steps

1. **Start minikube cluster**
   ```bash
   cd ../hub
   make start-minikube
   ```
   This downloads and runs the start-minikube.sh script from the tackle2-operator repository.

2. **Install Tackle**
   ```bash
   make install-tackle
   ```
   This downloads and runs the install-tackle.sh script that installs the Tackle operator and creates the necessary resources.

3. **Access the UI**
   Once installation is complete, you can access the Tackle UI through the exposed route or ingress depending on your cluster setup.

### Using Custom Images

You can customize the Tackle deployment by patching the Tackle Custom Resource with custom images. The Tackle CR supports various image fields:

```yaml
apiVersion: tackle.konveyor.io/v1alpha1
kind: Tackle
metadata:
  name: tackle
  namespace: konveyor-tackle
spec:
  # Hub image
  hub_image_fqin: "quay.io/konveyor/tackle2-hub:latest"

  # UI image
  ui_image_fqin: "quay.io/konveyor/tackle2-ui:latest"

  # Analyzer addon image
  analyzer_fqin: "quay.io/konveyor/tackle2-addon-analyzer:latest"

  # Language discovery addon image
  language_discovery_fqin: "quay.io/konveyor/tackle2-addon-discovery:latest"

  # Platform addon image
  platform_fqin: "quay.io/konveyor/tackle2-addon-platform:latest"

  # Provider images
  provider_java_image_fqin: "quay.io/konveyor/java-external-provider:latest"
  provider_python_image_fqin: "quay.io/konveyor/generic-external-provider:latest"
  provider_nodejs_image_fqin: "quay.io/konveyor/generic-external-provider:latest"

  # Keycloak and database images
  keycloak_sso_image_fqin: "quay.io/keycloak/keycloak:latest"
  keycloak_init_image_fqin: "quay.io/konveyor/tackle2-keycloak-init:latest"
  keycloak_database_image_fqin: "postgres:15"

  # OAuth proxy image
  oauth_image_fqin: "quay.io/openshift/origin-oauth-proxy:latest"

  # Image pull policy
  image_pull_policy: "Always"
```

To apply a custom image configuration:

```bash
kubectl patch tackle tackle -n konveyor-tackle --type=merge -p='{"spec":{"hub_image_fqin":"your-custom-hub-image:tag"}}'
```

Or create a custom Tackle CR file and apply it:

```bash
kubectl apply -f custom-tackle-cr.yaml
```

## Running tests

### API tests

Stored in `go-konveyor-tests` repository, executed from Makefile `make test-tier0`, there are tiers 0, 1 and 2. Tests need point to `hub` path from Konveyor installation with `HUB_BASE_URL` environment variable.


## Working with rules and rulesets


# kantra CLI


